package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"

	"orders/internal/domain"
)

type PaymentResult struct {
	MessageID uuid.UUID `json:"message_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Status    string    `json:"status"` // "SUCCESS"/"FAILED"
}

type PaymentResultConsumer struct {
	db *sql.DB
}

func NewPaymentResultConsumer(db *sql.DB) *PaymentResultConsumer {
	return &PaymentResultConsumer{db: db}
}

type paymentResultHandler struct {
	c *PaymentResultConsumer
}

func (h *paymentResultHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *paymentResultHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *paymentResultHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.c.handle(sess.Context(), msg.Value); err != nil {
			log.Printf("payments.result handle error: %v", err)
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (c *PaymentResultConsumer) Run(ctx context.Context) {
	cg, err := newConsumerGroup("orders-service")
	if err != nil {
		log.Fatal(err)
	}
	defer cg.Close()

	h := &paymentResultHandler{c: c}

	for {
		if err := cg.Consume(ctx, []string{"payments.result"}, h); err != nil {
			log.Printf("orders consumer error: %v", err)
			time.Sleep(500 * time.Millisecond)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *PaymentResultConsumer) handle(ctx context.Context, payload []byte) error {
	var ev PaymentResult
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}

	newStatus := domain.OrderCancelled
	if ev.Status == "SUCCESS" {
		newStatus = domain.OrderFinished
	}

	
	_, err := c.db.ExecContext(ctx, `update orders set status = $2 where id = $1`, ev.OrderID, string(newStatus))
	return err
}
