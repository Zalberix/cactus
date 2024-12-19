package server

import (
	"cactus/internal/config"
	sqlxconect "cactus/internal/pkg/db"
	"cactus/internal/route"
	chat_service "cactus/internal/service/chat"
	"cactus/internal/service/core"
	email_service "cactus/internal/service/email"
	"cactus/internal/storage/db"
	filestorage "cactus/internal/storage/file"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/jmoiron/sqlx"
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

	minioClient, err := minio.New(
		conf.Minio.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(conf.Minio.PublicKey, conf.Minio.PrivateKey, ""),
			Secure: conf.Minio.UseSSL,
		},
	)
	if err != nil {
		slog.Error(err.Error())
	}

	fileStorage := filestorage.New(minioClient)

	emailService := email_service.New(databaseConect, DBStorage, fileStorage)
	coreService := core.New(databaseConect, DBStorage, fileStorage)
	chatServer := chat_service.NewService()

	// TODO: сделать общий обработчик ошибок на уровне middleware для http(s)
	r := route.New(
		coreService,
		emailService,
		chatServer,
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

	dieServer := make(chan error, 1)
	defer close(dieServer)

	go func() {
		if err := s.app.ListenAndServe(); err != nil {
			dieServer <- err
		}
	}()

	err := <-dieServer
	return err
}
