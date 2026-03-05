package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"cactus/apps/core/config"
	wshub "cactus/apps/core/internal/pkg/wshub"
	"cactus/apps/core/internal/plugin/smtp"
	"cactus/apps/core/internal/plugin/telegram"
	"cactus/apps/core/internal/route"
	"cactus/apps/core/internal/server/meta"
	"cactus/apps/core/internal/service/core"
	"cactus/apps/core/internal/service/pipeline"
	"cactus/apps/core/internal/storage/broker"
	filestorage "cactus/apps/core/internal/storage/file"
	pluginstorage "cactus/apps/core/internal/storage/plugin"
	"cactus/apps/core/internal/storage/store"
	sqlxconect "cactus/apps/core/pkg/db"
	"cactus/apps/core/storage/db"
	rdb "cactus/libs/shared/redis"
)

type Server struct {
	db   *sqlx.DB
	rdb  *redis.Client
	app  *http.Server
	Meta meta.ServerMeta
}

func Create(conf config.Config) (Server, error) {
	ctx := context.Background()

	databaseConnect, err := sqlxconect.New(
		ctx,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Name,
		conf.Database.User,
		conf.Database.Pass,
	)
	if err != nil {
		return Server{}, fmt.Errorf("create database: %w", err)
	}

	DBStorage := db.New(databaseConnect)
	brokerApp, err := broker.New(ctx, &redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		Username: conf.Redis.User,
	})
	if err != nil {
		return Server{}, fmt.Errorf("create broker connection: %w", err)
	}

	RDBStorage, err := rdb.New(ctx, &redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		Username: conf.Redis.User,
	})
	if err != nil {
		return Server{}, fmt.Errorf("create redis connection: %w", err)
	}

	fileStorage, _ := filestorage.New("storages/local") // TODO path вынести в конфиг

	pluginStorage := pluginstorage.New()

	pluginStorage.Add("smtp", smtp.New())
	pluginStorage.Add("telegram", telegram.New())
	// pluginStorage.Add("push", push.New())

	host, err := os.Hostname()
	if err != nil {
		return Server{}, fmt.Errorf("неудалось получить hostname приложения: %w", err)
	}

	endpoint := fmt.Sprintf("http://%v:%v/api/register/worker", host, conf.HTTPServer.Port)

	var maxPriority int32
	maxPriority, err = DBStorage.GetMaxPriorityWeight(ctx)
	if err != nil {
		maxPriority = 0
	}

	metaServer := &meta.ServerMeta{
		HostName: host,
		Port:     conf.HTTPServer.Port,
	}

	var pipelineService *pipeline.Service
	// TODO сомнительное решение, тип подождать пока загрузится

	storeApp := store.New(databaseConnect, DBStorage)

	coreService := core.New(storeApp, brokerApp, fileStorage, pluginStorage, metaServer)
	pipelineService = pipeline.New(storeApp, pluginStorage)

	pipelineHub := wshub.NewPipelineHub(ctx, RDBStorage, pipelineService)
	pipelineService.SetHub(pipelineHub)
	go pipelineHub.Run()

	// TODO сделать WS сервис для отслеживания pipeline сообщений в реальном времени
	// chatServer := chat_service.NewService()

	// TODO: сделать общий обработчик ошибок на уровне middleware для http(s)
	r := route.New(
		coreService,
		pipelineService,
		pluginStorage,
	)
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	err = brokerApp.SendMetaEvent(ctx, maxPriority, endpoint)
	if err != nil {
		return Server{}, fmt.Errorf("неудалось записать hostname в redis: %w", err)
	}

	app := &http.Server{
		Addr:         fmt.Sprintf("%v:%v", conf.HTTPServer.Host, conf.HTTPServer.Port),
		Handler:      r,
		IdleTimeout:  conf.HTTPServer.IdleTimeout,
		ReadTimeout:  conf.HTTPServer.Timeout,
		WriteTimeout: conf.HTTPServer.Timeout,
	}

	return Server{
		db:  databaseConnect,
		rdb: RDBStorage,
		app: app,
	}, nil
}

func (s *Server) Start() error {
	defer s.db.Close()
	defer s.rdb.Close()

	if err := s.app.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
