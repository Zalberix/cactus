package plugin

import "cactus/internal/pkg/pipeline"

type Plugin interface {
	// TODO Переписать на состояние REQUEST, создаю каждый раз только из за того что отличить
	// REQEST'ы не могу друг от друга, нужен REQUEST ID.
	New() Plugin
	GetSchema() any
	ExtendPipeline([]pipeline.Step) ([]pipeline.Step, error)
}
