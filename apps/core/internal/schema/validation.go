package schema

import (
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func ValidateRaw(schemaJSON []byte, payloadJSON []byte) error {
	var payload any
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	return ValidateValue(schemaJSON, payload)
}

func ValidateValue(schemaJSON []byte, payload any) error {
	if len(schemaJSON) == 0 {
		return nil
	}

	schemaAny, err := NormalizeRequired(schemaJSON)
	if err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schemaAny); err != nil {
		return fmt.Errorf("compile schema resource: %w", err)
	}

	compiled, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	if err := compiled.Validate(payload); err != nil {
		return fmt.Errorf("validate payload: %w", err)
	}
	return nil
}
