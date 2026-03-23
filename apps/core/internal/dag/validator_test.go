package dag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zalberix/cactus/apps/core/internal/dag"
)

// makeStep создаёт тестовый шаг.
func makeStep(id int32, stepType dag.StepType, mappings ...dag.MappingEntry) dag.Step {
	return dag.Step{
		ID:           id,
		StepType:     stepType,
		InputMapping: mappings,
	}
}

// makeDep создаёт тестовую зависимость.
func makeDep(stepID, dependsOn int32, outcome string) dag.Dependency {
	return dag.Dependency{
		StepID:          stepID,
		DependsOnStepID: dependsOn,
		Outcome:         outcome,
	}
}

// TestValidateDAG_EmptyDAG — пустой DAG не должен давать ошибок.
func TestValidateDAG_EmptyDAG(t *testing.T) {
	errs := dag.ValidateDAG(nil, nil)
	assert.Empty(t, errs, "пустой DAG не должен содержать ошибок")
}

// TestValidateDAG_SingleStep — один шаг без зависимостей.
func TestValidateDAG_SingleStep(t *testing.T) {
	steps := []dag.Step{makeStep(1, dag.StepTypeTask)}
	errs := dag.ValidateDAG(steps, nil)
	assert.Empty(t, errs)
}

// TestValidateDAG_LinearDAG — линейный DAG A -> B -> C.
func TestValidateDAG_LinearDAG(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"),
		makeDep(3, 2, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs, "линейный DAG не должен содержать ошибок")
}

// TestValidateDAG_CycleDetection — цикл A -> B -> C -> A.
func TestValidateDAG_CycleDetection(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"),
		makeDep(3, 2, "success"),
		makeDep(1, 3, "success"), // замыкает цикл
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.NotEmpty(t, errs, "цикл должен быть обнаружен")
	for _, e := range errs {
		assert.Equal(t, "cycle", e.Type)
	}
}

// TestValidateDAG_DiamondDAG — ромбовидный DAG A -> B, A -> C, B -> D, C -> D.
func TestValidateDAG_DiamondDAG(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask), // A
		makeStep(2, dag.StepTypeTask), // B
		makeStep(3, dag.StepTypeTask), // C
		makeStep(4, dag.StepTypeTask), // D
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"), // B depends on A
		makeDep(3, 1, "success"), // C depends on A
		makeDep(4, 2, "success"), // D depends on B
		makeDep(4, 3, "success"), // D depends on C
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs, "ромбовидный DAG не должен содержать ошибок")
}

// TestValidateDAG_InvalidOutcome — зависимость с невалидным исходом для task-шага.
func TestValidateDAG_InvalidOutcome(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "nonexistent"), // task может иметь только "success"
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Type == "invalid_outcome" {
			found = true
			break
		}
	}
	assert.True(t, found, "должна быть ошибка invalid_outcome")
}

// TestValidateDAG_InvalidMapping — input_mapping ссылается на не-зависимость.
func TestValidateDAG_InvalidMapping(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask, dag.MappingEntry{
			Target: "email",
			Source: "$.steps.2.output.result", // шаг 2 — не зависимость шага 3
		}),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"),
		makeDep(3, 1, "success"), // шаг 3 зависит только от шага 1
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Type == "invalid_mapping" {
			found = true
			break
		}
	}
	assert.True(t, found, "должна быть ошибка invalid_mapping")
}

// TestValidateDAG_MessageSourceAlwaysValid — $.message.value.* всегда валиден.
func TestValidateDAG_MessageSourceAlwaysValid(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask, dag.MappingEntry{
			Target: "email",
			Source: "$.message.value.email",
		}),
	}
	errs := dag.ValidateDAG(steps, nil)
	assert.Empty(t, errs, "$.message.value.* всегда должен быть валидным источником")
}

// TestValidateDAG_StepsDependencyMappingValid — $.steps.{id}.output.* для прямой зависимости.
func TestValidateDAG_StepsDependencyMappingValid(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask, dag.MappingEntry{
			Target: "result",
			Source: "$.steps.1.output.data", // шаг 1 — прямая зависимость шага 2
		}),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs, "маппинг на прямую зависимость должен быть валидным")
}

// TestValidateDAG_DisconnectedSubgraphs — два несвязанных подграфа оба валидируются.
func TestValidateDAG_DisconnectedSubgraphs(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
		makeStep(4, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"), // подграф 1: 1 -> 2
		makeDep(4, 3, "success"), // подграф 2: 3 -> 4
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs, "два несвязанных подграфа без цикла должны быть валидны")
}

// TestValidateDAG_ControlStepOutcomeValid — control-шаг может иметь любой непустой исход.
func TestValidateDAG_ControlStepOutcomeValid(t *testing.T) {
	steps := []dag.Step{
		makeStep(1, dag.StepTypeControl),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "yes"),  // control шаг может иметь custom outcome
		makeDep(3, 1, "no"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs, "control-шаги могут иметь любые непустые исходы")
}

// TestValidateDAG_AllErrorsReturned — все ошибки возвращаются сразу (не stop on first).
func TestValidateDAG_AllErrorsReturned(t *testing.T) {
	// Два невалидных маппинга в разных шагах
	steps := []dag.Step{
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask, dag.MappingEntry{
			Target: "x",
			Source: "$.steps.99.output.data", // несуществующая зависимость
		}),
		makeStep(3, dag.StepTypeTask, dag.MappingEntry{
			Target: "y",
			Source: "$.steps.99.output.data", // тоже несуществующая зависимость
		}),
	}
	deps := []dag.Dependency{
		makeDep(2, 1, "success"),
		makeDep(3, 1, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.GreaterOrEqual(t, len(errs), 2, "должно быть как минимум 2 ошибки (по одной для каждого шага)")
}
