package worker

import "encoding/json"

// SchemaField describes one property in the worker schema dialect.
type SchemaField struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// SchemaProperty binds a field definition to its property name.
type SchemaProperty struct {
	Name  string
	Field SchemaField
}

// ObjectSchema builds an object schema with property-level required flags.
func ObjectSchema(fields ...SchemaProperty) json.RawMessage {
	properties := make(map[string]SchemaField, len(fields))
	for _, field := range fields {
		properties[field.Name] = field.Field
	}

	data, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": properties,
	})
	if err != nil {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}
	return data
}

// Field assigns a schema field to a property name.
func Field(name string, field SchemaField) SchemaProperty {
	return SchemaProperty{Name: name, Field: field}
}

// FieldOption modifies a schema field.
type FieldOption func(*SchemaField)

// StringField creates a string property.
func StringField(options ...FieldOption) SchemaField {
	return applyFieldOptions(SchemaField{Type: "string"}, options...)
}

// IntegerField creates an integer property.
func IntegerField(options ...FieldOption) SchemaField {
	return applyFieldOptions(SchemaField{Type: "integer"}, options...)
}

// NumberField creates a number property.
func NumberField(options ...FieldOption) SchemaField {
	return applyFieldOptions(SchemaField{Type: "number"}, options...)
}

// BooleanField creates a boolean property.
func BooleanField(options ...FieldOption) SchemaField {
	return applyFieldOptions(SchemaField{Type: "boolean"}, options...)
}

// Required marks a property as required in the worker schema dialect.
func Required() FieldOption {
	return func(field *SchemaField) {
		field.Required = true
	}
}

// Description adds help text to a property.
func Description(description string) FieldOption {
	return func(field *SchemaField) {
		field.Description = description
	}
}

// Enum constrains a string-like property to a known option set.
func Enum(values ...string) FieldOption {
	return func(field *SchemaField) {
		field.Enum = append([]string(nil), values...)
	}
}

func applyFieldOptions(field SchemaField, options ...FieldOption) SchemaField {
	for _, option := range options {
		option(&field)
	}
	return field
}
