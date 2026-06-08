package workflow

import (
	"encoding/json"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type CreateWorkflowRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Priority    int    `json:"priority" binding:"min=0,max=3"`
	Description string `json:"description"`
}

type UpdateWorkflowRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Priority    int    `json:"priority" binding:"min=0,max=3"`
	Description string `json:"description"`
}

type WorkflowListResponse struct {
	db.Workflow
	VersionCount       int32 `json:"version_count"`
	ActiveVersionCount int32 `json:"active_version_count"`
}

type WorkflowInputSchemaStatus string

const (
	WorkflowInputSchemaStatusDraft      WorkflowInputSchemaStatus = "draft"
	WorkflowInputSchemaStatusActive     WorkflowInputSchemaStatus = "active"
	WorkflowInputSchemaStatusDeprecated WorkflowInputSchemaStatus = "deprecated"
	WorkflowInputSchemaStatusArchived   WorkflowInputSchemaStatus = "archived"
)

type CompatibilityType string

const (
	CompatibilityTypeNative  CompatibilityType = "native"
	CompatibilityTypeAdapter CompatibilityType = "adapter"
	CompatibilityTypePartial CompatibilityType = "partial"
)

type ExperimentType string

const (
	ExperimentTypeCanary     ExperimentType = "canary"
	ExperimentTypeExperiment ExperimentType = "experiment"
	ExperimentTypeRollout    ExperimentType = "rollout"
)

type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusActive    ExperimentStatus = "active"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "cancelled"
)

type CreateWorkflowInputSchemaRequest struct {
	Code          string          `json:"code" binding:"required,min=1,max=255"`
	VersionNumber int32           `json:"version_number" binding:"min=1"`
	SchemaJSON    json.RawMessage `json:"schema_json" binding:"required"`
	IsDefault     bool            `json:"is_default"`
}

type UpdateWorkflowInputSchemaRequest struct {
	Code          string          `json:"code" binding:"required,min=1,max=255"`
	VersionNumber int32           `json:"version_number" binding:"min=1"`
	SchemaJSON    json.RawMessage `json:"schema_json" binding:"required"`
}

type WorkflowInputSchemaStatusRequest struct {
	Status WorkflowInputSchemaStatus `json:"status" binding:"required,oneof=draft active deprecated archived"`
}

type CreateWorkflowInputMapperRequest struct {
	MapperType string          `json:"mapper_type" binding:"required,oneof=internal jsonata jq javascript"`
	Rules      json.RawMessage `json:"rules" binding:"required"`
	IsActive   bool            `json:"is_active"`
}

type UpdateWorkflowInputMapperRequest = CreateWorkflowInputMapperRequest

type CreateCompatibilityRequest struct {
	WorkflowInputSchemaID int32             `json:"workflow_input_schema_id" binding:"required"`
	CompatibilityType     CompatibilityType `json:"compatibility_type" binding:"required,oneof=native adapter partial"`
	WorkflowInputMapperID *int32            `json:"workflow_input_mapper_id,omitempty"`
	DefaultValues         json.RawMessage   `json:"default_values,omitempty"`
	IsActive              bool              `json:"is_active"`
	IsDefaultRoute        bool              `json:"is_default_route"`
}

type UpdateCompatibilityRequest struct {
	CompatibilityType     CompatibilityType `json:"compatibility_type" binding:"required,oneof=native adapter partial"`
	WorkflowInputMapperID *int32            `json:"workflow_input_mapper_id,omitempty"`
	DefaultValues         json.RawMessage   `json:"default_values,omitempty"`
	IsActive              bool              `json:"is_active"`
	IsDefaultRoute        bool              `json:"is_default_route"`
}

type RoutingSupportedSchemaResponse struct {
	CompatibilityID     int32  `json:"compatibility_id"`
	SchemaID            int32  `json:"schema_id"`
	SchemaCode          string `json:"schema_code"`
	SchemaVersionNumber int32  `json:"schema_version_number"`
	CompatibilityType   string `json:"compatibility_type"`
	MapperID            *int32 `json:"mapper_id,omitempty"`
	SupportMode         string `json:"support_mode"`
	IsDefaultRoute      bool   `json:"is_default_route"`
}

type RoutingActiveTestResponse struct {
	ID             int32  `json:"id"`
	Name           string `json:"name"`
	ExperimentType string `json:"experiment_type"`
	Status         string `json:"status"`
}

type RoutingVersionRowResponse struct {
	WorkflowVersionID        int32                            `json:"workflow_version_id"`
	WorkflowID               int32                            `json:"workflow_id"`
	WorkflowVersionName      string                           `json:"workflow_version_name"`
	WorkflowVersionNumber    int32                            `json:"workflow_version_number"`
	IsValid                  bool                             `json:"is_valid"`
	IsActive                 bool                             `json:"is_active"`
	NativeInputSchemaID      int32                            `json:"native_input_schema_id"`
	NativeInputSchemaCode    string                           `json:"native_input_schema_code"`
	NativeInputSchemaVersion int32                            `json:"native_input_schema_version_number"`
	SupportedSchemas         []RoutingSupportedSchemaResponse `json:"supported_schemas"`
	ActiveTests              []RoutingActiveTestResponse      `json:"active_tests"`
}

type InputSchemaUsageResponse struct {
	UsedByMessage    bool     `json:"used_by_message"`
	UsedByMapper     bool     `json:"used_by_mapper"`
	UsedByExperiment bool     `json:"used_by_experiment"`
	IsReadOnly       bool     `json:"is_readonly"`
	Reasons          []string `json:"reasons"`
}

type ArchiveInputSchemaRequest struct {
	ConfirmationName string `json:"confirmation_name" binding:"required"`
}

type CompatibilityMappingFieldRequest struct {
	TargetPath string          `json:"target_path"`
	SourcePath string          `json:"source_path,omitempty"`
	Default    json.RawMessage `json:"default,omitempty"`
	Ignore     bool            `json:"ignore,omitempty"`
}

type ValidateAndCreateCompatibilityRequest struct {
	TargetInputSchemaID int32                              `json:"target_input_schema_id" binding:"required"`
	Fields              []CompatibilityMappingFieldRequest `json:"fields" binding:"required"`
	IsDefaultRoute      bool                               `json:"is_default_route"`
}

type CompatibilityValidationResponse struct {
	IsValid bool                   `json:"is_valid"`
	Errors  []response.ErrorDetail `json:"errors,omitempty"`
}

type InputSchemaCompatibilityRowResponse struct {
	ID                    int32  `json:"id"`
	WorkflowVersionID     int32  `json:"workflow_version_id"`
	WorkflowVersionName   string `json:"workflow_version_name"`
	WorkflowVersionNumber int32  `json:"workflow_version_number"`
	WorkflowInputSchemaID int32  `json:"workflow_input_schema_id"`
	CompatibilityType     string `json:"compatibility_type"`
	WorkflowInputMapperID *int32 `json:"workflow_input_mapper_id,omitempty"`
	IsActive              bool   `json:"is_active"`
	IsDefaultRoute        bool   `json:"is_default_route"`
}

type CreateExperimentRequest struct {
	Name           string         `json:"name" binding:"required,min=1,max=255"`
	Description    string         `json:"description"`
	ExperimentType ExperimentType `json:"experiment_type" binding:"required,oneof=canary experiment rollout"`
	StartedAt      string         `json:"started_at,omitempty"`
	EndedAt        string         `json:"ended_at,omitempty"`
}

type UpdateExperimentRequest = CreateExperimentRequest

type CreateExperimentScopeRequest struct {
	WorkflowInputSchemaID     int32           `json:"workflow_input_schema_id" binding:"required"`
	TrafficConditions         json.RawMessage `json:"traffic_conditions"`
	TrafficPercent            int32           `json:"traffic_percent" binding:"min=1,max=100"`
	FallbackPolicy            string          `json:"fallback_policy" binding:"omitempty,oneof=error default_route explicit_version"`
	FallbackWorkflowVersionID *int32          `json:"fallback_workflow_version_id,omitempty"`
}

type UpdateExperimentScopeRequest = CreateExperimentScopeRequest

type CreateExperimentVariantRequest struct {
	WorkflowVersionID int32 `json:"workflow_version_id" binding:"required"`
	TrafficWeight     int32 `json:"traffic_weight" binding:"min=0,max=100"`
	IsControlGroup    bool  `json:"is_control_group"`
	IsActive          bool  `json:"is_active"`
}

type UpdateExperimentVariantRequest = CreateExperimentVariantRequest

type CreateStepRequest struct {
	Name                     string          `json:"name" binding:"omitempty,max=255"`
	StepType                 string          `json:"step_type" binding:"required,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	WorkerSettingsSchemaID   *int32          `json:"worker_settings_schema_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
	CanvasPosition           json.RawMessage `json:"canvas_position"`
}

type UpdateStepRequest struct {
	Name                     *string         `json:"name" binding:"omitempty,min=1,max=255"`
	StepType                 *string         `json:"step_type" binding:"omitempty,oneof=task control"`
	WorkTypeID               *int32          `json:"work_type_id"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id"`
	WorkerSettingsSchemaID   *int32          `json:"worker_settings_schema_id"`
	ControlKind              *string         `json:"control_kind"`
	ControlSettings          json.RawMessage `json:"control_settings"`
	InputMapping             json.RawMessage `json:"input_mapping"`
	CanvasPosition           json.RawMessage `json:"canvas_position"`
}

type UpdateStepPositionRequest struct {
	CanvasPosition json.RawMessage `json:"canvas_position" binding:"required"`
}

type CreateDependencyRequest struct {
	DependsOnStepID int32  `json:"depends_on_step_id" binding:"required"`
	Outcome         string `json:"outcome" binding:"required"`
	OutputIndex     int32  `json:"output_index"`
}

type EnrichedStepResponse struct {
	ID                       int32           `json:"id"`
	WorkflowVersionID        int32           `json:"workflow_version_id"`
	Name                     string          `json:"name"`
	StepType                 string          `json:"step_type"`
	WorkTypeID               *int32          `json:"work_type_id,omitempty"`
	WorkerSettingsRevisionID *int32          `json:"worker_settings_revision_id,omitempty"`
	ControlKind              *string         `json:"control_kind,omitempty"`
	ControlSettings          json.RawMessage `json:"control_settings,omitempty"`
	InputMapping             json.RawMessage `json:"input_mapping,omitempty"`
	CanvasPosition           json.RawMessage `json:"canvas_position,omitempty"`
	WorkTypeName             *string         `json:"work_type_name,omitempty"`
	WorkTypeCode             *string         `json:"work_type_code,omitempty"`
	WorkTypeMeta             json.RawMessage `json:"work_type_meta,omitempty"`
	SettingsSchema           json.RawMessage `json:"settings_schema,omitempty"`
	InputSchema              json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema             json.RawMessage `json:"output_schema,omitempty"`
	Config                   json.RawMessage `json:"config,omitempty"`
}

type ValidateVersionResponse struct {
	IsValid bool                     `json:"is_valid"`
	Errors  []dagpkg.ValidationError `json:"errors,omitempty"`
}

type VersionSummaryResponse struct {
	ID                    int32  `json:"id"`
	WorkflowID            int32  `json:"workflow_id"`
	Name                  string `json:"name"`
	VersionNumber         int32  `json:"version_number"`
	IsValid               bool   `json:"is_valid"`
	IsActive              bool   `json:"is_active"`
	CompatibleSchemaCount int64  `json:"compatible_schema_count"`
	RunCount              int64  `json:"run_count"`
	LockedAt              string `json:"locked_at,omitempty"`
	PublishedAt           string `json:"published_at,omitempty"`
	ArchivedAt            string `json:"archived_at,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at,omitempty"`
	DeletedAt             string `json:"deleted_at,omitempty"`
}

type UpdateVersionNameRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

type InputSchemaFieldRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Type        string `json:"type" binding:"required,oneof=string number integer boolean object array"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type InputSchemaFieldResponse struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

type InputSchemaResponse struct {
	Schema map[string]any `json:"schema"`
}

type UpdateTaskSettingsRequest struct {
	SettingsData json.RawMessage `json:"settings_data" binding:"required"`
}

type UpdateTaskInputMappingRequest struct {
	InputMapping json.RawMessage `json:"input_mapping" binding:"required"`
}
