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

// CustomDomain model
type CustomDomain struct {
	ID        string    `json:"id"`
	UserUID   string    `json:"user_uid"`
	Domain    string    `json:"domain"`
	Target    string    `json:"target"`
	TXTName   string    `json:"txt_name"`
	TXTValue  string    `json:"txt_value"`
	Status    string    `json:"status"` // pending, success, error
	CreatedAt time.Time `json:"created_at"`
}
