package telegram

import (
	"github.com/zalberix/cactus/apps/core/internal/plugin"
	"github.com/zalberix/cactus/libs/pipeline"
)

type Plugin struct {
	plugin.CorePlugin
	Schema Schema
	Slug   string
}

func NewFx() *Plugin {
	return New()
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
