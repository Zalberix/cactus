package workflow_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zalberix/cactus/apps/core/internal/domain/workflow"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// mockStorage — минимальный мок Storage для тестов.
type mockStorage struct {
	version           db.WorkflowVersion
	getVersionErr     error
	activateErr       error
	updateValidErr    error
	updateSchemaErr   error
	updateNameErr     error
	updateTrafficErr  error
	steps             []db.WorkflowStep
	enrichedSteps     []db.ListEnrichedStepsByVersionIDRow
	deps              []db.WorkflowStepDependency
	activeVersions    []db.WorkflowVersion
	listVersions      []db.WorkflowVersion
	maxVersionNumber  int32
	workflow          db.Workflow
	revisionsBySchema map[int32][]db.WorkerSettingsRevision

	listVersionSummaries               []db.ListWorkflowVersionSummariesByWorkflowIDRow
	listVersionSummariesErr            error
	createdWorkflowVersion             db.WorkflowVersion
	createWorkflowVersionErr           error
	createdWorkflowSteps               []db.WorkflowStep
	createWorkflowStepParams           []db.CreateWorkflowStepParams
	createStepErr                      error
	createStepNextID                   int32
	createDependencyParams             []db.CreateWorkflowStepDependencyParams
	createDependencyErr                error
	createdRevisionSettings            []byte
	updateTrafficArgs                  []updateTrafficArg
	updateRevisionSettingsArg          db.UpdateWorkerSettingsRevisionSettingsParams
	lastUpdateVersionNameArg           db.UpdateWorkflowVersionNameParams
	lastCloneWorkerSettingsRevisionArg db.CloneWorkerSettingsRevisionParams
	cloneWorkerSettingsRevisionResp    db.WorkerSettingsRevision
	cloneWorkerSettingsRevisionErr     error
	getWorkflowStepErr                 error
	workflowStep                       db.WorkflowStep
	// capture args
	lastUpdateValidArg    db.UpdateWorkflowVersionValidParams
	lastUpdateActiveArg   db.UpdateWorkflowVersionActiveParams
	lastUpdateSchemaArg   db.UpdateWorkflowInputSchemaParams
	invalidatedWorkflowID int32
}

type updateTrafficArg struct {
	ID            int32
	TrafficWeight int32
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
	return nil, nil
}

func (m *mockStorage) UpdateWorkflow(_ context.Context, _ db.UpdateWorkflowParams) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) UpdateWorkflowInputSchema(_ context.Context, arg db.UpdateWorkflowInputSchemaParams) (db.Workflow, error) {
	m.lastUpdateSchemaArg = arg
	return db.Workflow{ID: arg.ID, InputSchema: arg.InputSchema}, m.updateSchemaErr
}
func (m *mockStorage) SoftDeleteWorkflow(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) CreateWorkflowVersion(_ context.Context, arg db.CreateWorkflowVersionParams) (db.WorkflowVersion, error) {
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
			TrafficWeight:  arg.TrafficWeight,
			IsControlGroup: arg.IsControlGroup,
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

func (m *mockStorage) UpdateWorkflowVersionTrafficWeight(_ context.Context, arg db.UpdateWorkflowVersionTrafficWeightParams) (db.WorkflowVersion, error) {
	m.updateTrafficArgs = append(m.updateTrafficArgs, updateTrafficArg{ID: arg.ID, TrafficWeight: arg.TrafficWeight})
	return db.WorkflowVersion{ID: arg.ID, TrafficWeight: arg.TrafficWeight}, m.updateTrafficErr
}

func (m *mockStorage) UpdateWorkflowVersionTrafficWeightIncludingDeleted(_ context.Context, arg db.UpdateWorkflowVersionTrafficWeightIncludingDeletedParams) (db.WorkflowVersion, error) {
	m.updateTrafficArgs = append(m.updateTrafficArgs, updateTrafficArg{ID: arg.ID, TrafficWeight: arg.TrafficWeight})
	return db.WorkflowVersion{ID: arg.ID, TrafficWeight: arg.TrafficWeight}, m.updateTrafficErr
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

func (m *mockStorage) UpdateWorkflowStep(_ context.Context, _ db.UpdateWorkflowStepParams) (db.WorkflowStep, error) {
	return db.WorkflowStep{}, nil
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

func TestGetWorkflowInputSchema_ReturnsStoredSchema(t *testing.T) {
	store := &mockStorage{
		workflow: db.Workflow{ID: 30, InputSchema: []byte(`{"type":"object","properties":{"email":{"type":"string","required":true}}}`)},
	}
	svc := workflow.NewService(store)
	schema, err := svc.GetWorkflowInputSchema(context.Background(), 30)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"object","properties":{"email":{"type":"string","required":true}}}`, string(schema))
}

func TestUpsertWorkflowInputSchemaField_UpdatesSchemaAndInvalidatesVersions(t *testing.T) {
	store := &mockStorage{
		workflow:       db.Workflow{ID: 40, InputSchema: []byte(`{"type":"object","properties":{}}`)},
		activeVersions: []db.WorkflowVersion{}, // пустой список после деактивации
	}
	svc := workflow.NewService(store)
	schema, err := svc.UpsertWorkflowInputSchemaField(context.Background(), 40, workflow.WorkflowInputSchemaFieldRequest{
		Name:     "email",
		Type:     "string",
		Required: true,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"object","properties":{"email":{"type":"string","required":true}}}`, string(schema))
	assert.Equal(t, int32(40), store.lastUpdateSchemaArg.ID)
	assert.JSONEq(t, `{"type":"object","properties":{"email":{"type":"string","required":true}}}`, string(store.lastUpdateSchemaArg.InputSchema))
	assert.Equal(t, int32(40), store.invalidatedWorkflowID)
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
			ID:             11,
			WorkflowID:     100,
			VersionNumber:  11,
			Name:           pgtype.Text{String: "Source copy", Valid: true},
			IsValid:        false,
			IsActive:       false,
			TrafficWeight:  0,
			IsControlGroup: false,
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

func TestUpdateVersionName_PassesNameToStorage(t *testing.T) {
	store := &mockStorage{}
	svc := workflow.NewService(store)

	_, err := svc.UpdateVersionName(context.Background(), 10, workflow.UpdateVersionNameRequest{Name: "Release 42"})
	require.NoError(t, err)
	assert.Equal(t, int32(10), store.lastUpdateVersionNameArg.ID)
	assert.Equal(t, "Release 42", store.lastUpdateVersionNameArg.Name.String)
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

func TestSemanticStepMutation_InvalidatesVersion(t *testing.T) {
	store := &mockStorage{
		workflowStep: db.WorkflowStep{ID: 12, WorkflowVersionID: 99},
	}
	svc := workflow.NewService(store)

	_, err := svc.UpdateStep(context.Background(), 12, workflow.UpdateStepRequest{StepType: "task"})
	require.NoError(t, err)
	assert.Equal(t, int32(99), store.lastUpdateValidArg.ID)
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

func TestUpdateTraffic_EqualModeSplitsActiveVersions(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true},
			{ID: 2, WorkflowID: 100, IsActive: true},
			{ID: 3, WorkflowID: 100, IsActive: true},
			{ID: 4, WorkflowID: 100, IsActive: false},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Mode: "equal",
	})
	require.NoError(t, err)
	require.Len(t, store.updateTrafficArgs, 4)

	got := make(map[int32]int32, 4)
	for _, arg := range store.updateTrafficArgs {
		got[arg.ID] = arg.TrafficWeight
	}
	assert.Equal(t, int32(34), got[1])
	assert.Equal(t, int32(33), got[2])
	assert.Equal(t, int32(33), got[3])
	assert.Equal(t, int32(0), got[4])
}

func TestUpdateTraffic_CustomModeRejectsTotalAbove100(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true},
			{ID: 2, WorkflowID: 100, IsActive: true},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Mode: "custom",
		Weights: []workflow.TrafficWeightInput{
			{VersionID: 1, Weight: 90},
			{VersionID: 2, Weight: 20},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, workflow.ErrTrafficWeightInvalid)
	assert.Empty(t, store.updateTrafficArgs)
}

func TestUpdateTraffic_PerVersionFixedAndShareDistributesRemaining(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true},
			{ID: 2, WorkflowID: 100, IsActive: true},
			{ID: 3, WorkflowID: 100, IsActive: true},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Versions: []workflow.TrafficVersionInput{
			{VersionID: 3, Mode: "share"},
			{VersionID: 2, Mode: "fixed", Weight: 70},
			{VersionID: 1, Mode: "share"},
		},
	})
	require.NoError(t, err)
	require.Len(t, store.updateTrafficArgs, 3)

	got := make(map[int32]int32, 3)
	for _, arg := range store.updateTrafficArgs {
		got[arg.ID] = arg.TrafficWeight
	}
	assert.Equal(t, int32(15), got[1])
	assert.Equal(t, int32(70), got[2])
	assert.Equal(t, int32(15), got[3])
}

func TestUpdateTraffic_PerVersionClearsInactiveVersionTraffic(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true, TrafficWeight: 50},
			{ID: 2, WorkflowID: 100, IsActive: true, TrafficWeight: 50},
			{ID: 3, WorkflowID: 100, IsActive: false, TrafficWeight: 100},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Versions: []workflow.TrafficVersionInput{
			{VersionID: 1, Mode: "fixed", Weight: 70},
			{VersionID: 2, Mode: "share"},
		},
	})
	require.NoError(t, err)
	require.Len(t, store.updateTrafficArgs, 3)

	got := make(map[int32]int32, 3)
	for _, arg := range store.updateTrafficArgs {
		got[arg.ID] = arg.TrafficWeight
	}
	assert.Equal(t, int32(70), got[1])
	assert.Equal(t, int32(30), got[2])
	assert.Equal(t, int32(0), got[3])
}

func TestUpdateTraffic_PerVersionClearsDeletedVersionTraffic(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true, TrafficWeight: 50},
			{ID: 2, WorkflowID: 100, IsActive: true, TrafficWeight: 50},
			{
				ID:            3,
				WorkflowID:    100,
				IsActive:      false,
				TrafficWeight: 100,
				DeletedAt:     pgtype.Timestamp{Valid: true},
			},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Versions: []workflow.TrafficVersionInput{
			{VersionID: 1, Mode: "fixed", Weight: 70},
			{VersionID: 2, Mode: "share"},
		},
	})
	require.NoError(t, err)
	require.Len(t, store.updateTrafficArgs, 3)

	got := make(map[int32]int32, 3)
	for _, arg := range store.updateTrafficArgs {
		got[arg.ID] = arg.TrafficWeight
	}
	assert.Equal(t, int32(70), got[1])
	assert.Equal(t, int32(30), got[2])
	assert.Equal(t, int32(0), got[3])
}

func TestUpdateTraffic_PerVersionRejectsFixedTotalAbove100(t *testing.T) {
	store := &mockStorage{
		listVersions: []db.WorkflowVersion{
			{ID: 1, WorkflowID: 100, IsActive: true},
			{ID: 2, WorkflowID: 100, IsActive: true},
		},
	}
	svc := workflow.NewService(store)

	err := svc.UpdateWorkflowTraffic(context.Background(), 100, workflow.UpdateTrafficRequest{
		Versions: []workflow.TrafficVersionInput{
			{VersionID: 1, Mode: "fixed", Weight: 80},
			{VersionID: 2, Mode: "fixed", Weight: 30},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, workflow.ErrTrafficWeightInvalid)
	assert.Empty(t, store.updateTrafficArgs)
}

func TestDeleteWorkflowInputSchemaField_UpdatesSchemaAndInvalidatesVersions(t *testing.T) {
	store := &mockStorage{
		workflow: db.Workflow{ID: 40, InputSchema: []byte(`{
			"type":"object",
			"properties":{
				"email":{"type":"string","required":true},
				"subject":{"type":"string"}
			}
		}`)},
	}
	svc := workflow.NewService(store)
	schema, err := svc.DeleteWorkflowInputSchemaField(context.Background(), 40, "email")
	require.NoError(t, err)

	assert.JSONEq(t, `{"type":"object","properties":{"subject":{"type":"string"}}}`, string(schema))
	assert.Equal(t, int32(40), store.lastUpdateSchemaArg.ID)
	assert.JSONEq(t, `{"type":"object","properties":{"subject":{"type":"string"}}}`, string(store.lastUpdateSchemaArg.InputSchema))
	assert.Equal(t, int32(40), store.invalidatedWorkflowID)
}
