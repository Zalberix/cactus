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

// Step is a workflow DAG validation step.
type Step struct {
	ID           int32
	StepType     StepType
	ControlKind  string
	InputMapping []MappingEntry
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
	directDeps := buildDirectDeps(steps, deps)

	errors := validateStart(steps, inDegree, adj)
	errors = append(errors, validateCycles(steps, inDegree, adj)...)
	errors = append(errors, validateOutcomes(deps, stepMap)...)
	errors = append(errors, validateMappings(steps, directDeps)...)

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

func buildDirectDeps(steps []Step, deps []Dependency) map[int32]map[int32]struct{} {
	directDeps := make(map[int32]map[int32]struct{}, len(steps))
	for _, dep := range deps {
		if directDeps[dep.StepID] == nil {
			directDeps[dep.StepID] = make(map[int32]struct{})
		}
		directDeps[dep.StepID][dep.DependsOnStepID] = struct{}{}
	}
	return directDeps
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
	if isStartStep(dependsOnStep) && dep.Outcome != "success" {
		errors = append(errors, ValidationError{
			Type:    "invalid_outcome",
			StepID:  dep.StepID,
			Message: fmt.Sprintf("Шаг %d: start-шаг (id=%d) поддерживает только исход 'success', получено '%s'", dep.StepID, dep.DependsOnStepID, dep.Outcome),
		})
	}
	if dependsOnStep.StepType == StepTypeControl && dep.Outcome == "" {
		errors = append(errors, ValidationError{
			Type:    "invalid_outcome",
			StepID:  dep.StepID,
			Message: fmt.Sprintf("Шаг %d: исход зависимости от control-шага (id=%d) не может быть пустым", dep.StepID, dep.DependsOnStepID),
		})
	}
	return errors
}

func validateMappings(steps []Step, directDeps map[int32]map[int32]struct{}) []ValidationError {
	var errors []ValidationError
	for _, step := range steps {
		for _, mapping := range step.InputMapping {
			if err := validateMappingSource(step.ID, mapping.Source, directDeps[step.ID]); err != nil {
				errors = append(errors, *err)
			}
		}
	}
	return errors
}

func isStartStep(step Step) bool {
	return step.StepType == StepTypeControl && step.ControlKind == ControlKindStart
}

func validateMappingSource(stepID int32, source string, directDeps map[int32]struct{}) *ValidationError {
	if strings.HasPrefix(source, "$.message.value.") {
		return nil
	}

	if strings.HasPrefix(source, "$.steps.") {
		return validateStepMappingSource(stepID, source, directDeps)
	}

	return &ValidationError{
		Type:    "invalid_mapping",
		StepID:  stepID,
		Message: fmt.Sprintf("Шаг %d: неизвестный формат источника маппинга: %s", stepID, source),
	}
}

func validateStepMappingSource(stepID int32, source string, directDeps map[int32]struct{}) *ValidationError {
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
	if _, ok := directDeps[refID]; !ok {
		return &ValidationError{
			Type:    "invalid_mapping",
			StepID:  stepID,
			Message: fmt.Sprintf("Шаг %d: маппинг ссылается на шаг %d, который не является прямой зависимостью", stepID, refID),
		}
	}
	return nil
}
