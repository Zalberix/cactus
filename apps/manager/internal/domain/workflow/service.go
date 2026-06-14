package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	dagpkg "github.com/zalberix/cactus/apps/manager/internal/dag"
	schemadialect "github.com/zalberix/cactus/apps/manager/internal/schema"
	db "github.com/zalberix/cactus/libs/storage/db"
)

// ErrValidationRequired — версия не прошла валидацию DAG.
var ErrValidationRequired = errors.New("VALIDATION_REQUIRED: версия должна пройти валидацию DAG перед активацией")

var ErrStartStepProtected = errors.New("START_STEP_PROTECTED: start step is managed by the system")

var ErrWorkflowVersionLocked = errors.New("WORKFLOW_VERSION_LOCKED: active, locked, or published workflow version cannot be semantically edited")

var ErrWorkflowVersionInUse = errors.New("WORKFLOW_VERSION_IN_USE: workflow version is referenced and cannot be deleted")

var ErrTaskStepRequired = errors.New("TASK_STEP_REQUIRED: step must be a task")

var ErrControlSettingsInvalid = errors.New("CONTROL_SETTINGS_INVALID: invalid control settings")

var ErrControlKindInvalid = errors.New("CONTROL_KIND_INVALID: unsupported control kind")

var ErrInvalidInputMapping = errors.New("INVALID_INPUT_MAPPING")

var ErrInvalidSettings = errors.New("INVALID_SETTINGS")

type workflowVersionReferenceQueries interface {
	IsWorkflowVersionReferenced(ctx context.Context, id int32) (bool, error)
}

type nativeInputSchemaWriter interface {
	CreateWorkflowInputSchema(ctx context.Context, arg db.CreateWorkflowInputSchemaParams) (db.WorkflowInputSchema, error)
	CreateWorkflowVersionInputSchemaCompatibility(ctx context.Context, arg db.CreateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error)
}

type copyVersionQueries interface {
	nativeInputSchemaWriter
	CreateWorkflowVersion(ctx context.Context, arg db.CreateWorkflowVersionParams) (db.WorkflowVersion, error)
	GetMaxVersionNumberByWorkflowID(ctx context.Context, workflowID int32) (int32, error)
	ListWorkflowVersionsByWorkflowID(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error)
	ListWorkflowStepsByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStep, error)
	CloneWorkerSettingsRevision(ctx context.Context, arg db.CloneWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error)
	CreateWorkflowStep(ctx context.Context, arg db.CreateWorkflowStepParams) (db.WorkflowStep, error)
	ListDependenciesByVersionID(ctx context.Context, workflowVersionID int32) ([]db.WorkflowStepDependency, error)
	CreateWorkflowStepDependency(ctx context.Context, arg db.CreateWorkflowStepDependencyParams) error
	GetNativeInputSchemaForWorkflowVersion(ctx context.Context, workflowVersionID int32) (db.WorkflowInputSchema, error)
}

var ErrStepNameRequired = errors.New("STEP_NAME_REQUIRED")

var ErrStepNameInvalid = errors.New("STEP_NAME_INVALID")

var ErrStepNameDuplicate = errors.New("STEP_NAME_DUPLICATE")

var ErrVersionNameDuplicate = errors.New("VERSION_NAME_DUPLICATE")

const maxStepNameLength = 255

const maxVersionNameLength = 255

var allowedControlKinds = map[string]struct{}{
	dagpkg.ControlKindStart:     {},
	dagpkg.ControlKindCondition: {},
	dagpkg.ControlKindSwitch:    {},
	dagpkg.ControlKindDelay:     {},
}

var allowedConditionOperators = map[string]struct{}{
	"eq":           {},
	"neq":          {},
	"gt":           {},
	"gte":          {},
	"lt":           {},
	"lte":          {},
	"contains":     {},
	"not_contains": {},
	"exists":       {},
	"not_exists":   {},
}

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
	if req.Priority > math.MaxInt32 || req.Priority < math.MinInt32 {
		return db.Workflow{}, fmt.Errorf("priority %d overflows int32", req.Priority)
	}
	return s.store.CreateWorkflow(ctx, db.CreateWorkflowParams{
		SystemID:    systemID,
		Name:        req.Name,
		Priority:    int32(req.Priority),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// ListWorkflows возвращает список workflow системы.
func (s *Service) ListWorkflows(ctx context.Context, systemID int32) ([]WorkflowListResponse, error) {
	workflows, err := s.store.ListWorkflowsBySystemID(ctx, systemID)
	if err != nil {
		return nil, err
	}

	responses := make([]WorkflowListResponse, 0, len(workflows))
	for _, wf := range workflows {
		versions, err := s.store.ListWorkflowVersionsByWorkflowID(ctx, wf.ID)
		if err != nil {
			return nil, fmt.Errorf("list workflow versions for workflow %d: %w", wf.ID, err)
		}

		var versionCount int32
		var activeVersionCount int32
		for _, version := range versions {
			if version.DeletedAt.Valid {
				continue
			}
			versionCount++
			if version.IsActive {
				activeVersionCount++
			}
		}

		responses = append(responses, WorkflowListResponse{
			Workflow:           wf,
			VersionCount:       versionCount,
			ActiveVersionCount: activeVersionCount,
		})
	}
	return responses, nil
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
		Priority:    int32(req.Priority), // #nosec G115 -- priority is validated by request binding.
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
	var version db.WorkflowVersion
	err := s.store.WithTx(ctx, func(q *db.Queries) error {
		maxVersion, err := q.GetMaxVersionNumberByWorkflowID(ctx, workflowID)
		if err != nil {
			return fmt.Errorf("get max version number: %w", err)
		}

		var createErr error
		version, createErr = q.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
			WorkflowID:      workflowID,
			CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
			VersionNumber:   maxVersion + 1,
			Name:            pgtype.Text{String: fmt.Sprintf("Version %d", maxVersion+1), Valid: true},
			IsValid:         false,
			IsActive:        false,
		})
		if createErr != nil {
			return fmt.Errorf("create workflow version: %w", createErr)
		}

		_, createErr = q.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
			WorkflowVersionID: version.ID,
			Name:              "System Trigger",
			StepType:          string(dagpkg.StepTypeControl),
			ControlKind:       pgtype.Text{String: dagpkg.ControlKindStart, Valid: true},
			ControlSettings:   []byte(`{"trigger":"system_message"}`),
			CanvasPosition:    []byte(`{"x":80,"y":200}`),
		})
		if createErr != nil {
			return fmt.Errorf("create start step: %w", createErr)
		}

		if _, createErr = createNativeInputSchemaForVersion(ctx, q, version, userID, []byte(`{"type":"object","properties":{}}`)); createErr != nil {
			return createErr
		}

		return nil
	})
	if err != nil {
		return db.WorkflowVersion{}, err
	}
	return version, nil
}

// ListVersions возвращает все версии workflow.
func nativeInputSchemaCode(versionNumber int32) string {
	return fmt.Sprintf("v%d", versionNumber)
}

func createNativeInputSchemaForVersion(ctx context.Context, q nativeInputSchemaWriter, version db.WorkflowVersion, actorUserID int32, schemaJSON []byte) (db.WorkflowInputSchema, error) {
	if len(bytes.TrimSpace(schemaJSON)) == 0 {
		schemaJSON = []byte(`{"type":"object","properties":{}}`)
	}

	schema, err := q.CreateWorkflowInputSchema(ctx, db.CreateWorkflowInputSchemaParams{
		WorkflowID:      version.WorkflowID,
		Code:            nativeInputSchemaCode(version.VersionNumber),
		VersionNumber:   version.VersionNumber,
		SchemaJson:      schemaJSON,
		Status:          "draft",
		IsDefault:       false,
		CreatedByUserID: pgInt4(actorUserID),
		UpdatedByUserID: pgInt4(actorUserID),
	})
	if err != nil {
		return db.WorkflowInputSchema{}, fmt.Errorf("create native input schema: %w", err)
	}

	if _, err := q.CreateWorkflowVersionInputSchemaCompatibility(ctx, db.CreateWorkflowVersionInputSchemaCompatibilityParams{
		WorkflowVersionID:     version.ID,
		WorkflowInputSchemaID: schema.ID,
		CompatibilityType:     "native",
		WorkflowInputMapperID: pgtype.Int4{},
		DefaultValues:         []byte(`{}`),
		IsActive:              true,
		IsDefaultRoute:        true,
		CreatedByUserID:       pgInt4(actorUserID),
		UpdatedByUserID:       pgInt4(actorUserID),
	}); err != nil {
		return db.WorkflowInputSchema{}, fmt.Errorf("create native input schema compatibility: %w", err)
	}

	return schema, nil
}

func (s *Service) ListVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error) {
	return s.store.ListWorkflowVersionsByWorkflowID(ctx, workflowID)
}

// ListVersionSummaries возвращает сводные данные по версиям workflow.
func (s *Service) ListVersionSummaries(ctx context.Context, workflowID int32) ([]VersionSummaryResponse, error) {
	rows, err := s.store.ListWorkflowVersionSummariesByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("list version summaries: %w", err)
	}

	summaries := make([]VersionSummaryResponse, 0, len(rows))
	for _, r := range rows {
		if r.DeletedAt.Valid {
			continue
		}

		summary := VersionSummaryResponse{
			ID:            r.ID,
			WorkflowID:    r.WorkflowID,
			Name:          pgTextString(r.Name),
			VersionNumber: r.VersionNumber,
			IsValid:       r.IsValid,
			IsActive:      r.IsActive,
			RunCount:      r.RunCount,
			CreatedAt:     timestampString(r.CreatedAt),
			UpdatedAt:     timestampString(r.UpdatedAt),
		}
		if r.DeletedAt.Valid {
			summary.DeletedAt = timestampString(r.DeletedAt)
		}
		summaries = append(summaries, summary)
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		if summaries[i].IsActive != summaries[j].IsActive {
			return summaries[i].IsActive
		}
		if summaries[i].VersionNumber != summaries[j].VersionNumber {
			return summaries[i].VersionNumber > summaries[j].VersionNumber
		}
		return summaries[i].ID < summaries[j].ID
	})

	return summaries, nil
}

// UpdateVersionName обновляет название версии.
func (s *Service) UpdateVersionName(ctx context.Context, versionID int32, req UpdateVersionNameRequest) (db.WorkflowVersion, error) {
	name := normalizeVersionName(req.Name)
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("get workflow version: %w", err)
	}

	versions, err := s.store.ListWorkflowVersionsByWorkflowID(ctx, version.WorkflowID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("list workflow versions: %w", err)
	}
	if hasDuplicateVersionName(name, versions, versionID) {
		return db.WorkflowVersion{}, ErrVersionNameDuplicate
	}

	updated, err := s.store.UpdateWorkflowVersionName(ctx, db.UpdateWorkflowVersionNameParams{
		ID:   versionID,
		Name: pgtype.Text{String: name, Valid: true},
	})
	if err != nil {
		if isUniqueConstraintViolation(err, "workflow_version_workflow_name_active_uq") {
			return db.WorkflowVersion{}, ErrVersionNameDuplicate
		}
		return db.WorkflowVersion{}, err
	}
	return updated, nil
}

// CopyVersion создаёт копию версии.
func (s *Service) CopyVersion(ctx context.Context, versionID, userID int32) (db.WorkflowVersion, error) {
	source, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("get source version: %w", err)
	}

	sourceName := source.Name.String
	if !source.Name.Valid || strings.TrimSpace(sourceName) == "" {
		sourceName = fmt.Sprintf("Version %d", source.VersionNumber)
	}

	var newVersion db.WorkflowVersion
	copyWith := func(q copyVersionQueries) error {
		maxVersion, err := q.GetMaxVersionNumberByWorkflowID(ctx, source.WorkflowID)
		if err != nil {
			return fmt.Errorf("get max version number: %w", err)
		}

		versions, err := q.ListWorkflowVersionsByWorkflowID(ctx, source.WorkflowID)
		if err != nil {
			return fmt.Errorf("list workflow versions: %w", err)
		}
		copyName := uniqueVersionName(sourceName+" copy", versions, 0)

		newVersion, err = q.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
			WorkflowID:      source.WorkflowID,
			CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
			VersionNumber:   maxVersion + 1,
			Name:            pgtype.Text{String: copyName, Valid: true},
			IsValid:         false,
			IsActive:        false,
		})
		if err != nil {
			return fmt.Errorf("create copied version: %w", err)
		}

		sourceSteps, err := q.ListWorkflowStepsByVersionID(ctx, versionID)
		if err != nil {
			return fmt.Errorf("list source steps: %w", err)
		}

		stepIDMap := make(map[int32]int32, len(sourceSteps))
		copiedNames := make([]db.WorkflowStep, 0, len(sourceSteps))
		for _, sourceStep := range sourceSteps {
			stepName := normalizeStepName(sourceStep.Name)
			if stepName == "" {
				stepName = defaultStepNameForStep(sourceStep)
			}
			stepName = uniqueStepName(stepName, copiedNames, 0)

			params := db.CreateWorkflowStepParams{
				WorkflowVersionID: newVersion.ID,
				Name:              stepName,
				StepType:          sourceStep.StepType,
				ControlKind:       sourceStep.ControlKind,
				ControlSettings:   sourceStep.ControlSettings,
				InputMapping:      sourceStep.InputMapping,
				CanvasPosition:    sourceStep.CanvasPosition,
			}

			if sourceStep.WorkTypeID.Valid {
				params.WorkTypeID = sourceStep.WorkTypeID
			}
			if sourceStep.WorkerSettingsRevisionID.Valid && sourceStep.StepType == string(dagpkg.StepTypeTask) {
				revision, err := q.CloneWorkerSettingsRevision(ctx, db.CloneWorkerSettingsRevisionParams{
					ID:              sourceStep.WorkerSettingsRevisionID.Int32,
					CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
				})
				if err != nil {
					return fmt.Errorf("clone settings revision for step %d: %w", sourceStep.ID, err)
				}
				params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revision.ID, Valid: true}
			}

			copiedStep, err := q.CreateWorkflowStep(ctx, params)
			if err != nil {
				return fmt.Errorf("copy step %d: %w", sourceStep.ID, err)
			}
			stepIDMap[sourceStep.ID] = copiedStep.ID
			copiedNames = append(copiedNames, db.WorkflowStep{ID: copiedStep.ID, Name: stepName})
		}

		sourceDependencies, err := q.ListDependenciesByVersionID(ctx, versionID)
		if err != nil {
			return fmt.Errorf("list source dependencies: %w", err)
		}
		for _, dep := range sourceDependencies {
			stepID, ok := stepIDMap[dep.StepID]
			if !ok {
				return fmt.Errorf("missing copied step for source step_id=%d", dep.StepID)
			}
			depStepID, ok := stepIDMap[dep.DependsOnStepID]
			if !ok {
				return fmt.Errorf("missing copied step for depends_on_step_id=%d", dep.DependsOnStepID)
			}
			if err := q.CreateWorkflowStepDependency(ctx, db.CreateWorkflowStepDependencyParams{
				StepID:          stepID,
				DependsOnStepID: depStepID,
				Outcome:         dep.Outcome,
				OutputIndex:     dep.OutputIndex,
			}); err != nil {
				return fmt.Errorf("create dependency copy %d->%d: %w", dep.DependsOnStepID, dep.StepID, err)
			}
		}

		sourceSchema, err := q.GetNativeInputSchemaForWorkflowVersion(ctx, source.ID)
		if err != nil {
			sourceSchema.SchemaJson = []byte(`{"type":"object","properties":{}}`)
		}
		if _, err := createNativeInputSchemaForVersion(ctx, q, newVersion, userID, sourceSchema.SchemaJson); err != nil {
			return err
		}
		return nil
	}

	err = s.store.WithTx(ctx, func(q *db.Queries) error {
		if q == nil {
			return nil
		}
		return copyWith(q)
	})
	if err != nil {
		return db.WorkflowVersion{}, err
	}
	if newVersion.ID == 0 {
		if err := copyWith(s.store); err != nil {
			return db.WorkflowVersion{}, err
		}
	}

	return newVersion, nil
}

// --- Step CRUD ---

// CreateStep создаёт новый шаг в версии workflow (WF-03).
// Если step_type=task и work_type_id указан, но worker_settings_revision_id нет —
// автоматически находит последнюю ревизию для данного work_type.
func (s *Service) CreateStep(ctx context.Context, versionID int32, req CreateStepRequest) (db.WorkflowStep, error) {
	if err := s.ensureVersionEditable(ctx, versionID); err != nil {
		return db.WorkflowStep{}, err
	}
	if isStartStepRequest(req.StepType, req.ControlKind) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if err := validateControlKind(req.StepType, req.ControlKind); err != nil {
		return db.WorkflowStep{}, err
	}
	controlSettings := req.ControlSettings
	if req.ControlKind != nil {
		controlSettings = defaultControlSettingsIfEmpty(*req.ControlKind, controlSettings)
		if err := validateControlSettings(*req.ControlKind, controlSettings); err != nil {
			return db.WorkflowStep{}, err
		}
	}

	stepName := normalizeStepName(req.Name)
	if stepName == "" {
		stepName = defaultStepNameForCreate(req)
	}
	if err := validateStepName(stepName); err != nil {
		return db.WorkflowStep{}, err
	}
	currentSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return db.WorkflowStep{}, fmt.Errorf("list workflow steps: %w", err)
	}
	stepName = uniqueStepName(stepName, currentSteps, 0)

	params := db.CreateWorkflowStepParams{
		WorkflowVersionID: versionID,
		Name:              stepName,
		StepType:          req.StepType,
		InputMapping:      req.InputMapping,
		ControlSettings:   controlSettings,
		CanvasPosition:    req.CanvasPosition,
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
	if shouldAutoResolveTaskRevision(req) {
		revID, err := s.resolveTaskRevision(ctx, req)
		if err != nil {
			return db.WorkflowStep{}, err
		}
		params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revID, Valid: true}
	}

	step, err := s.store.CreateWorkflowStep(ctx, params)
	if err != nil {
		return db.WorkflowStep{}, err
	}
	if err := s.invalidateVersion(ctx, versionID); err != nil {
		return db.WorkflowStep{}, err
	}
	return step, nil
}

// resolveLatestRevision находит последнюю ревизию настроек для work_type.
func shouldAutoResolveTaskRevision(req CreateStepRequest) bool {
	return req.StepType == "task" &&
		req.WorkerSettingsRevisionID == nil &&
		(req.WorkerSettingsSchemaID != nil || req.WorkTypeID != nil)
}

func (s *Service) resolveTaskRevision(ctx context.Context, req CreateStepRequest) (int32, error) {
	if req.WorkerSettingsSchemaID != nil {
		revID, err := s.createPrivateRevisionForSchema(ctx, *req.WorkerSettingsSchemaID)
		if err != nil {
			return 0, fmt.Errorf("create private revision: %w", err)
		}
		return revID, nil
	}

	revID, err := s.resolvePrivateRevision(ctx, *req.WorkTypeID)
	if err != nil {
		return 0, fmt.Errorf("auto-resolve revision: %w", err)
	}
	return revID, nil
}

func (s *Service) resolvePrivateRevision(ctx context.Context, workTypeID int32) (int32, error) {
	schemas, err := s.store.ListWorkerSettingsSchemasByWorkTypeID(ctx, workTypeID)
	if err != nil {
		return 0, fmt.Errorf("list schemas: %w", err)
	}
	if len(schemas) == 0 {
		return 0, fmt.Errorf("no settings schema found for work_type_id=%d", workTypeID)
	}
	return s.createPrivateRevisionForSchema(ctx, schemas[0].ID)
}

func (s *Service) createPrivateRevisionForSchema(ctx context.Context, schemaID int32) (int32, error) {
	revisions, err := s.store.ListWorkerSettingsRevisionsBySchemaID(ctx, schemaID)
	if err != nil {
		return 0, fmt.Errorf("list revisions: %w", err)
	}
	settingsData := []byte(`{}`)
	if len(revisions) > 0 {
		settingsData = revisions[0].SettingsData
	}
	revision, err := s.store.CreateWorkerSettingsRevision(ctx, db.CreateWorkerSettingsRevisionParams{
		WorkerSettingsSchemaID: schemaID,
		CreatedByUserID:        pgtype.Int4{},
		SettingsData:           settingsData,
	})
	if err != nil {
		return 0, fmt.Errorf("create worker settings revision: %w", err)
	}
	return revision.ID, nil
}

// ListSteps возвращает все шаги версии workflow.
func (s *Service) ListSteps(ctx context.Context, versionID int32) ([]db.WorkflowStep, error) {
	return s.store.ListWorkflowStepsByVersionID(ctx, versionID)
}

// UpdateStep обновляет шаг (включая input_mapping — WF-05).
func (s *Service) UpdateStep(ctx context.Context, stepID int32, req UpdateStepRequest) (db.WorkflowStep, error) {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return db.WorkflowStep{}, fmt.Errorf("get step: %w", err)
	}
	if isStartWorkflowStep(current) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return db.WorkflowStep{}, err
	}

	stepType := current.StepType
	if req.StepType != nil {
		stepType = *req.StepType
	}

	controlKind := (*string)(nil)
	if req.ControlKind != nil {
		controlKind = req.ControlKind
	} else if current.ControlKind.Valid {
		controlKind = &current.ControlKind.String
	}

	if isStartStepRequest(stepType, controlKind) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if err := validateControlKind(stepType, controlKind); err != nil {
		return db.WorkflowStep{}, err
	}
	controlSettings := current.ControlSettings
	if req.ControlSettings != nil {
		controlSettings = req.ControlSettings
	}
	if stepType == string(dagpkg.StepTypeControl) && controlKind != nil {
		if err := validateControlSettings(*controlKind, controlSettings); err != nil {
			return db.WorkflowStep{}, err
		}
	}

	stepName := normalizeStepName(current.Name)
	if stepName == "" {
		stepName = defaultStepNameForStep(current)
	}
	if req.Name != nil {
		stepName = normalizeStepName(*req.Name)
		if err := validateStepName(stepName); err != nil {
			return db.WorkflowStep{}, err
		}
		currentSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, current.WorkflowVersionID)
		if err != nil {
			return db.WorkflowStep{}, fmt.Errorf("list workflow steps: %w", err)
		}
		if hasDuplicateStepName(stepName, currentSteps, stepID) {
			return db.WorkflowStep{}, ErrStepNameDuplicate
		}
	}

	params := db.UpdateWorkflowStepParams{
		ID:                       stepID,
		Name:                     stepName,
		StepType:                 stepType,
		WorkTypeID:               current.WorkTypeID,
		WorkerSettingsRevisionID: current.WorkerSettingsRevisionID,
		ControlKind:              current.ControlKind,
		InputMapping:             current.InputMapping,
		CanvasPosition:           current.CanvasPosition,
		ControlSettings:          current.ControlSettings,
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
	if req.ControlSettings != nil {
		params.ControlSettings = req.ControlSettings
	}
	if req.InputMapping != nil {
		params.InputMapping = req.InputMapping
	}
	if req.CanvasPosition != nil {
		params.CanvasPosition = req.CanvasPosition
	}
	updated, err := s.store.UpdateWorkflowStep(ctx, params)
	if err != nil {
		return db.WorkflowStep{}, err
	}
	if err := s.invalidateVersionForStep(ctx, stepID); err != nil {
		return db.WorkflowStep{}, err
	}
	return updated, nil
}

// DeleteStep мягко удаляет шаг и каскадно удаляет его зависимости.
func (s *Service) UpdateTaskSettings(ctx context.Context, stepID int32, req UpdateTaskSettingsRequest) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if current.StepType != string(dagpkg.StepTypeTask) || !current.WorkerSettingsRevisionID.Valid {
		return ErrTaskStepRequired
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return err
	}

	settingsSchema, err := s.settingsSchemaForStep(ctx, current)
	if err != nil {
		return err
	}
	if err := schemadialect.ValidateRaw(settingsSchema.SettingsSchema, req.SettingsData); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}

	revisionID := current.WorkerSettingsRevisionID.Int32
	steps, err := s.store.ListWorkflowStepsByVersionID(ctx, current.WorkflowVersionID)
	if err != nil {
		return fmt.Errorf("list workflow steps: %w", err)
	}
	for _, step := range steps {
		if step.ID == current.ID {
			continue
		}
		if step.StepType != string(dagpkg.StepTypeTask) || !step.WorkerSettingsRevisionID.Valid {
			continue
		}
		if step.WorkerSettingsRevisionID.Int32 == revisionID {
			revision, err := s.store.CloneWorkerSettingsRevision(ctx, db.CloneWorkerSettingsRevisionParams{
				ID:              revisionID,
				CreatedByUserID: pgtype.Int4{},
			})
			if err != nil {
				return fmt.Errorf("clone task settings revision: %w", err)
			}

			if _, err := s.store.UpdateWorkflowStep(ctx, db.UpdateWorkflowStepParams{
				ID:                       current.ID,
				Name:                     current.Name,
				StepType:                 current.StepType,
				WorkTypeID:               current.WorkTypeID,
				WorkerSettingsRevisionID: pgtype.Int4{Int32: revision.ID, Valid: true},
				ControlKind:              current.ControlKind,
				ControlSettings:          current.ControlSettings,
				InputMapping:             current.InputMapping,
				CanvasPosition:           current.CanvasPosition,
			}); err != nil {
				return fmt.Errorf("rebind task settings revision: %w", err)
			}

			revisionID = revision.ID
			break
		}
	}

	if _, err := s.store.UpdateWorkerSettingsRevisionSettings(ctx, db.UpdateWorkerSettingsRevisionSettingsParams{
		ID:           revisionID,
		SettingsData: req.SettingsData,
	}); err != nil {
		return fmt.Errorf("update task settings revision: %w", err)
	}

	return s.invalidateVersion(ctx, current.WorkflowVersionID)
}

func (s *Service) UpdateTaskInputMapping(ctx context.Context, stepID int32, req UpdateTaskInputMappingRequest) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if current.StepType != string(dagpkg.StepTypeTask) {
		return ErrTaskStepRequired
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return err
	}
	if err := validateInputMappingShape(req.InputMapping); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInputMapping, err)
	}
	if err := s.validateTaskInputMapping(ctx, current, req.InputMapping); err != nil {
		return err
	}

	_, err = s.store.UpdateWorkflowStep(ctx, db.UpdateWorkflowStepParams{
		ID:                       current.ID,
		Name:                     current.Name,
		StepType:                 current.StepType,
		WorkTypeID:               current.WorkTypeID,
		WorkerSettingsRevisionID: current.WorkerSettingsRevisionID,
		ControlKind:              current.ControlKind,
		ControlSettings:          current.ControlSettings,
		InputMapping:             req.InputMapping,
		CanvasPosition:           current.CanvasPosition,
	})
	if err != nil {
		return fmt.Errorf("update task input mapping: %w", err)
	}
	return s.invalidateVersion(ctx, current.WorkflowVersionID)
}

func (s *Service) settingsSchemaForStep(ctx context.Context, current db.WorkflowStep) (db.WorkerSettingsSchema, error) {
	enrichedSteps, err := s.store.ListEnrichedStepsByVersionID(ctx, current.WorkflowVersionID)
	if err != nil {
		return db.WorkerSettingsSchema{}, fmt.Errorf("list enriched steps: %w", err)
	}

	for _, step := range enrichedSteps {
		if step.ID != current.ID {
			continue
		}
		if len(step.SettingsSchema) > 0 {
			return db.WorkerSettingsSchema{
				ID:             step.WorkerSettingsSchemaID.Int32,
				SettingsSchema: step.SettingsSchema,
			}, nil
		}
		if step.WorkerSettingsSchemaID.Valid {
			schema, err := s.store.GetWorkerSettingsSchemaByID(ctx, step.WorkerSettingsSchemaID.Int32)
			if err != nil {
				return db.WorkerSettingsSchema{}, fmt.Errorf("get worker settings schema: %w", err)
			}
			return schema, nil
		}
	}

	return db.WorkerSettingsSchema{}, fmt.Errorf("settings schema not found for step %d", current.ID)
}

func validateInputMappingShape(raw json.RawMessage) error {
	mapping, err := parseInputMapping(raw)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(mapping))
	for _, entry := range mapping {
		target := strings.TrimSpace(entry.Target)
		source := strings.TrimSpace(entry.Source)
		if target == "" {
			return fmt.Errorf("mapping target is required")
		}
		if strings.Contains(target, ".") {
			return fmt.Errorf("nested mapping target is not supported: %s", target)
		}
		if _, ok := seen[target]; ok {
			return fmt.Errorf("duplicate mapping target %q", target)
		}
		seen[target] = struct{}{}
		if strings.HasPrefix(source, "$.message.value.") {
			field := strings.TrimPrefix(source, "$.message.value.")
			if field == "" || strings.Contains(field, ".") {
				return fmt.Errorf("nested message source is not supported: %s", source)
			}
			continue
		}
		if strings.HasPrefix(source, "$.steps.") {
			if _, _, err := parseStepOutputSource(source); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) validateTaskInputMapping(ctx context.Context, current db.WorkflowStep, nextMapping json.RawMessage) error {
	inputSchemas, compatibilityErrors, err := s.versionValidationInputSchemas(ctx, current.WorkflowVersionID)
	if err != nil {
		return err
	}
	if len(compatibilityErrors) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidInputMapping, compatibilityErrors[0].Message)
	}
	steps, err := s.store.ListEnrichedStepsByVersionID(ctx, current.WorkflowVersionID)
	if err != nil {
		return fmt.Errorf("list enriched steps: %w", err)
	}
	for i := range steps {
		if steps[i].ID == current.ID {
			steps[i].InputMapping = nextMapping
			break
		}
	}
	deps, err := s.store.ListDependenciesByVersionID(ctx, current.WorkflowVersionID)
	if err != nil {
		return fmt.Errorf("list dependencies: %w", err)
	}
	for _, inputSchema := range inputSchemas {
		for _, validationErr := range validateStepInputsFilled(steps, deps, inputSchema) {
			if validationErr.StepID == current.ID {
				return fmt.Errorf("%w: %s", ErrInvalidInputMapping, validationErr.Message)
			}
		}
	}
	return nil
}

func (s *Service) DeleteStep(ctx context.Context, stepID int32) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if isStartWorkflowStep(current) {
		return ErrStartStepProtected
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return err
	}

	if err := s.removeInboundInputMappings(ctx, current.WorkflowVersionID, stepID); err != nil {
		return fmt.Errorf("delete step input mappings: %w", err)
	}

	// Удаляем зависимости шага перед его удалением
	if err := s.store.DeleteDependenciesByStepID(ctx, stepID); err != nil {
		return fmt.Errorf("delete step dependencies: %w", err)
	}
	if err := s.store.SoftDeleteWorkflowStep(ctx, stepID); err != nil {
		return fmt.Errorf("delete step: %w", err)
	}
	return s.invalidateVersion(ctx, current.WorkflowVersionID)
}

func (s *Service) removeInboundInputMappings(ctx context.Context, versionID, deletedStepID int32) error {
	steps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("list steps: %w", err)
	}

	prefix := fmt.Sprintf("$.steps.%d", deletedStepID)
	expectedPrefix := prefix + "."

	for _, step := range steps {
		if step.ID == deletedStepID {
			continue
		}
		if len(step.InputMapping) == 0 {
			continue
		}

		mapping, parseErr := parseInputMapping(step.InputMapping)
		if parseErr != nil {
			return fmt.Errorf("parse input mapping for step %d: %w", step.ID, parseErr)
		}

		kept := make([]dagpkg.MappingEntry, 0, len(mapping))
		changed := false
		for _, entry := range mapping {
			if strings.TrimSpace(entry.Source) == prefix || strings.HasPrefix(strings.TrimSpace(entry.Source), expectedPrefix) {
				changed = true
				continue
			}
			kept = append(kept, entry)
		}
		if !changed {
			continue
		}

		serialized, err := json.Marshal(kept)
		if err != nil {
			return fmt.Errorf("marshal filtered input mapping for step %d: %w", step.ID, err)
		}
		if _, err := s.store.UpdateWorkflowStep(ctx, db.UpdateWorkflowStepParams{
			ID:                       step.ID,
			Name:                     step.Name,
			StepType:                 step.StepType,
			WorkTypeID:               step.WorkTypeID,
			WorkerSettingsRevisionID: step.WorkerSettingsRevisionID,
			ControlKind:              step.ControlKind,
			ControlSettings:          step.ControlSettings,
			InputMapping:             serialized,
			CanvasPosition:           step.CanvasPosition,
		}); err != nil {
			return fmt.Errorf("clear inbound input mappings for step %d: %w", step.ID, err)
		}
	}
	return nil
}

func parseInputMapping(raw json.RawMessage) ([]dagpkg.MappingEntry, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}

	var arr []dagpkg.MappingEntry
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}

	var mapping map[string]string
	if err := json.Unmarshal(raw, &mapping); err == nil {
		entries := make([]dagpkg.MappingEntry, 0, len(mapping))
		for target, source := range mapping {
			entries = append(entries, dagpkg.MappingEntry{
				Target: target,
				Source: source,
			})
		}
		return entries, nil
	}

	return nil, fmt.Errorf("invalid input_mapping format")
}

// ListDependencies возвращает все зависимости версии workflow.
func (s *Service) ListDependencies(ctx context.Context, versionID int32) ([]db.WorkflowStepDependency, error) {
	return s.store.ListDependenciesByVersionID(ctx, versionID)
}

// --- Dependency management ---

// CreateDependency создаёт зависимость между шагами (WF-04).
func (s *Service) CreateDependency(ctx context.Context, stepID int32, req CreateDependencyRequest) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	dependency, err := s.store.GetWorkflowStepByID(ctx, req.DependsOnStepID)
	if err != nil {
		return fmt.Errorf("get dependency step: %w", err)
	}
	if dependency.WorkflowVersionID != current.WorkflowVersionID {
		return fmt.Errorf("dependency step belongs to another workflow version")
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return err
	}
	if err := s.store.CreateWorkflowStepDependency(ctx, db.CreateWorkflowStepDependencyParams{
		StepID:          stepID,
		DependsOnStepID: req.DependsOnStepID,
		Outcome:         pgtype.Text{String: req.Outcome, Valid: true},
		OutputIndex:     req.OutputIndex,
	}); err != nil {
		return err
	}
	return s.invalidateVersionForStep(ctx, stepID)
}

// UpdateStepPosition обновляет только позицию шага на холсте.
func (s *Service) UpdateStepPosition(ctx context.Context, stepID int32, req UpdateStepPositionRequest) error {
	return s.store.UpdateWorkflowStepPosition(ctx, db.UpdateWorkflowStepPositionParams{
		ID:             stepID,
		CanvasPosition: req.CanvasPosition,
	})
}

// ListEnrichedSteps возвращает шаги с подгруженными work_type meta и settings schemas.
// Маппит []byte JSONB-поля в json.RawMessage для корректной JSON-сериализации.
func (s *Service) ListEnrichedSteps(ctx context.Context, versionID int32) ([]EnrichedStepResponse, error) {
	rows, err := s.store.ListEnrichedStepsByVersionID(ctx, versionID)
	if err != nil {
		return nil, err
	}

	result := make([]EnrichedStepResponse, 0, len(rows))
	for _, r := range rows {
		step := EnrichedStepResponse{
			ID:                r.ID,
			WorkflowVersionID: r.WorkflowVersionID,
			Name:              r.Name,
			StepType:          r.StepType,
			ControlSettings:   toRawMessage(r.ControlSettings),
			InputMapping:      toRawMessage(r.InputMapping),
			CanvasPosition:    toRawMessage(r.CanvasPosition),
			WorkTypeMeta:      toRawMessage(r.WorkTypeMeta),
			Config:            toRawMessage(r.Config),
			SettingsSchema:    toRawMessage(r.SettingsSchema),
			InputSchema:       toRawMessage(r.InputSchema),
			OutputSchema:      toRawMessage(r.OutputSchema),
		}
		if r.WorkTypeID.Valid {
			id := r.WorkTypeID.Int32
			step.WorkTypeID = &id
		}
		if r.WorkerSettingsRevisionID.Valid {
			id := r.WorkerSettingsRevisionID.Int32
			step.WorkerSettingsRevisionID = &id
		}
		if r.ControlKind.Valid {
			step.ControlKind = &r.ControlKind.String
		}
		if r.WorkTypeName.Valid {
			step.WorkTypeName = &r.WorkTypeName.String
		}
		if r.WorkTypeCode.Valid {
			step.WorkTypeCode = &r.WorkTypeCode.String
		}
		result = append(result, step)
	}
	return result, nil
}

// toRawMessage конвертирует []byte в json.RawMessage, возвращает nil для пустых значений.
func toRawMessage(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return json.RawMessage(b)
}

func normalizeStepName(name string) string {
	return strings.TrimSpace(name)
}

func validateStepName(name string) error {
	if name == "" {
		return ErrStepNameRequired
	}
	if utf8.RuneCountInString(name) > maxStepNameLength {
		return ErrStepNameInvalid
	}
	return nil
}

func defaultStepNameForCreate(req CreateStepRequest) string {
	if req.StepType == string(dagpkg.StepTypeControl) && req.ControlKind != nil {
		return defaultStepNameForControlKind(*req.ControlKind)
	}
	return "Task"
}

func defaultStepNameForStep(step db.WorkflowStep) string {
	if step.StepType == string(dagpkg.StepTypeControl) && step.ControlKind.Valid {
		return defaultStepNameForControlKind(step.ControlKind.String)
	}
	return fmt.Sprintf("Step %d", step.ID)
}

func defaultStepNameForControlKind(kind string) string {
	if kind == dagpkg.ControlKindStart {
		return "System Trigger"
	}
	return kind
}

func uniqueStepName(base string, steps []db.WorkflowStep, excludeID int32) string {
	base = truncateStepName(base)
	used := activeStepNames(steps, excludeID)
	if _, exists := used[base]; !exists {
		return base
	}

	maxSuffix := int64(1)
	prefix := base + " "
	for name := range used {
		if name == base {
			continue
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimPrefix(name, prefix), 10, 32)
		if err != nil || n < 2 {
			continue
		}
		if n > maxSuffix {
			maxSuffix = n
		}
	}

	for suffix := maxSuffix + 1; ; suffix++ {
		suffixText := fmt.Sprintf(" %d", suffix)
		candidate := truncateStepNameForSuffix(base, suffixText) + suffixText
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func hasDuplicateStepName(name string, steps []db.WorkflowStep, excludeID int32) bool {
	_, exists := activeStepNames(steps, excludeID)[name]
	return exists
}

func activeStepNames(steps []db.WorkflowStep, excludeID int32) map[string]struct{} {
	names := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		if step.ID == excludeID || step.DeletedAt.Valid {
			continue
		}
		name := normalizeStepName(step.Name)
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	return names
}

func truncateStepName(name string) string {
	return truncateRunes(name, maxStepNameLength)
}

func truncateStepNameForSuffix(name, suffix string) string {
	return truncateRunes(name, maxStepNameLength-utf8.RuneCountInString(suffix))
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}

func (s *Service) invalidateVersion(ctx context.Context, versionID int32) error {
	_, err := s.store.UpdateWorkflowVersionValid(ctx, db.UpdateWorkflowVersionValidParams{
		ID:      versionID,
		IsValid: false,
	})
	return err
}

func (s *Service) invalidateVersionForStep(ctx context.Context, stepID int32) error {
	step, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step for invalidation: %w", err)
	}
	return s.invalidateVersion(ctx, step.WorkflowVersionID)
}

func timestampString(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.Format(time.RFC3339)
}

func pgTextString(text pgtype.Text) string {
	if !text.Valid {
		return ""
	}
	return text.String
}

func normalizeVersionName(name string) string {
	return strings.TrimSpace(name)
}

func hasDuplicateVersionName(name string, versions []db.WorkflowVersion, excludeID int32) bool {
	_, exists := activeVersionNames(versions, excludeID)[normalizeVersionName(name)]
	return exists
}

func uniqueVersionName(base string, versions []db.WorkflowVersion, excludeID int32) string {
	base = truncateVersionName(normalizeVersionName(base))
	used := activeVersionNames(versions, excludeID)
	if _, exists := used[base]; !exists {
		return base
	}

	maxSuffix := int64(1)
	prefix := base + " "
	for name := range used {
		if name == base || !strings.HasPrefix(name, prefix) {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimPrefix(name, prefix), 10, 32)
		if err != nil || n < 2 {
			continue
		}
		if n > maxSuffix {
			maxSuffix = n
		}
	}

	for suffix := maxSuffix + 1; ; suffix++ {
		suffixText := fmt.Sprintf(" %d", suffix)
		candidate := truncateVersionNameForSuffix(base, suffixText) + suffixText
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func activeVersionNames(versions []db.WorkflowVersion, excludeID int32) map[string]struct{} {
	names := make(map[string]struct{}, len(versions))
	for _, version := range versions {
		if version.ID == excludeID || version.DeletedAt.Valid || !version.Name.Valid {
			continue
		}
		name := normalizeVersionName(version.Name.String)
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	return names
}

func truncateVersionName(name string) string {
	return truncateRunes(name, maxVersionNameLength)
}

func truncateVersionNameForSuffix(name, suffix string) string {
	return truncateRunes(name, maxVersionNameLength-utf8.RuneCountInString(suffix))
}

func isUniqueConstraintViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}

func isStartWorkflowStep(step db.WorkflowStep) bool {
	return step.StepType == string(dagpkg.StepTypeControl) &&
		step.ControlKind.Valid &&
		step.ControlKind.String == dagpkg.ControlKindStart
}

func isStartStepRequest(stepType string, controlKind *string) bool {
	return stepType == string(dagpkg.StepTypeControl) &&
		controlKind != nil &&
		*controlKind == dagpkg.ControlKindStart
}

func validateControlKind(stepType string, controlKind *string) error {
	if stepType != string(dagpkg.StepTypeControl) {
		return nil
	}
	if controlKind == nil {
		return ErrControlKindInvalid
	}
	if _, ok := allowedControlKinds[*controlKind]; !ok {
		return ErrControlKindInvalid
	}
	return nil
}

func validateControlSettings(controlKind string, raw json.RawMessage) error {
	if controlKind != dagpkg.ControlKindDelay {
		return nil
	}
	return validateDelayControlSettings(raw)
}

func defaultControlSettingsIfEmpty(controlKind string, raw json.RawMessage) json.RawMessage {
	if !isEmptyControlSettings(raw) {
		return raw
	}
	switch controlKind {
	case dagpkg.ControlKindCondition:
		return json.RawMessage(`{"left":"$.message.value","operator":"exists","right":""}`)
	case dagpkg.ControlKindSwitch:
		return json.RawMessage(`{"expression":"$.message.value","cases":[]}`)
	case dagpkg.ControlKindDelay:
		return json.RawMessage(`{"count":1,"unit":"sec"}`)
	default:
		return raw
	}
}

func isEmptyControlSettings(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) == 0 || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte("{}"))
}

func validateDelayControlSettings(raw json.RawMessage) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ErrControlSettingsInvalid
	}

	type delaySettings struct {
		Count float64 `json:"count"`
		Unit  string  `json:"unit"`
	}
	var settings delaySettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("%w: %w", ErrControlSettingsInvalid, err)
	}

	if settings.Count <= 0 {
		return ErrControlSettingsInvalid
	}
	if settings.Unit == "" {
		return ErrControlSettingsInvalid
	}
	if _, ok := delayDurationMultiplier(strings.ToLower(strings.TrimSpace(settings.Unit))); ok {
		return nil
	}
	return ErrControlSettingsInvalid
}

func validateVersionControlSteps(steps []db.WorkflowStep) []dagpkg.ValidationError {
	var errors []dagpkg.ValidationError
	for _, step := range steps {
		if step.StepType != string(dagpkg.StepTypeControl) {
			continue
		}
		if !step.ControlKind.Valid || strings.TrimSpace(step.ControlKind.String) == "" {
			errors = append(errors, controlValidationError("invalid_control_kind", step.ID, "control kind is required"))
			continue
		}

		controlKind := strings.TrimSpace(step.ControlKind.String)
		if _, ok := allowedControlKinds[controlKind]; !ok {
			errors = append(errors, controlValidationError("invalid_control_kind", step.ID, fmt.Sprintf("unsupported control kind %q", controlKind)))
			continue
		}
		if err := validateControlSettingsForVersion(controlKind, step.ControlSettings); err != nil {
			errors = append(errors, controlValidationError("invalid_control_settings", step.ID, err.Error()))
		}
	}
	return errors
}

func validateControlSettingsForVersion(controlKind string, raw json.RawMessage) error {
	switch controlKind {
	case dagpkg.ControlKindStart:
		return nil
	case dagpkg.ControlKindCondition:
		return validateConditionControlSettings(raw)
	case dagpkg.ControlKindSwitch:
		return validateSwitchControlSettings(raw)
	case dagpkg.ControlKindDelay:
		return validateDelayControlSettings(raw)
	default:
		return ErrControlKindInvalid
	}
}

func validateConditionControlSettings(raw json.RawMessage) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ErrControlSettingsInvalid
	}

	type conditionSettings struct {
		Left     string `json:"left"`
		Operator string `json:"operator"`
	}
	var settings conditionSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("%w: %w", ErrControlSettingsInvalid, err)
	}
	if strings.TrimSpace(settings.Left) == "" {
		return ErrControlSettingsInvalid
	}
	if _, ok := allowedConditionOperators[strings.TrimSpace(settings.Operator)]; !ok {
		return ErrControlSettingsInvalid
	}
	return nil
}

func validateSwitchControlSettings(raw json.RawMessage) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ErrControlSettingsInvalid
	}

	type switchSettings struct {
		Expression string            `json:"expression"`
		Cases      []json.RawMessage `json:"cases"`
	}
	var settings switchSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("%w: %w", ErrControlSettingsInvalid, err)
	}
	if strings.TrimSpace(settings.Expression) == "" {
		return ErrControlSettingsInvalid
	}
	seenIDs := make(map[string]struct{}, len(settings.Cases))
	for _, item := range settings.Cases {
		var legacy string
		if err := json.Unmarshal(item, &legacy); err == nil {
			if strings.TrimSpace(legacy) == "" {
				return ErrControlSettingsInvalid
			}
			continue
		}

		var objectCase struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(item, &objectCase); err != nil {
			return fmt.Errorf("%w: %w", ErrControlSettingsInvalid, err)
		}
		id := strings.TrimSpace(objectCase.ID)
		if id == "" || id == "default" {
			return ErrControlSettingsInvalid
		}
		if strings.TrimSpace(objectCase.Label) == "" || strings.TrimSpace(objectCase.Value) == "" {
			return ErrControlSettingsInvalid
		}
		if _, exists := seenIDs[id]; exists {
			return ErrControlSettingsInvalid
		}
		seenIDs[id] = struct{}{}
	}
	return nil
}

func controlValidationError(errorType string, stepID int32, message string) dagpkg.ValidationError {
	return dagpkg.ValidationError{
		Type:    errorType,
		StepID:  stepID,
		Message: fmt.Sprintf("control step %d: %s", stepID, message),
	}
}

func delayDurationMultiplier(unit string) (time.Duration, bool) {
	switch unit {
	case "sec", "s", "seconds", "second":
		return time.Second, true
	case "min", "m", "minutes", "minute":
		return time.Minute, true
	case "hour", "h", "hours":
		return time.Hour, true
	case "day", "d", "days":
		return 24 * time.Hour, true
	}
	return 0, false
}

// DeleteDependency удаляет зависимость между шагами.
func (s *Service) DeleteDependency(ctx context.Context, stepID, dependsOnStepID int32) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if err := s.ensureVersionEditable(ctx, current.WorkflowVersionID); err != nil {
		return err
	}
	if err := s.store.DeleteWorkflowStepDependency(ctx, db.DeleteWorkflowStepDependencyParams{
		StepID:          stepID,
		DependsOnStepID: dependsOnStepID,
	}); err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}
	return s.invalidateVersionForStep(ctx, stepID)
}

// --- Validation ---

// ValidateVersion запускает валидацию DAG для версии (WF-06, WF-07).
// Загружает шаги и зависимости, вызывает dag.ValidateDAG,
// обновляет is_valid у версии.
// Возвращает результат валидации со всеми ошибками сразу (D-12).
func (s *Service) ValidateVersion(ctx context.Context, versionID int32) (ValidateVersionResponse, error) {
	if _, err := s.store.GetWorkflowVersionByID(ctx, versionID); err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("get workflow version: %w", err)
	}
	inputSchemas, compatibilityErrors, err := s.versionValidationInputSchemas(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, err
	}
	dbSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list steps: %w", err)
	}

	enrichedSteps, err := s.store.ListEnrichedStepsByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list enriched steps: %w", err)
	}

	dbDeps, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list dependencies: %w", err)
	}

	primarySchema := json.RawMessage(`{"type":"object","properties":{}}`)
	if len(inputSchemas) > 0 {
		primarySchema = inputSchemas[0]
	}
	inputNames := workflowInputNamesFromSchema(primarySchema)

	// Конвертируем в типы dag пакета
	dagSteps := make([]dagpkg.Step, 0, len(dbSteps))
	for _, step := range dbSteps {
		dagStep := dagpkg.Step{
			ID:              step.ID,
			StepType:        dagpkg.StepType(step.StepType),
			ControlSettings: step.ControlSettings,
			WorkflowInputs:  inputNames,
		}
		if step.ControlKind.Valid {
			dagStep.ControlKind = step.ControlKind.String
		}
		// Парсим input_mapping из JSONB
		if len(step.InputMapping) > 0 {
			var mappings []dagpkg.MappingEntry
			if err := json.Unmarshal(step.InputMapping, &mappings); err == nil {
				for _, mapping := range mappings {
					if strings.HasPrefix(strings.TrimSpace(mapping.Source), "$.") {
						dagStep.InputMapping = append(dagStep.InputMapping, mapping)
					}
				}
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
	validationErrors := compatibilityErrors
	validationErrors = append(validationErrors, dagpkg.ValidateDAG(dagSteps, dagDeps)...)
	validationErrors = append(validationErrors, validateVersionControlSteps(dbSteps)...)
	for _, inputSchema := range inputSchemas {
		validationErrors = append(validationErrors, validateStepInputsFilled(enrichedSteps, dbDeps, inputSchema)...)
	}
	validationErrors = append(validationErrors, validateStepSettings(enrichedSteps)...)
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

func (s *Service) versionValidationInputSchemas(ctx context.Context, versionID int32) ([]json.RawMessage, []dagpkg.ValidationError, error) {
	compatQueries, err := s.compatibilityQueries()
	if err != nil {
		return nil, nil, err
	}
	inputSchemaQueries, err := s.inputSchemaQueries()
	if err != nil {
		return nil, nil, err
	}
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return nil, nil, fmt.Errorf("get workflow version: %w", err)
	}
	compatibilities, err := compatQueries.ListCompatibilitiesByVersionID(ctx, versionID)
	if err != nil {
		return nil, nil, fmt.Errorf("list version compatibilities: %w", err)
	}

	var validationErrors []dagpkg.ValidationError
	inputSchemas := make([]json.RawMessage, 0, len(compatibilities))
	for _, compatibility := range compatibilities {
		if !compatibility.IsActive {
			continue
		}
		validationErrors = append(validationErrors, validateVersionCompatibilityConfig(compatibility)...)

		schema, err := inputSchemaQueries.GetWorkflowInputSchemaByID(ctx, compatibility.WorkflowInputSchemaID)
		if err != nil {
			return nil, nil, fmt.Errorf("get compatibility input schema %d: %w", compatibility.WorkflowInputSchemaID, err)
		}
		if schema.Status != "active" && !isDraftNativeInputSchemaForVersion(schema, compatibility, version) {
			validationErrors = append(validationErrors, dagpkg.ValidationError{
				Type:    "inactive_input_schema",
				Message: fmt.Sprintf("input schema %s v%d is not active", schema.Code, schema.VersionNumber),
			})
		}
		inputSchemas = append(inputSchemas, json.RawMessage(schema.SchemaJson))
	}
	if len(inputSchemas) == 0 {
		validationErrors = append(validationErrors, dagpkg.ValidationError{
			Type:    "missing_compatibility",
			Message: "workflow version must have at least one active input schema compatibility",
		})
	}
	return inputSchemas, validationErrors, nil
}

func validateVersionCompatibilityConfig(compatibility db.WorkflowVersionInputSchemaCompatibility) []dagpkg.ValidationError {
	hasMapper := compatibility.WorkflowInputMapperID.Valid
	hasDefaults := hasCompatibilityDefaultValues(compatibility.DefaultValues)
	switch compatibility.CompatibilityType {
	case "native":
		if hasMapper || hasDefaults {
			return []dagpkg.ValidationError{{
				Type:    "invalid_compatibility",
				Message: fmt.Sprintf("native compatibility %d must not define mapper or default values", compatibility.ID),
			}}
		}
	case "adapter", "partial":
		if !hasMapper && !hasDefaults {
			return []dagpkg.ValidationError{{
				Type:    "invalid_compatibility",
				Message: fmt.Sprintf("%s compatibility %d requires mapper or default values", compatibility.CompatibilityType, compatibility.ID),
			}}
		}
	default:
		return []dagpkg.ValidationError{{
			Type:    "invalid_compatibility",
			Message: fmt.Sprintf("compatibility %d has unsupported type %q", compatibility.ID, compatibility.CompatibilityType),
		}}
	}
	return nil
}

func isDraftNativeInputSchemaForVersion(schema db.WorkflowInputSchema, compatibility db.WorkflowVersionInputSchemaCompatibility, version db.WorkflowVersion) bool {
	return schema.Status == "draft" &&
		compatibility.CompatibilityType == "native" &&
		compatibility.WorkflowVersionID == version.ID &&
		compatibility.WorkflowInputSchemaID == schema.ID &&
		schema.WorkflowID == version.WorkflowID &&
		schema.VersionNumber == version.VersionNumber
}

func hasCompatibilityDefaultValues(raw []byte) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return false
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return true
	}
	return len(obj) > 0
}

func validateStepSettings(steps []db.ListEnrichedStepsByVersionIDRow) []dagpkg.ValidationError {
	var errors []dagpkg.ValidationError
	for _, step := range steps {
		if step.StepType != string(dagpkg.StepTypeTask) || len(step.SettingsSchema) == 0 {
			continue
		}

		payload := bytes.TrimSpace(step.Config)
		if len(payload) == 0 {
			payload = []byte(`{}`)
		}
		if err := schemadialect.ValidateRaw(step.SettingsSchema, payload); err != nil {
			errors = append(errors, dagpkg.ValidationError{
				Type:    "invalid_settings",
				StepID:  step.ID,
				Message: fmt.Sprintf("task step %d settings are invalid: %v", step.ID, err),
			})
		}
	}
	return errors
}

func workflowInputNamesFromSchema(schemaJSON []byte) []string {
	props, err := parseTopLevelProperties(schemaJSON)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
	_, compatibilityErrors, err := s.versionValidationInputSchemas(ctx, versionID)
	if err != nil {
		return err
	}
	if len(compatibilityErrors) > 0 {
		return fmt.Errorf("%w: active input schema compatibility is required and must be valid", ErrWorkflowConfigurationInvalid)
	}

	if !version.IsValid {
		return ErrValidationRequired
	}

	nativeSchema, err := s.store.GetNativeInputSchemaForWorkflowVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("%w: native input schema is required", ErrWorkflowConfigurationInvalid)
	}
	if nativeSchema.Status == "draft" {
		if _, err := s.store.UpdateWorkflowInputSchemaStatus(ctx, db.UpdateWorkflowInputSchemaStatusParams{
			ID:              nativeSchema.ID,
			Status:          "active",
			UpdatedByUserID: pgtype.Int4{},
		}); err != nil {
			return fmt.Errorf("activate native input schema: %w", err)
		}
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: true,
	}); err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	return nil
}

// DeactivateVersion деактивирует версию workflow (WF-08).
func (s *Service) DeactivateVersion(ctx context.Context, versionID int32) error {
	if _, err := s.store.GetWorkflowVersionByID(ctx, versionID); err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: false,
	}); err != nil {
		return fmt.Errorf("deactivate version: %w", err)
	}

	return nil
}

// DeleteVersion soft-deletes a workflow version and its steps.
func (s *Service) DeleteVersion(ctx context.Context, versionID int32) error {
	if err := s.ensureVersionDeletable(ctx, versionID); err != nil {
		return err
	}

	if err := s.store.WithTx(ctx, func(q *db.Queries) error {
		if err := q.DeleteWorkflowStepsByVersionID(ctx, versionID); err != nil {
			return fmt.Errorf("delete version steps: %w", err)
		}
		if err := q.SoftDeleteWorkflowVersion(ctx, versionID); err != nil {
			return fmt.Errorf("delete version: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) ensureVersionEditable(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}
	if version.IsActive || version.LockedAt.Valid || version.PublishedAt.Valid {
		return ErrWorkflowVersionLocked
	}
	return nil
}

func (s *Service) ensureVersionDeletable(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}
	if version.IsActive || version.LockedAt.Valid || version.PublishedAt.Valid {
		return ErrWorkflowVersionLocked
	}
	q, ok := s.store.(workflowVersionReferenceQueries)
	if !ok {
		return nil
	}
	referenced, err := q.IsWorkflowVersionReferenced(ctx, versionID)
	if err != nil {
		return fmt.Errorf("check version references: %w", err)
	}
	if referenced {
		return ErrWorkflowVersionInUse
	}
	return nil
}
