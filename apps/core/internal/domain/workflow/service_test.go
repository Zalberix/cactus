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
	version          db.WorkflowVersion
	getVersionErr    error
	activateErr      error
	updateValidErr   error
	updateSchemaErr  error
	steps            []db.WorkflowStep
	deps             []db.WorkflowStepDependency
	activeVersions   []db.WorkflowVersion
	maxVersionNumber int32
	// capture args
	lastUpdateValidArg  db.UpdateWorkflowVersionValidParams
	lastUpdateActiveArg db.UpdateWorkflowVersionActiveParams
	lastUpdateSchemaArg db.UpdateWorkflowInputValidationParams
}

func (m *mockStorage) WithTx(_ context.Context, _ func(q *db.Queries) error) error {
	return nil
}

func (m *mockStorage) CreateWorkflow(_ context.Context, _ db.CreateWorkflowParams) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) GetWorkflowByID(_ context.Context, _ int32) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) ListWorkflowsBySystemID(_ context.Context, _ int32) ([]db.Workflow, error) {
	return nil, nil
}

func (m *mockStorage) UpdateWorkflow(_ context.Context, _ db.UpdateWorkflowParams) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (m *mockStorage) UpdateWorkflowInputValidation(_ context.Context, arg db.UpdateWorkflowInputValidationParams) (db.Workflow, error) {
	m.lastUpdateSchemaArg = arg
	return db.Workflow{}, m.updateSchemaErr
}
func (m *mockStorage) SoftDeleteWorkflow(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) CreateWorkflowVersion(_ context.Context, _ db.CreateWorkflowVersionParams) (db.WorkflowVersion, error) {
	return db.WorkflowVersion{}, nil
}

func (m *mockStorage) GetWorkflowVersionByID(_ context.Context, _ int32) (db.WorkflowVersion, error) {
	return m.version, m.getVersionErr
}

func (m *mockStorage) GetMaxVersionNumberByWorkflowID(_ context.Context, _ int32) (int32, error) {
	return m.maxVersionNumber, nil
}

func (m *mockStorage) ListWorkflowVersionsByWorkflowID(_ context.Context, _ int32) ([]db.WorkflowVersion, error) {
	return nil, nil
}

func (m *mockStorage) ListActiveWorkflowVersions(_ context.Context, _ int32) ([]db.WorkflowVersion, error) {
	return m.activeVersions, nil
}

func (m *mockStorage) UpdateWorkflowVersionValid(_ context.Context, arg db.UpdateWorkflowVersionValidParams) (db.WorkflowVersion, error) {
	m.lastUpdateValidArg = arg
	return db.WorkflowVersion{IsValid: arg.IsValid}, m.updateValidErr
}

func (m *mockStorage) UpdateWorkflowVersionActive(_ context.Context, arg db.UpdateWorkflowVersionActiveParams) (db.WorkflowVersion, error) {
	m.lastUpdateActiveArg = arg
	return db.WorkflowVersion{IsActive: arg.IsActive}, m.activateErr
}
func (m *mockStorage) SoftDeleteWorkflowVersion(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) CreateWorkflowStep(_ context.Context, _ db.CreateWorkflowStepParams) (db.WorkflowStep, error) {
	return db.WorkflowStep{}, nil
}

func (m *mockStorage) GetWorkflowStepByID(_ context.Context, _ int32) (db.WorkflowStep, error) {
	return db.WorkflowStep{}, nil
}

func (m *mockStorage) ListWorkflowStepsByVersionID(_ context.Context, _ int32) ([]db.WorkflowStep, error) {
	return m.steps, nil
}

func (m *mockStorage) UpdateWorkflowStep(_ context.Context, _ db.UpdateWorkflowStepParams) (db.WorkflowStep, error) {
	return db.WorkflowStep{}, nil
}
func (m *mockStorage) DeleteWorkflowStepsByVersionID(_ context.Context, _ int32) error { return nil }
func (m *mockStorage) SoftDeleteWorkflowStep(_ context.Context, _ int32) error         { return nil }
func (m *mockStorage) CreateWorkflowStepDependency(_ context.Context, _ db.CreateWorkflowStepDependencyParams) error {
	return nil
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

func (m *mockStorage) ListWorkerSettingsRevisionsBySchemaID(_ context.Context, _ int32) ([]db.WorkerSettingsRevision, error) {
	return nil, nil
}

func (m *mockStorage) UpdateWorkflowStepPosition(_ context.Context, _ db.UpdateWorkflowStepPositionParams) error {
	return nil
}

func (m *mockStorage) ListEnrichedStepsByVersionID(_ context.Context, _ int32) ([]db.ListEnrichedStepsByVersionIDRow, error) {
	return nil, nil
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

// TestRegenerateInputValidation_Schema — проверяем корректность генерации JSON Schema из input_mapping.
func TestRegenerateInputValidation_Schema(t *testing.T) {
	// Шаг с маппингом из $.message.value.email и $.message.value.name
	mapping := []map[string]string{
		{"target": "email", "source": "$.message.value.email"},
		{"target": "name", "source": "$.message.value.name"},
	}
	mappingJSON, _ := json.Marshal(mapping)

	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         3,
			WorkflowID: 30,
			IsValid:    true,
		},
		activeVersions: []db.WorkflowVersion{
			{ID: 3, WorkflowID: 30, IsValid: true, IsActive: true},
		},
		steps: []db.WorkflowStep{
			{
				ID:                3,
				WorkflowVersionID: 3,
				StepType:          "task",
				InputMapping:      mappingJSON,
			},
		},
	}
	svc := workflow.NewService(store)
	err := svc.ActivateVersion(context.Background(), 3)
	require.NoError(t, err)

	// Проверяем, что схема была сохранена
	require.NotNil(t, store.lastUpdateSchemaArg.InputValidation)

	var schema map[string]interface{}
	err = json.Unmarshal(store.lastUpdateSchemaArg.InputValidation, &schema)
	require.NoError(t, err)

	assert.Equal(t, "object", schema["type"])
	props, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "properties должны быть map")
	assert.Contains(t, props, "email")
	assert.Contains(t, props, "name")
}

// TestRegenerateInputValidation_NoActiveVersions — если нет активных версий, input_validation=nil.
func TestRegenerateInputValidation_NoActiveVersions(t *testing.T) {
	store := &mockStorage{
		version: db.WorkflowVersion{
			ID:         4,
			WorkflowID: 40,
			IsValid:    true,
		},
		activeVersions: []db.WorkflowVersion{}, // пустой список после деактивации
	}
	svc := workflow.NewService(store)
	err := svc.DeactivateVersion(context.Background(), 4)
	require.NoError(t, err)
	assert.Nil(t, store.lastUpdateSchemaArg.InputValidation)
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
