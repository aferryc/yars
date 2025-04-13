package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App            AppConfig
	Port           string
	DatabaseURL    string
	BankAPIBaseURL string
	Bucket         BucketConfig
	Kafka          KafkaConfig
}

type AppConfig struct {
	Port     string
	Compiler CompilerConfig
	Server   ServerConfig
}

type ServerConfig struct {
	Address string
}

type CompilerConfig struct {
	BatchSize int
}

type KafkaConfig struct {
	BrokerList []string
	Topic      TopicConfig
	GroupID    string
	ClientID   string
}

type TopicConfig struct {
	CompilerTopic string
	ReconTopic    string
}

type BucketConfig struct {
	Name string
	URL  string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	batchSize, err := strconv.Atoi(getEnv("COMPILER_BATCH_SIZE", "100"))
	if err != nil {
		batchSize = 100
	}

	// Create full config
	config := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname"),
		BankAPIBaseURL: getEnv("BANK_API_BASE_URL", "https://api.bank.com"),
		App: AppConfig{
			Port: getEnv("PORT", "8080"),
			Compiler: CompilerConfig{
				BatchSize: batchSize,
			},
			Server: ServerConfig{
				Address: getEnv("SERVER_ADDRESS", ":8080"),
			},
		},
		Bucket: BucketConfig{
			Name: getEnv("BUCKET_NAME", "default-bucket"),
			URL:  getEnv("BUCKET_URL", "https://storage.googleapis.com"),
		},
		Kafka: KafkaConfig{
			BrokerList: kafkaBrokers,
			GroupID:    getEnv("KAFKA_GROUP_ID", "yars-group"),
			ClientID:   getEnv("KAFKA_CLIENT_ID", "yars-client"),
			Topic: TopicConfig{
				CompilerTopic: getEnv("KAFKA_COMPILER_TOPIC", "compiler-events"),
				ReconTopic:    getEnv("KAFKA_RECON_TOPIC", "reconciliation-events"),
			},
		},
	}

	return config, nil
}

// LoadConfig loads the configuration and panics on error (maintained for backward compatibility)
func LoadConfig() *Config {
	config, err := Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
