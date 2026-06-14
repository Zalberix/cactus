package temporalworker

import (
	"context"
	"fmt"
	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/config"
	"github.com/zalberix/cactus/apps/core/internal/temporal/activity"
	dagworkflow "github.com/zalberix/cactus/apps/core/internal/temporal/workflow"
)

// NewTemporalClient создаёт Temporal client для fx DI.
func NewTemporalClient(cfg *config.Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal client dial: %w", err)
	}
	slog.Info("Temporal client connected",
		slog.String("host", cfg.Temporal.HostPort),
		slog.String("namespace", cfg.Temporal.Namespace),
	)
	return c, nil
}

// BuildWorkerOptions maps cactus config to Temporal SDK worker options.
// Zero values are intentionally passed through so the Temporal SDK keeps its defaults.
func BuildWorkerOptions(cfg config.TemporalWorker) worker.Options {
	return worker.Options{
		MaxConcurrentActivityExecutionSize:     cfg.MaxConcurrentActivityExecutionSize,
		MaxConcurrentWorkflowTaskExecutionSize: cfg.MaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityTaskPollers:       cfg.MaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers:       cfg.MaxConcurrentWorkflowTaskPollers,
		WorkerActivitiesPerSecond:              cfg.WorkerActivitiesPerSecond,
	}
}

// RegisterTemporalWorker регистрирует Temporal worker в fx lifecycle.
func RegisterTemporalWorker(lc fx.Lifecycle, c client.Client, cfg *config.Config, acts *activity.Activities) {
	w := worker.New(c, cfg.Temporal.TaskQueue, BuildWorkerOptions(cfg.Temporal.Worker))
	w.RegisterWorkflow(dagworkflow.DAGExecutorWorkflow)
	w.RegisterActivity(acts)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			slog.Info("Starting Temporal worker", slog.String("task_queue", cfg.Temporal.TaskQueue))
			return w.Start()
		},
		OnStop: func(_ context.Context) error {
			slog.Info("Stopping Temporal worker")
			w.Stop()
			return nil
		},
	})
}
