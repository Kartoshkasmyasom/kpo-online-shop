package main

import (
	"github.com/google/uuid"
)

type Money int64

type OrderStatus string

const (
	OrderPaymentPending OrderStatus = "PAYMENT_PENDING"
	OrderPaid           OrderStatus = "PAID"
	OrderCancelled      OrderStatus = "CANCELLED"
)

type Order struct {
	ID 	   uuid.UUID   `json:"id"`
	UserID string      `json:"user_id"`
	Amount Money       `json:"amount"`
	Status OrderStatus `json:"status"`
}