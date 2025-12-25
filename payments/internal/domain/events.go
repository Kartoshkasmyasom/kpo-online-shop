package domain

import "encoding/json"

type CreateAccountReq struct {
	UserID string `json:"user_id"`
}

type CreateAccountResp struct {
	UserID  string      `json:"user_id"`
	Balance json.Number `json:"balance"`
}

type TopUpReq struct {
	UserID string       `json:"user_id"`
	Amount json.Number  `json:"amount"`
}

type TopUpResp struct {
	Balance json.Number `json:"balance"`
}

type BalanceReq struct {
	UserID string `json:"user_id"`
}

type BalanceResp struct {
	Balance json.Number `json:"balance"`
}

type PayReq struct {
	OrderID string       `json:"order_id"`
	UserID  string       `json:"user_id"`
	Amount  json.Number  `json:"amount"`
}

type ErrResp struct {
	Error string `json:"error"`
}