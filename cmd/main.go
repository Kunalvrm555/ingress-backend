package main

import (
	"fmt"
	"ingress_backend/database"
	"ingress_backend/middleware"
	"ingress_backend/routes"
	"ingress_backend/util"
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

	r.HandleFunc("/student/{rollno}", middleware.JwtAuthenticationMiddleware(routes.GetStudent(db))).Methods("GET")
	r.HandleFunc("/logs", middleware.JwtAuthenticationMiddleware(routes.GetLogs(db))).Methods("GET")
	r.HandleFunc("/statistics", middleware.JwtAuthenticationMiddleware(routes.GetStatistics(db))).Methods("GET")
	r.HandleFunc("/login", routes.Login).Methods("POST")
	r.HandleFunc("/student/add", middleware.JwtAuthenticationMiddleware(routes.AddStudents(db))).Methods("POST")

	// Seed users
	util.SeedUsers()

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
