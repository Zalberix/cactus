package worktype

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalberix/cactus/apps/core/internal/natsauth"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

var (
	ErrWorkflowTokenSystemMismatch              = errors.New("system token and workflow belong to different systems")
	ErrWorkerBootstrapActiveWorkerLimitExceeded = errors.New("worker bootstrap token active worker limit exceeded")
)

// HeartbeatTimeout — таймаут для определения offline-статуса воркера.
const HeartbeatTimeout = 90 * time.Second

// Service — бизнес-логика управления типами работ, воркерами, системами.
type Service struct {
	store      Storage
	natsAuth   natsauth.Manager
	natsURL    string
	natsCAFile string
}

type ServiceOption func(*Service)

func WithNATSURL(url string) ServiceOption {
	return func(s *Service) {
		s.natsURL = url
	}
}

func WithNATSCAFile(caFile string) ServiceOption {
	return func(s *Service) {
		s.natsCAFile = caFile
	}
}

type workerManifest struct {
	Kind           string          `json:"kind"`
	NameKind       string          `json:"name_kind"`
	Type           string          `json:"type"`
	NameType       string          `json:"name_type"`
	SettingsSchema json.RawMessage `json:"settings_schema"`
	InputSchema    json.RawMessage `json:"input_schema"`
	OutputSchema   json.RawMessage `json:"output_schema"`
}

// NewService создаёт новый worktype.Service.
func NewService(store Storage, natsAuth natsauth.Manager, opts ...ServiceOption) *Service {
	if natsAuth == nil {
		natsAuth = natsauth.NoopManager{}
	}
	s := &Service{store: store, natsAuth: natsAuth}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// --- Helper functions ---

// GenerateBootstrapToken генерирует bootstrap-токен.
// Возвращает plaintext (64-char hex) и hash (64-char hex sha256).
func GenerateBootstrapToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	plaintext = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(h[:])
	return plaintext, hash, nil
}

// ManifestHash вычисляет SHA256 хэш манифеста (JSON).
// Используется как version при хранении WorkerSettingsSchema.
func ManifestHash(manifest json.RawMessage) (string, error) {
	h := sha256.Sum256(manifest)
	return hex.EncodeToString(h[:]), nil
}

func parseWorkerManifest(manifest json.RawMessage) (workerManifest, error) {
	var parsed workerManifest
	if err := json.Unmarshal(manifest, &parsed); err != nil {
		return workerManifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	for name, schema := range map[string]json.RawMessage{
		"settings_schema": parsed.SettingsSchema,
		"input_schema":    parsed.InputSchema,
		"output_schema":   parsed.OutputSchema,
	} {
		if err := validateObjectSchema(name, schema); err != nil {
			return workerManifest{}, err
		}
	}
	return parsed, nil
}

func validateObjectSchema(name string, schema json.RawMessage) error {
	if len(schema) == 0 {
		return fmt.Errorf("manifest.%s is required", name)
	}
	var object map[string]any
	if err := json.Unmarshal(schema, &object); err != nil {
		return fmt.Errorf("manifest.%s must be a JSON object: %w", name, err)
	}
	if object["type"] != "object" {
		return fmt.Errorf("manifest.%s must have type=object", name)
	}
	if _, ok := object["properties"].(map[string]any); !ok {
		return fmt.Errorf("manifest.%s must have properties object", name)
	}
	return nil
}

// ComputeWorkerStatus вычисляет статус воркера на лету.
// hasRunningStep=true → working; last_heartbeat < 90s → ready; иначе → offline.
func ComputeWorkerStatus(lastHeartbeat time.Time, hasRunningStep bool) WorkerStatus {
	if hasRunningStep {
		return WorkerStatusWorking
	}
	if time.Since(lastHeartbeat) < HeartbeatTimeout {
		return WorkerStatusReady
	}
	return WorkerStatusOffline
}

// hashPrivateToken хэширует приватный токен через SHA256.
func hashPrivateToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// VerifyPrivateToken проверяет приватный токен через constant-time comparison.
func VerifyPrivateToken(storedHash, incoming string) bool {
	incomingHash := hashPrivateToken(incoming)
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(incomingHash)) == 1
}

// --- Work Type methods ---

type schemaWorkerStats struct {
	WorkerCount  int64
	ReadyWorkers int64
}

// CreateWorkType создаёт тип работы и генерирует bootstrap-токен.
func (s *Service) CreateWorkType(ctx context.Context, req CreateWorkTypeRequest) (db.WorkType, error) {
	wt, err := s.store.CreateWorkType(ctx, db.CreateWorkTypeParams{
		Name: req.Name,
		Code: req.Code,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Meta: req.Meta,
	})
	if err != nil {
		return db.WorkType{}, fmt.Errorf("create work type: %w", err)
	}
	return wt, nil
}

// ListWorkTypes возвращает все типы работ.
func (s *Service) ListWorkTypes(ctx context.Context) ([]Response, error) {
	rows, err := s.store.ListWorkTypes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Response, 0, len(rows))
	for _, r := range rows {
		wt := Response{
			ID:   r.ID,
			Name: r.Name,
			Code: r.Code,
		}
		if r.Description.Valid {
			wt.Description = r.Description.String
		}
		if len(r.Meta) > 0 {
			wt.Meta = json.RawMessage(r.Meta)
		}
		result = append(result, wt)
	}
	return result, nil
}

func (s *Service) ListWorkTypeCatalog(ctx context.Context) ([]CatalogItem, error) {
	workTypes, err := s.store.ListWorkTypes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CatalogItem, 0, len(workTypes))
	for _, wt := range workTypes {
		if isControlWorkType(wt.Meta) {
			continue
		}

		workers, err := s.store.ListNewWorkersByWorkTypeID(ctx, wt.ID)
		if err != nil {
			return nil, fmt.Errorf("list workers for work type %d: %w", wt.ID, err)
		}
		schemas, err := s.store.ListWorkerSettingsSchemasByWorkTypeID(ctx, wt.ID)
		if err != nil {
			return nil, fmt.Errorf("list schemas for work type %d: %w", wt.ID, err)
		}

		schemaStats := make(map[int32]schemaWorkerStats, len(schemas))
		item := CatalogItem{
			ID:          wt.ID,
			Name:        wt.Name,
			Code:        wt.Code,
			WorkerCount: int64(len(workers)),
		}
		if wt.Description.Valid {
			item.Description = wt.Description.String
		}
		if len(wt.Meta) > 0 {
			item.Meta = json.RawMessage(wt.Meta)
		}
		for _, worker := range workers {
			var lastHB time.Time
			if worker.LastHeartbeatAt.Valid {
				lastHB = worker.LastHeartbeatAt.Time
			}
			stats := schemaStats[worker.WorkerSettingsSchemaID]
			stats.WorkerCount++
			if ComputeWorkerStatus(lastHB, false) == WorkerStatusReady {
				stats.ReadyWorkers++
				item.ReadyWorkers++
			}
			schemaStats[worker.WorkerSettingsSchemaID] = stats
		}
		item.Schemas = settingsSchemaBriefs(schemas, schemaStats)
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) ListSettingsSchemasByWorkType(ctx context.Context, workTypeID int32) ([]db.WorkerSettingsSchema, error) {
	return s.store.ListWorkerSettingsSchemasByWorkTypeID(ctx, workTypeID)
}

func isControlWorkType(meta []byte) bool {
	if len(meta) == 0 {
		return false
	}
	var object map[string]interface{}
	if err := json.Unmarshal(meta, &object); err != nil {
		return false
	}
	return object["kind"] == "control"
}

func settingsSchemaBriefs(schemas []db.WorkerSettingsSchema, statsBySchema map[int32]schemaWorkerStats) []SettingsSchemaBrief {
	briefs := make([]SettingsSchemaBrief, 0, len(schemas))
	for _, schema := range schemas {
		stats := statsBySchema[schema.ID]
		if stats.WorkerCount == 0 {
			continue
		}
		briefs = append(briefs, SettingsSchemaBrief{
			ID:             schema.ID,
			Version:        schema.Version,
			CreatedAt:      timestampString(schema.CreatedAt),
			SettingsSchema: json.RawMessage(schema.SettingsSchema),
			WorkerCount:    stats.WorkerCount,
			ReadyWorkers:   stats.ReadyWorkers,
		})
	}
	return briefs
}

func timestampString(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.Format(time.RFC3339)
}

// --- Worker methods ---

// RegisterWorker регистрирует воркер по bootstrap-токену и манифесту.
// 1. Хэшируем bootstrap_token и ищем worker_bootstrap_token.
// 2. Вычисляем манифест-хэш → ищем/создаём WorkerSettingsSchema
// 3. Ищем/создаём Worker по (work_type_id, name)
// 4. Обновляем heartbeat
func (s *Service) RegisterWorker(ctx context.Context, req RegisterWorkerRequest) (RegisterWorkerResponse, error) { //nolint:gocognit // Registration keeps schema, worker, heartbeat, and NATS session in one transaction.
	tokenHash := sha256hex(req.BootstrapToken)

	manifest, err := parseWorkerManifest(req.Manifest)
	if err != nil {
		return RegisterWorkerResponse{}, err
	}

	// 2. Вычисляем манифест хэш (используется как version)
	manifestHash, err := ManifestHash(req.Manifest)
	if err != nil {
		return RegisterWorkerResponse{}, fmt.Errorf("compute manifest hash: %w", err)
	}

	// 3. Ищем или создаём WorkerSettingsSchema
	var registration RegisterWorkerResponse
	err = s.store.WithRegistrationTx(ctx, func(tx RegistrationTx) error {
		token, err := tx.GetActiveWorkerBootstrapTokenByHashForUpdate(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("invalid bootstrap token: %w", err)
		}

		schema, err := tx.GetWorkerSettingsSchemaByVersion(ctx, db.GetWorkerSettingsSchemaByVersionParams{
			WorkTypeID: token.WorkTypeID,
			Version:    manifestHash,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("get schema: %w", err)
			}
			// Создаём новую схему
			schema, err = tx.CreateWorkerSettingsSchema(ctx, db.CreateWorkerSettingsSchemaParams{
				WorkTypeID:     token.WorkTypeID,
				Version:        manifestHash,
				SettingsSchema: manifest.SettingsSchema,
				InputSchema:    manifest.InputSchema,
				OutputSchema:   manifest.OutputSchema,
			})
			if err != nil {
				return fmt.Errorf("create worker settings schema: %w", err)
			}
		}

		// 4. Ищем или создаём воркер по (work_type_id, name)
		revisionID, err := s.resolveRegistrationRevisionIDWithTx(ctx, tx, schema.ID)
		if err != nil {
			return err
		}

		worker, err := tx.GetWorkerByOrgWorkTypeAndName(ctx, db.GetWorkerByOrgWorkTypeAndNameParams{
			OrganizationID: token.OrganizationID,
			WorkTypeID:     token.WorkTypeID,
			Name:           req.Name,
		})
		if err != nil { //nolint:nestif // pgx.ErrNoRows creates a worker; other lookup errors bubble up.
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("get worker: %w", err)
			}
			// Создаём нового воркера
			activeWorkerCount, err := tx.CountActiveWorkersByBootstrapTokenExcludingWorker(ctx, db.CountActiveWorkersByBootstrapTokenExcludingWorkerParams{
				BootstrapTokenID:        token.ID,
				HeartbeatTimeoutSeconds: int32(HeartbeatTimeout / time.Second),
				ExcludeWorkerID:         0,
			})
			if err != nil {
				return fmt.Errorf("count active workers for bootstrap token: %w", err)
			}
			if activeWorkerCount >= token.MaxActiveWorkers {
				return ErrWorkerBootstrapActiveWorkerLimitExceeded
			}

			worker, err = tx.CreateNewWorker(ctx, db.CreateNewWorkerParams{
				OrganizationID:         token.OrganizationID,
				WorkTypeID:             token.WorkTypeID,
				WorkerSettingsSchemaID: schema.ID,
				Name:                   req.Name,
				Metadata:               []byte("{}"),
			})
			if err != nil {
				return fmt.Errorf("create worker: %w", err)
			}
		} else {
			activeWorkerCount, err := tx.CountActiveWorkersByBootstrapTokenExcludingWorker(ctx, db.CountActiveWorkersByBootstrapTokenExcludingWorkerParams{
				BootstrapTokenID:        token.ID,
				HeartbeatTimeoutSeconds: int32(HeartbeatTimeout / time.Second),
				ExcludeWorkerID:         worker.ID,
			})
			if err != nil {
				return fmt.Errorf("count active workers for bootstrap token: %w", err)
			}
			if activeWorkerCount >= token.MaxActiveWorkers {
				return ErrWorkerBootstrapActiveWorkerLimitExceeded
			}
		}

		if worker.ID != 0 && worker.WorkerSettingsSchemaID != schema.ID {
			// Обновляем схему если манифест изменился
			worker, err = tx.UpdateNewWorkerSchema(ctx, db.UpdateNewWorkerSchemaParams{
				ID:                     worker.ID,
				WorkerSettingsSchemaID: schema.ID,
			})
			if err != nil {
				return fmt.Errorf("update worker schema: %w", err)
			}
		}

		// 5. Обновляем heartbeat (WORK-05 — и для новых, и для повторно регистрирующихся)
		if err = tx.UpdateNewWorkerHeartbeat(ctx, worker.ID); err != nil {
			return fmt.Errorf("update worker heartbeat: %w", err)
		}
		if _, err = tx.TouchWorkerBootstrapTokenUse(ctx, token.ID); err != nil {
			return fmt.Errorf("touch bootstrap token use: %w", err)
		}

		creds, err := s.natsAuth.IssueWorker(ctx, natsauth.WorkerScope{
			OrganizationID: token.OrganizationID,
			WorkTypeID:     token.WorkTypeID,
			WorkerID:       worker.ID,
		})
		if err != nil {
			return fmt.Errorf("issue nats credentials: %w", err)
		}
		permissionsJSON, err := json.Marshal(creds.Permissions)
		if err != nil {
			return fmt.Errorf("marshal nats permissions: %w", err)
		}
		if _, err = tx.CreateWorkerNATSSession(ctx, db.CreateWorkerNATSSessionParams{
			WorkerID:             worker.ID,
			BootstrapTokenID:     token.ID,
			NatsAccountPublicKey: creds.AccountPublicKey,
			NatsUserPublicKey:    creds.UserPublicKey,
			NatsUserJwt:          creds.UserJWT,
			Permissions:          permissionsJSON,
		}); err != nil {
			return fmt.Errorf("create nats session: %w", err)
		}

		registration = RegisterWorkerResponse{
			Worker:     worker,
			RevisionID: revisionID,
			NATS: NATSCredentialsResponse{
				URL:         s.natsURL,
				CAFile:      s.natsCAFile,
				UserJWT:     creds.UserJWT,
				UserSeed:    creds.UserSeed,
				Credentials: string(creds.Creds),
			},
		}
		return nil
	})
	if err != nil {
		return RegisterWorkerResponse{}, err
	}
	return registration, nil
}

func (s *Service) resolveRegistrationRevisionIDWithTx(ctx context.Context, tx RegistrationTx, schemaID int32) (int32, error) {
	revisions, err := tx.ListWorkerSettingsRevisionsBySchemaID(ctx, schemaID)
	if err != nil {
		return 0, fmt.Errorf("list worker settings revisions: %w", err)
	}
	if len(revisions) > 0 {
		return revisions[0].ID, nil
	}

	revision, err := tx.CreateWorkerSettingsRevision(ctx, db.CreateWorkerSettingsRevisionParams{
		WorkerSettingsSchemaID: schemaID,
		CreatedByUserID:        pgtype.Int4{},
		SettingsData:           []byte(`{}`),
	})
	if err != nil {
		return 0, fmt.Errorf("create default worker settings revision: %w", err)
	}
	return revision.ID, nil
}

func (s *Service) resolveRegistrationRevisionID(ctx context.Context, schemaID int32) (int32, error) {
	revisions, err := s.store.ListWorkerSettingsRevisionsBySchemaID(ctx, schemaID)
	if err != nil {
		return 0, fmt.Errorf("list worker settings revisions: %w", err)
	}
	if len(revisions) > 0 {
		return revisions[0].ID, nil
	}

	revision, err := s.store.CreateWorkerSettingsRevision(ctx, db.CreateWorkerSettingsRevisionParams{
		WorkerSettingsSchemaID: schemaID,
		CreatedByUserID:        pgtype.Int4{},
		SettingsData:           []byte(`{}`),
	})
	if err != nil {
		return 0, fmt.Errorf("create default worker settings revision: %w", err)
	}
	return revision.ID, nil
}

// ListWorkers возвращает список воркеров с вычисленным статусом.
func (s *Service) CreateWorkerBootstrapToken(ctx context.Context, orgID, userID int32, req CreateWorkerBootstrapTokenRequest) (CreateWorkerBootstrapTokenResponse, error) {
	plaintext, hash, err := GenerateBootstrapToken()
	if err != nil {
		return CreateWorkerBootstrapTokenResponse{}, fmt.Errorf("generate bootstrap token: %w", err)
	}

	var expiresAt pgtype.Timestamp
	if req.ExpiresAt != nil {
		expiresAt = pgtype.Timestamp{Time: *req.ExpiresAt, Valid: true}
	}
	token, err := s.store.CreateWorkerBootstrapToken(ctx, db.CreateWorkerBootstrapTokenParams{
		OrganizationID:   orgID,
		WorkTypeID:       req.WorkTypeID,
		Name:             req.Name,
		Description:      pgtype.Text{String: req.Description, Valid: req.Description != ""},
		TokenHash:        hash,
		MaxActiveWorkers: req.MaxActiveWorkers,
		ExpiresAt:        expiresAt,
		CreatedByUserID:  pgtype.Int4{Int32: userID, Valid: userID > 0},
	})
	if err != nil {
		return CreateWorkerBootstrapTokenResponse{}, fmt.Errorf("create worker bootstrap token: %w", err)
	}
	return CreateWorkerBootstrapTokenResponse{
		Token:     workerBootstrapTokenResponse(token, 0),
		Plaintext: plaintext,
	}, nil
}

func (s *Service) ListWorkerBootstrapTokens(ctx context.Context, orgID int32) ([]WorkerBootstrapTokenResponse, error) {
	tokens, err := s.store.ListWorkerBootstrapTokensByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list worker bootstrap tokens: %w", err)
	}
	result := make([]WorkerBootstrapTokenResponse, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, workerBootstrapTokenResponse(db.WorkerBootstrapToken{
			ID:                     token.ID,
			OrganizationID:         token.OrganizationID,
			WorkTypeID:             token.WorkTypeID,
			Name:                   token.Name,
			Description:            token.Description,
			TokenHash:              token.TokenHash,
			Status:                 token.Status,
			MaxActiveWorkers:       token.MaxActiveWorkers,
			TotalRegistrationCount: token.TotalRegistrationCount,
			ExpiresAt:              token.ExpiresAt,
			LastUsedAt:             token.LastUsedAt,
			CreatedByUserID:        token.CreatedByUserID,
			RevokedAt:              token.RevokedAt,
			RevokedByUserID:        token.RevokedByUserID,
			CreatedAt:              token.CreatedAt,
			UpdatedAt:              token.UpdatedAt,
			DeletedAt:              token.DeletedAt,
		}, token.ActiveWorkerCount))
	}
	return result, nil
}

func (s *Service) RevokeWorkerBootstrapToken(ctx context.Context, tokenID, userID int32, revokeActive bool) error {
	if _, err := s.store.RevokeWorkerBootstrapToken(ctx, db.RevokeWorkerBootstrapTokenParams{
		ID:              tokenID,
		RevokedByUserID: pgtype.Int4{Int32: userID, Valid: userID > 0},
	}); err != nil {
		return fmt.Errorf("revoke worker bootstrap token: %w", err)
	}
	if !revokeActive {
		return nil
	}

	sessions, err := s.store.ListActiveWorkerNATSSessionsByBootstrapToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("list active nats sessions: %w", err)
	}
	for _, session := range sessions {
		if err := s.natsAuth.RevokeWorker(ctx, session.NatsUserPublicKey); err != nil {
			return fmt.Errorf("revoke nats user %s: %w", session.NatsUserPublicKey, err)
		}
	}
	if _, err := s.store.RevokeWorkerNATSSessionsByBootstrapToken(ctx, db.RevokeWorkerNATSSessionsByBootstrapTokenParams{
		BootstrapTokenID: tokenID,
		RevokedByUserID:  pgtype.Int4{Int32: userID, Valid: userID > 0},
	}); err != nil {
		return fmt.Errorf("revoke nats sessions: %w", err)
	}
	return nil
}

func workerBootstrapTokenResponse(token db.WorkerBootstrapToken, activeWorkerCount int32) WorkerBootstrapTokenResponse {
	resp := WorkerBootstrapTokenResponse{
		ID:                     token.ID,
		OrganizationID:         token.OrganizationID,
		WorkTypeID:             token.WorkTypeID,
		Name:                   token.Name,
		Status:                 token.Status,
		MaxActiveWorkers:       token.MaxActiveWorkers,
		ActiveWorkerCount:      activeWorkerCount,
		TotalRegistrationCount: token.TotalRegistrationCount,
	}
	if token.Description.Valid {
		resp.Description = token.Description.String
	}
	if token.ExpiresAt.Valid {
		resp.ExpiresAt = &token.ExpiresAt.Time
	}
	if token.LastUsedAt.Valid {
		resp.LastUsedAt = &token.LastUsedAt.Time
	}
	if token.CreatedAt.Valid {
		resp.CreatedAt = token.CreatedAt.Time
	}
	if token.RevokedAt.Valid {
		resp.RevokedAt = &token.RevokedAt.Time
	}
	return resp
}

func (s *Service) ListWorkers(ctx context.Context, workTypeID int32) ([]WorkerResponse, error) {
	workers, err := s.store.ListNewWorkersByWorkTypeID(ctx, workTypeID)
	if err != nil {
		return nil, fmt.Errorf("list workers: %w", err)
	}

	result := make([]WorkerResponse, 0, len(workers))
	for _, w := range workers {
		var lastHB time.Time
		if w.LastHeartbeatAt.Valid {
			lastHB = w.LastHeartbeatAt.Time
		}
		result = append(result, WorkerResponse{
			ID:              w.ID,
			Name:            w.Name,
			WorkTypeID:      w.WorkTypeID,
			SchemaID:        w.WorkerSettingsSchemaID,
			Status:          ComputeWorkerStatus(lastHB, false),
			LastHeartbeatAt: lastHB,
		})
	}
	return result, nil
}

func (s *Service) DeleteWorker(ctx context.Context, workerID int32) (DeleteWorkerResponse, error) {
	worker, err := s.store.GetNewWorkerByID(ctx, workerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeleteWorkerResponse{
				Result:   DeleteWorkerResultNotFound,
				Deleted:  false,
				Message:  "Worker not found",
				WorkerID: workerID,
			}, nil
		}
		return DeleteWorkerResponse{}, fmt.Errorf("get worker: %w", err)
	}

	var lastHB time.Time
	if worker.LastHeartbeatAt.Valid {
		lastHB = worker.LastHeartbeatAt.Time
	}
	status := ComputeWorkerStatus(lastHB, false)
	if status != WorkerStatusOffline {
		return DeleteWorkerResponse{
			Result:   DeleteWorkerResultOnline,
			Deleted:  false,
			Message:  "Worker is online",
			WorkerID: worker.ID,
			Status:   status,
		}, nil
	}

	rows, err := s.store.ListWorkflowUsagesByWorkerID(ctx, worker.ID)
	if err != nil {
		return DeleteWorkerResponse{}, fmt.Errorf("list worker workflow usages: %w", err)
	}
	if len(rows) > 0 {
		usages := make([]WorkerWorkflowUsage, 0, len(rows))
		for _, row := range rows {
			usages = append(usages, WorkerWorkflowUsage{
				WorkflowID:            row.WorkflowID,
				WorkflowName:          row.WorkflowName,
				SystemID:              row.SystemID,
				WorkflowVersionID:     row.WorkflowVersionID,
				WorkflowVersionNumber: row.WorkflowVersionNumber,
			})
		}
		return DeleteWorkerResponse{
			Result:   DeleteWorkerResultInUse,
			Deleted:  false,
			Message:  "Worker is used by workflow definitions",
			WorkerID: worker.ID,
			Status:   status,
			Usages:   usages,
		}, nil
	}

	if err := s.store.DeleteWorker(ctx, worker.ID); err != nil {
		return DeleteWorkerResponse{}, fmt.Errorf("delete worker: %w", err)
	}
	return DeleteWorkerResponse{
		Result:   DeleteWorkerResultDeleted,
		Deleted:  true,
		Message:  "Worker deleted",
		WorkerID: worker.ID,
		Status:   status,
	}, nil
}

// --- System methods ---

// CreateSystem создаёт систему в организации.
func (s *Service) CreateSystem(ctx context.Context, orgID int32, req CreateSystemRequest) (db.System, error) {
	return s.store.CreateSystem(ctx, db.CreateSystemParams{
		OrganizationID: pgtype.Int4{Int32: orgID, Valid: true},
		UserCreatorID:  pgtype.Int4{},
		Name:           req.Name,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		IsActive:       true,
		Priority:       req.Priority,
		PublicToken:    pgtype.Text{},
		PrivateToken:   pgtype.Text{},
	})
}

// UpdateSystem обновляет данные системы.
func (s *Service) UpdateSystem(ctx context.Context, id int32, req UpdateSystemRequest) (db.System, error) {
	return s.store.UpdateSystem(ctx, db.UpdateSystemParams{
		ID:          id,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		IsActive:    req.IsActive,
	})
}

// DeleteSystem мягко удаляет систему.
func (s *Service) DeleteSystem(ctx context.Context, id int32) error {
	return s.store.SoftDeleteSystem(ctx, id)
}

// ListSystems возвращает системы организации с количеством активных токенов.
func (s *Service) ListSystems(ctx context.Context, orgID int32) ([]db.ListSystemsByOrganizationIDRow, error) {
	return s.store.ListSystemsByOrganizationID(ctx, pgtype.Int4{Int32: orgID, Valid: true})
}

// ListSystemsPaginated возвращает страницу систем организации с количеством активных токенов.
func (s *Service) ListSystemsPaginated(ctx context.Context, orgID int32, page, perPage int) ([]db.ListSystemsByOrganizationIDPaginatedRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	orgPg := pgtype.Int4{Int32: orgID, Valid: true}
	systems, err := s.store.ListSystemsByOrganizationIDPaginated(ctx, db.ListSystemsByOrganizationIDPaginatedParams{
		OrganizationID: orgPg,
		Limit:          int32(perPage),              // #nosec G115 -- perPage is bounded above.
		Offset:         int32((page - 1) * perPage), // #nosec G115 -- page/perPage are bounded by pagination rules.
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list systems: %w", err)
	}

	total, err := s.store.CountSystemsByOrganizationID(ctx, orgPg)
	if err != nil {
		return nil, 0, fmt.Errorf("count systems: %w", err)
	}

	return systems, total, nil
}

// --- System Token methods ---

// CreateSystemToken генерирует public + private токен, хранит хэш private токена.
// Возвращает plaintext public + private ОДИН раз.
// Per D-05/D-18: вся криптологическая логика в сервисе.
func (s *Service) ListSystemTokens(ctx context.Context, systemID int32) ([]SystemTokenResponse, error) {
	tokens, err := s.store.ListSystemTokensBySystemID(ctx, systemID)
	if err != nil {
		return nil, err
	}
	result := make([]SystemTokenResponse, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, SystemTokenResponse{
			ID:          token.ID,
			SystemID:    token.SystemID,
			Name:        token.Name,
			PublicToken: token.PublicToken,
			IsActive:    token.IsActive,
			CreatedAt:   token.CreatedAt,
			UpdatedAt:   token.UpdatedAt,
		})
	}
	return result, nil
}

func (s *Service) CreateSystemToken(ctx context.Context, systemID int32, name string) (CreateSystemTokenResponse, error) {
	// Генерируем public token (16 random bytes → hex)
	pubBytes := make([]byte, 16)
	if _, err := rand.Read(pubBytes); err != nil {
		return CreateSystemTokenResponse{}, fmt.Errorf("generate public token: %w", err)
	}
	publicToken := hex.EncodeToString(pubBytes)

	// Генерируем private token (32 random bytes → hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return CreateSystemTokenResponse{}, fmt.Errorf("generate private token: %w", err)
	}
	privateTokenPlain := hex.EncodeToString(b)

	// Хэшируем private token для хранения
	privateTokenHash := hashPrivateToken(privateTokenPlain)

	token, err := s.store.CreateSystemToken(ctx, db.CreateSystemTokenParams{
		SystemID:     systemID,
		Name:         name,
		PublicToken:  publicToken,
		PrivateToken: privateTokenHash,
		IsActive:     true,
	})
	if err != nil {
		return CreateSystemTokenResponse{}, fmt.Errorf("create system token: %w", err)
	}

	return CreateSystemTokenResponse{
		ID:           token.ID,
		SystemID:     token.SystemID,
		Name:         name,
		PublicToken:  publicToken,
		PrivateToken: privateTokenPlain,
	}, nil
}

// DeactivateSystemToken деактивирует токен системы.
func (s *Service) DeactivateSystemToken(ctx context.Context, tokenID int32) error {
	return s.store.DeactivateSystemToken(ctx, tokenID)
}

// ActivateSystemToken активирует токен системы.
func (s *Service) ActivateSystemToken(ctx context.Context, tokenID int32) error {
	return s.store.ActivateSystemToken(ctx, tokenID)
}

// BindWorkflowToToken привязывает workflow к токену (RBAC-08).
func (s *Service) BindWorkflowToToken(ctx context.Context, tokenID, workflowID int32) error {
	if err := s.ensureTokenWorkflowSameSystem(ctx, tokenID, workflowID); err != nil {
		return err
	}
	return s.store.GrantWorkflowToken(ctx, db.GrantWorkflowTokenParams{
		SystemTokenID: tokenID,
		WorkflowID:    workflowID,
	})
}

// UnbindWorkflowFromToken отвязывает workflow от токена.
func (s *Service) UnbindWorkflowFromToken(ctx context.Context, tokenID, workflowID int32) error {
	if err := s.ensureTokenWorkflowSameSystem(ctx, tokenID, workflowID); err != nil {
		return err
	}
	return s.store.RevokeWorkflowToken(ctx, db.RevokeWorkflowTokenParams{
		SystemTokenID: tokenID,
		WorkflowID:    workflowID,
	})
}

func (s *Service) ListTokenWorkflows(ctx context.Context, tokenID int32) ([]db.WorkflowToken, error) {
	links, err := s.store.ListWorkflowTokensBySystemTokenID(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return []db.WorkflowToken{}, nil
	}
	return links, nil
}

func (s *Service) ListWorkflowTokens(ctx context.Context, workflowID int32) ([]db.WorkflowToken, error) {
	links, err := s.store.ListWorkflowTokensByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return []db.WorkflowToken{}, nil
	}
	return links, nil
}

func (s *Service) ensureTokenWorkflowSameSystem(ctx context.Context, tokenID, workflowID int32) error {
	token, err := s.store.GetSystemTokenByID(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("get system token: %w", err)
	}
	workflow, err := s.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("get workflow: %w", err)
	}
	if token.SystemID != workflow.SystemID {
		return ErrWorkflowTokenSystemMismatch
	}
	return nil
}

// --- Settings Schema methods ---

// GetSettingsSchema возвращает схему настроек воркера по ID.
func (s *Service) GetSettingsSchema(ctx context.Context, schemaID int32) (db.WorkerSettingsSchema, error) {
	return s.store.GetWorkerSettingsSchemaByID(ctx, schemaID)
}

// --- Settings Revision methods ---

// CreateSettingsRevision создаёт ревизию настроек для схемы воркера.
func (s *Service) CreateSettingsRevision(ctx context.Context, schemaID int32, req CreateRevisionRequest) (db.WorkerSettingsRevision, error) {
	return s.store.CreateWorkerSettingsRevision(ctx, db.CreateWorkerSettingsRevisionParams{
		WorkerSettingsSchemaID: schemaID,
		CreatedByUserID:        pgtype.Int4{},
		SettingsData:           req.SettingsData,
	})
}

// ListSettingsRevisions возвращает ревизии настроек для схемы.
func (s *Service) ListSettingsRevisions(ctx context.Context, schemaID int32) ([]db.WorkerSettingsRevision, error) {
	return s.store.ListWorkerSettingsRevisionsBySchemaID(ctx, schemaID)
}

// sha256hex вычисляет SHA256 и возвращает hex строку.
func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
