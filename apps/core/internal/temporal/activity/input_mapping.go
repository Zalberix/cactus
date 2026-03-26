package activity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

// ResolveInput формирует входные данные шага по его input_mapping.
// Источники (per D-05 from CONTEXT):
//   - $.message.value.{field} — из messageValue (JSON payload клиента)
//   - $.steps.{id}.output.{field} — из результатов предыдущих шагов
//
// messageValue — raw JSON bytes из message.value.
// stepOutputs — map[stepID]map[string]any с output предыдущих шагов.
func ResolveInput(mapping []temporaltypes.MappingEntry, messageValue []byte, stepOutputs map[int32]map[string]any) (map[string]any, error) {
	if len(mapping) == 0 {
		return nil, nil
	}

	// Десериализуем message value
	var msgValue map[string]any
	if err := json.Unmarshal(messageValue, &msgValue); err != nil {
		return nil, fmt.Errorf("unmarshal message value: %w", err)
	}

	result := make(map[string]any, len(mapping))
	for _, entry := range mapping {
		val, err := resolveSource(entry.Source, msgValue, stepOutputs)
		if err != nil {
			return nil, fmt.Errorf("resolve %s -> %s: %w", entry.Source, entry.Target, err)
		}
		result[entry.Target] = val
	}
	return result, nil
}

func resolveSource(source string, msgValue map[string]any, stepOutputs map[int32]map[string]any) (any, error) {
	switch {
	case strings.HasPrefix(source, "$.message.value."):
		field := strings.TrimPrefix(source, "$.message.value.")
		val, ok := msgValue[field]
		if !ok {
			return nil, fmt.Errorf("field %q not found in message value", field)
		}
		return val, nil

	case strings.HasPrefix(source, "$.steps."):
		// Format: $.steps.{id}.output.{field}
		rest := strings.TrimPrefix(source, "$.steps.")
		parts := strings.SplitN(rest, ".", 3) // id, "output", field
		if len(parts) < 3 || parts[1] != "output" {
			return nil, fmt.Errorf("invalid step source format: %s", source)
		}
		stepID64, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid step ID in source: %s", source)
		}
		stepID := int32(stepID64)
		field := parts[2]

		outputs, ok := stepOutputs[stepID]
		if !ok {
			return nil, fmt.Errorf("step %d output not available", stepID)
		}
		val, ok := outputs[field]
		if !ok {
			return nil, fmt.Errorf("field %q not found in step %d output", field, stepID)
		}
		return val, nil

	default:
		return nil, fmt.Errorf("unknown source format: %s", source)
	}
}
