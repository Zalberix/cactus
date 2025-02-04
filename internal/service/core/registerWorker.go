package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	dto "cactus/internal/DTO"
	configschema "cactus/internal/pkg/configSchema"
	"cactus/internal/storage/db"
)

type RegisterWorkerParams struct {
	WorkerUUID   uuid.UUID
	Kind         string
	Type         string
	ConfigSchema []configschema.ConfigField
}

func (s *Service) RegisterWorker(
	ctx context.Context,
	arg RegisterWorkerParams,
) (dto.RegisteWorker, error) {
	if _, ok := s.plugins.Get(arg.Kind); !ok {
		return dto.RegisteWorker{}, fmt.Errorf("воркеры с таким типом не включены или не поддерживаются")
	}

	storageTx := s.storage
	err := s.storage.SetContext(ctx, &storageTx)
	if err != nil {
		slog.Error("Ошибка создания транзакции при регистрации воркера:", slog.String("error", err.Error()))
		return dto.RegisteWorker{}, fmt.Errorf("невозможно зарегистрировать воркер: %w", err)
	}
	defer storageTx.Rollback()

	kindWorker, err := storageTx.GetKindWorkerBySlug(ctx, arg.Kind)
	if errors.Is(err, sql.ErrNoRows) {
		configSchemaByte, err := json.Marshal(arg.ConfigSchema)
		if err != nil {
			slog.Error("Ошибка создания json настроке воркера:", slog.String("error", err.Error()))
			return dto.RegisteWorker{}, fmt.Errorf("ошибка создания json настроек: %w", err)
		}
		kindWorker, err = storageTx.CreateKindWorker(ctx, db.CreateKindWorkerParams{
			Name:         "",
			Slug:         arg.Kind,
			ConfigSchema: json.RawMessage(configSchemaByte),
			Config: pqtype.NullRawMessage{
				Valid: false,
			},
		})
		if err != nil {
			slog.Error("Ошибка при создании вида воркера:", slog.String("error", err.Error()))
			return dto.RegisteWorker{}, fmt.Errorf("ошибка при создании вида воркера: %w", err)
		}
	} else if err != nil {
		slog.Error("Ошибка при получении вида воркера:", slog.String("error", err.Error()))
		return dto.RegisteWorker{}, fmt.Errorf("ошибка при получении вида воркер: %w", err)
	}

	config := map[string]interface{}{}
	if kindWorker.Config.Valid {
		err := json.Unmarshal(kindWorker.Config.RawMessage, &config)
		if err != nil {
			slog.Error("неудалось обработать конфиг из базы данных:", slog.String("error", err.Error()))
			return dto.RegisteWorker{}, fmt.Errorf("неудалось обработать конфиг из базы данных: %w", err)
		}
	}

	typeWorker, err := storageTx.GetTypeWorkerBySlug(ctx, arg.Type)
	if errors.Is(err, sql.ErrNoRows) {
		typeWorker, err = storageTx.CreateTypeWorker(ctx, db.CreateTypeWorkerParams{
			Name: "",
			Slug: arg.Kind,
		})
		if err != nil {
			slog.Error("Ошибка при создании типа воркера:", slog.String("error", err.Error()))
			return dto.RegisteWorker{}, fmt.Errorf("ошибка при создании типа воркера: %w", err)
		}
	} else if err != nil {
		slog.Error("Ошибка при получении типа воркера:", slog.String("error", err.Error()))
		return dto.RegisteWorker{}, fmt.Errorf("ошибка при получении типа воркер: %w", err)
	}

	created := false
	worker, err := storageTx.GetWorkerByUUID(ctx, arg.WorkerUUID)
	if errors.Is(err, sql.ErrNoRows) {
		created = true
		worker, err = storageTx.CreateWorker(ctx, db.CreateWorkerParams{
			Uuid:         arg.WorkerUUID,
			IsActive:     false,
			IDTypeWorker: typeWorker.ID,
			IDKindWorker: kindWorker.ID,
		})
		if err != nil {
			slog.Error("Ошибка при регистрации воркера:", slog.String("error", err.Error()))
			return dto.RegisteWorker{}, fmt.Errorf("ошибка при регистрации воркера: ")
		}
	} else if err != nil {
		slog.Error("Ошибка при получении воркера:", slog.String("error", err.Error()))
		return dto.RegisteWorker{}, fmt.Errorf("ошибка при получении воркер: %w", err)
	}

	// Проверка целостности воркера.
	if !created && worker.IDKindWorker != kindWorker.ID && worker.IDTypeWorker != typeWorker.ID {
		// TODO сделать востановление
		slog.Error("Ошибка целостности воркера:", slog.String("error", err.Error()))
		return dto.RegisteWorker{}, fmt.Errorf("ошибка целостности воркера: %w", err)
	}

	err = storageTx.Commit()
	return dto.RegisteWorker{
		Created: created,
		ID:      worker.ID,
		Config:  config,
	}, err
}
