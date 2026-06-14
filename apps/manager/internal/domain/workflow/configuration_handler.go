package workflow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/manager/internal/http/middleware"
	"github.com/zalberix/cactus/apps/manager/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
	db "github.com/zalberix/cactus/libs/storage/db"
)

type inputSchemaBody struct {
	Code          string          `json:"code"`
	VersionNumber int32           `json:"version_number"`
	Schema        json.RawMessage `json:"schema"`
	SchemaJSON    json.RawMessage `json:"schema_json"`
	Status        string          `json:"status"`
	IsDefault     *bool           `json:"is_default"`
}

type inputSchemaStatusBody struct {
	Status string `json:"status" binding:"required"`
}

type inputMapperBody struct {
	MapperType string          `json:"mapper_type"`
	Rules      json.RawMessage `json:"rules"`
	IsActive   *bool           `json:"is_active"`
}

type activeBody struct {
	IsActive bool `json:"is_active"`
}

type compatibilityBody struct {
	WorkflowInputSchemaID int32           `json:"workflow_input_schema_id"`
	CompatibilityType     string          `json:"compatibility_type"`
	WorkflowInputMapperID *int32          `json:"workflow_input_mapper_id"`
	DefaultValues         json.RawMessage `json:"default_values"`
	IsActive              *bool           `json:"is_active"`
	IsDefaultRoute        *bool           `json:"is_default_route"`
}

type defaultRouteBody struct {
	IsDefaultRoute bool `json:"is_default_route"`
}

func (h *Handler) registerWorkflowConfigurationRoutes(v1 *gin.RouterGroup) {
	v1.GET("/workflows/:workflowId/routing/versions",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowRoutingVersionRows)
	v1.GET("/workflows/:workflowId/input-schemas",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowInputSchemas)
	v1.GET("/workflows/:workflowId/input-schemas/search",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.SearchWorkflowInputSchemas)
	v1.POST("/workflows/:workflowId/input-schemas",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowInputSchemaRecord)
	v1.GET("/input-schemas/:inputSchemaId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.GetWorkflowInputSchemaRecord)
	v1.GET("/input-schemas/:inputSchemaId/usage",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.GetWorkflowInputSchemaUsage)
	v1.PUT("/input-schemas/:inputSchemaId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowInputSchemaRecord)
	v1.PATCH("/input-schemas/:inputSchemaId/status",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowInputSchemaRecordStatus)
	v1.PATCH("/input-schemas/:inputSchemaId/default",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.SetDefaultWorkflowInputSchemaRecord)
	v1.POST("/input-schemas/:inputSchemaId/archive",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.ArchiveWorkflowInputSchemaRecord)
	v1.DELETE("/input-schemas/:inputSchemaId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowInputSchemaRecord)

	v1.GET("/workflows/:workflowId/input-mappers",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowInputMappers)
	v1.POST("/workflows/:workflowId/input-mappers",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowInputMapperRecord)
	v1.PUT("/input-mappers/:inputMapperId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowInputMapperRecord)
	v1.PATCH("/input-mappers/:inputMapperId/active",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowInputMapperActive)
	v1.DELETE("/input-mappers/:inputMapperId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowInputMapperRecord)

	v1.GET("/workflows/:workflowId/schema-compatibilities",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListWorkflowCompatibilities)
	v1.GET("/versions/:versionId/schema-compatibilities",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListVersionCompatibilities)
	v1.GET("/input-schemas/:inputSchemaId/compatibilities",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowRead), h.ListInputSchemaCompatibilities)
	v1.POST("/input-schemas/:inputSchemaId/compatibilities",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.ValidateAndCreateInputSchemaCompatibility)
	v1.POST("/versions/:versionId/schema-compatibilities",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.CreateWorkflowCompatibility)
	v1.PUT("/schema-compatibilities/:compatibilityId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowCompatibility)
	v1.PATCH("/schema-compatibilities/:compatibilityId/default-route",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.UpdateWorkflowCompatibilityDefaultRoute)
	v1.PATCH("/schema-compatibilities/:compatibilityId/deactivate",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeactivateWorkflowCompatibility)
	v1.DELETE("/schema-compatibilities/:compatibilityId",
		middleware.RequirePermission(h.permChecker, permissions.WorkflowWrite), h.DeleteWorkflowCompatibility)
}

func (h *Handler) ListWorkflowRoutingVersionRows(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	page, perPage := parsePaginationQuery(c)
	rows, total, err := h.service.ListWorkflowRoutingVersionRows(c.Request.Context(), workflowID, page, perPage)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OKPaginated(c, rows, total, page, perPage)
}

func (h *Handler) ListWorkflowInputSchemas(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowInputSchemas(c.Request.Context(), workflowID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponses(items))
}

func (h *Handler) SearchWorkflowInputSchemas(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	excludeID := int32(0)
	if raw := c.Query("exclude_input_schema_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err == nil && parsed > 0 {
			excludeID = int32(parsed)
		}
	}
	items, err := h.service.SearchWorkflowInputSchemas(c.Request.Context(), workflowID, c.Query("query"), excludeID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponses(items))
}

func (h *Handler) CreateWorkflowInputSchemaRecord(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var body inputSchemaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowInputSchemaRecord(c.Request.Context(), workflowID, actorUserID(c), body.inputSchemaMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, inputSchemaResponse(item))
}

func (h *Handler) GetWorkflowInputSchemaRecord(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	item, err := h.service.GetWorkflowInputSchemaRecord(c.Request.Context(), inputSchemaID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	usage, err := h.service.GetWorkflowInputSchemaUsage(c.Request.Context(), inputSchemaID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponseWithUsage(item, usage))
}

func (h *Handler) UpdateWorkflowInputSchemaRecord(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	var body inputSchemaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowInputSchemaRecord(c.Request.Context(), inputSchemaID, actorUserID(c), body.inputSchemaMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponse(item))
}

func (h *Handler) GetWorkflowInputSchemaUsage(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	usage, err := h.service.GetWorkflowInputSchemaUsage(c.Request.Context(), inputSchemaID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, usage)
}

func (h *Handler) UpdateWorkflowInputSchemaRecordStatus(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	var body inputSchemaStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowInputSchemaRecordStatus(c.Request.Context(), inputSchemaID, actorUserID(c), body.Status)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponse(item))
}

func (h *Handler) SetDefaultWorkflowInputSchemaRecord(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	item, err := h.service.SetDefaultWorkflowInputSchemaRecord(c.Request.Context(), inputSchemaID, actorUserID(c))
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponse(item))
}

func (h *Handler) ArchiveWorkflowInputSchemaRecord(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	var body ArchiveInputSchemaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.ArchiveWorkflowInputSchemaRecord(c.Request.Context(), inputSchemaID, actorUserID(c), body.ConfirmationName)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputSchemaResponse(item))
}

func (h *Handler) DeleteWorkflowInputSchemaRecord(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	var body ArchiveInputSchemaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", "delete requires archive confirmation")
		return
	}
	if _, err := h.service.ArchiveWorkflowInputSchemaRecord(c.Request.Context(), inputSchemaID, actorUserID(c), body.ConfirmationName); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "input schema archived"})
}

func (h *Handler) ListWorkflowInputMappers(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowInputMappers(c.Request.Context(), workflowID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputMapperResponses(items))
}

func (h *Handler) CreateWorkflowInputMapperRecord(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	var body inputMapperBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowInputMapperRecord(c.Request.Context(), workflowID, actorUserID(c), body.inputMapperMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, inputMapperResponse(item))
}

func (h *Handler) UpdateWorkflowInputMapperRecord(c *gin.Context) {
	mapperID, ok := parseID(c, "inputMapperId")
	if !ok {
		return
	}
	var body inputMapperBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowInputMapperRecord(c.Request.Context(), mapperID, actorUserID(c), body.inputMapperMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputMapperResponse(item))
}

func (h *Handler) UpdateWorkflowInputMapperActive(c *gin.Context) {
	mapperID, ok := parseID(c, "inputMapperId")
	if !ok {
		return
	}
	var body activeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowInputMapperActive(c.Request.Context(), mapperID, actorUserID(c), body.IsActive)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, inputMapperResponse(item))
}

func (h *Handler) DeleteWorkflowInputMapperRecord(c *gin.Context) {
	mapperID, ok := parseID(c, "inputMapperId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowInputMapperRecord(c.Request.Context(), mapperID, actorUserID(c)); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "input mapper deleted"})
}

func (h *Handler) ListWorkflowCompatibilities(c *gin.Context) {
	workflowID, ok := parseID(c, "workflowId")
	if !ok {
		return
	}
	items, err := h.service.ListWorkflowCompatibilities(c.Request.Context(), workflowID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, compatibilityResponses(items))
}

func (h *Handler) ListVersionCompatibilities(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	items, err := h.service.ListVersionCompatibilities(c.Request.Context(), versionID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, compatibilityResponses(items))
}

func (h *Handler) ListInputSchemaCompatibilities(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	items, err := h.service.ListInputSchemaCompatibilityRows(c.Request.Context(), inputSchemaID)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *Handler) CreateWorkflowCompatibility(c *gin.Context) {
	versionID, ok := parseID(c, "versionId")
	if !ok {
		return
	}
	var body compatibilityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateWorkflowCompatibility(c.Request.Context(), versionID, actorUserID(c), body.compatibilityMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.Created(c, compatibilityResponse(item))
}

func (h *Handler) ValidateAndCreateInputSchemaCompatibility(c *gin.Context) {
	inputSchemaID, ok := parseID(c, "inputSchemaId")
	if !ok {
		return
	}
	var body ValidateAndCreateCompatibilityRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, validation, err := h.service.ValidateAndCreateInputSchemaCompatibility(c.Request.Context(), inputSchemaID, actorUserID(c), body)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	if !validation.IsValid {
		response.Fail(c, http.StatusUnprocessableEntity, "COMPATIBILITY_VALIDATION_FAILED", "compatibility validation failed", validation.Errors...)
		return
	}
	response.Created(c, compatibilityResponse(item))
}

func (h *Handler) UpdateWorkflowCompatibility(c *gin.Context) {
	compatibilityID, ok := parseID(c, "compatibilityId")
	if !ok {
		return
	}
	var body compatibilityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowCompatibility(c.Request.Context(), compatibilityID, actorUserID(c), body.compatibilityMutation())
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, compatibilityResponse(item))
}

func (h *Handler) UpdateWorkflowCompatibilityDefaultRoute(c *gin.Context) {
	compatibilityID, ok := parseID(c, "compatibilityId")
	if !ok {
		return
	}
	var body defaultRouteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateWorkflowCompatibilityDefaultRoute(c.Request.Context(), compatibilityID, actorUserID(c), body.IsDefaultRoute)
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, compatibilityResponse(item))
}

func (h *Handler) DeactivateWorkflowCompatibility(c *gin.Context) {
	compatibilityID, ok := parseID(c, "compatibilityId")
	if !ok {
		return
	}
	item, err := h.service.DeactivateWorkflowCompatibility(c.Request.Context(), compatibilityID, actorUserID(c))
	if err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, compatibilityResponse(item))
}

func (h *Handler) DeleteWorkflowCompatibility(c *gin.Context) {
	compatibilityID, ok := parseID(c, "compatibilityId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowCompatibility(c.Request.Context(), compatibilityID, actorUserID(c)); err != nil {
		writeWorkflowConfigurationError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "compatibility deleted or deactivated"})
}

func (b inputSchemaBody) inputSchemaMutation() inputSchemaMutation {
	schemaJSON := b.SchemaJSON
	if len(schemaJSON) == 0 {
		schemaJSON = b.Schema
	}
	return inputSchemaMutation{
		Code:          b.Code,
		VersionNumber: b.VersionNumber,
		SchemaJSON:    schemaJSON,
		Status:        b.Status,
		IsDefault:     b.IsDefault,
	}
}

func (b inputMapperBody) inputMapperMutation() inputMapperMutation {
	return inputMapperMutation{
		MapperType: b.MapperType,
		Rules:      b.Rules,
		IsActive:   b.IsActive,
	}
}

func (b compatibilityBody) compatibilityMutation() compatibilityMutation {
	return compatibilityMutation{
		WorkflowInputSchemaID: b.WorkflowInputSchemaID,
		CompatibilityType:     b.CompatibilityType,
		WorkflowInputMapperID: b.WorkflowInputMapperID,
		DefaultValues:         b.DefaultValues,
		IsActive:              b.IsActive,
		IsDefaultRoute:        b.IsDefaultRoute,
	}
}

func actorUserID(c *gin.Context) int32 {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

func parsePaginationQuery(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	return normalizeRoutingPagination(page, perPage)
}

func writeWorkflowConfigurationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrWorkflowConfigurationInvalid):
		response.Fail(c, http.StatusUnprocessableEntity, "INVALID_WORKFLOW_CONFIGURATION", err.Error())
	case errors.Is(err, ErrWorkflowConfigurationInUse):
		response.Fail(c, http.StatusConflict, "WORKFLOW_CONFIGURATION_IN_USE", err.Error())
	case errors.Is(err, ErrWorkflowConfigurationNotFound):
		response.NotFound(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}

func inputSchemaResponses(items []db.WorkflowInputSchema) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, inputSchemaResponse(item))
	}
	return result
}

func inputSchemaResponse(item db.WorkflowInputSchema) gin.H {
	return gin.H{
		"id":                 item.ID,
		"workflow_id":        item.WorkflowID,
		"code":               item.Code,
		"version_number":     item.VersionNumber,
		"schema_json":        json.RawMessage(item.SchemaJson),
		"status":             item.Status,
		"is_default":         item.IsDefault,
		"created_by_user_id": item.CreatedByUserID,
		"updated_by_user_id": item.UpdatedByUserID,
		"created_at":         item.CreatedAt,
		"updated_at":         item.UpdatedAt,
	}
}

func inputSchemaResponseWithUsage(item db.WorkflowInputSchema, usage InputSchemaUsageResponse) gin.H {
	resp := inputSchemaResponse(item)
	resp["usage"] = usage
	return resp
}

func inputMapperResponses(items []db.WorkflowInputMapper) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, inputMapperResponse(item))
	}
	return result
}

func inputMapperResponse(item db.WorkflowInputMapper) gin.H {
	return gin.H{
		"id":                 item.ID,
		"workflow_id":        item.WorkflowID,
		"mapper_type":        item.MapperType,
		"rules":              json.RawMessage(item.Rules),
		"is_active":          item.IsActive,
		"created_by_user_id": item.CreatedByUserID,
		"updated_by_user_id": item.UpdatedByUserID,
		"created_at":         item.CreatedAt,
		"updated_at":         item.UpdatedAt,
	}
}

func compatibilityResponses(items []db.WorkflowVersionInputSchemaCompatibility) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, compatibilityResponse(item))
	}
	return result
}

func compatibilityResponse(item db.WorkflowVersionInputSchemaCompatibility) gin.H {
	return gin.H{
		"id":                       item.ID,
		"workflow_version_id":      item.WorkflowVersionID,
		"workflow_input_schema_id": item.WorkflowInputSchemaID,
		"compatibility_type":       item.CompatibilityType,
		"workflow_input_mapper_id": item.WorkflowInputMapperID,
		"default_values":           json.RawMessage(item.DefaultValues),
		"is_active":                item.IsActive,
		"is_default_route":         item.IsDefaultRoute,
		"created_by_user_id":       item.CreatedByUserID,
		"updated_by_user_id":       item.UpdatedByUserID,
		"created_at":               item.CreatedAt,
		"updated_at":               item.UpdatedAt,
	}
}
