package worktype

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type RegistrationTx interface {
	GetActiveWorkerBootstrapTokenByHashForUpdate(ctx context.Context, tokenHash string) (db.WorkerBootstrapToken, error)
	GetWorkerSettingsSchemaByVersion(ctx context.Context, arg db.GetWorkerSettingsSchemaByVersionParams) (db.WorkerSettingsSchema, error)
	CreateWorkerSettingsSchema(ctx context.Context, arg db.CreateWorkerSettingsSchemaParams) (db.WorkerSettingsSchema, error)
	ListWorkerSettingsRevisionsBySchemaID(ctx context.Context, workerSettingsSchemaID int32) ([]db.WorkerSettingsRevision, error)
	CreateWorkerSettingsRevision(ctx context.Context, arg db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error)
	GetWorkerByOrgWorkTypeAndName(ctx context.Context, arg db.GetWorkerByOrgWorkTypeAndNameParams) (db.Worker, error)
	CreateNewWorker(ctx context.Context, arg db.CreateNewWorkerParams) (db.Worker, error)
	UpdateNewWorkerSchema(ctx context.Context, arg db.UpdateNewWorkerSchemaParams) (db.Worker, error)
	CountActiveWorkersByBootstrapTokenExcludingWorker(ctx context.Context, arg db.CountActiveWorkersByBootstrapTokenExcludingWorkerParams) (int32, error)
	UpdateNewWorkerHeartbeat(ctx context.Context, id int32) error
	TouchWorkerBootstrapTokenUse(ctx context.Context, id int32) (db.WorkerBootstrapToken, error)
	CreateWorkerNATSSession(ctx context.Context, arg db.CreateWorkerNATSSessionParams) (db.WorkerNatsSession, error)
}

// Storage — интерфейс хранилища для worktype домена.
// Реализуется *store.Store через встроенный *db.Queries.
type Storage interface {
	// Work Type CRUD
	CreateWorkType(ctx context.Context, arg db.CreateWorkTypeParams) (db.WorkType, error)
	GetWorkTypeByID(ctx context.Context, id int32) (db.WorkType, error)
	ListWorkTypes(ctx context.Context) ([]db.WorkType, error)
	SoftDeleteWorkType(ctx context.Context, id int32) error

	// Worker Bootstrap Token
	CreateWorkerBootstrapToken(ctx context.Context, arg db.CreateWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error)
	GetActiveWorkerBootstrapTokenByHash(ctx context.Context, tokenHash string) (db.WorkerBootstrapToken, error)
	ListWorkerBootstrapTokensByOrganization(ctx context.Context, organizationID int32) ([]db.ListWorkerBootstrapTokensByOrganizationRow, error)
	TouchWorkerBootstrapTokenUse(ctx context.Context, id int32) (db.WorkerBootstrapToken, error)
	RevokeWorkerBootstrapToken(ctx context.Context, arg db.RevokeWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error)
	WithRegistrationTx(ctx context.Context, fn func(RegistrationTx) error) error

	// Worker
	CreateNewWorker(ctx context.Context, arg db.CreateNewWorkerParams) (db.Worker, error)
	GetNewWorkerByID(ctx context.Context, id int32) (db.Worker, error)
	GetWorkerByOrgWorkTypeAndName(ctx context.Context, arg db.GetWorkerByOrgWorkTypeAndNameParams) (db.Worker, error)
	ListNewWorkersByWorkTypeID(ctx context.Context, workTypeID int32) ([]db.Worker, error)
	ListNewWorkersByOrganizationID(ctx context.Context, organizationID int32) ([]db.Worker, error)
	UpdateNewWorkerHeartbeat(ctx context.Context, id int32) error
	UpdateNewWorkerSchema(ctx context.Context, arg db.UpdateNewWorkerSchemaParams) (db.Worker, error)
	DeleteWorker(ctx context.Context, id int32) error
	ListWorkflowUsagesByWorkerID(ctx context.Context, id int32) ([]db.ListWorkflowUsagesByWorkerIDRow, error)

	// Worker NATS Session
	CreateWorkerNATSSession(ctx context.Context, arg db.CreateWorkerNATSSessionParams) (db.WorkerNatsSession, error)
	ListActiveWorkerNATSSessionsByBootstrapToken(ctx context.Context, bootstrapTokenID int32) ([]db.WorkerNatsSession, error)
	RevokeWorkerNATSSession(ctx context.Context, arg db.RevokeWorkerNATSSessionParams) (db.WorkerNatsSession, error)
	RevokeWorkerNATSSessionsByBootstrapToken(ctx context.Context, arg db.RevokeWorkerNATSSessionsByBootstrapTokenParams) ([]db.WorkerNatsSession, error)

	// Worker Settings Schema
	CreateWorkerSettingsSchema(ctx context.Context, arg db.CreateWorkerSettingsSchemaParams) (db.WorkerSettingsSchema, error)
	GetWorkerSettingsSchemaByID(ctx context.Context, id int32) (db.WorkerSettingsSchema, error)
	GetWorkerSettingsSchemaByVersion(ctx context.Context, arg db.GetWorkerSettingsSchemaByVersionParams) (db.WorkerSettingsSchema, error)
	ListWorkerSettingsSchemasByWorkTypeID(ctx context.Context, workTypeID int32) ([]db.WorkerSettingsSchema, error)

	// Worker Settings Revision
	CreateWorkerSettingsRevision(ctx context.Context, arg db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error)
	GetWorkerSettingsRevisionByID(ctx context.Context, id int32) (db.WorkerSettingsRevision, error)
	ListWorkerSettingsRevisionsBySchemaID(ctx context.Context, workerSettingsSchemaID int32) ([]db.WorkerSettingsRevision, error)

	// System CRUD
	CreateSystem(ctx context.Context, arg db.CreateSystemParams) (db.System, error)
	GetSystemByID(ctx context.Context, id int32) (db.System, error)
	CountSystemsByOrganizationID(ctx context.Context, organizationID pgtype.Int4) (int64, error)
	ListSystemsByOrganizationID(ctx context.Context, organizationID pgtype.Int4) ([]db.ListSystemsByOrganizationIDRow, error)
	ListSystemsByOrganizationIDPaginated(ctx context.Context, arg db.ListSystemsByOrganizationIDPaginatedParams) ([]db.ListSystemsByOrganizationIDPaginatedRow, error)
	UpdateSystem(ctx context.Context, arg db.UpdateSystemParams) (db.System, error)
	SoftDeleteSystem(ctx context.Context, id int32) error

	// System Token
	CreateSystemToken(ctx context.Context, arg db.CreateSystemTokenParams) (db.SystemToken, error)
	GetSystemTokenByID(ctx context.Context, id int32) (db.SystemToken, error)
	GetSystemTokenByPublicToken(ctx context.Context, publicToken string) (db.SystemToken, error)
	ListSystemTokensBySystemID(ctx context.Context, systemID int32) ([]db.SystemToken, error)
	DeactivateSystemToken(ctx context.Context, id int32) error
	ActivateSystemToken(ctx context.Context, id int32) error

	// Workflow Token
	GetWorkflowByID(ctx context.Context, id int32) (db.Workflow, error)
	GrantWorkflowToken(ctx context.Context, arg db.GrantWorkflowTokenParams) error
	RevokeWorkflowToken(ctx context.Context, arg db.RevokeWorkflowTokenParams) error
	ListWorkflowTokensBySystemTokenID(ctx context.Context, systemTokenID int32) ([]db.WorkflowToken, error)
	ListWorkflowTokensByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowToken, error)
}
