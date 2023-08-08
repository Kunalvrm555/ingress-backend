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

// Define the Student struct.
type Student struct {
	RollNo       string `json:"rollNo"`
	Name         string `json:"name"`
	CheckInTime  string `json:"checkInTime"`
	CheckoutTime string `json:"checkoutTime"`
}

// GetStudent handles the GET request for retrieving student information.
func GetStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		rollNo := params["rollno"]

		var name string
		err := db.QueryRow("SELECT name FROM students WHERE rollno = $1", rollNo).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Student not found.", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to retrieve the student name.", http.StatusInternalServerError)
			}
			return
		}

		// Begin a transaction.
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction.", http.StatusInternalServerError)
			return
		}

		// Set the current time.
		currentTime := time.Now()
		// Fetch the log for the current date.
		row := tx.QueryRow("SELECT checkintime, checkouttime FROM logs WHERE rollno = $1 AND DATE(checkintime) = CURRENT_DATE", rollNo)
		var checkInTime, checkoutTime sql.NullString
		err = row.Scan(&checkInTime, &checkoutTime)

		if err == sql.ErrNoRows {
			// Insert a new row if the entry does not exist.
			_, err := tx.Exec("INSERT INTO logs (rollno, name, checkintime) VALUES ($1, $2, $3)", rollNo, name, currentTime)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to insert log.", http.StatusInternalServerError)
				return
			}
			checkInTime.String = currentTime.Format(time.RFC3339Nano)
		} else if err == nil && checkoutTime.Valid {
			// Update checkout time to null and checkin time to current time.
			_, err := tx.Exec("UPDATE logs SET checkouttime = NULL, checkintime = $1 WHERE rollno = $2 AND DATE(checkintime) = CURRENT_DATE", currentTime, rollNo)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to update log.", http.StatusInternalServerError)
				return
			}
			checkInTime.String = currentTime.Format(time.RFC3339Nano)
			checkoutTime.Valid = false
		} else if err == nil && !checkoutTime.Valid {
			// Update checkout time to current time.
			_, err := tx.Exec("UPDATE logs SET checkouttime = $1 WHERE rollno = $2 AND DATE(checkintime) = CURRENT_DATE", currentTime, rollNo)
			if err != nil {
				tx.Rollback()
				http.Error(w, "Failed to update log.", http.StatusInternalServerError)
				return
			}
			checkoutTime.String = currentTime.Format(time.RFC3339Nano)
			checkoutTime.Valid = true
		} else {
			tx.Rollback()
			http.Error(w, "Failed to retrieve log.", http.StatusInternalServerError)
			return
		}

		// Commit the transaction.
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to commit transaction.", http.StatusInternalServerError)
			return
		}

		// Format checkin and checkout times.
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
		}

		// Create the student object.
		student := Student{
			RollNo:       rollNo,
			Name:         name,
			CheckInTime:  checkInTimeStr.Format("03:04 PM"),
			CheckoutTime: checkoutTimeStr,
		}

		// Set the response content type to JSON.
		w.Header().Set("Content-Type", "application/json")

		// Encode the student object and send it in the response.
		err = json.NewEncoder(w).Encode(student)
		if err != nil {
			http.Error(w, "Failed to encode the student object.", http.StatusInternalServerError)
			return
		}
	}
}

// ParseDatabaseTime parses the time string stored in the database (RFC3339Nano format)
// and returns it as a time.Time object.
func parseDatabaseTime(timeStr string) (time.Time, error) {
	// The format used in the database for the time is RFC3339Nano.
	return time.Parse(time.RFC3339Nano, timeStr)
}

// StudentInput defines the structure of the incoming JSON data.
type StudentInput struct {
	Type   string `json:"type"`
	RollNo string `json:"rollno"`
	Dept   string `json:"dept"`
	Name   string `json:"name"`
}
type RowError struct {
	RowIndex int    `json:"rowIndex"`
	Message  string `json:"message"`
}

func AddStudents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var students []StudentInput
		if err := json.NewDecoder(r.Body).Decode(&students); err != nil {
			http.Error(w, "Failed to decode request body.", http.StatusBadRequest)
			return
		}

		var rowErrors []RowError
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction.", http.StatusInternalServerError)
			return
		}

		for i, student := range students {
			// Validate student data.
			if student.RollNo == "" || student.Name == "" || student.Type == "" || student.Dept == "" {
				rowErrors = append(rowErrors, RowError{RowIndex: i, Message: "Incomplete data in row."})
				continue
			}

			result, err := tx.Exec(`INSERT INTO students (rollno, name, type, dept) VALUES ($1, $2, $3, $4)
				ON CONFLICT (rollno) DO NOTHING;`,
				student.RollNo, student.Name, student.Type, student.Dept)
			fmt.Println(result)
			if err != nil {
				rowErrors = append(rowErrors, RowError{RowIndex: i, Message: err.Error()})
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction.", http.StatusInternalServerError)
			return
		}

		if len(rowErrors) > 0 {
			w.WriteHeader(http.StatusPartialContent)
			json.NewEncoder(w).Encode(rowErrors)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "All students added successfully."})
	}
}
