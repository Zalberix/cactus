package message

import "testing"

func TestValidatePayloadTreatsPropertyRequiredAsRequired(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"properties": {
			"to": {"type": "string", "required": true},
			"subject": {"type": "string", "required": true},
			"body": {"type": "string"}
		}
	}`)

	errors := ValidatePayload(schema, map[string]any{
		"to": "user@example.com",
	})

	if len(errors) == 0 {
		t.Fatal("expected missing required subject to fail validation")
	}
}

func TestValidatePayloadAcceptsPropertyRequiredDialectWhenPresent(t *testing.T) {
	schema := []byte(`{
		"type": "object",
		"properties": {
			"to": {"type": "string", "required": true},
			"subject": {"type": "string", "required": true}
		}
	}`)

	errors := ValidatePayload(schema, map[string]any{
		"to":      "user@example.com",
		"subject": "Hello",
	})

	if len(errors) != 0 {
		t.Fatalf("expected valid payload, got errors: %#v", errors)
	}
}
