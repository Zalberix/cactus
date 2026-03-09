package route

import (
	pipeline_controller "github.com/zalberix/cactus/apps/core/internal/http/controllers/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/pkg/router"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
)

func addRoutePipeline(
	r *router.ServerRouter,
	pipelineService *pipeline.Service,
) *router.ServerRouter {
	r.HandleFunc("GET /ws/pipeline/{uuid}", pipeline_controller.Listen(pipelineService))

	return r
}
