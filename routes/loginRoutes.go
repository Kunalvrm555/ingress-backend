package routes

import (
	"encoding/json"
	"fmt"
	"ingress_backend/database"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserType string `json:"userType"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	db := database.Connect()

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		fmt.Println(err, 1)
		return
	}

	row := db.QueryRow("SELECT password, userType FROM users WHERE username = $1", user.Username)
	var hashedPassword string
	var userType string
	err = row.Scan(&hashedPassword, &userType)
	if err != nil {
		http.Error(w, "Invalid username or password.", http.StatusBadRequest)
		fmt.Println(err, 2)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid username or password.", http.StatusBadRequest)
		fmt.Println(err, 3)
		return
	}

	token, err := createToken(user.Username, userType)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, 4)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func createToken(username string, userType string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"userType": userType,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}
