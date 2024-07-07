package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB_USER=your_db_user
// DB_PASSWORD=your_db_password
// DB_HOST=localhost
// DB_PORT=5432
// DB_NAME=hotel_locator

func ConnectDB() (*sql.DB, error) {
	connStr := "user=your_db_user dbname=hotel_locator sslmode=disable password=your_db_password"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PgSQL db")

	return db, nil
}
