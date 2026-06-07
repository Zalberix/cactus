package workflow

import (
	"context"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// Storage — интерфейс хранилища для workflow домена.
type Storage interface {
	WithTx(ctx context.Context, fn func(q *db.Queries) error) error

	// Workflow
	CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) (db.Workflow, error)
	GetWorkflowByID(ctx context.Context, id int32) (db.Workflow, error)
	ListWorkflowsBySystemID(ctx context.Context, systemID int32) ([]db.Workflow, error)
	UpdateWorkflow(ctx context.Context, arg db.UpdateWorkflowParams) (db.Workflow, error)
	SoftDeleteWorkflow(ctx context.Context, id int32) error

	// WorkflowVersion
	CreateWorkflowVersion(ctx context.Context, arg db.CreateWorkflowVersionParams) (db.WorkflowVersion, error)
	GetWorkflowVersionByID(ctx context.Context, id int32) (db.WorkflowVersion, error)
	GetMaxVersionNumberByWorkflowID(ctx context.Context, workflowID int32) (int32, error)
	ListWorkflowVersionsByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error)
	ListAllWorkflowVersionsByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error)
	ListWorkflowVersionSummariesByWorkflowID(ctx context.Context, workflowID int32) ([]db.ListWorkflowVersionSummariesByWorkflowIDRow, error)
	ListActiveWorkflowVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error)
	UpdateWorkflowVersionValid(ctx context.Context, arg db.UpdateWorkflowVersionValidParams) (db.WorkflowVersion, error)
	UpdateWorkflowVersionActive(ctx context.Context, arg db.UpdateWorkflowVersionActiveParams) (db.WorkflowVersion, error)
	UpdateWorkflowVersionName(ctx context.Context, arg db.UpdateWorkflowVersionNameParams) (db.WorkflowVersion, error)
	InvalidateWorkflowVersionsByWorkflowID(ctx context.Context, workflowID int32) error
	SoftDeleteWorkflowVersion(ctx context.Context, id int32) error

	// Workflow input schema
	CreateWorkflowInputSchema(ctx context.Context, arg db.CreateWorkflowInputSchemaParams) (db.WorkflowInputSchema, error)
	UpdateWorkflowInputSchemaStatus(ctx context.Context, arg db.UpdateWorkflowInputSchemaStatusParams) (db.WorkflowInputSchema, error)
	CreateWorkflowVersionInputSchemaCompatibility(ctx context.Context, arg db.CreateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error)
	GetNativeInputSchemaForWorkflowVersion(ctx context.Context, workflowVersionID int32) (db.WorkflowInputSchema, error)

	// WorkflowStep
	CreateWorkflowStep(ctx context.Context, arg db.CreateWorkflowStepParams) (db.WorkflowStep, error)
	GetWorkflowStepByID(ctx context.Context, id int32) (db.WorkflowStep, error)
	ListWorkflowStepsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStep, error)
	UpdateWorkflowStep(ctx context.Context, arg db.UpdateWorkflowStepParams) (db.WorkflowStep, error)
	DeleteWorkflowStepsByVersionID(ctx context.Context, workflowVersionID int32) error
	SoftDeleteWorkflowStep(ctx context.Context, id int32) error

	// Worker settings (for auto-resolving revision in CreateStep)
	ListWorkerSettingsSchemasByWorkTypeID(ctx context.Context, workTypeID int32) ([]db.WorkerSettingsSchema, error)
	GetWorkerSettingsSchemaByID(ctx context.Context, id int32) (db.WorkerSettingsSchema, error)
	ListWorkerSettingsRevisionsBySchemaID(ctx context.Context, schemaID int32) ([]db.WorkerSettingsRevision, error)
	CreateWorkerSettingsRevision(ctx context.Context, arg db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error)
	CloneWorkerSettingsRevision(ctx context.Context, arg db.CloneWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error)
	UpdateWorkerSettingsRevisionSettings(ctx context.Context, arg db.UpdateWorkerSettingsRevisionSettingsParams) (db.WorkerSettingsRevision, error)

	// WorkflowStep (extended)
	UpdateWorkflowStepPosition(ctx context.Context, arg db.UpdateWorkflowStepPositionParams) error
	ListEnrichedStepsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.ListEnrichedStepsByVersionIDRow, error)

	// WorkflowStepDependency
	CreateWorkflowStepDependency(ctx context.Context, arg db.CreateWorkflowStepDependencyParams) error
	ListDependenciesByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStepDependency, error)
	DeleteDependenciesByStepID(ctx context.Context, stepID int32) error
	DeleteWorkflowStepDependency(ctx context.Context, arg db.DeleteWorkflowStepDependencyParams) error
}
