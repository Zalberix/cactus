package plugin

import "github.com/zalberix/cactus/libs/pipeline"

type CorePlugin struct{}

func (c *CorePlugin) ExtendPipeline(p []pipeline.Step) ([]pipeline.Step, error) {
	return p, nil
}
