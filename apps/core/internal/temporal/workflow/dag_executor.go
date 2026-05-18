package workflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

// DAGExecutorWorkflow executes DAG workflow sequentially inside Temporal.
// For compatibility with deterministic execution, no mutable external I/O is used
// directly inside this workflow code.
func DAGExecutorWorkflow(ctx workflow.Context, input temporaltypes.DAGInput) error { //nolint:gocognit // Temporal workflow logic must remain deterministic and linear.
	logger := workflow.GetLogger(ctx)

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    5 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    2 * time.Minute,
		MaximumAttempts:    3, // 1 initial attempt + 2 retries
	}
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	logger.Info("DAGExecutorWorkflow started",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("message_id", int(input.MessageID)),
		slog.Int("steps_count", len(input.Steps)),
	)

	stepMap, children, inDegree, _ := buildGraph(input.Steps, input.Deps)

	for _, step := range input.Steps {
		logger.Info("RecordStep pending scheduled",
			slog.Int("workflow_run_id", int(input.WorkflowRunID)),
			slog.Int("step_id", int(step.ID)),
			slog.String("step_type", step.StepType),
			slog.Int("work_type_id", int(step.WorkTypeID)),
		)
		_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusPending,
		}).Get(ctx, nil)
	}

	stepResults := make(map[int32]temporaltypes.StepResult, len(input.Steps))

	remaining := make(map[int32]bool, len(input.Steps))
	for _, step := range input.Steps {
		remaining[step.ID] = true
	}

	type stepFuture struct {
		stepID int32
		future workflow.Future
	}

	futures := make([]stepFuture, 0)

	launchTaskStep := func(step temporaltypes.StepDef) {
		logger.Info("Launching RunTaskStep",
			slog.Int("workflow_run_id", int(input.WorkflowRunID)),
			slog.Int("step_id", int(step.ID)),
			slog.String("step_type", step.StepType),
			slog.Int("work_type_id", int(step.WorkTypeID)),
			slog.Int("revision_id", int(step.WorkerSettingsRevisionID)),
			slog.Int("attempt", 1),
		)
		f := workflow.ExecuteActivity(ctx, "RunTaskStep", temporaltypes.RunTaskStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			Step:          step,
			Attempt:       1,
			MessageValue:  input.MessageValue,
			StepOutputs:   collectOutputs(stepResults),
		})
		sf := stepFuture{stepID: step.ID, future: f}
		futures = append(futures, sf)
	}

	executeControlStep := func(step temporaltypes.StepDef) error {
		if err := workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusRunning,
		}).Get(ctx, nil); err != nil {
			return err
		}

		if step.ControlKind == "delay" {
			delay, parseErr := parseDelayDuration(step.ControlSettings)
			if parseErr != nil {
				_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
					WorkflowRunID: input.WorkflowRunID,
					MessageID:     input.MessageID,
					StepID:        step.ID,
					Status:        temporaltypes.StepStatusFailed,
					ErrorMessage:  parseErr.Error(),
				}).Get(ctx, nil)
				return parseErr
			}
			workflow.Sleep(ctx, delay)
		}

		if err := workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusCompleted,
			Outcome:       "success",
		}).Get(ctx, nil); err != nil {
			return err
		}
		return nil
	}

	var failErr error
	var processReadyStep func(stepID int32)
	processReadyStep = func(stepID int32) {
		step, ok := stepMap[stepID]
		if !ok {
			return
		}
		if isControlStep(step) && !isStartStep(step) {
			if err := executeControlStep(step); err != nil {
				failErr = err
				return
			}
			stepResults[step.ID] = temporaltypes.StepResult{
				StepID:  step.ID,
				Success: true,
				Outcome: "success",
			}
			delete(remaining, step.ID)
			for _, childID := range children[step.ID] {
				inDegree[childID]--
				if inDegree[childID] == 0 {
					processReadyStep(childID)
				}
			}
			return
		}
		if !isStartStep(step) {
			launchTaskStep(step)
			return
		}

		_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusCompleted,
			Outcome:       "success",
		}).Get(ctx, nil)

		stepResults[step.ID] = temporaltypes.StepResult{
			StepID:  step.ID,
			Success: true,
			Outcome: "success",
		}
		delete(remaining, step.ID)

		for _, childID := range children[step.ID] {
			inDegree[childID]--
			if inDegree[childID] == 0 {
				processReadyStep(childID)
			}
		}
	}

	roots := findRoots(inDegree)
	for _, rootID := range roots {
		processReadyStep(rootID)
	}

	buildSelector := func() workflow.Selector {
		selector := workflow.NewSelector(ctx)
		for _, sf := range futures {
			sfCopy := sf
			selector.AddFuture(sfCopy.future, func(_ workflow.Future) {})
		}
		return selector
	}

	selector := buildSelector()

	for len(futures) > 0 && failErr == nil {
		selector.Select(ctx)

		currentFutures := futures
		futures = make([]stepFuture, 0, len(currentFutures))
		for _, sf := range currentFutures {
			if !sf.future.IsReady() {
				futures = append(futures, sf)
				continue
			}

			var result temporaltypes.StepResult
			err := sf.future.Get(ctx, &result)
			logger.Info("RunTaskStep completed",
				slog.Int("workflow_run_id", int(input.WorkflowRunID)),
				slog.Int("step_id", int(sf.stepID)),
				slog.Bool("future_error", err != nil),
				slog.Bool("step_success", result.Success),
				slog.String("result_error", result.Error),
				slog.Int("worker_id", int(result.WorkerID)),
				slog.Int("remaining_before", len(remaining)),
			)

			if err != nil || !result.Success {
				if err != nil {
					logger.Error("RunTaskStep execution returned error",
						slog.Int("workflow_run_id", int(input.WorkflowRunID)),
						slog.Int("step_id", int(sf.stepID)),
						slog.String("error", err.Error()),
					)
				} else {
					logger.Error("RunTaskStep returned failed result",
						slog.Int("workflow_run_id", int(input.WorkflowRunID)),
						slog.Int("step_id", int(sf.stepID)),
						slog.String("step_error", result.Error),
					)
				}

				delete(remaining, sf.stepID)
				for remID := range remaining {
					logger.Info("Marking step as skipped",
						slog.Int("workflow_run_id", int(input.WorkflowRunID)),
						slog.Int("failed_step_id", int(sf.stepID)),
						slog.Int("skipped_step_id", int(remID)),
					)
					_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
						WorkflowRunID: input.WorkflowRunID,
						MessageID:     input.MessageID,
						StepID:        remID,
						Status:        temporaltypes.StepStatusSkipped,
					}).Get(ctx, nil)
				}
				failErr = temporal.NewApplicationError("step failed", "STEP_FAILED")
				break
			}

			stepResults[sf.stepID] = result
			delete(remaining, sf.stepID)

			for _, childID := range children[sf.stepID] {
				inDegree[childID]--
				if inDegree[childID] == 0 {
					processReadyStep(childID)
				}
			}
		}

		if len(futures) > 0 && failErr == nil {
			selector = buildSelector()
		}
	}

	if failErr != nil {
		logger.Info("Workflow failed, updating status",
			slog.Int("workflow_run_id", int(input.WorkflowRunID)),
			slog.Int("message_id", int(input.MessageID)),
			slog.String("status", temporaltypes.RunStatusFailed),
		)
		_ = workflow.ExecuteActivity(ctx, "UpdateRunStatus", input.WorkflowRunID, input.MessageID, temporaltypes.RunStatusFailed, failErr.Error()).Get(ctx, nil)
		return failErr
	}

	logger.Info("Workflow completed, updating status",
		slog.Int("workflow_run_id", int(input.WorkflowRunID)),
		slog.Int("message_id", int(input.MessageID)),
		slog.String("status", temporaltypes.RunStatusCompleted),
	)
	_ = workflow.ExecuteActivity(ctx, "UpdateRunStatus", input.WorkflowRunID, input.MessageID, temporaltypes.RunStatusCompleted, "").Get(ctx, nil)

	return nil
}

func isStartStep(step temporaltypes.StepDef) bool {
	return step.StepType == "control" && step.ControlKind == "start"
}

func isControlStep(step temporaltypes.StepDef) bool {
	return step.StepType == "control"
}

func parseDelayDuration(raw json.RawMessage) (time.Duration, error) {
	type delaySettings struct {
		Count float64 `json:"count"`
		Unit  string  `json:"unit"`
	}
	var settings delaySettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return 0, err
	}
	if settings.Count <= 0 {
		return 0, fmt.Errorf("invalid delay count: %g", settings.Count)
	}
	if strings.TrimSpace(settings.Unit) == "" {
		return 0, fmt.Errorf("delay unit is required")
	}

	unit, ok := delayUnitMultiplier(strings.ToLower(strings.TrimSpace(settings.Unit)))
	if !ok {
		return 0, fmt.Errorf("unsupported delay unit: %s", settings.Unit)
	}
	return time.Duration(settings.Count * float64(unit)), nil
}

func delayUnitMultiplier(unit string) (time.Duration, bool) {
	switch unit {
	case "sec", "s", "seconds", "second":
		return time.Second, true
	case "min", "m", "minutes", "minute":
		return time.Minute, true
	case "hour", "h", "hours":
		return time.Hour, true
	case "day", "d", "days":
		return 24 * time.Hour, true
	default:
		return 0, false
	}
}

func buildGraph(
	steps []temporaltypes.StepDef,
	deps []temporaltypes.DepDef,
) (
	stepMap map[int32]temporaltypes.StepDef,
	children map[int32][]int32,
	inDegree map[int32]int,
	depOutcome map[int32]map[int32]string,
) {
	stepMap = make(map[int32]temporaltypes.StepDef, len(steps))
	children = make(map[int32][]int32, len(steps))
	inDegree = make(map[int32]int, len(steps))
	depOutcome = make(map[int32]map[int32]string)

	for _, s := range steps {
		stepMap[s.ID] = s
		inDegree[s.ID] = 0
	}

	for _, d := range deps {
		inDegree[d.StepID]++
		children[d.DependsOnStepID] = append(children[d.DependsOnStepID], d.StepID)
		if depOutcome[d.StepID] == nil {
			depOutcome[d.StepID] = make(map[int32]string)
		}
		depOutcome[d.StepID][d.DependsOnStepID] = d.Outcome
	}

	return
}

func findRoots(inDegree map[int32]int) []int32 {
	roots := make([]int32, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			roots = append(roots, id)
		}
	}
	return roots
}

func collectOutputs(stepResults map[int32]temporaltypes.StepResult) map[int32]map[string]any {
	result := make(map[int32]map[string]any, len(stepResults))
	for id, sr := range stepResults {
		result[id] = sr.Output
	}
	return result
}
