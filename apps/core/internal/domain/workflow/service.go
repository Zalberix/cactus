package workflow

import (
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
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// ErrValidationRequired РІР‚вЂќ Р Р†Р ВµРЎР‚РЎРѓР С‘РЎРЏ Р Р…Р Вµ Р С—РЎР‚Р С•РЎв‚¬Р В»Р В° Р Р†Р В°Р В»Р С‘Р Т‘Р В°РЎвЂ Р С‘РЎР‹ DAG.
var ErrValidationRequired = errors.New("VALIDATION_REQUIRED: Р Р†Р ВµРЎР‚РЎРѓР С‘РЎРЏ Р Т‘Р С•Р В»Р В¶Р Р…Р В° Р С—РЎР‚Р С•Р в„–РЎвЂљР С‘ Р Р†Р В°Р В»Р С‘Р Т‘Р В°РЎвЂ Р С‘РЎР‹ DAG Р С—Р ВµРЎР‚Р ВµР Т‘ Р В°Р С”РЎвЂљР С‘Р Р†Р В°РЎвЂ Р С‘Р ВµР в„–")

var ErrStartStepProtected = errors.New("START_STEP_PROTECTED: start step is managed by the system")

var ErrTrafficWeightInvalid = errors.New("TRAFFIC_WEIGHT_INVALID: custom traffic weight total must be <= 100")

var ErrTaskStepRequired = errors.New("TASK_STEP_REQUIRED: step must be a task")

var ErrControlKindInvalid = errors.New("CONTROL_KIND_INVALID: unsupported control kind")

var allowedControlKinds = map[string]struct{}{
	dagpkg.ControlKindStart: {},
	"condition":             {},
	"switch":                {},
	"delay":                 {},
}

// Service РІР‚вЂќ Р В±Р С‘Р В·Р Р…Р ВµРЎРѓ-Р В»Р С•Р С–Р С‘Р С”Р В° РЎС“Р С—РЎР‚Р В°Р Р†Р В»Р ВµР Р…Р С‘РЎРЏ workflow.
type Service struct {
	store Storage
}

// NewService РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р Р…Р С•Р Р†РЎвЂ№Р в„– workflow.Service.
func NewService(store Storage) *Service {
	return &Service{store: store}
}

// --- Workflow CRUD ---

// CreateWorkflow РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р Р…Р С•Р Р†РЎвЂ№Р в„– workflow Р Р† РЎРѓР С‘РЎРѓРЎвЂљР ВµР СР Вµ (WF-01).
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

// ListWorkflows Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎРѓР С—Р С‘РЎРѓР С•Р С” workflow РЎРѓР С‘РЎРѓРЎвЂљР ВµР СРЎвЂ№.
func (s *Service) ListWorkflows(ctx context.Context, systemID int32) ([]db.Workflow, error) {
	return s.store.ListWorkflowsBySystemID(ctx, systemID)
}

// GetWorkflow Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ workflow Р С—Р С• ID.
func (s *Service) GetWorkflow(ctx context.Context, id int32) (db.Workflow, error) {
	return s.store.GetWorkflowByID(ctx, id)
}

// UpdateWorkflow Р С•Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ Р Р…Р В°Р В·Р Р†Р В°Р Р…Р С‘Р Вµ Р С‘ Р С—РЎР‚Р С‘Р С•РЎР‚Р С‘РЎвЂљР ВµРЎвЂљ workflow.
func (s *Service) UpdateWorkflow(ctx context.Context, id int32, req UpdateWorkflowRequest) (db.Workflow, error) {
	return s.store.UpdateWorkflow(ctx, db.UpdateWorkflowParams{
		ID:          id,
		Name:        req.Name,
		Priority:    int32(req.Priority), // #nosec G115 -- priority is validated by request binding.
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
}

// DeleteWorkflow Р СРЎРЏР С–Р С”Р С• РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ workflow.
func (s *Service) DeleteWorkflow(ctx context.Context, id int32) error {
	return s.store.SoftDeleteWorkflow(ctx, id)
}

// --- Version management ---

// CreateVersion РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р Р…Р С•Р Р†РЎС“РЎР‹ Р Р†Р ВµРЎР‚РЎРѓР С‘РЎР‹ workflow (WF-02).
// Р СњР С•Р СР ВµРЎР‚ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘ Р С‘Р Р…Р С”РЎР‚Р ВµР СР ВµР Р…РЎвЂљР С‘РЎР‚РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ.
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

// ListVersions Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ workflow.
func (s *Service) ListVersions(ctx context.Context, workflowID int32) ([]db.WorkflowVersion, error) {
	return s.store.ListWorkflowVersionsByWorkflowID(ctx, workflowID)
}

// ListVersionSummaries Р В Р вЂ Р В РЎвЂўР В Р вЂ Р В Р вЂ Р В РЎвЂўР В РўвЂР РЋР С“Р РЋРЎвЂњР РЋРІР‚В°Р В РЎвЂР В Р’Вµ Р РЋРЎвЂњР В Р вЂ Р В Р’В°Р В Р’В»Р РЋР Р‰Р В Р вЂ¦Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂР В Р вЂ¦Р В Р’В°Р РЋР С“Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р вЂ¦Р В РўвЂР В Р’В°.
func (s *Service) ListVersionSummaries(ctx context.Context, workflowID int32) ([]VersionSummaryResponse, error) {
	rows, err := s.store.ListWorkflowVersionSummariesByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("list version summaries: %w", err)
	}

	summaries := make([]VersionSummaryResponse, 0, len(rows))
	for _, r := range rows {
		summary := VersionSummaryResponse{
			ID:             r.ID,
			WorkflowID:     r.WorkflowID,
			Name:           r.Name,
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

// UpdateVersionName Р В РЎвЂўР В Р’В±Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В Р’В»Р РЋР РЏР РЋРІР‚С™Р РЋР Р‰ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР В РЎВР В Р’В°Р РЋРІР‚С™Р В Р’В° Р В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР С“Р В РЎвЂР В РЎвЂ.
func (s *Service) UpdateVersionName(ctx context.Context, versionID int32, req UpdateVersionNameRequest) (db.WorkflowVersion, error) {
	return s.store.UpdateWorkflowVersionName(ctx, db.UpdateWorkflowVersionNameParams{
		ID:   versionID,
		Name: pgtype.Text{String: req.Name, Valid: true},
	})
}

// CopyVersion Р В РЎвЂ“Р В РЎвЂўР В Р’В»Р РЋР Р‰Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР РЋРЎвЂњР РЋР РЏ Р В РЎвЂР В Р’В»Р В РЎвЂР РЋР С“Р В Р вЂ Р РЋРІР‚в„–Р В РІвЂћвЂ“ copy Р В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР С“Р В РЎвЂР В РЎвЂ.
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

// UpdateWorkflowTraffic Р В РўвЂР В Р’В°Р РЋРІвЂљВ¬Р В Р’В»Р РЋР Р‰Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р РЋР Р‰ traffic Р“С’Р’Р†Р“С’Р’С• Р“С’Р’Р†Р“С’Р’С—Р“вЂР–вЂ™Р“С’Р’В»Р“С’Р’Р…Р“С’Р’ВµР“С’Р’Р…Р“С’Р’С‘Р“С’Р’С‘ workflow.
func (s *Service) UpdateWorkflowTraffic(ctx context.Context, workflowID int32, req UpdateTrafficRequest) error {
	versions, err := s.store.ListAllWorkflowVersionsByWorkflowID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("list versions: %w", err)
	}

	activeVersions := make([]db.WorkflowVersion, 0, len(versions))
	activeSet := make(map[int32]db.WorkflowVersion, len(versions))
	for _, version := range versions {
		if version.IsActive {
			activeVersions = append(activeVersions, version)
			activeSet[version.ID] = version
		}
	}

	traffic := make(map[int32]int32, len(versions))
	if len(req.Versions) > 0 {
		traffic, err = distributePerVersionTraffic(activeVersions, activeSet, req.Versions)
		if err != nil {
			return err
		}
	} else {
		switch req.Mode {
		case "equal":
			if len(activeVersions) > 0 {
				ids := make([]int32, 0, len(activeVersions))
				for _, version := range activeVersions {
					ids = append(ids, version.ID)
				}
				sort.SliceStable(ids, func(i, j int) bool {
					return ids[i] < ids[j]
				})

				base := int32(100 / len(activeVersions))
				rem := int32(100 % len(activeVersions))
				for _, id := range ids {
					traffic[id] = base
					if rem > 0 {
						traffic[id]++
						rem--
					}
				}
			}
		case "custom":
			assigned := make(map[int32]int32, len(req.Weights))
			for _, w := range req.Weights {
				if _, ok := activeSet[w.VersionID]; !ok {
					return fmt.Errorf("version %d is not active", w.VersionID)
				}
				assigned[w.VersionID] = w.Weight
			}

			var sum int32
			unassigned := 0
			for _, version := range activeVersions {
				w := assigned[version.ID]
				traffic[version.ID] = w
				sum += w
				if _, ok := assigned[version.ID]; !ok {
					unassigned++
				}
			}
			if sum > 100 {
				return ErrTrafficWeightInvalid
			}
			remaining := 100 - sum
			if unassigned > 0 && remaining > 0 {
				base := remaining / int32(unassigned)
				rem := remaining % int32(unassigned)
				for _, version := range activeVersions {
					if _, ok := assigned[version.ID]; ok {
						continue
					}
					traffic[version.ID] += base
					if rem > 0 {
						traffic[version.ID]++
						rem--
					}
				}
			}
		default:
			return fmt.Errorf("unsupported traffic mode: %s", req.Mode)
		}
	}

	for _, version := range versions {
		weight := int32(0)
		if version.IsActive {
			weight = traffic[version.ID]
		}
		if _, err := s.store.UpdateWorkflowVersionTrafficWeightIncludingDeleted(ctx, db.UpdateWorkflowVersionTrafficWeightParams{
			ID:            version.ID,
			TrafficWeight: weight,
		}); err != nil {
			return fmt.Errorf("update traffic weight for version %d: %w", version.ID, err)
		}
	}

	return nil
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
		sort.SliceStable(shareIDs, func(i, j int) bool {
			return shareIDs[i] < shareIDs[j]
		})
		base := remaining / int32(len(shareIDs))
		rem := remaining % int32(len(shareIDs))
		for _, id := range shareIDs {
			traffic[id] = base
			if rem > 0 {
				traffic[id]++
				rem--
			}
		}
	}

	return traffic, nil
}

func (s *Service) ListWorkflowInputs(ctx context.Context, versionID int32) ([]WorkflowInputResponse, error) {
	inputs, err := s.store.ListWorkflowVersionInputs(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list workflow inputs: %w", err)
	}
	return workflowInputResponses(inputs), nil
}

func (s *Service) CreateWorkflowInput(ctx context.Context, versionID int32, req WorkflowInputRequest) (WorkflowInputResponse, error) {
	input, err := s.store.CreateWorkflowVersionInput(ctx, db.CreateWorkflowVersionInputParams{
		WorkflowVersionID: versionID,
		Name:              req.Name,
		Type:              req.Type,
		Required:          req.Required,
		Description:       pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return WorkflowInputResponse{}, fmt.Errorf("create workflow input: %w", err)
	}
	if err := s.invalidateVersion(ctx, versionID); err != nil {
		return WorkflowInputResponse{}, err
	}
	return workflowInputResponse(input), nil
}

func (s *Service) UpdateWorkflowInput(ctx context.Context, inputID int32, req WorkflowInputRequest) (WorkflowInputResponse, error) {
	input, err := s.store.UpdateWorkflowVersionInput(ctx, db.UpdateWorkflowVersionInputParams{
		ID:          inputID,
		Name:        req.Name,
		Type:        req.Type,
		Required:    req.Required,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return WorkflowInputResponse{}, fmt.Errorf("update workflow input: %w", err)
	}
	if err := s.invalidateVersion(ctx, input.WorkflowVersionID); err != nil {
		return WorkflowInputResponse{}, err
	}
	return workflowInputResponse(input), nil
}

func (s *Service) DeleteWorkflowInput(ctx context.Context, inputID int32) error {
	input, err := s.store.GetWorkflowVersionInputByID(ctx, inputID)
	if err != nil {
		return fmt.Errorf("get workflow input: %w", err)
	}
	if err := s.store.SoftDeleteWorkflowVersionInput(ctx, inputID); err != nil {
		return fmt.Errorf("delete workflow input: %w", err)
	}
	return s.invalidateVersion(ctx, input.WorkflowVersionID)
}

func (s *Service) GetVersionInputSchema(ctx context.Context, versionID int32) (json.RawMessage, error) {
	inputs, err := s.store.ListWorkflowVersionInputs(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("list workflow inputs: %w", err)
	}
	schema, err := workflowInputsSchema(inputs)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(schema), nil
}

// --- Step CRUD ---

// CreateStep РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р Р…Р С•Р Р†РЎвЂ№Р в„– РЎв‚¬Р В°Р С– Р Р† Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ workflow (WF-03).
// Р вЂўРЎРѓР В»Р С‘ step_type=task Р С‘ work_type_id РЎС“Р С”Р В°Р В·Р В°Р Р…, Р Р…Р С• worker_settings_revision_id Р Р…Р ВµРЎвЂљ РІР‚вЂќ
// Р В°Р Р†РЎвЂљР С•Р СР В°РЎвЂљР С‘РЎвЂЎР ВµРЎРѓР С”Р С‘ Р Р…Р В°РЎвЂ¦Р С•Р Т‘Р С‘РЎвЂљ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…РЎР‹РЎР‹ РЎР‚Р ВµР Р†Р С‘Р В·Р С‘РЎР‹ Р Т‘Р В»РЎРЏ Р Т‘Р В°Р Р…Р Р…Р С•Р С–Р С• work_type.
func (s *Service) CreateStep(ctx context.Context, versionID int32, req CreateStepRequest) (db.WorkflowStep, error) {
	if isStartStepRequest(req.StepType, req.ControlKind) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if err := validateControlKind(req.StepType, req.ControlKind); err != nil {
		return db.WorkflowStep{}, err
	}

	params := db.CreateWorkflowStepParams{
		WorkflowVersionID: versionID,
		StepType:          req.StepType,
		InputMapping:      req.InputMapping,
		ControlSettings:   req.ControlSettings,
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

	// Р С’Р Р†РЎвЂљР С•Р С—Р С•Р Т‘РЎРѓРЎвЂљР В°Р Р…Р С•Р Р†Р С”Р В° revision Р Т‘Р В»РЎРЏ task-РЎв‚¬Р В°Р С–Р С•Р Р†
	if req.StepType == "task" && req.WorkerSettingsRevisionID == nil {
		if req.WorkerSettingsSchemaID != nil {
			revID, err := s.createPrivateRevisionForSchema(ctx, *req.WorkerSettingsSchemaID)
			if err != nil {
				return db.WorkflowStep{}, fmt.Errorf("create private revision: %w", err)
			}
			params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revID, Valid: true}
		} else if req.WorkTypeID != nil {
			revID, err := s.resolvePrivateRevision(ctx, *req.WorkTypeID)
			if err != nil {
				return db.WorkflowStep{}, fmt.Errorf("auto-resolve revision: %w", err)
			}
			params.WorkerSettingsRevisionID = pgtype.Int4{Int32: revID, Valid: true}
		}
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

// resolveLatestRevision Р Р…Р В°РЎвЂ¦Р С•Р Т‘Р С‘РЎвЂљ Р С—Р С•РЎРѓР В»Р ВµР Т‘Р Р…РЎР‹РЎР‹ РЎР‚Р ВµР Р†Р С‘Р В·Р С‘РЎР‹ Р Р…Р В°РЎРѓРЎвЂљРЎР‚Р С•Р ВµР С” Р Т‘Р В»РЎРЏ work_type.
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
	if len(revisions) == 0 {
		settingsData = []byte(`{}`)
	} else {
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

// ListSteps Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ РЎв‚¬Р В°Р С–Р С‘ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ workflow.
func (s *Service) ListSteps(ctx context.Context, versionID int32) ([]db.WorkflowStep, error) {
	return s.store.ListWorkflowStepsByVersionID(ctx, versionID)
}

// UpdateStep Р С•Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎв‚¬Р В°Р С– (Р Р†Р С”Р В»РЎР‹РЎвЂЎР В°РЎРЏ input_mapping РІР‚вЂќ WF-05).
func (s *Service) UpdateStep(ctx context.Context, stepID int32, req UpdateStepRequest) (db.WorkflowStep, error) {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return db.WorkflowStep{}, fmt.Errorf("get step: %w", err)
	}
	if isStartWorkflowStep(current) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if isStartStepRequest(req.StepType, req.ControlKind) {
		return db.WorkflowStep{}, ErrStartStepProtected
	}
	if err := validateControlKind(req.StepType, req.ControlKind); err != nil {
		return db.WorkflowStep{}, err
	}

	params := db.UpdateWorkflowStepParams{
		ID:              stepID,
		StepType:        req.StepType,
		InputMapping:    req.InputMapping,
		ControlSettings: req.ControlSettings,
		CanvasPosition:  req.CanvasPosition,
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
	updated, err := s.store.UpdateWorkflowStep(ctx, params)
	if err != nil {
		return db.WorkflowStep{}, err
	}
	if err := s.invalidateVersionForStep(ctx, stepID); err != nil {
		return db.WorkflowStep{}, err
	}
	return updated, nil
}

// DeleteStep Р СРЎРЏР С–Р С”Р С• РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ РЎв‚¬Р В°Р С– Р С‘ Р С”Р В°РЎРѓР С”Р В°Р Т‘Р Р…Р С• РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ Р ВµР С–Р С• Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљР С‘.
func (s *Service) UpdateTaskSettings(ctx context.Context, stepID int32, req UpdateTaskSettingsRequest) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if current.StepType != string(dagpkg.StepTypeTask) || !current.WorkerSettingsRevisionID.Valid {
		return ErrTaskStepRequired
	}

	if _, err := s.store.UpdateWorkerSettingsRevisionSettings(ctx, db.UpdateWorkerSettingsRevisionSettingsParams{
		ID:           current.WorkerSettingsRevisionID.Int32,
		SettingsData: req.SettingsData,
	}); err != nil {
		return fmt.Errorf("update task settings revision: %w", err)
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
func (s *Service) DeleteStep(ctx context.Context, stepID int32) error {
	current, err := s.store.GetWorkflowStepByID(ctx, stepID)
	if err != nil {
		return fmt.Errorf("get step: %w", err)
	}
	if isStartWorkflowStep(current) {
		return ErrStartStepProtected
	}

	// Р Р€Р Т‘Р В°Р В»РЎРЏР ВµР С Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљР С‘ РЎв‚¬Р В°Р С–Р В° Р С—Р ВµРЎР‚Р ВµР Т‘ Р ВµР С–Р С• РЎС“Р Т‘Р В°Р В»Р ВµР Р…Р С‘Р ВµР С
	if err := s.store.DeleteDependenciesByStepID(ctx, stepID); err != nil {
		return fmt.Errorf("delete step dependencies: %w", err)
	}
	if err := s.store.SoftDeleteWorkflowStep(ctx, stepID); err != nil {
		return fmt.Errorf("delete step: %w", err)
	}
	return s.invalidateVersion(ctx, current.WorkflowVersionID)
}

// ListDependencies Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљР С‘ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ workflow.
func (s *Service) ListDependencies(ctx context.Context, versionID int32) ([]db.WorkflowStepDependency, error) {
	return s.store.ListDependenciesByVersionID(ctx, versionID)
}

// --- Dependency management ---

// CreateDependency РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљРЎРЉ Р СР ВµР В¶Р Т‘РЎС“ РЎв‚¬Р В°Р С–Р В°Р СР С‘ (WF-04).
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

// UpdateStepPosition Р С•Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎвЂљР С•Р В»РЎРЉР С”Р С• Р С—Р С•Р В·Р С‘РЎвЂ Р С‘РЎР‹ РЎв‚¬Р В°Р С–Р В° Р Р…Р В° РЎвЂ¦Р С•Р В»РЎРѓРЎвЂљР Вµ.
func (s *Service) UpdateStepPosition(ctx context.Context, stepID int32, req UpdateStepPositionRequest) error {
	return s.store.UpdateWorkflowStepPosition(ctx, db.UpdateWorkflowStepPositionParams{
		ID:             stepID,
		CanvasPosition: req.CanvasPosition,
	})
}

// ListEnrichedSteps Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎв‚¬Р В°Р С–Р С‘ РЎРѓ Р С—Р С•Р Т‘Р С–РЎР‚РЎС“Р В¶Р ВµР Р…Р Р…РЎвЂ№Р СР С‘ work_type meta Р С‘ settings schemas.
// Р СљР В°Р С—Р С—Р С‘РЎвЂљ []byte JSONB-Р С—Р С•Р В»РЎРЏ Р Р† json.RawMessage Р Т‘Р В»РЎРЏ Р С”Р С•РЎР‚РЎР‚Р ВµР С”РЎвЂљР Р…Р С•Р в„– JSON-РЎРѓР ВµРЎР‚Р С‘Р В°Р В»Р С‘Р В·Р В°РЎвЂ Р С‘Р С‘.
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

// toRawMessage Р С”Р С•Р Р…Р Р†Р ВµРЎР‚РЎвЂљР С‘РЎР‚РЎС“Р ВµРЎвЂљ []byte Р Р† json.RawMessage, Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ nil Р Т‘Р В»РЎРЏ Р С—РЎС“РЎРѓРЎвЂљРЎвЂ№РЎвЂ¦ Р В·Р Р…Р В°РЎвЂЎР ВµР Р…Р С‘Р в„–.
func toRawMessage(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return json.RawMessage(b)
}

func workflowInputResponse(input db.WorkflowVersionInput) WorkflowInputResponse {
	resp := WorkflowInputResponse{
		ID:                input.ID,
		WorkflowVersionID: input.WorkflowVersionID,
		Name:              input.Name,
		Type:              input.Type,
		Required:          input.Required,
	}
	if input.Description.Valid {
		resp.Description = input.Description.String
	}
	return resp
}

func workflowInputResponses(inputs []db.WorkflowVersionInput) []WorkflowInputResponse {
	responses := make([]WorkflowInputResponse, 0, len(inputs))
	for _, input := range inputs {
		responses = append(responses, workflowInputResponse(input))
	}
	return responses
}

func workflowInputNames(inputs []db.WorkflowVersionInput) []string {
	names := make([]string, 0, len(inputs))
	for _, input := range inputs {
		names = append(names, input.Name)
	}
	return names
}

func workflowInputsSchema(inputs []db.WorkflowVersionInput) ([]byte, error) {
	properties := make(map[string]interface{}, len(inputs))
	required := make([]string, 0, len(inputs))
	for _, input := range inputs {
		properties[input.Name] = map[string]interface{}{"type": input.Type}
		if input.Required {
			required = append(required, input.Name)
		}
	}
	sort.Strings(required)

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return json.Marshal(schema)
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

// DeleteDependency РЎС“Р Т‘Р В°Р В»РЎРЏР ВµРЎвЂљ Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљРЎРЉ Р СР ВµР В¶Р Т‘РЎС“ РЎв‚¬Р В°Р С–Р В°Р СР С‘.
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

// ValidateVersion Р В·Р В°Р С—РЎС“РЎРѓР С”Р В°Р ВµРЎвЂљ Р Р†Р В°Р В»Р С‘Р Т‘Р В°РЎвЂ Р С‘РЎР‹ DAG Р Т‘Р В»РЎРЏ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ (WF-06, WF-07).
// Р вЂ”Р В°Р С–РЎР‚РЎС“Р В¶Р В°Р ВµРЎвЂљ РЎв‚¬Р В°Р С–Р С‘ Р С‘ Р В·Р В°Р Р†Р С‘РЎРѓР С‘Р СР С•РЎРѓРЎвЂљР С‘, Р Р†РЎвЂ№Р В·РЎвЂ№Р Р†Р В°Р ВµРЎвЂљ dag.ValidateDAG,
// Р С•Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµРЎвЂљ is_valid РЎС“ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘.
// Р вЂ™Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎР‚Р ВµР В·РЎС“Р В»РЎРЉРЎвЂљР В°РЎвЂљ Р Р†Р В°Р В»Р С‘Р Т‘Р В°РЎвЂ Р С‘Р С‘ РЎРѓР С• Р Р†РЎРѓР ВµР СР С‘ Р С•РЎв‚¬Р С‘Р В±Р С”Р В°Р СР С‘ РЎРѓРЎР‚Р В°Р В·РЎС“ (D-12).
func (s *Service) ValidateVersion(ctx context.Context, versionID int32) (ValidateVersionResponse, error) {
	dbSteps, err := s.store.ListWorkflowStepsByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list steps: %w", err)
	}

	dbDeps, err := s.store.ListDependenciesByVersionID(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list dependencies: %w", err)
	}

	versionInputs, err := s.store.ListWorkflowVersionInputs(ctx, versionID)
	if err != nil {
		return ValidateVersionResponse{}, fmt.Errorf("list workflow inputs: %w", err)
	}
	inputNames := workflowInputNames(versionInputs)

	// Р С™Р С•Р Р…Р Р†Р ВµРЎР‚РЎвЂљР С‘РЎР‚РЎС“Р ВµР С Р Р† РЎвЂљР С‘Р С—РЎвЂ№ dag Р С—Р В°Р С”Р ВµРЎвЂљР В°
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
		// Р СџР В°РЎР‚РЎРѓР С‘Р С input_mapping Р С‘Р В· JSONB
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

	// Р вЂ™Р В°Р В»Р С‘Р Т‘Р С‘РЎР‚РЎС“Р ВµР С DAG
	validationErrors := dagpkg.ValidateDAG(dagSteps, dagDeps)
	isValid := len(validationErrors) == 0

	// Р С›Р В±Р Р…Р С•Р Р†Р В»РЎРЏР ВµР С is_valid РЎС“ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘
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

// ActivateVersion Р В°Р С”РЎвЂљР С‘Р Р†Р С‘РЎР‚РЎС“Р ВµРЎвЂљ Р Р†Р ВµРЎР‚РЎРѓР С‘РЎР‹ workflow (WF-08).
// Р вЂњР В°РЎР‚Р Т‘: Р Р†Р ВµРЎР‚РЎРѓР С‘РЎРЏ Р Т‘Р С•Р В»Р В¶Р Р…Р В° Р В±РЎвЂ№РЎвЂљРЎРЉ is_valid=true (Pitfall 3).
func (s *Service) ActivateVersion(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	// Р вЂњР В°РЎР‚Р Т‘: Р Р…Р ВµР В»РЎРЉР В·РЎРЏ Р В°Р С”РЎвЂљР С‘Р Р†Р С‘РЎР‚Р С•Р Р†Р В°РЎвЂљРЎРЉ Р Р…Р ВµР Р†Р В°Р В»Р С‘Р Т‘Р Р…РЎС“РЎР‹ Р Р†Р ВµРЎР‚РЎРѓР С‘РЎР‹
	if !version.IsValid {
		return ErrValidationRequired
	}

	if _, err := s.store.UpdateWorkflowVersionActive(ctx, db.UpdateWorkflowVersionActiveParams{
		ID:       versionID,
		IsActive: true,
	}); err != nil {
		return fmt.Errorf("activate version: %w", err)
	}

	// WF-09: Р С—Р ВµРЎР‚Р ВµРЎРѓРЎвЂЎР С‘РЎвЂљРЎвЂ№Р Р†Р В°Р ВµР С input_validation
	return s.regenerateInputValidation(ctx, version.WorkflowID)
}

// DeactivateVersion Р Т‘Р ВµР В°Р С”РЎвЂљР С‘Р Р†Р С‘РЎР‚РЎС“Р ВµРЎвЂљ Р Р†Р ВµРЎР‚РЎРѓР С‘РЎР‹ workflow (WF-08).
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

	// WF-09: Р С—Р ВµРЎР‚Р ВµРЎРѓРЎвЂЎР С‘РЎвЂљРЎвЂ№Р Р†Р В°Р ВµР С input_validation
	return s.regenerateInputValidation(ctx, version.WorkflowID)
}

// DeleteVersion soft-deletes a workflow version and its steps.
func (s *Service) DeleteVersion(ctx context.Context, versionID int32) error {
	version, err := s.store.GetWorkflowVersionByID(ctx, versionID)
	if err != nil {
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

	if version.IsActive {
		return s.regenerateInputValidation(ctx, version.WorkflowID)
	}
	return nil
}

// regenerateInputValidation Р С—Р ВµРЎР‚Р ВµРЎРѓРЎвЂЎР С‘РЎвЂљРЎвЂ№Р Р†Р В°Р ВµРЎвЂљ input_validation workflow (WF-09).
// Р ВРЎвЂљР ВµРЎР‚Р С‘РЎР‚РЎС“Р ВµРЎвЂљ Р вЂ™Р РЋР вЂў Р В°Р С”РЎвЂљР С‘Р Р†Р Р…РЎвЂ№Р Вµ Р Р†Р ВµРЎР‚РЎРѓР С‘Р С‘ (Pitfall 5), РЎРѓР С•Р В±Р С‘РЎР‚Р В°Р ВµРЎвЂљ union JSON Schema
// Р С‘Р В· Р Р†РЎРѓР ВµРЎвЂ¦ $.message.value.* Р С—Р С•Р В»Р ВµР в„– Р Р† input_mapping.
func (s *Service) regenerateInputValidation(ctx context.Context, workflowID int32) error {
	activeVersions, err := s.store.ListActiveWorkflowVersions(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("list active versions: %w", err)
	}

	if len(activeVersions) == 0 {
		emptySchema := []byte(`{"type":"object","properties":{}}`)
		_, err = s.store.UpdateWorkflowInputValidation(ctx, db.UpdateWorkflowInputValidationParams{
			ID:              workflowID,
			InputValidation: emptySchema,
		})
		return err
	}

	inputByName := make(map[string]db.WorkflowVersionInput)
	for _, version := range activeVersions {
		inputs, err := s.store.ListWorkflowVersionInputs(ctx, version.ID)
		if err != nil {
			return fmt.Errorf("list workflow inputs for version %d: %w", version.ID, err)
		}
		for _, input := range inputs {
			existing, ok := inputByName[input.Name]
			if !ok {
				inputByName[input.Name] = input
				continue
			}
			existing.Required = existing.Required || input.Required
			if existing.Type == "" {
				existing.Type = input.Type
			}
			inputByName[input.Name] = existing
		}
	}

	inputs := make([]db.WorkflowVersionInput, 0, len(inputByName))
	for _, input := range inputByName {
		inputs = append(inputs, input)
	}
	schemaJSON, err := workflowInputsSchema(inputs)
	if err != nil {
		return fmt.Errorf("marshal input_validation schema: %w", err)
	}

	_, err = s.store.UpdateWorkflowInputValidation(ctx, db.UpdateWorkflowInputValidationParams{
		ID:              workflowID,
		InputValidation: schemaJSON,
	})
	return err
}
