package schema

import (
	"encoding/json"
	"testing"
)

func TestNormalizeRequiredMovesPropertyRequiredToParent(t *testing.T) {
	normalized, err := NormalizeRequired([]byte(`{
		"type": "object",
		"properties": {
			"outer": {
				"type": "object",
				"required": true,
				"properties": {
					"inner": {"type": "string", "required": true}
				}
			}
		}
	}`))
	if err != nil {
		t.Fatalf("NormalizeRequired error: %v", err)
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal normalized schema: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("unmarshal normalized schema: %v", err)
	}

	assertRequired(t, schema, "outer")

	props := schema["properties"].(map[string]any)
	outer := props["outer"].(map[string]any)
	if _, exists := outer["required"].(bool); exists {
		t.Fatalf("outer property must not keep dialect required bool: %#v", outer)
	}
	assertRequired(t, outer, "inner")
}

func assertRequired(t *testing.T, schema map[string]any, field string) {
	t.Helper()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("expected required array on schema: %#v", schema)
	}
	for _, item := range required {
		if item == field {
			return
		}
	}
	t.Fatalf("expected %q in required array %#v", field, required)
}
