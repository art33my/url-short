package models

import "time"

type Link struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ClickCount  int       `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}
