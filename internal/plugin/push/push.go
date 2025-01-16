package push

import (
	"cactus/internal/plugin"
)

type PushPlugin struct {
	plugin.CorePlugin
	Schema PushSchema
	Slug   string
}

func New() *PushPlugin {
	return &PushPlugin{
		Schema: PushSchema{},
		Slug:   "push",
	}
}

func (ep *PushPlugin) New() plugin.Plugin {
	return New()
}

func (ep *PushPlugin) GetSchema() any {
	return &PushSchema{}
}
