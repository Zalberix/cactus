package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dagpkg "github.com/zalberix/cactus/apps/core/internal/dag"
	db "github.com/zalberix/cactus/apps/core/storage/db"
)

const (
	jsonSchemaStringType      = "string"
	messageValueSourcePrefix  = "$.message.value."
	stepOutputSourcePrefix    = "$.steps."
	unknownMappingSourceStart = "$."
)

type workflowInputSchemaField struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type schemaProperty struct {
	Type     string
	Required bool
}

func emptyWorkflowInputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func parseWorkflowInputSchema(schemaJSON []byte) (map[string]workflowInputSchemaField, error) {
	var schema map[string]any
	if len(schemaJSON) == 0 {
		schema = emptyWorkflowInputSchema()
	} else if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal workflow input schema: %w", err)
	}

	propsAny, _ := schema["properties"].(map[string]any)
	fields := make(map[string]workflowInputSchemaField, len(propsAny))
	for name, raw := range propsAny {
		if strings.Contains(name, ".") {
			return nil, fmt.Errorf("nested workflow input field %q is not supported", name)
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("workflow input field %q schema must be object", name)
		}
		typ, _ := prop["type"].(string)
		if typ == "" {
			typ = jsonSchemaStringType
		}
		field := workflowInputSchemaField{
			Name:     name,
			Type:     typ,
			Required: prop["required"] == true,
		}
		if description, ok := prop["description"].(string); ok {
			field.Description = description
		}
		fields[name] = field
	}
	return fields, nil
}

func marshalWorkflowInputSchema(fields map[string]workflowInputSchemaField) ([]byte, error) {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)

	properties := make(map[string]any, len(names))
	for _, name := range names {
		field := fields[name]
		prop := map[string]any{"type": field.Type}
		if field.Required {
			prop["required"] = true
		}
		if field.Description != "" {
			prop["description"] = field.Description
		}
		properties[name] = prop
	}

	return json.Marshal(map[string]any{
		"type":       "object",
		"properties": properties,
	})
}

func upsertWorkflowInputSchemaField(schemaJSON []byte, req InputSchemaFieldRequest) ([]byte, error) {
	if strings.Contains(req.Name, ".") {
		return nil, fmt.Errorf("nested workflow input field %q is not supported", req.Name)
	}
	fields, err := parseWorkflowInputSchema(schemaJSON)
	if err != nil {
		return nil, err
	}
	fields[req.Name] = workflowInputSchemaField(req)
	return marshalWorkflowInputSchema(fields)
}

func deleteWorkflowInputSchemaField(schemaJSON []byte, name string) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("workflow input field name is required")
	}
	if strings.Contains(name, ".") {
		return nil, fmt.Errorf("nested workflow input field %q is not supported", name)
	}
	fields, err := parseWorkflowInputSchema(schemaJSON)
	if err != nil {
		return nil, err
	}
	if _, ok := fields[name]; !ok {
		return nil, fmt.Errorf("workflow input field %q is not declared", name)
	}
	delete(fields, name)
	return marshalWorkflowInputSchema(fields)
}

func parseTopLevelProperties(schemaJSON []byte) (map[string]schemaProperty, error) {
	if len(schemaJSON) == 0 {
		return map[string]schemaProperty{}, nil
	}

	var schema map[string]any
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal schema: %w", err)
	}

	propsAny, ok := schema["properties"].(map[string]any)
	if !ok {
		return map[string]schemaProperty{}, nil
	}

	props := make(map[string]schemaProperty, len(propsAny))
	for name, raw := range propsAny {
		if strings.Contains(name, ".") {
			return nil, fmt.Errorf("nested schema property %q is not supported", name)
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("schema property %q must be object", name)
		}
		typ, _ := prop["type"].(string)
		if typ == "" {
			typ = jsonSchemaStringType
		}
		props[name] = schemaProperty{
			Type:     typ,
			Required: prop["required"] == true,
		}
	}
	return props, nil
}

func validateStepInputsFilled(steps []db.ListEnrichedStepsByVersionIDRow, deps []db.WorkflowStepDependency, workflowInputSchema []byte) []dagpkg.ValidationError {
	workflowProps, err := parseTopLevelProperties(workflowInputSchema)
	if err != nil {
		return []dagpkg.ValidationError{invalidMappingError(0, err.Error())}
	}

	stepByID := make(map[int32]db.ListEnrichedStepsByVersionIDRow, len(steps))
	predecessors := buildStepPredecessors(deps)
	for _, step := range steps {
		stepByID[step.ID] = step
	}

	var errs []dagpkg.ValidationError
	for _, step := range steps {
		if step.StepType != string(dagpkg.StepTypeTask) {
			continue
		}
		errs = append(errs, validateOneStepInputs(step, stepByID, predecessors[step.ID], workflowProps)...)
	}
	return errs
}

func validateOneStepInputs(
	step db.ListEnrichedStepsByVersionIDRow,
	stepByID map[int32]db.ListEnrichedStepsByVersionIDRow,
	predecessors map[int32]struct{},
	workflowProps map[string]schemaProperty,
) []dagpkg.ValidationError {
	inputProps, err := parseTopLevelProperties(step.InputSchema)
	if err != nil {
		return []dagpkg.ValidationError{invalidMappingError(step.ID, err.Error())}
	}
	mapping, err := parseInputMapping(step.InputMapping)
	if err != nil {
		return []dagpkg.ValidationError{invalidMappingError(step.ID, err.Error())}
	}

	mappedTargets := make(map[string]struct{}, len(mapping))
	var errs []dagpkg.ValidationError
	for _, entry := range mapping {
		target := strings.TrimSpace(entry.Target)
		source := strings.TrimSpace(entry.Source)
		if target == "" {
			errs = append(errs, invalidMappingError(step.ID, "mapping target is required"))
			continue
		}
		if strings.Contains(target, ".") {
			errs = append(errs, invalidMappingError(step.ID, fmt.Sprintf("nested mapping target is not supported: %s", target)))
			continue
		}
		if _, ok := mappedTargets[target]; ok {
			errs = append(errs, invalidMappingError(step.ID, fmt.Sprintf("duplicate mapping target: %s", target)))
			continue
		}
		mappedTargets[target] = struct{}{}

		targetProp, ok := inputProps[target]
		if !ok {
			errs = append(errs, invalidMappingError(step.ID, fmt.Sprintf("mapping target %q is not declared in step input_schema", target)))
			continue
		}
		if err := validateMappingSourceProperty(source, targetProp, workflowProps, stepByID, predecessors); err != nil {
			errs = append(errs, invalidMappingError(step.ID, err.Error()))
		}
	}

	for name, prop := range inputProps {
		if !prop.Required {
			continue
		}
		if _, ok := mappedTargets[name]; !ok {
			errs = append(errs, invalidMappingError(step.ID, fmt.Sprintf("required input %q has no mapping", name)))
		}
	}
	return errs
}

func validateMappingSourceProperty(
	source string,
	target schemaProperty,
	workflowProps map[string]schemaProperty,
	stepByID map[int32]db.ListEnrichedStepsByVersionIDRow,
	predecessors map[int32]struct{},
) error {
	if strings.HasPrefix(source, messageValueSourcePrefix) {
		field := strings.TrimPrefix(source, messageValueSourcePrefix)
		if field == "" || strings.Contains(field, ".") {
			return fmt.Errorf("nested message source is not supported: %s", source)
		}
		sourceProp, ok := workflowProps[field]
		if !ok {
			return fmt.Errorf("workflow input %q is not declared", field)
		}
		return validateSourceCompatibility(target, sourceProp)
	}

	if strings.HasPrefix(source, stepOutputSourcePrefix) {
		return validateStepOutputSourceProperty(source, target, stepByID, predecessors)
	}

	if strings.HasPrefix(source, unknownMappingSourceStart) {
		return fmt.Errorf("unknown mapping source: %s", source)
	}
	return validateStaticLiteral(source, target.Type)
}

func validateStepOutputSourceProperty(
	source string,
	target schemaProperty,
	stepByID map[int32]db.ListEnrichedStepsByVersionIDRow,
	predecessors map[int32]struct{},
) error {
	sourceStepID, field, err := parseStepOutputSource(source)
	if err != nil {
		return err
	}
	if _, ok := predecessors[sourceStepID]; !ok {
		return fmt.Errorf("step %d is not a predecessor", sourceStepID)
	}
	sourceStep, ok := stepByID[sourceStepID]
	if !ok {
		return fmt.Errorf("step %d is not found", sourceStepID)
	}
	outputProps, err := parseTopLevelProperties(sourceStep.OutputSchema)
	if err != nil {
		return err
	}
	sourceProp, ok := outputProps[field]
	if !ok {
		return fmt.Errorf("output %q is not declared on step %d", field, sourceStepID)
	}
	return validateSourceCompatibility(target, sourceProp)
}

func parseStepOutputSource(source string) (int32, string, error) {
	rest := strings.TrimPrefix(source, stepOutputSourcePrefix)
	parts := strings.SplitN(rest, ".", 3)
	if len(parts) != 3 || parts[1] != "output" {
		return 0, "", fmt.Errorf("invalid step output source: %s", source)
	}
	stepID64, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return 0, "", fmt.Errorf("invalid step id in source: %s", source)
	}
	field := parts[2]
	if field == "" || strings.Contains(field, ".") {
		return 0, "", fmt.Errorf("nested step output source is not supported: %s", source)
	}
	return int32(stepID64), field, nil
}

func validateSourceCompatibility(target, source schemaProperty) error {
	if target.Type != "" && source.Type != "" && target.Type != source.Type {
		return fmt.Errorf("mapping type mismatch: target %s, source %s", target.Type, source.Type)
	}
	if target.Required && !source.Required {
		return fmt.Errorf("required target is mapped from optional source")
	}
	return nil
}

func validateStaticLiteral(source, targetType string) error {
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("static mapping source is empty")
	}
	switch targetType {
	case "", "string":
		return nil
	case "number":
		if _, err := strconv.ParseFloat(source, 64); err != nil {
			return fmt.Errorf("static source is not a number")
		}
	case "integer":
		if _, err := strconv.ParseInt(source, 10, 64); err != nil {
			return fmt.Errorf("static source is not an integer")
		}
	case "boolean":
		if _, err := strconv.ParseBool(source); err != nil {
			return fmt.Errorf("static source is not a boolean")
		}
	default:
		return fmt.Errorf("static source is not supported for %s input", targetType)
	}
	return nil
}

func buildStepPredecessors(deps []db.WorkflowStepDependency) map[int32]map[int32]struct{} {
	parents := make(map[int32][]int32)
	for _, dep := range deps {
		parents[dep.StepID] = append(parents[dep.StepID], dep.DependsOnStepID)
	}

	predecessors := make(map[int32]map[int32]struct{}, len(parents))
	for stepID := range parents {
		visited := make(map[int32]struct{})
		queue := append([]int32(nil), parents[stepID]...)
		for len(queue) > 0 {
			parentID := queue[0]
			queue = queue[1:]
			if _, ok := visited[parentID]; ok {
				continue
			}
			visited[parentID] = struct{}{}
			queue = append(queue, parents[parentID]...)
		}
		predecessors[stepID] = visited
	}
	return predecessors
}

func invalidMappingError(stepID int32, message string) dagpkg.ValidationError {
	return dagpkg.ValidationError{
		Type:    "invalid_mapping",
		StepID:  stepID,
		Message: message,
	}
}
