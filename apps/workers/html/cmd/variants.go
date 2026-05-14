package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	"github.com/zalberix/cactus/libs/worker"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

type htmlRenderer interface {
	Render(templateName string, fields map[string]any) (string, error)
}

type HTMLHandler struct {
	renderer htmlRenderer
}

func (h HTMLHandler) Handle(ctx context.Context, task worker.TaskMessage) (worker.Result, error) {
	_ = ctx

	templateName, _ := task.Input["template"].(string)
	if templateName == "" {
		return worker.Result{}, fmt.Errorf("missing required field: template")
	}

	fields, err := fieldsFromInput(task.Input["fields"])
	if err != nil {
		return worker.Result{}, err
	}

	body, err := h.renderer.Render(templateName, fields)
	if err != nil {
		return worker.Result{}, err
	}

	return worker.Result{
		Success: true,
		Output: map[string]any{
			"body": body,
		},
	}, nil
}

func fieldsFromInput(value any) (map[string]any, error) {
	switch typed := value.(type) {
	case nil:
		return nil, fmt.Errorf("missing required field: fields")
	case map[string]any:
		return typed, nil
	case map[string]string:
		result := make(map[string]any, len(typed))
		for key, fieldValue := range typed {
			result[key] = fieldValue
		}
		return result, nil
	case string:
		return parseFieldsJSON([]byte(typed))
	case []byte:
		return parseFieldsJSON(typed)
	case json.RawMessage:
		return parseFieldsJSON(typed)
	default:
		return nil, fmt.Errorf("fields must be a JSON object")
	}
}

func parseFieldsJSON(data []byte) (map[string]any, error) {
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("parse fields JSON: %w", err)
	}
	if fields == nil {
		return nil, fmt.Errorf("fields must be a JSON object")
	}
	return fields, nil
}

func htmlVariants() map[string]worker.Variant {
	return map[string]worker.Variant{
		"basic": {
			Name:     "basic",
			Manifest: basicHTMLManifest(),
			Handler:  HTMLHandler{renderer: TemplateRenderer{templates: embeddedTemplates}},
		},
	}
}

func basicHTMLManifest() worker.ManifestSpec {
	return worker.Manifest().
		Kind("html-template", "HTML Template").
		Type("html", "HTML Generation").
		SettingsSchema(func(sb *worker.SchemaBuilder) {}).
		InputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("template").Required().Description("Template file name without path")
			sb.Object("fields").Required().Description("Template values as a JSON object")
		}).
		OutputSchema(func(sb *worker.SchemaBuilder) {
			sb.String("body").Description("Rendered HTML body")
		}).
		Build()
}
