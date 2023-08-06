package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func JwtAuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the 'Authorization' header
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" {
			http.Error(w, "Missing auth token", http.StatusForbidden)
			return
		}

		// Split the token to retrieve just the token value without the "Bearer" prefix
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			http.Error(w, "Invalid/Malformed auth token", http.StatusForbidden)
			return
		}
		tokenPart := splitted[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenPart, func(token *jwt.Token) (interface{}, error) {
			// Ensure token method conforms to "SigningMethodHMAC"
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			http.Error(w, "Malformed authentication token", http.StatusForbidden)
			return
		}

		if !token.Valid {
			http.Error(w, "Token is not valid.", http.StatusForbidden)
			return
		}

		// Token is valid, call the next handler
		next.ServeHTTP(w, r)
	})
}
