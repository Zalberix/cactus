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

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

var ErrWorkflowTokenSystemMismatch = errors.New("system token and workflow belong to different systems")

// HeartbeatTimeout — таймаут для определения offline-статуса воркера.
const HeartbeatTimeout = 90 * time.Second

// Service — бизнес-логика управления типами работ, воркерами, системами.
type Service struct {
	store Storage
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
func NewService(store Storage) *Service {
	return &Service{store: store}
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
// Bootstrap-токен хэшируется SHA256 и хранится в work_type_token.
// Plaintext возвращается один раз.
func (s *Service) CreateWorkType(ctx context.Context, req CreateWorkTypeRequest) (CreateWorkTypeResponse, error) {
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
		return CreateWorkTypeResponse{}, fmt.Errorf("create work type: %w", err)
	}

	plaintext, hash, err := GenerateBootstrapToken()
	if err != nil {
		return CreateWorkTypeResponse{}, fmt.Errorf("generate bootstrap token: %w", err)
	}

	_, err = s.store.CreateWorkTypeToken(ctx, db.CreateWorkTypeTokenParams{
		WorkTypeID: wt.ID,
		TokenHash:  hash,
		IsActive:   true,
	})
	if err != nil {
		return CreateWorkTypeResponse{}, fmt.Errorf("create work type token: %w", err)
	}

	return CreateWorkTypeResponse{
		WorkType:       wt,
		BootstrapToken: plaintext,
	}, nil
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

func (s *Service) ListWorkTypeCatalog(ctx context.Context) ([]WorkTypeCatalogItem, error) {
	workTypes, err := s.store.ListWorkTypes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]WorkTypeCatalogItem, 0, len(workTypes))
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
		item := WorkTypeCatalogItem{
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
// 1. Хэшируем bootstrap_token → ищем work_type_token
// 2. Вычисляем манифест-хэш → ищем/создаём WorkerSettingsSchema
// 3. Ищем/создаём Worker по (work_type_id, name)
// 4. Обновляем heartbeat
func (s *Service) RegisterWorker(ctx context.Context, req RegisterWorkerRequest) (RegisterWorkerResponse, error) {
	// 1. Хэшируем bootstrap token и ищем work_type_token
	tokenHash := sha256hex(req.BootstrapToken)
	wtt, err := s.store.GetActiveWorkTypeTokenByHash(ctx, tokenHash)
	if err != nil {
		return RegisterWorkerResponse{}, fmt.Errorf("invalid bootstrap token: %w", err)
	}

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
	schema, err := s.store.GetWorkerSettingsSchemaByVersion(ctx, db.GetWorkerSettingsSchemaByVersionParams{
		WorkTypeID: wtt.WorkTypeID,
		Version:    manifestHash,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return RegisterWorkerResponse{}, fmt.Errorf("get schema: %w", err)
		}
		// Создаём новую схему
		schema, err = s.store.CreateWorkerSettingsSchema(ctx, db.CreateWorkerSettingsSchemaParams{
			WorkTypeID:     wtt.WorkTypeID,
			Version:        manifestHash,
			SettingsSchema: manifest.SettingsSchema,
			InputSchema:    manifest.InputSchema,
			OutputSchema:   manifest.OutputSchema,
		})
		if err != nil {
			return RegisterWorkerResponse{}, fmt.Errorf("create worker settings schema: %w", err)
		}
	}

	// 4. Ищем или создаём воркер по (work_type_id, name)
	revisionID, err := s.resolveRegistrationRevisionID(ctx, schema.ID)
	if err != nil {
		return RegisterWorkerResponse{}, err
	}

	worker, err := s.store.GetWorkerByWorkTypeAndName(ctx, db.GetWorkerByWorkTypeAndNameParams{
		WorkTypeID: wtt.WorkTypeID,
		Name:       req.Name,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return RegisterWorkerResponse{}, fmt.Errorf("get worker: %w", err)
		}
		// Создаём нового воркера
		worker, err = s.store.CreateNewWorker(ctx, db.CreateNewWorkerParams{
			WorkTypeID:             wtt.WorkTypeID,
			WorkerSettingsSchemaID: schema.ID,
			Name:                   req.Name,
			Metadata:               []byte("{}"),
		})
		if err != nil {
			return RegisterWorkerResponse{}, fmt.Errorf("create worker: %w", err)
		}
	} else if worker.WorkerSettingsSchemaID != schema.ID {
		// Обновляем схему если манифест изменился
		worker, err = s.store.UpdateNewWorkerSchema(ctx, db.UpdateNewWorkerSchemaParams{
			ID:                     worker.ID,
			WorkerSettingsSchemaID: schema.ID,
		})
		if err != nil {
			return RegisterWorkerResponse{}, fmt.Errorf("update worker schema: %w", err)
		}
	}

	// 5. Обновляем heartbeat (WORK-05 — и для новых, и для повторно регистрирующихся)
	if err = s.store.UpdateNewWorkerHeartbeat(ctx, worker.ID); err != nil {
		return RegisterWorkerResponse{}, fmt.Errorf("update worker heartbeat: %w", err)
	}

	return RegisterWorkerResponse{
		Worker:     worker,
		RevisionID: revisionID,
	}, nil
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
