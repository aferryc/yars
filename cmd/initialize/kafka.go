package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Set default values for simplicity
const maxKafkaRetries = 3
const retryBackoff = 100 * time.Millisecond
const connectRetry = 5
const connectBackoff = 500 * time.Millisecond

func NewKafkaProducer(brokers []string, clientID string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
	}

	opts = append(opts, kgo.ClientID(clientID))
	opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	opts = append(opts, kgo.RetryBackoffFn(func(attempt int) time.Duration {
		return retryBackoff * time.Duration(attempt+1)
	}))
	opts = append(opts, kgo.RetryTimeout(retryBackoff*time.Duration(maxKafkaRetries+1)))

	var client *kgo.Client
	var err error

	for i := 0; i < connectRetry; i++ {
		client, err = kgo.NewClient(opts...)
		if err == nil {
			// Test the connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = client.Ping(ctx)
			cancel()

			if err == nil {
				break
			}
			client.Close()
		}

		if i < connectRetry-1 {
			time.Sleep(connectBackoff * time.Duration(i+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %v", err)
	}

	return client, nil

}
