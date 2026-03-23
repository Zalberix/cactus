package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/core/internal/http/response"
)

// Handler реализует HTTP обработчики для аутентификации.
type Handler struct {
	service *Service
}

// NewHandler создаёт новый auth.Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login — POST /api/v1/auth/login
// Принимает email + пароль, возвращает JWT access + refresh токены.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// Refresh — POST /api/v1/auth/refresh
// Принимает refresh токен, возвращает новую пару access + refresh токенов.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, resp)
}
