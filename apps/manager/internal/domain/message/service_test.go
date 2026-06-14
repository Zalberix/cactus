package message

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"

	temporaltypes "github.com/zalberix/cactus/apps/manager/internal/temporal"
	db "github.com/zalberix/cactus/libs/storage/db"
)

type detailStore struct{}

func (detailStore) CreateNewMessage(context.Context, db.CreateNewMessageParams) (db.Message, error) {
	return db.Message{}, nil
}

func (detailStore) GetWorkflowByID(context.Context, int32) (db.Workflow, error) {
	return db.Workflow{}, nil
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
	messageArg       db.CreateNewMessageParams
	workflowRunArg   db.CreateWorkflowRunParams
	dagVersionID     int32
	schemasByVersion map[[2]int32]db.WorkflowInputSchema
	scopes           []db.ListActiveExperimentScopesForRoutingRow
	variants         map[int32][]db.WorkflowExperimentVariant
	pairRoutes       map[[2]int32]db.WorkflowVersionInputSchemaCompatibility
	routingParams    db.ListActiveExperimentScopesForRoutingParams
}

func (s *sendMessageStore) CreateNewMessage(_ context.Context, arg db.CreateNewMessageParams) (db.Message, error) {
	s.messageArg = arg
	return db.Message{ID: 501, WorkflowID: arg.WorkflowID, WorkflowInputSchemaID: arg.WorkflowInputSchemaID}, nil
}

func (s *sendMessageStore) GetWorkflowByID(_ context.Context, id int32) (db.Workflow, error) {
	return db.Workflow{ID: id}, nil
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

func (s *sendMessageStore) GetWorkflowInputSchemaByID(_ context.Context, id int32) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{ID: id, WorkflowID: 100, Code: "public", VersionNumber: 1, SchemaJson: []byte(`{"type":"object","properties":{}}`), Status: "active"}, nil
}

func (s *sendMessageStore) GetWorkflowInputSchemaByCode(_ context.Context, arg db.GetWorkflowInputSchemaByCodeParams) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{ID: 2, WorkflowID: arg.WorkflowID, Code: arg.Code, VersionNumber: 1, SchemaJson: []byte(`{"type":"object","properties":{}}`), Status: "active"}, nil
}

func (s *sendMessageStore) GetWorkflowInputSchemaByVersionNumber(_ context.Context, arg db.GetWorkflowInputSchemaByVersionNumberParams) (db.WorkflowInputSchema, error) {
	if s.schemasByVersion != nil {
		if schema, ok := s.schemasByVersion[[2]int32{arg.WorkflowID, arg.VersionNumber}]; ok {
			return schema, nil
		}
	}
	return db.WorkflowInputSchema{ID: arg.VersionNumber + 10, WorkflowID: arg.WorkflowID, Code: "v2", VersionNumber: arg.VersionNumber, SchemaJson: []byte(`{"type":"object","properties":{}}`), Status: "active"}, nil
}

func (s *sendMessageStore) GetDefaultWorkflowInputSchema(_ context.Context, workflowID int32) (db.WorkflowInputSchema, error) {
	return db.WorkflowInputSchema{ID: 2, WorkflowID: workflowID, Code: "public", VersionNumber: 1, SchemaJson: []byte(`{"type":"object","properties":{}}`), Status: "active"}, nil
}

func (s *sendMessageStore) ListActiveExperimentScopesForRouting(_ context.Context, arg db.ListActiveExperimentScopesForRoutingParams) ([]db.ListActiveExperimentScopesForRoutingRow, error) {
	s.routingParams = arg
	return s.scopes, nil
}

func (s *sendMessageStore) ListActiveWorkflowExperimentVariantsByScopeID(_ context.Context, workflowExperimentScopeID int32) ([]db.WorkflowExperimentVariant, error) {
	return s.variants[workflowExperimentScopeID], nil
}

func (s *sendMessageStore) CountExperimentVariantRunsSince(context.Context, db.CountExperimentVariantRunsSinceParams) ([]db.CountExperimentVariantRunsSinceRow, error) {
	return nil, nil
}

func (s *sendMessageStore) ListActiveRoutingCompatibilitiesByInputSchemaID(context.Context, int32) ([]db.ListActiveRoutingCompatibilitiesByInputSchemaIDRow, error) {
	return []db.ListActiveRoutingCompatibilitiesByInputSchemaIDRow{{
		ID:                1,
		WorkflowVersionID: 20,
		CompatibilityType: "native",
		IsActive:          true,
		IsDefaultRoute:    true,
	}}, nil
}

func (s *sendMessageStore) GetDefaultRouteForInputSchema(context.Context, int32) (db.WorkflowVersionInputSchemaCompatibility, error) {
	return db.WorkflowVersionInputSchemaCompatibility{
		ID:                1,
		WorkflowVersionID: 20,
		CompatibilityType: "native",
		IsActive:          true,
		IsDefaultRoute:    true,
	}, nil
}

func (s *sendMessageStore) GetActiveWorkflowVersionInputSchemaCompatibilityByPair(_ context.Context, arg db.GetActiveWorkflowVersionInputSchemaCompatibilityByPairParams) (db.WorkflowVersionInputSchemaCompatibility, error) {
	if s.pairRoutes != nil {
		if route, ok := s.pairRoutes[[2]int32{arg.WorkflowVersionID, arg.WorkflowInputSchemaID}]; ok {
			return route, nil
		}
	}
	return db.WorkflowVersionInputSchemaCompatibility{
		ID:                    1,
		WorkflowVersionID:     arg.WorkflowVersionID,
		WorkflowInputSchemaID: arg.WorkflowInputSchemaID,
		CompatibilityType:     "native",
		IsActive:              true,
	}, nil
}

func (s *sendMessageStore) GetWorkflowInputMapperByID(context.Context, int32) (db.WorkflowInputMapper, error) {
	return db.WorkflowInputMapper{}, nil
}

type executeWorkflowClient struct {
	client.Client
}

func (executeWorkflowClient) ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, ...interface{}) (client.WorkflowRun, error) {
	return nil, nil
}

func TestSendMessageUsesDefaultRuntimeRouteToSelectVersion(t *testing.T) {
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

func TestSendMessageResolvesProcessToWorkflowAndSchemaVersion(t *testing.T) {
	store := &sendMessageStore{
		schemasByVersion: map[[2]int32]db.WorkflowInputSchema{
			{1, 2}: {
				ID:            12,
				WorkflowID:    1,
				Code:          "v2",
				VersionNumber: 2,
				SchemaJson:    []byte(`{"type":"object","properties":{}}`),
				Status:        "active",
			},
		},
	}
	svc := NewService(store, executeWorkflowClient{})

	resp, validationErrors, err := svc.SendMessage(context.Background(), SendMessageRequest{
		Process: "1.v2",
		Value:   map[string]any{"email": "ada@example.com"},
	}, "")

	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.NotNil(t, resp)
	require.Equal(t, int32(1), store.messageArg.WorkflowID)
	require.Equal(t, int32(12), store.messageArg.WorkflowInputSchemaID.Int32)
	require.True(t, store.messageArg.WorkflowInputSchemaID.Valid)
	require.Equal(t, int32(12), *resp.WorkflowInputSchemaID)
	require.Equal(t, int32(2), *resp.WorkflowInputSchemaVersionNumber)
}

func TestSendMessageUsesExplicitExperiment(t *testing.T) {
	experimentID := int32(3)
	store := &sendMessageStore{
		schemasByVersion: map[[2]int32]db.WorkflowInputSchema{
			{1, 2}: {
				ID:            12,
				WorkflowID:    1,
				Code:          "v2",
				VersionNumber: 2,
				SchemaJson:    []byte(`{"type":"object","properties":{}}`),
				Status:        "active",
			},
		},
		scopes: []db.ListActiveExperimentScopesForRoutingRow{{
			WorkflowExperimentID:      3,
			WorkflowID:                1,
			ExperimentType:            "rollout",
			WorkflowExperimentScopeID: 30,
			WorkflowInputSchemaID:     12,
			TrafficConditions:         []byte(`{}`),
			TrafficPercent:            100,
			FallbackPolicy:            "default_route",
		}},
		variants: map[int32][]db.WorkflowExperimentVariant{
			30: {
				{ID: 88, WorkflowExperimentScopeID: 30, WorkflowVersionID: 20, TrafficWeight: 100, IsActive: true},
			},
		},
		pairRoutes: map[[2]int32]db.WorkflowVersionInputSchemaCompatibility{
			{20, 12}: {ID: 91, WorkflowVersionID: 20, WorkflowInputSchemaID: 12, CompatibilityType: "native", IsActive: true},
		},
	}
	svc := NewService(store, executeWorkflowClient{})

	resp, validationErrors, err := svc.SendMessage(context.Background(), SendMessageRequest{
		Process:      "1.v2",
		Experimental: &experimentID,
		Value:        map[string]any{"email": "ada@example.com"},
	}, "")

	require.NoError(t, err)
	require.Empty(t, validationErrors)
	require.NotNil(t, resp)
	require.Equal(t, int32(3), store.routingParams.ExperimentID.Int32)
	require.True(t, store.routingParams.ExperimentID.Valid)
	require.Equal(t, int32(3), store.workflowRunArg.WorkflowExperimentID.Int32)
	require.Equal(t, int32(20), store.workflowRunArg.WorkflowVersionID)
	require.Equal(t, int32(88), store.workflowRunArg.WorkflowExperimentVariantID.Int32)
}

var _ Storage = (*sendMessageStore)(nil)
