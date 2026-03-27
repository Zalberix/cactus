package message

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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

// GetMessageStatus returns message status (per D-20).
// GET /api/v1/messages/:id/status
func (h *Handler) GetMessageStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "Invalid message ID")
		return
	}

	resp, err := h.service.GetMessageStatus(c.Request.Context(), int32(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			response.Fail(c, http.StatusNotFound, "MESSAGE_NOT_FOUND",
				fmt.Sprintf("Message with ID %d not found", id))
			return
		}
		response.InternalError(c, "Error fetching message status")
		return
	}

	response.OK(c, resp)
}

// ListMessages returns paginated messages for an organization (per UI-12).
// GET /api/v1/organizations/:orgId/messages?page=1&per_page=20
func (h *Handler) ListMessages(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "Invalid organization ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	items, total, err := h.service.ListMessages(c.Request.Context(), int32(orgID), page, perPage)
	if err != nil {
		response.InternalError(c, "Ошибка получения списка сообщений")
		return
	}

	response.OKPaginated(c, items, total, page, perPage)
}

// RegisterRoutes регистрирует маршруты message domain.
// Per D-11: endpoint доступен через system token auth (M2M) И JWT auth (UI).
//
// Поскольку Gin не допускает регистрацию одного пути с разными middleware,
// используем два пути:
//   - POST /api/v1/messages/send — M2M (system token auth)
//   - POST /api/v1/messages/send-user — UI (JWT auth)
//   - GET /api/v1/messages/:id/status-system — M2M (system token auth, per D-22)
//   - GET /api/v1/messages/:id/status — UI (JWT auth, per D-22)
//
// Оба пути ведут в один handler SendMessage/GetMessageStatus.
func (h *Handler) RegisterRoutes(r *gin.Engine, jwtAuthMw gin.HandlerFunc, systemTokenAuthMw gin.HandlerFunc) {
	v1 := r.Group("/api/v1")

	// M2M endpoints (system token auth, per D-11, D-22)
	v1.POST("/messages/send", systemTokenAuthMw, h.SendMessage)
	v1.GET("/messages/:id/status-system", systemTokenAuthMw, h.GetMessageStatus)

	// JWT endpoints (per D-11, D-22: JWT auth also accepted)
	jwtGroup := v1.Group("", jwtAuthMw)
	jwtGroup.POST("/messages/send-user", h.SendMessage)
	jwtGroup.GET("/messages/:id/status", h.GetMessageStatus)
	jwtGroup.GET("/organizations/:orgId/messages", h.ListMessages)
}
