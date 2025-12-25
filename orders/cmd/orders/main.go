package main

import (
	// "context"
	// "encoding/json"
	// "log"
	// "net/http"
	// "strconv"
	// "strings"

	// "github.com/google/uuid"
	"orders/internal/app"
)

// type createOrderReq struct {
// 	UserID      string      `json:"user_id"`
// 	Amount      json.Number `json:"amount"`
// 	Description string      `json:"description"`
// }

// type listOrderReq struct {
// 	UserID string `json:"user_id"`
// }

// type statusReq struct {
// 	ID string `json:"id"`
// }

// type statusResp struct {
// 	Status OrderStatus `json:"status"`
// }

// type errResp struct {
// 	Error string `json:"error"`
// }

// func writeJSON(w http.ResponseWriter, code int, v any) error {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	err := json.NewEncoder(w).Encode(v)
// 	return err
// }

// func parseAmount(amount json.Number) (int64, error) {
// 	if val, err := amount.Int64(); err == nil {
// 		return val, nil
// 	}
// 	str := amount.String()

// 	if strings.Contains(str, ".") {
// 		parts := strings.Split(str, ".")
// 		str = parts[0]
// 	}

// 	return strconv.ParseInt(str, 10, 64)
// }

// func makeHandleCreateOrder(s *OrdersStore) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			writeJSON(w, http.StatusMethodNotAllowed, errResp{Error: "method not allowed"})
// 			return
// 		}

// 		var req createOrderReq
// 		err := json.NewDecoder(r.Body).Decode(&req)
// 		if err != nil {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "bad json: " + err.Error()})
// 			return
// 		}
// 		if req.UserID == "" {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "empty user_id"})
// 			return
// 		}
// 		var amount int64
// 		amount, err = parseAmount(req.Amount)

// 		if amount <= 0 {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "amount should be greater than 0"})
// 			return
// 		}
// 		if len(req.Description) > 200 {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "description must contain less than 200 symbols"})
// 			return
// 		}

// 		var o Order
// 		o, err = s.CreateOrder(req.UserID, Money(amount), req.Description)
// 		if err != nil {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: err.Error()})
// 			return
// 		}

// 		writeJSON(w, http.StatusOK, o)

// 	}
// }

// func makeHandleListOrders(s *OrdersStore) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			writeJSON(w, http.StatusMethodNotAllowed, errResp{Error: "method not allowed"})
// 			return
// 		}

// 		var req listOrderReq
// 		err := json.NewDecoder(r.Body).Decode(&req)
// 		if err != nil {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "bad json: " + err.Error()})
// 			return
// 		}
// 		if req.UserID == "" {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "empty user_id"})
// 			return
// 		}
// 		orders, _ := s.ListOrders(req.UserID)

// 		if err = writeJSON(w, http.StatusOK, orders); err != nil {
// 			writeJSON(w, http.StatusInternalServerError, errResp{Error: err.Error()})
// 		}
// 	}
// }

// func makeHandleGetStatus(s *OrdersStore) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			writeJSON(w, http.StatusMethodNotAllowed, errResp{Error: "method not allowed"})
// 			return
// 		}

// 		var req statusReq
// 		err := json.NewDecoder(r.Body).Decode(&req)
// 		if err != nil {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "bad json: " + err.Error()})
// 			return
// 		}

// 		if req.ID == "" {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "empty id"})
// 			return
// 		}

// 		orderUUID, err := uuid.Parse(req.ID)
// 		if err != nil {
// 			writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid orderID format"})
// 			return
// 		}
// 		var o OrderStatus
// 		o, err = s.GetStatus(orderUUID)
// 		resp := statusResp{
// 			Status: o,
// 		}
// 		if err = writeJSON(w, http.StatusOK, resp); err != nil {
// 			writeJSON(w, http.StatusInternalServerError, errResp{Error: err.Error()})
// 		}

// 	}
// }

func main() {
	// db, err := OpenDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// store := NewOrdersStore(db)

	// ctx := context.Background()

	// prod, err := newSyncProducer()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pub := NewOutboxPublisher(db, prod)
	// go pub.Run(ctx)

	// resCons := NewPaymentResultConsumer(db)
	// go resCons.Run(ctx)

	// mux := http.NewServeMux()

	// mux.HandleFunc("/create", makeHandleCreateOrder(store))
	// mux.HandleFunc("/list", makeHandleListOrders(store))
	// mux.HandleFunc("/status", makeHandleGetStatus(store))

	// log.Println("orders listening on :8080")
	// log.Fatal(http.ListenAndServe(":8080", mux))
	app.Run()
}
