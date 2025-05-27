package models

// ErrorResponse: Универсальный ответ для ошибок
// swagger:response ErrorResponse
type ErrorResponse struct {
	// Описание ошибки
	// example: Неверные данные
	Error string `json:"error"`
}
