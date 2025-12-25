package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"orders/internal/domain"
)

var (
	ErrNoOrder          = errors.New("no order")
	ErrInvalidPrice     = errors.New("order price should be greater than 0")
	ErrDescriptionLimit = errors.New("description should contain maximum of 200 symbols")
)

type OrdersStore struct {
	db *sql.DB
}

func NewOrdersStore(db *sql.DB) *OrdersStore {
	return &OrdersStore{db: db}
}

type PaymentRequested struct {
	MessageID    uuid.UUID `json:"message_id"`
	OrderID      uuid.UUID `json:"order_id"`
	UserID       string    `json:"user_id"`
	Amount       int64     `json:"amount"`
	Description  string    `json:"description"`
}

func (s *OrdersStore) CreateOrder(userID string, amount domain.Money, description string) (domain.Order, error) {
	if userID == "" {
		return domain.Order{}, errors.New("empty user_id")
	}
	if amount <= 0 {
		return domain.Order{}, ErrInvalidPrice
	}
	if len(description) > 200 {
		return domain.Order{}, ErrDescriptionLimit
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	orderID := uuid.New()
	o := domain.Order{
		ID:          orderID,
		UserID:      userID,
		Amount:      amount,
		Description: description,
		Status:      domain.OrderNew,
	}

	_, err = tx.ExecContext(ctx,
		`insert into orders(id, user_id, amount, description, status) values ($1,$2,$3,$4,$5)`,
		o.ID, o.UserID, int64(o.Amount), o.Description, string(o.Status),
	)
	if err != nil {
		return domain.Order{}, err
	}

	msgID := uuid.New()
	ev := PaymentRequested{
		MessageID:   msgID,
		OrderID:     o.ID,
		UserID:      o.UserID,
		Amount:      int64(o.Amount),
		Description: o.Description,
	}
	payload, _ := json.Marshal(ev)

	_, err = tx.ExecContext(ctx,
		`insert into orders_outbox(message_id, topic, key, payload) values ($1,$2,$3,$4)`,
		msgID, "payments.request", o.ID.String(), payload,
	)
	if err != nil {
		return domain.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return o, nil
}

func (s *OrdersStore) GetStatus(id uuid.UUID) (domain.OrderStatus, error) {
	var st string
	err := s.db.QueryRow(`select status from orders where id = $1`, id).Scan(&st)
	if err == sql.ErrNoRows {
		return "", ErrNoOrder
	}
	if err != nil {
		return "", err
	}
	return domain.OrderStatus(st), nil
}

func (s *OrdersStore) ListOrders(userID string) ([]domain.Order, error) {
	rows, err := s.db.Query(`select id, user_id, amount, description, status from orders where user_id = $1 order by created_at desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Order{}
	for rows.Next() {
		var o domain.Order
		var amt int64
		var st string
		if err := rows.Scan(&o.ID, &o.UserID, &amt, &o.Description, &st); err != nil {
			return nil, err
		}
		o.Amount = domain.Money(amt)
		o.Status = domain.OrderStatus(st)
		out = append(out, o)
	}
	return out, nil
}
