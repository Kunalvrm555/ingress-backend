package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Student struct {
	RollNo      string `json:"rollNo"`
	Name        string `json:"name"`
	CheckInTime string `json:"checkInTime"`
}

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:postgres@localhost/ingress?sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func GetStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rollNo := params["rollno"]

	fmt.Println(params["rollno"])

	time.Sleep(2 * time.Second)

	row := db.QueryRow("SELECT name FROM students WHERE rollno = $1", rollNo)

	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No matching student found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error.", http.StatusInternalServerError)
		return
	}

	currentTime := time.Now().Format("03:04 PM")

	student := Student{RollNo: rollNo, Name: name, CheckInTime: currentTime}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)

	fmt.Println(student.RollNo, student.Name, student.CheckInTime)
}
