package worktype

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/natsauth"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// TestManifestHash проверяет детерминированность хэша манифеста.
func TestManifestHash(t *testing.T) {
	t.Run("same input produces same hash", func(t *testing.T) {
		manifest := json.RawMessage(`{"version":"1.0","settings":{"key":"value"}}`)

		hash1, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		hash2, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if hash1 != hash2 {
			t.Errorf("expected same hash, got %q and %q", hash1, hash2)
		}
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		manifest1 := json.RawMessage(`{"version":"1.0"}`)
		manifest2 := json.RawMessage(`{"version":"2.0"}`)

		hash1, err := ManifestHash(manifest1)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		hash2, err := ManifestHash(manifest2)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if hash1 == hash2 {
			t.Errorf("expected different hashes but got same: %q", hash1)
		}
	})

	t.Run("hash is 64 char hex string", func(t *testing.T) {
		manifest := json.RawMessage(`{"test":"data"}`)
		hash, err := ManifestHash(manifest)
		if err != nil {
			t.Fatalf("ManifestHash error: %v", err)
		}
		if len(hash) != 64 {
			t.Errorf("expected 64-char hash, got len=%d: %q", len(hash), hash)
		}
	})
}

// TestComputeWorkerStatus проверяет вычисление статуса воркера.
func TestComputeWorkerStatus(t *testing.T) {
	t.Run("hasRunningStep=true returns working", func(t *testing.T) {
		// Даже с устаревшим heartbeat — если есть running step → working
		staleHeartbeat := time.Now().Add(-5 * time.Minute)
		status := ComputeWorkerStatus(staleHeartbeat, true)
		if status != WorkerStatusWorking {
			t.Errorf("expected working, got %q", status)
		}
	})

	t.Run("recent heartbeat returns ready", func(t *testing.T) {
		// Heartbeat < 90 секунд назад
		recentHeartbeat := time.Now().Add(-30 * time.Second)
		status := ComputeWorkerStatus(recentHeartbeat, false)
		if status != WorkerStatusReady {
			t.Errorf("expected ready, got %q", status)
		}
	})

	t.Run("stale heartbeat returns offline", func(t *testing.T) {
		// Heartbeat > 90 секунд назад
		staleHeartbeat := time.Now().Add(-2 * time.Minute)
		status := ComputeWorkerStatus(staleHeartbeat, false)
		if status != WorkerStatusOffline {
			t.Errorf("expected offline, got %q", status)
		}
	})
}

type catalogStore struct {
	registerWorkerStore
	workTypes         []db.WorkType
	workersByWorkType map[int32][]db.Worker
	schemasByWorkType map[int32][]db.WorkerSettingsSchema
}

func (s *catalogStore) ListWorkTypes(context.Context) ([]db.WorkType, error) {
	return s.workTypes, nil
}

func (s *catalogStore) ListNewWorkersByWorkTypeID(_ context.Context, workTypeID int32) ([]db.Worker, error) {
	return s.workersByWorkType[workTypeID], nil
}

func (s *catalogStore) ListWorkerSettingsSchemasByWorkTypeID(_ context.Context, workTypeID int32) ([]db.WorkerSettingsSchema, error) {
	return s.schemasByWorkType[workTypeID], nil
}

func TestListWorkTypeCatalogFiltersSchemasWithoutWorkersAndIncludesDetails(t *testing.T) {
	createdAt := time.Date(2026, 5, 11, 9, 30, 0, 0, time.UTC)
	store := &catalogStore{
		workTypes: []db.WorkType{
			{ID: 7, Name: "SMTP", Code: "smtp"},
		},
		workersByWorkType: map[int32][]db.Worker{
			7: {
				{
					ID:                     100,
					WorkTypeID:             7,
					WorkerSettingsSchemaID: 11,
					LastHeartbeatAt:        pgtype.Timestamp{Time: time.Now(), Valid: true},
				},
			},
		},
		schemasByWorkType: map[int32][]db.WorkerSettingsSchema{
			7: {
				{
					ID:             11,
					WorkTypeID:     7,
					Version:        "hash-with-worker",
					SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string"}}}`),
					CreatedAt:      pgtype.Timestamp{Time: createdAt, Valid: true},
				},
				{
					ID:             12,
					WorkTypeID:     7,
					Version:        "hash-without-worker",
					SettingsSchema: []byte(`{"type":"object","properties":{"unused":{"type":"string"}}}`),
					CreatedAt:      pgtype.Timestamp{Time: createdAt.Add(time.Hour), Valid: true},
				},
			},
		},
	}

	catalog, err := NewService(store, natsauth.NoopManager{}).ListWorkTypeCatalog(context.Background())
	if err != nil {
		t.Fatalf("ListWorkTypeCatalog error: %v", err)
	}
	if len(catalog) != 1 {
		t.Fatalf("expected one catalog item, got %d", len(catalog))
	}
	schemas := catalog[0].Schemas
	if len(schemas) != 1 {
		t.Fatalf("expected only worker-backed schemas, got %#v", schemas)
	}
	if schemas[0].ID != 11 {
		t.Fatalf("expected schema 11, got %#v", schemas[0])
	}
	if schemas[0].CreatedAt != "2026-05-11T09:30:00Z" {
		t.Fatalf("expected RFC3339 created_at, got %q", schemas[0].CreatedAt)
	}
	if schemas[0].WorkerCount != 1 || schemas[0].ReadyWorkers != 1 {
		t.Fatalf("expected worker counts 1/1, got %#v", schemas[0])
	}
	assertJSONEqual(t, `{"type":"object","properties":{"host":{"type":"string"}}}`, schemas[0].SettingsSchema)
}

func TestListWorkTypesFiltersControlWorkTypes(t *testing.T) {
	store := &catalogStore{
		workTypes: []db.WorkType{
			{
				ID:   1,
				Name: "SMTP",
				Code: "smtp",
				Meta: []byte(`{"color":"#1976d2"}`),
			},
			{
				ID:   2,
				Name: "Delay",
				Code: "delay",
				Meta: []byte(`{"kind":"control"}`),
			},
		},
	}

	workTypes, err := NewService(store, natsauth.NoopManager{}).ListWorkTypes(context.Background())
	if err != nil {
		t.Fatalf("ListWorkTypes error: %v", err)
	}
	if len(workTypes) != 1 {
		t.Fatalf("expected only non-control work types, got %#v", workTypes)
	}
	if workTypes[0].ID != 1 || workTypes[0].Code != "smtp" {
		t.Fatalf("expected SMTP work type, got %#v", workTypes[0])
	}
}

func TestCreateWorkTypeRejectsControlWorkType(t *testing.T) {
	_, err := NewService(&registerWorkerStore{}, natsauth.NoopManager{}).CreateWorkType(context.Background(), CreateWorkTypeRequest{
		Name: "Delay",
		Code: "delay",
		Meta: json.RawMessage(`{"kind":"control"}`),
	})

	if !errors.Is(err, ErrControlWorkTypeReserved) {
		t.Fatalf("expected ErrControlWorkTypeReserved, got %v", err)
	}
}

func TestCreateSettingsRevisionRejectsMissingRequiredSettings(t *testing.T) {
	store := &registerWorkerStore{
		settingsSchema: db.WorkerSettingsSchema{
			ID:             22,
			SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		},
	}
	svc := NewService(store, natsauth.NoopManager{})

	_, err := svc.CreateSettingsRevision(context.Background(), 22, CreateRevisionRequest{
		SettingsData: json.RawMessage(`{}`),
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateSettingsRevisionAcceptsValidSettings(t *testing.T) {
	store := &registerWorkerStore{
		settingsSchema: db.WorkerSettingsSchema{
			ID:             22,
			SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		},
	}
	svc := NewService(store, natsauth.NoopManager{})

	revision, err := svc.CreateSettingsRevision(context.Background(), 22, CreateRevisionRequest{
		SettingsData: json.RawMessage(`{"host":"smtp.local"}`),
	})
	if err != nil {
		t.Fatalf("CreateSettingsRevision error: %v", err)
	}
	if revision.ID != 88 {
		t.Fatalf("expected created revision 88, got %d", revision.ID)
	}
	assertJSONEqual(t, `{"host":"smtp.local"}`, store.createdRevision.SettingsData)
}

// TestGenerateBootstrapToken проверяет формат bootstrap-токена.
func TestGenerateBootstrapToken(t *testing.T) {
	plaintext, hash, err := GenerateBootstrapToken()
	if err != nil {
		t.Fatalf("GenerateBootstrapToken error: %v", err)
	}

	t.Run("plaintext is 64-char hex string", func(t *testing.T) {
		if len(plaintext) != 64 {
			t.Errorf("expected 64-char plaintext, got len=%d: %q", len(plaintext), plaintext)
		}
		if !isHex(plaintext) {
			t.Errorf("expected hex string, got: %q", plaintext)
		}
	})

	t.Run("hash is 64-char hex string", func(t *testing.T) {
		if len(hash) != 64 {
			t.Errorf("expected 64-char hash, got len=%d: %q", len(hash), hash)
		}
		if !isHex(hash) {
			t.Errorf("expected hex string, got: %q", hash)
		}
	})

	t.Run("plaintext and hash are different", func(t *testing.T) {
		if plaintext == hash {
			t.Error("plaintext and hash should be different")
		}
	})

	t.Run("two calls produce different tokens", func(t *testing.T) {
		p2, h2, err := GenerateBootstrapToken()
		if err != nil {
			t.Fatalf("GenerateBootstrapToken error: %v", err)
		}
		if plaintext == p2 {
			t.Error("two calls should produce different plaintexts")
		}
		if hash == h2 {
			t.Error("two calls should produce different hashes")
		}
	})
}

// isHex проверяет что строка состоит из hex-символов.
func isHex(s string) bool {
	s = strings.ToLower(s)
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

type deleteWorkerStore struct {
	registerWorkerStore
	worker       db.Worker
	getWorkerErr error
	usages       []db.ListWorkflowUsagesByWorkerIDRow
	deleteCalled bool
}

func (s *deleteWorkerStore) GetNewWorkerByID(context.Context, int32) (db.Worker, error) {
	return s.worker, s.getWorkerErr
}

func (s *deleteWorkerStore) ListWorkflowUsagesByWorkerID(context.Context, int32) ([]db.ListWorkflowUsagesByWorkerIDRow, error) {
	return s.usages, nil
}

func (s *deleteWorkerStore) DeleteWorker(context.Context, int32) error {
	s.deleteCalled = true
	return nil
}

func TestDeleteWorkerDeletesOfflineUnusedWorker(t *testing.T) {
	store := &deleteWorkerStore{
		worker: db.Worker{
			ID:              10,
			LastHeartbeatAt: pgtype.Timestamp{Time: time.Now().Add(-2 * time.Minute), Valid: true},
		},
	}
	resp, err := NewService(store, natsauth.NoopManager{}).DeleteWorker(context.Background(), 10)
	if err != nil {
		t.Fatalf("DeleteWorker error: %v", err)
	}
	if resp.Result != DeleteWorkerResultDeleted || !resp.Deleted {
		t.Fatalf("expected deleted response, got %#v", resp)
	}
	if !store.deleteCalled {
		t.Fatal("expected DeleteWorker to be called")
	}
}

func TestDeleteWorkerKeepsOnlineWorker(t *testing.T) {
	store := &deleteWorkerStore{
		worker: db.Worker{
			ID:              11,
			LastHeartbeatAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}
	resp, err := NewService(store, natsauth.NoopManager{}).DeleteWorker(context.Background(), 11)
	if err != nil {
		t.Fatalf("DeleteWorker error: %v", err)
	}
	if resp.Result != DeleteWorkerResultOnline || resp.Deleted {
		t.Fatalf("expected online response, got %#v", resp)
	}
	if store.deleteCalled {
		t.Fatal("DeleteWorker must not be called for online worker")
	}
}

func TestDeleteWorkerReturnsWorkflowUsages(t *testing.T) {
	store := &deleteWorkerStore{
		worker: db.Worker{
			ID:              12,
			LastHeartbeatAt: pgtype.Timestamp{Time: time.Now().Add(-2 * time.Minute), Valid: true},
		},
		usages: []db.ListWorkflowUsagesByWorkerIDRow{
			{WorkflowID: 101, WorkflowName: "Welcome", SystemID: 7, WorkflowVersionID: 201, WorkflowVersionNumber: 3},
		},
	}
	resp, err := NewService(store, natsauth.NoopManager{}).DeleteWorker(context.Background(), 12)
	if err != nil {
		t.Fatalf("DeleteWorker error: %v", err)
	}
	if resp.Result != DeleteWorkerResultInUse || resp.Deleted {
		t.Fatalf("expected in_use response, got %#v", resp)
	}
	if len(resp.Usages) != 1 || resp.Usages[0].WorkflowID != 101 {
		t.Fatalf("expected workflow usage in response, got %#v", resp.Usages)
	}
	if store.deleteCalled {
		t.Fatal("DeleteWorker must not be called when worker is used by workflow definitions")
	}
}

func TestDeleteWorkerReturnsNotFoundBody(t *testing.T) {
	store := &deleteWorkerStore{getWorkerErr: pgx.ErrNoRows}
	resp, err := NewService(store, natsauth.NoopManager{}).DeleteWorker(context.Background(), 99)
	if err != nil {
		t.Fatalf("DeleteWorker error: %v", err)
	}
	if resp.Result != DeleteWorkerResultNotFound || resp.Deleted {
		t.Fatalf("expected not_found response, got %#v", resp)
	}
}
