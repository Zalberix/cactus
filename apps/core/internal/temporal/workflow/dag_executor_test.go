package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

func TestDAGExecutorWaitsForTaskLaunchedByCompletedTask(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(DAGExecutorWorkflow)
	env.RegisterActivityWithOptions(func(context.Context, temporaltypes.RecordStepInput) error { return nil }, activity.RegisterOptions{Name: "RecordStep"})
	env.RegisterActivityWithOptions(func(context.Context, temporaltypes.RunTaskStepInput) (temporaltypes.StepResult, error) {
		return temporaltypes.StepResult{}, nil
	}, activity.RegisterOptions{Name: "RunTaskStep"})
	env.RegisterActivityWithOptions(func(context.Context, int32, int32, string, string) error { return nil }, activity.RegisterOptions{Name: "UpdateRunStatus"})

	env.OnActivity("RecordStep", mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("UpdateRunStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("RunTaskStep", mock.Anything, mock.MatchedBy(func(input temporaltypes.RunTaskStepInput) bool {
		return input.Step.ID == 2
	})).Return(temporaltypes.StepResult{
		StepID:  2,
		Success: true,
		Output: map[string]any{
			"body": "rendered html",
		},
		Outcome: "success",
	}, nil)
	env.OnActivity("RunTaskStep", mock.Anything, mock.MatchedBy(func(input temporaltypes.RunTaskStepInput) bool {
		return input.Step.ID == 3 && input.StepOutputs[2]["body"] == "rendered html"
	})).Return(temporaltypes.StepResult{}, errors.New("smtp failed"))

	env.ExecuteWorkflow(DAGExecutorWorkflow, temporaltypes.DAGInput{
		WorkflowRunID: 10,
		MessageID:     20,
		MessageValue:  []byte(`{"to":"a@example.test"}`),
		Steps: []temporaltypes.StepDef{
			{ID: 1, StepType: "control", ControlKind: "start"},
			{ID: 2, StepType: "task", WorkTypeID: 6, WorkerSettingsRevisionID: 29},
			{ID: 3, StepType: "task", WorkTypeID: 1, WorkerSettingsRevisionID: 30},
		},
		Deps: []temporaltypes.DepDef{
			{StepID: 2, DependsOnStepID: 1, Outcome: "success"},
			{StepID: 3, DependsOnStepID: 2, Outcome: "success"},
		},
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
