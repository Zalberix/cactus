package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/domain/auth"
)

const claimsKey = "claims"

// SetClaims сохраняет JWT claims в gin.Context.
func SetClaims(c *gin.Context, claims *auth.Claims) {
	c.Set(claimsKey, claims)
}

// GetClaims извлекает JWT claims из gin.Context.
// Возвращает nil, если claims отсутствуют или имеют неверный тип.
func GetClaims(c *gin.Context) *auth.Claims {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil
	}
	claims, ok := v.(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
