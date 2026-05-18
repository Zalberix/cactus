package workflow

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
)

// Handler вЂ” HTTP-РѕР±СЂР°Р±РѕС‚С‡РёРєРё РґР»СЏ workflow-РґРѕРјРµРЅР°.
type Handler struct {
	service     *Service
	permChecker middleware.PermissionChecker
}

// NewHandler СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ workflow Handler.
func NewHandler(service *Service, permChecker middleware.PermissionChecker) *Handler {
	return &Handler{
		service:     service,
		permChecker: permChecker,
	}
}

// RegisterRoutes СЂРµРіРёСЃС‚СЂРёСЂСѓРµС‚ РІСЃРµ РјР°СЂС€СЂСѓС‚С‹ workflow-РґРѕРјРµРЅР°.
func (h *Handler) RegisterRoutes(r *gin.Engine, authMw gin.HandlerFunc) {
	v1 := r.Group("/api/v1", authMw)

	// --- Workflow ---
	// GET /api/v1/systems/:systemId/workflows
	v1.GET("/systems/:systemId/workflows",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflows)
	// POST /api/v1/systems/:systemId/workflows
	v1.POST("/systems/:systemId/workflows",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflow)
	// GET /api/v1/workflows/:workflowId
	v1.GET("/workflows/:workflowId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.GetWorkflow)
	// PUT /api/v1/workflows/:workflowId
	v1.PUT("/workflows/:workflowId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflow)
	// GET /api/v1/workflows/:workflowId/input-schema
	v1.GET("/workflows/:workflowId/input-schema",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.GetWorkflowInputSchema)
	// POST /api/v1/workflows/:workflowId/input-schema/fields
	v1.POST("/workflows/:workflowId/input-schema/fields",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpsertWorkflowInputSchemaField)
	// DELETE /api/v1/workflows/:workflowId/input-schema/fields/:fieldName
	v1.DELETE("/workflows/:workflowId/input-schema/fields/:fieldName",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowInputSchemaField)
	// DELETE /api/v1/workflows/:workflowId
	v1.DELETE("/workflows/:workflowId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflow)

	// --- Versions ---
	// GET /api/v1/workflows/:workflowId/versions
	v1.GET("/workflows/:workflowId/versions",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListVersions)
	// GET /api/v1/workflows/:workflowId/version-summaries
	v1.GET("/workflows/:workflowId/version-summaries",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListVersionSummaries)
	// POST /api/v1/workflows/:workflowId/versions
	v1.POST("/workflows/:workflowId/versions",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateVersion)
	// PATCH /api/v1/versions/:versionId/name
	v1.PATCH("/versions/:versionId/name",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateVersionName)
	// POST /api/v1/versions/:versionId/copy
	v1.POST("/versions/:versionId/copy",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CopyVersion)
	// PUT /api/v1/workflows/:workflowId/traffic
	v1.PUT("/workflows/:workflowId/traffic",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowTraffic)
	// POST /api/v1/versions/:versionId/validate
	v1.POST("/versions/:versionId/validate",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.ValidateVersion)
	// PATCH /api/v1/versions/:versionId/activate
	v1.PATCH("/versions/:versionId/activate",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.ActivateVersion)
	// PATCH /api/v1/versions/:versionId/deactivate
	v1.PATCH("/versions/:versionId/deactivate",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeactivateVersion)
	// DELETE /api/v1/versions/:versionId
	v1.DELETE("/versions/:versionId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteVersion)

	// --- Steps ---
	// GET /api/v1/versions/:versionId/steps
	v1.GET("/versions/:versionId/steps",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListSteps)
	// POST /api/v1/versions/:versionId/steps
	v1.POST("/versions/:versionId/steps",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateStep)
	// PUT /api/v1/steps/:stepId
	v1.PUT("/steps/:stepId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateStep)
	// PUT /api/v1/steps/:stepId/task-settings
	v1.PUT("/steps/:stepId/task-settings",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateTaskSettings)
	// PUT /api/v1/steps/:stepId/input-mapping
	v1.PUT("/steps/:stepId/input-mapping",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateTaskInputMapping)
	// PATCH /api/v1/steps/:stepId/position
	v1.PATCH("/steps/:stepId/position",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateStepPosition)
	// DELETE /api/v1/steps/:stepId
	v1.DELETE("/steps/:stepId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteStep)

	// --- Dependencies ---
	// POST /api/v1/steps/:stepId/dependencies
	v1.POST("/steps/:stepId/dependencies",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateDependency)
	// DELETE /api/v1/steps/:stepId/dependencies/:dependsOnStepId
	v1.DELETE("/steps/:stepId/dependencies/:dependsOnStepId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteDependency)
}

// parseID РёР·РІР»РµРєР°РµС‚ int32 ID РёР· РїР°СЂР°РјРµС‚СЂР° URL.
func parseID(c *gin.Context, param string) (int32, bool) {
	raw := c.Param(param)
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		response.BadRequest(c, "INVALID_PARAM", "РќРµРІРµСЂРЅС‹Р№ С„РѕСЂРјР°С‚ ID: "+param)
		return 0, false
	}
	return int32(id), true
}

// --- Workflow handlers ---

// ListWorkflows godoc
// GET /api/v1/systems/:systemId/workflows
func (h *Handler) ListWorkflows(c *gin.Context) {
	systemID, ok := parseID(c, "systemId")
	if !ok {
		return
	}
	workflows, err := h.service.ListWorkflows(c.Request.Context(), systemID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ СЃРїРёСЃРєР° workflow")
		return
	}
	response.OK(c, workflows)
}

// CreateWorkflow godoc
// POST /api/v1/systems/:systemId/workflows
func (h *Handler) CreateWorkflow(c *gin.Context) {
	systemID, ok := parseID(c, "systemId")
	if !ok {
		return
	}
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	wf, err := h.service.CreateWorkflow(c.Request.Context(), systemID, req)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СЃРѕР·РґР°РЅРёСЏ workflow")
		return
	}
	response.Created(c, wf)
}

// GetWorkflow godoc
// GET /api/v1/workflows/:workflowId
func (h *Handler) GetWorkflow(c *gin.Context) {
	id, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	wf, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Workflow РЅРµ РЅР°Р№РґРµРЅ")
		return
	}
	response.OK(c, wf)
}

// UpdateWorkflow godoc
// PUT /api/v1/workflows/:workflowId
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	id, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	wf, err := h.service.UpdateWorkflow(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ workflow")
		return
	}
	response.OK(c, wf)
}

// DeleteWorkflow godoc
// DELETE /api/v1/workflows/:workflowId
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	id, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflow(c.Request.Context(), id); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СѓРґР°Р»РµРЅРёСЏ workflow")
		return
	}
	response.OK(c, gin.H{"message": "Workflow СѓРґР°Р»С‘РЅ"})
}

func (h *Handler) GetWorkflowInputSchema(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	schema, err := h.service.GetWorkflowInputSchema(c.Request.Context(), workflowID)
	if err != nil {
		response.InternalError(c, "Р С›РЎв‚¬Р С‘Р В±Р С”Р В° Р С—Р С•Р В»РЎС“РЎвЂЎР ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂ¦Р ВµР СРЎвЂ№ Р Р†РЎвЂ¦Р С•Р Т‘Р Р…РЎвЂ№РЎвЂ¦ Р С—Р В°РЎР‚Р В°Р СР ВµРЎвЂљРЎР‚Р С•Р Р†")
		return
	}
	response.OK(c, gin.H{"schema": schema})
}

func (h *Handler) UpsertWorkflowInputSchemaField(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var req InputSchemaFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	schema, err := h.service.UpsertWorkflowInputSchemaField(c.Request.Context(), workflowID, req)
	if err != nil {
		response.Fail(c, http.StatusUnprocessableEntity, "INVALID_INPUT_SCHEMA", err.Error())
		return
	}
	response.OK(c, gin.H{"schema": schema})
}

func (h *Handler) DeleteWorkflowInputSchemaField(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	fieldName := c.Param("fieldName")
	schema, err := h.service.DeleteWorkflowInputSchemaField(c.Request.Context(), workflowID, fieldName)
	if err != nil {
		response.Fail(c, http.StatusUnprocessableEntity, "INVALID_INPUT_SCHEMA", err.Error())
		return
	}
	response.OK(c, gin.H{"schema": schema})
}

// --- Version handlers ---

// ListVersions godoc
// GET /api/v1/workflows/:workflowId/versions
func (h *Handler) ListVersions(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	versions, err := h.service.ListVersions(c.Request.Context(), workflowID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ РІРµСЂСЃРёР№")
		return
	}
	response.OK(c, versions)
}

// ListVersionSummaries godoc
// GET /api/v1/workflows/:workflowId/version-summaries
func (h *Handler) ListVersionSummaries(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	summaries, err := h.service.ListVersionSummaries(c.Request.Context(), workflowID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ СЃРІРѕРґРєРё РІРµСЂСЃРёР№")
		return
	}
	response.OK(c, summaries)
}

// CreateVersion godoc
// POST /api/v1/workflows/:workflowId/versions
func (h *Handler) CreateVersion(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	claims := middleware.GetClaims(c)
	var userID int32
	if claims != nil {
		userID = claims.UserID
	}
	version, err := h.service.CreateVersion(c.Request.Context(), workflowID, userID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СЃРѕР·РґР°РЅРёСЏ РІРµСЂСЃРёРё")
		return
	}
	response.Created(c, version)
}

// UpdateVersionName godoc
// PATCH /api/v1/versions/:versionId/name
func (h *Handler) UpdateVersionName(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	var req UpdateVersionNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	version, err := h.service.UpdateVersionName(c.Request.Context(), versionID, req)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ РёРјРµРЅРё РІРµСЂСЃРёРё")
		return
	}
	response.OK(c, version)
}

// CopyVersion godoc
// POST /api/v1/versions/:versionId/copy
func (h *Handler) CopyVersion(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	claims := middleware.GetClaims(c)
	userID := int32(0)
	if claims != nil {
		userID = claims.UserID
	}
	version, err := h.service.CopyVersion(c.Request.Context(), versionID, userID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РєРѕРїРёСЂРѕРІР°РЅРёСЏ РІРµСЂСЃРёРё")
		return
	}
	response.Created(c, version)
}

// ValidateVersion godoc
// POST /api/v1/versions/:versionId/validate
// Р—Р°РїСѓСЃРєР°РµС‚ DAG-РІР°Р»РёРґР°С†РёСЋ. РџСЂРё РѕС€РёР±РєР°С… РІРѕР·РІСЂР°С‰Р°РµС‚ DAG_VALIDATION_FAILED.
func (h *Handler) ValidateVersion(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	result, err := h.service.ValidateVersion(c.Request.Context(), versionID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РІР°Р»РёРґР°С†РёРё РІРµСЂСЃРёРё")
		return
	}
	if !result.IsValid {
		details := make([]response.ErrorDetail, 0, len(result.Errors))
		for _, e := range result.Errors {
			details = append(details, response.ErrorDetail{
				Type:    e.Type,
				StepID:  e.StepID,
				Message: e.Message,
			})
		}
		response.Fail(c, http.StatusUnprocessableEntity, "DAG_VALIDATION_FAILED", "Р’РµСЂСЃРёСЏ СЃРѕРґРµСЂР¶РёС‚ РѕС€РёР±РєРё DAG", details...)
		return
	}
	response.OK(c, result)
}

// ActivateVersion godoc
// PATCH /api/v1/versions/:versionId/activate
func (h *Handler) ActivateVersion(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	if err := h.service.ActivateVersion(c.Request.Context(), versionID); err != nil {
		if errors.Is(err, ErrValidationRequired) {
			response.Fail(c, http.StatusUnprocessableEntity, "VALIDATION_REQUIRED", "Р’РµСЂСЃРёСЏ РґРѕР»Р¶РЅР° РїСЂРѕР№С‚Рё РІР°Р»РёРґР°С†РёСЋ DAG РїРµСЂРµРґ Р°РєС‚РёРІР°С†РёРµР№")
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° Р°РєС‚РёРІР°С†РёРё РІРµСЂСЃРёРё")
		return
	}
	response.OK(c, gin.H{"message": "Р’РµСЂСЃРёСЏ Р°РєС‚РёРІРёСЂРѕРІР°РЅР°"})
}

// DeactivateVersion godoc
// PATCH /api/v1/versions/:versionId/deactivate
func (h *Handler) DeactivateVersion(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	if err := h.service.DeactivateVersion(c.Request.Context(), versionID); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РґРµР°РєС‚РёРІР°С†РёРё РІРµСЂСЃРёРё")
		return
	}
	response.OK(c, gin.H{"message": "Р’РµСЂСЃРёСЏ РґРµР°РєС‚РёРІРёСЂРѕРІР°РЅР°"})
}

// DeleteVersion godoc
// DELETE /api/v1/versions/:versionId
func (h *Handler) DeleteVersion(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	if err := h.service.DeleteVersion(c.Request.Context(), versionID); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СѓРґР°Р»РµРЅРёСЏ РІРµСЂСЃРёРё")
		return
	}
	response.OK(c, gin.H{"message": "Р’РµСЂСЃРёСЏ СѓРґР°Р»РµРЅР°"})
}

// UpdateWorkflowTraffic godoc
// PUT /api/v1/workflows/:workflowId/traffic
func (h *Handler) UpdateWorkflowTraffic(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var req UpdateTrafficRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.service.UpdateWorkflowTraffic(c.Request.Context(), workflowID, req); err != nil {
		if errors.Is(err, ErrTrafficWeightInvalid) {
			response.Fail(c, http.StatusUnprocessableEntity, "TRAFFIC_WEIGHT_INVALID", "РЎСѓРјРјР° РІРµСЃРѕРІ С‚СЂР°С„РёРєР° РЅРµ РјРѕР¶РµС‚ РїСЂРµРІС‹С€Р°С‚СЊ 100")
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ РЅР°СЃС‚СЂРѕРµРє С‚СЂР°С„РёРєР°")
		return
	}
	response.OK(c, gin.H{"message": "РќР°СЃС‚СЂРѕР№РєРё С‚СЂР°С„РёРєР° СЃРѕС…СЂР°РЅРµРЅС‹"})
}

// --- Step handlers ---

// ListSteps godoc
// GET /api/v1/versions/:versionId/steps
// Р’РѕР·РІСЂР°С‰Р°РµС‚ enriched-С€Р°РіРё СЃ work_type meta Рё settings schemas.
func (h *Handler) ListSteps(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	steps, err := h.service.ListEnrichedSteps(c.Request.Context(), versionID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ С€Р°РіРѕРІ")
		return
	}
	deps, err := h.service.ListDependencies(c.Request.Context(), versionID)
	if err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РїРѕР»СѓС‡РµРЅРёСЏ Р·Р°РІРёСЃРёРјРѕСЃС‚РµР№")
		return
	}
	response.OK(c, gin.H{
		"steps":        steps,
		"dependencies": deps,
	})
}

// UpdateStepPosition godoc
// PATCH /api/v1/steps/:stepId/position
func (h *Handler) UpdateStepPosition(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	var req UpdateStepPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.service.UpdateStepPosition(c.Request.Context(), stepID, req); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ РїРѕР·РёС†РёРё С€Р°РіР°")
		return
	}
	response.OK(c, gin.H{"message": "РџРѕР·РёС†РёСЏ РѕР±РЅРѕРІР»РµРЅР°"})
}

// CreateStep godoc
// POST /api/v1/versions/:versionId/steps
func (h *Handler) CreateStep(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	var req CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	step, err := h.service.CreateStep(c.Request.Context(), versionID, req)
	if err != nil {
		if errors.Is(err, ErrStartStepProtected) {
			response.Fail(c, http.StatusUnprocessableEntity, "START_STEP_PROTECTED", "РЎС‚Р°СЂС‚РѕРІС‹Р№ Р±Р»РѕРє СѓРїСЂР°РІР»СЏРµС‚СЃСЏ СЃРёСЃС‚РµРјРѕР№")
			return
		}
		if errors.Is(err, ErrControlKindInvalid) {
			response.Fail(c, http.StatusUnprocessableEntity, "CONTROL_KIND_INVALID", "РќРµРїРѕРґРґРµСЂР¶РёРІР°РµРјС‹Р№ С‚РёРї control")
			return
		}
		if errors.Is(err, ErrControlSettingsInvalid) {
			response.Fail(c, http.StatusUnprocessableEntity, "CONTROL_SETTINGS_INVALID", "Invalid control settings")
			return
		}
		fmt.Printf("[CreateStep ERROR] versionID=%d stepType=%s workTypeID=%v controlKind=%v err=%v\n",
			versionID, req.StepType, req.WorkTypeID, req.ControlKind, err)
		response.InternalError(c, "РћС€РёР±РєР° СЃРѕР·РґР°РЅРёСЏ С€Р°РіР°: "+err.Error())
		return
	}
	response.Created(c, step)
}

// UpdateStep godoc
// PUT /api/v1/steps/:stepId
func (h *Handler) UpdateStep(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	var req UpdateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	step, err := h.service.UpdateStep(c.Request.Context(), stepID, req)
	if err != nil {
		if errors.Is(err, ErrStartStepProtected) {
			response.Fail(c, http.StatusUnprocessableEntity, "START_STEP_PROTECTED", "РЎС‚Р°СЂС‚РѕРІС‹Р№ Р±Р»РѕРє СѓРїСЂР°РІР»СЏРµС‚СЃСЏ СЃРёСЃС‚РµРјРѕР№")
			return
		}
		if errors.Is(err, ErrControlKindInvalid) {
			response.Fail(c, http.StatusUnprocessableEntity, "CONTROL_KIND_INVALID", "РќРµРїРѕРґРґРµСЂР¶РёРІР°РµРјС‹Р№ С‚РёРї control")
			return
		}
		if errors.Is(err, ErrControlSettingsInvalid) {
			response.Fail(c, http.StatusUnprocessableEntity, "CONTROL_SETTINGS_INVALID", "Invalid control settings")
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ С€Р°РіР°")
		return
	}
	response.OK(c, step)
}

func (h *Handler) UpdateTaskSettings(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	var req UpdateTaskSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.service.UpdateTaskSettings(c.Request.Context(), stepID, req); err != nil {
		if errors.Is(err, ErrTaskStepRequired) {
			response.Fail(c, http.StatusUnprocessableEntity, "TASK_STEP_REQUIRED", "РЁР°Рі РґРѕР»Р¶РµРЅ Р±С‹С‚СЊ task-С€Р°РіРѕРј")
			return
		}
		if errors.Is(err, ErrInvalidSettings) {
			response.Fail(c, http.StatusUnprocessableEntity, "INVALID_SETTINGS", err.Error())
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ РЅР°СЃС‚СЂРѕРµРє Р·Р°РґР°С‡Рё")
		return
	}
	response.OK(c, gin.H{"message": "РќР°СЃС‚СЂРѕР№РєРё Р·Р°РґР°С‡Рё СЃРѕС…СЂР°РЅРµРЅС‹"})
}

func (h *Handler) UpdateTaskInputMapping(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	var req UpdateTaskInputMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.service.UpdateTaskInputMapping(c.Request.Context(), stepID, req); err != nil {
		if errors.Is(err, ErrTaskStepRequired) {
			response.Fail(c, http.StatusUnprocessableEntity, "TASK_STEP_REQUIRED", "РЁР°Рі РґРѕР»Р¶РµРЅ Р±С‹С‚СЊ task-С€Р°РіРѕРј")
			return
		}
		if errors.Is(err, ErrInvalidInputMapping) {
			response.Fail(c, http.StatusUnprocessableEntity, "INVALID_INPUT_MAPPING", err.Error())
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° РѕР±РЅРѕРІР»РµРЅРёСЏ input mapping")
		return
	}
	response.OK(c, gin.H{"message": "Input mapping saved"})
}

// DeleteStep godoc
// DELETE /api/v1/steps/:stepId
func (h *Handler) DeleteStep(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	if err := h.service.DeleteStep(c.Request.Context(), stepID); err != nil {
		if errors.Is(err, ErrStartStepProtected) {
			response.Fail(c, http.StatusUnprocessableEntity, "START_STEP_PROTECTED", "РЎС‚Р°СЂС‚РѕРІС‹Р№ Р±Р»РѕРє СѓРїСЂР°РІР»СЏРµС‚СЃСЏ СЃРёСЃС‚РµРјРѕР№")
			return
		}
		response.InternalError(c, "РћС€РёР±РєР° СѓРґР°Р»РµРЅРёСЏ С€Р°РіР°")
		return
	}
	response.OK(c, gin.H{"message": "РЁР°Рі СѓРґР°Р»С‘РЅ"})
}

// --- Dependency handlers ---

// CreateDependency godoc
// POST /api/v1/steps/:stepId/dependencies
func (h *Handler) CreateDependency(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	var req CreateDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.service.CreateDependency(c.Request.Context(), stepID, req); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СЃРѕР·РґР°РЅРёСЏ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё")
		return
	}
	response.Created(c, gin.H{"message": "Р—Р°РІРёСЃРёРјРѕСЃС‚СЊ СЃРѕР·РґР°РЅР°"})
}

// DeleteDependency godoc
// DELETE /api/v1/steps/:stepId/dependencies/:dependsOnStepId
func (h *Handler) DeleteDependency(c *gin.Context) {
	stepID, ok := parseID(c, "stepId")
	if !ok {
		return
	}
	dependsOnStepID, ok := parseID(c, "dependsOnStepId")
	if !ok {
		return
	}
	if err := h.service.DeleteDependency(c.Request.Context(), stepID, dependsOnStepID); err != nil {
		response.InternalError(c, "РћС€РёР±РєР° СѓРґР°Р»РµРЅРёСЏ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё")
		return
	}
	response.OK(c, gin.H{"message": "Р—Р°РІРёСЃРёРјРѕСЃС‚СЊ СѓРґР°Р»РµРЅР°"})
}
