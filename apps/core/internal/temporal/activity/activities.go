package activity

import (
	"context"
	"log/slog"

	"github.com/zalberix/cactus/apps/core/internal/store"
	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
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

// RunTaskStep — activity для выполнения task-шага.
// Stub — реализация в Plan 02.
func (a *Activities) RunTaskStep(ctx context.Context, input temporaltypes.RunTaskStepInput) (temporaltypes.StepResult, error) {
	// TODO: Plan 02 — NATS publish + result wait
	return temporaltypes.StepResult{
		StepID:  input.Step.ID,
		Success: true,
		Outcome: "success",
	}, nil
}

// RecordStep — activity для записи телеметрии шага в БД.
// Stub — реализация в Plan 02.
func (a *Activities) RecordStep(ctx context.Context, input temporaltypes.RecordStepInput) error {
	// TODO: Plan 02 — SQLC записи в workflow_run_step + workflow_run_step_attempt
	a.logger.Info("RecordStep stub",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("step_id", int(input.StepID)),
		slog.String("status", input.Status),
	)
	return nil
}

// UpdateRunStatus — activity для обновления статуса workflow_run.
// Stub — реализация в Plan 02.
func (a *Activities) UpdateRunStatus(ctx context.Context, workflowRunID int32, status string, errorMsg string) error {
	// TODO: Plan 02
	a.logger.Info("UpdateRunStatus stub",
		slog.Int("workflow_run_id", int(workflowRunID)),
		slog.String("status", status),
	)
	return nil
}
