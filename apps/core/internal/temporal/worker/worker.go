package temporalworker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	// "github.com/zalberix/cactus/apps/core/internal/temporal/workflows"
)

type Opts struct {
	fx.In
	Client client.Client
}

func NewFx(opts Opts) Worker {
	return New(opts.Client)
}

// Worker is a re-export of the Temporal SDK worker type for use in fx wiring.
type Worker = worker.Worker

const TaskQueue = "cactus-core"

// New creates a Temporal worker bound to TaskQueue and registers all workflows
// and activities. Add registrations below as you implement them.
func New(c client.Client) Worker {
	w := worker.New(c, TaskQueue, worker.Options{})

	// Register workflows:
	// w.RegisterWorkflow(workflows.ExampleWorkflow)

	// Register activities:
	// w.RegisterActivity(...)

	return w
}
