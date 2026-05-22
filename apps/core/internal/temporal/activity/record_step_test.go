package activity

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type recordStepStore struct {
	steps  []db.WorkflowRunStep
	nextID int32
	now    time.Time
}

func newRecordStepStore(now time.Time) *recordStepStore {
	return &recordStepStore{
		nextID: 1,
		now:    now,
	}
}

func (s *recordStepStore) CreateWorkflowRunStep(_ context.Context, arg db.CreateWorkflowRunStepParams) (db.WorkflowRunStep, error) {
	step := db.WorkflowRunStep{
		ID:             s.nextID,
		WorkflowRunID:  arg.WorkflowRunID,
		WorkflowStepID: arg.WorkflowStepID,
		Status:         arg.Status,
		InputData:      arg.InputData,
	}
	s.nextID++
	s.steps = append(s.steps, step)
	return step, nil
}

func (s *recordStepStore) ListWorkflowRunStepsByRunID(_ context.Context, workflowRunID int32) ([]db.WorkflowRunStep, error) {
	result := make([]db.WorkflowRunStep, 0, len(s.steps))
	for _, step := range s.steps {
		if step.WorkflowRunID == workflowRunID {
			result = append(result, step)
		}
	}
	return result, nil
}

func (s *recordStepStore) UpdateWorkflowRunStepStarted(_ context.Context, arg db.UpdateWorkflowRunStepStartedParams) (db.WorkflowRunStep, error) {
	for i := range s.steps {
		if s.steps[i].ID != arg.ID {
			continue
		}
		s.steps[i].Status = temporaltypes.StepStatusRunning
		s.steps[i].InputData = arg.InputData
		s.steps[i].StartedAt = pgtype.Timestamp{Time: s.now, Valid: true}
		return s.steps[i], nil
	}
	return db.WorkflowRunStep{}, nil
}

func (s *recordStepStore) UpdateWorkflowRunStepStatus(_ context.Context, arg db.UpdateWorkflowRunStepStatusParams) (db.WorkflowRunStep, error) {
	for i := range s.steps {
		if s.steps[i].ID != arg.ID {
			continue
		}
		s.steps[i].Status = arg.Status
		s.steps[i].Outcome = arg.Outcome
		s.steps[i].OutputData = arg.OutputData
		s.steps[i].CompletedAt = arg.CompletedAt
		s.steps[i].ErrorMessage = arg.ErrorMessage
		if !s.steps[i].StartedAt.Valid {
			s.steps[i].StartedAt = arg.StartedAt
		}
		return s.steps[i], nil
	}
	return db.WorkflowRunStep{}, nil
}

func (s *recordStepStore) GetWorkflowRunStepByRunAndStepID(context.Context, db.GetWorkflowRunStepByRunAndStepIDParams) (db.WorkflowRunStep, error) {
	return db.WorkflowRunStep{}, nil
}

func (s *recordStepStore) CreateWorkflowRunStepAttempt(context.Context, db.CreateWorkflowRunStepAttemptParams) (db.WorkflowRunStepAttempt, error) {
	return db.WorkflowRunStepAttempt{}, nil
}

func (s *recordStepStore) GetWorkerSettingsRevisionByID(context.Context, int32) (db.WorkerSettingsRevision, error) {
	return db.WorkerSettingsRevision{}, nil
}

func (s *recordStepStore) UpdateWorkflowRunStepAttemptStatus(context.Context, db.UpdateWorkflowRunStepAttemptStatusParams) (db.WorkflowRunStepAttempt, error) {
	return db.WorkflowRunStepAttempt{}, nil
}

func (s *recordStepStore) UpdateWorkflowRunStatus(context.Context, db.UpdateWorkflowRunStatusParams) (db.WorkflowRun, error) {
	return db.WorkflowRun{}, nil
}

func TestRecordStepControlLifecyclePreservesStartedAt(t *testing.T) {
	startedAt := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	store := newRecordStepStore(startedAt)
	activities := &Activities{store: store, logger: slog.Default()}

	err := activities.RecordStep(context.Background(), temporaltypes.RecordStepInput{
		WorkflowRunID: 100,
		StepID:        20,
		Status:        temporaltypes.StepStatusPending,
	})
	require.NoError(t, err)

	err = activities.RecordStep(context.Background(), temporaltypes.RecordStepInput{
		WorkflowRunID: 100,
		StepID:        20,
		Status:        temporaltypes.StepStatusRunning,
	})
	require.NoError(t, err)
	require.Len(t, store.steps, 1)
	require.Equal(t, temporaltypes.StepStatusRunning, store.steps[0].Status)
	require.True(t, store.steps[0].StartedAt.Valid)
	require.Equal(t, startedAt, store.steps[0].StartedAt.Time)

	store.now = startedAt.Add(30 * time.Second)
	err = activities.RecordStep(context.Background(), temporaltypes.RecordStepInput{
		WorkflowRunID: 100,
		StepID:        20,
		Status:        temporaltypes.StepStatusCompleted,
		Outcome:       "success",
	})
	require.NoError(t, err)

	require.Equal(t, temporaltypes.StepStatusCompleted, store.steps[0].Status)
	require.Equal(t, "success", store.steps[0].Outcome.String)
	require.Equal(t, startedAt, store.steps[0].StartedAt.Time)
	require.True(t, store.steps[0].CompletedAt.Valid)
	require.True(t, store.steps[0].CompletedAt.Time.After(store.steps[0].StartedAt.Time))
}
