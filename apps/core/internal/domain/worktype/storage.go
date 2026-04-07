package worktype

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// Storage — интерфейс хранилища для worktype домена.
// Реализуется *store.Store через встроенный *db.Queries.
type Storage interface {
	// Work Type CRUD
	CreateWorkType(ctx context.Context, arg db.CreateWorkTypeParams) (db.WorkType, error)
	GetWorkTypeByID(ctx context.Context, id int32) (db.WorkType, error)
	ListWorkTypes(ctx context.Context) ([]db.WorkType, error)
	SoftDeleteWorkType(ctx context.Context, id int32) error

	// Work Type Token
	CreateWorkTypeToken(ctx context.Context, arg db.CreateWorkTypeTokenParams) (db.WorkTypeToken, error)
	GetActiveWorkTypeTokenByHash(ctx context.Context, tokenHash string) (db.WorkTypeToken, error)

	// Worker
	CreateNewWorker(ctx context.Context, arg db.CreateNewWorkerParams) (db.Worker, error)
	GetNewWorkerByID(ctx context.Context, id int32) (db.Worker, error)
	GetWorkerByWorkTypeAndName(ctx context.Context, arg db.GetWorkerByWorkTypeAndNameParams) (db.Worker, error)
	ListNewWorkersByWorkTypeID(ctx context.Context, workTypeID int32) ([]db.Worker, error)
	UpdateNewWorkerHeartbeat(ctx context.Context, id int32) error
	UpdateNewWorkerSchema(ctx context.Context, arg db.UpdateNewWorkerSchemaParams) (db.Worker, error)

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
	ListSystemsByOrganizationID(ctx context.Context, organizationID pgtype.Int4) ([]db.ListSystemsByOrganizationIDRow, error)
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
	GrantWorkflowToken(ctx context.Context, arg db.GrantWorkflowTokenParams) error
	RevokeWorkflowToken(ctx context.Context, arg db.RevokeWorkflowTokenParams) error
}
