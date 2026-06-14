package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/domain/auth"
	"github.com/zalberix/cactus/apps/manager/internal/http/response"
)

// Auth — JWT Bearer token validation middleware.
// Извлекает токен из заголовка Authorization, валидирует и кладёт Claims в контекст.
func Auth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "Отсутствует токен авторизации")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "Неверный формат токена авторизации")
			c.Abort()
			return
		}

		claims, err := authService.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Неверный токен")
			c.Abort()
			return
		}

		SetClaims(c, claims)
		c.Next()
	}
}
