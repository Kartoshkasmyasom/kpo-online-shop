package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"payments/internal/store"
	"payments/internal/domain"
)

type PaymentRequested struct {
	MessageID   uuid.UUID `json:"message_id"`
	OrderID     uuid.UUID `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
}

type PaymentRequestConsumer struct {
	db    *sql.DB
	store *store.Store
}

func NewPaymentRequestConsumer(db *sql.DB, store *store.Store) *PaymentRequestConsumer {
	return &PaymentRequestConsumer{db: db, store: store}
}

type paymentRequestHandler struct {
	c *PaymentRequestConsumer
}

func (h *paymentRequestHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *paymentRequestHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *paymentRequestHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.c.handleMessage(sess.Context(), msg); err != nil {
			log.Printf("payments.request handle error: %v", err)
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (c *PaymentRequestConsumer) Run(ctx context.Context) {
	cg, err := newConsumerGroup("payments-service")
	if err != nil {
		log.Fatal(err)
	}
	defer cg.Close()

	handler := &paymentRequestHandler{c: c}

	for {
		if err := cg.Consume(ctx, []string{"payments.request"}, handler); err != nil {
			log.Printf("consumer error: %v", err)
			time.Sleep(500 * time.Millisecond)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *PaymentRequestConsumer) handleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var ev PaymentRequested
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		return err
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`insert into payments_inbox(message_id) values ($1) on conflict do nothing`,
		ev.MessageID,
	)
	if err != nil {
		return err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return tx.Commit()
	}

	p, _ := store.PayInTx(ctx, tx, ev.OrderID, ev.UserID, domain.Money(ev.Amount))
	_ = store.InsertPaymentResultOutbox(ctx, tx, ev.OrderID, p.Status)

	return tx.Commit()
}
