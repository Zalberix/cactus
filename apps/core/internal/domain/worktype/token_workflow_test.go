package worktype

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/natsauth"
	"github.com/zalberix/cactus/apps/core/storage/db"
)

type tokenWorkflowStore struct {
	token       db.SystemToken
	workflow    db.Workflow
	grantArg    db.GrantWorkflowTokenParams
	revokeArg   db.RevokeWorkflowTokenParams
	grantCalled bool
}

func (s *tokenWorkflowStore) CreateWorkType(context.Context, db.CreateWorkTypeParams) (db.WorkType, error) {
	return db.WorkType{}, nil
}

func (s *tokenWorkflowStore) GetWorkTypeByID(context.Context, int32) (db.WorkType, error) {
	return db.WorkType{}, nil
}
func (s *tokenWorkflowStore) ListWorkTypes(context.Context) ([]db.WorkType, error) { return nil, nil }
func (s *tokenWorkflowStore) SoftDeleteWorkType(context.Context, int32) error      { return nil }

func (s *tokenWorkflowStore) CreateWorkerBootstrapToken(context.Context, db.CreateWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *tokenWorkflowStore) GetActiveWorkerBootstrapTokenByHash(context.Context, string) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *tokenWorkflowStore) GetActiveWorkerBootstrapTokenByHashForUpdate(context.Context, string) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *tokenWorkflowStore) ListWorkerBootstrapTokensByOrganization(context.Context, int32) ([]db.ListWorkerBootstrapTokensByOrganizationRow, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) CountActiveWorkersByBootstrapTokenExcludingWorker(context.Context, db.CountActiveWorkersByBootstrapTokenExcludingWorkerParams) (int32, error) {
	return 0, nil
}

func (s *tokenWorkflowStore) TouchWorkerBootstrapTokenUse(context.Context, int32) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *tokenWorkflowStore) GetActiveWorkerNATSSessionByWorkerAndBootstrapTokenForUpdate(context.Context, db.GetActiveWorkerNATSSessionByWorkerAndBootstrapTokenForUpdateParams) (db.WorkerNatsSession, error) {
	return db.WorkerNatsSession{}, nil
}

func (s *tokenWorkflowStore) ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapToken(context.Context, db.ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapTokenParams) (db.WorkerNatsSession, error) {
	return db.WorkerNatsSession{}, nil
}

func (s *tokenWorkflowStore) WithRegistrationTx(_ context.Context, fn func(RegistrationTx) error) error {
	return fn(s)
}

func (s *tokenWorkflowStore) RevokeWorkerBootstrapToken(context.Context, db.RevokeWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *tokenWorkflowStore) CreateNewWorker(context.Context, db.CreateNewWorkerParams) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *tokenWorkflowStore) GetNewWorkerByID(context.Context, int32) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *tokenWorkflowStore) GetWorkerByOrgWorkTypeAndName(context.Context, db.GetWorkerByOrgWorkTypeAndNameParams) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *tokenWorkflowStore) ListNewWorkersByWorkTypeID(context.Context, int32) ([]db.Worker, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) ListNewWorkersByOrganizationID(context.Context, int32) ([]db.Worker, error) {
	return nil, nil
}
func (s *tokenWorkflowStore) UpdateNewWorkerHeartbeat(context.Context, int32) error { return nil }
func (s *tokenWorkflowStore) UpdateNewWorkerSchema(context.Context, db.UpdateNewWorkerSchemaParams) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *tokenWorkflowStore) DeleteWorker(context.Context, int32) error { return nil }

func (s *tokenWorkflowStore) ListWorkflowUsagesByWorkerID(context.Context, int32) ([]db.ListWorkflowUsagesByWorkerIDRow, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) CreateWorkerNATSSession(context.Context, db.CreateWorkerNATSSessionParams) (db.WorkerNatsSession, error) {
	return db.WorkerNatsSession{}, nil
}

func (s *tokenWorkflowStore) ListActiveWorkerNATSSessionsByBootstrapToken(context.Context, int32) ([]db.WorkerNatsSession, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) RevokeWorkerNATSSession(context.Context, db.RevokeWorkerNATSSessionParams) (db.WorkerNatsSession, error) {
	return db.WorkerNatsSession{}, nil
}

func (s *tokenWorkflowStore) RevokeWorkerNATSSessionsByBootstrapToken(context.Context, db.RevokeWorkerNATSSessionsByBootstrapTokenParams) ([]db.WorkerNatsSession, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) CreateWorkerSettingsSchema(context.Context, db.CreateWorkerSettingsSchemaParams) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, nil
}

func (s *tokenWorkflowStore) GetWorkerSettingsSchemaByID(context.Context, int32) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, nil
}

func (s *tokenWorkflowStore) GetWorkerSettingsSchemaByVersion(context.Context, db.GetWorkerSettingsSchemaByVersionParams) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, nil
}

func (s *tokenWorkflowStore) ListWorkerSettingsSchemasByWorkTypeID(context.Context, int32) ([]db.WorkerSettingsSchema, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) CreateWorkerSettingsRevision(context.Context, db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (s *tokenWorkflowStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (s *tokenWorkflowStore) ListWorkerSettingsRevisionsBySchemaID(context.Context, int32) ([]db.WorkerSettingsRevision, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) CreateSystem(context.Context, db.CreateSystemParams) (db.System, error) {
	return db.System{}, nil
}

func (s *tokenWorkflowStore) GetSystemByID(context.Context, int32) (db.System, error) {
	return db.System{}, nil
}

func (s *tokenWorkflowStore) CountSystemsByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 0, nil
}

func (s *tokenWorkflowStore) ListSystemsByOrganizationID(context.Context, pgtype.Int4) ([]db.ListSystemsByOrganizationIDRow, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) ListSystemsByOrganizationIDPaginated(context.Context, db.ListSystemsByOrganizationIDPaginatedParams) ([]db.ListSystemsByOrganizationIDPaginatedRow, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) UpdateSystem(context.Context, db.UpdateSystemParams) (db.System, error) {
	return db.System{}, nil
}
func (s *tokenWorkflowStore) SoftDeleteSystem(context.Context, int32) error { return nil }
func (s *tokenWorkflowStore) CreateSystemToken(context.Context, db.CreateSystemTokenParams) (db.SystemToken, error) {
	return db.SystemToken{}, nil
}

func (s *tokenWorkflowStore) GetSystemTokenByID(_ context.Context, id int32) (db.SystemToken, error) {
	s.token.ID = id
	return s.token, nil
}

func (s *tokenWorkflowStore) GetSystemTokenByPublicToken(context.Context, string) (db.SystemToken, error) {
	return db.SystemToken{}, nil
}

func (s *tokenWorkflowStore) ListSystemTokensBySystemID(context.Context, int32) ([]db.SystemToken, error) {
	return nil, nil
}
func (s *tokenWorkflowStore) DeactivateSystemToken(context.Context, int32) error { return nil }
func (s *tokenWorkflowStore) ActivateSystemToken(context.Context, int32) error   { return nil }
func (s *tokenWorkflowStore) GetWorkflowByID(_ context.Context, id int32) (db.Workflow, error) {
	s.workflow.ID = id
	return s.workflow, nil
}

func (s *tokenWorkflowStore) GrantWorkflowToken(_ context.Context, arg db.GrantWorkflowTokenParams) error {
	s.grantCalled = true
	s.grantArg = arg
	return nil
}

func (s *tokenWorkflowStore) RevokeWorkflowToken(_ context.Context, arg db.RevokeWorkflowTokenParams) error {
	s.revokeArg = arg
	return nil
}

func (s *tokenWorkflowStore) ListWorkflowTokensBySystemTokenID(context.Context, int32) ([]db.WorkflowToken, error) {
	return nil, nil
}

func (s *tokenWorkflowStore) ListWorkflowTokensByWorkflowID(context.Context, int32) ([]db.WorkflowToken, error) {
	return nil, nil
}

func TestBindWorkflowToTokenRejectsDifferentSystems(t *testing.T) {
	store := &tokenWorkflowStore{
		token:    db.SystemToken{SystemID: 10},
		workflow: db.Workflow{SystemID: 20},
	}
	service := NewService(store, natsauth.NoopManager{})

	err := service.BindWorkflowToToken(context.Background(), 1, 2)

	if !errors.Is(err, ErrWorkflowTokenSystemMismatch) {
		t.Fatalf("expected ErrWorkflowTokenSystemMismatch, got %v", err)
	}
	if store.grantCalled {
		t.Fatal("GrantWorkflowToken should not be called for different systems")
	}
}

func TestBindWorkflowToTokenGrantsWhenSystemsMatch(t *testing.T) {
	store := &tokenWorkflowStore{
		token:    db.SystemToken{SystemID: 10},
		workflow: db.Workflow{SystemID: 10},
	}
	service := NewService(store, natsauth.NoopManager{})

	err := service.BindWorkflowToToken(context.Background(), 7, 8)
	if err != nil {
		t.Fatalf("BindWorkflowToToken error: %v", err)
	}
	if store.grantArg.SystemTokenID != 7 || store.grantArg.WorkflowID != 8 {
		t.Fatalf("unexpected grant args: %+v", store.grantArg)
	}
}

func TestListTokenWorkflowsReturnsEmptySliceWhenNoLinks(t *testing.T) {
	service := NewService(&tokenWorkflowStore{}, natsauth.NoopManager{})

	links, err := service.ListTokenWorkflows(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListTokenWorkflows error: %v", err)
	}
	if links == nil {
		t.Fatal("expected empty slice, got nil")
	}
}

func TestListWorkflowTokensReturnsEmptySliceWhenNoLinks(t *testing.T) {
	service := NewService(&tokenWorkflowStore{}, natsauth.NoopManager{})

	links, err := service.ListWorkflowTokens(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListWorkflowTokens error: %v", err)
	}
	if links == nil {
		t.Fatal("expected empty slice, got nil")
	}
}
