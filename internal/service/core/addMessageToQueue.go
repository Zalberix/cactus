package core

import (
	"cactus/internal/storage/db"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type AddMessageToQueueParams struct {
	db.Message
	SlugKindWorker        string
	SlugTypeWorker        string
	WeightPriorityMessage int32
}

func (s *Service) AddMessageToQueue(ctx context.Context, arg AddMessageToQueueParams) error {
	nameQueue := fmt.Sprintf("%v_%v_%v", arg.SlugTypeWorker, arg.SlugKindWorker, arg.WeightPriorityMessage)
	nameListQueue := nameQueue + ":list"
	nameMapQueue := nameQueue + ":hash"
	_ = nameMapQueue

	pipe := s.rdb.TxPipeline()

	// Логика такая: добавиляем в список UUID, а рядом ложим hash для хранения значений, чтобы можно было удалять значения по UUID. В списке этого не сделать без LUA, а для получения данных их hash нужен ключ или получать сразу все.
	pipe.RPush(ctx, nameListQueue, arg.Uuid.String())

	valueHash, err := json.Marshal(arg.Message)
	if err != nil {
		slog.Error("Не удалось сформировать JSON из сообщения: ", slog.Any("err", err))
		return fmt.Errorf("не удалось сформировать JSON из сообщения: %w", err)
	}
	pipe.HSet(ctx, nameMapQueue, arg.Uuid.String(), valueHash)

	_, err = pipe.Exec(ctx)
	if err != nil {
		slog.Error("Ошибка добавления в очередь: ", slog.Any("err", err))
		// TODO прервать pipeline (а как? снова pipelineService сюда передавать? Мне кажется надо сервис pipeline как поле для coreService добавить, так как отменять пайплайны будем часто)
		return fmt.Errorf("ошибка добавления в очередь")
	}

	return nil
}
