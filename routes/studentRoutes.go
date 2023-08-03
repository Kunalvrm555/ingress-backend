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
	RollNo       string `json:"rollNo"`
	Name         string `json:"name"`
	CheckInTime  string `json:"checkInTime"`
	CheckoutTime string `json:"checkoutTime"`
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

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "1")
		return
	}

	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")
	_, err = tx.Exec("UPDATE logs SET checkouttime = $1 WHERE rollno = $2 AND DATE(checkintime) = DATE($1)", currentTime, rollNo)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "2")
		return
	}

	row := db.QueryRow("SELECT name, dept, type FROM students WHERE rollno = $1", rollNo)

	//type is a keyword in Go, so we use type_ instead
	var name, dept, type_ string
	err = row.Scan(&name, &dept, &type_)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No matching student found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error.", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO logs (rollno, name, dept, type, checkintime) SELECT $1, $2, $3, $4, $5 WHERE NOT EXISTS (SELECT 1 FROM logs WHERE rollno = $1 AND DATE(checkintime) = $6)", rollNo, name, dept, type_, currentTime, currentDate)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "3")
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "4")
		return
	}

	row = db.QueryRow("SELECT checkintime, checkouttime FROM logs WHERE rollno = $1 AND DATE(checkintime) = DATE($2)", rollNo, currentTime)

	// Use sql.NullString to handle NULL values for checkouttime
	var checkInTime, checkoutTime sql.NullString
	err = row.Scan(&checkInTime, &checkoutTime)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "5")
		return
	}

	// If checkoutTime is NULL, set an empty string for the response
	var checkoutTimeStr string
	if checkoutTime.Valid {
		checkoutTimeStr = checkoutTime.String

		// Format the checkoutTime as "03:04 PM" format
		parsedTime, err := time.Parse(time.RFC3339Nano, checkoutTimeStr)
		if err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			fmt.Println(err, "6")
			return
		}
		checkoutTimeStr = parsedTime.Format("03:04 PM")
	}

	// Format the checkInTime as "03:04 PM" format
	parsedTime, err := time.Parse(time.RFC3339Nano, checkInTime.String)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "7")
		return
	}
	checkInTimeFormatted := parsedTime.Format("03:04 PM")

	student := Student{
		RollNo:       rollNo,
		Name:         name,
		CheckInTime:  checkInTimeFormatted,
		CheckoutTime: checkoutTimeStr,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)

	fmt.Println(student.RollNo, student.Name, student.CheckInTime, student.CheckoutTime)
}
