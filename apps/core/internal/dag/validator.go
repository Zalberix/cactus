package dag

import (
	"fmt"
	"strconv"
	"strings"
)

// StepType represents a workflow step type.
type StepType string

const (
	StepTypeTask    StepType = "task"
	StepTypeControl StepType = "control"
)

const ControlKindStart = "start"
const ControlKindCondition = "condition"
const ControlKindSwitch = "switch"
const ControlKindDelay = "delay"

// Step is a workflow DAG validation step.
type Step struct {
	ID             int32
	StepType       StepType
	ControlKind    string
	InputMapping   []MappingEntry
	WorkflowInputs []string
}

// MappingEntry is an input mapping entry.
type MappingEntry struct {
	Target string
	Source string
}

// Dependency is a dependency between workflow steps.
type Dependency struct {
	StepID          int32
	DependsOnStepID int32
	Outcome         string
}

// ValidationError is a DAG validation error.
type ValidationError struct {
	Type    string `json:"type"`
	StepID  int32  `json:"step_id,omitempty"`
	Message string `json:"message"`
}

// ValidateDAG validates a directed acyclic workflow graph.
func ValidateDAG(steps []Step, deps []Dependency) []ValidationError {
	if len(steps) == 0 {
		return nil
	}

	stepMap := buildStepMap(steps)
	inDegree, adj := buildGraph(steps, deps)
	predecessors := buildPredecessors(steps, deps)
	workflowInputs := buildWorkflowInputSet(steps)

	errors := validateStart(steps, inDegree, adj)
	errors = append(errors, validateCycles(steps, inDegree, adj)...)
	errors = append(errors, validateOutcomes(deps, stepMap)...)
	errors = append(errors, validateRequiredControlOutcomes(steps, deps)...)
	errors = append(errors, validateMappings(steps, predecessors, workflowInputs)...)

	return errors
}

func buildStepMap(steps []Step) map[int32]Step {
	stepMap := make(map[int32]Step, len(steps))
	for _, step := range steps {
		stepMap[step.ID] = step
	}
	return stepMap
}

func buildGraph(steps []Step, deps []Dependency) (map[int32]int, map[int32][]int32) {
	inDegree := make(map[int32]int, len(steps))
	adj := make(map[int32][]int32, len(steps))
	for _, step := range steps {
		inDegree[step.ID] = 0
	}
	for _, dep := range deps {
		inDegree[dep.StepID]++
		adj[dep.DependsOnStepID] = append(adj[dep.DependsOnStepID], dep.StepID)
	}
	return inDegree, adj
}

func buildPredecessors(steps []Step, deps []Dependency) map[int32]map[int32]struct{} {
	parents := make(map[int32][]int32, len(steps))
	for _, dep := range deps {
		parents[dep.StepID] = append(parents[dep.StepID], dep.DependsOnStepID)
	}

	predecessors := make(map[int32]map[int32]struct{}, len(steps))
	for _, step := range steps {
		visited := make(map[int32]struct{})
		queue := append([]int32(nil), parents[step.ID]...)
		for len(queue) > 0 {
			parentID := queue[0]
			queue = queue[1:]
			if _, ok := visited[parentID]; ok {
				continue
			}
			visited[parentID] = struct{}{}
			queue = append(queue, parents[parentID]...)
		}
		predecessors[step.ID] = visited
	}
	return predecessors
}

func buildWorkflowInputSet(steps []Step) map[string]struct{} {
	inputs := make(map[string]struct{})
	for _, step := range steps {
		for _, input := range step.WorkflowInputs {
			if input != "" {
				inputs[input] = struct{}{}
			}
		}
	}
	return inputs
}

func validateStart(steps []Step, inDegree map[int32]int, adj map[int32][]int32) []ValidationError {
	var errors []ValidationError

	startIDs := findStartIDs(steps)
	if len(startIDs) == 0 {
		errors = append(errors, ValidationError{
			Type:    "missing_start",
			Message: "DAG должен содержать один стартовый системный блок",
		})
	}
	if len(startIDs) > 1 {
		for _, id := range startIDs {
			errors = append(errors, ValidationError{
				Type:    "multiple_start",
				StepID:  id,
				Message: "DAG может содержать только один стартовый системный блок",
			})
		}
	}
	if len(startIDs) != 1 {
		return errors
	}

	startID := startIDs[0]
	if inDegree[startID] > 0 {
		errors = append(errors, ValidationError{
			Type:    "start_has_input",
			StepID:  startID,
			Message: fmt.Sprintf("Стартовый блок %d не может иметь входящие связи", startID),
		})
	}
	errors = append(errors, validateStartReachability(steps, startID, inDegree, adj)...)
	return errors
}

func findStartIDs(steps []Step) []int32 {
	startIDs := make([]int32, 0, 1)
	for _, step := range steps {
		if isStartStep(step) {
			startIDs = append(startIDs, step.ID)
		}
	}
	return startIDs
}

func validateStartReachability(
	steps []Step,
	startID int32,
	inDegree map[int32]int,
	adj map[int32][]int32,
) []ValidationError {
	var errors []ValidationError
	for _, step := range steps {
		if step.ID != startID && inDegree[step.ID] == 0 {
			errors = append(errors, unreachableFromStartError(step.ID))
		}
	}

	visited := reachableFrom(startID, adj)
	for _, step := range steps {
		if _, ok := visited[step.ID]; !ok {
			errors = append(errors, unreachableFromStartError(step.ID))
		}
	}
	return errors
}

func reachableFrom(startID int32, adj map[int32][]int32) map[int32]struct{} {
	visited := map[int32]struct{}{startID: {}}
	queue := []int32{startID}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for _, child := range adj[node] {
			if _, ok := visited[child]; ok {
				continue
			}
			visited[child] = struct{}{}
			queue = append(queue, child)
		}
	}
	return visited
}

func unreachableFromStartError(stepID int32) ValidationError {
	return ValidationError{
		Type:    "unreachable_from_start",
		StepID:  stepID,
		Message: fmt.Sprintf("Шаг %d должен быть достижим от стартового блока", stepID),
	}
}

func validateCycles(steps []Step, inDegree map[int32]int, adj map[int32][]int32) []ValidationError {
	degrees := copyInDegree(inDegree)
	queue := make([]int32, 0, len(steps))
	for _, step := range steps {
		if degrees[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}

	processed := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		processed++
		for _, neighbor := range adj[node] {
			degrees[neighbor]--
			if degrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if processed == len(steps) {
		return nil
	}

	var errors []ValidationError
	for _, step := range steps {
		if degrees[step.ID] > 0 {
			errors = append(errors, ValidationError{
				Type:    "cycle",
				StepID:  step.ID,
				Message: fmt.Sprintf("Шаг %d участвует в цикле", step.ID),
			})
		}
	}
	return errors
}

func copyInDegree(inDegree map[int32]int) map[int32]int {
	degrees := make(map[int32]int, len(inDegree))
	for stepID, degree := range inDegree {
		degrees[stepID] = degree
	}
	return degrees
}

func validateOutcomes(deps []Dependency, stepMap map[int32]Step) []ValidationError {
	var errors []ValidationError
	for _, dep := range deps {
		dependsOnStep, ok := stepMap[dep.DependsOnStepID]
		if !ok {
			continue
		}
		errors = append(errors, validateDependencyOutcome(dep, dependsOnStep)...)
	}
	return errors
}

func validateDependencyOutcome(dep Dependency, dependsOnStep Step) []ValidationError {
	var errors []ValidationError
	if dependsOnStep.StepType == StepTypeTask && dep.Outcome != "success" {
		errors = append(errors, ValidationError{
			Type:    "invalid_outcome",
			StepID:  dep.StepID,
			Message: fmt.Sprintf("Шаг %d: task-шаг (id=%d) поддерживает только исход 'success', получено '%s'", dep.StepID, dep.DependsOnStepID, dep.Outcome),
		})
	}
	if dependsOnStep.StepType == StepTypeControl {
		switch {
		case dep.Outcome == "":
			errors = append(errors, ValidationError{
				Type:    "invalid_outcome",
				StepID:  dep.StepID,
				Message: fmt.Sprintf("Шаг %d: исход зависимости от control-шага (id=%d) не может быть пустым", dep.StepID, dep.DependsOnStepID),
			})
		case !isControlOutcomeAllowed(dependsOnStep, dep.Outcome):
			errors = append(errors, ValidationError{
				Type:    "invalid_outcome",
				StepID:  dep.StepID,
				Message: fmt.Sprintf("Шаг %d: control-шаг (id=%d, kind=%s) не поддерживает исход '%s'", dep.StepID, dep.DependsOnStepID, dependsOnStep.ControlKind, dep.Outcome),
			})
		}
	}
	return errors
}

func isControlOutcomeAllowed(step Step, outcome string) bool {
	switch step.ControlKind {
	case ControlKindStart, ControlKindDelay:
		return outcome == "success"
	case ControlKindCondition:
		return outcome == "true" || outcome == "false"
	case ControlKindSwitch:
		return outcome == "default"
	default:
		return true
	}
}

func validateRequiredControlOutcomes(steps []Step, deps []Dependency) []ValidationError {
	outcomesByStep := make(map[int32]map[string]struct{}, len(steps))
	for _, dep := range deps {
		if outcomesByStep[dep.DependsOnStepID] == nil {
			outcomesByStep[dep.DependsOnStepID] = make(map[string]struct{})
		}
		outcomesByStep[dep.DependsOnStepID][dep.Outcome] = struct{}{}
	}

	var errors []ValidationError
	for _, step := range steps {
		for _, outcome := range requiredControlOutcomes(step) {
			if _, ok := outcomesByStep[step.ID][outcome]; ok {
				continue
			}
			errors = append(errors, ValidationError{
				Type:    "missing_control_outcome",
				StepID:  step.ID,
				Message: fmt.Sprintf("Control-шаг %d (%s) должен иметь исход '%s'", step.ID, step.ControlKind, outcome),
			})
		}
	}
	return errors
}

func requiredControlOutcomes(step Step) []string {
	if step.StepType != StepTypeControl {
		return nil
	}
	switch step.ControlKind {
	case ControlKindCondition:
		return []string{"true", "false"}
	case ControlKindSwitch:
		return []string{"default"}
	default:
		return nil
	}
}

func validateMappings(steps []Step, predecessors map[int32]map[int32]struct{}, workflowInputs map[string]struct{}) []ValidationError {
	var errors []ValidationError
	for _, step := range steps {
		for _, mapping := range step.InputMapping {
			if err := validateMappingSource(step.ID, mapping.Source, predecessors[step.ID], workflowInputs); err != nil {
				errors = append(errors, *err)
			}
		}
	}
	return errors
}

func isStartStep(step Step) bool {
	return step.StepType == StepTypeControl && step.ControlKind == ControlKindStart
}

func validateMappingSource(stepID int32, source string, predecessors map[int32]struct{}, workflowInputs map[string]struct{}) *ValidationError {
	if strings.HasPrefix(source, "$.message.value.") {
		fieldPath := strings.TrimPrefix(source, "$.message.value.")
		fieldName := fieldPath
		if dotIdx := strings.Index(fieldName, "."); dotIdx != -1 {
			fieldName = fieldName[:dotIdx]
		}
		if fieldName != "" {
			if _, ok := workflowInputs[fieldName]; ok {
				return nil
			}
		}
		return &ValidationError{
			Type:    "invalid_mapping",
			StepID:  stepID,
			Message: fmt.Sprintf("Шаг %d: маппинг ссылается на необъявленный workflow input: %s", stepID, source),
		}
	}

	if strings.HasPrefix(source, "$.steps.") {
		return validateStepMappingSource(stepID, source, predecessors)
	}

	return &ValidationError{
		Type:    "invalid_mapping",
		StepID:  stepID,
		Message: fmt.Sprintf("Шаг %d: неизвестный формат источника маппинга: %s", stepID, source),
	}
}

func validateStepMappingSource(stepID int32, source string, predecessors map[int32]struct{}) *ValidationError {
	rest := strings.TrimPrefix(source, "$.steps.")
	parts := strings.SplitN(rest, ".", 2)
	if len(parts) < 1 {
		return &ValidationError{
			Type:    "invalid_mapping",
			StepID:  stepID,
			Message: fmt.Sprintf("Шаг %d: некорректный формат источника маппинга: %s", stepID, source),
		}
	}

	refID64, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return &ValidationError{
			Type:    "invalid_mapping",
			StepID:  stepID,
			Message: fmt.Sprintf("Шаг %d: некорректный ID шага в маппинге: %s", stepID, source),
		}
	}

	refID := int32(refID64)
	if _, ok := predecessors[refID]; !ok {
		return &ValidationError{
			Type:    "invalid_mapping",
			StepID:  stepID,
			Message: fmt.Sprintf("Шаг %d: маппинг ссылается на шаг %d, который не является прямой зависимостью", stepID, refID),
		}
	}
	return nil
}
