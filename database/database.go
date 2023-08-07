package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func Connect() *sql.DB {
	once.Do(func() {
		var err error
		connStr := fmt.Sprintf("user=%s password=%s host=%s port=5432 dbname=ingress sslmode=disable",
			os.Getenv("POSTGRES_USERNAME"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"))
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Failed to connect to the database: %v", err)
		}

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
	})
	return db
}
