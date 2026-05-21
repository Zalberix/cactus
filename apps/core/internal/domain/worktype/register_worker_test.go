package worktype

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/natsauth"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type registerWorkerStore struct {
	bootstrapToken         db.WorkerBootstrapToken
	revisions              []db.WorkerSettingsRevision
	existingWorker         db.Worker
	activeNATSSessions     []db.WorkerNatsSession
	latestNATSSession      db.WorkerNatsSession
	activeWorkerCount      int32
	settingsSchema         db.WorkerSettingsSchema
	createdSchemaArg       db.CreateWorkerSettingsSchemaParams
	createdRevision        db.CreateWorkerSettingsRevisionParams
	createdWorkerArg       db.CreateNewWorkerParams
	createdNATSSessionArg  db.CreateWorkerNATSSessionParams
	replacedNATSSessionArg db.ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapTokenParams
	touchedTokenID         int32
}

func (s *registerWorkerStore) CreateWorkType(context.Context, db.CreateWorkTypeParams) (db.WorkType, error) {
	return db.WorkType{}, nil
}

func (s *registerWorkerStore) GetWorkTypeByID(context.Context, int32) (db.WorkType, error) {
	return db.WorkType{}, nil
}
func (s *registerWorkerStore) ListWorkTypes(context.Context) ([]db.WorkType, error) { return nil, nil }
func (s *registerWorkerStore) SoftDeleteWorkType(context.Context, int32) error      { return nil }

func (s *registerWorkerStore) CreateWorkerBootstrapToken(context.Context, db.CreateWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *registerWorkerStore) GetActiveWorkerBootstrapTokenByHash(context.Context, string) (db.WorkerBootstrapToken, error) {
	if s.bootstrapToken.ID == 0 {
		return db.WorkerBootstrapToken{}, pgx.ErrNoRows
	}
	return s.bootstrapToken, nil
}

func (s *registerWorkerStore) GetActiveWorkerBootstrapTokenByHashForUpdate(context.Context, string) (db.WorkerBootstrapToken, error) {
	return s.GetActiveWorkerBootstrapTokenByHash(context.Background(), "")
}

func (s *registerWorkerStore) ListWorkerBootstrapTokensByOrganization(context.Context, int32) ([]db.ListWorkerBootstrapTokensByOrganizationRow, error) {
	return nil, nil
}

func (s *registerWorkerStore) CountActiveWorkersByBootstrapTokenExcludingWorker(context.Context, db.CountActiveWorkersByBootstrapTokenExcludingWorkerParams) (int32, error) {
	return s.activeWorkerCount, nil
}

func (s *registerWorkerStore) TouchWorkerBootstrapTokenUse(context.Context, int32) (db.WorkerBootstrapToken, error) {
	s.touchedTokenID = s.bootstrapToken.ID
	return s.bootstrapToken, nil
}

func (s *registerWorkerStore) GetActiveWorkerNATSSessionByWorkerAndBootstrapTokenForUpdate(_ context.Context, arg db.GetActiveWorkerNATSSessionByWorkerAndBootstrapTokenForUpdateParams) (db.WorkerNatsSession, error) {
	for _, session := range s.activeNATSSessions {
		if session.WorkerID == arg.WorkerID && session.BootstrapTokenID == arg.BootstrapTokenID {
			return session, nil
		}
	}
	return db.WorkerNatsSession{}, pgx.ErrNoRows
}

func (s *registerWorkerStore) ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapToken(_ context.Context, arg db.ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapTokenParams) (db.WorkerNatsSession, error) {
	s.replacedNATSSessionArg = arg
	session := s.latestNATSSession
	if session.ID == 0 && len(s.activeNATSSessions) > 0 {
		session = s.activeNATSSessions[0]
	}
	if session.ID == 0 {
		return db.WorkerNatsSession{}, pgx.ErrNoRows
	}
	session.WorkerID = arg.WorkerID
	session.BootstrapTokenID = arg.BootstrapTokenID
	session.NatsAccountPublicKey = arg.NatsAccountPublicKey
	session.NatsUserPublicKey = arg.NatsUserPublicKey
	session.NatsUserJwt = arg.NatsUserJwt
	session.NatsUserSeed = arg.NatsUserSeed
	session.NatsUserCredentials = arg.NatsUserCredentials
	session.Permissions = arg.Permissions
	session.RevokedAt = pgtype.Timestamp{}
	session.RevokedByUserID = pgtype.Int4{}
	return session, nil
}

func (s *registerWorkerStore) WithRegistrationTx(_ context.Context, fn func(RegistrationTx) error) error {
	return fn(s)
}

func (s *registerWorkerStore) RevokeWorkerBootstrapToken(context.Context, db.RevokeWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	return db.WorkerBootstrapToken{}, nil
}

func (s *registerWorkerStore) CreateNewWorker(_ context.Context, arg db.CreateNewWorkerParams) (db.Worker, error) {
	s.createdWorkerArg = arg
	return db.Worker{ID: 77, OrganizationID: arg.OrganizationID, WorkTypeID: arg.WorkTypeID, WorkerSettingsSchemaID: arg.WorkerSettingsSchemaID, Name: arg.Name}, nil
}

func (s *registerWorkerStore) GetNewWorkerByID(context.Context, int32) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *registerWorkerStore) GetWorkerByOrgWorkTypeAndName(context.Context, db.GetWorkerByOrgWorkTypeAndNameParams) (db.Worker, error) {
	if s.existingWorker.ID != 0 {
		return s.existingWorker, nil
	}
	return db.Worker{}, pgx.ErrNoRows
}

func (s *registerWorkerStore) ListNewWorkersByWorkTypeID(context.Context, int32) ([]db.Worker, error) {
	return nil, nil
}

func (s *registerWorkerStore) ListNewWorkersByOrganizationID(context.Context, int32) ([]db.Worker, error) {
	return nil, nil
}
func (s *registerWorkerStore) UpdateNewWorkerHeartbeat(context.Context, int32) error { return nil }
func (s *registerWorkerStore) UpdateNewWorkerSchema(context.Context, db.UpdateNewWorkerSchemaParams) (db.Worker, error) {
	return db.Worker{}, nil
}

func (s *registerWorkerStore) DeleteWorker(context.Context, int32) error { return nil }

func (s *registerWorkerStore) ListWorkflowUsagesByWorkerID(context.Context, int32) ([]db.ListWorkflowUsagesByWorkerIDRow, error) {
	return nil, nil
}

func (s *registerWorkerStore) CreateWorkerNATSSession(_ context.Context, arg db.CreateWorkerNATSSessionParams) (db.WorkerNatsSession, error) {
	s.createdNATSSessionArg = arg
	return db.WorkerNatsSession{
		ID:                   1,
		WorkerID:             arg.WorkerID,
		BootstrapTokenID:     arg.BootstrapTokenID,
		NatsAccountPublicKey: arg.NatsAccountPublicKey,
		NatsUserPublicKey:    arg.NatsUserPublicKey,
		NatsUserJwt:          arg.NatsUserJwt,
		NatsUserSeed:         arg.NatsUserSeed,
		NatsUserCredentials:  arg.NatsUserCredentials,
		Permissions:          arg.Permissions,
	}, nil
}

func (s *registerWorkerStore) ListActiveWorkerNATSSessionsByBootstrapToken(context.Context, int32) ([]db.WorkerNatsSession, error) {
	return s.activeNATSSessions, nil
}

func (s *registerWorkerStore) RevokeWorkerNATSSession(context.Context, db.RevokeWorkerNATSSessionParams) (db.WorkerNatsSession, error) {
	return db.WorkerNatsSession{}, nil
}

func (s *registerWorkerStore) RevokeWorkerNATSSessionsByBootstrapToken(context.Context, db.RevokeWorkerNATSSessionsByBootstrapTokenParams) ([]db.WorkerNatsSession, error) {
	return nil, nil
}

func (s *registerWorkerStore) CreateWorkerSettingsSchema(_ context.Context, arg db.CreateWorkerSettingsSchemaParams) (db.WorkerSettingsSchema, error) {
	s.createdSchemaArg = arg
	return db.WorkerSettingsSchema{ID: 55, WorkTypeID: arg.WorkTypeID, SettingsSchema: arg.SettingsSchema, InputSchema: arg.InputSchema, OutputSchema: arg.OutputSchema}, nil
}

func (s *registerWorkerStore) GetWorkerSettingsSchemaByID(context.Context, int32) (db.WorkerSettingsSchema, error) {
	return s.settingsSchema, nil
}

func (s *registerWorkerStore) GetWorkerSettingsSchemaByVersion(context.Context, db.GetWorkerSettingsSchemaByVersionParams) (db.WorkerSettingsSchema, error) {
	return db.WorkerSettingsSchema{}, pgx.ErrNoRows
}

func (s *registerWorkerStore) ListWorkerSettingsSchemasByWorkTypeID(context.Context, int32) ([]db.WorkerSettingsSchema, error) {
	return nil, nil
}

func (s *registerWorkerStore) CreateWorkerSettingsRevision(_ context.Context, arg db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error) {
	s.createdRevision = arg
	return db.WorkerSettingsRevision{ID: 88, WorkerSettingsSchemaID: arg.WorkerSettingsSchemaID, SettingsData: arg.SettingsData}, nil
}

func (s *registerWorkerStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (s *registerWorkerStore) ListWorkerSettingsRevisionsBySchemaID(context.Context, int32) ([]db.WorkerSettingsRevision, error) {
	return s.revisions, nil
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

func (s *registerWorkerStore) ListSystemsByOrganizationIDPaginated(context.Context, db.ListSystemsByOrganizationIDPaginatedParams) ([]db.ListSystemsByOrganizationIDPaginatedRow, error) {
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

type issuingNATSManager struct {
	issueCalls       int
	accountPublicKey string
	creds            natsauth.WorkerCredentials
}

func (m *issuingNATSManager) AccountPublicKey(context.Context) (string, error) {
	return m.accountPublicKey, nil
}

func (m *issuingNATSManager) IssueWorker(_ context.Context, scope natsauth.WorkerScope) (natsauth.WorkerCredentials, error) {
	m.issueCalls++
	if m.creds.UserJWT == "" {
		return natsauth.WorkerCredentials{
			AccountPublicKey: "ANEW",
			UserPublicKey:    "UNEW",
			UserJWT:          "new.jwt",
			UserSeed:         "SUnewseed",
			Creds:            []byte("new.creds"),
			Permissions:      natsauth.WorkerPermissions(scope),
		}, nil
	}
	return m.creds, nil
}

func (m *issuingNATSManager) RevokeWorker(context.Context, string) error {
	return nil
}

func mustPermissionsJSON(t *testing.T, perms natsauth.Permissions) []byte {
	t.Helper()
	raw, err := json.Marshal(perms)
	if err != nil {
		t.Fatalf("marshal permissions: %v", err)
	}
	return raw
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
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{ID: 9, OrganizationID: 2, WorkTypeID: 12, MaxActiveWorkers: 1},
		revisions:      []db.WorkerSettingsRevision{{ID: 99, WorkerSettingsSchemaID: 55}},
	}
	service := NewService(store, natsauth.NoopManager{})

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       manifest,
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if registration.ID != 77 {
		t.Fatalf("expected worker ID 77, got %d", registration.ID)
	}
	if registration.WorkTypeID != 12 {
		t.Fatalf("expected work_type_id 12, got %d", registration.WorkTypeID)
	}
	if registration.OrganizationID != 2 {
		t.Fatalf("expected organization_id 2, got %d", registration.OrganizationID)
	}
	if registration.RevisionID != 99 {
		t.Fatalf("expected revision_id 99, got %d", registration.RevisionID)
	}
	assertJSONEqual(t, `{"type":"object","properties":{"host":{"type":"string","required":true}}}`, store.createdSchemaArg.SettingsSchema)
	assertJSONEqual(t, `{"type":"object","properties":{"to":{"type":"string","required":true}}}`, store.createdSchemaArg.InputSchema)
	assertJSONEqual(t, `{"type":"object","properties":{"message_id":{"type":"string"}}}`, store.createdSchemaArg.OutputSchema)
}

func TestRegisterWorkerCreatesDefaultRevisionWhenSchemaHasNoRevisions(t *testing.T) {
	manifest := json.RawMessage(`{
		"kind": "smtp",
		"name_kind": "SMTP Email",
		"type": "email",
		"name_type": "Email Delivery",
		"settings_schema": {"type":"object","properties":{}},
		"input_schema": {"type":"object","properties":{}},
		"output_schema": {"type":"object","properties":{}}
	}`)
	store := &registerWorkerStore{bootstrapToken: db.WorkerBootstrapToken{ID: 9, OrganizationID: 2, WorkTypeID: 12, MaxActiveWorkers: 1}}
	service := NewService(store, natsauth.NoopManager{})

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       manifest,
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if registration.RevisionID != 88 {
		t.Fatalf("expected created revision_id 88, got %d", registration.RevisionID)
	}
	if store.createdRevision.WorkerSettingsSchemaID != 55 {
		t.Fatalf("expected default revision for schema 55, got %d", store.createdRevision.WorkerSettingsSchemaID)
	}
	assertJSONEqual(t, `{}`, store.createdRevision.SettingsData)
}

func TestRegisterWorkerUsesOrgScopedBootstrapTokenAndReturnsNATSCredentials(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 1,
		},
		revisions: []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	service := NewService(store, natsauth.NoopManager{})

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if registration.OrganizationID != 12 {
		t.Fatalf("organization_id mismatch: %d", registration.OrganizationID)
	}
	if registration.WorkTypeID != 3 {
		t.Fatalf("work_type_id mismatch: %d", registration.WorkTypeID)
	}
	if registration.NATS.UserJWT == "" || registration.NATS.UserSeed == "" {
		t.Fatalf("expected nats credentials in response")
	}
	if store.createdNATSSessionArg.WorkerID != registration.ID {
		t.Fatalf("expected nats session for worker %d, got %#v", registration.ID, store.createdNATSSessionArg)
	}
	if store.createdNATSSessionArg.BootstrapTokenID != 9 {
		t.Fatalf("expected bootstrap token 9 in nats session, got %#v", store.createdNATSSessionArg)
	}
}

func TestRegisterWorkerReusesActiveNATSSessionForSameWorkerAndBootstrap(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 1,
		},
		existingWorker: db.Worker{
			ID:                     77,
			OrganizationID:         12,
			WorkTypeID:             3,
			Name:                   "smtp-worker",
			WorkerSettingsSchemaID: 55,
		},
		activeNATSSessions: []db.WorkerNatsSession{{
			ID:                   44,
			WorkerID:             77,
			BootstrapTokenID:     9,
			NatsAccountPublicKey: "AOLD",
			NatsUserPublicKey:    "UOLD",
			NatsUserJwt:          "old.jwt",
			NatsUserSeed:         "SUoldseed",
			NatsUserCredentials:  "old.creds",
			Permissions: mustPermissionsJSON(t, natsauth.WorkerPermissions(natsauth.WorkerScope{
				OrganizationID: 12,
				WorkTypeID:     3,
				WorkerID:       77,
			})),
		}},
		revisions: []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	manager := &issuingNATSManager{}
	service := NewService(store, manager)

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if manager.issueCalls != 0 {
		t.Fatalf("expected active NATS session reuse without issuing credentials, got %d calls", manager.issueCalls)
	}
	if store.createdNATSSessionArg.WorkerID != 0 {
		t.Fatalf("expected no new NATS session, got %#v", store.createdNATSSessionArg)
	}
	if registration.NATS.UserJWT != "old.jwt" {
		t.Fatalf("expected previous NATS JWT, got %q", registration.NATS.UserJWT)
	}
	if registration.NATS.UserSeed != "SUoldseed" || registration.NATS.Credentials != "old.creds" {
		t.Fatalf("expected previous NATS credentials, got %#v", registration.NATS)
	}
}

func TestRegisterWorkerReplacesActiveNATSSessionWhenAccountChanged(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 1,
		},
		existingWorker: db.Worker{
			ID:                     77,
			OrganizationID:         12,
			WorkTypeID:             3,
			Name:                   "smtp-worker",
			WorkerSettingsSchemaID: 55,
		},
		activeNATSSessions: []db.WorkerNatsSession{{
			ID:                   44,
			WorkerID:             77,
			BootstrapTokenID:     9,
			NatsAccountPublicKey: "AOLD",
			NatsUserPublicKey:    "UOLD",
			NatsUserJwt:          "old.jwt",
			NatsUserSeed:         "SUoldseed",
			NatsUserCredentials:  "old.creds",
			Permissions: mustPermissionsJSON(t, natsauth.WorkerPermissions(natsauth.WorkerScope{
				OrganizationID: 12,
				WorkTypeID:     3,
				WorkerID:       77,
			})),
		}},
		revisions: []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	manager := &issuingNATSManager{accountPublicKey: "ANEW"}
	service := NewService(store, manager)

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if manager.issueCalls != 1 {
		t.Fatalf("expected stale account session to trigger credential replacement, got %d issues", manager.issueCalls)
	}
	if registration.NATS.UserJWT != "new.jwt" || registration.NATS.UserSeed != "SUnewseed" {
		t.Fatalf("expected new NATS credentials after account change, got %#v", registration.NATS)
	}
	if store.replacedNATSSessionArg.NatsAccountPublicKey != "ANEW" {
		t.Fatalf("expected replacement with current account, got %#v", store.replacedNATSSessionArg)
	}
}

func TestRegisterWorkerReplacesActiveNATSSessionWhenPermissionsAreStale(t *testing.T) {
	stalePerms := natsauth.Permissions{
		PublishAllow: []string{
			"result.org.12.>",
			"config.request.org.12.work_type.3.>",
			"$JS.API.CONSUMER.CREATE.TASKS.worker-77.task.org.12.work_type.3.>",
			"$JS.API.CONSUMER.MSG.NEXT.TASKS.worker-77",
			"$JS.ACK.TASKS.worker-77.>",
			"_INBOX.>",
		},
		SubscribeAllow: []string{"task.org.12.work_type.3.>", "config.org.12.work_type.3.>", "_INBOX.>"},
	}
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 1,
		},
		existingWorker: db.Worker{
			ID:                     77,
			OrganizationID:         12,
			WorkTypeID:             3,
			Name:                   "smtp-worker",
			WorkerSettingsSchemaID: 55,
		},
		activeNATSSessions: []db.WorkerNatsSession{{
			ID:                   44,
			WorkerID:             77,
			BootstrapTokenID:     9,
			NatsAccountPublicKey: "AOLD",
			NatsUserPublicKey:    "UOLD",
			NatsUserJwt:          "old.jwt",
			NatsUserSeed:         "SUoldseed",
			NatsUserCredentials:  "old.creds",
			Permissions:          mustPermissionsJSON(t, stalePerms),
		}},
		revisions: []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	manager := &issuingNATSManager{}
	service := NewService(store, manager)

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if manager.issueCalls != 1 {
		t.Fatalf("expected stale NATS permissions to trigger credential replacement, got %d issues", manager.issueCalls)
	}
	if store.createdNATSSessionArg.WorkerID != 0 {
		t.Fatalf("expected stale active NATS session replacement without create, got %#v", store.createdNATSSessionArg)
	}
	if store.replacedNATSSessionArg.WorkerID != 77 || store.replacedNATSSessionArg.BootstrapTokenID != 9 {
		t.Fatalf("expected replacement for worker/bootstrap, got %#v", store.replacedNATSSessionArg)
	}
	if registration.NATS.UserJWT != "new.jwt" || registration.NATS.UserSeed != "SUnewseed" {
		t.Fatalf("expected new NATS credentials from replacement, got %#v", registration.NATS)
	}
}

func TestRegisterWorkerReplacesInactiveNATSSessionForSameWorkerAndBootstrap(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 1,
		},
		existingWorker: db.Worker{
			ID:                     77,
			OrganizationID:         12,
			WorkTypeID:             3,
			Name:                   "smtp-worker",
			WorkerSettingsSchemaID: 55,
		},
		latestNATSSession: db.WorkerNatsSession{
			ID:               44,
			WorkerID:         77,
			BootstrapTokenID: 9,
			RevokedAt:        pgtype.Timestamp{Valid: true},
		},
		revisions: []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	manager := &issuingNATSManager{}
	service := NewService(store, manager)

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}

	if manager.issueCalls != 1 {
		t.Fatalf("expected one NATS credential issue for inactive session replacement, got %d", manager.issueCalls)
	}
	if store.createdNATSSessionArg.WorkerID != 0 {
		t.Fatalf("expected inactive NATS session replacement without create, got %#v", store.createdNATSSessionArg)
	}
	if store.replacedNATSSessionArg.WorkerID != 77 || store.replacedNATSSessionArg.BootstrapTokenID != 9 {
		t.Fatalf("expected replacement for worker/bootstrap, got %#v", store.replacedNATSSessionArg)
	}
	if registration.NATS.UserJWT != "new.jwt" || registration.NATS.UserSeed != "SUnewseed" {
		t.Fatalf("expected new NATS credentials from replacement, got %#v", registration.NATS)
	}
}

func TestRegisterWorkerDeniesNewWorkerWhenActiveWorkerLimitReached(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 4,
		},
		activeWorkerCount: 4,
		revisions:         []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	service := NewService(store, natsauth.NoopManager{})

	_, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker-5",
		Manifest:       validManifestJSON(),
	})
	if !errors.Is(err, ErrWorkerBootstrapActiveWorkerLimitExceeded) {
		t.Fatalf("expected active worker limit error, got %v", err)
	}
	if store.createdNATSSessionArg.WorkerID != 0 {
		t.Fatalf("expected no NATS session when capacity is exceeded, got %#v", store.createdNATSSessionArg)
	}
}

func TestRegisterWorkerAllowsExistingActiveWorkerToRestartAtCapacity(t *testing.T) {
	store := &registerWorkerStore{
		bootstrapToken: db.WorkerBootstrapToken{
			ID:               9,
			OrganizationID:   12,
			WorkTypeID:       3,
			Status:           "active",
			MaxActiveWorkers: 4,
		},
		existingWorker: db.Worker{
			ID:                     77,
			OrganizationID:         12,
			WorkTypeID:             3,
			Name:                   "smtp-worker-1",
			WorkerSettingsSchemaID: 55,
		},
		activeWorkerCount: 3,
		revisions:         []db.WorkerSettingsRevision{{ID: 73, WorkerSettingsSchemaID: 55}},
	}
	service := NewService(store, natsauth.NoopManager{})

	registration, err := service.RegisterWorker(context.Background(), RegisterWorkerRequest{
		BootstrapToken: "token",
		Name:           "smtp-worker-1",
		Manifest:       validManifestJSON(),
	})
	if err != nil {
		t.Fatalf("RegisterWorker error: %v", err)
	}
	if registration.ID != 77 {
		t.Fatalf("expected existing worker 77, got %d", registration.ID)
	}
	if store.createdNATSSessionArg.WorkerID != 77 {
		t.Fatalf("expected NATS session for restarted worker, got %#v", store.createdNATSSessionArg)
	}
}

func validManifestJSON() json.RawMessage {
	return json.RawMessage(`{
		"kind": "smtp",
		"name_kind": "SMTP Email",
		"type": "email",
		"name_type": "Email Delivery",
		"settings_schema": {"type":"object","properties":{}},
		"input_schema": {"type":"object","properties":{}},
		"output_schema": {"type":"object","properties":{}}
	}`)
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
