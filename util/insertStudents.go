package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Open the CSV file
	f, err := os.Open("student_list.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic(err)
	}

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/ingress?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Prepare insert statement
	stmt, err := db.Prepare("INSERT INTO students(type,rollno,dept,name) VALUES($1, $2, $3, $4) ON CONFLICT (rollno) DO NOTHING")
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		_, err := stmt.Exec(line[0], line[1], line[2], line[3])
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Data successfully imported into PostgreSQL database!")
}
