package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"payments/internal/domain"
)

var (
	ErrNoAccount      = errors.New("no account")
	ErrNotEnoughMoney = errors.New("not enough money")
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type PaymentResult struct {
	MessageID uuid.UUID `json:"message_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Status    string    `json:"status"`
}

func (s *Store) CreateAccount(userID string) {
	_, _ = s.db.Exec(`insert into accounts(user_id, balance) values ($1, 0)
					  on conflict (user_id) do nothing`, userID)
}

func (s *Store) TopUp(userID string, amount domain.Money) error {
	if amount <= 0 {
		return errors.New("amount must be > 0")
	}

	res, err := s.db.Exec(`update accounts set balance = balance + $2 where user_id = $1`, userID, int64(amount))
	if err != nil {
		return err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return ErrNoAccount
	}
	return nil
}

func (s *Store) Balance(userID string) (domain.Money, error) {
	var b int64
	err := s.db.QueryRow(`select balance from accounts where user_id = $1`, userID).Scan(&b)
	if err == sql.ErrNoRows {
		return 0, ErrNoAccount
	}
	if err != nil {
		return 0, err
	}
	return domain.Money(b), nil
}

func InsertPaymentResultOutbox(ctx context.Context, tx *sql.Tx, orderID uuid.UUID, status domain.PaymentStatus) error {
	msgID := uuid.New()
	ev := PaymentResult{
		MessageID: msgID,
		OrderID:   orderID,
		Status:    string(status),
	}
	payload, _ := json.Marshal(ev)

	_, err := tx.ExecContext(ctx,
		`insert into payments_outbox(message_id, topic, key, payload) values ($1,$2,$3,$4)`,
		msgID, "payments.result", orderID.String(), payload,
	)
	return err
}


func (s *Store) Pay(orderID uuid.UUID, userID string, amount domain.Money) (domain.Payment, error) {
	if amount <= 0 {
		return domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PayFailed}, errors.New("amount must be > 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Payment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	{
		var status string
		var uid string
		var amt int64
		err := tx.QueryRowContext(ctx,
			`select user_id, amount, status from payments where order_id = $1`, orderID,
		).Scan(&uid, &amt, &status)

		if err == nil {
			return domain.Payment{OrderID: orderID, UserID: uid, Amount: domain.Money(amt), Status: domain.PaymentStatus(status)}, nil
		}
		if err != sql.ErrNoRows {
			return domain.Payment{}, err
		}
	}

	var bal int64
	err = tx.QueryRowContext(ctx,
		`select balance from accounts where user_id = $1 for update`, userID,
	).Scan(&bal)
	if err == sql.ErrNoRows {
		p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PayFailed}
		_, _ = tx.ExecContext(ctx,
			`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
			orderID, userID, int64(amount), string(p.Status),
		)
		_ = InsertPaymentResultOutbox(ctx, tx, orderID, p.Status)
		_ = tx.Commit()
		return p, ErrNoAccount
	}
	if err != nil {
		return domain.Payment{}, err
	}

	if bal < int64(amount) {
		p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PayFailed}
		_, _ = tx.ExecContext(ctx,
			`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
			orderID, userID, int64(amount), string(p.Status),
		)
		_ = InsertPaymentResultOutbox(ctx, tx, orderID, p.Status)
		_ = tx.Commit()
		return p, ErrNotEnoughMoney
	}

	_, err = tx.ExecContext(ctx,
		`update accounts set balance = balance - $2 where user_id = $1`, userID, int64(amount),
	)
	if err != nil {
		return domain.Payment{}, err
	}

	p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PaySuccess}
	_, err = tx.ExecContext(ctx,
		`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
		orderID, userID, int64(amount), string(p.Status),
	)
	if err != nil {
		return domain.Payment{}, err
	}
	if err = InsertPaymentResultOutbox(ctx, tx, orderID, p.Status); err != nil {
		return domain.Payment{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Payment{}, err
	}
	return p, nil
}

func PayInTx(ctx context.Context, tx *sql.Tx, orderID uuid.UUID, userID string, amount domain.Money) (domain.Payment, error) {
	{
		var status string
		var uid string
		var amt int64
		err := tx.QueryRowContext(ctx,
			`select user_id, amount, status from payments where order_id = $1`, orderID,
		).Scan(&uid, &amt, &status)

		if err == nil {
			return domain.Payment{OrderID: orderID, UserID: uid, Amount: domain.Money(amt), Status: domain.PaymentStatus(status)}, nil
		}
		if err != sql.ErrNoRows {
			return domain.Payment{}, err
		}
	}

	var bal int64
	err := tx.QueryRowContext(ctx,
		`select balance from accounts where user_id = $1 for update`, userID,
	).Scan(&bal)
	if err == sql.ErrNoRows {
		p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PayFailed}
		_, e := tx.ExecContext(ctx,
			`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
			orderID, userID, int64(amount), string(p.Status),
		)
		if e != nil {
			return domain.Payment{}, e
		}
		return p, ErrNoAccount
	}
	if err != nil {
		return domain.Payment{}, err
	}

	if bal < int64(amount) {
		p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PayFailed}
		_, e := tx.ExecContext(ctx,
			`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
			orderID, userID, int64(amount), string(p.Status),
		)
		if e != nil {
			return domain.Payment{}, e
		}
		return p, ErrNotEnoughMoney
	}

	_, err = tx.ExecContext(ctx,
		`update accounts set balance = balance - $2 where user_id = $1`,
		userID, int64(amount),
	)
	if err != nil {
		return domain.Payment{}, err
	}

	p := domain.Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: domain.PaySuccess}
	_, err = tx.ExecContext(ctx,
		`insert into payments(order_id, user_id, amount, status) values ($1,$2,$3,$4)`,
		orderID, userID, int64(amount), string(p.Status),
	)
	if err != nil {
		return domain.Payment{}, err
	}
	return p, nil
}
