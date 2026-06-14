package dag_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zalberix/cactus/apps/manager/internal/dag"
)

func makeStep(id int32, stepType dag.StepType, mappings ...dag.MappingEntry) dag.Step {
	return dag.Step{
		ID:           id,
		StepType:     stepType,
		InputMapping: mappings,
	}
}

func withControlSettings(step dag.Step, raw string) dag.Step {
	step.ControlSettings = json.RawMessage(raw)
	return step
}

func withWorkflowInputs(step dag.Step, inputs ...string) dag.Step {
	step.WorkflowInputs = inputs
	return step
}

func makeDep(stepID, dependsOn int32, outcome string) dag.Dependency {
	return dag.Dependency{
		StepID:          stepID,
		DependsOnStepID: dependsOn,
		Outcome:         outcome,
	}
}

func makeStart() dag.Step {
	step := makeStep(100, dag.StepTypeControl)
	step.ControlKind = dag.ControlKindStart
	return step
}

func TestValidateDAG_EmptyDAG(t *testing.T) {
	errs := dag.ValidateDAG(nil, nil)
	assert.Empty(t, errs)
}

func TestValidateDAG_MissingStart(t *testing.T) {
	errs := dag.ValidateDAG([]dag.Step{makeStep(1, dag.StepTypeTask)}, nil)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "missing_start", errs[0].Type)
}

func TestValidateDAG_SingleTaskWithStart(t *testing.T) {
	steps := []dag.Step{makeStart(), makeStep(1, dag.StepTypeTask)}
	deps := []dag.Dependency{makeDep(1, 100, "success")}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_LinearDAG(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 2, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_CycleDetection(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 2, "success"),
		makeDep(1, 3, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.NotEmpty(t, errs)
	assert.Contains(t, collectTypes(errs), "cycle")
}

func TestValidateDAG_DiamondDAG(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
		makeStep(4, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 1, "success"),
		makeDep(4, 2, "success"),
		makeDep(4, 3, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_InvalidTaskOutcome(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "nonexistent"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "invalid_outcome")
}

func TestValidateDAG_InvalidStartOutcome(t *testing.T) {
	steps := []dag.Step{makeStart(), makeStep(1, dag.StepTypeTask)}
	deps := []dag.Dependency{makeDep(1, 100, "manual")}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "invalid_outcome")
}

func TestValidateDAG_InvalidMapping(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask, dag.MappingEntry{
			Target: "email",
			Source: "$.steps.2.output.result",
		}),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 1, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "invalid_mapping")
}

func TestValidateDAG_DeclaredMessageSourceValid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		withWorkflowInputs(makeStep(1, dag.StepTypeTask, dag.MappingEntry{
			Target: "email",
			Source: "$.message.value.email",
		}), "email"),
	}
	deps := []dag.Dependency{makeDep(1, 100, "success")}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_UndeclaredMessageSourceInvalid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask, dag.MappingEntry{
			Target: "email",
			Source: "$.message.value.email",
		}),
	}
	deps := []dag.Dependency{makeDep(1, 100, "success")}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "invalid_mapping")
}

func TestValidateDAG_StepsDependencyMappingValid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask, dag.MappingEntry{
			Target: "result",
			Source: "$.steps.1.output.data",
		}),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_TransitivePredecessorMappingValid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask, dag.MappingEntry{
			Target: "result",
			Source: "$.steps.1.output.data",
		}),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 2, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_DisconnectedSubgraphInvalid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
		makeStep(4, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(4, 3, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "unreachable_from_start")
}

func TestValidateDAG_ControlStepOutcomeValid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeControl),
		makeStep(2, dag.StepTypeTask),
		makeStep(3, dag.StepTypeTask),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "yes"),
		makeDep(3, 1, "no"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.Empty(t, errs)
}

func TestValidateDAG_ConditionTrueFalseOutcomesValid(t *testing.T) {
	condition := makeStep(1, dag.StepTypeControl)
	condition.ControlKind = "condition"
	steps := []dag.Step{makeStart(), condition, makeStep(2, dag.StepTypeTask), makeStep(3, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "true"),
		makeDep(3, 1, "false"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Empty(t, errs)
}

func TestValidateDAG_ConditionRequiresTrueAndFalseOutcomes(t *testing.T) {
	condition := makeStep(1, dag.StepTypeControl)
	condition.ControlKind = "condition"
	steps := []dag.Step{makeStart(), condition, makeStep(2, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "true"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Contains(t, collectTypes(errs), "missing_control_outcome")
}

func TestValidateDAG_SwitchCustomCaseOutcomeValid(t *testing.T) {
	switchStep := makeStep(1, dag.StepTypeControl)
	switchStep.ControlKind = dag.ControlKindSwitch
	switchStep = withControlSettings(switchStep, `{"expression":"$.message.value.type","cases":[{"id":"case-vip","label":"VIP","value":"vip"}]}`)
	steps := []dag.Step{makeStart(), switchStep, makeStep(2, dag.StepTypeTask), makeStep(3, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "case-vip"),
		makeDep(3, 1, "default"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Empty(t, errs)
}

func TestValidateDAG_SwitchRejectsUnknownOutcome(t *testing.T) {
	switchStep := makeStep(1, dag.StepTypeControl)
	switchStep.ControlKind = dag.ControlKindSwitch
	switchStep = withControlSettings(switchStep, `{"expression":"$.message.value.type","cases":[{"id":"case-vip","label":"VIP","value":"vip"}]}`)
	steps := []dag.Step{makeStart(), switchStep, makeStep(2, dag.StepTypeTask), makeStep(3, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "case-missing"),
		makeDep(3, 1, "default"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Contains(t, collectTypes(errs), "invalid_outcome")
}

func TestValidateDAG_SwitchRequiresCustomCaseOutcomes(t *testing.T) {
	switchStep := makeStep(1, dag.StepTypeControl)
	switchStep.ControlKind = dag.ControlKindSwitch
	switchStep = withControlSettings(switchStep, `{"expression":"$.message.value.type","cases":[{"id":"case-vip","label":"VIP","value":"vip"}]}`)
	steps := []dag.Step{makeStart(), switchStep, makeStep(2, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "default"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Contains(t, collectTypes(errs), "missing_control_outcome")
}

func TestValidateDAG_SwitchLegacyStringCasesRemainValid(t *testing.T) {
	switchStep := makeStep(1, dag.StepTypeControl)
	switchStep.ControlKind = dag.ControlKindSwitch
	switchStep = withControlSettings(switchStep, `{"expression":"$.message.value.type","cases":["vip"]}`)
	steps := []dag.Step{makeStart(), switchStep, makeStep(2, dag.StepTypeTask), makeStep(3, dag.StepTypeTask)}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "case-vip"),
		makeDep(3, 1, "default"),
	}

	errs := dag.ValidateDAG(steps, deps)

	assert.Empty(t, errs)
}

func TestValidateDAG_MultipleStartInvalid(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		func() dag.Step {
			step := makeStep(101, dag.StepTypeControl)
			step.ControlKind = dag.ControlKindStart
			return step
		}(),
		makeStep(1, dag.StepTypeTask),
	}
	deps := []dag.Dependency{makeDep(1, 100, "success")}
	errs := dag.ValidateDAG(steps, deps)
	assert.Contains(t, collectTypes(errs), "multiple_start")
}

func TestValidateDAG_AllErrorsReturned(t *testing.T) {
	steps := []dag.Step{
		makeStart(),
		makeStep(1, dag.StepTypeTask),
		makeStep(2, dag.StepTypeTask, dag.MappingEntry{
			Target: "x",
			Source: "$.steps.99.output.data",
		}),
		makeStep(3, dag.StepTypeTask, dag.MappingEntry{
			Target: "y",
			Source: "$.steps.99.output.data",
		}),
	}
	deps := []dag.Dependency{
		makeDep(1, 100, "success"),
		makeDep(2, 1, "success"),
		makeDep(3, 1, "success"),
	}
	errs := dag.ValidateDAG(steps, deps)
	assert.GreaterOrEqual(t, len(errs), 2)
}

func collectTypes(errs []dag.ValidationError) []string {
	types := make([]string, 0, len(errs))
	for _, err := range errs {
		types = append(types, err.Type)
	}
	return types
}
