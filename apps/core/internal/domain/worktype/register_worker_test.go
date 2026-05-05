package worktype

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type registerWorkerStore struct {
	token            db.WorkTypeToken
	createdSchemaArg db.CreateWorkerSettingsSchemaParams
}

func (s *registerWorkerStore) CreateWorkType(context.Context, db.CreateWorkTypeParams) (db.WorkType, error) {
	return db.WorkType{}, nil
}
func (s *registerWorkerStore) GetWorkTypeByID(context.Context, int32) (db.WorkType, error) {
	return db.WorkType{}, nil
}
func (s *registerWorkerStore) ListWorkTypes(context.Context) ([]db.WorkType, error) { return nil, nil }
func (s *registerWorkerStore) SoftDeleteWorkType(context.Context, int32) error      { return nil }
func (s *registerWorkerStore) CreateWorkTypeToken(context.Context, db.CreateWorkTypeTokenParams) (db.WorkTypeToken, error) {
	return db.WorkTypeToken{}, nil
}
func (s *registerWorkerStore) GetActiveWorkTypeTokenByHash(context.Context, string) (db.WorkTypeToken, error) {
	return s.token, nil
}
func (s *registerWorkerStore) CreateNewWorker(_ context.Context, arg db.CreateNewWorkerParams) (db.Worker, error) {
	return db.Worker{ID: 77, WorkTypeID: arg.WorkTypeID, WorkerSettingsSchemaID: arg.WorkerSettingsSchemaID, Name: arg.Name}, nil
}
func (s *registerWorkerStore) GetNewWorkerByID(context.Context, int32) (db.Worker, error) {
	return db.Worker{}, nil
}
func (s *registerWorkerStore) GetWorkerByWorkTypeAndName(context.Context, db.GetWorkerByWorkTypeAndNameParams) (db.Worker, error) {
	return db.Worker{}, pgx.ErrNoRows
}
func (s *registerWorkerStore) ListNewWorkersByWorkTypeID(context.Context, int32) ([]db.Worker, error) {
	return nil, nil
}
func (s *registerWorkerStore) UpdateNewWorkerHeartbeat(context.Context, int32) error { return nil }
func (s *registerWorkerStore) UpdateNewWorkerSchema(context.Context, db.UpdateNewWorkerSchemaParams) (db.Worker, error) {
	return db.Worker{}, nil
}
func (s *registerWorkerStore) CreateWorkerSettingsSchema(_ context.Context, arg db.CreateWorkerSettingsSchemaParams) (db.WorkerSettingsSchema, error) {
	s.createdSchemaArg = arg
	return db.WorkerSettingsSchema{ID: 55, WorkTypeID: arg.WorkTypeID, SettingsSchema: arg.SettingsSchema, InputSchema: arg.InputSchema, OutputSchema: arg.OutputSchema}, nil
}
func (s *registerWorkerStore) GetWorkerSettingsSchemaByID(context.Context, int32) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, nil
}
func (s *registerWorkerStore) GetWorkerSettingsSchemaByVersion(context.Context, db.GetWorkerSettingsSchemaByVersionParams) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, pgx.ErrNoRows
}
func (s *registerWorkerStore) ListWorkerSettingsSchemasByWorkTypeID(context.Context, int32) ([]db.WorkerSettingsSchema, error) {
	return nil, nil
}
func (s *registerWorkerStore) CreateWorkerSettingsRevision(context.Context, db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}
func (s *registerWorkerStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}
func (s *registerWorkerStore) ListWorkerSettingsRevisionsBySchemaID(context.Context, int32) ([]db.WorkerSettingsRevision, error) {
	return nil, nil
}
func (s *registerWorkerStore) CreateSystem(context.Context, db.CreateSystemParams) (db.System, error) {
	return db.System{}, nil
}
func (s *registerWorkerStore) GetSystemByID(context.Context, int32) (db.System, error) {
	return db.System{}, nil
}
func (s *registerWorkerStore) CountSystemsByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 0, nil
}
func (s *registerWorkerStore) ListSystemsByOrganizationID(context.Context, pgtype.Int4) ([]db.ListSystemsByOrganizationIDRow, error) {
	return nil, nil
}
func (s *registerWorkerStore) ListSystemsByOrganizationIDPaginated(context.Context, db.ListSystemsByOrganizationIDPaginatedParams) ([]db.ListSystemsByOrganizationIDRow, error) {
	return nil, nil
}
func (s *registerWorkerStore) UpdateSystem(context.Context, db.UpdateSystemParams) (db.System, error) {
	return db.System{}, nil
}
func (s *registerWorkerStore) SoftDeleteSystem(context.Context, int32) error { return nil }
func (s *registerWorkerStore) CreateSystemToken(context.Context, db.CreateSystemTokenParams) (db.SystemToken, error) {
	return db.SystemToken{}, nil
}
func (s *registerWorkerStore) GetSystemTokenByID(context.Context, int32) (db.SystemToken, error) {
	return db.SystemToken{}, nil
}
func (s *registerWorkerStore) GetSystemTokenByPublicToken(context.Context, string) (db.SystemToken, error) {
	return db.SystemToken{}, nil
}
func (s *registerWorkerStore) ListSystemTokensBySystemID(context.Context, int32) ([]db.SystemToken, error) {
	return nil, nil
}
func (s *registerWorkerStore) DeactivateSystemToken(context.Context, int32) error { return nil }
func (s *registerWorkerStore) ActivateSystemToken(context.Context, int32) error   { return nil }
func (s *registerWorkerStore) GetWorkflowByID(context.Context, int32) (db.Workflow, error) {
	return db.Workflow{}, nil
}
func (s *registerWorkerStore) GrantWorkflowToken(context.Context, db.GrantWorkflowTokenParams) error {
	return nil
}
func (s *registerWorkerStore) RevokeWorkflowToken(context.Context, db.RevokeWorkflowTokenParams) error {
	return nil
}
func (s *registerWorkerStore) ListWorkflowTokensBySystemTokenID(context.Context, int32) ([]db.WorkflowToken, error) {
	return nil, nil
}
func (s *registerWorkerStore) ListWorkflowTokensByWorkflowID(context.Context, int32) ([]db.WorkflowToken, error) {
	return nil, nil
}

func TestRegisterWorkerSplitsManifestSchemasIntoColumns(t *testing.T) {
	manifest := json.RawMessage(`{
		"kind": "smtp",
		"name_kind": "SMTP Email",
		"type": "email",
		"name_type": "Email Delivery",
		"settings_schema": {"type":"object","properties":{"host":{"type":"string","required":true}}},
		"input_schema": {"type":"object","properties":{"to":{"type":"string","required":true}}},
		"output_schema": {"type":"object","properties":{"message_id":{"type":"string"}}}
	}`)
	store := &registerWorkerStore{token: db.WorkTypeToken{WorkTypeID: 12}}
	service := NewService(store)

	worker, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       manifest,
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if worker.ID != 77 {
		t.Fatalf("expected worker ID 77, got %d", worker.ID)
	}
	assertJSONEqual(t, `{"type":"object","properties":{"host":{"type":"string","required":true}}}`, store.createdSchemaArg.SettingsSchema)
	assertJSONEqual(t, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, store.createdSchemaArg.InputSchema)
	assertJSONEqual(t, `{"type":"object","properties":{"message_id":{"type":"string"}}}`, store.createdSchemaArg.OutputSchema)
}

func assertJSONEqual(t *testing.T, want string, got []byte) {
	t.Helper()
	var wantAny any
	if err := json.Unmarshal([]byte(want), &wantAny); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	var gotAny any
	if err := json.Unmarshal(got, &gotAny); err != nil {
		t.Fatalf("unmarshal got %s: %v", got, err)
	}
	wantJSON, _ := json.Marshal(wantAny)
	gotJSON, _ := json.Marshal(gotAny)
	if string(wantJSON) != string(gotJSON) {
		t.Fatalf("json mismatch\nwant: %s\ngot:  %s", wantJSON, gotJSON)
	}
}
