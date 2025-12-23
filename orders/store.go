package main

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"payments/models.go"
)

var (
	ErrNoAccount      = errors.New("no account")
	ErrNotEnoughMoney = errors.New("not enough money")
)

type Store struct {
	mu       sync.Mutex
	accounts map[string]Money
	payments map[uuid.UUID]Payment
}