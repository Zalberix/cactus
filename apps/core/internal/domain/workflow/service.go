package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
	schemadialect "github.com/zalberix/cactus/apps/core/internal/schema"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// ErrValidationRequired вЂ” РІРµСЂСЃРёСЏ РЅРµ РїСЂРѕС€Р»Р° РІР°Р»РёРґР°С†РёСЋ DAG.
var ErrValidationRequired = errors.New("VALIDATION_REQUIRED: РІРµСЂСЃРёСЏ РґРѕР»Р¶РЅР° РїСЂРѕР№С‚Рё РІР°Р»РёРґР°С†РёСЋ DAG РїРµСЂРµРґ Р°РєС‚РёРІР°С†РёРµР№")

var ErrStartStepProtected = errors.New("START_STEP_PROTECTED: start step is managed by the system")

var ErrTrafficWeightInvalid = errors.New("TRAFFIC_WEIGHT_INVALID: custom traffic weight total must be <= 100")

var ErrTaskStepRequired = errors.New("TASK_STEP_REQUIRED: step must be a task")

var ErrControlSettingsInvalid = errors.New("CONTROL_SETTINGS_INVALID: invalid control settings")

var ErrControlKindInvalid = errors.New("CONTROL_KIND_INVALID: unsupported control kind")

var ErrInvalidInputMapping = errors.New("INVALID_INPUT_MAPPING")

var ErrInvalidSettings = errors.New("INVALID_SETTINGS")

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

// Service вЂ” Р±РёР·РЅРµСЃ-Р»РѕРіРёРєР° СѓРїСЂР°РІР»РµРЅРёСЏ workflow.
type Service struct {
	store Storage
}

// NewService СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ workflow.Service.
func NewService(store Storage) *Service {
	return &Service{store: store}
}

// --- Workflow CRUD ---

// CreateWorkflow СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ workflow РІ СЃРёСЃС‚РµРјРµ (WF-01).
func (s *Service) CreateWorkflow(ctx context.Context, systemID int32, req CreateWorkflowRequest) (db.Workflow, error) {
	if req.Priority > math.MaxInt32 || req.Priority < math.MinInt32 {
		return db.Workflow{}, fmt.Errorf("priority %d overflows int32", req.Priority)
	}
	return s.store.CreateWorkflow(ctx, db.CreateWorkflowParams{
		SystemID:    systemID,
		Name:        req.Name,
		Priority:    int32(req.Priority),
		InputSchema: []byte(`{"type":"object","properties":{}}`),
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// ListWorkflows РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРїРёСЃРѕРє workflow СЃРёСЃС‚РµРјС‹.
func (s *Service) ListWorkflows(ctx context.Context, systemID int32) ([]db.Workflow, error) {
	return s.store.ListWorkflowsBySystemID(ctx, systemID)
}

// GetWorkflow РІРѕР·РІСЂР°С‰Р°РµС‚ workflow РїРѕ ID.
func (s *Service) GetWorkflow(ctx context.Context, id int32) (db.Workflow, error) {
	return s.store.GetWorkflowByID(ctx, id)
}

// UpdateWorkflow РѕР±РЅРѕРІР»СЏРµС‚ РЅР°Р·РІР°РЅРёРµ Рё РїСЂРёРѕСЂРёС‚РµС‚ workflow.
func (s *Service) UpdateWorkflow(ctx context.Context, id int32, req UpdateWorkflowRequest) (db.Workflow, error) {
	return s.store.UpdateWorkflow(ctx, db.UpdateWorkflowParams{
		ID:          id,
		Name:        req.Name,
		Priority:    int32(req.Priority), // #nosec G115 -- priority is validated by request binding.
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// DeleteWorkflow РјСЏРіРєРѕ СѓРґР°Р»СЏРµС‚ workflow.
func (s *Service) DeleteWorkflow(ctx context.Context, id int32) error {
	return s.store.SoftDeleteWorkflow(ctx, id)
}

// --- Version management ---

// CreateVersion СЃРѕР·РґР°С‘С‚ РЅРѕРІСѓСЋ РІРµСЂСЃРёСЋ workflow (WF-02).
// РќРѕРјРµСЂ РІРµСЂСЃРёРё Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРё РёРЅРєСЂРµРјРµРЅС‚РёСЂСѓРµС‚СЃСЏ.
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
			TrafficWeight:   100,
			IsControlGroup:  false,
		})
		if createErr != nil {
			return fmt.Errorf("create workflow version: %w", createErr)
		}

		_, createErr = q.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
			WorkflowVersionID: version.ID,
			StepType:          string(dagpkg.StepTypeControl),
			ControlKind:       pgtype.Text{String: dagpkg.ControlKindStart, Valid: true},
			ControlSettings:   []byte(`{"trigger":"system_message"}`),
			CanvasPosition:    []byte(`{"x":80,"y":200}`),
		})
		if createErr != nil {
			return fmt.Errorf("create start step: %w", createErr)
		}

		return nil
	})
	if err != nil {
		return db.WorkflowVersion{}, err
	}
	return version, nil
}

// ListVersions РІРѕР·РІСЂР°С‰Р°РµС‚ РІСЃРµ РІРµСЂСЃРёРё workflow.
func (s *Service) ListVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error) {
	return s.store.ListWorkflowVersionsByWorkflowID(ctx, workflowID)
}

// ListVersionSummaries РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРІРѕРґРЅС‹Рµ РґР°РЅРЅС‹Рµ РїРѕ РІРµСЂСЃРёСЏРј workflow.
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
			ID:             r.ID,
			WorkflowID:     r.WorkflowID,
			Name:           pgTextString(r.Name),
			VersionNumber:  r.VersionNumber,
			IsValid:        r.IsValid,
			IsActive:       r.IsActive,
			TrafficWeight:  r.TrafficWeight,
			IsControlGroup: r.IsControlGroup,
			RunCount:       r.RunCount,
			CreatedAt:      timestampString(r.CreatedAt),
			UpdatedAt:      timestampString(r.UpdatedAt),
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

// UpdateVersionName РѕР±РЅРѕРІР»СЏРµС‚ РЅР°Р·РІР°РЅРёРµ РІРµСЂСЃРёРё.
func (s *Service) UpdateVersionName(ctx context.Context, versionID int32, req UpdateVersionNameRequest) (db.WorkflowVersion, error) {
	return s.store.UpdateWorkflowVersionName(ctx, db.UpdateWorkflowVersionNameParams{
		ID:   versionID,
		Name: pgtype.Text{String: req.Name, Valid: true},
	})
}

// CopyVersion СЃРѕР·РґР°С‘С‚ РєРѕРїРёСЋ РІРµСЂСЃРёРё.
func (s *Service) CopyVersion(ctx context.Context, versionID, userID int32) (db.WorkflowVersion, error) {
	source, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("get source version: %w", err)
	}

	maxVersion, err := s.store.GetMaxVersionNumberByWorkflowID(ctx, source.WorkflowID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("get max version number: %w", err)
	}

	sourceName := source.Name.String
	if !source.Name.Valid || strings.TrimSpace(sourceName) == "" {
		sourceName = fmt.Sprintf("Version %d", source.VersionNumber)
	}

	newVersion, err := s.store.CreateWorkflowVersion(ctx, db.CreateWorkflowVersionParams{
		WorkflowID:      source.WorkflowID,
		CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
		VersionNumber:   maxVersion + 1,
		Name:            pgtype.Text{String: sourceName + " copy", Valid: true},
		IsValid:         false,
		IsActive:        false,
		TrafficWeight:   0,
		IsControlGroup:  false,
	})
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("create copied version: %w", err)
	}

	sourceSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("list source steps: %w", err)
	}

	stepIDMap := make(map[int32]int32, len(sourceSteps))
	for _, sourceStep := range sourceSteps {
		params := db.CreateWorkflowStepParams{
			WorkflowVersionID: newVersion.ID,
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
			revision, err := s.store.CloneWorkerSettingsRevision(ctx, db.CloneWorkerSettingsRevisionParams{
				ID:              sourceStep.WorkerSettingsRevisionID.Int32,
				CreatedByUserID: pgtype.Int4{Int32: userID, Valid: true},
			})
			if err != nil {
				return db.WorkflowVersion{}, fmt.Errorf("clone settings revision for step %d: %w", sourceStep.ID, err)
			}
			params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revision.ID, Valid: true}
		}

		copiedStep, err := s.store.CreateWorkflowStep(ctx, params)
		if err != nil {
			return db.WorkflowVersion{}, fmt.Errorf("copy step %d: %w", sourceStep.ID, err)
		}
		stepIDMap[sourceStep.ID] = copiedStep.ID
	}

	sourceDependencies, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return db.WorkflowVersion{}, fmt.Errorf("list source dependencies: %w", err)
	}
	for _, dep := range sourceDependencies {
		stepID, ok := stepIDMap[dep.StepID]
		if !ok {
			return db.WorkflowVersion{}, fmt.Errorf("missing copied step for source step_id=%d", dep.StepID)
		}
		depStepID, ok := stepIDMap[dep.DependsOnStepID]
		if !ok {
			return db.WorkflowVersion{}, fmt.Errorf("missing copied step for depends_on_step_id=%d", dep.DependsOnStepID)
		}
		if err := s.store.CreateWorkflowStepDependency(ctx, db.CreateWorkflowStepDependencyParams{
			StepID:          stepID,
			DependsOnStepID: depStepID,
			Outcome:         dep.Outcome,
			OutputIndex:     dep.OutputIndex,
		}); err != nil {
			return db.WorkflowVersion{}, fmt.Errorf("create dependency copy %d->%d: %w", dep.DependsOnStepID, dep.StepID, err)
		}
	}

	return newVersion, nil
}

// UpdateWorkflowTraffic РѕР±РЅРѕРІР»СЏРµС‚ СЂР°СЃРїСЂРµРґРµР»РµРЅРёРµ traffic РґР»СЏ workflow.
func (s *Service) UpdateWorkflowTraffic(ctx context.Context, workflowID int32, req UpdateTrafficRequest) error {
	versions, err := s.store.ListAllWorkflowVersionsByWorkflowID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("list versions: %w", err)
	}

	activeVersions, activeSet := activeWorkflowVersions(versions)
	traffic, err := buildWorkflowTraffic(req, activeVersions, activeSet)
	if err != nil {
		return err
	}

	for _, version := range versions {
		weight := int32(0)
		if version.IsActive {
			weight = traffic[version.ID]
		}
		if _, err := s.store.UpdateWorkflowVersionTrafficWeightIncludingDeleted(ctx, db.UpdateWorkflowVersionTrafficWeightIncludingDeletedParams{
			ID:            version.ID,
			TrafficWeight: weight,
		}); err != nil {
			return fmt.Errorf("update traffic weight for version %d: %w", version.ID, err)
		}
	}

	return nil
}

func activeWorkflowVersions(versions []db.WorkflowVersion) ([]db.WorkflowVersion, map[int32]db.WorkflowVersion) {
	activeVersions := make([]db.WorkflowVersion, 0, len(versions))
	activeSet := make(map[int32]db.WorkflowVersion, len(versions))
	for _, version := range versions {
		if !version.IsActive {
			continue
		}
		activeVersions = append(activeVersions, version)
		activeSet[version.ID] = version
	}
	return activeVersions, activeSet
}

func buildWorkflowTraffic(
	req UpdateTrafficRequest,
	activeVersions []db.WorkflowVersion,
	activeSet map[int32]db.WorkflowVersion,
) (map[int32]int32, error) {
	if len(req.Versions) > 0 {
		return distributePerVersionTraffic(activeVersions, activeSet, req.Versions)
	}
	switch req.Mode {
	case "equal":
		return distributeEqualTraffic(activeVersions), nil
	case "custom":
		return distributeCustomTraffic(activeVersions, activeSet, req.Weights)
	default:
		return nil, fmt.Errorf("unsupported traffic mode: %s", req.Mode)
	}
}

func distributeEqualTraffic(activeVersions []db.WorkflowVersion) map[int32]int32 {
	traffic := make(map[int32]int32, len(activeVersions))
	if len(activeVersions) == 0 {
		return traffic
	}

	ids := make([]int32, 0, len(activeVersions))
	for _, version := range activeVersions {
		ids = append(ids, version.ID)
	}
	assignEvenTraffic(traffic, ids, 100)
	return traffic
}

func distributeCustomTraffic(
	activeVersions []db.WorkflowVersion,
	activeSet map[int32]db.WorkflowVersion,
	weights []TrafficWeightInput,
) (map[int32]int32, error) {
	assigned, err := assignedTrafficWeights(activeSet, weights)
	if err != nil {
		return nil, err
	}

	traffic := make(map[int32]int32, len(activeVersions))
	var sum int32
	unassignedIDs := make([]int32, 0, len(activeVersions))
	for _, version := range activeVersions {
		weight, ok := assigned[version.ID]
		traffic[version.ID] = weight
		sum += weight
		if !ok {
			unassignedIDs = append(unassignedIDs, version.ID)
		}
	}
	if sum > 100 {
		return nil, ErrTrafficWeightInvalid
	}
	if remaining := 100 - sum; len(unassignedIDs) > 0 && remaining > 0 {
		assignEvenTraffic(traffic, unassignedIDs, remaining)
	}
	return traffic, nil
}

func assignedTrafficWeights(activeSet map[int32]db.WorkflowVersion, weights []TrafficWeightInput) (map[int32]int32, error) {
	assigned := make(map[int32]int32, len(weights))
	for _, weight := range weights {
		if _, ok := activeSet[weight.VersionID]; !ok {
			return nil, fmt.Errorf("version %d is not active", weight.VersionID)
		}
		assigned[weight.VersionID] = weight.Weight
	}
	return assigned, nil
}

func distributePerVersionTraffic(
	activeVersions []db.WorkflowVersion,
	activeSet map[int32]db.WorkflowVersion,
	inputs []TrafficVersionInput,
) (map[int32]int32, error) {
	traffic := make(map[int32]int32, len(activeVersions))
	inputByID := make(map[int32]TrafficVersionInput, len(inputs))

	for _, input := range inputs {
		if _, ok := activeSet[input.VersionID]; !ok {
			return nil, fmt.Errorf("version %d is not active", input.VersionID)
		}
		inputByID[input.VersionID] = input
	}

	var fixedTotal int32
	shareIDs := make([]int32, 0, len(activeVersions))
	for _, version := range activeVersions {
		input, ok := inputByID[version.ID]
		if !ok || input.Mode == "share" {
			shareIDs = append(shareIDs, version.ID)
			continue
		}
		if input.Mode != "fixed" {
			return nil, fmt.Errorf("unsupported traffic version mode: %s", input.Mode)
		}
		traffic[version.ID] = input.Weight
		fixedTotal += input.Weight
	}

	if fixedTotal > 100 {
		return nil, ErrTrafficWeightInvalid
	}

	remaining := 100 - fixedTotal
	if len(shareIDs) > 0 && remaining > 0 {
		assignEvenTraffic(traffic, shareIDs, remaining)
	}

	return traffic, nil
}

func assignEvenTraffic(traffic map[int32]int32, ids []int32, total int32) {
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	base, rem := splitEvenly(total, len(ids))
	for _, id := range ids {
		traffic[id] += base
		if rem > 0 {
			traffic[id]++
			rem--
		}
	}
}

func splitEvenly(total int32, count int) (int32, int32) {
	var count32 int32
	for range count {
		count32++
	}
	return total / count32, total % count32
}

func (s *Service) GetWorkflowInputSchema(ctx context.Context, workflowID int32) (json.RawMessage, error) {
	wf, err := s.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get workflow: %w", err)
	}
	if len(wf.InputSchema) == 0 {
		schema, err := marshalWorkflowInputSchema(map[string]workflowInputSchemaField{})
		if err != nil {
			return nil, err
		}
		return json.RawMessage(schema), nil
	}
	return json.RawMessage(wf.InputSchema), nil
}

func (s *Service) UpsertWorkflowInputSchemaField(ctx context.Context, workflowID int32, req InputSchemaFieldRequest) (json.RawMessage, error) {
	wf, err := s.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get workflow: %w", err)
	}
	schemaJSON, err := upsertWorkflowInputSchemaField(wf.InputSchema, req)
	if err != nil {
		return nil, err
	}
	updated, err := s.store.UpdateWorkflowInputSchema(ctx, db.UpdateWorkflowInputSchemaParams{
		ID:          workflowID,
		InputSchema: schemaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("update workflow input schema: %w", err)
	}
	if err := s.store.InvalidateWorkflowVersionsByWorkflowID(ctx, workflowID); err != nil {
		return nil, fmt.Errorf("invalidate workflow versions: %w", err)
	}
	return json.RawMessage(updated.InputSchema), nil
}

func (s *Service) DeleteWorkflowInputSchemaField(ctx context.Context, workflowID int32, fieldName string) (json.RawMessage, error) {
	wf, err := s.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get workflow: %w", err)
	}
	schemaJSON, err := deleteWorkflowInputSchemaField(wf.InputSchema, fieldName)
	if err != nil {
		return nil, err
	}
	updated, err := s.store.UpdateWorkflowInputSchema(ctx, db.UpdateWorkflowInputSchemaParams{
		ID:          workflowID,
		InputSchema: schemaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("update workflow input schema: %w", err)
	}
	if err := s.store.InvalidateWorkflowVersionsByWorkflowID(ctx, workflowID); err != nil {
		return nil, fmt.Errorf("invalidate workflow versions: %w", err)
	}
	return json.RawMessage(updated.InputSchema), nil
}

// --- Step CRUD ---

// CreateStep СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ С€Р°Рі РІ РІРµСЂСЃРёРё workflow (WF-03).
// Р•СЃР»Рё step_type=task Рё work_type_id СѓРєР°Р·Р°РЅ, РЅРѕ worker_settings_revision_id РЅРµС‚ вЂ”
// Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРё РЅР°С…РѕРґРёС‚ РїРѕСЃР»РµРґРЅСЋСЋ СЂРµРІРёР·РёСЋ РґР»СЏ РґР°РЅРЅРѕРіРѕ work_type.
func (s *Service) CreateStep(ctx context.Context, versionID int32, req CreateStepRequest) (db.WorkflowStep, error) {
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

	params := db.CreateWorkflowStepParams{
		WorkflowVersionID: versionID,
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

	// РђРІС‚РѕРїРѕРґСЃС‚Р°РЅРѕРІРєР° revision РґР»СЏ task-С€Р°РіРѕРІ
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

// resolveLatestRevision РЅР°С…РѕРґРёС‚ РїРѕСЃР»РµРґРЅСЋСЋ СЂРµРІРёР·РёСЋ РЅР°СЃС‚СЂРѕРµРє РґР»СЏ work_type.
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

// ListSteps РІРѕР·РІСЂР°С‰Р°РµС‚ РІСЃРµ С€Р°РіРё РІРµСЂСЃРёРё workflow.
func (s *Service) ListSteps(ctx context.Context, versionID int32) ([]db.WorkflowStep, error) {
	return s.store.ListWorkflowStepsByVersionID(ctx, versionID)
}

// UpdateStep РѕР±РЅРѕРІР»СЏРµС‚ С€Р°Рі (РІРєР»СЋС‡Р°СЏ input_mapping вЂ” WF-05).
func (s *Service) UpdateStep(ctx context.Context, stepID int32, req UpdateStepRequest) (db.WorkflowStep, error) {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return db.WorkflowStep{}, fmt.Errorf("get step: %w", err)
	}
	if isStartWorkflowStep(current) {
		return db.WorkflowStep{}, ErrStartStepProtected
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

	params := db.UpdateWorkflowStepParams{
		ID:                       stepID,
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

// DeleteStep РјСЏРіРєРѕ СѓРґР°Р»СЏРµС‚ С€Р°Рі Рё РєР°СЃРєР°РґРЅРѕ СѓРґР°Р»СЏРµС‚ РµРіРѕ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё.
func (s *Service) UpdateTaskSettings(ctx context.Context, stepID int32, req UpdateTaskSettingsRequest) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if current.StepType != string(dagpkg.StepTypeTask) || !current.WorkerSettingsRevisionID.Valid {
		return ErrTaskStepRequired
	}

	settingsSchema, err := s.settingsSchemaForStep(ctx, current)
	if err != nil {
		return err
	}
	if err := schemadialect.ValidateRaw(settingsSchema.SettingsSchema, req.SettingsData); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSettings, err)
	}

	if _, err := s.store.UpdateWorkerSettingsRevisionSettings(ctx, db.UpdateWorkerSettingsRevisionSettingsParams{
		ID:           current.WorkerSettingsRevisionID.Int32,
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
	if err := validateInputMappingShape(req.InputMapping); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInputMapping, err)
	}
	if err := s.validateTaskInputMapping(ctx, current, req.InputMapping); err != nil {
		return err
	}

	_, err = s.store.UpdateWorkflowStep(ctx, db.UpdateWorkflowStepParams{
		ID:                       current.ID,
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
	version, err := s.store.GetWorkflowVersionByID(ctx, current.WorkflowVersionID)
	if err != nil {
		return fmt.Errorf("get workflow version: %w", err)
	}
	workflow, err := s.store.GetWorkflowByID(ctx, version.WorkflowID)
	if err != nil {
		return fmt.Errorf("get workflow: %w", err)
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
	for _, validationErr := range validateStepInputsFilled(steps, deps, workflow.InputSchema) {
		if validationErr.StepID == current.ID {
			return fmt.Errorf("%w: %s", ErrInvalidInputMapping, validationErr.Message)
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

	if err := s.removeInboundInputMappings(ctx, current.WorkflowVersionID, stepID); err != nil {
		return fmt.Errorf("delete step input mappings: %w", err)
	}

	// РЈРґР°Р»СЏРµРј Р·Р°РІРёСЃРёРјРѕСЃС‚Рё С€Р°РіР° РїРµСЂРµРґ РµРіРѕ СѓРґР°Р»РµРЅРёРµРј
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

// ListDependencies РІРѕР·РІСЂР°С‰Р°РµС‚ РІСЃРµ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё РІРµСЂСЃРёРё workflow.
func (s *Service) ListDependencies(ctx context.Context, versionID int32) ([]db.WorkflowStepDependency, error) {
	return s.store.ListDependenciesByVersionID(ctx, versionID)
}

// --- Dependency management ---

// CreateDependency СЃРѕР·РґР°С‘С‚ Р·Р°РІРёСЃРёРјРѕСЃС‚СЊ РјРµР¶РґСѓ С€Р°РіР°РјРё (WF-04).
func (s *Service) CreateDependency(ctx context.Context, stepID int32, req CreateDependencyRequest) error {
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

// UpdateStepPosition РѕР±РЅРѕРІР»СЏРµС‚ С‚РѕР»СЊРєРѕ РїРѕР·РёС†РёСЋ С€Р°РіР° РЅР° С…РѕР»СЃС‚Рµ.
func (s *Service) UpdateStepPosition(ctx context.Context, stepID int32, req UpdateStepPositionRequest) error {
	return s.store.UpdateWorkflowStepPosition(ctx, db.UpdateWorkflowStepPositionParams{
		ID:             stepID,
		CanvasPosition: req.CanvasPosition,
	})
}

// ListEnrichedSteps РІРѕР·РІСЂР°С‰Р°РµС‚ С€Р°РіРё СЃ РїРѕРґРіСЂСѓР¶РµРЅРЅС‹РјРё work_type meta Рё settings schemas.
// РњР°РїРїРёС‚ []byte JSONB-РїРѕР»СЏ РІ json.RawMessage РґР»СЏ РєРѕСЂСЂРµРєС‚РЅРѕР№ JSON-СЃРµСЂРёР°Р»РёР·Р°С†РёРё.
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
			StepType:          r.StepType,
			ControlSettings:   toRawMessage(r.ControlSettings),
			InputMapping:      toRawMessage(r.InputMapping),
			CanvasPosition:    toRawMessage(r.CanvasPosition),
			WorkTypeMeta:      toRawMessage(r.WorkTypeMeta),
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

// toRawMessage РєРѕРЅРІРµСЂС‚РёСЂСѓРµС‚ []byte РІ json.RawMessage, РІРѕР·РІСЂР°С‰Р°РµС‚ nil РґР»СЏ РїСѓСЃС‚С‹С… Р·РЅР°С‡РµРЅРёР№.
func toRawMessage(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return json.RawMessage(b)
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
		Expression string   `json:"expression"`
		Cases      []string `json:"cases"`
	}
	var settings switchSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("%w: %w", ErrControlSettingsInvalid, err)
	}
	if strings.TrimSpace(settings.Expression) == "" {
		return ErrControlSettingsInvalid
	}
	for _, item := range settings.Cases {
		if strings.TrimSpace(item) == "" {
			return ErrControlSettingsInvalid
		}
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

// DeleteDependency СѓРґР°Р»СЏРµС‚ Р·Р°РІРёСЃРёРјРѕСЃС‚СЊ РјРµР¶РґСѓ С€Р°РіР°РјРё.
func (s *Service) DeleteDependency(ctx context.Context, stepID, dependsOnStepID int32) error {
	if err := s.store.DeleteWorkflowStepDependency(ctx, db.DeleteWorkflowStepDependencyParams{
		StepID:          stepID,
		DependsOnStepID: dependsOnStepID,
	}); err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}
	return s.invalidateVersionForStep(ctx, stepID)
}

// --- Validation ---

// ValidateVersion Р·Р°РїСѓСЃРєР°РµС‚ РІР°Р»РёРґР°С†РёСЋ DAG РґР»СЏ РІРµСЂСЃРёРё (WF-06, WF-07).
// Р—Р°РіСЂСѓР¶Р°РµС‚ С€Р°РіРё Рё Р·Р°РІРёСЃРёРјРѕСЃС‚Рё, РІС‹Р·С‹РІР°РµС‚ dag.ValidateDAG,
// РѕР±РЅРѕРІР»СЏРµС‚ is_valid Сѓ РІРµСЂСЃРёРё.
// Р’РѕР·РІСЂР°С‰Р°РµС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІР°Р»РёРґР°С†РёРё СЃРѕ РІСЃРµРјРё РѕС€РёР±РєР°РјРё СЃСЂР°Р·Сѓ (D-12).
func (s *Service) ValidateVersion(ctx context.Context, versionID int32) (ValidateVersionResponse, error) {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("get workflow version: %w", err)
	}
	workflow, err := s.store.GetWorkflowByID(ctx, version.WorkflowID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("get workflow: %w", err)
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

	inputNames := workflowInputNamesFromSchema(workflow.InputSchema)

	// РљРѕРЅРІРµСЂС‚РёСЂСѓРµРј РІ С‚РёРїС‹ dag РїР°РєРµС‚Р°
	dagSteps := make([]dagpkg.Step, 0, len(dbSteps))
	for _, step := range dbSteps {
		dagStep := dagpkg.Step{
			ID:             step.ID,
			StepType:       dagpkg.StepType(step.StepType),
			WorkflowInputs: inputNames,
		}
		if step.ControlKind.Valid {
			dagStep.ControlKind = step.ControlKind.String
		}
		// РџР°СЂСЃРёРј input_mapping РёР· JSONB
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

	// Р’Р°Р»РёРґРёСЂСѓРµРј DAG
	validationErrors := dagpkg.ValidateDAG(dagSteps, dagDeps)
	validationErrors = append(validationErrors, validateVersionControlSteps(dbSteps)...)
	validationErrors = append(validationErrors, validateStepInputsFilled(enrichedSteps, dbDeps, workflow.InputSchema)...)
	validationErrors = append(validationErrors, validateStepSettings(enrichedSteps)...)
	isValid := len(validationErrors) == 0

	// РћР±РЅРѕРІР»СЏРµРј is_valid Сѓ РІРµСЂСЃРёРё
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

// ActivateVersion Р°РєС‚РёРІРёСЂСѓРµС‚ РІРµСЂСЃРёСЋ workflow (WF-08).
// Р“Р°СЂРґ: РІРµСЂСЃРёСЏ РґРѕР»Р¶РЅР° Р±С‹С‚СЊ is_valid=true (Pitfall 3).
func (s *Service) ActivateVersion(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	// Р“Р°СЂРґ: РЅРµР»СЊР·СЏ Р°РєС‚РёРІРёСЂРѕРІР°С‚СЊ РЅРµРІР°Р»РёРґРЅСѓСЋ РІРµСЂСЃРёСЋ
	if !version.IsValid {
		return ErrValidationRequired
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: true,
	}); err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	return nil
}

// DeactivateVersion РґРµР°РєС‚РёРІРёСЂСѓРµС‚ РІРµСЂСЃРёСЋ workflow (WF-08).
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
	if _, err := s.store.GetWorkflowVersionByID(ctx, versionID); err != nil {
		return fmt.Errorf("get version: %w", err)
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
