package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type errResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func mustReadBody(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	b = bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))
	return b, nil
}

type Front struct {
	ordersURL   string
	paymentsURL string
	client      *http.Client
}

func (f *Front) proxyPostJSON(w http.ResponseWriter, r *http.Request, target string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResp{Error: "method not allowed"})
		return
	}

	body, err := mustReadBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "cannot read body: " + err.Error()})
		return
	}

	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{Error: err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errResp{Error: "backend request failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (f *Front) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func main() {
	ordersURL := os.Getenv("ORDERS_URL")
	if ordersURL == "" {
		ordersURL = "http://orders:8080"
	}
	paymentsURL := os.Getenv("PAYMENTS_URL")
	if paymentsURL == "" {
		paymentsURL = "http://payments:8080"
	}

	f := &Front{
		ordersURL:   ordersURL,
		paymentsURL: paymentsURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", f.handleIndex)

	mux.HandleFunc("/api/payments/create", func(w http.ResponseWriter, r *http.Request) {
		f.proxyPostJSON(w, r, f.paymentsURL+"/create")
	})
	mux.HandleFunc("/api/payments/topup", func(w http.ResponseWriter, r *http.Request) {
		f.proxyPostJSON(w, r, f.paymentsURL+"/topup")
	})
	mux.HandleFunc("/api/payments/balance", func(w http.ResponseWriter, r *http.Request) {
		f.proxyPostJSON(w, r, f.paymentsURL+"/balance")
	})

	mux.HandleFunc("/api/orders/create", func(w http.ResponseWriter, r *http.Request) {
		f.proxyPostJSON(w, r, f.ordersURL+"/create")
	})
	mux.HandleFunc("/api/orders/list", func(w http.ResponseWriter, r *http.Request) {
		f.proxyPostJSON(w, r, f.ordersURL+"/list")
	})
	mux.HandleFunc("/api/orders/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errResp{Error: "method not allowed"})
			return
		}
		body, err := mustReadBody(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResp{Error: err.Error()})
			return
		}

		var m map[string]any
		if json.Unmarshal(body, &m) == nil {
			if v, ok := m["id"]; ok {
				m["order_id"] = v
			}
			if v, ok := m["order_id"]; ok {
				m["id"] = v
			}
			body, _ = json.Marshal(m)
		}

		req, _ := http.NewRequest(http.MethodPost, f.ordersURL+"/status", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := f.client.Do(req)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, errResp{Error: err.Error()})
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	log.Println("frontend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

const indexHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Gozon UI</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; max-width: 900px; }
    .row { display: flex; gap: 16px; flex-wrap: wrap; }
    .card { border: 1px solid #ddd; border-radius: 10px; padding: 14px; width: 500px; }
    input, textarea { width: 90%; padding: 8px; margin: 6px 0; }
    button { padding: 8px 12px; cursor: pointer; }
    pre { background: #f7f7f7; padding: 10px; border-radius: 8px; overflow: auto; }
    .small { color: #666; font-size: 12px; }
  </style>
</head>
<body>
  <h2>Фронтенд для нашего магазина</h2>

  <div class="row">
    <div class="card">
      <h3>Payments: create account</h3>
      <input id="p_user_create" placeholder="user_id (например u1)" />
      <button onclick="callApi('/api/payments/create', {user_id: val('p_user_create')})">Create</button>
      <pre id="out_p_create"></pre>
    </div>

    <div class="card">
      <h3>Payments: topup</h3>
      <input id="p_user_topup" placeholder="user_id" />
      <input id="p_amount_topup" placeholder="amount (например 100)" />
      <button onclick="callApi('/api/payments/topup', {user_id: val('p_user_topup'), amount: num('p_amount_topup')})">TopUp</button>
      <pre id="out_p_topup"></pre>
    </div>

    <div class="card">
      <h3>Payments: balance</h3>
      <input id="p_user_balance" placeholder="user_id" />
      <button onclick="callApi('/api/payments/balance', {user_id: val('p_user_balance')})">Balance</button>
      <pre id="out_p_balance"></pre>
    </div>

    <div class="card">
      <h3>Orders: create</h3>
      <input id="o_user_create" placeholder="user_id" />
      <input id="o_amount_create" placeholder="amount (например 30)" />
      <textarea id="o_desc_create" placeholder="description (<=200 символов)"></textarea>
      <button onclick="createOrder()">Create order</button>
      <pre id="out_o_create"></pre>
      <div class="small">После create — вставь order_id ниже, чтобы смотреть статус.</div>
    </div>

    <div class="card">
      <h3>Orders: status</h3>
      <input id="o_id_status" placeholder="order_id (uuid)" />
      <button onclick="callApi('/api/orders/status', {id: val('o_id_status')})">Get status</button>
      <pre id="out_o_status"></pre>
    </div>

    <div class="card">
      <h3>Orders: list</h3>
      <input id="o_user_list" placeholder="user_id" />
      <button onclick="callApi('/api/orders/list', {user_id: val('o_user_list')})">List</button>
      <pre id="out_o_list"></pre>
    </div>
  </div>

<script>
function val(id){ return document.getElementById(id).value.trim(); }
function num(id){ const v = val(id); return v === "" ? 0 : Number(v); }

async function callApi(path, payload){
  const outId = {
    "/api/payments/create":"out_p_create",
    "/api/payments/topup":"out_p_topup",
    "/api/payments/balance":"out_p_balance",
    "/api/orders/create":"out_o_create",
    "/api/orders/status":"out_o_status",
    "/api/orders/list":"out_o_list"
  }[path];

  const out = document.getElementById(outId);
  out.textContent = "loading...";

  try {
    const resp = await fetch(path, {
      method: "POST",
      headers: {"Content-Type":"application/json"},
      body: JSON.stringify(payload)
    });
    const text = await resp.text();
    out.textContent = text;
    return text;
  } catch(e){
    out.textContent = "ERROR: " + e;
    return "";
  }
}

async function createOrder(){
  const text = await callApi("/api/orders/create", {
    user_id: val("o_user_create"),
    amount: num("o_amount_create"),
    description: val("o_desc_create")
  });

  // если ответ содержит id — подсунем в status
  try {
    const obj = JSON.parse(text);
    if (obj && obj.id) document.getElementById("o_id_status").value = obj.id;
  } catch(e) {}
}
</script>
</body>
</html>`
