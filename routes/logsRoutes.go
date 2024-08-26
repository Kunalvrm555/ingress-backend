package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"ingress_backend/database"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type StudentLog struct {
	RollNo      string `json:"rollno"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Department  string `json:"department"`
	CheckInTime string `json:"checkInTime"`
}

func GetLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := database.Connect()
		today := time.Now().Format("2006-01-02")
		rows, err := db.Query(`SELECT l.rollno, s.name, s.type, s.dept, l.checkintime 
                            FROM logs l 
                            LEFT JOIN students s 
                            ON l.rollno = s.rollno 
							WHERE l.checkouttime IS NULL AND DATE(l.checkintime) = $1
							ORDER BY l.checkintime DESC`, today)
		if err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			fmt.Println(err, 1)
			return
		}
		defer rows.Close()

		var studentLogs []StudentLog
		for rows.Next() {
			var log StudentLog
			var checkInTimeDB string
			err = rows.Scan(&log.RollNo, &log.Name, &log.Type, &log.Department, &checkInTimeDB)
			if err != nil {
				http.Error(w, "Server error.", http.StatusInternalServerError)
				fmt.Println(err, 2)
				return
			}

			checkInTime, err := time.Parse(time.RFC3339, checkInTimeDB)
			if err != nil {
				http.Error(w, "Server error.", http.StatusInternalServerError)
				fmt.Println(err, 3)
				return
			}

			log.CheckInTime = checkInTime.Format("03:04 PM")

			studentLogs = append(studentLogs, log)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(studentLogs)
	}
}
