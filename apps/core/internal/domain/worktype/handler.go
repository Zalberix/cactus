package worktype

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
)

// Handler — HTTP-обработчики для типов работ, воркеров и систем.
type Handler struct {
	service     *Service
	permChecker middleware.PermissionChecker
}

// NewHandler создаёт новый worktype Handler.
func NewHandler(service *Service, permChecker middleware.PermissionChecker) *Handler {
	return &Handler{
		service:     service,
		permChecker: permChecker,
	}
}

// RegisterRoutes регистрирует все маршруты worktype домена.
func (h *Handler) RegisterRoutes(r *gin.Engine, authMw gin.HandlerFunc, _ middleware.SystemTokenStore) {
	v1 := r.Group("/api/v1", authMw)

	// Systems (под организацией)
	v1.GET("/organizations/:orgId/systems",
		middleware.RequirePermission(h.permChecker, permissions.SystemRead), h.ListSystems)
	v1.POST("/organizations/:orgId/systems",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.CreateSystem)
	v1.PUT("/systems/:systemId",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.UpdateSystem)
	v1.DELETE("/systems/:systemId",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.DeleteSystem)

	// System Tokens
	v1.POST("/systems/:systemId/tokens",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.CreateToken)
	v1.GET("/systems/:systemId/tokens",
		middleware.RequirePermission(h.permChecker, permissions.SystemRead), h.ListTokens)
	v1.DELETE("/tokens/:tokenId",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.DeactivateToken)
	v1.POST("/tokens/:tokenId/activate",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.ActivateToken)
	v1.GET("/tokens/:tokenId/workflows",
		middleware.RequirePermission(h.permChecker, permissions.SystemRead), h.ListTokenWorkflows)
	v1.GET("/workflows/:workflowId/tokens",
		middleware.RequirePermission(h.permChecker, permissions.SystemRead), h.ListWorkflowTokens)
	v1.POST("/tokens/:tokenId/workflows/:workflowId",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.BindWorkflow)
	v1.DELETE("/tokens/:tokenId/workflows/:workflowId",
		middleware.RequirePermission(h.permChecker, permissions.SystemWrite), h.UnbindWorkflow)

	// Work Types
	v1.GET("/work-types",
		middleware.RequirePermission(h.permChecker, permissions.WorkerRead), h.ListWorkTypes)
	v1.POST("/work-types",
		middleware.RequirePermission(h.permChecker, permissions.WorkerWrite), h.CreateWorkType)
	v1.GET("/work-types/:workTypeId/workers",
		middleware.RequirePermission(h.permChecker, permissions.WorkerRead), h.ListWorkers)

	// Settings Schemas
	v1.GET("/worker-settings-schemas/:schemaId",
		middleware.RequirePermission(h.permChecker, permissions.WorkerRead), h.GetSettingsSchema)

	// Settings Revisions
	v1.GET("/worker-settings-schemas/:schemaId/revisions",
		middleware.RequirePermission(h.permChecker, permissions.WorkerRead), h.ListRevisions)
	v1.POST("/worker-settings-schemas/:schemaId/revisions",
		middleware.RequirePermission(h.permChecker, permissions.WorkerWrite), h.CreateRevision)

	// Публичный маршрут — регистрация воркера по bootstrap token (D-06, вне JWT-группы)
	r.POST("/api/v1/register/worker", h.RegisterWorker)
}

// parseID извлекает int32 ID из параметра URL.
func parseID(c *gin.Context, param string) (int32, bool) {
	raw := c.Param(param)
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		response.BadRequest(c, "INVALID_PARAM", "Неверный формат ID: "+param)
		return 0, false
	}
	return int32(id), true
}

// --- System handlers ---

// ListSystems godoc
// GET /api/v1/organizations/:orgId/systems
func (h *Handler) ListSystems(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}

	if c.Query("page") != "" || c.Query("per_page") != "" {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		systems, total, err := h.service.ListSystemsPaginated(c.Request.Context(), orgID, page, perPage)
		if err != nil {
			response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ СЃРёСЃС‚РµРј")
			return
		}
		response.OKPaginated(c, systems, total, page, perPage)
		return
	}

	systems, err := h.service.ListSystems(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Ошибка получения систем")
		return
	}
	response.OK(c, systems)
}

// CreateSystem godoc
// POST /api/v1/organizations/:orgId/systems
func (h *Handler) CreateSystem(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}
	var req CreateSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	system, err := h.service.CreateSystem(c.Request.Context(), orgID, req)
	if err != nil {
		response.InternalError(c, "Ошибка создания системы")
		return
	}
	response.Created(c, system)
}

// UpdateSystem godoc
// PUT /api/v1/systems/:systemId
func (h *Handler) UpdateSystem(c *gin.Context) {
	id, ok := parseID(c, "systemId")
	if !ok {
		return
	}
	var req UpdateSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	system, err := h.service.UpdateSystem(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, "Ошибка обновления системы")
		return
	}
	response.OK(c, system)
}

// DeleteSystem godoc
// DELETE /api/v1/systems/:systemId
func (h *Handler) DeleteSystem(c *gin.Context) {
	id, ok := parseID(c, "systemId")
	if !ok {
		return
	}

	if err := h.service.DeleteSystem(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Ошибка удаления системы")
		return
	}
	response.OK(c, gin.H{"message": "Система удалена"})
}

// --- System Token handlers ---

// CreateToken godoc
// POST /api/v1/systems/:systemId/tokens
// Токены генерируются в сервисном слое (D-18).
func (h *Handler) CreateToken(c *gin.Context) {
	systemID, ok := parseID(c, "systemId")
	if !ok {
		return
	}
	var req CreateSystemTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	tokenResp, err := h.service.CreateSystemToken(c.Request.Context(), systemID, req.Name)
	if err != nil {
		response.InternalError(c, "Ошибка создания токена")
		return
	}
	response.Created(c, tokenResp)
}

// ListTokens godoc
// GET /api/v1/systems/:systemId/tokens
func (h *Handler) ListTokens(c *gin.Context) {
	systemID, ok := parseID(c, "systemId")
	if !ok {
		return
	}

	tokens, err := h.service.ListSystemTokens(c.Request.Context(), systemID)
	if err != nil {
		response.InternalError(c, "Ошибка получения токенов")
		return
	}
	response.OK(c, tokens)
}

// DeactivateToken godoc
// DELETE /api/v1/tokens/:tokenId
func (h *Handler) DeactivateToken(c *gin.Context) {
	tokenID, ok := parseID(c, "tokenId")
	if !ok {
		return
	}

	if err := h.service.DeactivateSystemToken(c.Request.Context(), tokenID); err != nil {
		response.InternalError(c, "Ошибка деактивации токена")
		return
	}
	response.OK(c, gin.H{"message": "Токен деактивирован"})
}

// ActivateToken godoc
// POST /api/v1/tokens/:tokenId/activate
func (h *Handler) ActivateToken(c *gin.Context) {
	tokenID, ok := parseID(c, "tokenId")
	if !ok {
		return
	}

	if err := h.service.ActivateSystemToken(c.Request.Context(), tokenID); err != nil {
		response.InternalError(c, "Ошибка активации токена")
		return
	}
	response.OK(c, gin.H{"message": "Токен активирован"})
}

// BindWorkflow godoc
// POST /api/v1/tokens/:tokenId/workflows/:workflowId
func (h *Handler) BindWorkflow(c *gin.Context) {
	tokenID, ok := parseID(c, "tokenId")
	if !ok {
		return
	}
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}

	if err := h.service.BindWorkflowToToken(c.Request.Context(), tokenID, workflowID); err != nil {
		if errors.Is(err, ErrWorkflowTokenSystemMismatch) {
			response.BadRequest(c, "SYSTEM_MISMATCH", "System token and workflow must belong to the same system")
			return
		}
		response.InternalError(c, "Ошибка привязки workflow к токену")
		return
	}
	response.OK(c, gin.H{"message": "Workflow привязан к токену"})
}

// UnbindWorkflow godoc
// DELETE /api/v1/tokens/:tokenId/workflows/:workflowId
func (h *Handler) UnbindWorkflow(c *gin.Context) {
	tokenID, ok := parseID(c, "tokenId")
	if !ok {
		return
	}
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}

	if err := h.service.UnbindWorkflowFromToken(c.Request.Context(), tokenID, workflowID); err != nil {
		if errors.Is(err, ErrWorkflowTokenSystemMismatch) {
			response.BadRequest(c, "SYSTEM_MISMATCH", "System token and workflow must belong to the same system")
			return
		}
		response.InternalError(c, "Ошибка отвязки workflow от токена")
		return
	}
	response.OK(c, gin.H{"message": "Workflow отвязан от токена"})
}

// ListTokenWorkflows godoc
// GET /api/v1/tokens/:tokenId/workflows
func (h *Handler) ListTokenWorkflows(c *gin.Context) {
	tokenID, ok := parseID(c, "tokenId")
	if !ok {
		return
	}

	links, err := h.service.ListTokenWorkflows(c.Request.Context(), tokenID)
	if err != nil {
		response.InternalError(c, "Error fetching token workflows")
		return
	}
	response.OK(c, links)
}

// ListWorkflowTokens godoc
// GET /api/v1/workflows/:workflowId/tokens
func (h *Handler) ListWorkflowTokens(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}

	links, err := h.service.ListWorkflowTokens(c.Request.Context(), workflowID)
	if err != nil {
		response.InternalError(c, "Error fetching workflow tokens")
		return
	}
	response.OK(c, links)
}

// --- Work Type handlers ---

// ListWorkTypes godoc
// GET /api/v1/work-types
func (h *Handler) ListWorkTypes(c *gin.Context) {
	workTypes, err := h.service.ListWorkTypes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Ошибка получения типов работ")
		return
	}
	response.OK(c, workTypes)
}

// CreateWorkType godoc
// POST /api/v1/work-types
// Возвращает тип работы и bootstrap-токен (plaintext, один раз).
func (h *Handler) CreateWorkType(c *gin.Context) {
	var req CreateWorkTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	result, err := h.service.CreateWorkType(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Ошибка создания типа работы")
		return
	}
	response.Created(c, result)
}

// ListWorkers godoc
// GET /api/v1/work-types/:workTypeId/workers
func (h *Handler) ListWorkers(c *gin.Context) {
	workTypeID, ok := parseID(c, "workTypeId")
	if !ok {
		return
	}

	workers, err := h.service.ListWorkers(c.Request.Context(), workTypeID)
	if err != nil {
		response.InternalError(c, "Ошибка получения воркеров")
		return
	}
	response.OK(c, workers)
}

// --- Worker Registration handler ---

// RegisterWorker godoc
// POST /api/v1/register/worker
// Публичный endpoint — аутентификация по bootstrap token (не JWT).
func (h *Handler) RegisterWorker(c *gin.Context) {
	var req RegisterWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	worker, err := h.service.RegisterWorker(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "INVALID_TOKEN", "Неверный bootstrap токен или ошибка регистрации")
		return
	}
	response.OK(c, worker)
}

// --- Settings Schema handlers ---

// GetSettingsSchema godoc
// GET /api/v1/worker-settings-schemas/:schemaId
func (h *Handler) GetSettingsSchema(c *gin.Context) {
	schemaID, ok := parseID(c, "schemaId")
	if !ok {
		return
	}

	schema, err := h.service.GetSettingsSchema(c.Request.Context(), schemaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "Схема настроек не найдена")
			return
		}
		response.InternalError(c, "Ошибка получения схемы настроек")
		return
	}
	response.OK(c, schema)
}

// --- Settings Revision handlers ---

// ListRevisions godoc
// GET /api/v1/worker-settings-schemas/:schemaId/revisions
func (h *Handler) ListRevisions(c *gin.Context) {
	schemaID, ok := parseID(c, "schemaId")
	if !ok {
		return
	}

	revisions, err := h.service.ListSettingsRevisions(c.Request.Context(), schemaID)
	if err != nil {
		response.InternalError(c, "Ошибка получения ревизий настроек")
		return
	}
	response.OK(c, revisions)
}

// CreateRevision godoc
// POST /api/v1/worker-settings-schemas/:schemaId/revisions
func (h *Handler) CreateRevision(c *gin.Context) {
	schemaID, ok := parseID(c, "schemaId")
	if !ok {
		return
	}
	var req CreateRevisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	revision, err := h.service.CreateSettingsRevision(c.Request.Context(), schemaID, req)
	if err != nil {
		response.InternalError(c, "Ошибка создания ревизии настроек")
		return
	}
	response.Created(c, revision)
}

// ActivateRevision godoc
// PATCH /api/v1/worker-settings-revisions/:revisionId/activate
// Заглушка — activate логика будет реализована при наличии is_active поля в схеме.
func (h *Handler) ActivateRevision(c *gin.Context) {
	_, ok := parseID(c, "revisionId")
	if !ok {
		return
	}
	// TODO: implement when worker_settings_revision gains is_active field
	response.OK(c, gin.H{"message": "Ревизия активирована"})
}
