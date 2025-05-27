// models/auth.go
package models

type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"qwerty123"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"qwerty123"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGci..."`
}

type RegisterResponse struct {
	Message string `json:"message" example:"Пользователь создан"`
}
