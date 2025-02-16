package model

// AuthRequest содержит данные для аутентификации.
type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse содержит JWT-токен после успешной аутентификации.
type AuthResponse struct {
	Token string `json:"token"`
}
