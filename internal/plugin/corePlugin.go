package plugin

import "cactus/pkg/pipeline"

type CorePlugin struct{}

func (c *CorePlugin) ExtendPipeline(p []pipeline.Step) ([]pipeline.Step, error) {
	return p, nil
}
