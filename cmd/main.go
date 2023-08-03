package main

import (
	"ingress_backend/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/student/{rollno}", routes.GetStudent).Methods("GET")
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
