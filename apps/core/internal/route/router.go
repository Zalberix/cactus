package route

import (
	"github.com/go-chi/cors"

	"cactus/apps/core/internal/pkg/router"
	"cactus/apps/core/internal/service/core"
	"cactus/apps/core/internal/service/pipeline"
	"cactus/apps/core/internal/storage/plugin"
)

func New(
	coreService *core.Service,
	pipelineService *pipeline.Service,
	plugins *plugin.Storage,
) *router.ServerRouter {
	r := router.NewServerRouter()

	// TODO вынести в middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	addRouteAPI(r, coreService, pipelineService, plugins)

	addRoutePipeline(r, pipelineService)

	return r
}
