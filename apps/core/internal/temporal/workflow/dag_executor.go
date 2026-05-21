package workflow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
	"github.com/zalberix/cactus/apps/core/internal/temporal/expression"
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

	stepMap, children, inDegree, depOutcome := buildGraph(input.Steps, input.Deps)

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

	executeControlStep := func(step temporaltypes.StepDef) (string, error) {
		if err := workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusRunning,
		}).Get(ctx, nil); err != nil {
			return "", err
		}

		outcome := "success"
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
				return "", parseErr
			}
			workflow.Sleep(ctx, delay)
		} else {
			var evalErr error
			outcome, evalErr = evaluateControlOutcome(step, input.MessageValue, collectOutputs(stepResults))
			if evalErr != nil {
				_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
					WorkflowRunID: input.WorkflowRunID,
					MessageID:     input.MessageID,
					StepID:        step.ID,
					Status:        temporaltypes.StepStatusFailed,
					ErrorMessage:  evalErr.Error(),
				}).Get(ctx, nil)
				return "", evalErr
			}
		}

		if err := workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusCompleted,
			Outcome:       outcome,
		}).Get(ctx, nil); err != nil {
			return "", err
		}
		return outcome, nil
	}

	var failErr error
	blockedByBranch := make(map[int32]bool, len(input.Steps))
	var skipBranch func(stepID int32)
	var advanceChildren func(stepID int32, outcome string)

	skipBranch = func(stepID int32) {
		if !remaining[stepID] {
			return
		}
		delete(remaining, stepID)
		_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			MessageID:     input.MessageID,
			StepID:        stepID,
			Status:        temporaltypes.StepStatusSkipped,
		}).Get(ctx, nil)
		for _, childID := range children[stepID] {
			if !remaining[childID] {
				continue
			}
			blockedByBranch[childID] = true
			inDegree[childID]--
			if inDegree[childID] == 0 {
				skipBranch(childID)
			}
		}
	}

	var processReadyStep func(stepID int32)
	advanceChildren = func(stepID int32, outcome string) {
		for _, childID := range children[stepID] {
			expectedOutcome := depOutcome[childID][stepID]
			if expectedOutcome != "" && expectedOutcome != outcome {
				blockedByBranch[childID] = true
			}
			inDegree[childID]--
			if inDegree[childID] != 0 {
				continue
			}
			if blockedByBranch[childID] {
				skipBranch(childID)
				continue
			}
			processReadyStep(childID)
		}
	}

	processReadyStep = func(stepID int32) {
		step, ok := stepMap[stepID]
		if !ok {
			return
		}
		if isControlStep(step) && !isStartStep(step) {
			outcome, err := executeControlStep(step)
			if err != nil {
				failErr = err
				return
			}
			stepResults[step.ID] = temporaltypes.StepResult{
				StepID:  step.ID,
				Success: true,
				Outcome: outcome,
			}
			delete(remaining, step.ID)
			advanceChildren(step.ID, outcome)
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

		advanceChildren(step.ID, "success")
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

			if result.Outcome == "" {
				result.Outcome = "success"
			}
			stepResults[sf.stepID] = result
			delete(remaining, sf.stepID)

			advanceChildren(sf.stepID, result.Outcome)
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

func evaluateControlOutcome(step temporaltypes.StepDef, messageValue []byte, stepOutputs map[int32]map[string]any) (string, error) {
	switch step.ControlKind {
	case "delay":
		return "success", nil
	case "condition":
		return evaluateConditionOutcome(step.ControlSettings, messageValue, stepOutputs)
	case "switch":
		return evaluateSwitchOutcome(step.ControlSettings, messageValue, stepOutputs)
	default:
		return "success", nil
	}
}

func evaluateConditionOutcome(raw json.RawMessage, messageValue []byte, stepOutputs map[int32]map[string]any) (string, error) {
	var settings struct {
		Left     string `json:"left"`
		Operator string `json:"operator"`
		Right    string `json:"right"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return "", err
	}

	operator := strings.TrimSpace(settings.Operator)
	if operator == "exists" || operator == "not_exists" {
		_, exists, err := resolveControlValuePresence(settings.Left, messageValue, stepOutputs)
		if err != nil {
			return "", err
		}
		result := exists
		if operator == "not_exists" {
			result = !result
		}
		return boolOutcome(result), nil
	}

	left, err := resolveControlValue(settings.Left, messageValue, stepOutputs)
	if err != nil {
		return "", err
	}
	right, err := resolveControlValue(settings.Right, messageValue, stepOutputs)
	if err != nil {
		return "", err
	}

	result, err := compareConditionValues(left, right, operator)
	if err != nil {
		return "", err
	}
	return boolOutcome(result), nil
}

func evaluateSwitchOutcome(raw json.RawMessage, messageValue []byte, stepOutputs map[int32]map[string]any) (string, error) {
	var settings struct {
		Expression string            `json:"expression"`
		Cases      []json.RawMessage `json:"cases"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return "", err
	}
	value, err := resolveControlValue(settings.Expression, messageValue, stepOutputs)
	if err != nil {
		return "", err
	}
	for _, item := range settings.Cases {
		caseID, caseValue, err := parseSwitchCaseForRuntime(item)
		if err != nil {
			return "", err
		}
		if scalarEqual(value, caseValue) {
			return caseID, nil
		}
	}
	return "default", nil
}

func parseSwitchCaseForRuntime(raw json.RawMessage) (string, string, error) {
	var legacy string
	if err := json.Unmarshal(raw, &legacy); err == nil {
		value := strings.TrimSpace(legacy)
		if value == "" {
			return "", "", fmt.Errorf("empty switch case")
		}
		return "case-" + slugSwitchCase(value), value, nil
	}

	var objectCase struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &objectCase); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(objectCase.ID) == "" || strings.TrimSpace(objectCase.Value) == "" {
		return "", "", fmt.Errorf("invalid switch case")
	}
	return strings.TrimSpace(objectCase.ID), strings.TrimSpace(objectCase.Value), nil
}

func resolveControlValue(source string, messageValue []byte, stepOutputs map[int32]map[string]any) (any, error) {
	if strings.HasPrefix(strings.TrimSpace(source), "$.") {
		return expression.ResolveSource(source, messageValue, stepOutputs)
	}
	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("empty control value")
	}
	return source, nil
}

func resolveControlValuePresence(source string, messageValue []byte, stepOutputs map[int32]map[string]any) (any, bool, error) {
	if strings.HasPrefix(strings.TrimSpace(source), "$.") {
		return expression.ResolveSourcePresence(source, messageValue, stepOutputs)
	}
	if strings.TrimSpace(source) == "" {
		return nil, false, nil
	}
	return source, true, nil
}

func compareConditionValues(left, right any, operator string) (bool, error) {
	switch operator {
	case "eq":
		return scalarEqual(left, right), nil
	case "neq":
		return !scalarEqual(left, right), nil
	case "gt", "gte", "lt", "lte":
		leftNumber, ok := toNumber(left)
		if !ok {
			return false, fmt.Errorf("left value is not numeric")
		}
		rightNumber, ok := toNumber(right)
		if !ok {
			return false, fmt.Errorf("right value is not numeric")
		}
		switch operator {
		case "gt":
			return leftNumber > rightNumber, nil
		case "gte":
			return leftNumber >= rightNumber, nil
		case "lt":
			return leftNumber < rightNumber, nil
		default:
			return leftNumber <= rightNumber, nil
		}
	case "contains", "not_contains":
		contains := valueContains(left, right)
		if operator == "not_contains" {
			return !contains, nil
		}
		return contains, nil
	default:
		return false, fmt.Errorf("unsupported condition operator: %s", operator)
	}
}

func boolOutcome(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func scalarEqual(left, right any) bool {
	if leftNumber, ok := toNumber(left); ok {
		if rightNumber, ok := toNumber(right); ok {
			return leftNumber == rightNumber
		}
	}
	return fmt.Sprint(left) == fmt.Sprint(right)
}

func toNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		n, err := typed.Float64()
		return n, err == nil
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func valueContains(left, right any) bool {
	if leftString, ok := left.(string); ok {
		return strings.Contains(leftString, fmt.Sprint(right))
	}
	if leftArray, ok := left.([]any); ok {
		for _, item := range leftArray {
			if scalarEqual(item, right) {
				return true
			}
		}
	}
	return false
}

func slugSwitchCase(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
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
