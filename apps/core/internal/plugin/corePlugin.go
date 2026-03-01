package plugin

import "cactus/libs/shared/pipeline"

type CorePlugin struct{}

func (c *CorePlugin) ExtendPipeline(p []pipeline.Step) ([]pipeline.Step, error) {
	return p, nil
}
