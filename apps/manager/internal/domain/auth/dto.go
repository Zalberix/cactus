package auth

// LoginRequest — запрос на вход по email + пароль.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse — ответ с JWT access + refresh токенами.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshRequest — запрос на обновление access токена.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
