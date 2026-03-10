package plugin

import (
	"github.com/zalberix/cactus/apps/core/internal/plugin"
	"github.com/zalberix/cactus/apps/core/internal/plugin/smtp"
	"github.com/zalberix/cactus/apps/core/internal/plugin/telegram"
	"go.uber.org/fx"
)

type Opts struct {
	fx.In
	SMTP     *smtp.Plugin
	Telegram *telegram.Plugin
}

func NewFx(opts Opts) *Storage {
	s := New()
	s.Add("smtp", opts.SMTP)
	s.Add("telegram", opts.Telegram)
	return s
}

type Storage struct {
	plugins map[string]plugin.Plugin
}

func New() *Storage {
	return &Storage{
		plugins: make(map[string]plugin.Plugin),
	}
}

func (s *Storage) Add(slug string, plugin plugin.Plugin) {
	s.plugins[slug] = plugin
}

func (s *Storage) Delete(slug string) {
	delete(s.plugins, slug)
}

func (s *Storage) Get(slug string) (p plugin.Plugin, ok bool) {
	p, ok = s.plugins[slug]
	if !ok {
		return nil, ok
	}
	p = p.New()
	return
}
