package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

// DAGExecutorWorkflow выполняет DAG шагов workflow в топологическом порядке.
// Независимые шаги запускаются параллельно через workflow.Go.
// При провале любого шага все оставшиеся помечаются skipped (fail-fast, D-18).
//
// КРИТИЧНО — весь код детерминистичен: нет time.Now, нет go func, нет IO.
func DAGExecutorWorkflow(ctx workflow.Context, input temporaltypes.DAGInput) error {
	// Настройки retry и activity options
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    5 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    2 * time.Minute,
		MaximumAttempts:    3, // D-16: 1 основная + 2 retry
	}
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// Строим граф зависимостей
	stepMap, children, inDegree, _ := buildGraph(input.Steps, input.Deps)

	// Записываем начальные pending записи для всех шагов
	for _, step := range input.Steps {
		_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
			WorkflowRunID: input.WorkflowRunID,
			StepID:        step.ID,
			Status:        temporaltypes.StepStatusPending,
		}).Get(ctx, nil)
	}

	// Хранилище результатов выполненных шагов
	stepResults := make(map[int32]temporaltypes.StepResult, len(input.Steps))

	// Отслеживаем незавершённые шаги
	remaining := make(map[int32]bool, len(input.Steps))
	for _, step := range input.Steps {
		remaining[step.ID] = true
	}

	// Счётчик запущенных и завершённых goroutine
	type stepFuture struct {
		stepID int32
		future workflow.Future
	}

	futures := make([]stepFuture, 0)
	selector := workflow.NewSelector(ctx)

	// Функция для запуска шага
	launchStep := func(step temporaltypes.StepDef) {
		f := workflow.ExecuteActivity(ctx, "RunTaskStep", temporaltypes.RunTaskStepInput{
			WorkflowRunID: input.WorkflowRunID,
			Step:          step,
			Attempt:       1,
			MessageValue:  input.MessageValue,
			StepOutputs:   collectOutputs(stepResults),
		})
		sf := stepFuture{stepID: step.ID, future: f}
		futures = append(futures, sf)
		selector.AddFuture(f, func(f workflow.Future) {})
	}

	// Находим и запускаем корневые шаги (без зависимостей)
	roots := findRoots(inDegree)
	for _, rootID := range roots {
		if step, ok := stepMap[rootID]; ok {
			launchStep(step)
		}
	}

	// Переменная для хранения ошибки провала шага
	var failErr error

	// Основной цикл обработки результатов
	for len(futures) > 0 && failErr == nil {
		selector.Select(ctx)

		// Обрабатываем все завершившиеся futures
		newFutures := make([]stepFuture, 0, len(futures))
		for _, sf := range futures {
			if !sf.future.IsReady() {
				newFutures = append(newFutures, sf)
				continue
			}

			var result temporaltypes.StepResult
			err := sf.future.Get(ctx, &result)

			if err != nil || !result.Success {
				// Fail-fast (D-18): помечаем все оставшиеся шаги как skipped
				delete(remaining, sf.stepID)
				for remID := range remaining {
					_ = workflow.ExecuteActivity(ctx, "RecordStep", temporaltypes.RecordStepInput{
						WorkflowRunID: input.WorkflowRunID,
						StepID:        remID,
						Status:        temporaltypes.StepStatusSkipped,
					}).Get(ctx, nil)
				}
				failErr = temporal.NewApplicationError("step failed", "STEP_FAILED")
				break
			}

			// Шаг успешно завершён
			stepResults[sf.stepID] = result
			delete(remaining, sf.stepID)

			// Уменьшаем inDegree потомков и запускаем готовые
			for _, childID := range children[sf.stepID] {
				inDegree[childID]--
				if inDegree[childID] == 0 {
					if childStep, ok := stepMap[childID]; ok {
						launchStep(childStep)
					}
				}
			}
		}
		futures = newFutures

		// Пересоздаём selector с оставшимися futures
		if len(futures) > 0 && failErr == nil {
			selector = workflow.NewSelector(ctx)
			for _, sf := range futures {
				sfCopy := sf
				selector.AddFuture(sfCopy.future, func(f workflow.Future) {})
			}
		}
	}

	if failErr != nil {
		return failErr
	}

	// Все шаги успешно завершены — обновляем статус workflow run
	_ = workflow.ExecuteActivity(ctx, "UpdateRunStatus", input.WorkflowRunID, temporaltypes.RunStatusCompleted, "").Get(ctx, nil)

	return nil
}

// buildGraph строит структуры данных для обхода DAG.
// Возвращает: stepMap, children, inDegree, depOutcome.
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

// findRoots возвращает список шагов без зависимостей (корневые узлы DAG).
func findRoots(inDegree map[int32]int) []int32 {
	roots := make([]int32, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			roots = append(roots, id)
		}
	}
	return roots
}

// collectOutputs собирает выходные данные завершённых шагов для передачи в следующий шаг.
func collectOutputs(stepResults map[int32]temporaltypes.StepResult) map[int32]map[string]any {
	result := make(map[int32]map[string]any, len(stepResults))
	for id, sr := range stepResults {
		result[id] = sr.Output
	}
	return result
}
