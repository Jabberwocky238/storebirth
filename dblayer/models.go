package dblayer

import "time"

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

// VerificationCode model
type VerificationCode struct {
	ID        int       `json:"-"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// ConfigTask model
type ConfigTask struct {
	ID        int       `json:"-"`
	UserUID   string    `json:"user_uid"`
	TaskType  string    `json:"task_type"`
	Status    string    `json:"status"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
