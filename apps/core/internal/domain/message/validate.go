package message

import (
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/zalberix/cactus/apps/core/internal/http/response"
	schemadialect "github.com/zalberix/cactus/apps/core/internal/schema"
)

// ValidatePayload валидирует payload по JSON Schema из workflow_input_schema.schema_json.
// schemaJSON — JSONB из поля workflow_input_schema.schema_json.
// payload — значение из SendMessageRequest.Value.
// Возвращает nil если валидация пройдена, или slice ErrorDetail с описанием ошибок полей.
func ValidatePayload(schemaJSON []byte, payload map[string]any) []response.ErrorDetail {
	if len(schemaJSON) == 0 {
		// Нет схемы валидации — пропускаем
		return nil
	}

	// Десериализуем JSON Schema из JSONB
	schemaAny, err := schemadialect.NormalizeRequired(schemaJSON)
	if err != nil {
		return []response.ErrorDetail{{
			Message: fmt.Sprintf("invalid input_schema schema: %v", err),
		}}
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schemaAny); err != nil {
		return []response.ErrorDetail{{
			Message: fmt.Sprintf("failed to compile validation schema: %v", err),
		}}
	}

	sch, err := compiler.Compile("schema.json")
	if err != nil {
		return []response.ErrorDetail{{
			Message: fmt.Sprintf("failed to compile validation schema: %v", err),
		}}
	}

	if err := sch.Validate(payload); err != nil {
		// Извлекаем детали ошибок
		var ve *jsonschema.ValidationError
		if errors.As(err, &ve) {
			return extractValidationErrors(ve)
		}
		return []response.ErrorDetail{{
			Message: fmt.Sprintf("validation failed: %v", err),
		}}
	}

	return nil
}

// extractValidationErrors рекурсивно извлекает ошибки валидации.
// InstanceLocation в jsonschema/v6 — это []string (path segments).
func extractValidationErrors(ve *jsonschema.ValidationError) []response.ErrorDetail {
	if len(ve.Causes) == 0 {
		field := ""
		if len(ve.InstanceLocation) > 0 {
			field = strings.Join(ve.InstanceLocation, ".")
		}
		return []response.ErrorDetail{{
			Field:   field,
			Message: ve.Error(),
		}}
	}
	var details []response.ErrorDetail
	for _, cause := range ve.Causes {
		details = append(details, extractValidationErrors(cause)...)
	}
	return details
}
