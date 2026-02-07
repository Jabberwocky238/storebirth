package dblayer

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

// DB connection
var DB *sql.DB

func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("Connection err: %s", err.Error())
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	return createTables()
}

// createTables creates all necessary database tables
func createTables() error {
	// Read SQL file
	sqlBytes, err := os.ReadFile("scripts/init.sql")
	if err != nil {
		return err
	}

	_, err = DB.Exec(string(sqlBytes))
	return err
}
