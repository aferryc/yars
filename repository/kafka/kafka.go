package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaProducer struct {
	client *kgo.Client
}

func NewKafkaRepository(client *kgo.Client) *KafkaProducer {
	return &KafkaProducer{
		client: client,
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, message any) error {
	// Validate topic name
	if topic == "" {
		return errors.New("topic cannot be empty")
	}

	var msgBytes []byte
	var err error

	switch v := message.(type) {
	case string:
		msgBytes = []byte(v)
	case []byte:
		msgBytes = v
	default:
		msgBytes, err = json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %v", err)
		}
	}

	var keyBytes []byte
	if key != "" {
		keyBytes = []byte(key)
	}

	record := &kgo.Record{
		Topic: topic,
		Value: msgBytes,
		Key:   keyBytes,
	}

	err = p.client.ProduceSync(ctx, record).FirstErr()
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}
