package main

import (
	"github.com/google/uuid"
)

type Money int64

type PaymentStatus string

const (
	PaySuccess PaymentStatus = "SUCCESS"
	PayFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	OrderID uuid.UUID     `json:"order_id"`
	UserID  string		  `json:"user_id"`
	Amount  Money		  `json:"amount"`
	Status  PaymentStatus `json:"status"`
}