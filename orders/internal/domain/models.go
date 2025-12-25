package domain

import (
	"github.com/google/uuid"
)

type Money int64

type OrderStatus string

const (
	OrderNew OrderStatus = "NEW"
	OrderFinished       OrderStatus = "FINISHED"
	OrderCancelled      OrderStatus = "CANCELLED"
)

type Order struct {
	ID 	        uuid.UUID   `json:"id"`
	UserID      string      `json:"user_id"`
	Amount 	    Money       `json:"amount"`
	Description string      `json:"description"`
	Status      OrderStatus `json:"status"`
}