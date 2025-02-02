package main

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/sqlc-dev/pqtype"

	"cactus/internal/config"
	"cactus/internal/logger"
	configschema "cactus/internal/pkg/configSchema"
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

	// RDBStorage, err := rdb.New(ctx, &redis.Options{
	// 	Addr: cfg.Redis.Address,
	// })

	// fileStorage := filestorage.New()

	// TODO добавить TRANCATE всех таблиц, включаемый по флагу --refresh в консоли.

	CreateDefaultTypesWorker(ctx, DBStorage)

	prioritys, err := CreateDefaultPrioritys(ctx, DBStorage)
	if err != nil {
		slog.Info("не смогли создать приоритеты, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}

	testUser, err := CreateUser(ctx, DBStorage, db.CreateUserParams{
		Fio:                     "Demo user #1",
		Login:                   "demo",
		Email:                   "demo@mail.ru",
		Password:                "password",
		ResetPasswordAfterLogin: sql.NullBool{Valid: true, Bool: false},
		CreateAt:                sql.NullTime{Valid: true, Time: time.Now()},
	})
	if err != nil {
		slog.Info("не смогли создать тестового пользователя, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}

	testSystem, err := CreateSystem(ctx, DBStorage, db.CreateSystemParams{
		CreateUser:  sql.NullInt32{Valid: true, Int32: testUser.ID},
		IDPriority:  prioritys[0].ID,
		Name:        "Тестовая система",
		Description: sql.NullString{Valid: true, String: "demo для работы с системой"},
		IsActive:    true,
	})
	if err != nil {
		slog.Info("не смогли создать тестовую систему, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}

	configSchema, err := json.Marshal([]configschema.ConfigField{
		{
			Type: "numeric",
			Slug: "host",
			Name: "ip адрес сервера SMTP",
		},
		{
			Type: "numeric",
			Slug: "port",
			Name: "Порт сервера",
		},
		{
			Type: "text",
			Slug: "from",
			Name: "Адрем отправителя",
		},
	})
	if err != nil {
		slog.Info("ошибка сериализации JSON для запроса в endpoint для регистрации worker")
		return
	}
	SMTPKindWorker, err := CreateKindWorker(ctx, DBStorage, db.CreateKindWorkerParams{
		Name:         "smtp сервер",
		Slug:         "smtp",
		ConfigSchema: configSchema,
		Config:       pqtype.NullRawMessage{Valid: false}, // Значит что еще не настроен
	})
	if err != nil {
		slog.Info("не смогли создать тестовую систему, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}

	kindWorkerSystem, err := CreateKindWorkerSystem(ctx, DBStorage, db.AddKindWorkerForSystemParams{
		IDSystem:     testSystem.ID,
		IDKindWorker: SMTPKindWorker.ID,
	})
	if err != nil {
		slog.Info("не смогли создать связь системы с видом воркера, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}

	_, err = CreateToken(ctx, DBStorage, db.CreateTokenParams{
		IDSystem:     kindWorkerSystem.IDSystem,
		IDKindWorker: kindWorkerSystem.IDKindWorker,
		IsActive:     true,
		PublicToken:  "12345678910",
		SecretToken:  "10987654321",
	})
	if err != nil {
		slog.Info("не смогли создать токен для системы, дальнейшая работа невозможна:", slog.String("error", err.Error()))
		return
	}
}

func CreateDefaultTypesWorker(ctx context.Context, storage *db.Queries) ([]db.TypeWorker, error) {
	typesWorker := []db.TypeWorker{
		{ID: 1, Name: "Рассылка писем", Slug: "email"},
		{ID: 2, Name: "Телеграмм уведомления", Slug: "telegram"},
		{ID: 3, Name: "Push уведомления", Slug: "push"},
		{ID: 4, Name: "SMS-сообщения", Slug: "sms"},
		{ID: 5, Name: "1С", Slug: "oneC"},
	}

	var createdTypesWorker []db.TypeWorker

	for _, typeWorker := range typesWorker {
		createTypeWorker, err := storage.CreateTypeWorker(ctx, db.CreateTypeWorkerParams{
			ID:   typeWorker.ID,
			Name: typeWorker.Name,
			Slug: typeWorker.Slug,
		})
		if err != nil {
			return []db.TypeWorker{}, fmt.Errorf("ошибка при заполнении типов воркеров: %w", err)
		}
		createdTypesWorker = append(createdTypesWorker, createTypeWorker)
	}

	return createdTypesWorker, nil
}

func CreateDefaultPrioritys(ctx context.Context, storage *db.Queries) ([]db.Priority, error) {
	prioritys := []db.CreatePriorityParams{
		{Weight: 0, Name: "Низкий", Slug: "low"},
		{Weight: 1, Name: "Средний", Slug: "middle"},
		{Weight: 2, Name: "Высокий", Slug: "high"},
		{Weight: 3, Name: "Экстренный", Slug: "extra"},
		{Weight: 4, Name: "Черезвычайный", Slug: "emergency"},
	}

	var createdPrioritys []db.Priority

	for _, priority := range prioritys {
		createdPriority, err := storage.CreatePriority(ctx, priority)
		if err != nil {
			return []db.Priority{}, fmt.Errorf("ошибка при заполнении типов воркеров: %w", err)
		}
		createdPrioritys = append(createdPrioritys, createdPriority)
	}

	return createdPrioritys, nil
}

func CreateKindWorker(ctx context.Context, storage *db.Queries, kind db.CreateKindWorkerParams) (db.KindWorker, error) {
	createdKind, err := storage.CreateKindWorker(ctx, kind)
	if err != nil {
		return db.KindWorker{}, fmt.Errorf("ошибка при заполнении типов воркеров: %w", err)
	}

	return createdKind, nil
}

func CreateKindWorkerSystem(ctx context.Context, storage *db.Queries, kind db.AddKindWorkerForSystemParams) (db.KindWorkerSystem, error) {
	err := storage.AddKindWorkerForSystem(ctx, kind)
	if err != nil {
		return db.KindWorkerSystem{}, fmt.Errorf("ошибка при добавлениии вида воркера для системы: %w", err)
	}

	return db.KindWorkerSystem(kind), nil
}

func CreateSystem(ctx context.Context, storage *db.Queries, system db.CreateSystemParams) (db.System, error) {
	createSystem, err := storage.CreateSystem(ctx, system)
	if err != nil {
		return db.System{}, fmt.Errorf("ошибка при создании систем: %w", err)
	}

	return createSystem, nil
}

func CreateUser(ctx context.Context, storage *db.Queries, user db.CreateUserParams) (db.User, error) {
	hasher := sha512.New()
	hasher.Write([]byte(user.Password))
	hashPassword := hasher.Sum(nil)
	user.Password = hex.EncodeToString(hashPassword)
	createUser, err := storage.CreateUser(ctx, user)
	if err != nil {
		return db.User{}, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return createUser, nil
}

func CreateToken(ctx context.Context, storage *db.Queries, token db.CreateTokenParams) (db.Token, error) {
	createToken, err := storage.CreateToken(ctx, token)
	if err != nil {
		return db.Token{}, fmt.Errorf("ошибка при создании токена: %w", err)
	}
	return createToken, nil
}
