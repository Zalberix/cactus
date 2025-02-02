package plugin

import "cactus/internal/pkg/pipeline"

type CorePlugin struct{}

func (c *CorePlugin) ExtendPipline(_ *[]pipeline.Step) error {
	return nil
}
