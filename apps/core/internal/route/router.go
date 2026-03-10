package route

import (
	"github.com/go-chi/cors"
	"go.uber.org/fx"

	"github.com/zalberix/cactus/apps/core/internal/pkg/router"
	"github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/storage/plugin"
)

type Opts struct {
	fx.In
	CoreService     *core.Service
	PipelineService *pipeline.Service
	Plugins         *plugin.Storage
}

func NewFx(opts Opts) *router.ServerRouter {
	return New(opts.CoreService, opts.PipelineService, opts.Plugins)
}

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
