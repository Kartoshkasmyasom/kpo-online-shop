package kafka

import (
	"errors"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

func kafkaBrokers() ([]string, error) {
	b := os.Getenv("KAFKA_BROKERS")
	if b == "" {
		return nil, errors.New("KAFKA_BROKERS is empty")
	}
	parts := strings.Split(b, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("KAFKA_BROKERS is empty")
	}
	return out, nil
}

func newConsumerGroup(group string) (sarama.ConsumerGroup, error) {
	brokers, err := kafkaBrokers()
	if err != nil {
		return nil, err
	}
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	cfg.Version = sarama.V3_6_0_0

	cfg.Consumer.Offsets.AutoCommit.Enable = true

	return sarama.NewConsumerGroup(brokers, group, cfg)
}

func NewSyncProducer() (sarama.SyncProducer, error) {
	brokers, err := kafkaBrokers()
	if err != nil {
		return nil, err
	}

	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Return.Successes = true
	cfg.Version = sarama.V3_6_0_0

	return sarama.NewSyncProducer(brokers, cfg)
}
