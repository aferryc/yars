package main

import (
	"log"
	"net/http"
	"os"
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
	// Set up logging with timestamp
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Server starting...")

	// Log environment for debugging
	log.Println("Environment variables:")
	log.Printf("  PORT=%s", os.Getenv("PORT"))
	log.Printf("  DATABASE_URL=%s", maskPassword(os.Getenv("DATABASE_URL")))
	log.Printf("  BUCKET_NAME=%s", os.Getenv("BUCKET_NAME"))
	log.Printf("  BUCKET_URL=%s", os.Getenv("BUCKET_URL"))
	log.Printf("  KAFKA_BROKERS=%s", os.Getenv("KAFKA_BROKERS"))
	log.Printf("  STORAGE_EMULATOR_HOST=%s", os.Getenv("STORAGE_EMULATOR_HOST"))

	// Load application configuration
	log.Println("Loading configuration...")
	cfg := config.LoadConfig()
	log.Printf("Configuration loaded: Port=%s, BucketName=%s, KafkaClientID=%s",
		cfg.Port, cfg.Bucket.Name, cfg.Kafka.ClientID)

	// Connect to GCS
	log.Println("Connecting to GCS storage...")
	gcsClient, err := initialize.ConnectGCS(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to GCS: %v", err)
	}
	log.Println("Successfully connected to GCS")

	// Initialize GCS repository
	log.Printf("Initializing GCS repository with bucket: %s", cfg.Bucket.Name)
	gcsRepo, err := gcs.NewGCSRepository(cfg.Bucket.Name, gcsClient)
	if err != nil {
		log.Fatalf("Failed to create GCS repository: %v", err)
	}
	log.Println("GCS repository initialized")

	// Connect to Kafka
	log.Printf("Connecting to Kafka at %s with client ID %s...", cfg.Kafka.BrokerList, cfg.Kafka.ClientID)
	kafkaConn, err := initialize.NewKafkaProducer(cfg.Kafka.BrokerList, cfg.Kafka.ClientID)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	log.Println("Successfully connected to Kafka")

	// Initialize Kafka repository
	log.Println("Initializing Kafka repository...")
	kafkaRepo := kafka.NewKafkaRepository(kafkaConn)
	log.Println("Kafka repository initialized")

	// Connect to database
	log.Printf("Connecting to database at %s...", maskPassword(cfg.DatabaseURL))
	dbConn := initialize.ConnectDB(cfg.DatabaseURL)
	log.Println("Successfully connected to database")

	// Perform DB ping to verify connection
	log.Println("Testing database connection...")
	err = dbConn.Ping()
	if err != nil {
		log.Fatalf("Database connection check failed: %v", err)
	}
	log.Println("Database connection verified")

	// Initialize repository and use cases
	log.Println("Initializing repositories and use cases...")
	listRepo := postgres.NewDBReconResultRepository(dbConn)
	reconUC := usecase.NewReconManager(gcsRepo, kafkaRepo, cfg)
	listUC := usecase.NewListUsecase(listRepo)

	// Set up the router
	log.Println("Setting up HTTP router...")
	handler := transport.NewHandler(reconUC, listUC)
	router := initialize.SetupRouter(*handler)
	log.Println("Router setup complete")

	// Start the HTTP server
	log.Printf("Starting server on port %s...", cfg.Port)
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Log ready message just before starting to listen
	log.Printf("Server ready to accept connections on port %s", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

// maskPassword hides the password in connection strings for safe logging
func maskPassword(connString string) string {
	// Simple password masking for common connection string formats
	// This won't catch all cases but works for most standard formats
	if connString == "" {
		return ""
	}

	// For postgres://user:password@host:port/dbname format
	passwordStart := -1
	passwordEnd := -1

	for i := 0; i < len(connString); i++ {
		if connString[i] == ':' && passwordStart == -1 {
			// Look for the pattern "://" to avoid matching the port colon
			if i+2 < len(connString) && connString[i:i+3] != "://" {
				passwordStart = i + 1
			}
		} else if connString[i] == '@' && passwordStart != -1 && passwordEnd == -1 {
			passwordEnd = i
		}
	}

	if passwordStart != -1 && passwordEnd != -1 {
		return connString[:passwordStart] + "******" + connString[passwordEnd:]
	}

	return connString
}
