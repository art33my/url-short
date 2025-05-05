package models

import "time"

type ClickAnalytic struct {
	ID         int       `json:"id"`
	LinkID     int       `json:"link_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Location   string    `json:"location"`
	DeviceType string    `json:"device_type"`
	OS         string    `json:"os"`
	Browser    string    `json:"browser"`
	ClickedAt  time.Time `json:"clicked_at"`
}
