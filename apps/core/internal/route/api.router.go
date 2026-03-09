package route

import (
	core_controller "github.com/zalberix/cactus/apps/core/internal/http/controllers/core"
	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/pkg/router"
	"github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/apps/core/internal/service/pipeline"
	"github.com/zalberix/cactus/apps/core/internal/storage/plugin"
)

func addRouteAPI(
	r *router.ServerRouter,
	coreService *core.Service,
	pipelineService *pipeline.Service,
	plugins *plugin.Storage,
) *router.ServerRouter {
	// TODO: Добавить middleware для авторизованных действий
	// middleware.AuthSession()

	// TODO исправить на message вместо email, и core_controller
	// TODO: Обработать slug чтобы там были только буквы, и небыло спец симовлов

	r.HandleFunc("GET /api/{slug}/list", core_controller.GetMessages(coreService))
	r.HandleFunc("GET /api/app/types-worker/list", core_controller.GetTypesWorker(coreService))

	r.Group(func(sr router.ServerRouter) {
		sr.Use(middleware.CheckDomainToken(coreService))

		// TODO: исправить на message вместо email, и core_controller
		sr.HandleFunc("POST /api/{slug}/send", core_controller.Send(coreService, pipelineService, plugins))
		sr.HandleFunc("POST /api/status", core_controller.GetStatus(coreService))
		sr.HandleFunc("POST /api/abort", core_controller.AbortMessage(coreService))
		sr.HandleFunc("POST /api/link", core_controller.GetLink(coreService))
	})

	r.Group(func(sr router.ServerRouter) {
		// TODO: Добавить проверку на регистрацию воркера
		// sr.Use(middleware.ChechWorkerToken(emailService))

		sr.HandleFunc("POST /api/register/worker", core_controller.RegisterWorker(coreService))
		sr.HandleFunc("GET /api/file/get", core_controller.GetFile(coreService))
		// sr.HandleFunc("POST /api/files/set", core_controller.SetFile(emailService))
	})

	return r
}
