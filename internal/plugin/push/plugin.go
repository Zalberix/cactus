package push

import (
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
		Slug:   "push",
	}
}

func (ep *Plugin) New() plugin.Plugin {
	return New()
}

func (ep *Plugin) GetSchema() any {
	return &Schema{}
}
