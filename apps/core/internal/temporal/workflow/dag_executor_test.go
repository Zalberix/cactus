package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

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

func TestParseDelayDurationUsesNumericCountAndUnit(t *testing.T) {
	delay, err := parseDelayDuration([]byte(`{"count":1.5,"unit":"hour"}`))

	require.NoError(t, err)
	require.Equal(t, 90*time.Minute, delay)
}

func TestDAGExecutorSkipsNonSelectedConditionBranch(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(DAGExecutorWorkflow)

	var recorded []temporaltypes.RecordStepInput
	env.RegisterActivityWithOptions(func(_ context.Context, input temporaltypes.RecordStepInput) error {
		recorded = append(recorded, input)
		return nil
	}, activity.RegisterOptions{Name: "RecordStep"})
	env.RegisterActivityWithOptions(func(context.Context, temporaltypes.RunTaskStepInput) (temporaltypes.StepResult, error) {
		return temporaltypes.StepResult{}, nil
	}, activity.RegisterOptions{Name: "RunTaskStep"})
	env.RegisterActivityWithOptions(func(context.Context, int32, int32, string, string) error { return nil }, activity.RegisterOptions{Name: "UpdateRunStatus"})

	env.OnActivity("RunTaskStep", mock.Anything, mock.MatchedBy(func(input temporaltypes.RunTaskStepInput) bool {
		return input.Step.ID == 3
	})).Return(temporaltypes.StepResult{
		StepID:  3,
		Success: true,
		Outcome: "success",
	}, nil)

	env.ExecuteWorkflow(DAGExecutorWorkflow, temporaltypes.DAGInput{
		WorkflowRunID: 10,
		MessageID:     20,
		MessageValue:  []byte(`{"age":21}`),
		Steps: []temporaltypes.StepDef{
			{ID: 1, StepType: "control", ControlKind: "start"},
			{ID: 2, StepType: "control", ControlKind: "condition", ControlSettings: json.RawMessage(`{"left":"$.message.value.age","operator":"gte","right":"18"}`)},
			{ID: 3, StepType: "task", WorkTypeID: 6, WorkerSettingsRevisionID: 29},
			{ID: 4, StepType: "task", WorkTypeID: 7, WorkerSettingsRevisionID: 30},
		},
		Deps: []temporaltypes.DepDef{
			{StepID: 2, DependsOnStepID: 1, Outcome: "success"},
			{StepID: 3, DependsOnStepID: 2, Outcome: "true"},
			{StepID: 4, DependsOnStepID: 2, Outcome: "false"},
		},
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
	require.Contains(t, recordedStepStatuses(recorded), "4:skipped")
}

func TestEvaluateControlOutcomeConditionTrueAndFalse(t *testing.T) {
	step := temporaltypes.StepDef{
		ID:              2,
		StepType:        "control",
		ControlKind:     "condition",
		ControlSettings: json.RawMessage(`{"left":"$.message.value.age","operator":"gte","right":"18"}`),
	}

	outcome, err := evaluateControlOutcome(step, []byte(`{"age":21}`), nil)
	require.NoError(t, err)
	require.Equal(t, "true", outcome)

	outcome, err = evaluateControlOutcome(step, []byte(`{"age":16}`), nil)
	require.NoError(t, err)
	require.Equal(t, "false", outcome)
}

func recordedStepStatuses(recorded []temporaltypes.RecordStepInput) []string {
	statuses := make([]string, 0, len(recorded))
	for _, item := range recorded {
		statuses = append(statuses, fmt.Sprintf("%d:%s", item.StepID, item.Status))
	}
	return statuses
}

func TestEvaluateControlOutcomeConditionExistsMissingFieldIsFalse(t *testing.T) {
	step := temporaltypes.StepDef{
		ID:              2,
		StepType:        "control",
		ControlKind:     "condition",
		ControlSettings: json.RawMessage(`{"left":"$.message.value.missing","operator":"exists"}`),
	}

	outcome, err := evaluateControlOutcome(step, []byte(`{"age":21}`), nil)

	require.NoError(t, err)
	require.Equal(t, "false", outcome)
}

func TestEvaluateControlOutcomeSwitchCustomAndDefaultBranches(t *testing.T) {
	step := temporaltypes.StepDef{
		ID:              2,
		StepType:        "control",
		ControlKind:     "switch",
		ControlSettings: json.RawMessage(`{"expression":"$.message.value.type","cases":[{"id":"case-vip","label":"VIP","value":"vip"}]}`),
	}

	outcome, err := evaluateControlOutcome(step, []byte(`{"type":"vip"}`), nil)
	require.NoError(t, err)
	require.Equal(t, "case-vip", outcome)

	outcome, err = evaluateControlOutcome(step, []byte(`{"type":"regular"}`), nil)
	require.NoError(t, err)
	require.Equal(t, "default", outcome)
}
