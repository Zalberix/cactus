package workflows

import "go.temporal.io/sdk/workflow"

// Add workflow functions here and register them in
// apps/core/internal/temporal/worker/worker.go via w.RegisterWorkflow().
//
// Example:
//
//	func ExampleWorkflow(ctx workflow.Context, input string) (string, error) {
//	    ao := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second}
//	    ctx = workflow.WithActivityOptions(ctx, ao)
//	    var result string
//	    err := workflow.ExecuteActivity(ctx, ExampleActivity, input).Get(ctx, &result)
//	    return result, err
//	}

var _ workflow.Context // prevent unused import
