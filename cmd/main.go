package main

import (
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := mux.NewRouter()

	r.HandleFunc("/student/{rollno}", middleware.JwtAuthenticationMiddleware(routes.GetStudent)).Methods("GET")
	r.HandleFunc("/logs", middleware.JwtAuthenticationMiddleware(routes.GetLogs)).Methods("GET")
	r.HandleFunc("/statistics", middleware.JwtAuthenticationMiddleware(routes.GetStatistics)).Methods("GET")
	r.HandleFunc("/login", routes.Login).Methods("POST")

	// Seed users
	util.SeedUsers()

	log.Println("Server running on port 8000")
	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Insert the middleware
	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe("127.0.0.1:8000", handler))
}
