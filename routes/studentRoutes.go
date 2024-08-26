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

// Student struct defines the data model for student records.
type Student struct {
	RollNo       string `json:"rollNo"`
	Name         string `json:"name"`
	CheckInTime  string `json:"checkInTime"`
	CheckoutTime string `json:"checkoutTime"`
}

// GetStudent returns an HTTP handler function that manages student check-in and check-out logs.
func GetStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the roll number from URL parameters.
		params := mux.Vars(r)
		rollNo := params["rollno"]

		// Retrieve the student's name from the database using the roll number.
		var name string
		err := db.QueryRow("SELECT name FROM students WHERE rollno = $1", rollNo).Scan(&name)
		if err != nil {
			// Handle errors such as no student found or database issues.
			if err == sql.ErrNoRows {
				http.Error(w, "Student not found.", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to retrieve the student name.", http.StatusInternalServerError)
			}
			return
		}

		// Start a transaction for managing log entries.
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction.", http.StatusInternalServerError)
			return
		}

		// Current time and date used for logs.
		currentTime := time.Now()
		currentDate := currentTime.Format("2006-01-02")

		// Query for existing log entries for the student on the current date.
		logQuery := fmt.Sprintf("SELECT checkintime, checkouttime FROM logs WHERE rollno = '%s' AND DATE(checkintime) = '%s'", rollNo, currentDate)
		row := tx.QueryRow(logQuery)

		// Check if the student has a log entry for the current date
		var checkInTime, checkoutTime sql.NullString
		err = row.Scan(&checkInTime, &checkoutTime)
		// Determine the necessary action based on log entry existence and checkout time status.
		if err == sql.ErrNoRows {
			fmt.Println("Case 1")
			// Case 1: No entry exists, insert new log with check-in time.
			insertQuery := fmt.Sprintf("INSERT INTO logs (rollno, name, checkintime) VALUES ('%s', '%s', '%s')", rollNo, name, currentTime.Format(time.RFC3339Nano))
			_, err = tx.Exec(insertQuery)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to insert log.", http.StatusInternalServerError)
				return
			}
			checkInTime.String = currentTime.Format(time.RFC3339Nano)
		} else if err == nil && !checkoutTime.Valid {
			// Case 2: Entry exists, update checkout time if it's null.
			fmt.Println("Case 2")
			updateCheckoutQuery := fmt.Sprintf("UPDATE logs SET checkouttime = '%s' WHERE rollno = '%s' AND DATE(checkintime) = '%s' AND checkouttime IS NULL", currentTime.Format(time.RFC3339Nano), rollNo, currentDate)
			_, err = tx.Exec(updateCheckoutQuery)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to update checkout time.", http.StatusInternalServerError)
				return
			}
			checkoutTime.String = currentTime.Format(time.RFC3339Nano)
			checkoutTime.Valid = true
		} else if err == nil && checkoutTime.Valid {
			fmt.Println("Case 3")
			// Case 3: Entry exists and checkout time is not null, reset checkout time to null and update check-in time.
			resetCheckoutQuery := fmt.Sprintf("UPDATE logs SET checkouttime = NULL, checkintime = '%s' WHERE rollno = '%s' AND DATE(checkintime) = '%s'", currentTime.Format(time.RFC3339Nano), rollNo, currentDate)
			_, err = tx.Exec(resetCheckoutQuery)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to reset and update log.", http.StatusInternalServerError)
				return
			}
			checkInTime.String = currentTime.Format(time.RFC3339Nano)
			checkoutTime.Valid = false
		} else {
			// Handle unexpected database errors.
			tx.Rollback()
			http.Error(w, "Failed to retrieve log.", http.StatusInternalServerError)
			return
		}

		// Commit the transaction to ensure all changes are saved to the database.
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to commit transaction.", http.StatusInternalServerError)
			return
		}

		// Parse and format the check-in and check-out times for the response.
		checkInTimeStr, err := parseDatabaseTime(checkInTime.String)
		if err != nil {
			http.Error(w, "Failed to parse check-in time.", http.StatusInternalServerError)
			return
		}
		var checkoutTimeStr string
		if checkoutTime.Valid {
			parsedCheckoutTime, err := parseDatabaseTime(checkoutTime.String)
			if err != nil {
				http.Error(w, "Failed to parse check-out time.", http.StatusInternalServerError)
				return
			}
			checkoutTimeStr = parsedCheckoutTime.Format("03:04 PM")
		} else {
			checkoutTimeStr = ""
		}

		// Create the student object with formatted times for JSON response.
		student := Student{
			RollNo:       rollNo,
			Name:         name,
			CheckInTime:  checkInTimeStr.Format("03:04 PM"),
			CheckoutTime: checkoutTimeStr,
		}

		// Set response header to JSON and encode the student object into the response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(student)
		if err != nil {
			http.Error(w, "Failed to encode the student object.", http.StatusInternalServerError)
			return
		}
	}
}

// ParseDatabaseTime parses the RFC3339Nano formatted time string from the database and returns a time.Time object.
func parseDatabaseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, timeStr)
}

// StudentInput defines the structure of the incoming JSON data for adding a new student.
type StudentInput struct {
	Type   string `json:"type"`
	RollNo string `json:"rollno"`
	Dept   string `json:"dept"`
	Name   string `json:"name"`
}

// AddStudent returns an HTTP handler function that adds a new student record to the database.
func AddStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var student StudentInput
		// Decode the incoming JSON data into the StudentInput struct.
		if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
			http.Error(w, "Failed to decode request body.", http.StatusBadRequest)
			return
		}

		// Validate that the necessary fields are provided.
		if student.RollNo == "" || student.Name == "" || student.Type == "" || student.Dept == "" {
			http.Error(w, "Incomplete data provided.", http.StatusBadRequest)
			return
		}

		// Begin a transaction for adding the student.
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction.", http.StatusInternalServerError)
			return
		}

		// Execute the SQL command to insert a new student record.
		_, err = tx.Exec(`INSERT INTO students (rollno, name, type, dept) VALUES ($1, $2, $3, $4)
                ON CONFLICT (rollno) DO NOTHING;`, student.RollNo, student.Name, student.Type, student.Dept)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert student.", http.StatusInternalServerError)
			return
		}

		// Commit the transaction if all operations are successful.
		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction.", http.StatusInternalServerError)
			return
		}

		// Set the HTTP status to 201 (Created) and encode the added student's data in the response.
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(student)
	}
}
