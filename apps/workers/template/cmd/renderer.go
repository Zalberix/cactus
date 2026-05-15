package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	fieldsPlaceholderPattern          = regexp.MustCompile(`{{\s*fields\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*"((?:\\.|[^"\\])*)"\s*}}`)
	malformedFieldsPlaceholderPattern = regexp.MustCompile(`{{\s*fields\s*\.`)
)

type TemplateRenderer struct {
	templates fs.FS
}

func (r TemplateRenderer) Render(templateName string, fields map[string]any) (string, error) {
	fileName, err := templateFileName(templateName)
	if err != nil {
		return "", err
	}

	data, err := fs.ReadFile(r.templates, path.Join("templates", fileName))
	if err != nil {
		return "", fmt.Errorf("read template %q: %w", templateName, err)
	}

	templateBody := string(data)
	if containsMalformedFieldsPlaceholder(templateBody) {
		return "", fmt.Errorf("template %q contains malformed fields placeholder", templateName)
	}

	body := fieldsPlaceholderPattern.ReplaceAllStringFunc(templateBody, func(token string) string {
		matches := fieldsPlaceholderPattern.FindStringSubmatch(token)
		fieldName := matches[1]
		defaultValue := unquoteDefault(matches[2])

		value, exists := fields[fieldName]
		if !exists {
			return html.EscapeString(defaultValue)
		}
		return html.EscapeString(formatFieldValue(value))
	})

	return body, nil
}

func containsMalformedFieldsPlaceholder(templateBody string) bool {
	withoutValidPlaceholders := fieldsPlaceholderPattern.ReplaceAllString(templateBody, "")
	return malformedFieldsPlaceholderPattern.MatchString(withoutValidPlaceholders)
}

func templateFileName(templateName string) (string, error) {
	name := strings.TrimSpace(templateName)
	if name == "" {
		return "", fmt.Errorf("template is required")
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\`) || strings.Contains(name, "..") {
		return "", fmt.Errorf("template name must be a file basename")
	}
	if path.Ext(name) == "" {
		name += ".html"
	}
	if path.Ext(name) != ".html" {
		return "", fmt.Errorf("template must use .html extension")
	}
	return name, nil
}

func unquoteDefault(raw string) string {
	value, err := strconv.Unquote(`"` + raw + `"`)
	if err != nil {
		return raw
	}
	return value
}

func formatFieldValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(data)
	}
}
