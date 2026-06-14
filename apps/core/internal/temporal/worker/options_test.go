package temporalworker

import (
	"testing"

	"github.com/zalberix/cactus/apps/core/config"
)

func TestBuildWorkerOptionsMapsConfiguredLimits(t *testing.T) {
	cfg := config.TemporalWorker{
		MaxConcurrentActivityExecutionSize:     64,
		MaxConcurrentWorkflowTaskExecutionSize: 32,
		MaxConcurrentActivityTaskPollers:       8,
		MaxConcurrentWorkflowTaskPollers:       4,
		WorkerActivitiesPerSecond:              250,
	}

	opts := BuildWorkerOptions(cfg)

	if opts.MaxConcurrentActivityExecutionSize != 64 {
		t.Fatalf("activity execution limit = %d", opts.MaxConcurrentActivityExecutionSize)
	}
	if opts.MaxConcurrentWorkflowTaskExecutionSize != 32 {
		t.Fatalf("workflow task execution limit = %d", opts.MaxConcurrentWorkflowTaskExecutionSize)
	}
	if opts.MaxConcurrentActivityTaskPollers != 8 {
		t.Fatalf("activity pollers = %d", opts.MaxConcurrentActivityTaskPollers)
	}
	if opts.MaxConcurrentWorkflowTaskPollers != 4 {
		t.Fatalf("workflow task pollers = %d", opts.MaxConcurrentWorkflowTaskPollers)
	}
	if opts.WorkerActivitiesPerSecond != 250 {
		t.Fatalf("activity rate = %f", opts.WorkerActivitiesPerSecond)
	}
}

func TestBuildWorkerOptionsLeavesZeroValuesAsTemporalDefaults(t *testing.T) {
	opts := BuildWorkerOptions(config.TemporalWorker{})

	if opts.MaxConcurrentActivityExecutionSize != 0 {
		t.Fatalf("activity execution limit should keep Temporal default, got %d", opts.MaxConcurrentActivityExecutionSize)
	}
	if opts.MaxConcurrentWorkflowTaskExecutionSize != 0 {
		t.Fatalf("workflow task execution limit should keep Temporal default, got %d", opts.MaxConcurrentWorkflowTaskExecutionSize)
	}
	if opts.MaxConcurrentActivityTaskPollers != 0 {
		t.Fatalf("activity pollers should keep Temporal default, got %d", opts.MaxConcurrentActivityTaskPollers)
	}
	if opts.MaxConcurrentWorkflowTaskPollers != 0 {
		t.Fatalf("workflow pollers should keep Temporal default, got %d", opts.MaxConcurrentWorkflowTaskPollers)
	}
	if opts.WorkerActivitiesPerSecond != 0 {
		t.Fatalf("activity rate should keep Temporal default, got %f", opts.WorkerActivitiesPerSecond)
	}
}
