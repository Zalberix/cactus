package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/activity"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	"github.com/zalberix/cactus/apps/core/internal/store"
	db "github.com/zalberix/cactus/apps/core/storage/db"
	"github.com/zalberix/cactus/libs/bus"
)

// Activities содержит все Temporal activities с инжектированными зависимостями.
// Методы этого struct регистрируются как activities в Temporal worker.
type Activities struct {
	store  *store.Store
	bus    *bus.Bus
	logger *slog.Logger
}

// New создаёт Activities с зависимостями.
func New(store *store.Store, bus *bus.Bus) *Activities {
	return &Activities{
		store:  store,
		bus:    bus,
		logger: slog.Default(),
	}
}

// RunTaskStep — activity для выполнения task-шага (per EXEC-05, EXEC-06 dispatch side).
//
// Алгоритм:
// 1. Resolve input mapping из message value + step outputs
// 2. Создать workflow_run_step запись со статусом "running"
// 3. Создать workflow_run_step_attempt запись
// 4. Сформировать TaskMessage (per D-06)
// 5. Опубликовать в NATS TASKS stream: subject "task.{work_type_id}.{revision_id}.{run_id}"
//    NOTE: D-02 deviation — используется integer revision ID вместо configRevisionHash.
//    Hash добавляет сложность без явной пользы в v1, integer revision ID проще и достаточен.
// 6. Ждать результат через waitForResult на "result.{run_id}.{step_id}"
//    Per D-05: timeout берётся из input.Step.Timeout (populated from WorkerSettingsRevision
//    in buildDAGInput), falls back to defaultWorkerTimeout (5 min)
// 7. Обновить workflow_run_step и attempt с результатом
// 8. Вернуть StepResult
func (a *Activities) RunTaskStep(ctx context.Context, input temporaltypes.RunTaskStepInput) (temporaltypes.StepResult, error) {
	info := activity.GetInfo(ctx)
	attempt := int32(info.Attempt) + 1 // Temporal attempts 0-based, мы 1-based

	a.logger.Info("RunTaskStep starting",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
	)

	// 1. Resolve input mapping
	resolvedInput, err := ResolveInput(input.Step.InputMapping, input.MessageValue, input.StepOutputs)
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("resolve input mapping for step %d: %w", input.Step.ID, err)
	}

	inputDataJSON, _ := json.Marshal(resolvedInput)

	// 2. Создать workflow_run_step
	runStep, err := a.store.CreateWorkflowRunStep(ctx, db.CreateWorkflowRunStepParams{
		WorkflowRunID:  input.WorkflowRunID,
		WorkflowStepID: input.Step.ID,
		WorkerID:       pgtype.Int4{},
		TemporalStepID: pgtype.Text{String: fmt.Sprintf("step-%d", input.Step.ID), Valid: true},
		Status:         temporaltypes.StepStatusRunning,
		InputData:      inputDataJSON,
	})
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("create workflow_run_step: %w", err)
	}

	// 3. Создать attempt
	stepAttempt, err := a.store.CreateWorkflowRunStepAttempt(ctx, db.CreateWorkflowRunStepAttemptParams{
		WorkflowRunStepID: runStep.ID,
		WorkerID:          pgtype.Int4{},
		AttemptNumber:     attempt,
		Status:            temporaltypes.StepStatusRunning,
	})
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("create step attempt: %w", err)
	}

	// 4. Формируем TaskMessage (per D-06)
	replyTo := fmt.Sprintf("result.%d.%d", input.WorkflowRunID, input.Step.ID)
	idempotencyKey := fmt.Sprintf("%d.%d.%d", input.WorkflowRunID, input.Step.ID, attempt)

	taskMsg := temporaltypes.TaskMessage{
		WorkflowRunID:  input.WorkflowRunID,
		StepID:         input.Step.ID,
		Attempt:        attempt,
		ReplyTo:        replyTo,
		Input:          resolvedInput,
		IdempotencyKey: idempotencyKey,
	}

	// 5. Публикуем в NATS TASKS stream
	// NOTE: D-02 deviation — integer revision ID вместо configRevisionHash (simpler for v1).
	subject := fmt.Sprintf("task.%d.%d.%d",
		input.Step.WorkTypeID,
		input.Step.WorkerSettingsRevisionID,
		input.WorkflowRunID,
	)
	if err := a.bus.PublishJS(ctx, subject, taskMsg); err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("publish task to NATS: %w", err)
	}

	// 6. Ждём результат (per D-05: timeout из Step.Timeout с fallback на default)
	workerResult, err := waitForResult(ctx, a.bus.JS(), replyTo, input.Step.Timeout)
	if err != nil {
		// Обновляем attempt как failed
		now := pgtype.Timestamp{Time: time.Now(), Valid: true}
		_, _ = a.store.UpdateWorkflowRunStepAttemptStatus(ctx, db.UpdateWorkflowRunStepAttemptStatusParams{
			ID:           stepAttempt.ID,
			Status:       temporaltypes.StepStatusFailed,
			ErrorCode:    pgtype.Text{},
			ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			OutputData:   nil,
			CompletedAt:  now,
		})
		return temporaltypes.StepResult{}, fmt.Errorf("wait for result step %d: %w", input.Step.ID, err)
	}

	// 7. Обновляем records
	now := pgtype.Timestamp{Time: time.Now(), Valid: true}
	outputJSON, _ := json.Marshal(workerResult.Output)

	if workerResult.Success {
		// Обновляем step как completed
		_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
			ID:          runStep.ID,
			Status:      temporaltypes.StepStatusCompleted,
			Outcome:     pgtype.Text{String: "success", Valid: true},
			OutputData:  outputJSON,
			CompletedAt: now,
		})
		// Обновляем attempt как completed
		_, _ = a.store.UpdateWorkflowRunStepAttemptStatus(ctx, db.UpdateWorkflowRunStepAttemptStatusParams{
			ID:          stepAttempt.ID,
			Status:      temporaltypes.StepStatusCompleted,
			ErrorCode:   pgtype.Text{},
			OutputData:  outputJSON,
			CompletedAt: now,
		})

		return temporaltypes.StepResult{
			StepID:   input.Step.ID,
			Success:  true,
			Output:   workerResult.Output,
			WorkerID: workerResult.WorkerID,
			Outcome:  "success",
		}, nil
	}

	// Worker reported failure
	_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
		ID:           runStep.ID,
		Status:       temporaltypes.StepStatusFailed,
		CompletedAt:  now,
		ErrorMessage: pgtype.Text{String: workerResult.Error, Valid: true},
	})
	_, _ = a.store.UpdateWorkflowRunStepAttemptStatus(ctx, db.UpdateWorkflowRunStepAttemptStatusParams{
		ID:           stepAttempt.ID,
		Status:       temporaltypes.StepStatusFailed,
		ErrorCode:    pgtype.Text{},
		ErrorMessage: pgtype.Text{String: workerResult.Error, Valid: true},
		CompletedAt:  now,
	})

	return temporaltypes.StepResult{
		StepID:  input.Step.ID,
		Success: false,
		Error:   workerResult.Error,
	}, fmt.Errorf("worker reported failure for step %d: %s", input.Step.ID, workerResult.Error)
}

// RecordStep — activity для записи статуса шага в БД (per EXEC-07).
// Используется для начальных pending записей и skip пропагации.
func (a *Activities) RecordStep(ctx context.Context, input temporaltypes.RecordStepInput) error {
	a.logger.Info("RecordStep",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.StepID)),
		slog.String("status", input.Status),
	)

	switch input.Status {
	case temporaltypes.StepStatusPending:
		// Создаём начальную запись workflow_run_step
		_, err := a.store.CreateWorkflowRunStep(ctx, db.CreateWorkflowRunStepParams{
			WorkflowRunID:  input.WorkflowRunID,
			WorkflowStepID: input.StepID,
			Status:         temporaltypes.StepStatusPending,
		})
		if err != nil {
			return fmt.Errorf("create pending run step %d: %w", input.StepID, err)
		}

	case temporaltypes.StepStatusSkipped:
		// Обновляем существующую запись как skipped
		// Находим run_step по workflow_run_id + workflow_step_id
		steps, err := a.store.ListWorkflowRunStepsByRunID(ctx, input.WorkflowRunID)
		if err != nil {
			return fmt.Errorf("list run steps for skip: %w", err)
		}
		now := pgtype.Timestamp{Time: time.Now(), Valid: true}
		for _, s := range steps {
			if s.WorkflowStepID == input.StepID {
				_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
					ID:          s.ID,
					Status:      temporaltypes.StepStatusSkipped,
					CompletedAt: now,
				})
				break
			}
		}
	}
	return nil
}

// UpdateRunStatus — activity для обновления статуса workflow_run (per D-20).
func (a *Activities) UpdateRunStatus(ctx context.Context, workflowRunID int32, status string, errorMsg string) error {
	a.logger.Info("UpdateRunStatus",
		slog.Int("workflow_run_id", int(workflowRunID)),
		slog.String("status", status),
	)

	now := pgtype.Timestamp{Time: time.Now(), Valid: true}
	_, err := a.store.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
		ID:           workflowRunID,
		Status:       status,
		CompletedAt:  now,
		ErrorMessage: pgtype.Text{String: errorMsg, Valid: errorMsg != ""},
	})
	return err
}
