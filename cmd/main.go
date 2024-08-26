package main

import (
	"fmt"
	"ingress_backend/database"
	"ingress_backend/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := database.Connect()

	// Create a new router
	r := mux.NewRouter()

	r.HandleFunc("/student/{rollno}", routes.GetStudent(db)).Methods("GET")
	r.HandleFunc("/student", routes.AddStudent(db)).Methods("POST")
	r.HandleFunc("/logs", routes.GetLogs(db)).Methods("GET")
	r.HandleFunc("/statistics", routes.GetStatistics(db)).Methods("GET")

	log.Println("Server running on port 8000")
	// CORS middleware
	c := cors.New(cors.Options{
		// Allow all origins
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Insert the middleware
	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":8000", handler))
	fmt.Println("Server running on port 8000")
}
