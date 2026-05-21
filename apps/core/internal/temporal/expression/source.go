package expression

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	temporaltypes "github.com/zalberix/cactus/apps/core/internal/temporal"
)

// ResolveInput builds task input from mapping expressions.
func ResolveInput(mapping []temporaltypes.MappingEntry, messageValue []byte, stepOutputs map[int32]map[string]any) (map[string]any, error) {
	if len(mapping) == 0 {
		return nil, nil
	}

	result := make(map[string]any, len(mapping))
	seenTargets := make(map[string]struct{}, len(mapping))
	for _, entry := range mapping {
		if strings.Contains(entry.Target, ".") {
			return nil, fmt.Errorf("nested target path is not supported: %s", entry.Target)
		}
		if _, ok := seenTargets[entry.Target]; ok {
			return nil, fmt.Errorf("duplicate mapping target %q", entry.Target)
		}
		seenTargets[entry.Target] = struct{}{}

		val, err := ResolveSource(entry.Source, messageValue, stepOutputs)
		if err != nil {
			return nil, fmt.Errorf("resolve %s -> %s: %w", entry.Source, entry.Target, err)
		}
		result[entry.Target] = val
	}
	return result, nil
}

// ResolveSource resolves a single expression or static scalar string.
func ResolveSource(source string, messageValue []byte, stepOutputs map[int32]map[string]any) (any, error) {
	value, exists, err := ResolveSourcePresence(source, messageValue, stepOutputs)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("source not found: %s", source)
	}
	return value, nil
}

// ResolveSourcePresence resolves a source and returns exists=false for missing paths.
func ResolveSourcePresence(source string, messageValue []byte, stepOutputs map[int32]map[string]any) (any, bool, error) {
	source = strings.TrimSpace(source)
	switch {
	case source == "$.message.value":
		value, err := decodeJSONValue(messageValue)
		return value, err == nil, err
	case strings.HasPrefix(source, "$.message.value."):
		root, err := decodeJSONValue(messageValue)
		if err != nil {
			return nil, false, fmt.Errorf("unmarshal message value: %w", err)
		}
		return resolvePath(root, strings.TrimPrefix(source, "$.message.value."))
	case strings.HasPrefix(source, "$.steps."):
		return resolveStepSource(source, stepOutputs)
	case strings.HasPrefix(source, "$."):
		return nil, false, fmt.Errorf("unknown source format: %s", source)
	default:
		if source == "" {
			return nil, false, fmt.Errorf("empty static source")
		}
		return source, true, nil
	}
}

func decodeJSONValue(raw []byte) (any, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func resolveStepSource(source string, stepOutputs map[int32]map[string]any) (any, bool, error) {
	rest := strings.TrimPrefix(source, "$.steps.")
	parts := strings.SplitN(rest, ".", 3)
	if len(parts) < 2 || parts[1] != "output" {
		return nil, false, fmt.Errorf("invalid step source format: %s", source)
	}
	stepID64, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return nil, false, fmt.Errorf("invalid step ID in source: %s", source)
	}
	outputs, ok := stepOutputs[int32(stepID64)]
	if !ok {
		return nil, false, nil
	}
	if len(parts) == 2 {
		return outputs, true, nil
	}
	return resolvePath(outputs, parts[2])
}

func resolvePath(root any, path string) (any, bool, error) {
	if strings.TrimSpace(path) == "" {
		return nil, false, fmt.Errorf("empty source path")
	}
	current := root
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, false, fmt.Errorf("invalid source path: %s", path)
		}
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false, nil
		}
		next, ok := obj[part]
		if !ok {
			return nil, false, nil
		}
		current = next
	}
	return current, true, nil
}
