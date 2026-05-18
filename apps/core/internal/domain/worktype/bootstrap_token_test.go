package worktype

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/natsauth"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type bootstrapTokenStore struct {
	registerWorkerStore
	createArg      db.CreateWorkerBootstrapTokenParams
	tokens         []db.ListWorkerBootstrapTokensByOrganizationRow
	revokeArg      db.RevokeWorkerBootstrapTokenParams
	activeSessions []db.WorkerNatsSession
	revokeSessions db.RevokeWorkerNATSSessionsByBootstrapTokenParams
}

func (s *bootstrapTokenStore) CreateWorkerBootstrapToken(_ context.Context, arg db.CreateWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	s.createArg = arg
	return db.WorkerBootstrapToken{
		ID:               7,
		OrganizationID:   arg.OrganizationID,
		WorkTypeID:       arg.WorkTypeID,
		Name:             arg.Name,
		Description:      arg.Description,
		TokenHash:        arg.TokenHash,
		MaxActiveWorkers: arg.MaxActiveWorkers,
		ExpiresAt:        arg.ExpiresAt,
		CreatedAt:        pgtype.Timestamp{Time: time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC), Valid: true},
	}, nil
}

func (s *bootstrapTokenStore) ListWorkerBootstrapTokensByOrganization(context.Context, int32) ([]db.ListWorkerBootstrapTokensByOrganizationRow, error) {
	return s.tokens, nil
}

func (s *bootstrapTokenStore) RevokeWorkerBootstrapToken(_ context.Context, arg db.RevokeWorkerBootstrapTokenParams) (db.WorkerBootstrapToken, error) {
	s.revokeArg = arg
	return db.WorkerBootstrapToken{ID: arg.ID, Status: "revoked"}, nil
}

func (s *bootstrapTokenStore) ListActiveWorkerNATSSessionsByBootstrapToken(context.Context, int32) ([]db.WorkerNatsSession, error) {
	return s.activeSessions, nil
}

func (s *bootstrapTokenStore) RevokeWorkerNATSSessionsByBootstrapToken(_ context.Context, arg db.RevokeWorkerNATSSessionsByBootstrapTokenParams) ([]db.WorkerNatsSession, error) {
	s.revokeSessions = arg
	return s.activeSessions, nil
}

type recordingNATSManager struct {
	natsauth.NoopManager
	revoked []string
}

func (m *recordingNATSManager) RevokeWorker(_ context.Context, userPublicKey string) error {
	m.revoked = append(m.revoked, userPublicKey)
	return nil
}

func TestCreateWorkerBootstrapTokenHashesPlaintext(t *testing.T) {
	store := &bootstrapTokenStore{}
	service := NewService(store, natsauth.NoopManager{})

	result, err := service.CreateWorkerBootstrapToken(context.Background(), 12, 34, CreateWorkerBootstrapTokenRequest{
		Name:             "smtp-prod",
		WorkTypeID:       3,
		Description:      "SMTP pool",
		MaxActiveWorkers: 10,
	})
	if err != nil {
		t.Fatalf("CreateWorkerBootstrapToken error: %v", err)
	}
	if result.Plaintext == "" {
		t.Fatal("expected plaintext token")
	}
	if store.createArg.TokenHash == "" || store.createArg.TokenHash == result.Plaintext {
		t.Fatalf("expected stored hash to differ from plaintext, got %q", store.createArg.TokenHash)
	}
	if store.createArg.OrganizationID != 12 || store.createArg.CreatedByUserID.Int32 != 34 {
		t.Fatalf("unexpected create args: %#v", store.createArg)
	}
}

func TestRevokeWorkerBootstrapTokenRevokesActiveNATSSessions(t *testing.T) {
	store := &bootstrapTokenStore{
		activeSessions: []db.WorkerNatsSession{
			{ID: 1, NatsUserPublicKey: "U1"},
			{ID: 2, NatsUserPublicKey: "U2"},
		},
	}
	manager := &recordingNATSManager{}
	service := NewService(store, manager)

	if err := service.RevokeWorkerBootstrapToken(context.Background(), 9, 34, true); err != nil {
		t.Fatalf("RevokeWorkerBootstrapToken error: %v", err)
	}
	if store.revokeArg.ID != 9 || store.revokeArg.RevokedByUserID.Int32 != 34 {
		t.Fatalf("unexpected revoke args: %#v", store.revokeArg)
	}
	if len(manager.revoked) != 2 || manager.revoked[0] != "U1" || manager.revoked[1] != "U2" {
		t.Fatalf("expected both nats users revoked, got %#v", manager.revoked)
	}
	if store.revokeSessions.BootstrapTokenID != 9 {
		t.Fatalf("expected db session revoke for token 9, got %#v", store.revokeSessions)
	}
}
