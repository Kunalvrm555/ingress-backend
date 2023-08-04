package util

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers() {
	db, err := sql.Open("postgres", "user="+os.Getenv("POSTGRES_USERNAME")+" password="+os.Getenv("POSTGRES_PASSWORD")+" dbname=ingress sslmode=disable")
	if err != nil {
		log.Fatal(err, 1)
	}
	defer db.Close()

	adminPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err, 2)
	}

	_, err = db.Exec("INSERT INTO users (username, password, usertype) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING", os.Getenv("ADMIN_USERNAME"), string(adminPassword), "admin")
	if err != nil {
		log.Fatal(err, 3)
	}

	userPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err, 4)
	}

	_, err = db.Exec("INSERT INTO users (username, password, usertype) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING", os.Getenv("USERNAME"), string(userPassword), "user")
	if err != nil {
		log.Fatal(err, 5)
	}
	fmt.Println("User seed successful.")
}
