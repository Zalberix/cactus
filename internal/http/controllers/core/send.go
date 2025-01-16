package core

import (
	dto "cactus/internal/DTO"
	"cactus/internal/error/validation"
	"cactus/internal/http/request"
	"cactus/internal/http/response"
	"cactus/internal/pkg/contextkeys"
	"cactus/internal/service/core"
	"cactus/internal/service/pipeline"
	"cactus/internal/storage/plugin"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-playground/form"

	"context"
	"net/http"
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

			response.ResponseValidationJSON(w, "Доступ запрещен", map[string]string{
				"Token": "Доступ запрещен",
			})
			return
		}

		IDKindWorkerValue := ctx.Value(contextkeys.KindIDKey)
		IDKindWorker, ok := IDKindWorkerValue.(int32)
		if !ok {
			slog.Error("Доступ запрещен")

			response.ResponseValidationJSON(w, "Доступ запрещен", map[string]string{
				"Token": "Доступ запрещен",
			})
			return
		}

		pluginSlug := r.PathValue("slug")
		plugin, exist := plugins.Get(pluginSlug)
		if !exist {
			response.ResponseFailJSON(w, "Неизвестное название канала рассылки")
			return
		}

		err := r.ParseMultipartForm(32 << 20) // 32 МБ
		if err != nil {
			slog.Error(err.Error())
			response.ResponseFailJSON(w, "Превышен размер файлов или формат запроса неверный в запроса.")
			return
		}
		var decoder = form.NewDecoder()
		decoder.SetTagName("json")

		var req request.SendMessageRequest

		err = decoder.Decode(&req, r.MultipartForm.Value)
		if err != nil {
			slog.Error("Ошибка декодирования: ", slog.String("message", err.Error()))
			response.ResponseFailJSON(w, "Проверьте поля на правильность написания, недолжно быть неизвестных полей.")
			return
		}

		// Валидация полей запроса
		errors, err := validation.ValidateStructure(&req)
		if err != nil {
			slog.Error("Ошибка структуры", slog.Any("err", err))
			response.ResponseFailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ResponseValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", errors)
			return
		}

		// Валидация схемы для канала связи (плагина)
		schema := plugin.GetSchema()
		err = json.Unmarshal([]byte(req.Value), schema)
		if err != nil {
			slog.Error("Ошибка при работе с полем value", slog.String("error", err.Error()))
			response.ResponseFailJSON(w, "Ошибка при работе с полем value")
			return
		}

		errors, err = validation.ValidateStructure(schema)
		if err != nil {
			slog.Error("Ошибка структуры", slog.Any("err", err))
			response.ResponseFailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ResponseValidationJSON(w, "Ошибка валидации поля Value, проверьте отправляемые поля", errors)
			return
		}

		// TODO добавить проверку наличия struce tag'ов для Schema так как валидация должна присутсвовать. Не допускаем поведения когда пользователь может не писать валидацию данных!

		allowedExtensions := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx", ".xls", ".xlsx", ".zip"} // TODO вынести в plugin или в базу данных (или в базу а брать через плагин, так через настройки плагина можно будет настраивать это поведение)

		// TODO вынести бы в функцию, но только если будет гдето еще использоваться, а так пусть тут.
		var files []core.SetFileParams
		if len(r.MultipartForm.File) == len(req.Files) && len(req.Files) != 0 {
			for i, file := range req.Files {
				key := fmt.Sprintf("Files.%v", i)
				MultipartFiles, ok := r.MultipartForm.File[file.Form]
				if !ok {
					message := fmt.Sprintf("Файл по ключу %v не найден", file.Form)
					response.ResponseValidationJSON(w, message, map[string]string{
						key: message,
					})
					return
				}
				ff, _ := MultipartFiles[0].Open()
				mtype, err := mimetype.DetectReader(ff)

				isAllowedExtensions := slices.Contains(allowedExtensions, mtype.Extension())

				if err != nil || !isAllowedExtensions {
					message := "Файл может быть расширения pdf,jpg,jpeg,png,doc,docx,xls,xlsx,zip"
					response.ResponseValidationJSON(w, message, map[string]string{
						key: message,
					})
					return
				}

				_, err = ff.Seek(0, io.SeekStart)
				if err != nil {
					response.ResponseValidationJSON(w, "Не удалось прочитать файл", map[string]string{
						key: "Не удалось прочитать файл",
					})
				}

				var newFile core.SetFileParams = core.SetFileParams{
					File:  ff,
					Title: file.Title,
					Ext:   *mtype,
				}

				files = append(files, newFile)
			}
		}

		message, err := service.CreateMessage(
			ctx,
			core.CreateMessageParams{
				Plugin:       plugin,
				IDKindWorker: IDKindWorker,
				IDSystem:     IDSystem,
				PrioritySlug: req.PrioritySlug,
				ChangelSlug:  pluginSlug,
				Schema:       schema,
				Title:        req.Title,
				Message:      req.Message,
				SendLater:    req.SendLater,
				Subject:      req.Subject,
				Files:        files,
			},
			*piplineService,
		)
		if err != nil {
			slog.Error("ошибка создания сообщения", slog.String("error-message", err.Error()), slog.String("slug", pluginSlug))
			response.ResponseFailJSON(w, "Возникли неполадки при создании сообщения. Попробуйте выполнить запрос чуть позже.")
			return
		}

		response.ResponseOKJSON(w, response.SendMessageResponse{
			UUID: message.Uuid.String(),
		})
	}
}
