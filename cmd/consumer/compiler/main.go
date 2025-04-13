package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aferryc/yars/cmd/initialize"
	"github.com/aferryc/yars/internal/config"
	"github.com/aferryc/yars/repository/gcs"
	"github.com/aferryc/yars/repository/kafka"
	"github.com/aferryc/yars/repository/postgres"
	"github.com/aferryc/yars/transport"
	"github.com/aferryc/yars/usecase"
)

func main() {
	cfg := config.LoadConfig()

	gcsClient, err := initialize.ConnectGCS(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to GCS: %v", err)
	}

	gcsRepo, err := gcs.NewGCSRepository(cfg.Bucket.Name, gcsClient)
	if err != nil {
		log.Fatalf("Failed to create GCS repository: %v", err)
	}

	pgConn := initialize.ConnectDB(cfg.DatabaseURL)
	if pgConn == nil {
		log.Fatal("Failed to connect to PostgreSQL")
	}

	bankRepo := postgres.NewDBBankStatementRepository(pgConn)
	transactionRepo := postgres.NewDBInternalTransactionRepository(pgConn)

	kafkaConn, err := initialize.NewKafkaProducer(cfg.Kafka.BrokerList, cfg.Kafka.ClientID)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	kafkaRepo := kafka.NewKafkaRepository(kafkaConn)

	uc := usecase.NewFileCompiler(cfg, gcsRepo, bankRepo, transactionRepo, kafkaRepo)
	consumer, err := transport.NewConsumer(&cfg.Kafka, cfg.Kafka.Topic.CompilerTopic, uc)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start the consumer in a goroutine
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
			cancel() // Cancel context to signal shutdown
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)
	cancel() // Cancel context to signal shutdown

	// Allow some time for cleanup
	shutdownTimeout := 10 * time.Second
	time.Sleep(shutdownTimeout)
}
