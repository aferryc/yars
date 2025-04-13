package transport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aferryc/yars/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer handles the consumption of Kafka messages
type Consumer struct {
	client  *kgo.Client
	topic   string
	handler EventHandler
	cfg     *config.KafkaConfig
}

type ConsumerConfig struct {
	Brokers  []string
	GroupID  string
	Topic    string
	ClientID string
	Handler  EventHandler
}

type EventHandler interface {
	ProcessEvent([]byte) error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.KafkaConfig, topic string, handler EventHandler) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.BrokerList...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(topic),
		kgo.ClientID(cfg.ClientID),
		kgo.FetchMinBytes(1e3), // 1KB
		kgo.FetchMaxBytes(1e6), // 1MB
		kgo.FetchMaxWait(5 * time.Second),
		kgo.AutoCommitInterval(5 * time.Second),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Consumer{
		client:  client,
		topic:   topic,
		handler: handler,
		cfg:     cfg,
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting consumer for topic: %s", c.topic)

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return errors.New("client closed")
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			// Log all errors, but continue processing
			for _, err := range errs {
				log.Printf("Error polling: %v", err)
			}
		}

		// Process all fetched records
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			p.EachRecord(func(record *kgo.Record) {
				// Process each record
				log.Printf("Received message: partition=%d offset=%d", record.Partition, record.Offset)

				// Process the event
				if err := c.handler.ProcessEvent(record.Value); err != nil {
					log.Printf("Error processing event: %v", err)
					// Depending on your strategy, you might want to:
					// - Skip this message and continue
					// - Retry a certain number of times
					// - Send to a dead-letter queue
					// - Pause consumption and alert
				}
			})
		})
	}
}

// Close properly shuts down the consumer
func (c *Consumer) Close() {
	if c.client != nil {
		c.client.Close()
		log.Println("Kafka consumer closed")
	}
}
