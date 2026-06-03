package message

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type detailStore struct{}

func (detailStore) CreateNewMessage(context.Context, db.CreateNewMessageParams) (db.Message, error) {
	return db.Message{}, nil
}

func (detailStore) GetWorkflowByID(context.Context, int32) (db.Workflow, error) {
	return db.Workflow{}, nil
}

func (detailStore) ListWorkflowTrafficCandidatesByWorkflowID(context.Context, int32) ([]db.ListWorkflowTrafficCandidatesByWorkflowIDRow, error) {
	return nil, nil
}

func (detailStore) ListWorkflowStepsByVersionID(context.Context, int32) ([]db.WorkflowStep, error) {
	return nil, nil
}

func (detailStore) ListEnrichedStepsByVersionID(context.Context, int32) ([]db.ListEnrichedStepsByVersionIDRow, error) {
	return []db.ListEnrichedStepsByVersionIDRow{
		{
			ID:             10,
			StepType:       "control",
			ControlKind:    pgtype.Text{String: "start", Valid: true},
			InputMapping:   []byte(`[{"target":"email","source":"$.message.value.email"}]`),
			CanvasPosition: []byte(`{"x":20,"y":40}`),
			InputSchema:    []byte(`{"type":"object"}`),
			OutputSchema:   []byte(`{"type":"object"}`),
		},
		{
			ID:             20,
			StepType:       "task",
			WorkTypeID:     pgtype.Int4{Int32: 5, Valid: true},
			WorkTypeName:   pgtype.Text{String: "SMTP", Valid: true},
			WorkTypeCode:   pgtype.Text{String: "smtp", Valid: true},
			WorkTypeMeta:   []byte(`{"icon":"mail"}`),
			InputMapping:   []byte(`[{"target":"body","source":"$.steps.10.output.body"}]`),
			CanvasPosition: []byte(`{"x":240,"y":40}`),
		},
	}, nil
}

func (detailStore) ListDependenciesByVersionID(context.Context, int32) ([]db.WorkflowStepDependency, error) {
	return []db.WorkflowStepDependency{{
		StepID:          20,
		DependsOnStepID: 10,
		Outcome:         pgtype.Text{String: "success", Valid: true},
		OutputIndex:     0,
	}}, nil
}

func (detailStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (detailStore) CreateWorkflowRun(context.Context, db.CreateWorkflowRunParams) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (detailStore) UpdateWorkflowRunStarted(context.Context, int32) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (detailStore) CheckWorkflowAccess(context.Context, db.CheckWorkflowAccessParams) (bool, error) {
	return false, nil
}

func (detailStore) GetMessageStatusByID(context.Context, int32) (db.GetMessageStatusByIDRow, error) {
	return db.GetMessageStatusByIDRow{}, nil
}

func (detailStore) ListWorkflowRunStepStatusesByRunID(context.Context, int32) ([]db.ListWorkflowRunStepStatusesByRunIDRow, error) {
	return nil, nil
}

func (detailStore) GetMessageDetailByID(context.Context, int32) (db.GetMessageDetailByIDRow, error) {
	now := time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)
	return db.GetMessageDetailByIDRow{
		ID:                    100,
		WorkflowID:            7,
		WorkflowName:          "Welcome",
		MessageValue:          []byte(`{"email":"ada@example.com"}`),
		MessageStatus:         "created",
		CreatedAt:             pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:             pgtype.Timestamp{Time: now, Valid: true},
		WorkflowRunID:         pgtype.Int4{Int32: 91, Valid: true},
		WorkflowVersionID:     pgtype.Int4{Int32: 55, Valid: true},
		WorkflowVersionNumber: pgtype.Int4{Int32: 3, Valid: true},
		WorkflowVersionName:   pgtype.Text{String: "Release A", Valid: true},
		WorkflowStatus:        pgtype.Text{String: "completed", Valid: true},
		RunStartedAt:          pgtype.Timestamp{Time: now, Valid: true},
		RunCompletedAt:        pgtype.Timestamp{Time: now.Add(time.Second), Valid: true},
	}, nil
}

func (detailStore) ListWorkflowRunStepDetailsByRunID(context.Context, int32) ([]db.WorkflowRunStep, error) {
	now := time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)
	return []db.WorkflowRunStep{{
		ID:             400,
		WorkflowRunID:  91,
		WorkflowStepID: 20,
		Status:         "completed",
		Outcome:        pgtype.Text{String: "success", Valid: true},
		InputData:      []byte(`{"email":"ada@example.com"}`),
		OutputData:     []byte(`{"message_id":"abc"}`),
		StartedAt:      pgtype.Timestamp{Time: now, Valid: true},
		CompletedAt:    pgtype.Timestamp{Time: now.Add(time.Second), Valid: true},
	}}, nil
}

func (detailStore) ListMessagesByOrganizationID(context.Context, db.ListMessagesByOrganizationIDParams) ([]db.ListMessagesByOrganizationIDRow, error) {
	return nil, nil
}

func (detailStore) CountMessagesByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 0, nil
}

func TestGetMessageDetailBuildsRunGraphAndStepData(t *testing.T) {
	svc := NewService(detailStore{}, client.Client(nil))

	got, err := svc.GetMessageDetail(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetMessageDetail returned error: %v", err)
	}

	if got.Graph.VersionID != 55 {
		t.Fatalf("version id = %d", got.Graph.VersionID)
	}
	require.NotNil(t, got.WorkflowVersionID)
	require.NotNil(t, got.WorkflowVersionNumber)
	require.NotNil(t, got.WorkflowVersionName)
	require.Equal(t, int32(55), *got.WorkflowVersionID)
	require.Equal(t, int32(3), *got.WorkflowVersionNumber)
	require.Equal(t, "Release A", *got.WorkflowVersionName)
	if len(got.Graph.Steps) != 2 {
		t.Fatalf("steps = %d", len(got.Graph.Steps))
	}
	if len(got.Graph.Dependencies) != 1 {
		t.Fatalf("dependencies = %d", len(got.Graph.Dependencies))
	}
	if got.RunSteps[0].InputData["email"] != "ada@example.com" {
		t.Fatalf("input not mapped")
	}
	if got.RunSteps[0].OutputData["message_id"] != "abc" {
		t.Fatalf("output not mapped")
	}
	require.NotNil(t, got.RunSteps[0].DurationMs)
	require.Equal(t, int64(1000), *got.RunSteps[0].DurationMs)
}

var _ Storage = detailStore{}

type listMessagesStore struct {
	detailStore
}

func (listMessagesStore) CountMessagesByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 1, nil
}

func (listMessagesStore) ListMessagesByOrganizationID(context.Context, db.ListMessagesByOrganizationIDParams) ([]db.ListMessagesByOrganizationIDRow, error) {
	now := time.Date(2026, 5, 18, 8, 0, 0, 0, time.UTC)
	return []db.ListMessagesByOrganizationIDRow{{
		ID:                    200,
		WorkflowID:            7,
		WorkflowName:          "Welcome",
		Status:                "running",
		CreatedAt:             pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:             pgtype.Timestamp{Time: now.Add(time.Minute), Valid: true},
		WorkflowVersionID:     pgtype.Int4{Int32: 55, Valid: true},
		WorkflowVersionNumber: pgtype.Int4{Int32: 3, Valid: true},
		WorkflowVersionName:   pgtype.Text{String: "Release A", Valid: true},
	}}, nil
}

func TestListMessagesIncludesWorkflowVersion(t *testing.T) {
	svc := NewService(listMessagesStore{}, client.Client(nil))

	got, total, err := svc.ListMessages(context.Background(), 12, 1, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].WorkflowVersionID)
	require.NotNil(t, got[0].WorkflowVersionNumber)
	require.NotNil(t, got[0].WorkflowVersionName)
	require.Equal(t, int32(55), *got[0].WorkflowVersionID)
	require.Equal(t, int32(3), *got[0].WorkflowVersionNumber)
	require.Equal(t, "Release A", *got[0].WorkflowVersionName)
}

type sendMessageStore struct {
	workflowRunArg db.CreateWorkflowRunParams
	dagVersionID   int32
}

func (s *sendMessageStore) CreateNewMessage(context.Context, db.CreateNewMessageParams) (db.Message, error) {
	return db.Message{ID: 501}, nil
}

func (s *sendMessageStore) GetWorkflowByID(context.Context, int32) (db.Workflow, error) {
	return db.Workflow{ID: 100}, nil
}

func (s *sendMessageStore) ListWorkflowTrafficCandidatesByWorkflowID(context.Context, int32) ([]db.ListWorkflowTrafficCandidatesByWorkflowIDRow, error) {
	return []db.ListWorkflowTrafficCandidatesByWorkflowIDRow{
		{ID: 10, WorkflowID: 100, IsActive: true, TrafficWeight: 50, RunCount: 1},
		{ID: 20, WorkflowID: 100, IsActive: true, TrafficWeight: 50, RunCount: 0},
	}, nil
}

func (s *sendMessageStore) ListWorkflowStepsByVersionID(context.Context, int32) ([]db.WorkflowStep, error) {
	return nil, nil
}

func (s *sendMessageStore) ListEnrichedStepsByVersionID(_ context.Context, workflowVersionID int32) ([]db.ListEnrichedStepsByVersionIDRow, error) {
	s.dagVersionID = workflowVersionID
	return nil, nil
}

func (s *sendMessageStore) ListDependenciesByVersionID(context.Context, int32) ([]db.WorkflowStepDependency, error) {
	return nil, nil
}

func (s *sendMessageStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (s *sendMessageStore) CreateWorkflowRun(_ context.Context, arg db.CreateWorkflowRunParams) (db.WorkflowRun, error) {
	s.workflowRunArg = arg
	return db.WorkflowRun{ID: 601, Status: temporaltypes.RunStatusRunning}, nil
}

func (s *sendMessageStore) UpdateWorkflowRunStarted(context.Context, int32) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func (s *sendMessageStore) CheckWorkflowAccess(context.Context, db.CheckWorkflowAccessParams) (bool, error) {
	return false, nil
}

func (s *sendMessageStore) GetMessageStatusByID(context.Context, int32) (db.GetMessageStatusByIDRow, error) {
	return db.GetMessageStatusByIDRow{}, nil
}

func (s *sendMessageStore) ListWorkflowRunStepStatusesByRunID(context.Context, int32) ([]db.ListWorkflowRunStepStatusesByRunIDRow, error) {
	return nil, nil
}

func (s *sendMessageStore) GetMessageDetailByID(context.Context, int32) (db.GetMessageDetailByIDRow, error) {
	return db.GetMessageDetailByIDRow{}, nil
}

func (s *sendMessageStore) ListWorkflowRunStepDetailsByRunID(context.Context, int32) ([]db.WorkflowRunStep, error) {
	return nil, nil
}

func (s *sendMessageStore) ListMessagesByOrganizationID(context.Context, db.ListMessagesByOrganizationIDParams) ([]db.ListMessagesByOrganizationIDRow, error) {
	return nil, nil
}

func (s *sendMessageStore) CountMessagesByOrganizationID(context.Context, pgtype.Int4) (int64, error) {
	return 0, nil
}

type executeWorkflowClient struct {
	client.Client
}

func (executeWorkflowClient) ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, ...interface{}) (client.WorkflowRun, error) {
	return nil, nil
}

func TestSelectWorkflowVersionByTrafficDeficitChoosesUnderAllocatedVersion(t *testing.T) {
	version, err := selectWorkflowVersionByTrafficDeficit([]db.ListWorkflowTrafficCandidatesByWorkflowIDRow{
		{ID: 10, WorkflowID: 100, IsActive: true, TrafficWeight: 50, RunCount: 1},
		{ID: 20, WorkflowID: 100, IsActive: true, TrafficWeight: 50, RunCount: 0},
	})

	require.NoError(t, err)
	require.Equal(t, int32(20), version.ID)
}

func TestSendMessageUsesTrafficDeficitToSelectVersion(t *testing.T) {
	store := &sendMessageStore{}
	svc := NewService(store, executeWorkflowClient{})

	_, validationErrors, err := svc.SendMessage(context.Background(), SendMessageRequest{
		WorkflowID: 100,
		Value:      map[string]any{"email": "ada@example.com"},
	}, "")

	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.Equal(t, int32(20), store.workflowRunArg.WorkflowVersionID)
	require.Equal(t, int32(20), store.dagVersionID)
}

var _ Storage = (*sendMessageStore)(nil)
