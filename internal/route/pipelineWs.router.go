package route

import (
	pipeline_controller "cactus/internal/http/controllers/pipeline"
	"cactus/internal/pkg/router"
	"cactus/internal/service/pipeline"
)

func addRoutePipeline(
	r *router.ServerRouter,
	pipelineService *pipeline.Service,
) *router.ServerRouter {
	r.HandleFunc("GET /ws/pipeline/{uuid}", pipeline_controller.Listen(pipelineService))

	return r
}
