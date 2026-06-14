package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	db "github.com/zalberix/cactus/libs/storage/db"
)

const (
	headerPublicToken  = "X-Public-Token"
	headerPrivateToken = "X-Private-Token"
	systemTokenKey     = "system_token"
)

// SystemTokenStore — интерфейс для поиска системного токена по публичному токену.
type SystemTokenStore interface {
	GetSystemTokenByPublicToken(ctx context.Context, publicToken string) (db.SystemToken, error)
}

// SystemTokenAuth — middleware аутентификации M2M по X-Public-Token + X-Private-Token.
// Алгоритм:
//  1. Читаем заголовки X-Public-Token и X-Private-Token
//  2. Ищем system_token по public_token (WHERE is_active=TRUE)
//  3. Хэшируем incoming private token: sha256hex(incoming)
//  4. Сравниваем с сохранённым хэшем через constant-time comparison (crypto/subtle)
//  5. Устанавливаем system_token_id и system_id в gin.Context
func SystemTokenAuth(store SystemTokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		publicToken := c.GetHeader(headerPublicToken)
		privateToken := c.GetHeader(headerPrivateToken)

		if publicToken == "" || privateToken == "" {
			response.Unauthorized(c, "Требуются заголовки X-Public-Token и X-Private-Token")
			c.Abort()
			return
		}

		token, err := store.GetSystemTokenByPublicToken(c.Request.Context(), publicToken)
		if err != nil {
			response.Unauthorized(c, "Недействительный токен")
			c.Abort()
			return
		}

		// Хэшируем входящий приватный токен
		incomingHash := sha256hex(privateToken)

		// Constant-time сравнение хэшей (защита от timing attacks)
		if subtle.ConstantTimeCompare([]byte(token.PrivateToken), []byte(incomingHash)) != 1 {
			response.Unauthorized(c, "Недействительный приватный токен")
			c.Abort()
			return
		}

		// Устанавливаем данные токена в контекст
		c.Set(systemTokenKey, token)
		c.Next()
	}
}

// GetSystemToken извлекает системный токен из gin.Context.
func GetSystemToken(c *gin.Context) *db.SystemToken {
	v, exists := c.Get(systemTokenKey)
	if !exists {
		return nil
	}
	token, ok := v.(db.SystemToken)
	if !ok {
		return nil
	}
	return &token
}

// sha256hex вычисляет SHA256 хэш и возвращает hex строку.
func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
