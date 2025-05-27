package models

import "time"

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url" example:"https://google.com"`
	CustomCode  string `json:"custom_code" example:"my_custom_code"`
}

type LinkResponse struct {
	ShortCode string `json:"short_code" example:"a1b2c3"`
	FullURL   string `json:"full_url" example:"http://localhost:8080/a1b2c3"`
}

type Link struct {
	ID          int       `json:"-"`
	UserID      int       `json:"-"`
	OriginalURL string    `json:"-"`
	ShortCode   string    `json:"-"`
	ClickCount  int       `json:"-"`
	CreatedAt   time.Time `json:"-"`
}
