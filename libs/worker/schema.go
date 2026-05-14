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

// SchemaBuilder collects object schema properties through a fluent API.
type SchemaBuilder struct {
	fields []SchemaProperty
}

// FieldBuilder modifies a field created by SchemaBuilder.
type FieldBuilder struct {
	builder *SchemaBuilder
	index   int
}

// SettingsSchema builds a worker settings object schema.
func SettingsSchema(build func(*SchemaBuilder)) json.RawMessage {
	return buildObjectSchema(build)
}

// InputSchema builds a worker input object schema.
func InputSchema(build func(*SchemaBuilder)) json.RawMessage {
	return buildObjectSchema(build)
}

// OutputSchema builds a worker output object schema.
func OutputSchema(build func(*SchemaBuilder)) json.RawMessage {
	return buildObjectSchema(build)
}

// ObjectSchema builds an object schema with property-level required flags.
func ObjectSchema(fields ...SchemaProperty) json.RawMessage {
	return marshalObjectSchema(fields)
}

// String creates a string property.
func (builder *SchemaBuilder) String(name string) *FieldBuilder {
	return builder.field(name, "string")
}

// Integer creates an integer property.
func (builder *SchemaBuilder) Integer(name string) *FieldBuilder {
	return builder.field(name, "integer")
}

// Number creates a number property.
func (builder *SchemaBuilder) Number(name string) *FieldBuilder {
	return builder.field(name, "number")
}

// Boolean creates a boolean property.
func (builder *SchemaBuilder) Boolean(name string) *FieldBuilder {
	return builder.field(name, "boolean")
}

// Object creates an object property.
func (builder *SchemaBuilder) Object(name string) *FieldBuilder {
	return builder.field(name, "object")
}

// Required marks a property as required in the worker schema dialect.
func (builder *FieldBuilder) Required() *FieldBuilder {
	builder.field().Required = true
	return builder
}

// Description adds help text to a property.
func (builder *FieldBuilder) Description(description string) *FieldBuilder {
	builder.field().Description = description
	return builder
}

// Enum constrains a string-like property to a known option set.
func (builder *FieldBuilder) Enum(values ...string) *FieldBuilder {
	builder.field().Enum = append([]string(nil), values...)
	return builder
}

func (builder *SchemaBuilder) field(name string, fieldType string) *FieldBuilder {
	builder.fields = append(builder.fields, SchemaProperty{
		Name: name,
		Field: SchemaField{
			Type: fieldType,
		},
	})
	return &FieldBuilder{
		builder: builder,
		index:   len(builder.fields) - 1,
	}
}

func (builder *FieldBuilder) field() *SchemaField {
	return &builder.builder.fields[builder.index].Field
}

func buildObjectSchema(build func(*SchemaBuilder)) json.RawMessage {
	builder := &SchemaBuilder{}
	build(builder)
	return marshalObjectSchema(builder.fields)
}

func marshalObjectSchema(fields []SchemaProperty) json.RawMessage {
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
