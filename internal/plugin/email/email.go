package email

import (
	"cactus/internal/plugin"
)

type EmailPlugin struct {
	plugin.CorePlugin
	Schema EmailSchema
	Slug   string
}

func New() *EmailPlugin {
	return &EmailPlugin{
		Schema: EmailSchema{},
		Slug:   "email",
	}
}

func (ep *EmailPlugin) New() plugin.Plugin {
	return New()
}

func (ep *EmailPlugin) GetSchema() any {
	return &EmailSchema{}
}
