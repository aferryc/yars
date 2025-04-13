package initialize

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const maxRetries = 5
const retryDelay = 2 * time.Second

// ConnectDBWithRetry establishes a connection to the database with retry logic
func ConnectDB(databaseURL string) *sqlx.DB {
	var db *sqlx.DB
	var err error

	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", databaseURL)
		if err == nil {
			break
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatalf("Could not connect to database after %d attempts: %v", maxRetries, err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	log.Println("Successfully connected to the database")
	return db
}
