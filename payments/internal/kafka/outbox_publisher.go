package kafka

import (
	"context"
	"database/sql"
	"time"

	"github.com/IBM/sarama"
)

type OutboxPublisher struct {
	db       *sql.DB
	producer sarama.SyncProducer
}

func NewOutboxPublisher(db *sql.DB, producer sarama.SyncProducer) *OutboxPublisher {
	return &OutboxPublisher{db: db, producer: producer}
}

func (p *OutboxPublisher) Run(ctx context.Context) {
	t := time.NewTicker(400 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = p.tick(ctx)
		}
	}
}

func (p *OutboxPublisher) tick(ctx context.Context) error {
	var (
		id      int64
		topic   string
		key     string
		payload []byte
	)

	err := p.db.QueryRowContext(ctx,
		`select id, topic, key, payload from payments_outbox
		 where published_at is null
		 order by id
		 limit 1`,
	).Scan(&id, &topic, &key, &payload)

	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	_, _ = p.db.ExecContext(ctx, `update payments_outbox set published_at = now() where id = $1`, id)
	return nil
}
