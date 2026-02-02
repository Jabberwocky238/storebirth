package storebirth

import (
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
