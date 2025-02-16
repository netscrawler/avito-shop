package model

// ErrorResponse используется для возврата сообщений об ошибках.
type ErrorResponse struct {
	Errors string `json:"errors"`
}
