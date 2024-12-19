package email

import (
	validation "cactus/internal/error/validation"
	request "cactus/internal/http/request/email"
	response "cactus/internal/http/response/email"
	"cactus/internal/pkg/contextkeys"
	schemaApp "cactus/internal/schema"
	"cactus/internal/service/email"
	"cactus/internal/storage/db"
	"log/slog"

	"github.com/gorilla/schema"

	"context"
	"encoding/json"
	"net/http"
)

type sendService interface {
	CreateMessage(
		ctx context.Context,
		arg email.CreateMessageParams,
		files []schemaApp.File, // TODO реализовать загрузку файлов отдельно
	) (db.Message, error)
	SaveFiles(ctx context.Context, r *http.Request) ([]schemaApp.File, error)
}

// Отправка сообщения
func Send(s sendService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var unmarshalErr *json.UnmarshalTypeError
		ctx := r.Context()

		IDSystemValue := ctx.Value(contextkeys.SystemIDKey)
		IDSystem, ok := IDSystemValue.(int32)
		if !ok {
			slog.Error("Ошибка получения ID системы")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			jsonErrors, _ := json.Marshal(response.CreateResponseError("Ошибка получения ID системы", map[string]string{
				"IDSystem": "Ошибка получения ID системы",
			}))
			w.Write(jsonErrors)
			return
		}

		err := r.ParseMultipartForm(32 << 20) // 32 МБ
		if err != nil {
			slog.Error(err.Error())
		}

		// TODO Вынести выше
		var decoder = schema.NewDecoder()

		err = r.ParseForm()
		if err != nil {
			// Handle error
		}

		var req request.SendRequest

		// r.PostForm is a map of our POST form values
		err = decoder.Decode(&req, r.PostForm)
		if err != nil {
			slog.Error("Ошибка декодирования: ", slog.String("message", err.Error()))
			return
		}

		//_ = json.NewDecoder(r.Body).Decode(&req)
		errors := validation.ValidateStructure(&req)
		if errors != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			jsonErrors, _ := json.Marshal(response.CreateResponseError("Ошибка валидации, проверьте отправляемые поля", errors))
			w.Write(jsonErrors)
			return
		}

		files, err := s.SaveFiles(ctx, r)
		if err != nil {
			slog.Error(err.Error())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			jsonErrors, _ := json.Marshal(response.CreateResponseError(err.Error(), nil))
			w.Write(jsonErrors)
			return
		}

		message, _ := s.CreateMessage(
			ctx,
			email.CreateMessageParams{
				IDSystem:     IDSystem,
				PrioritySlug: req.PrioritySlug,
				Title:        req.Title,
				Message:      req.Message,
				Subjects:     req.Subjects,
				TypeMessage:  req.TypeMessage,
				Data:         req.Data,
				Theme:        req.Theme,
				SendLater:    req.SendLater,
			},
			files, // TODO: вынести в отделбный метод сервиса
		)
		// TODO вызов сервис методов и преобразование в response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(response.CreateResponseOK(response.SendResponse{
			Status: "ok",
			UUID:   message.Uuid.String(),
		}))
		w.Write(jsonResp)
	}
}
