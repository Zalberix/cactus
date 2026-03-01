package plugin

import "cactus/apps/core/internal/plugin"

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
