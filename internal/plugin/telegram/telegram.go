package telegram

import (
	"cactus/internal/plugin"
)

type TelegramPlugin struct {
	plugin.CorePlugin
	Schema TelegramSchema
	Slug   string
}

func New() *TelegramPlugin {
	return &TelegramPlugin{
		Schema: TelegramSchema{},
		Slug:   "telegram",
	}
}

func (ep *TelegramPlugin) New() plugin.Plugin {
	return New()
}

func (ep *TelegramPlugin) GetSchema() any {
	return &TelegramSchema{}
}
