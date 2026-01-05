package telegram

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
		Slug:   "telegram",
	}
}

func (ep *Plugin) New() plugin.Plugin {
	return New()
}

func (ep *Plugin) GetSchema() any {
	return &Schema{}
}

func (ep *Plugin) ExtendPipeline(steps []pipeline.Step) ([]pipeline.Step, error) {
	_ = ep
	steps = append(steps, pipeline.Step{
		Step: 1,
		Name: "Телеграм воркер",
	})

	return steps, nil
}
