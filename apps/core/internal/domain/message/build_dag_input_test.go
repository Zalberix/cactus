package message

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type buildDAGInputStore struct{}

func (buildDAGInputStore) CreateNewMessage(context.Context, db.CreateNewMessageParams) (db.Message, error) {
	return db.Message{}, nil
}

func (buildDAGInputStore) GetWorkflowByID(context.Context, int32) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (buildDAGInputStore) ListWorkflowTrafficCandidatesByWorkflowID(context.Context, int32) ([]db.ListWorkflowTrafficCandidatesByWorkflowIDRow, error) {
	return nil, nil
}

func (buildDAGInputStore) ListWorkflowStepsByVersionID(context.Context, int32) ([]db.WorkflowStep, error) {
	return nil, nil
}

func (buildDAGInputStore) ListEnrichedStepsByVersionID(context.Context, int32) ([]db.ListEnrichedStepsByVersionIDRow, error) {
	return []db.ListEnrichedStepsByVersionIDRow{{
		ID:                       51,
		StepType:                 "task",
		OrganizationID:           pgtype.Int4{Int32: 99, Valid: true},
		WorkTypeID:               pgtype.Int4{Int32: 1, Valid: true},
		WorkerSettingsRevisionID: pgtype.Int4{Int32: 25, Valid: true},
		WorkerSettingsSchemaID:   pgtype.Int4{Int32: 7, Valid: true},
	}}, nil
}

func (buildDAGInputStore) ListDependenciesByVersionID(context.Context, int32) ([]db.WorkflowStepDependency, error) {
	return nil, nil
}

func (buildDAGInputStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{ID: 25, WorkerSettingsSchemaID: 7}, nil
}

func (buildDAGInputStore) CreateWorkflowRun(context.Context, db.CreateWorkflowRunParams) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (buildDAGInputStore) UpdateWorkflowRunStarted(context.Context, int32) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (buildDAGInputStore) CheckWorkflowAccess(context.Context, db.CheckWorkflowAccessParams) (bool, error) {
	return false, nil
}

func (buildDAGInputStore) GetMessageStatusByID(context.Context, int32) (db.GetMessageStatusByIDRow, error) {
	return db.GetMessageStatusByIDRow{}, nil
}

func (buildDAGInputStore) ListWorkflowRunStepStatusesByRunID(context.Context, int32) ([]db.ListWorkflowRunStepStatusesByRunIDRow, error) {
	return nil, nil
}

func (buildDAGInputStore) GetMessageDetailByID(context.Context, int32) (db.GetMessageDetailByIDRow, error) {
	return db.GetMessageDetailByIDRow{}, nil
}

func (buildDAGInputStore) ListWorkflowRunStepDetailsByRunID(context.Context, int32) ([]db.WorkflowRunStep, error) {
	return nil, nil
}

func (buildDAGInputStore) ListMessagesByOrganizationID(context.Context, db.ListMessagesByOrganizationIDParams) ([]db.ListMessagesByOrganizationIDRow, error) {
	return nil, nil
}

func (buildDAGInputStore) CountMessagesByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 0, nil
}

func TestBuildDAGInputIncludesWorkerSettingsSchemaID(t *testing.T) {
	svc := NewService(buildDAGInputStore{}, client.Client(nil))

	input, err := svc.buildDAGInput(context.Background(), 8, 20, []byte(`{}`))
	require.NoError(t, err)
	require.Len(t, input.Steps, 1)
	require.Equal(t, int32(7), input.Steps[0].WorkerSettingsSchemaID)
	require.Equal(t, int32(99), input.Steps[0].OrganizationID)
}

var _ Storage = buildDAGInputStore{}
