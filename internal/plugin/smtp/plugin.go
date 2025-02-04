package smtp

import (
	"cactus/internal/pkg/pipeline"
	"cactus/internal/plugin"
)

type Plugin struct {
	plugin.CorePlugin
	Schema Schema
	Slug   string
}

func New() *Plugin {
	return &Plugin{
		Schema: Schema{},
		Slug:   "email",
	}
}

func (p *Plugin) New() plugin.Plugin {
	return New()
}

func (p *Plugin) GetSchema() any {
	return &Schema{}
}

func (p *Plugin) ExtendPipeline(steps []pipeline.Step) ([]pipeline.Step, error) {
	steps = append(steps, pipeline.Step{
		Step: 1,
		Name: "SMTP воркер",
	})

	return steps, nil
}
