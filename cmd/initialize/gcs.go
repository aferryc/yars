package initialize

import (
	"context"
	"errors"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aferryc/yars/internal/config"
	"google.golang.org/api/option"
)

const (
	maxGCSRetries = 5
	gcsRetryDelay = 2 * time.Second
)

func ConnectGCS(cfg *config.Config) (*storage.Client, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	clientOpts := []option.ClientOption{
		option.WithEndpoint(cfg.Bucket.URL),
		option.WithoutAuthentication(),
	}

	for i := 0; i < maxGCSRetries; i++ {
		client, err = storage.NewClient(ctx, clientOpts...)

		if err == nil {
			log.Println("Successfully connected to Google Cloud Storage")
			return client, nil
		}

		log.Printf("Failed to connect to GCS (attempt %d/%d): %v", i+1, maxGCSRetries, err)

		if i < maxGCSRetries-1 {
			log.Printf("Retrying in %v...", gcsRetryDelay)
			time.Sleep(gcsRetryDelay)
		}
	}

	return nil, errors.New("failed to connect to GCS")
}
