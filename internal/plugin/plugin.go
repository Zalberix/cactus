package plugin

import "cactus/internal/pkg/pipeline"

//go:generate mockgen -package=mocks -destination=mocks/mock_plugin.go cactus/internal/plugin Plugin
type Plugin interface {
	// TODO Переписать на состояние REQUEST, создаю каждый раз только из за того что отличить
	// REQEST'ы не могу друг от друга, нужен REQUEST ID.
	New() Plugin
	GetSchema() any
	ExtendPipeline([]pipeline.Step) ([]pipeline.Step, error)
}
