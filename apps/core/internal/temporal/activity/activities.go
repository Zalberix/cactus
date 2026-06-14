package activity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/activity"

	"github.com/zalberix/cactus/apps/core/internal/configpub"
	"github.com/zalberix/cactus/apps/core/internal/natssubjects"
	"github.com/zalberix/cactus/apps/core/internal/store"
	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	"github.com/zalberix/cactus/libs/bus"
	"github.com/zalberix/cactus/libs/storage/db"
)

// Activities contains activity implementations for Temporal worker.
type Activities struct {
	store           activityStore
	bus             *bus.Bus
	configPublisher *configpub.Service
	logger          *slog.Logger
}

type activityStore interface {
	GetWorkflowRunStepByRunAndStepID(ctx context.Context, arg db.GetWorkflowRunStepByRunAndStepIDParams) (db.WorkflowRunStep, error)
	CreateWorkflowRunStep(ctx context.Context, arg db.CreateWorkflowRunStepParams) (db.WorkflowRunStep, error)
	UpdateWorkflowRunStepStarted(ctx context.Context, arg db.UpdateWorkflowRunStepStartedParams) (db.WorkflowRunStep, error)
	CreateWorkflowRunStepAttempt(ctx context.Context, arg db.CreateWorkflowRunStepAttemptParams) (db.WorkflowRunStepAttempt, error)
	GetWorkerSettingsRevisionByID(ctx context.Context, id int32) (db.WorkerSettingsRevision, error)
	UpdateWorkflowRunStepAttemptStatus(ctx context.Context, arg db.UpdateWorkflowRunStepAttemptStatusParams) (db.WorkflowRunStepAttempt, error)
	UpdateWorkflowRunStepStatus(ctx context.Context, arg db.UpdateWorkflowRunStepStatusParams) (db.WorkflowRunStep, error)
	ListWorkflowRunStepsByRunID(ctx context.Context, workflowRunID int32) ([]db.WorkflowRunStep, error)
	UpdateWorkflowRunStatus(ctx context.Context, arg db.UpdateWorkflowRunStatusParams) (db.WorkflowRun, error)
}

func activityAttemptNumber(temporalAttempt int32) int32 {
	return temporalAttempt + 1
}

// New creates Activities with dependencies.
func New(store *store.Store, bus *bus.Bus, configPublisher *configpub.Service) *Activities {
	return &Activities{
		store:           store,
		bus:             bus,
		configPublisher: configPublisher,
		logger:          slog.Default(),
	}
}

func configHash(settings []byte) string {
	sum := sha256.Sum256(settings)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func buildTaskDispatch(input temporaltypes.RunTaskStepInput, attempt int32, resolvedInput map[string]any, settingsData []byte) (string, temporaltypes.TaskMessage) {
	scope := natssubjects.WorkerScope{
		OrganizationID: input.Step.OrganizationID,
		WorkTypeID:     input.Step.WorkTypeID,
	}
	replyTo := natssubjects.Result(input.Step.OrganizationID, input.WorkflowRunID, input.Step.ID)
	configSubject := natssubjects.Config(
		input.Step.OrganizationID,
		input.Step.WorkTypeID,
		input.Step.WorkerSettingsRevisionID,
	)
	taskMsg := temporaltypes.TaskMessage{
		WorkflowRunID: input.WorkflowRunID,
		StepID:        input.Step.ID,
		Attempt:       attempt,
		ReplyTo:       replyTo,
		Input:         resolvedInput,
		ConfigRef: temporaltypes.ConfigRef{
			OrganizationID: input.Step.OrganizationID,
			WorkTypeID:     input.Step.WorkTypeID,
			SchemaID:       input.Step.WorkerSettingsSchemaID,
			RevisionID:     input.Step.WorkerSettingsRevisionID,
			ConfigHash:     configHash(settingsData),
			ConfigSubject:  configSubject,
		},
		IdempotencyKey: fmt.Sprintf("%d.%d.%d", input.WorkflowRunID, input.Step.ID, attempt),
	}
	subject := natssubjects.Task(
		scope,
		input.Step.WorkerSettingsSchemaID,
		input.Step.WorkerSettingsRevisionID,
		input.WorkflowRunID,
	)
	return subject, taskMsg
}

// publishWorkflowEvent publishes a workflow event to NATS for WebSocket Hub.
// Subject: event.workflow.{messageID}
// This is fire-and-forget — event publishing failure should not break activity execution.
func (a *Activities) publishWorkflowEvent(ctx context.Context, messageID int32, event temporaltypes.WorkflowEvent) {
	if messageID == 0 {
		return // no messageID available, skip silently
	}
	event.Timestamp = eventTime(time.Now())
	subject := fmt.Sprintf("event.workflow.%d", messageID)
	if err := a.bus.PublishJS(ctx, subject, event); err != nil {
		a.logger.Warn("failed to publish workflow event",
			slog.String("subject", subject),
			slog.String("error", err.Error()),
		)
	}
}

func eventTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func eventDurationMs(startedAt pgtype.Timestamp, end time.Time) *int64 {
	if !startedAt.Valid {
		return nil
	}
	ms := end.UTC().Sub(startedAt.Time).Milliseconds()
	if ms < 0 {
		ms = 0
	}
	return &ms
}

// RunTaskStep is the activity for dispatching task steps.
func (a *Activities) RunTaskStep(ctx context.Context, input temporaltypes.RunTaskStepInput) (temporaltypes.StepResult, error) {
	info := activity.GetInfo(ctx)
	attempt := activityAttemptNumber(info.Attempt)

	a.logger.Info("RunTaskStep starting",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.Int("work_type_id", int(input.Step.WorkTypeID)),
		slog.Int("schema_id", int(input.Step.WorkerSettingsSchemaID)),
		slog.Int("revision_id", int(input.Step.WorkerSettingsRevisionID)),
	)

	resolvedInput, err := ResolveInput(input.Step.InputMapping, input.MessageValue, input.StepOutputs)
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("resolve input mapping for step %d: %w", input.Step.ID, err)
	}
	a.logger.Info("RunTaskStep resolved input",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Any("resolved_input", resolvedInput),
	)

	inputDataJSON, _ := json.Marshal(resolvedInput)

	runStep, err := a.store.GetWorkflowRunStepByRunAndStepID(ctx, db.GetWorkflowRunStepByRunAndStepIDParams{
		WorkflowRunID:  input.WorkflowRunID,
		WorkflowStepID: input.Step.ID,
	})
	if err != nil {
		runStep, err = a.store.CreateWorkflowRunStep(ctx, db.CreateWorkflowRunStepParams{
			WorkflowRunID:  input.WorkflowRunID,
			WorkflowStepID: input.Step.ID,
			WorkerID:       pgtype.Int4{},
			TemporalStepID: pgtype.Text{String: fmt.Sprintf("step-%d", input.Step.ID), Valid: true},
			Status:         temporaltypes.StepStatusPending,
			InputData:      inputDataJSON,
		})
		if err != nil {
			return temporaltypes.StepResult{}, fmt.Errorf("create workflow_run_step: %w", err)
		}
	}

	runStep, err = a.store.UpdateWorkflowRunStepStarted(ctx, db.UpdateWorkflowRunStepStartedParams{
		ID:        runStep.ID,
		WorkerID:  pgtype.Int4{},
		InputData: inputDataJSON,
	})
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("mark workflow_run_step started: %w", err)
	}
	a.logger.Info("RunTaskStep marked run_step running",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Any("run_step_id", runStep.ID),
	)

	a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
		Type:       "step_update",
		StepID:     input.Step.ID,
		RunStepID:  runStep.ID,
		StepType:   input.Step.StepType,
		Status:     temporaltypes.StepStatusRunning,
		InputData:  resolvedInput,
		StartedAt:  eventTime(runStep.StartedAt.Time),
		DurationMs: eventDurationMs(runStep.StartedAt, time.Now()),
	})

	stepAttempt, err := a.store.CreateWorkflowRunStepAttempt(ctx, db.CreateWorkflowRunStepAttemptParams{
		WorkflowRunStepID: runStep.ID,
		WorkerID:          pgtype.Int4{},
		AttemptNumber:     attempt,
		Status:            temporaltypes.StepStatusRunning,
	})
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("create step attempt: %w", err)
	}
	a.logger.Info("RunTaskStep created step attempt",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Any("attempt_id", stepAttempt.ID),
		slog.Int("attempt_number", int(attempt)),
	)

	revision, err := a.store.GetWorkerSettingsRevisionByID(ctx, input.Step.WorkerSettingsRevisionID)
	if err != nil {
		return temporaltypes.StepResult{}, fmt.Errorf("get worker settings revision %d: %w", input.Step.WorkerSettingsRevisionID, err)
	}
	subject, taskMsg := buildTaskDispatch(input, attempt, resolvedInput, revision.SettingsData)
	replyTo := taskMsg.ReplyTo
	if a.configPublisher != nil {
		if err := a.configPublisher.PublishRevision(
			ctx,
			input.Step.OrganizationID,
			input.Step.WorkTypeID,
			input.Step.WorkerSettingsSchemaID,
			input.Step.WorkerSettingsRevisionID,
			revision.SettingsData,
		); err != nil {
			return temporaltypes.StepResult{}, fmt.Errorf("publish config revision: %w", err)
		}
	}
	a.logger.Info("RunTaskStep dispatching task",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.String("subject", subject),
		slog.String("reply_to", replyTo),
		slog.String("idempotency_key", taskMsg.IdempotencyKey),
	)
	if err := a.bus.PublishJS(ctx, subject, taskMsg); err != nil {
		a.logger.Error("RunTaskStep failed to publish task",
			slog.Int("workflow_run_id", int(input.WorkflowRunID)),
			slog.Int("step_id", int(input.Step.ID)),
			slog.String("subject", subject),
			slog.String("error", err.Error()),
		)
		return temporaltypes.StepResult{}, fmt.Errorf("publish task to NATS: %w", err)
	}
	a.logger.Info("RunTaskStep published task",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.String("subject", subject),
	)

	a.logger.Info("RunTaskStep waiting for worker result",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.Duration("timeout", input.Step.Timeout),
	)
	workerResult, err := waitForResult(ctx, a.bus.JS(), replyTo, input.Step.Timeout)
	if err != nil {
		a.logger.Warn("RunTaskStep waitForResult error",
			slog.Int("workflow_run_id", int(input.WorkflowRunID)),
			slog.Int("step_id", int(input.Step.ID)),
			slog.Int("attempt", int(attempt)),
			slog.String("reply_to", replyTo),
			slog.String("error", err.Error()),
		)
		now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		_, _ = a.store.UpdateWorkflowRunStepAttemptStatus(ctx, db.UpdateWorkflowRunStepAttemptStatusParams{
			ID:           stepAttempt.ID,
			Status:       temporaltypes.StepStatusFailed,
			ErrorCode:    pgtype.Text{},
			ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			OutputData:   nil,
			CompletedAt:  now,
		})
		_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
			ID:           runStep.ID,
			Status:       temporaltypes.StepStatusFailed,
			CompletedAt:  now,
			ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
		})
		a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
			Type:        "step_update",
			StepID:      input.Step.ID,
			RunStepID:   runStep.ID,
			StepType:    input.Step.StepType,
			Status:      temporaltypes.StepStatusFailed,
			CompletedAt: eventTime(now.Time),
			DurationMs:  eventDurationMs(runStep.StartedAt, now.Time),
			Error:       err.Error(),
		})
		return temporaltypes.StepResult{}, fmt.Errorf("wait for result step %d: %w", input.Step.ID, err)
	}

	a.logger.Info("RunTaskStep worker result",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.Bool("worker_success", workerResult.Success),
		slog.Any("worker_output", workerResult.Output),
		slog.Int("worker_id", int(workerResult.WorkerID)),
		slog.String("worker_error", workerResult.Error),
	)

	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	outputJSON, _ := json.Marshal(workerResult.Output)

	if workerResult.Success {
		_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
			ID:          runStep.ID,
			Status:      temporaltypes.StepStatusCompleted,
			Outcome:     pgtype.Text{String: "success", Valid: true},
			OutputData:  outputJSON,
			CompletedAt: now,
		})
		_, _ = a.store.UpdateWorkflowRunStepAttemptStatus(ctx, db.UpdateWorkflowRunStepAttemptStatusParams{
			ID:          stepAttempt.ID,
			Status:      temporaltypes.StepStatusCompleted,
			ErrorCode:   pgtype.Text{},
			OutputData:  outputJSON,
			CompletedAt: now,
		})

		a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
			Type:        "step_update",
			StepID:      input.Step.ID,
			RunStepID:   runStep.ID,
			StepType:    input.Step.StepType,
			Status:      temporaltypes.StepStatusCompleted,
			OutputData:  workerResult.Output,
			CompletedAt: eventTime(now.Time),
			DurationMs:  eventDurationMs(runStep.StartedAt, now.Time),
		})

		return temporaltypes.StepResult{
			StepID:   input.Step.ID,
			Success:  true,
			Output:   workerResult.Output,
			WorkerID: workerResult.WorkerID,
			Outcome:  "success",
		}, nil
	}

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

	a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
		Type:        "step_update",
		StepID:      input.Step.ID,
		RunStepID:   runStep.ID,
		StepType:    input.Step.StepType,
		Status:      temporaltypes.StepStatusFailed,
		CompletedAt: eventTime(now.Time),
		DurationMs:  eventDurationMs(runStep.StartedAt, now.Time),
		Error:       workerResult.Error,
	})
	a.logger.Warn("RunTaskStep worker reported failure",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.Step.ID)),
		slog.Int("attempt", int(attempt)),
		slog.String("worker_error", workerResult.Error),
		slog.Int("worker_id", int(workerResult.WorkerID)),
	)

	return temporaltypes.StepResult{
		StepID:  input.Step.ID,
		Success: false,
		Error:   workerResult.Error,
	}, fmt.Errorf("worker reported failure for step %d: %s", input.Step.ID, workerResult.Error)
}

// RecordStep — activity for recording step status in DB.
func (a *Activities) RecordStep(ctx context.Context, input temporaltypes.RecordStepInput) error {
	a.logger.Info("RecordStep",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.StepID)),
		slog.String("status", input.Status),
	)

	switch input.Status {
	case temporaltypes.StepStatusPending:
		runStep, err := a.store.CreateWorkflowRunStep(ctx, db.CreateWorkflowRunStepParams{
			WorkflowRunID:  input.WorkflowRunID,
			WorkflowStepID: input.StepID,
			Status:         temporaltypes.StepStatusPending,
		})
		if err != nil {
			return fmt.Errorf("create pending run step %d: %w", input.StepID, err)
		}

		a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
			Type:      "step_update",
			StepID:    input.StepID,
			RunStepID: runStep.ID,
			Status:    temporaltypes.StepStatusPending,
		})

	case temporaltypes.StepStatusRunning:
		steps, err := a.store.ListWorkflowRunStepsByRunID(ctx, input.WorkflowRunID)
		if err != nil {
			return fmt.Errorf("list run steps for start: %w", err)
		}
		for _, s := range steps {
			if s.WorkflowStepID == input.StepID {
				inputDataJSON, _ := json.Marshal(input.InputData)
				runStep, err := a.store.UpdateWorkflowRunStepStarted(ctx, db.UpdateWorkflowRunStepStartedParams{
					ID:        s.ID,
					WorkerID:  pgtype.Int4{},
					InputData: inputDataJSON,
				})
				if err != nil {
					return fmt.Errorf("mark workflow_run_step started: %w", err)
				}

				a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
					Type:       "step_update",
					StepID:     input.StepID,
					RunStepID:  runStep.ID,
					StepType:   "control",
					Status:     temporaltypes.StepStatusRunning,
					InputData:  input.InputData,
					StartedAt:  eventTime(runStep.StartedAt.Time),
					DurationMs: eventDurationMs(runStep.StartedAt, time.Now()),
				})
				break
			}
		}

	case temporaltypes.StepStatusSkipped:
		steps, err := a.store.ListWorkflowRunStepsByRunID(ctx, input.WorkflowRunID)
		if err != nil {
			return fmt.Errorf("list run steps for skip: %w", err)
		}
		now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		for _, s := range steps {
			if s.WorkflowStepID == input.StepID {
				_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
					ID:          s.ID,
					Status:      temporaltypes.StepStatusSkipped,
					CompletedAt: now,
					StartedAt:   now,
				})

				a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
					Type:        "step_update",
					StepID:      input.StepID,
					RunStepID:   s.ID,
					Status:      temporaltypes.StepStatusSkipped,
					CompletedAt: eventTime(now.Time),
					DurationMs:  eventDurationMs(s.StartedAt, now.Time),
				})
				break
			}
		}
	case temporaltypes.StepStatusCompleted:
		steps, err := a.store.ListWorkflowRunStepsByRunID(ctx, input.WorkflowRunID)
		if err != nil {
			return fmt.Errorf("list run steps for complete: %w", err)
		}
		now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		for _, s := range steps {
			if s.WorkflowStepID == input.StepID {
				outputDataJSON, _ := json.Marshal(input.OutputData)
				_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
					ID:          s.ID,
					Status:      temporaltypes.StepStatusCompleted,
					Outcome:     pgtype.Text{String: input.Outcome, Valid: input.Outcome != ""},
					OutputData:  outputDataJSON,
					CompletedAt: now,
					StartedAt:   now,
				})

				a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
					Type:        "step_update",
					StepID:      input.StepID,
					RunStepID:   s.ID,
					Status:      temporaltypes.StepStatusCompleted,
					StepType:    "control",
					OutputData:  input.OutputData,
					CompletedAt: eventTime(now.Time),
					DurationMs:  eventDurationMs(s.StartedAt, now.Time),
				})
				break
			}
		}

	case temporaltypes.StepStatusFailed:
		steps, err := a.store.ListWorkflowRunStepsByRunID(ctx, input.WorkflowRunID)
		if err != nil {
			return fmt.Errorf("list run steps for fail: %w", err)
		}
		now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		for _, s := range steps {
			if s.WorkflowStepID == input.StepID {
				_, _ = a.store.UpdateWorkflowRunStepStatus(ctx, db.UpdateWorkflowRunStepStatusParams{
					ID:           s.ID,
					Status:       temporaltypes.StepStatusFailed,
					CompletedAt:  now,
					ErrorMessage: pgtype.Text{String: input.ErrorMessage, Valid: input.ErrorMessage != ""},
					StartedAt:    now,
				})

				a.publishWorkflowEvent(ctx, input.MessageID, temporaltypes.WorkflowEvent{
					Type:        "step_update",
					StepID:      input.StepID,
					RunStepID:   s.ID,
					StepType:    "control",
					Status:      temporaltypes.StepStatusFailed,
					CompletedAt: eventTime(now.Time),
					DurationMs:  eventDurationMs(s.StartedAt, now.Time),
					Error:       input.ErrorMessage,
				})
				break
			}
		}
	}
	return nil
}

// UpdateRunStatus updates workflow_run status.
func (a *Activities) UpdateRunStatus(ctx context.Context, workflowRunID int32, messageID int32, status string, errorMsg string) error {
	a.logger.Info("UpdateRunStatus",
		slog.Int("workflow_run_id", int(workflowRunID)),
		slog.Int("message_id", int(messageID)),
		slog.String("status", status),
		slog.String("error_msg", errorMsg),
	)

	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	_, err := a.store.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
		ID:           workflowRunID,
		Status:       status,
		CompletedAt:  now,
		ErrorMessage: pgtype.Text{String: errorMsg, Valid: errorMsg != ""},
	})
	if err != nil {
		return err
	}

	switch status {
	case temporaltypes.RunStatusCompleted:
		a.publishWorkflowEvent(ctx, messageID, temporaltypes.WorkflowEvent{
			Type: "workflow_done",
		})
	case temporaltypes.RunStatusFailed:
		a.publishWorkflowEvent(ctx, messageID, temporaltypes.WorkflowEvent{
			Type:  "workflow_failed",
			Error: errorMsg,
		})
	}

	return nil
}
