package schema

import (
	"encoding/json"
	"fmt"
	"sort"
)

// NormalizeRequired converts the Cactus schema dialect to standard JSON Schema.
// The dialect stores required on each property; JSON Schema expects a required
// string array on the containing object.
func NormalizeRequired(schemaJSON []byte) (any, error) {
	var value any
	if err := json.Unmarshal(schemaJSON, &value); err != nil {
		return nil, fmt.Errorf("unmarshal schema: %w", err)
	}
	return normalizeValue(value), nil
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeObject(typed)
	case []any:
		for i, item := range typed {
			typed[i] = normalizeValue(item)
		}
		return typed
	default:
		return value
	}
}

func normalizeObject(object map[string]any) map[string]any {
	for key, value := range object {
		if key == "properties" {
			continue
		}
		object[key] = normalizeValue(value)
	}

	propertiesAny, ok := object["properties"]
	if !ok {
		return object
	}
	properties, ok := propertiesAny.(map[string]any)
	if !ok {
		return object
	}

	required := make([]string, 0)
	for name, propertyAny := range properties {
		property, ok := propertyAny.(map[string]any)
		if !ok {
			continue
		}
		if isRequired, ok := property["required"].(bool); ok {
			if isRequired {
				required = append(required, name)
			}
			delete(property, "required")
		}
		properties[name] = normalizeValue(property)
	}
	if len(required) > 0 {
		sort.Strings(required)
		requiredAny := make([]any, len(required))
		for i, field := range required {
			requiredAny[i] = field
		}
		object["required"] = requiredAny
	}
	return object
}
