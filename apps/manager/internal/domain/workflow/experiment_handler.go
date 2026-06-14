package workflow

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/http/middleware"
	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
	db "github.com/zalberix/cactus/libs/storage/db"
)

type experimentBody struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ExperimentType string     `json:"experiment_type"`
	Status         string     `json:"status"`
	StartedAt      *time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
}

type experimentStatusBody struct {
	Status    string     `json:"status" binding:"required"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
}

type experimentScopeBody struct {
	WorkflowInputSchemaID     int32           `json:"workflow_input_schema_id"`
	TrafficConditions         json.RawMessage `json:"traffic_conditions"`
	TrafficPercent            *int32          `json:"traffic_percent"`
	FallbackPolicy            string          `json:"fallback_policy"`
	FallbackWorkflowVersionID *int32          `json:"fallback_workflow_version_id"`
}

type experimentVariantBody struct {
	WorkflowVersionID int32  `json:"workflow_version_id"`
	TrafficWeight     *int32 `json:"traffic_weight"`
	IsControlGroup    *bool  `json:"is_control_group"`
	IsActive          *bool  `json:"is_active"`
}

func (h *Handler) registerWorkflowExperimentRoutes(v1 *gin.RouterGroup) {
	v1.GET("/workflows/:workflowId/experiments",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowExperiments)
	v1.POST("/workflows/:workflowId/experiments",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowExperimentRecord)
	v1.GET("/experiments/:experimentId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.GetWorkflowExperimentRecord)
	v1.PUT("/experiments/:experimentId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowExperimentRecord)
	v1.PATCH("/experiments/:experimentId/status",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowExperimentRecordStatus)
	v1.DELETE("/experiments/:experimentId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowExperimentRecord)

	v1.GET("/experiments/:experimentId/scopes",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowExperimentScopes)
	v1.POST("/experiments/:experimentId/scopes",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowExperimentScopeRecord)
	v1.PUT("/experiment-scopes/:scopeId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowExperimentScopeRecord)
	v1.DELETE("/experiment-scopes/:scopeId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowExperimentScopeRecord)

	v1.GET("/experiment-scopes/:scopeId/variants",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowExperimentVariants)
	v1.POST("/experiment-scopes/:scopeId/variants",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowExperimentVariantRecord)
	v1.PUT("/experiment-variants/:variantId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowExperimentVariantRecord)
	v1.DELETE("/experiment-variants/:variantId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowExperimentVariantRecord)
}

func (h *Handler) ListWorkflowExperiments(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowExperiments(c.Request.Context(), workflowID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentResponses(items))
}

func (h *Handler) CreateWorkflowExperimentRecord(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var body experimentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowExperimentRecord(c.Request.Context(), workflowID, actorUserID(c), body.experimentMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, experimentResponse(item))
}

func (h *Handler) GetWorkflowExperimentRecord(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	item, err := h.service.GetWorkflowExperimentRecord(c.Request.Context(), experimentID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentResponse(item))
}

func (h *Handler) UpdateWorkflowExperimentRecord(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	var body experimentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowExperimentRecord(c.Request.Context(), experimentID, actorUserID(c), body.experimentMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentResponse(item))
}

func (h *Handler) UpdateWorkflowExperimentRecordStatus(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	var body experimentStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowExperimentRecordStatus(c.Request.Context(), experimentID, actorUserID(c), body.statusMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentResponse(item))
}

func (h *Handler) DeleteWorkflowExperimentRecord(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowExperimentRecord(c.Request.Context(), experimentID); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "experiment deleted"})
}

func (h *Handler) ListWorkflowExperimentScopes(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowExperimentScopes(c.Request.Context(), experimentID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentScopeResponses(items))
}

func (h *Handler) CreateWorkflowExperimentScopeRecord(c *gin.Context) {
	experimentID, ok := parseID(c, "experimentId")
	if !ok {
		return
	}
	var body experimentScopeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowExperimentScopeRecord(c.Request.Context(), experimentID, actorUserID(c), body.scopeMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, experimentScopeResponse(item))
}

func (h *Handler) UpdateWorkflowExperimentScopeRecord(c *gin.Context) {
	scopeID, ok := parseID(c, "scopeId")
	if !ok {
		return
	}
	var body experimentScopeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowExperimentScopeRecord(c.Request.Context(), scopeID, actorUserID(c), body.scopeMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentScopeResponse(item))
}

func (h *Handler) DeleteWorkflowExperimentScopeRecord(c *gin.Context) {
	scopeID, ok := parseID(c, "scopeId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowExperimentScopeRecord(c.Request.Context(), scopeID, actorUserID(c)); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "experiment scope deleted"})
}

func (h *Handler) ListWorkflowExperimentVariants(c *gin.Context) {
	scopeID, ok := parseID(c, "scopeId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowExperimentVariants(c.Request.Context(), scopeID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentVariantResponses(items))
}

func (h *Handler) CreateWorkflowExperimentVariantRecord(c *gin.Context) {
	scopeID, ok := parseID(c, "scopeId")
	if !ok {
		return
	}
	var body experimentVariantBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowExperimentVariantRecord(c.Request.Context(), scopeID, actorUserID(c), body.variantMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, experimentVariantResponse(item))
}

func (h *Handler) UpdateWorkflowExperimentVariantRecord(c *gin.Context) {
	variantID, ok := parseID(c, "variantId")
	if !ok {
		return
	}
	var body experimentVariantBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowExperimentVariantRecord(c.Request.Context(), variantID, actorUserID(c), body.variantMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, experimentVariantResponse(item))
}

func (h *Handler) DeleteWorkflowExperimentVariantRecord(c *gin.Context) {
	variantID, ok := parseID(c, "variantId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowExperimentVariantRecord(c.Request.Context(), variantID, actorUserID(c)); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "experiment variant deleted"})
}

func (b experimentBody) experimentMutation() workflowExperimentMutation {
	return workflowExperimentMutation{
		Name:           b.Name,
		Description:    b.Description,
		ExperimentType: b.ExperimentType,
		Status:         b.Status,
		StartedAt:      b.StartedAt,
		EndedAt:        b.EndedAt,
	}
}

func (b experimentStatusBody) statusMutation() workflowExperimentStatusMutation {
	return workflowExperimentStatusMutation{
		Status:    b.Status,
		StartedAt: b.StartedAt,
		EndedAt:   b.EndedAt,
	}
}

func (b experimentScopeBody) scopeMutation() workflowExperimentScopeMutation {
	return workflowExperimentScopeMutation{
		WorkflowInputSchemaID:     b.WorkflowInputSchemaID,
		TrafficConditions:         b.TrafficConditions,
		TrafficPercent:            b.TrafficPercent,
		FallbackPolicy:            b.FallbackPolicy,
		FallbackWorkflowVersionID: b.FallbackWorkflowVersionID,
	}
}

func (b experimentVariantBody) variantMutation() workflowExperimentVariantMutation {
	return workflowExperimentVariantMutation{
		WorkflowVersionID: b.WorkflowVersionID,
		TrafficWeight:     b.TrafficWeight,
		IsControlGroup:    b.IsControlGroup,
		IsActive:          b.IsActive,
	}
}

func experimentResponses(items []db.WorkflowExperiment) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, experimentResponse(item))
	}
	return result
}

func experimentResponse(item db.WorkflowExperiment) gin.H {
	return gin.H{
		"id":                 item.ID,
		"workflow_id":        item.WorkflowID,
		"name":               item.Name,
		"description":        item.Description,
		"experiment_type":    item.ExperimentType,
		"status":             item.Status,
		"started_at":         item.StartedAt,
		"ended_at":           item.EndedAt,
		"created_by_user_id": item.CreatedByUserID,
		"updated_by_user_id": item.UpdatedByUserID,
		"created_at":         item.CreatedAt,
		"updated_at":         item.UpdatedAt,
	}
}

func experimentScopeResponses(items []db.WorkflowExperimentScope) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, experimentScopeResponse(item))
	}
	return result
}

func experimentScopeResponse(item db.WorkflowExperimentScope) gin.H {
	return gin.H{
		"id":                           item.ID,
		"workflow_experiment_id":       item.WorkflowExperimentID,
		"workflow_input_schema_id":     item.WorkflowInputSchemaID,
		"traffic_conditions":           json.RawMessage(item.TrafficConditions),
		"conditions_hash":              item.ConditionsHash,
		"traffic_percent":              item.TrafficPercent,
		"fallback_policy":              item.FallbackPolicy,
		"fallback_workflow_version_id": item.FallbackWorkflowVersionID,
		"created_at":                   item.CreatedAt,
		"updated_at":                   item.UpdatedAt,
	}
}

func experimentVariantResponses(items []db.WorkflowExperimentVariant) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, experimentVariantResponse(item))
	}
	return result
}

func experimentVariantResponse(item db.WorkflowExperimentVariant) gin.H {
	return gin.H{
		"id":                           item.ID,
		"workflow_experiment_scope_id": item.WorkflowExperimentScopeID,
		"workflow_version_id":          item.WorkflowVersionID,
		"traffic_weight":               item.TrafficWeight,
		"is_control_group":             item.IsControlGroup,
		"is_active":                    item.IsActive,
		"created_at":                   item.CreatedAt,
		"updated_at":                   item.UpdatedAt,
	}
}
