package message

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
)

// Handler — HTTP-обработчики для message domain.
type Handler struct {
	service *Service
}

// NewHandler создаёт новый Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// SendMessage обрабатывает запрос на отправку сообщения (per D-10, D-11).
//
// Авторизация: system token (X-Public-Token + X-Private-Token) ИЛИ JWT.
// Для system token: middleware.SystemTokenAuth уже проверил credentials,
// публичный токен используется для CheckWorkflowAccess.
// Для JWT: пользователь аутентифицирован через middleware.Auth.
func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", "Некорректный формат запроса",
			response.ErrorDetail{Message: err.Error()})
		return
	}

	// Определяем public token если system token auth
	publicToken := ""
	if token := middleware.GetSystemToken(c); token != nil {
		publicToken = token.PublicToken
	}

	resp, validationErrors, err := h.service.SendMessage(c.Request.Context(), req, publicToken)
	if err != nil {
		slog.Error("SendMessage error", slog.String("error", err.Error()))

		// Различаем типы ошибок
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "not found"):
			response.NotFound(c, err.Error())
		case strings.Contains(errMsg, "access denied"):
			response.Forbidden(c, err.Error())
		case strings.Contains(errMsg, "no active version"):
			response.BadRequest(c, "NO_ACTIVE_VERSION", err.Error())
		default:
			response.InternalError(c, "Ошибка обработки сообщения")
		}
		return
	}

	if len(validationErrors) > 0 {
		response.BadRequest(c, "VALIDATION_ERROR", "Payload не прошёл валидацию", validationErrors...)
		return
	}

	response.OK(c, resp)
}

// RegisterRoutes регистрирует маршруты message domain.
// Per D-11: endpoint доступен через system token auth (M2M) И JWT auth (UI).
//
// Поскольку Gin не допускает регистрацию одного пути с разными middleware,
// используем два пути:
//   - POST /api/v1/messages/send — M2M (system token auth)
//   - POST /api/v1/messages/send-user — UI (JWT auth)
//
// Оба пути ведут в один handler SendMessage.
func (h *Handler) RegisterRoutes(r *gin.Engine, jwtAuthMw gin.HandlerFunc, systemTokenAuthMw gin.HandlerFunc) {
	v1 := r.Group("/api/v1")

	// M2M endpoint (system token auth, per D-11)
	v1.POST("/messages/send", systemTokenAuthMw, h.SendMessage)

	// JWT endpoint (per D-11: JWT auth also accepted)
	jwtGroup := v1.Group("", jwtAuthMw)
	jwtGroup.POST("/messages/send-user", h.SendMessage)
}
