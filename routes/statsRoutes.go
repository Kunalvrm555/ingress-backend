package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"ingress_backend/database"
	"net/http"
	"time"
)

type Statistics struct {
	CheckedInCount      int `json:"checkedInCount"`
	CheckedInTodayCount int `json:"checkedInTodayCount"`
}

func GetStatistics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stats Statistics
		today := time.Now()

		db := database.Connect()
		// Query to get the total number of students who have checked in today but haven't checked out
		err := db.QueryRow("SELECT COUNT(*) FROM logs WHERE checkouttime IS NULL AND DATE(checkintime) = $1", today).Scan(&stats.CheckedInCount)
		if err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		// Query to get total number of students who have checked in today
		err = db.QueryRow("SELECT COUNT(*) FROM logs WHERE DATE(checkintime) = $1", today).Scan(&stats.CheckedInTodayCount)
		if err != nil {
			http.Error(w, "Server error.", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		// Return the results as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
