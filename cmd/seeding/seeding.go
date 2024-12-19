package main

import (
	"cactus/internal/config"
	"cactus/internal/logger"
	"context"
	"fmt"
	"log/slog"

	sqlxconect "cactus/internal/pkg/db"
	"cactus/internal/storage/db"
)

func main() {

	cfg := config.MustLoad()

	// isDev := cfg.Env == config.AppEnvDevelopment || cfg.Env == config.AppEnvLocal

	slog.SetDefault(logger.SetupLogger(cfg.Env))

	ctx := context.Background()

	databaseConect, err := sqlxconect.New(
		ctx,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.User,
		cfg.Database.Pass,
	)

	if err != nil {
		slog.Error("Ошибка запуска сервера")
		return
	}

	DBStorage := db.New(databaseConect)

	// minioClient, err := minio.New(
	// 	cfg.Minio.Endpoint,
	// 	&minio.Options{
	// 		Creds:  credentials.NewStaticV4(cfg.Minio.PublicKey, cfg.Minio.PrivateKey, ""),
	// 		Secure: cfg.Minio.UseSSL,
	// 	},
	// )
	// if err != nil {
	// 	slog.Error(err.Error())
	// }

	// fileStorage := filestorage.New(minioClient)

	CreateTypesWorker(ctx, DBStorage)

}

func CreateTypesWorker(ctx context.Context, storage *db.Queries) ([]db.TypeWorker, error) {
	typesWorker := []db.TypeWorker{
		{ID: 1, Name: "Рассылка писем", Slug: "email"},
		{ID: 2, Name: "Телеграмм уведомления", Slug: "telegram"},
		{ID: 3, Name: "Push уведомления", Slug: "push"},
		{ID: 4, Name: "SMS-сообщения", Slug: "sms"},
		{ID: 5, Name: "1С", Slug: "1s"},
	}

	var createdTypesWorker []db.TypeWorker

	for _, typeWorker := range typesWorker {
		createTypeWorker, err := storage.CreateTypeWorker(ctx, db.CreateTypeWorkerParams{
			ID:   typeWorker.ID,
			Name: typeWorker.Name,
			Slug: typeWorker.Slug,
		})
		if err != nil {
			return []db.TypeWorker{}, fmt.Errorf("ошибка при заполнении типов воркеров")
		}
		createdTypesWorker = append(createdTypesWorker, createTypeWorker)
	}

	return createdTypesWorker, nil
}
