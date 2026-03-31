package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// ErrValidationRequired — версия не прошла валидацию DAG.
var ErrValidationRequired = errors.New("VALIDATION_REQUIRED: версия должна пройти валидацию DAG перед активацией")

// Service — бизнес-логика управления workflow.
type Service struct {
	store Storage
}

// NewService создаёт новый workflow.Service.
func NewService(store Storage) *Service {
	return &Service{store: store}
}

// --- Workflow CRUD ---

// CreateWorkflow создаёт новый workflow в системе (WF-01).
func (s *Service) CreateWorkflow(ctx context.Context, systemID int32, req CreateWorkflowRequest) (db.Workflow, error) {
	return s.store.CreateWorkflow(ctx, db.CreateWorkflowParams{
		SystemID:    systemID,
		Name:        req.Name,
		Priority:    int32(req.Priority),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// ListWorkflows возвращает список workflow системы.
func (s *Service) ListWorkflows(ctx context.Context, systemID int32) ([]db.Workflow, error) {
	return s.store.ListWorkflowsBySystemID(ctx, systemID)
}

// GetWorkflow возвращает workflow по ID.
func (s *Service) GetWorkflow(ctx context.Context, id int32) (db.Workflow, error) {
	return s.store.GetWorkflowByID(ctx, id)
}

// UpdateWorkflow обновляет название и приоритет workflow.
func (s *Service) UpdateWorkflow(ctx context.Context, id int32, req UpdateWorkflowRequest) (db.Workflow, error) {
	return s.store.UpdateWorkflow(ctx, db.UpdateWorkflowParams{
		ID:          id,
		Name:        req.Name,
		Priority:    int32(req.Priority),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// DeleteWorkflow мягко удаляет workflow.
func (s *Service) DeleteWorkflow(ctx context.Context, id int32) error {
	return s.store.SoftDeleteWorkflow(ctx, id)
}

// --- Version management ---

// CreateVersion создаёт новую версию workflow (WF-02).
// Номер версии автоматически инкрементируется.
func (s *Service) CreateVersion(ctx context.Context, workflowID, userID int32) (db.WorkflowVersion, error) {
	maxVersion, err := s.store.GetMaxVersionNumberByWorkflowID(ctx, workflowID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("get max version number: %w", err)
	}

	return s.store.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
		WorkflowID:      workflowID,
		CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
		VersionNumber:   maxVersion + 1,
		IsValid:         false,
		IsActive:        false,
		TrafficWeight:   100,
		IsControlGroup:  false,
	})
}

// ListVersions возвращает все версии workflow.
func (s *Service) ListVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error) {
	return s.store.ListWorkflowVersionsByWorkflowID(ctx, workflowID)
}

// --- Step CRUD ---

// CreateStep создаёт новый шаг в версии workflow (WF-03).
// Если step_type=task и work_type_id указан, но worker_settings_revision_id нет —
// автоматически находит последнюю ревизию для данного work_type.
func (s *Service) CreateStep(ctx context.Context, versionID int32, req CreateStepRequest) (db.WorkflowStep, error) {
	params := db.CreateWorkflowStepParams{
		WorkflowVersionID: versionID,
		StepType:          req.StepType,
		InputMapping:      req.InputMapping,
		ControlSettings:   req.ControlSettings,
	}
	if req.WorkTypeID != nil {
		params.WorkTypeID = pgtype.Int4{Int32: *req.WorkTypeID, Valid: true}
	}
	if req.WorkerSettingsRevisionID != nil {
		params.WorkerSettingsRevisionID = pgtype.Int4{Int32: *req.WorkerSettingsRevisionID, Valid: true}
	}
	if req.ControlKind != nil {
		params.ControlKind = pgtype.Text{String: *req.ControlKind, Valid: true}
	}

	// Автоподстановка revision для task-шагов
	if req.StepType == "task" && req.WorkTypeID != nil && req.WorkerSettingsRevisionID == nil {
		revID, err := s.resolveLatestRevision(ctx, *req.WorkTypeID)
		if err != nil {
			return db.WorkflowStep{}, fmt.Errorf("auto-resolve revision: %w", err)
		}
		params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revID, Valid: true}
	}

	return s.store.CreateWorkflowStep(ctx, params)
}

// resolveLatestRevision находит последнюю ревизию настроек для work_type.
func (s *Service) resolveLatestRevision(ctx context.Context, workTypeID int32) (int32, error) {
	schemas, err := s.store.ListWorkerSettingsSchemasByWorkTypeID(ctx, workTypeID)
	if err != nil {
		return 0, fmt.Errorf("list schemas: %w", err)
	}
	if len(schemas) == 0 {
		return 0, fmt.Errorf("no settings schema found for work_type_id=%d", workTypeID)
	}
	revisions, err := s.store.ListWorkerSettingsRevisionsBySchemaID(ctx, schemas[0].ID)
	if err != nil {
		return 0, fmt.Errorf("list revisions: %w", err)
	}
	if len(revisions) == 0 {
		return 0, fmt.Errorf("no settings revision found for schema_id=%d", schemas[0].ID)
	}
	return revisions[0].ID, nil
}

// ListSteps возвращает все шаги версии workflow.
func (s *Service) ListSteps(ctx context.Context, versionID int32) ([]db.WorkflowStep, error) {
	return s.store.ListWorkflowStepsByVersionID(ctx, versionID)
}

// UpdateStep обновляет шаг (включая input_mapping — WF-05).
func (s *Service) UpdateStep(ctx context.Context, stepID int32, req UpdateStepRequest) (db.WorkflowStep, error) {
	params := db.UpdateWorkflowStepParams{
		ID:              stepID,
		StepType:        req.StepType,
		InputMapping:    req.InputMapping,
		ControlSettings: req.ControlSettings,
	}
	if req.WorkTypeID != nil {
		params.WorkTypeID = pgtype.Int4{Int32: *req.WorkTypeID, Valid: true}
	}
	if req.WorkerSettingsRevisionID != nil {
		params.WorkerSettingsRevisionID = pgtype.Int4{Int32: *req.WorkerSettingsRevisionID, Valid: true}
	}
	if req.ControlKind != nil {
		params.ControlKind = pgtype.Text{String: *req.ControlKind, Valid: true}
	}
	return s.store.UpdateWorkflowStep(ctx, params)
}

// DeleteStep мягко удаляет шаг и каскадно удаляет его зависимости.
func (s *Service) DeleteStep(ctx context.Context, stepID int32) error {
	// Удаляем зависимости шага перед его удалением
	if err := s.store.DeleteDependenciesByStepID(ctx, stepID); err != nil {
		return fmt.Errorf("delete step dependencies: %w", err)
	}
	return s.store.SoftDeleteWorkflowStep(ctx, stepID)
}

// ListDependencies возвращает все зависимости версии workflow.
func (s *Service) ListDependencies(ctx context.Context, versionID int32) ([]db.WorkflowStepDependency, error) {
	return s.store.ListDependenciesByVersionID(ctx, versionID)
}

// --- Dependency management ---

// CreateDependency создаёт зависимость между шагами (WF-04).
func (s *Service) CreateDependency(ctx context.Context, stepID int32, req CreateDependencyRequest) error {
	return s.store.CreateWorkflowStepDependency(ctx, db.CreateWorkflowStepDependencyParams{
		StepID:          stepID,
		DependsOnStepID: req.DependsOnStepID,
		Outcome:         pgtype.Text{String: req.Outcome, Valid: true},
	})
}

// DeleteDependency удаляет зависимость между шагами.
func (s *Service) DeleteDependency(ctx context.Context, stepID, dependsOnStepID int32) error {
	return s.store.DeleteWorkflowStepDependency(ctx, db.DeleteWorkflowStepDependencyParams{
		StepID:          stepID,
		DependsOnStepID: dependsOnStepID,
	})
}

// --- Validation ---

// ValidateVersion запускает валидацию DAG для версии (WF-06, WF-07).
// Загружает шаги и зависимости, вызывает dag.ValidateDAG,
// обновляет is_valid у версии.
// Возвращает результат валидации со всеми ошибками сразу (D-12).
func (s *Service) ValidateVersion(ctx context.Context, versionID int32) (ValidateVersionResponse, error) {
	dbSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list steps: %w", err)
	}

	dbDeps, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list dependencies: %w", err)
	}

	// Конвертируем в типы dag пакета
	dagSteps := make([]dagpkg.Step, 0, len(dbSteps))
	for _, step := range dbSteps {
		dagStep := dagpkg.Step{
			ID:       step.ID,
			StepType: dagpkg.StepType(step.StepType),
		}
		// Парсим input_mapping из JSONB
		if len(step.InputMapping) > 0 {
			var mappings []dagpkg.MappingEntry
			if err := json.Unmarshal(step.InputMapping, &mappings); err == nil {
				dagStep.InputMapping = mappings
			}
		}
		dagSteps = append(dagSteps, dagStep)
	}

	dagDeps := make([]dagpkg.Dependency, 0, len(dbDeps))
	for _, dep := range dbDeps {
		outcome := "success"
		if dep.Outcome.Valid {
			outcome = dep.Outcome.String
		}
		dagDeps = append(dagDeps, dagpkg.Dependency{
			StepID:          dep.StepID,
			DependsOnStepID: dep.DependsOnStepID,
			Outcome:         outcome,
		})
	}

	// Валидируем DAG
	validationErrors := dagpkg.ValidateDAG(dagSteps, dagDeps)
	isValid := len(validationErrors) == 0

	// Обновляем is_valid у версии
	if _, err := s.store.UpdateWorkflowVersionValid(ctx, db.UpdateWorkflowVersionValidParams{
		ID:      versionID,
		IsValid: isValid,
	}); err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("update version valid: %w", err)
	}

	return ValidateVersionResponse{
		IsValid: isValid,
		Errors:  validationErrors,
	}, nil
}

// --- Activation ---

// ActivateVersion активирует версию workflow (WF-08).
// Гард: версия должна быть is_valid=true (Pitfall 3).
func (s *Service) ActivateVersion(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	// Гард: нельзя активировать невалидную версию
	if !version.IsValid {
		return ErrValidationRequired
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: true,
	}); err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	// WF-09: пересчитываем input_validation
	return s.regenerateInputValidation(ctx, version.WorkflowID)
}

// DeactivateVersion деактивирует версию workflow (WF-08).
func (s *Service) DeactivateVersion(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: false,
	}); err != nil {
		return fmt.Errorf("deactivate version: %w", err)
	}

	// WF-09: пересчитываем input_validation
	return s.regenerateInputValidation(ctx, version.WorkflowID)
}

// regenerateInputValidation пересчитывает input_validation workflow (WF-09).
// Итерирует ВСЕ активные версии (Pitfall 5), собирает union JSON Schema
// из всех $.message.value.* полей в input_mapping.
func (s *Service) regenerateInputValidation(ctx context.Context, workflowID int32) error {
	activeVersions, err := s.store.ListActiveWorkflowVersions(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("list active versions: %w", err)
	}

	if len(activeVersions) == 0 {
		// Нет активных версий — сбрасываем input_validation
		_, err = s.store.UpdateWorkflowInputValidation(ctx, db.UpdateWorkflowInputValidationParams{
			ID:              workflowID,
			InputValidation: nil,
		})
		return err
	}

	// Собираем уникальные поля из всех активных версий
	fieldSet := make(map[string]struct{})
	for _, version := range activeVersions {
		steps, err := s.store.ListWorkflowStepsByVersionID(ctx, version.ID)
		if err != nil {
			return fmt.Errorf("list steps for version %d: %w", version.ID, err)
		}
		for _, step := range steps {
			if len(step.InputMapping) == 0 {
				continue
			}
			var mappings []dagpkg.MappingEntry
			if err := json.Unmarshal(step.InputMapping, &mappings); err != nil {
				continue
			}
			for _, m := range mappings {
				if strings.HasPrefix(m.Source, "$.message.value.") {
					field := strings.TrimPrefix(m.Source, "$.message.value.")
					// Берём только первый сегмент пути (без вложенности для v1)
					if dotIdx := strings.Index(field, "."); dotIdx != -1 {
						field = field[:dotIdx]
					}
					if field != "" {
						fieldSet[field] = struct{}{}
					}
				}
			}
		}
	}

	// Строим JSON Schema
	properties := make(map[string]interface{}, len(fieldSet))
	required := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		properties[field] = map[string]string{"type": "string"}
		required = append(required, field)
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("marshal input_validation schema: %w", err)
	}

	_, err = s.store.UpdateWorkflowInputValidation(ctx, db.UpdateWorkflowInputValidationParams{
		ID:              workflowID,
		InputValidation: schemaJSON,
	})
	return err
}
