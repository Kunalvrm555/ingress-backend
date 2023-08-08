package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

const maxRetries = 5

func connectAndPing() (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=5432 dbname=ingress sslmode=disable",
		os.Getenv("POSTGRES_USERNAME"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"))

	tmpDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the database: %v", err)
	}

	for i := 0; i < maxRetries; i++ {
		err = tmpDB.Ping()
		if err == nil {
			return tmpDB, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d). Retrying in 2 seconds...\n", i+1, maxRetries)
		time.Sleep(time.Second * 2)
	}

	tmpDB.Close()
	return nil, fmt.Errorf("Failed to connect to database after %d tries: %v", maxRetries, err)
}

func Connect() *sql.DB {
	once.Do(func() {
		var err error
		db, err = connectAndPing()
		if err != nil {
			log.Fatalf("Connection failure after retries: %v", err)
		}
	})
	return db
}
