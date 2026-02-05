package dblayer

import "time"

// User model
type User struct {
	ID           int       `json:"-"`
	UID          string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	SecretKey    string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// VerificationCode model
type VerificationCode struct {
	ID        int       `json:"-"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}
