package dblayer

import (
	"database/sql"
	"errors"
	"fmt"

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

	return DB.Ping()
}
