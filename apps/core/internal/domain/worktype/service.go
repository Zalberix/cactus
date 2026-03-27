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

// HeartbeatTimeout — таймаут для определения offline-статуса воркера.
const HeartbeatTimeout = 90 * time.Second

// Service — бизнес-логика управления типами работ, воркерами, системами.
type Service struct {
	store Storage
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
func (s *Service) ListWorkTypes(ctx context.Context) ([]db.WorkType, error) {
	return s.store.ListWorkTypes(ctx)
}

// --- Worker methods ---

// RegisterWorker регистрирует воркер по bootstrap-токену и манифесту.
// 1. Хэшируем bootstrap_token → ищем work_type_token
// 2. Вычисляем манифест-хэш → ищем/создаём WorkerSettingsSchema
// 3. Ищем/создаём Worker по (work_type_id, name)
// 4. Обновляем heartbeat (WORK-05)
func (s *Service) RegisterWorker(ctx context.Context, req RegisterWorkerRequest) (db.Worker, error) {
	// 1. Хэшируем bootstrap token и ищем work_type_token
	tokenHash := sha256hex(req.BootstrapToken)
	wtt, err := s.store.GetActiveWorkTypeTokenByHash(ctx, tokenHash)
	if err != nil {
		return db.Worker{}, fmt.Errorf("invalid bootstrap token: %w", err)
	}

	// 2. Вычисляем манифест хэш (используется как version)
	manifestHash, err := ManifestHash(req.Manifest)
	if err != nil {
		return db.Worker{}, fmt.Errorf("compute manifest hash: %w", err)
	}

	// 3. Ищем или создаём WorkerSettingsSchema
	schema, err := s.store.GetWorkerSettingsSchemaByVersion(ctx, db.GetWorkerSettingsSchemaByVersionParams{
		WorkTypeID: wtt.WorkTypeID,
		Version:    manifestHash,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return db.Worker{}, fmt.Errorf("get schema: %w", err)
		}
		// Создаём новую схему
		schema, err = s.store.CreateWorkerSettingsSchema(ctx, db.CreateWorkerSettingsSchemaParams{
			WorkTypeID:     wtt.WorkTypeID,
			Version:        manifestHash,
			SettingsSchema: req.Manifest,
			InputSchema:    []byte("{}"),
			OutputSchema:   []byte("{}"),
		})
		if err != nil {
			return db.Worker{}, fmt.Errorf("create worker settings schema: %w", err)
		}
	}

	// 4. Ищем или создаём воркер по (work_type_id, name)
	worker, err := s.store.GetWorkerByWorkTypeAndName(ctx, db.GetWorkerByWorkTypeAndNameParams{
		WorkTypeID: wtt.WorkTypeID,
		Name:       req.Name,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return db.Worker{}, fmt.Errorf("get worker: %w", err)
		}
		// Создаём нового воркера
		worker, err = s.store.CreateNewWorker(ctx, db.CreateNewWorkerParams{
			WorkTypeID:             wtt.WorkTypeID,
			WorkerSettingsSchemaID: schema.ID,
			Name:                   req.Name,
			Metadata:               []byte("{}"),
		})
		if err != nil {
			return db.Worker{}, fmt.Errorf("create worker: %w", err)
		}
	} else {
		// Обновляем схему если манифест изменился
		if worker.WorkerSettingsSchemaID != schema.ID {
			worker, err = s.store.UpdateNewWorkerSchema(ctx, db.UpdateNewWorkerSchemaParams{
				ID:                     worker.ID,
				WorkerSettingsSchemaID: schema.ID,
			})
			if err != nil {
				return db.Worker{}, fmt.Errorf("update worker schema: %w", err)
			}
		}
	}

	// 5. Обновляем heartbeat (WORK-05 — и для новых, и для повторно регистрирующихся)
	if err = s.store.UpdateNewWorkerHeartbeat(ctx, worker.ID); err != nil {
		return db.Worker{}, fmt.Errorf("update worker heartbeat: %w", err)
	}

	return worker, nil
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

// ListSystems возвращает системы организации.
func (s *Service) ListSystems(ctx context.Context, orgID int32) ([]db.System, error) {
	return s.store.ListSystemsByOrganizationID(ctx, pgtype.Int4{Int32: orgID, Valid: true})
}

// --- System Token methods ---

// CreateSystemToken генерирует public + private токен, хранит хэш private токена.
// Возвращает plaintext public + private ОДИН раз.
// Per D-05/D-18: вся криптологическая логика в сервисе.
func (s *Service) CreateSystemToken(ctx context.Context, systemID int32) (CreateSystemTokenResponse, error) {
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
		PublicToken:  publicToken,
		PrivateToken: privateTokenPlain,
	}, nil
}

// DeactivateSystemToken деактивирует токен системы.
func (s *Service) DeactivateSystemToken(ctx context.Context, tokenID int32) error {
	return s.store.DeactivateSystemToken(ctx, tokenID)
}

// BindWorkflowToToken привязывает workflow к токену (RBAC-08).
func (s *Service) BindWorkflowToToken(ctx context.Context, tokenID, workflowID int32) error {
	return s.store.GrantWorkflowToken(ctx, db.GrantWorkflowTokenParams{
		SystemTokenID: tokenID,
		WorkflowID:    workflowID,
	})
}

// UnbindWorkflowFromToken отвязывает workflow от токена.
func (s *Service) UnbindWorkflowFromToken(ctx context.Context, tokenID, workflowID int32) error {
	return s.store.RevokeWorkflowToken(ctx, db.RevokeWorkflowTokenParams{
		SystemTokenID: tokenID,
		WorkflowID:    workflowID,
	})
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
