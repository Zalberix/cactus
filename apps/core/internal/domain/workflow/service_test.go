package workflow_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
	"github.com/zalberix/cactus/apps/core/internal/domain/workflow"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// mockStorage — минимальный мок Storage для тестов.
type mockStorage struct {
	version           db.WorkflowVersion
	getVersionErr     error
	activateErr       error
	updateValidErr    error
	updateNameErr     error
	steps             []db.WorkflowStep
	enrichedSteps     []db.ListEnrichedStepsByVersionIDRow
	deps              []db.WorkflowStepDependency
	activeVersions    []db.WorkflowVersion
	listVersions      []db.WorkflowVersion
	maxVersionNumber  int32
	workflow          db.Workflow
	revisionsBySchema map[int32][]db.WorkerSettingsRevision
	settingsSchema    db.WorkerSettingsSchema
	settingsSchemaErr error
	compatibilities   []db.WorkflowVersionInputSchemaCompatibility
	inputSchemas      map[int32]db.WorkflowInputSchema

	listVersionSummaries               []db.ListWorkflowVersionSummariesByWorkflowIDRow
	listVersionSummariesErr            error
	createdWorkflowVersion             db.WorkflowVersion
	lastCreateWorkflowVersionArg       db.CreateWorkflowVersionParams
	createWorkflowVersionErr           error
	createdWorkflowSteps               []db.WorkflowStep
	createWorkflowStepParams           []db.CreateWorkflowStepParams
	createStepErr                      error
	createStepNextID                   int32
	createDependencyParams             []db.CreateWorkflowStepDependencyParams
	createDependencyErr                error
	createdRevisionSettings            []byte
	updateRevisionSettingsArg          db.UpdateWorkerSettingsRevisionSettingsParams
	lastUpdateVersionNameArg           db.UpdateWorkflowVersionNameParams
	lastCloneWorkerSettingsRevisionArg db.CloneWorkerSettingsRevisionParams
	cloneWorkerSettingsRevisionResp    db.WorkerSettingsRevision
	cloneWorkerSettingsRevisionErr     error
	getWorkflowStepErr                 error
	workflowStep                       db.WorkflowStep
	// capture args
	lastUpdateValidArg             db.UpdateWorkflowVersionValidParams
	lastUpdateActiveArg            db.UpdateWorkflowVersionActiveParams
	lastUpdateInputSchemaStatusArg db.UpdateWorkflowInputSchemaStatusParams
	lastUpdateStepArg              db.UpdateWorkflowStepParams
	invalidatedWorkflowID          int32
}

func (m *mockStorage) WithTx(_ context.Context, _ func(q *db.Queries) error) error {
	return nil
}

func (m *mockStorage) CreateWorkflow(_ context.Context, _ db.CreateWorkflowParams) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) GetWorkflowByID(_ context.Context, _ int32) (db.Workflow, error) {
	return m.workflow, nil
}

func (m *mockStorage) ListWorkflowsBySystemID(_ context.Context, _ int32) ([]db.Workflow, error) {
	return []db.Workflow{{ID: 10, SystemID: 3, Name: "Process", Priority: 1}}, nil
}

func (m *mockStorage) UpdateWorkflow(_ context.Context, _ db.UpdateWorkflowParams) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) SoftDeleteWorkflow(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) CreateWorkflowVersion(_ context.Context, arg db.CreateWorkflowVersionParams) (db.WorkflowVersion, error) {
	m.lastCreateWorkflowVersionArg = arg
	if m.createWorkflowVersionErr != nil {
		return db.WorkflowVersion{}, m.createWorkflowVersionErr
	}
	if m.createdWorkflowVersion == (db.WorkflowVersion{}) {
		m.createdWorkflowVersion = db.WorkflowVersion{
			ID:             arg.WorkflowID + 1000,
			WorkflowID:     arg.WorkflowID,
			VersionNumber:  arg.VersionNumber,
			Name:           arg.Name,
			IsValid:        arg.IsValid,
			IsActive:       arg.IsActive,
		}
	}
	m.createdWorkflowVersion.WorkflowID = arg.WorkflowID
	return m.createdWorkflowVersion, nil
}

func (m *mockStorage) GetWorkflowVersionByID(_ context.Context, _ int32) (db.WorkflowVersion, error) {
	return m.version, m.getVersionErr
}

func (m *mockStorage) GetMaxVersionNumberByWorkflowID(_ context.Context, _ int32) (int32, error) {
	return m.maxVersionNumber, nil
}

func (m *mockStorage) ListWorkflowVersionsByWorkflowID(_ context.Context, _ int32) ([]db.WorkflowVersion, error) {
	versions := make([]db.WorkflowVersion, 0, len(m.listVersions))
	for _, version := range m.listVersions {
		if !version.DeletedAt.Valid {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

func (m *mockStorage) ListAllWorkflowVersionsByWorkflowID(_ context.Context, _ int32) ([]db.WorkflowVersion, error) {
	return m.listVersions, nil
}

func (m *mockStorage) ListWorkflowVersionSummariesByWorkflowID(_ context.Context, _ int32) ([]db.ListWorkflowVersionSummariesByWorkflowIDRow, error) {
	return m.listVersionSummaries, m.listVersionSummariesErr
}

func (m *mockStorage) ListActiveWorkflowVersions(_ context.Context, _ int32) ([]db.WorkflowVersion, error) {
	return m.activeVersions, nil
}

func (m *mockStorage) UpdateWorkflowVersionName(_ context.Context, arg db.UpdateWorkflowVersionNameParams) (db.WorkflowVersion, error) {
	m.lastUpdateVersionNameArg = arg
	return db.WorkflowVersion{ID: arg.ID}, m.updateNameErr
}

func (m *mockStorage) UpdateWorkflowVersionValid(_ context.Context, arg db.UpdateWorkflowVersionValidParams) (db.WorkflowVersion, error) {
	m.lastUpdateValidArg = arg
	return db.WorkflowVersion{IsValid: arg.IsValid}, m.updateValidErr
}

func (m *mockStorage) UpdateWorkflowVersionActive(_ context.Context, arg db.UpdateWorkflowVersionActiveParams) (db.WorkflowVersion, error) {
	m.lastUpdateActiveArg = arg
	return db.WorkflowVersion{IsActive: arg.IsActive}, m.activateErr
}

func (m *mockStorage) InvalidateWorkflowVersionsByWorkflowID(_ context.Context, workflowID int32) error {
	m.invalidatedWorkflowID = workflowID
	return nil
}

func (m *mockStorage) SoftDeleteWorkflowVersion(_ context.Context, _ int32) error { return nil }

func (m *mockStorage) CreateWorkflowStep(_ context.Context, arg db.CreateWorkflowStepParams) (db.WorkflowStep, error) {
	m.createWorkflowStepParams = append(m.createWorkflowStepParams, arg)
	if m.createStepErr != nil {
		return db.WorkflowStep{}, m.createStepErr
	}
	if m.createStepNextID == 0 {
		m.createStepNextID = 1000
	}
	step := db.WorkflowStep{
		ID:                       m.createStepNextID,
		WorkflowVersionID:        arg.WorkflowVersionID,
		Name:                     arg.Name,
		StepType:                 arg.StepType,
		WorkTypeID:               arg.WorkTypeID,
		WorkerSettingsRevisionID: arg.WorkerSettingsRevisionID,
		ControlKind:              arg.ControlKind,
		ControlSettings:          arg.ControlSettings,
		InputMapping:             arg.InputMapping,
		CanvasPosition:           arg.CanvasPosition,
	}
	m.createStepNextID++
	m.createdWorkflowSteps = append(m.createdWorkflowSteps, step)
	return step, nil
}

func (m *mockStorage) GetWorkflowStepByID(_ context.Context, _ int32) (db.WorkflowStep, error) {
	if m.getWorkflowStepErr != nil {
		return db.WorkflowStep{}, m.getWorkflowStepErr
	}
	return m.workflowStep, nil
}

func (m *mockStorage) ListWorkflowStepsByVersionID(_ context.Context, _ int32) ([]db.WorkflowStep, error) {
	return m.steps, nil
}

func (m *mockStorage) UpdateWorkflowStep(_ context.Context, arg db.UpdateWorkflowStepParams) (db.WorkflowStep, error) {
	m.lastUpdateStepArg = arg
	return db.WorkflowStep{
		ID:                       arg.ID,
		WorkflowVersionID:        m.workflowStep.WorkflowVersionID,
		Name:                     arg.Name,
		StepType:                 arg.StepType,
		WorkTypeID:               arg.WorkTypeID,
		WorkerSettingsRevisionID: arg.WorkerSettingsRevisionID,
		ControlKind:              arg.ControlKind,
		ControlSettings:          arg.ControlSettings,
		InputMapping:             arg.InputMapping,
		CanvasPosition:           arg.CanvasPosition,
	}, nil
}
func (m *mockStorage) DeleteWorkflowStepsByVersionID(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) SoftDeleteWorkflowStep(_ context.Context, _ int32) error         { return nil }
func (m *mockStorage) CreateWorkflowStepDependency(_ context.Context, arg db.CreateWorkflowStepDependencyParams) error {
	m.createDependencyParams = append(m.createDependencyParams, arg)
	return m.createDependencyErr
}

func (m *mockStorage) ListDependenciesByVersionID(_ context.Context, _ int32) ([]db.WorkflowStepDependency, error) {
	return m.deps, nil
}
func (m *mockStorage) DeleteDependenciesByStepID(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) DeleteWorkflowStepDependency(_ context.Context, _ db.DeleteWorkflowStepDependencyParams) error {
	return nil
}

func (m *mockStorage) ListWorkerSettingsSchemasByWorkTypeID(_ context.Context, _ int32) ([]db.WorkerSettingsSchema, error) {
	return nil, nil
}

func (m *mockStorage) GetWorkerSettingsSchemaByID(_ context.Context, _ int32) (db.WorkerSettingsSchema, error) {
	return m.settingsSchema, m.settingsSchemaErr
}

func (m *mockStorage) ListWorkerSettingsRevisionsBySchemaID(_ context.Context, schemaID int32) ([]db.WorkerSettingsRevision, error) {
	if m.revisionsBySchema != nil {
		return m.revisionsBySchema[schemaID], nil
	}
	return nil, nil
}

func (m *mockStorage) CreateWorkerSettingsRevision(_ context.Context, arg db.CreateWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error) {
	m.createdRevisionSettings = arg.SettingsData
	return db.WorkerSettingsRevision{
		ID:                     arg.WorkerSettingsSchemaID + 1000,
		WorkerSettingsSchemaID: arg.WorkerSettingsSchemaID,
		SettingsData:           arg.SettingsData,
	}, nil
}

func (m *mockStorage) UpdateWorkerSettingsRevisionSettings(_ context.Context, arg db.UpdateWorkerSettingsRevisionSettingsParams) (db.WorkerSettingsRevision, error) {
	m.updateRevisionSettingsArg = arg
	return db.WorkerSettingsRevision{ID: arg.ID, SettingsData: arg.SettingsData}, nil
}

func (m *mockStorage) CloneWorkerSettingsRevision(_ context.Context, arg db.CloneWorkerSettingsRevisionParams) (db.WorkerSettingsRevision, error) {
	m.lastCloneWorkerSettingsRevisionArg = arg
	if m.cloneWorkerSettingsRevisionErr != nil {
		return db.WorkerSettingsRevision{}, m.cloneWorkerSettingsRevisionErr
	}
	if m.cloneWorkerSettingsRevisionResp.ID == 0 && arg.ID != 0 {
		return db.WorkerSettingsRevision{ID: arg.ID + 5000, WorkerSettingsSchemaID: 0}, nil
	}
	return m.cloneWorkerSettingsRevisionResp, nil
}

func (m *mockStorage) UpdateWorkflowStepPosition(_ context.Context, _ db.UpdateWorkflowStepPositionParams) error {
	return nil
}

func (m *mockStorage) ListEnrichedStepsByVersionID(_ context.Context, _ int32) ([]db.ListEnrichedStepsByVersionIDRow, error) {
	return m.enrichedSteps, nil
}

func (m *mockStorage) ClearDefaultWorkflowInputSchema(context.Context, db.ClearDefaultWorkflowInputSchemaParams) error {
	return nil
}

func (m *mockStorage) CreateWorkflowInputSchema(_ context.Context, arg db.CreateWorkflowInputSchemaParams) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{
		ID:            1,
		WorkflowID:    arg.WorkflowID,
		Code:          arg.Code,
		VersionNumber: arg.VersionNumber,
		SchemaJson:    arg.SchemaJson,
		Status:        arg.Status,
		IsDefault:     arg.IsDefault,
	}, nil
}

func (m *mockStorage) GetDefaultWorkflowInputSchema(_ context.Context, workflowID int32) (db.WorkflowInputSchema, error) {
	if m.inputSchemas != nil {
		if schema, ok := m.inputSchemas[1]; ok {
			return schema, nil
		}
	}
	return db.WorkflowInputSchema{}, errors.New("no default input schema")
}

func (m *mockStorage) defaultWorkflowInputSchema(workflowID int32) db.WorkflowInputSchema {
	return db.WorkflowInputSchema{
		ID:            1,
		WorkflowID:    workflowID,
		Code:          "public",
		VersionNumber: 1,
		SchemaJson:    []byte(`{"type":"object","properties":{}}`),
		Status:        "active",
		IsDefault:     true,
	}
}

func (m *mockStorage) GetNextWorkflowInputSchemaVersionNumber(context.Context, int32) (int32, error) {
	return 1, nil
}

func (m *mockStorage) GetWorkflowInputSchemaByCode(_ context.Context, arg db.GetWorkflowInputSchemaByCodeParams) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{
		ID:            1,
		WorkflowID:    arg.WorkflowID,
		Code:          arg.Code,
		VersionNumber: 1,
		SchemaJson:    []byte(`{"type":"object","properties":{}}`),
		Status:        "active",
	}, nil
}

func (m *mockStorage) GetWorkflowInputSchemaByID(_ context.Context, id int32) (db.WorkflowInputSchema, error) {
	if m.inputSchemas != nil {
		if schema, ok := m.inputSchemas[id]; ok {
			return schema, nil
		}
	}
	return db.WorkflowInputSchema{
		ID:            id,
		WorkflowID:    m.version.WorkflowID,
		Code:          "public",
		VersionNumber: 1,
		SchemaJson:    []byte(`{"type":"object","properties":{}}`),
		Status:        "active",
	}, nil
}

func (m *mockStorage) HasInputSchemaUsage(context.Context, pgtype.Int4) (bool, error) {
	return false, nil
}

func (m *mockStorage) ListWorkflowInputSchemasByWorkflowID(_ context.Context, workflowID int32) ([]db.WorkflowInputSchema, error) {
	return []db.WorkflowInputSchema{{
		ID:            1,
		WorkflowID:    workflowID,
		Code:          "public",
		VersionNumber: 1,
		SchemaJson:    []byte(`{"type":"object","properties":{}}`),
		Status:        "active",
	}}, nil
}

func (m *mockStorage) SetDefaultWorkflowInputSchema(_ context.Context, arg db.SetDefaultWorkflowInputSchemaParams) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{ID: arg.ID, Status: "active", IsDefault: true}, nil
}

func (m *mockStorage) SoftDeleteWorkflowInputSchema(context.Context, int32) error {
	return nil
}

func (m *mockStorage) UpdateWorkflowInputSchemaRecord(_ context.Context, arg db.UpdateWorkflowInputSchemaRecordParams) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{
		ID:            arg.ID,
		Code:          arg.Code,
		VersionNumber: arg.VersionNumber,
		SchemaJson:    arg.SchemaJson,
		Status:        "active",
	}, nil
}

func (m *mockStorage) UpdateWorkflowInputSchemaStatus(_ context.Context, arg db.UpdateWorkflowInputSchemaStatusParams) (db.WorkflowInputSchema, error) {
	m.lastUpdateInputSchemaStatusArg = arg
	return db.WorkflowInputSchema{ID: arg.ID, Status: arg.Status}, nil
}

func (m *mockStorage) GetNativeInputSchemaForWorkflowVersion(_ context.Context, workflowVersionID int32) (db.WorkflowInputSchema, error) {
	compatibilities := m.compatibilities
	if compatibilities == nil {
		compatibilities = []db.WorkflowVersionInputSchemaCompatibility{{
			ID:                    1,
			WorkflowVersionID:     workflowVersionID,
			WorkflowInputSchemaID: 1,
			CompatibilityType:     "native",
			IsActive:              true,
		}}
	}
	for _, compatibility := range compatibilities {
		if compatibility.WorkflowVersionID != workflowVersionID || compatibility.CompatibilityType != "native" || !compatibility.IsActive {
			continue
		}
		if m.inputSchemas != nil {
			if schema, ok := m.inputSchemas[compatibility.WorkflowInputSchemaID]; ok {
				return schema, nil
			}
		}
		return db.WorkflowInputSchema{
			ID:            compatibility.WorkflowInputSchemaID,
			WorkflowID:    m.version.WorkflowID,
			Code:          "v1",
			VersionNumber: 1,
			SchemaJson:    []byte(`{"type":"object","properties":{}}`),
			Status:        "active",
		}, nil
	}
	return db.WorkflowInputSchema{}, errors.New("native input schema not found")
}

func (m *mockStorage) CreateWorkflowVersionInputSchemaCompatibility(_ context.Context, arg db.CreateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{
		ID:                    1,
		WorkflowVersionID:     arg.WorkflowVersionID,
		WorkflowInputSchemaID: arg.WorkflowInputSchemaID,
		CompatibilityType:     arg.CompatibilityType,
		IsActive:              arg.IsActive,
		IsDefaultRoute:        arg.IsDefaultRoute,
	}, nil
}

func (m *mockStorage) DeactivateWorkflowVersionInputSchemaCompatibility(_ context.Context, arg db.DeactivateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{ID: arg.ID, IsActive: false}, nil
}

func (m *mockStorage) GetWorkflowVersionInputSchemaCompatibilityByID(_ context.Context, id int32) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{ID: id, WorkflowVersionID: m.version.ID, WorkflowInputSchemaID: 1, CompatibilityType: "native", IsActive: true}, nil
}

func (m *mockStorage) HasCompatibilityHistoricalRun(context.Context, int32) (bool, error) {
	return false, nil
}

func (m *mockStorage) ListCompatibilitiesByInputSchemaID(context.Context, int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	return nil, nil
}

func (m *mockStorage) ListCompatibilitiesByVersionID(_ context.Context, versionID int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	if m.compatibilities != nil {
		return m.compatibilities, nil
	}
	return []db.WorkflowVersionInputSchemaCompatibility{{
		ID:                    1,
		WorkflowVersionID:     versionID,
		WorkflowInputSchemaID: 1,
		CompatibilityType:     "native",
		IsActive:              true,
	}}, nil
}

func (m *mockStorage) ListCompatibilitiesByWorkflowID(context.Context, int32) ([]db.WorkflowVersionInputSchemaCompatibility, error) {
	return m.compatibilities, nil
}

func (m *mockStorage) SoftDeleteWorkflowVersionInputSchemaCompatibility(context.Context, int32) error {
	return nil
}

func (m *mockStorage) UpdateCompatibilityDefaultRoute(_ context.Context, arg db.UpdateCompatibilityDefaultRouteParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{ID: arg.ID, IsDefaultRoute: arg.IsDefaultRoute}, nil
}

func (m *mockStorage) UpdateWorkflowVersionInputSchemaCompatibility(_ context.Context, arg db.UpdateWorkflowVersionInputSchemaCompatibilityParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{
		ID:                arg.ID,
		CompatibilityType: arg.CompatibilityType,
		IsActive:          arg.IsActive,
		IsDefaultRoute:    arg.IsDefaultRoute,
	}, nil
}

func makeDBStartStep(id, versionID int32) db.WorkflowStep {
	return db.WorkflowStep{
		ID:                id,
		WorkflowVersionID: versionID,
		StepType:          "control",
		ControlKind:       pgtype.Text{String: "start", Valid: true},
	}
}

func makeDBTaskStep(id, versionID, workTypeID, revision int32) db.WorkflowStep {
	return db.WorkflowStep{
		ID:                       id,
		WorkflowVersionID:        versionID,
		StepType:                 "task",
		WorkTypeID:               pgtype.Int4{Int32: workTypeID, Valid: true},
		WorkerSettingsRevisionID: pgtype.Int4{Int32: revision, Valid: true},
	}
}

func makeDBControlStep(id, versionID int32, controlKind string) db.WorkflowStep {
	return db.WorkflowStep{
		ID:                id,
		WorkflowVersionID: versionID,
		StepType:          "control",
		ControlKind:       pgtype.Text{String: controlKind, Valid: true},
	}
}

func makeDBDep(stepID, dependsOnStepID int32, outcome string) db.WorkflowStepDependency {
	return db.WorkflowStepDependency{
		StepID:          stepID,
		DependsOnStepID: dependsOnStepID,
		Outcome:         pgtype.Text{String: outcome, Valid: true},
	}
}

func ptrInt32(v int32) *int32 {
	return &v
}

func collectWorkflowErrorTypes(errs []dagpkg.ValidationError) []string {
	types := make([]string, 0, len(errs))
	for _, err := range errs {
		types = append(types, err.Type)
	}
	return types
}

// TestActivateVersion_GuardInvalid — нельзя активировать версию с is_valid=false.
func TestActivateVersion_GuardInvalid(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         1,
			WorkflowID: 10,
			IsValid:    false,
		},
	}
	svc := workflow.NewService(store)
	err := svc.ActivateVersion(context.Background(), 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, workflow.ErrValidationRequired), "должна быть ошибка ErrValidationRequired")
}

// TestActivateVersion_Success — активация версии с is_valid=true проходит успешно.
func TestActivateVersion_Success(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         2,
			WorkflowID: 20,
			IsValid:    true,
		},
		activeVersions: []db.WorkflowVersion{
			{ID: 2, WorkflowID: 20, IsValid: true, IsActive: true},
		},
		steps: []db.WorkflowStep{},
	}
	svc := workflow.NewService(store)
	err := svc.ActivateVersion(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int32(2), store.lastUpdateActiveArg.ID)
	assert.True(t, store.lastUpdateActiveArg.IsActive)
}

func TestActivateVersion_ActivatesDraftNativeInputSchema(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:            2,
			WorkflowID:    20,
			VersionNumber: 1,
			IsValid:       true,
		},
		compatibilities: []db.WorkflowVersionInputSchemaCompatibility{{
			ID:                    10,
			WorkflowVersionID:     2,
			WorkflowInputSchemaID: 30,
			CompatibilityType:     "native",
			IsActive:              true,
		}},
		inputSchemas: map[int32]db.WorkflowInputSchema{
			30: {
				ID:            30,
				WorkflowID:    20,
				Code:          "v1",
				VersionNumber: 1,
				SchemaJson:    []byte(`{"type":"object","properties":{}}`),
				Status:        "draft",
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.ActivateVersion(context.Background(), 2)

	require.NoError(t, err)
	assert.Equal(t, int32(30), store.lastUpdateInputSchemaStatusArg.ID)
	assert.Equal(t, "active", store.lastUpdateInputSchemaStatusArg.Status)
	assert.Equal(t, int32(2), store.lastUpdateActiveArg.ID)
	assert.True(t, store.lastUpdateActiveArg.IsActive)
}

func TestListWorkflows_ReturnsVersionCount(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 10, IsActive: false},
			{ID: 2, WorkflowID: 10, IsActive: true},
			{ID: 3, WorkflowID: 10, DeletedAt: pgtype.Timestamp{Valid: true}},
		},
	}
	svc := workflow.NewService(store)

	workflows, err := svc.ListWorkflows(context.Background(), 3)

	require.NoError(t, err)
	require.Len(t, workflows, 1)
	assert.Equal(t, int32(2), workflows[0].VersionCount)
}

// TestValidateVersion_CycleReturnsErrors — ValidateVersion обнаруживает цикл в DAG.
func TestValidateVersion_CycleReturnsErrors(t *testing.T) {
	store := &mockStorage{
		steps: []db.WorkflowStep{
			{ID: 1, WorkflowVersionID: 10, StepType: "control", ControlKind: pgtype.Text{String: "start", Valid: true}},
			{ID: 2, WorkflowVersionID: 10, StepType: "task"},
		},
		deps: []db.WorkflowStepDependency{
			{StepID: 2, DependsOnStepID: 1, Outcome: pgtype.Text{String: "success", Valid: true}},
			{StepID: 1, DependsOnStepID: 2, Outcome: pgtype.Text{String: "success", Valid: true}}, // цикл
		},
	}
	svc := workflow.NewService(store)
	resp, err := svc.ValidateVersion(context.Background(), 10)
	require.NoError(t, err)
	assert.False(t, resp.IsValid)
	assert.NotEmpty(t, resp.Errors)
	assert.False(t, store.lastUpdateValidArg.IsValid)
}

// TestValidateVersion_ValidDAGSetsIsValid — ValidateVersion помечает is_valid=true для корректного DAG.
func TestValidateVersion_ValidDAGSetsIsValid(t *testing.T) {
	store := &mockStorage{
		steps: []db.WorkflowStep{
			{ID: 1, WorkflowVersionID: 20, StepType: "control", ControlKind: pgtype.Text{String: "start", Valid: true}},
			{ID: 2, WorkflowVersionID: 20, StepType: "task"},
		},
		deps: []db.WorkflowStepDependency{
			{StepID: 2, DependsOnStepID: 1, Outcome: pgtype.Text{String: "success", Valid: true}},
		},
	}
	svc := workflow.NewService(store)
	resp, err := svc.ValidateVersion(context.Background(), 20)
	require.NoError(t, err)
	assert.True(t, resp.IsValid)
	assert.Empty(t, resp.Errors)
	assert.True(t, store.lastUpdateValidArg.IsValid)
}

func TestValidateVersion_AllowsDraftNativeInputSchemaForSameVersion(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:            20,
			WorkflowID:    100,
			VersionNumber: 2,
		},
		compatibilities: []db.WorkflowVersionInputSchemaCompatibility{{
			ID:                    10,
			WorkflowVersionID:     20,
			WorkflowInputSchemaID: 30,
			CompatibilityType:     "native",
			IsActive:              true,
		}},
		inputSchemas: map[int32]db.WorkflowInputSchema{
			30: {
				ID:            30,
				WorkflowID:    100,
				Code:          "v2",
				VersionNumber: 2,
				SchemaJson:    []byte(`{"type":"object","properties":{}}`),
				Status:        "draft",
			},
		},
		steps: []db.WorkflowStep{
			makeDBStartStep(1, 20),
			makeDBTaskStep(2, 20, 7, 70),
		},
		deps: []db.WorkflowStepDependency{
			makeDBDep(2, 1, "success"),
		},
	}
	svc := workflow.NewService(store)

	resp, err := svc.ValidateVersion(context.Background(), 20)

	require.NoError(t, err)
	assert.True(t, resp.IsValid)
	assert.Empty(t, resp.Errors)
	assert.True(t, store.lastUpdateValidArg.IsValid)
}

func TestListEnrichedStepsIncludesSavedTaskConfig(t *testing.T) {
	store := &mockStorage{
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:                       12,
				WorkflowVersionID:        99,
				StepType:                 "task",
				WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
				Config:                   []byte(`{"host":"smtp.local","port":25}`),
				SettingsSchema:           []byte(`{"type":"object","properties":{"host":{"type":"string"}}}`),
				InputSchema:              []byte(`{"type":"object","properties":{}}`),
			},
		},
	}
	svc := workflow.NewService(store)

	steps, err := svc.ListEnrichedSteps(context.Background(), 99)

	require.NoError(t, err)
	require.Len(t, steps, 1)
	assert.Equal(t, int32(12), steps[0].ID)
	assert.JSONEq(t, `{"host":"smtp.local","port":25}`, string(steps[0].Config))
}

func TestValidateVersion_InvalidPersistedControlSettingsReturnsError(t *testing.T) {
	store := &mockStorage{
		steps: []db.WorkflowStep{
			{ID: 1, WorkflowVersionID: 20, StepType: "control", ControlKind: pgtype.Text{String: "start", Valid: true}},
			{
				ID:                2,
				WorkflowVersionID: 20,
				StepType:          "control",
				ControlKind:       pgtype.Text{String: "delay", Valid: true},
				ControlSettings:   []byte(`{"count":0,"unit":"sec"}`),
			},
			{ID: 3, WorkflowVersionID: 20, StepType: "task"},
		},
		deps: []db.WorkflowStepDependency{
			{StepID: 2, DependsOnStepID: 1, Outcome: pgtype.Text{String: "success", Valid: true}},
			{StepID: 3, DependsOnStepID: 2, Outcome: pgtype.Text{String: "success", Valid: true}},
		},
	}
	svc := workflow.NewService(store)

	resp, err := svc.ValidateVersion(context.Background(), 20)

	require.NoError(t, err)
	assert.False(t, resp.IsValid)
	assert.Contains(t, collectWorkflowErrorTypes(resp.Errors), "invalid_control_settings")
	assert.False(t, store.lastUpdateValidArg.IsValid)
}

func TestValidateVersionRejectsTaskWithMissingRequiredSettings(t *testing.T) {
	store := &mockStorage{
		version:  db.WorkflowVersion{ID: 20, WorkflowID: 100},
		workflow: db.Workflow{ID: 100},
		steps: []db.WorkflowStep{
			makeDBStartStep(1, 20),
			makeDBTaskStep(2, 20, 7, 70),
		},
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:             2,
				StepType:       "task",
				SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
				Config:         []byte(`{}`),
			},
		},
		deps: []db.WorkflowStepDependency{
			makeDBDep(2, 1, "success"),
		},
	}
	svc := workflow.NewService(store)

	resp, err := svc.ValidateVersion(context.Background(), 20)

	require.NoError(t, err)
	assert.False(t, resp.IsValid)
	assert.Contains(t, collectWorkflowErrorTypes(resp.Errors), "invalid_settings")
}

func TestListVersionSummaries_ActiveFirstThenNewest(t *testing.T) {
	store := &mockStorage{
		listVersionSummaries: []db.ListWorkflowVersionSummariesByWorkflowIDRow{
			{
				ID:            1,
				WorkflowID:    100,
				VersionNumber: 1,
				IsActive:      false,
				IsValid:       true,
				RunCount:      2,
			},
			{
				ID:            2,
				WorkflowID:    100,
				VersionNumber: 2,
				IsActive:      true,
				IsValid:       true,
				RunCount:      10,
			},
			{
				ID:            3,
				WorkflowID:    100,
				VersionNumber: 3,
				IsActive:      true,
				IsValid:       false,
				RunCount:      0,
			},
		},
	}
	svc := workflow.NewService(store)
	got, err := svc.ListVersionSummaries(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, int32(3), got[0].ID)
	assert.Equal(t, int32(2), got[1].ID)
	assert.Equal(t, int32(1), got[2].ID)
	assert.Equal(t, int64(10), got[1].RunCount)
}

func TestCopyVersion_ClonesStepsDependenciesAndTaskRevisions(t *testing.T) {
	sourceVersion := db.WorkflowVersion{
		ID:         10,
		WorkflowID: 100,
		Name:       pgtype.Text{String: "Source", Valid: true},
	}
	store := &mockStorage{
		version:                         sourceVersion,
		maxVersionNumber:                10,
		steps:                           []db.WorkflowStep{makeDBStartStep(1, 10), makeDBTaskStep(2, 10, 7, 70), makeDBControlStep(3, 10, "condition")},
		deps:                            []db.WorkflowStepDependency{makeDBDep(2, 1, "success"), makeDBDep(3, 2, "success")},
		createStepNextID:                2000,
		cloneWorkerSettingsRevisionResp: db.WorkerSettingsRevision{ID: 701},
		createdWorkflowVersion: db.WorkflowVersion{
			ID:            11,
			WorkflowID:    100,
			VersionNumber: 11,
			Name:          pgtype.Text{String: "Source copy", Valid: true},
			IsValid:       false,
			IsActive:      false,
		},
	}
	svc := workflow.NewService(store)

	copied, err := svc.CopyVersion(context.Background(), 10, 55)
	require.NoError(t, err)
	assert.Equal(t, int32(11), copied.ID)
	assert.Len(t, store.createdWorkflowSteps, 3)
	assert.Equal(t, int32(2000), store.createdWorkflowSteps[0].ID)
	assert.Equal(t, int32(2001), store.createdWorkflowSteps[1].ID)
	assert.Equal(t, int32(2002), store.createdWorkflowSteps[2].ID)
	assert.NotEqual(t, int32(70), store.createdWorkflowSteps[1].WorkerSettingsRevisionID.Int32)
	assert.Equal(t, int32(2000), store.createDependencyParams[0].DependsOnStepID)
	assert.Equal(t, int32(2001), store.createDependencyParams[0].StepID)
	assert.Equal(t, int32(2001), store.createDependencyParams[1].DependsOnStepID)
	assert.Equal(t, int32(2002), store.createDependencyParams[1].StepID)
	assert.Equal(t, int32(70), store.lastCloneWorkerSettingsRevisionArg.ID)
	assert.Equal(t, int32(55), store.lastCloneWorkerSettingsRevisionArg.CreatedByUserID.Int32)
	assert.True(t, store.lastCloneWorkerSettingsRevisionArg.CreatedByUserID.Valid)
}

func TestCopyVersion_AssignsUniqueCopyNameWithinWorkflow(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:            10,
			WorkflowID:    100,
			VersionNumber: 1,
			Name:          pgtype.Text{String: "Source", Valid: true},
		},
		maxVersionNumber: 3,
		listVersions: []db.WorkflowVersion{
			{ID: 10, WorkflowID: 100, Name: pgtype.Text{String: "Source", Valid: true}},
			{ID: 11, WorkflowID: 100, Name: pgtype.Text{String: "Source copy", Valid: true}},
			{ID: 12, WorkflowID: 100, Name: pgtype.Text{String: "Source copy 2", Valid: true}},
			{
				ID:         13,
				WorkflowID: 100,
				Name:       pgtype.Text{String: "Source copy 3", Valid: true},
				DeletedAt:  pgtype.Timestamp{Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	copied, err := svc.CopyVersion(context.Background(), 10, 55)

	require.NoError(t, err)
	assert.Equal(t, "Source copy 3", store.lastCreateWorkflowVersionArg.Name.String)
	assert.Equal(t, "Source copy 3", copied.Name.String)
}

func TestUpdateVersionName_PassesNameToStorage(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	_, err := svc.UpdateVersionName(context.Background(), 10, workflow.UpdateVersionNameRequest{Name: "Release 42"})
	require.NoError(t, err)
	assert.Equal(t, int32(10), store.lastUpdateVersionNameArg.ID)
	assert.Equal(t, "Release 42", store.lastUpdateVersionNameArg.Name.String)
}

func TestUpdateVersionName_RejectsDuplicateNameWithinWorkflow(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         10,
			WorkflowID: 100,
			Name:       pgtype.Text{String: "Current", Valid: true},
		},
		listVersions: []db.WorkflowVersion{
			{ID: 10, WorkflowID: 100, Name: pgtype.Text{String: "Current", Valid: true}},
			{ID: 11, WorkflowID: 100, Name: pgtype.Text{String: "Release 42", Valid: true}},
			{
				ID:         12,
				WorkflowID: 100,
				Name:       pgtype.Text{String: "Release 42", Valid: true},
				DeletedAt:  pgtype.Timestamp{Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	_, err := svc.UpdateVersionName(context.Background(), 10, workflow.UpdateVersionNameRequest{Name: " Release 42 "})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "VERSION_NAME_DUPLICATE")
	assert.Zero(t, store.lastUpdateVersionNameArg.ID)
}

func TestUpdateVersionNameHandler_ReturnsRussianDuplicateMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         10,
			WorkflowID: 100,
			Name:       pgtype.Text{String: "Current", Valid: true},
		},
		listVersions: []db.WorkflowVersion{
			{ID: 10, WorkflowID: 100, Name: pgtype.Text{String: "Current", Valid: true}},
			{ID: 11, WorkflowID: 100, Name: pgtype.Text{String: "Release 42", Valid: true}},
		},
	}
	handler := workflow.NewHandler(workflow.NewService(store), nil)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/versions/10/name", bytes.NewBufferString(`{"name":"Release 42"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "versionId", Value: "10"}}

	handler.UpdateVersionName(c)

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "VERSION_NAME_DUPLICATE")
	assert.Contains(t, recorder.Body.String(), "Такое имя версии уже существует")
}

func TestCreateTaskStep_UsesSelectedSchemaAndCreatesPrivateRevision(t *testing.T) {
	store := &mockStorage{
		revisionsBySchema: map[int32][]db.WorkerSettingsRevision{
			22: {
				{
					ID:                     70,
					WorkerSettingsSchemaID: 22,
					SettingsData:           []byte(`{"host":"smtp.example.com"}`),
				},
			},
		},
	}
	svc := workflow.NewService(store)

	step, err := svc.CreateStep(context.Background(), 10, workflow.CreateStepRequest{
		StepType:               "task",
		WorkTypeID:             ptrInt32(7),
		WorkerSettingsSchemaID: ptrInt32(22),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(7), step.WorkTypeID.Int32)
	assert.NotEqual(t, int32(70), step.WorkerSettingsRevisionID.Int32)
	assert.JSONEq(t, `{"host":"smtp.example.com"}`, string(store.createdRevisionSettings))
}

func TestCreateStep_AssignsNextUniqueNameWithinVersion(t *testing.T) {
	store := &mockStorage{
		steps: []db.WorkflowStep{
			{ID: 1, WorkflowVersionID: 10, Name: "Название"},
			{ID: 2, WorkflowVersionID: 10, Name: "Название 2"},
			{
				ID:                3,
				WorkflowVersionID: 10,
				Name:              "Название 3",
				DeletedAt:         pgtype.Timestamp{Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	step, err := svc.CreateStep(context.Background(), 10, workflow.CreateStepRequest{
		Name:        " Название ",
		StepType:    "control",
		ControlKind: ptrString("delay"),
	})

	require.NoError(t, err)
	require.Len(t, store.createWorkflowStepParams, 1)
	assert.Equal(t, "Название 3", store.createWorkflowStepParams[0].Name)
	assert.Equal(t, "Название 3", step.Name)
}

func TestUpdateStep_RejectsDuplicateNameWithinVersion(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{
			ID:                12,
			WorkflowVersionID: 10,
			Name:              "Current",
			StepType:          "control",
			ControlKind:       pgtype.Text{String: "delay", Valid: true},
			ControlSettings:   []byte(`{"count":1,"unit":"sec"}`),
		},
		steps: []db.WorkflowStep{
			{ID: 11, WorkflowVersionID: 10, Name: "Existing"},
			{
				ID:                13,
				WorkflowVersionID: 10,
				Name:              "Existing",
				DeletedAt:         pgtype.Timestamp{Valid: true},
			},
			{ID: 12, WorkflowVersionID: 10, Name: "Current"},
		},
	}
	svc := workflow.NewService(store)

	_, err := svc.UpdateStep(context.Background(), 12, workflow.UpdateStepRequest{Name: ptrString(" Existing ")})

	assert.ErrorIs(t, err, workflow.ErrStepNameDuplicate)
	assert.Zero(t, store.lastUpdateStepArg.ID)
}

func TestCreateDelayStep_RequiresValidControlSettings(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	_, err := svc.CreateStep(context.Background(), 10, workflow.CreateStepRequest{
		StepType:        "control",
		ControlKind:     ptrString("delay"),
		ControlSettings: json.RawMessage(`{"count":"3","unit":"sec"}`),
	})

	assert.ErrorIs(t, err, workflow.ErrControlSettingsInvalid)
}

func TestCreateDelayStep_UsesDefaultControlSettingsWhenOmitted(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	step, err := svc.CreateStep(context.Background(), 10, workflow.CreateStepRequest{
		StepType:    "control",
		ControlKind: ptrString("delay"),
	})

	require.NoError(t, err)
	assert.JSONEq(t, `{"count":1,"unit":"sec"}`, string(step.ControlSettings))
	assert.JSONEq(t, `{"count":1,"unit":"sec"}`, string(store.createWorkflowStepParams[0].ControlSettings))
}

func TestCreateSwitchStep_UsesObjectCaseDefaultSettingsWhenOmitted(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	step, err := svc.CreateStep(context.Background(), 10, workflow.CreateStepRequest{
		StepType:    "control",
		ControlKind: ptrString("switch"),
	})

	require.NoError(t, err)
	assert.JSONEq(t, `{"expression":"$.message.value","cases":[]}`, string(step.ControlSettings))
	assert.JSONEq(t, `{"expression":"$.message.value","cases":[]}`, string(store.createWorkflowStepParams[0].ControlSettings))
}

func TestValidateVersion_AcceptsSwitchObjectCaseSettingsAndOutcomes(t *testing.T) {
	store := &mockStorage{
		version:  db.WorkflowVersion{ID: 20, WorkflowID: 100},
		workflow: db.Workflow{ID: 100},
		steps: []db.WorkflowStep{
			makeDBStartStep(1, 20),
			{
				ID:                2,
				WorkflowVersionID: 20,
				StepType:          "control",
				ControlKind:       pgtype.Text{String: "switch", Valid: true},
				ControlSettings:   []byte(`{"expression":"$.message.value.type","cases":[{"id":"case-vip","label":"VIP","value":"vip"}]}`),
			},
			makeDBTaskStep(3, 20, 7, 70),
			makeDBTaskStep(4, 20, 7, 71),
		},
		deps: []db.WorkflowStepDependency{
			makeDBDep(2, 1, "success"),
			makeDBDep(3, 2, "case-vip"),
			makeDBDep(4, 2, "default"),
		},
	}
	svc := workflow.NewService(store)

	resp, err := svc.ValidateVersion(context.Background(), 20)

	require.NoError(t, err)
	assert.True(t, resp.IsValid)
	assert.Empty(t, resp.Errors)
}

func TestValidateVersion_RejectsInvalidSwitchCaseSettings(t *testing.T) {
	store := &mockStorage{
		version:  db.WorkflowVersion{ID: 20, WorkflowID: 100},
		workflow: db.Workflow{ID: 100},
		steps: []db.WorkflowStep{
			makeDBStartStep(1, 20),
			{
				ID:                2,
				WorkflowVersionID: 20,
				StepType:          "control",
				ControlKind:       pgtype.Text{String: "switch", Valid: true},
				ControlSettings:   []byte(`{"expression":"$.message.value.type","cases":[{"id":"default","label":"VIP","value":"vip"}]}`),
			},
			makeDBTaskStep(3, 20, 7, 70),
		},
		deps: []db.WorkflowStepDependency{
			makeDBDep(2, 1, "success"),
			makeDBDep(3, 2, "default"),
		},
	}
	svc := workflow.NewService(store)

	resp, err := svc.ValidateVersion(context.Background(), 20)

	require.NoError(t, err)
	assert.False(t, resp.IsValid)
	assert.Contains(t, collectWorkflowErrorTypes(resp.Errors), "invalid_control_settings")
}

func ptrString(v string) *string {
	return &v
}

func TestSemanticStepMutation_InvalidatesVersion(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{ID: 12, WorkflowVersionID: 99},
	}
	svc := workflow.NewService(store)
	stepType := "task"

	_, err := svc.UpdateStep(context.Background(), 12, workflow.UpdateStepRequest{StepType: &stepType})
	require.NoError(t, err)
	assert.Equal(t, int32(99), store.lastUpdateValidArg.ID)
}

func TestUpdateStep_PartialPayloadPreservesExistingFields(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{
			ID:                       12,
			WorkflowVersionID:        99,
			StepType:                 "control",
			WorkTypeID:               pgtype.Int4{Int32: 7, Valid: true},
			WorkerSettingsRevisionID: pgtype.Int4{Int32: 99, Valid: true},
			ControlKind:              pgtype.Text{String: "condition", Valid: true},
			ControlSettings:          []byte(`{"count":10,"unit":"sec"}`),
			InputMapping:             []byte(`[{"target":"a","source":"b"}]`),
			CanvasPosition:           []byte(`{"x":1,"y":2}`),
		},
	}
	svc := workflow.NewService(store)

	updated, err := svc.UpdateStep(context.Background(), 12, workflow.UpdateStepRequest{
		ControlSettings: json.RawMessage(`{"count":20,"unit":"sec"}`),
	})
	require.NoError(t, err)

	assert.Equal(t, "control", store.lastUpdateStepArg.StepType)
	assert.Equal(t, int32(7), store.lastUpdateStepArg.WorkTypeID.Int32)
	assert.True(t, store.lastUpdateStepArg.WorkTypeID.Valid)
	assert.Equal(t, int32(99), store.lastUpdateStepArg.WorkerSettingsRevisionID.Int32)
	assert.True(t, store.lastUpdateStepArg.WorkerSettingsRevisionID.Valid)
	assert.Equal(t, "condition", store.lastUpdateStepArg.ControlKind.String)
	assert.True(t, store.lastUpdateStepArg.ControlKind.Valid)
	assert.JSONEq(t, `{"count":20,"unit":"sec"}`, string(store.lastUpdateStepArg.ControlSettings))
	assert.JSONEq(t, `[{"target":"a","source":"b"}]`, string(store.lastUpdateStepArg.InputMapping))
	assert.JSONEq(t, `{"x":1,"y":2}`, string(store.lastUpdateStepArg.CanvasPosition))

	assert.Equal(t, "control", updated.StepType)
	assert.Equal(t, int32(99), updated.WorkflowVersionID)
	assert.JSONEq(t, `{"count":20,"unit":"sec"}`, string(updated.ControlSettings))
	assert.Equal(t, int32(7), updated.WorkTypeID.Int32)
	assert.True(t, updated.WorkTypeID.Valid)
	assert.True(t, updated.WorkerSettingsRevisionID.Valid)
	assert.Equal(t, int32(99), updated.WorkerSettingsRevisionID.Int32)
}

func TestPositionMutation_DoesNotInvalidateVersion(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	err := svc.UpdateStepPosition(context.Background(), 12, workflow.UpdateStepPositionRequest{
		CanvasPosition: json.RawMessage(`{"x":100,"y":200}`),
	})
	require.NoError(t, err)
	assert.Zero(t, store.lastUpdateValidArg.ID)
}

func TestUpdateTaskSettingsValidatesAndDoesNotUpdateInputMapping(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{
			ID:                       12,
			WorkflowVersionID:        99,
			StepType:                 "task",
			WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
			InputMapping:             []byte(`[{"target":"to","source":"$.message.value.email"}]`),
		},
		settingsSchema: db.WorkerSettingsSchema{
			ID:             7,
			SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		},
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:                     12,
				WorkerSettingsSchemaID: pgtype.Int4{Int32: 7, Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateTaskSettings(context.Background(), 12, workflow.UpdateTaskSettingsRequest{
		SettingsData: json.RawMessage(`{"host":"smtp.local"}`),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(44), store.updateRevisionSettingsArg.ID)
	assert.JSONEq(t, `{"host":"smtp.local"}`, string(store.updateRevisionSettingsArg.SettingsData))
	assert.Zero(t, store.lastUpdateStepArg.ID)
	assert.Equal(t, int32(99), store.lastUpdateValidArg.ID)
}

func TestUpdateTaskSettingsClonesSharedTaskSettingsRevision(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{
			ID:                       12,
			WorkflowVersionID:        99,
			StepType:                 "task",
			WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
			InputMapping:             []byte(`[{"target":"to","source":"$.message.value.email"}]`),
		},
		steps: []db.WorkflowStep{
			{
				ID:                       12,
				WorkflowVersionID:        99,
				StepType:                 "task",
				WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
			},
			{
				ID:                       13,
				WorkflowVersionID:        99,
				StepType:                 "task",
				WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
			},
		},
		settingsSchema: db.WorkerSettingsSchema{
			ID:             7,
			SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		},
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:                     12,
				WorkerSettingsSchemaID: pgtype.Int4{Int32: 7, Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateTaskSettings(context.Background(), 12, workflow.UpdateTaskSettingsRequest{
		SettingsData: json.RawMessage(`{"host":"smtp.local"}`),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(44), store.lastCloneWorkerSettingsRevisionArg.ID)
	assert.Equal(t, int32(5000+44), store.lastUpdateStepArg.WorkerSettingsRevisionID.Int32)
	assert.Equal(t, int32(12), store.lastUpdateStepArg.ID)
	assert.Equal(t, int32(5000+44), store.updateRevisionSettingsArg.ID)
	assert.Equal(t, int32(99), store.lastUpdateValidArg.ID)
}

func TestUpdateTaskSettingsRejectsMissingRequiredSettings(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{
			ID:                       12,
			WorkflowVersionID:        99,
			StepType:                 "task",
			WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
		},
		settingsSchema: db.WorkerSettingsSchema{
			ID:             7,
			SettingsSchema: []byte(`{"type":"object","properties":{"host":{"type":"string","required":true}}}`),
		},
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:                     12,
				WorkerSettingsSchemaID: pgtype.Int4{Int32: 7, Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateTaskSettings(context.Background(), 12, workflow.UpdateTaskSettingsRequest{
		SettingsData: json.RawMessage(`{}`),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, workflow.ErrInvalidSettings)
	assert.Zero(t, store.updateRevisionSettingsArg.ID)
	assert.Zero(t, store.lastUpdateStepArg.ID)
}

func TestUpdateTaskInputMappingDoesNotUpdateSettingsRevision(t *testing.T) {
	store := &mockStorage{
		version:  db.WorkflowVersion{ID: 99, WorkflowID: 100},
		workflow: db.Workflow{ID: 100},
		workflowStep: db.WorkflowStep{
			ID:                       12,
			WorkflowVersionID:        99,
			StepType:                 "task",
			WorkerSettingsRevisionID: pgtype.Int4{Int32: 44, Valid: true},
			InputMapping:             []byte(`[{"target":"old","source":"$.message.value.old"}]`),
		},
		enrichedSteps: []db.ListEnrichedStepsByVersionIDRow{
			{
				ID:           12,
				StepType:     "task",
				InputSchema:  []byte(`{"type":"object","properties":{}}`),
				InputMapping: []byte(`[{"target":"old","source":"$.message.value.old"}]`),
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateTaskInputMapping(context.Background(), 12, workflow.UpdateTaskInputMappingRequest{
		InputMapping: json.RawMessage(`[]`),
	})

	require.NoError(t, err)
	assert.Zero(t, store.updateRevisionSettingsArg.ID)
	assert.Equal(t, int32(12), store.lastUpdateStepArg.ID)
	assert.JSONEq(t, `[]`, string(store.lastUpdateStepArg.InputMapping))
	assert.Equal(t, int32(99), store.lastUpdateValidArg.ID)
}
