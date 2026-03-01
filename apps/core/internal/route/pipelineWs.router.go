package route

import (
	pipeline_controller "cactus/apps/core/internal/http/controllers/pipeline"
	"cactus/apps/core/internal/pkg/router"
	"cactus/apps/core/internal/service/pipeline"
)

func addRoutePipeline(
	r *router.ServerRouter,
	pipelineService *pipeline.Service,
) *router.ServerRouter {
	r.HandleFunc("GET /ws/pipeline/{uuid}", pipeline_controller.Listen(pipelineService))

	return r
}
