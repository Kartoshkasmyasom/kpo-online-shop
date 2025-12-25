package domain

import "encoding/json"

type CreateOrderReq struct {
	UserID      string      `json:"user_id"`
	Amount      json.Number `json:"amount"`
	Description string      `json:"description"`
}

type ListOrderReq struct {
	UserID string `json:"user_id"`
}

type StatusReq struct {
	ID string `json:"id"`
}

type StatusResp struct {
	Status OrderStatus `json:"status"`
}

type ErrResp struct {
	Error string `json:"error"`
}
