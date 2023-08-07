package util

import (
	"fmt"
	"log"
	"os"

	"ingress_backend/database"

	"golang.org/x/crypto/bcrypt"
)

func SeedUsers() {
	// Use the shared database connection from the database package
	db := database.Connect()

	adminPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error generating password hash for admin:", err)
	}

	result, err := db.Exec("INSERT INTO users (username, password, usertype) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING", os.Getenv("ADMIN_USERNAME"), string(adminPassword), "admin")
	if err != nil {
		log.Fatal("Error inserting admin:", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("Error checking affected rows for admin:", err)
	}
	if rowsAffected > 0 {
		fmt.Println("Admin seed successful.")
	}

	userPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error generating password hash for user:", err)
	}

	result, err = db.Exec("INSERT INTO users (username, password, usertype) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING", os.Getenv("USERNAME"), string(userPassword), "user")
	if err != nil {
		log.Fatal("Error inserting user:", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		log.Fatal("Error checking affected rows for user:", err)
	}
	if rowsAffected > 0 {
		fmt.Println("User seed successful.")
	}
}
