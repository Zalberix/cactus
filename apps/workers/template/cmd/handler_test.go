package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/zalberix/cactus/libs/worker"
)

type fakeRenderer struct {
	templateName string
	fields       map[string]any
	body         string
}

func (r *fakeRenderer) Render(templateName string, fields map[string]any) (string, error) {
	r.templateName = templateName
	r.fields = fields
	return r.body, nil
}

func TestHTMLHandlerReturnsRenderedBody(t *testing.T) {
	renderer := &fakeRenderer{body: "<h1>Hello</h1>"}
	handler := HTMLHandler{renderer: renderer}

	result, err := handler.Handle(context.Background(), worker.TaskMessage{
		Input: map[string]any{
			"template": "welcome",
			"fields": map[string]any{
				"name": "Ignat",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success result: %#v", result)
	}
	if result.Output["body"] != "<h1>Hello</h1>" {
		t.Fatalf("expected rendered body, got %#v", result.Output)
	}
	if renderer.templateName != "welcome" {
		t.Fatalf("expected template welcome, got %q", renderer.templateName)
	}
	if renderer.fields["name"] != "Ignat" {
		t.Fatalf("expected field name, got %#v", renderer.fields)
	}
}

func TestHTMLHandlerAcceptsFieldsJSONString(t *testing.T) {
	renderer := &fakeRenderer{body: "<p>OK</p>"}
	handler := HTMLHandler{renderer: renderer}

	_, err := handler.Handle(context.Background(), worker.TaskMessage{
		Input: map[string]any{
			"template": "welcome",
			"fields":   `{"name":"Ignat"}`,
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if renderer.fields["name"] != "Ignat" {
		t.Fatalf("expected parsed JSON fields, got %#v", renderer.fields)
	}
}

func TestHTMLHandlerRequiresTemplateAndFields(t *testing.T) {
	handler := HTMLHandler{renderer: &fakeRenderer{}}

	_, err := handler.Handle(context.Background(), worker.TaskMessage{
		Input: map[string]any{"template": "welcome"},
	})
	if err == nil {
		t.Fatal("expected missing fields error")
	}
}

func TestHTMLManifestSchemas(t *testing.T) {
	variants := templateVariants()
	basic, err := worker.SelectVariant(variants, "html")
	if err != nil {
		t.Fatalf("select html: %v", err)
	}

	if basic.Name != "html" {
		t.Fatalf("expected variant html, got %q", basic.Name)
	}
	if basic.Manifest.Kind != "template-html" {
		t.Fatalf("expected kind template-html, got %q", basic.Manifest.Kind)
	}
	if basic.Manifest.Type != "html" {
		t.Fatalf("expected type html, got %q", basic.Manifest.Type)
	}

	var input map[string]any
	if err := json.Unmarshal(basic.Manifest.InputSchema, &input); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}
	props := input["properties"].(map[string]any)
	fields := props["fields"].(map[string]any)
	if fields["type"] != "object" {
		t.Fatalf("expected fields object schema, got %#v", fields)
	}

	var output map[string]any
	if err := json.Unmarshal(basic.Manifest.OutputSchema, &output); err != nil {
		t.Fatalf("unmarshal output schema: %v", err)
	}
	outputProps := output["properties"].(map[string]any)
	if _, exists := outputProps["body"]; !exists {
		t.Fatalf("expected output body field, got %#v", outputProps)
	}
}
