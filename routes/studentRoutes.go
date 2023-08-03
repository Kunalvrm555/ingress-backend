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

// GetStudent handles the GET request for retrieving student information
func GetStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rollNo := params["rollno"]

	fmt.Println(params["rollno"])

	// Introducing a small delay to simulate some processing time
	time.Sleep(2 * time.Second)

	// Begin a transaction to update the checkout time for the student
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "1")
		return
	}

	// Record the current time as the checkout time for the student
	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")
	_, err = tx.Exec("UPDATE logs SET checkouttime = $1 WHERE rollno = $2 AND DATE(checkintime) = DATE($1)", currentTime, rollNo)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "2")
		return
	}

	// Retrieve the student's name, department, and type from the database
	row := db.QueryRow("SELECT name, dept, type FROM students WHERE rollno = $1", rollNo)
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

	// Insert a new log entry for the student if it doesn't exist for the current date
	_, err = tx.Exec("INSERT INTO logs (rollno, name, dept, type, checkintime) SELECT $1, $2, $3, $4, $5 WHERE NOT EXISTS (SELECT 1 FROM logs WHERE rollno = $1 AND DATE(checkintime) = $6)", rollNo, name, dept, type_, currentTime, currentDate)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "3")
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "4")
		return
	}

	// Retrieve the check-in time and checkout time for the student
	row = db.QueryRow("SELECT checkintime, checkouttime FROM logs WHERE rollno = $1 AND DATE(checkintime) = DATE($2)", rollNo, currentTime)
	var checkInTime, checkoutTime sql.NullString
	err = row.Scan(&checkInTime, &checkoutTime)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "5")
		return
	}

	// Format the check-in time as "03:04 PM" format
	parsedCheckInTime, err := parseDatabaseTime(checkInTime.String)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "6")
		return
	}
	checkInTimeFormatted := parsedCheckInTime.Format("03:04 PM")

	// Format the checkout time as "03:04 PM" format
	var checkoutTimeStr string
	if checkoutTime.Valid {
		parsedCheckoutTime, err := parseDatabaseTime(checkoutTime.String)
		if err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			fmt.Println(err, "7")
			return
		}
		checkoutTimeStr = parsedCheckoutTime.Format("03:04 PM")
	}

	// Create the student object to be returned in the response
	student := Student{
		RollNo:       rollNo,
		Name:         name,
		CheckInTime:  checkInTimeFormatted,
		CheckoutTime: checkoutTimeStr,
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the student object and send it as the response
	err = json.NewEncoder(w).Encode(student)
	if err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		fmt.Println(err, "8")
		return
	}

	// Log the student information
	fmt.Println(student.RollNo, student.Name, student.CheckInTime, student.CheckoutTime)
}

// parseDatabaseTime parses the time string stored in the database (RFC3339Nano format)
// and returns it as a time.Time object.
func parseDatabaseTime(timeStr string) (time.Time, error) {
	// The format used in the database for the time is RFC3339Nano
	// e.g., "2023-08-03T21:20:56.051486Z"
	return time.Parse(time.RFC3339Nano, timeStr)
}
