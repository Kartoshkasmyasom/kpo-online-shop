package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"payments/internal/domain"
	"payments/internal/store"

	"github.com/google/uuid"
)

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	return err
}

func makeHandleCreatePayment(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.CreateAccountReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}

		if req.UserID == "" {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty user_id"})
			return
		}
		s.CreateAccount(req.UserID)
		var balance domain.Money
		balance, err = s.Balance(req.UserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "could not get balance: " + err.Error()})
			return
		}

		resp := domain.CreateAccountResp{
			UserID: req.UserID,
			Balance: json.Number(strconv.FormatInt(int64(balance), 10)),
		}
		err = writeJSON(w, http.StatusOK, resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "failed to write response: " + err.Error()})	 	
			return
		}
	}
}

func parseAmount(amount json.Number) (int64, error) {
    if val, err := amount.Int64(); err == nil {
        return val, nil
    }
    
    str := amount.String()
    
    if strings.Contains(str, ".") {
        parts := strings.Split(str, ".")
        str = parts[0]
    }
    
    return strconv.ParseInt(str, 10, 64)
}

func makeHandleTopUp(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.TopUpReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}

		if req.UserID == "" {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty user_id"})
			return
		}
		var amount int64
		amount, err = parseAmount(req.Amount)
		if amount < 0 {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "can not deposit negative amount"})
			return
		}

		if err = s.TopUp(req.UserID, domain.Money(amount)); err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "could not top up: " + err.Error()})
			return
		} 
		var balance domain.Money
		balance, err = s.Balance(req.UserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "could not get balance: " + err.Error()})
			return
		}

		resp := domain.TopUpResp {
			Balance: json.Number(strconv.FormatInt(int64(balance), 10)),
		}
		err = writeJSON(w, http.StatusOK, resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "failed to write response: " + err.Error()})	 	
			return
		}
	}
}

func makeHandleBalance(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.BalanceReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}

		if req.UserID == "" {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty user_id"})
			return
		}

		var balance domain.Money
		balance, err = s.Balance(req.UserID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "could not get balance: " + err.Error()})
			return
		}

		resp := domain.BalanceResp {
			Balance: json.Number(strconv.FormatInt(int64(balance), 10)),
		}
		err = writeJSON(w, http.StatusOK, resp)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "failed to write response: " + err.Error()})	 	
			return
		}
	}
}


func makeHandlePay(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.PayReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}

		if req.UserID == ""{
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty user_id"})
			return
		}
		if req.OrderID == ""{
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty order_id"})
			return
		}
		var amount int64
		amount, err = parseAmount(req.Amount)
		if amount < 0 {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "can not pay negative amount"})
			return
		}
		var orderUUID uuid.UUID
		orderUUID, err = uuid.Parse(req.OrderID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "invalid orderID format"})
			return
		}

		var payment domain.Payment
		if payment, err = s.Pay(orderUUID, req.UserID, domain.Money(amount)); err != nil {
			writeJSON(w, http.StatusOK, domain.ErrResp{Error: "could not pay: " + err.Error()})
			return
		} 
		err = writeJSON(w, http.StatusOK, payment)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: "failed to write response"})	 	
			return
		}
	}
}


func RegisterRoutes(mux *http.ServeMux, st *store.Store) {
	mux.HandleFunc("/create", makeHandleCreatePayment(st))
	mux.HandleFunc("/topup", makeHandleTopUp(st))
	mux.HandleFunc("/balance", makeHandleBalance(st))
	mux.HandleFunc("/pay", makeHandlePay(st))
}