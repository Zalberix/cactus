package message

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// Storage — интерфейс хранилища для message domain.
type Storage interface {
	// Message
	CreateNewMessage(ctx context.Context, arg db.CreateNewMessageParams) (db.Message, error)

	// Workflow (для получения активной версии и валидации)
	GetWorkflowByID(ctx context.Context, id int32) (db.Workflow, error)
	ListActiveWorkflowVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error)

	// Workflow steps и dependencies (для формирования DAGInput)
	ListWorkflowStepsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStep, error)
	ListEnrichedStepsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.ListEnrichedStepsByVersionIDRow, error)
	ListDependenciesByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStepDependency, error)
	GetWorkerSettingsRevisionByID(ctx context.Context, id int32) (db.WorkerSettingsRevision, error)

	// Workflow run
	CreateWorkflowRun(ctx context.Context, arg db.CreateWorkflowRunParams) (db.WorkflowRun, error)
	UpdateWorkflowRunStarted(ctx context.Context, id int32) (db.WorkflowRun, error)

	// Workflow token access check
	CheckWorkflowAccess(ctx context.Context, arg db.CheckWorkflowAccessParams) (bool, error)

	// Status API (per EXEC-09)
	GetMessageStatusByID(ctx context.Context, id int32) (db.GetMessageStatusByIDRow, error)
	ListWorkflowRunStepStatusesByRunID(ctx context.Context, workflowRunID int32) ([]db.ListWorkflowRunStepStatusesByRunIDRow, error)

	// Message detail API
	GetMessageDetailByID(ctx context.Context, id int32) (db.GetMessageDetailByIDRow, error)
	ListWorkflowRunStepDetailsByRunID(ctx context.Context, workflowRunID int32) ([]db.WorkflowRunStep, error)

	// Message listing (per UI-12)
	ListMessagesByOrganizationID(ctx context.Context, arg db.ListMessagesByOrganizationIDParams) ([]db.ListMessagesByOrganizationIDRow, error)
	CountMessagesByOrganizationID(ctx context.Context, organizationID pgtype.Int4) (int64, error)
}
