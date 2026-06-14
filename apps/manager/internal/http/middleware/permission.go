package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
	"github.com/zalberix/cactus/libs/storage/db"
)

// PermissionChecker — интерфейс для проверки прав пользователя.
// Сигнатура совпадает с SQLC-генерированным методом store.ListPermissionsByUserID.
type PermissionChecker interface {
	ListPermissionsByUserID(ctx context.Context, userID int32) ([]db.Permission, error)
}

// RequirePermission — RBAC middleware для проверки права доступа к endpoint.
// Должен применяться после Auth middleware (требует claims в контексте).
func RequirePermission(checker PermissionChecker, perm permissions.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Unauthorized(c, "Отсутствует токен авторизации")
			c.Abort()
			return
		}

		perms, err := checker.ListPermissionsByUserID(c.Request.Context(), claims.UserID)
		if err != nil {
			response.InternalError(c, "Ошибка проверки прав")
			c.Abort()
			return
		}

		for _, p := range perms {
			if p.Slug == string(perm) {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Недостаточно прав")
		c.Abort()
	}
}
