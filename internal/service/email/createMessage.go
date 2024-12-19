package email

import (
	"cactus/internal/schema"
	"cactus/internal/storage/db"
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateMessageParams struct {
	IDSystem       int32
	PrioritySlug   string
	Title, Message string
	Subjects       []string
	TypeMessage    *string
	Data           *map[string]string
	Theme          *string
	SendLater      *time.Time
}

func (s *Service) CreateMessage(
	ctx context.Context,
	arg CreateMessageParams,
	files []schema.File, // TODO реализовать загрузку файлов отдельно
) (db.Message, error) {

	var SendLater sql.NullTime
	if arg.SendLater == nil {
		SendLater = sql.NullTime{
			Valid: false,
		}
	} else {
		SendLater = sql.NullTime{
			Time:  *arg.SendLater,
			Valid: true,
		}
	}

	if arg.Theme == nil {
		defaultTheme := "default"
		arg.Theme = &defaultTheme
	}

	if arg.TypeMessage == nil {
		defaultMessage := "text"
		arg.TypeMessage = &defaultMessage
	}

	if arg.Data == nil || *arg.TypeMessage == "text" {
		defaultData := make(map[string]string)
		arg.Data = &defaultData
	}

	messagePriority, err := s.storage.GetPriorityBySlug(ctx, arg.PrioritySlug)
	if err != nil {
		return db.Message{}, err
	}

	Process, err := s.storage.GetTypeWorkerBySlug(ctx, SLUG_TYPE_WORKER)
	if err != nil {
		return db.Message{}, err
	}

	// TODO: [Вынести логику redis, это не относится к созданию сообщения] после добавления redis дополнить метод удалением из очереди в redis

	value, _ := json.Marshal(schema.Email{
		TypeMessage: *arg.TypeMessage,
		Title:       arg.Title,
		Message:     arg.Message,
		Subjects:    arg.Subjects,
		Data:        *arg.Data,
		Theme:       *arg.Theme,
	})

	newMessage, _ := s.storage.CreateMessage(
		ctx,
		db.CreateMessageParams{
			IDWorker: sql.NullInt32{
				Valid: false, // Пока NULL но надо определять по SendLater текущий Worker
			},
			IDTypeWorker: Process.ID,
			IDSystem:     arg.IDSystem,
			Uuid:         uuid.New(),
			Value:        value,
			IDPriority:   messagePriority.ID,
			SendLater:    SendLater,
		},
	)
	return newMessage, nil
}
