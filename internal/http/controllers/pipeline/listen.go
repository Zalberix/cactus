package pipeline

import (
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"cactus/internal/http/response"
	"cactus/internal/service/pipeline"
)

// Подключение к сокету.
func Listen(s *pipeline.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ws, err := s.Hub().Upgrader().Upgrade(w, r, nil)
		if err != nil {
			slog.Error(err.Error())
			response.FailJSON(w, "неудалось создать websocket соединение")
			return
		}
		defer ws.Close()

		uuid := r.PathValue("uuid")
		if uuid == "" {
			response.ValidationJSON(w, "в url должен быть uuid сообщения", map[string]string{
				"uuid": "является обязательным для создания websocket соединения.",
			})
			return
		}

		if err = validator.New().VarCtx(ctx, uuid, "uuid"); err != nil {
			response.ValidationJSON(w, "в url должен быть uuid сообщения", map[string]string{
				"uuid": "поле должно быть валидным uuid",
			})
			return
		}

		// TODO проверить что сообщение по uuid существует

		client := s.Hub().NewClient(ctx, uuid, ws)

		client.Listen()
	}
}
