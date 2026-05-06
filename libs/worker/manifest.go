package worker

// ManifestBuilder collects worker manifest metadata through a fluent API.
type ManifestBuilder struct {
	manifest ManifestSpec
}

// Manifest starts a worker manifest declaration.
func Manifest() *ManifestBuilder {
	return &ManifestBuilder{}
}

// Kind sets the worker kind and display name.
func (builder *ManifestBuilder) Kind(kind string, name string) *ManifestBuilder {
	builder.manifest.Kind = kind
	builder.manifest.NameKind = name
	return builder
}

// Type sets the worker type and display name.
func (builder *ManifestBuilder) Type(workerType string, name string) *ManifestBuilder {
	builder.manifest.Type = workerType
	builder.manifest.NameType = name
	return builder
}

// SettingsSchema builds and assigns the worker settings schema.
func (builder *ManifestBuilder) SettingsSchema(build func(*SchemaBuilder)) *ManifestBuilder {
	builder.manifest.SettingsSchema = SettingsSchema(build)
	return builder
}

// InputSchema builds and assigns the worker input schema.
func (builder *ManifestBuilder) InputSchema(build func(*SchemaBuilder)) *ManifestBuilder {
	builder.manifest.InputSchema = InputSchema(build)
	return builder
}

// OutputSchema builds and assigns the worker output schema.
func (builder *ManifestBuilder) OutputSchema(build func(*SchemaBuilder)) *ManifestBuilder {
	builder.manifest.OutputSchema = OutputSchema(build)
	return builder
}

// Build returns the final manifest value.
func (builder *ManifestBuilder) Build() ManifestSpec {
	return builder.manifest
}
