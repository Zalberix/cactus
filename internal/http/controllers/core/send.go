package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/form"

	dto "cactus/internal/DTO"
	"cactus/internal/error/validation"
	"cactus/internal/http/request"
	"cactus/internal/http/response"
	"cactus/internal/pkg"
	"cactus/internal/pkg/contextkeys"
	"cactus/internal/service/core"
	"cactus/internal/service/pipeline"
	"cactus/internal/storage/plugin"
)

type sendService interface {
	CreateMessage(
		ctx context.Context,
		arg core.CreateMessageParams,
		pipelineService pipeline.Service,
	) (dto.CreateMessage, error)
}

// Отправка сообщения
func Send(service sendService, piplineService *pipeline.Service, plugins *plugin.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		IDSystemValue := ctx.Value(contextkeys.SystemIDKey)
		IDSystem, ok := IDSystemValue.(int32)
		if !ok {
			slog.Error("Доступ запрещен")

			response.ValidationJSON(w, "Доступ запрещен", map[string]string{
				"Token": "Доступ запрещен",
			})
			return
		}

		IDKindWorkerValue := ctx.Value(contextkeys.KindIDKey)
		IDKindWorker, ok := IDKindWorkerValue.(int32)
		if !ok {
			slog.Error("Доступ запрещен")

			response.ValidationJSON(w, "Доступ запрещен", map[string]string{
				"Token": "Доступ запрещен",
			})
			return
		}

		pluginSlug := r.PathValue("slug")
		plugin, exist := plugins.Get(pluginSlug)
		if !exist {
			response.FailJSON(w, "Неизвестное название канала рассылки")
			return
		}

		err := r.ParseMultipartForm(32 << 20) // 32 МБ
		if err != nil {
			slog.Error(err.Error())
			response.FailJSON(w, "Превышен размер файлов или формат запроса неверный в запроса.")
			return
		}
		decoder := form.NewDecoder()
		decoder.SetTagName("json")

		var req request.SendMessageRequest

		err = decoder.Decode(&req, r.MultipartForm.Value)
		if err != nil {
			slog.Error("Ошибка декодирования: ", slog.String("message", err.Error()))
			response.FailJSON(w, "Проверьте поля на правильность написания, недолжно быть неизвестных полей.")
			return
		}

		// Валидация полей запроса
		errors, err := validation.ValidateStructure(&req)
		if err != nil {
			slog.Error("Ошибка структуры", slog.Any("err", err))
			response.FailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", errors)
			return
		}

		// Валидация схемы для канала связи (плагина)
		schema := plugin.GetSchema()
		err = json.Unmarshal([]byte(req.Value), schema)
		if err != nil {
			slog.Error("Ошибка при работе с полем value", slog.String("error", err.Error()))
			response.FailJSON(w, "Ошибка при работе с полем value")
			return
		}

		errors, err = validation.ValidateStructure(schema)
		if err != nil {
			slog.Error("Ошибка структуры", slog.Any("err", err))
			response.FailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ValidationJSON(w, "Ошибка валидации поля Value, проверьте отправляемые поля", errors)
			return
		}

		// TODO добавить проверку наличия struce tag'ов для Schema так как валидация должна присутсвовать.
		// Не допускаем поведения когда пользователь может не писать валидацию данных!

		// TODO вынести в plugin или в базу данных (или в базу а брать через плагин,
		// так через настройки плагина можно будет настраивать это поведение)
		allowedExtensions := []string{
			".pdf",
			".jpg", ".jpeg", ".png",
			".doc", ".docx",
			".xls", ".xlsx",
			".zip",
		}
		// TODO вынести бы в функцию, но только если будет гдето еще использоваться, а так пусть тут.
		var files []core.SetFileParams
		if len(r.MultipartForm.File) == len(req.Files) && len(req.Files) != 0 {
			var key string
			var err error
			files, key, err = pkg.ParseFileInMultipartForm(req.Files, r.MultipartForm.File, allowedExtensions)
			if err != nil {
				message := err.Error()
				response.ValidationJSON(w, message, map[string]string{
					key: message,
				})
			}
		}

		message, err := service.CreateMessage(
			ctx,
			core.CreateMessageParams{
				Plugin:       plugin,
				IDKindWorker: IDKindWorker,
				IDSystem:     IDSystem,
				PrioritySlug: req.PrioritySlug,
				ChanelSlug:   pluginSlug,
				Schema:       schema, // TODO удалить или перенести внутрь Value
				SendLater:    req.SendLater,
				Files:        files,
			},
			*piplineService,
		)
		if err != nil {
			slog.Error("ошибка создания сообщения", slog.String("error-message", err.Error()), slog.String("slug", pluginSlug))
			response.FailJSON(w, "Возникли неполадки при создании сообщения. Попробуйте выполнить запрос чуть позже.")
			return
		}

		response.OKJSON(w, response.SendMessageResponse{
			UUID: message.Uuid.String(),
		})
	}
}
