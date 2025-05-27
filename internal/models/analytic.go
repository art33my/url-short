package models

import "time"

// AnalyticsResponse представляет статистику кликов для Swagger
// swagger:response AnalyticsResponse
type AnalyticsResponse struct {
	// Общее количество кликов
	// example: 42
	TotalClicks int `json:"total_clicks"`

	// Список кликов
	Clicks []ClickStatistic `json:"clicks"`
}

// ClickStatistic представляет данные одного клика для Swagger
// swagger:model ClickStatistic
type ClickStatistic struct {
	// IP-адрес клиента
	// example: 192.168.1.1
	IPAddress string `json:"ip_address"`

	// Геолокация клиента
	// example: Moscow, Russia
	Location string `json:"location"`

	// Тип устройства
	// example: mobile
	DeviceType string `json:"device_type"`

	// Операционная система
	// example: Android 13
	OS string `json:"os"`

	// Браузер пользователя
	// example: Chrome 115
	Browser string `json:"browser"`

	// Время клика
	// example: 2024-02-20T15:04:05Z
	ClickedAt time.Time `json:"clicked_at"`
}
