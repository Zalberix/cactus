package main

import (
	"testing"
	"testing/fstest"
)

func TestTemplateRendererUsesFieldsAndDefaults(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/welcome.html": {
				Data: []byte(`<h1>{{fields.name|"Guest"}}</h1><p>{{fields.city|"Yekaterinburg"}}</p>`),
			},
		},
	}

	body, err := renderer.Render("welcome", map[string]any{"name": "Ignat"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := `<h1>Ignat</h1><p>Yekaterinburg</p>`
	if body != want {
		t.Fatalf("Render() = %q, want %q", body, want)
	}
}

func TestTemplateRendererEscapesInsertedValues(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/welcome.html": {
				Data: []byte(`<p>{{fields.name|"Guest"}}</p>`),
			},
		},
	}

	body, err := renderer.Render("welcome", map[string]any{"name": `<script>alert(1)</script>`})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := `<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>`
	if body != want {
		t.Fatalf("Render() = %q, want %q", body, want)
	}
}

func TestTemplateRendererAllowsRenderedTextContainingFieldsPlaceholder(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/welcome.html": {
				Data: []byte(`<p>{{fields.name|"Guest"}}</p>`),
			},
		},
	}

	body, err := renderer.Render("welcome", map[string]any{"name": `{{fields.name}}`})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := `<p>{{fields.name}}</p>`
	if body != want {
		t.Fatalf("Render() = %q, want %q", body, want)
	}
}

func TestTemplateRendererRejectsPathTraversal(t *testing.T) {
	renderer := TemplateRenderer{templates: fstest.MapFS{}}

	_, err := renderer.Render("../secret", map[string]any{})
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestTemplateRendererRejectsMalformedFieldsPlaceholder(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/bad.html": {
				Data: []byte(`<p>{{fields.name}}</p>`),
			},
		},
	}

	_, err := renderer.Render("bad", map[string]any{"name": "Ignat"})
	if err == nil {
		t.Fatal("expected malformed placeholder error")
	}
}

func TestTemplateRendererRejectsMalformedFieldsPlaceholderWithWhitespace(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/bad.html": {
				Data: []byte(`<p>{{ fields.name }}</p>`),
			},
		},
	}

	_, err := renderer.Render("bad", map[string]any{"name": "Ignat"})
	if err == nil {
		t.Fatal("expected malformed placeholder error")
	}
}

func TestTemplateRendererRejectsMalformedFieldsPlaceholderWithWhitespaceBeforeDot(t *testing.T) {
	renderer := TemplateRenderer{
		templates: fstest.MapFS{
			"templates/bad.html": {
				Data: []byte(`<p>{{ fields .name }}</p>`),
			},
		},
	}

	_, err := renderer.Render("bad", map[string]any{"name": "Ignat"})
	if err == nil {
		t.Fatal("expected malformed placeholder error")
	}
}
