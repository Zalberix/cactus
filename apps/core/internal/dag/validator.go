package dag

import (
	"fmt"
	"strconv"
	"strings"
)

// StepType представляет тип шага workflow.
type StepType string

const (
	StepTypeTask    StepType = "task"
	StepTypeControl StepType = "control"
)

const ControlKindStart = "start"

// Step — шаг workflow для валидации DAG.
type Step struct {
	ID           int32
	StepType     StepType
	ControlKind  string
	InputMapping []MappingEntry
}

// MappingEntry — одна запись маппинга входных данных.
type MappingEntry struct {
	Target string // поле назначения
	Source string // источник: $.message.value.* или $.steps.{id}.output.*
}

// Dependency — зависимость между шагами.
type Dependency struct {
	StepID          int32
	DependsOnStepID int32
	Outcome         string
}

// ValidationError — ошибка валидации DAG.
type ValidationError struct {
	Type    string `json:"type"` // "cycle", "invalid_outcome", "invalid_mapping"
	StepID  int32  `json:"step_id,omitempty"`
	Message string `json:"message"`
}

// ValidateDAG валидирует направленный ациклический граф workflow.
// Проверяет: отсутствие циклов (алгоритм Кана), корректность исходов,
// корректность ссылок в input_mapping.
// Возвращает все ошибки сразу (не останавливается на первой).
func ValidateDAG(steps []Step, deps []Dependency) []ValidationError {
	if len(steps) == 0 {
		return nil
	}

	var errors []ValidationError

	// Строим карту шагов для быстрого поиска
	stepMap := make(map[int32]Step, len(steps))
	for _, s := range steps {
		stepMap[s.ID] = s
	}

	// Строим граф для алгоритма Кана:
	// adj[dependsOn] -> [stepID] (ребро: dependsOn должен выполниться до stepID)
	// inDegree[stepID] = количество шагов, от которых зависит stepID
	inDegree := make(map[int32]int, len(steps))
	adj := make(map[int32][]int32, len(steps))
	// Инициализируем нулевыми входными степенями
	for _, s := range steps {
		inDegree[s.ID] = 0
	}
	for _, d := range deps {
		inDegree[d.StepID]++
		adj[d.DependsOnStepID] = append(adj[d.DependsOnStepID], d.StepID)
	}

	startIDs := make([]int32, 0, 1)
	for _, s := range steps {
		if isStartStep(s) {
			startIDs = append(startIDs, s.ID)
		}
	}
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
	if len(startIDs) == 1 {
		startID := startIDs[0]
		if inDegree[startID] > 0 {
			errors = append(errors, ValidationError{
				Type:    "start_has_input",
				StepID:  startID,
				Message: fmt.Sprintf("Стартовый блок %d не может иметь входящие связи", startID),
			})
		}
		for _, s := range steps {
			if s.ID != startID && inDegree[s.ID] == 0 {
				errors = append(errors, ValidationError{
					Type:    "unreachable_from_start",
					StepID:  s.ID,
					Message: fmt.Sprintf("Шаг %d должен быть достижим от стартового блока", s.ID),
				})
			}
		}

		visited := map[int32]struct{}{startID: {}}
		queueFromStart := []int32{startID}
		for len(queueFromStart) > 0 {
			node := queueFromStart[0]
			queueFromStart = queueFromStart[1:]
			for _, child := range adj[node] {
				if _, ok := visited[child]; ok {
					continue
				}
				visited[child] = struct{}{}
				queueFromStart = append(queueFromStart, child)
			}
		}
		for _, s := range steps {
			if _, ok := visited[s.ID]; !ok {
				errors = append(errors, ValidationError{
					Type:    "unreachable_from_start",
					StepID:  s.ID,
					Message: fmt.Sprintf("Шаг %d должен быть достижим от стартового блока", s.ID),
				})
			}
		}
	}

	// Алгоритм Кана: обнаружение циклов
	queue := make([]int32, 0, len(steps))
	for _, s := range steps {
		if inDegree[s.ID] == 0 {
			queue = append(queue, s.ID)
		}
	}

	processed := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		processed++
		for _, neighbor := range adj[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Если обработано меньше шагов, чем существует — есть цикл
	if processed < len(steps) {
		for _, s := range steps {
			if inDegree[s.ID] > 0 {
				errors = append(errors, ValidationError{
					Type:    "cycle",
					StepID:  s.ID,
					Message: fmt.Sprintf("Шаг %d участвует в цикле", s.ID),
				})
			}
		}
	}

	// Проверка исходов зависимостей
	for _, d := range deps {
		dependsOnStep, ok := stepMap[d.DependsOnStepID]
		if !ok {
			continue
		}
		// Task-шаги поддерживают только исход "success"
		if dependsOnStep.StepType == StepTypeTask && d.Outcome != "success" {
			errors = append(errors, ValidationError{
				Type:    "invalid_outcome",
				StepID:  d.StepID,
				Message: fmt.Sprintf("Шаг %d: task-шаг (id=%d) поддерживает только исход 'success', получено '%s'", d.StepID, d.DependsOnStepID, d.Outcome),
			})
		}
		if isStartStep(dependsOnStep) && d.Outcome != "success" {
			errors = append(errors, ValidationError{
				Type:    "invalid_outcome",
				StepID:  d.StepID,
				Message: fmt.Sprintf("Шаг %d: start-шаг (id=%d) поддерживает только исход 'success', получено '%s'", d.StepID, d.DependsOnStepID, d.Outcome),
			})
		}
		// Control-шаги могут иметь любой непустой исход
		if dependsOnStep.StepType == StepTypeControl && d.Outcome == "" {
			errors = append(errors, ValidationError{
				Type:    "invalid_outcome",
				StepID:  d.StepID,
				Message: fmt.Sprintf("Шаг %d: исход зависимости от control-шага (id=%d) не может быть пустым", d.StepID, d.DependsOnStepID),
			})
		}
	}

	// Строим множество прямых зависимостей для каждого шага
	// directDeps[stepID] = set of dependsOnStepID
	directDeps := make(map[int32]map[int32]struct{}, len(steps))
	for _, d := range deps {
		if directDeps[d.StepID] == nil {
			directDeps[d.StepID] = make(map[int32]struct{})
		}
		directDeps[d.StepID][d.DependsOnStepID] = struct{}{}
	}

	// Проверка input_mapping
	for _, s := range steps {
		for _, mapping := range s.InputMapping {
			if err := validateMappingSource(s.ID, mapping.Source, directDeps[s.ID]); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

func isStartStep(step Step) bool {
	return step.StepType == StepTypeControl && step.ControlKind == ControlKindStart
}

// validateMappingSource проверяет источник маппинга.
// Допустимые форматы:
//   - $.message.value.{field} — всегда валиден
//   - $.steps.{id}.output.{field} — id должен быть прямой зависимостью шага
func validateMappingSource(stepID int32, source string, directDeps map[int32]struct{}) *ValidationError {
	if strings.HasPrefix(source, "$.message.value.") {
		// Всегда валидный источник
		return nil
	}

	if strings.HasPrefix(source, "$.steps.") {
		// Формат: $.steps.{id}.output.{field}
		rest := strings.TrimPrefix(source, "$.steps.")
		parts := strings.SplitN(rest, ".", 2)
		if len(parts) < 1 {
			return &ValidationError{
				Type:    "invalid_mapping",
				StepID:  stepID,
				Message: fmt.Sprintf("Шаг %d: некорректный формат источника маппинга: %s", stepID, source),
			}
		}
		refIDStr := parts[0]
		refID64, err := strconv.ParseInt(refIDStr, 10, 32)
		if err != nil {
			return &ValidationError{
				Type:    "invalid_mapping",
				StepID:  stepID,
				Message: fmt.Sprintf("Шаг %d: некорректный ID шага в маппинге: %s", stepID, source),
			}
		}
		refID := int32(refID64)

		// Проверяем, что referenced шаг является прямой зависимостью
		if _, ok := directDeps[refID]; !ok {
			return &ValidationError{
				Type:    "invalid_mapping",
				StepID:  stepID,
				Message: fmt.Sprintf("Шаг %d: маппинг ссылается на шаг %d, который не является прямой зависимостью", stepID, refID),
			}
		}
		return nil
	}

	// Неизвестный формат источника
	return &ValidationError{
		Type:    "invalid_mapping",
		StepID:  stepID,
		Message: fmt.Sprintf("Шаг %d: неизвестный формат источника маппинга: %s", stepID, source),
	}
}
