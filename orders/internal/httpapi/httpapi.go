package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"orders/internal/domain"
	"orders/internal/store"

	"github.com/google/uuid"
)

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	return err
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

func makeHandleCreateOrder(s *store.OrdersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.CreateOrderReq
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

		if amount <= 0 {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "amount should be greater than 0"})
			return
		}
		if len(req.Description) > 200 {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "description must contain less than 200 symbols"})
			return
		}

		var o domain.Order
		o, err = s.CreateOrder(req.UserID, domain.Money(amount), req.Description)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, o)

	}
}

func makeHandleListOrders(s *store.OrdersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.ListOrderReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}
		if req.UserID == "" {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty user_id"})
			return
		}
		orders, _ := s.ListOrders(req.UserID)

		if err = writeJSON(w, http.StatusOK, orders); err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: err.Error()})
		}
	}
}

func makeHandleGetStatus(s *store.OrdersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, domain.ErrResp{Error: "method not allowed"})
			return
		}

		var req domain.StatusReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "bad json: " + err.Error()})
			return
		}

		if req.ID == "" {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "empty id"})
			return
		}

		orderUUID, err := uuid.Parse(req.ID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, domain.ErrResp{Error: "invalid orderID format"})
			return
		}
		var o domain.OrderStatus
		o, err = s.GetStatus(orderUUID)
		resp := domain.StatusResp{
			Status: o,
		}
		if err = writeJSON(w, http.StatusOK, resp); err != nil {
			writeJSON(w, http.StatusInternalServerError, domain.ErrResp{Error: err.Error()})
		}

	}
}

func RegisterRoutes(mux *http.ServeMux, st *store.OrdersStore) {
	mux.HandleFunc("/create", makeHandleCreateOrder(st))
	mux.HandleFunc("/status", makeHandleGetStatus(st))
	mux.HandleFunc("/list", makeHandleListOrders(st))
}