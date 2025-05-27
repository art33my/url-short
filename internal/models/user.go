// models/user.go
package models

import "time"

type User struct {
	ID           int       `json:"-"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}
