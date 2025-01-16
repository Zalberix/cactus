package plugin

import "cactus/internal/pkg/pipeline"

type CorePlugin struct {
}

func (c *CorePlugin) ExtendPipline(pipeline *[]pipeline.PipelineStep) error {
	return nil
}
