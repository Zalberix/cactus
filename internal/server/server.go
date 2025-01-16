package server

import (
	"cactus/internal/config"
	sqlxconect "cactus/internal/pkg/db"
	"cactus/internal/plugin/email"
	"cactus/internal/route"
	"cactus/internal/service/core"
	"cactus/internal/service/pipeline"
	"cactus/internal/storage/db"
	filestorage "cactus/internal/storage/file"
	plugin_storage "cactus/internal/storage/plugin"
	rdb "cactus/internal/storage/redis"
	"context"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	db  *sqlx.DB
	app *http.Server
}

func Create(conf config.Config) (Server, error) {
	ctx := context.Background()

	databaseConect, err := sqlxconect.New(
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

	DBStorage := db.New(databaseConect)
	RDBStorage, err := rdb.New(ctx, &redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		Username: conf.Redis.User,
	})
	if err != nil {
		return Server{}, fmt.Errorf("create redis conection: %w", err)
	}

	_ = RDBStorage // TODO передать в сервис

	fileStorage, _ := filestorage.New("app/files") // TODO path вынести в конфиг

	pluginStorage := plugin_storage.New()
	pluginStorage.Add("email", email.New())
	// pluginStorage.Add("telegram", telegram.New())
	// pluginStorage.Add("push", push.New())

	coreService := core.New(databaseConect, DBStorage, RDBStorage, fileStorage, pluginStorage)
	pipelineService := pipeline.New(databaseConect, DBStorage, RDBStorage, pluginStorage)

	// TODO сделать WS сервис для отслеживания pipeline сообщений в реальном времени
	// chatServer := chat_service.NewService()

	// TODO: сделать общий обработчик ошибок на уровне middleware для http(s)
	r := route.New(
		coreService,
		pipelineService,
		pluginStorage,
	)

	app := &http.Server{
		Addr:         conf.HTTPServer.Address,
		Handler:      r,
		IdleTimeout:  conf.HTTPServer.IdleTimeout,
		ReadTimeout:  conf.HTTPServer.Timeout,
		WriteTimeout: conf.HTTPServer.Timeout,
	}

	return Server{
		db:  databaseConect,
		app: app,
	}, nil
}

func (s *Server) Start() error {
	defer s.db.Close()

	if err := s.app.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
