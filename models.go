package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// User model
type User struct {
	ID           int       `json:"-"`
	UID          string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// RDB model
type RDB struct {
	ID      int    `json:"-"`
	UID     string `json:"id"`
	UserID  int    `json:"-"`
	Name    string `json:"name"`
	Type    string `json:"rdb_type"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// KV model
type KV struct {
	ID      int    `json:"-"`
	UID     string `json:"id"`
	UserID  int    `json:"-"`
	Name    string `json:"name"`
	Type    string `json:"kv_type"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// DB connection
var DB *sql.DB

func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
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
