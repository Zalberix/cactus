package core

import (
	dto "cactus/internal/DTO"
	"cactus/internal/error/validation"
	"cactus/internal/http/request"
	"cactus/internal/http/response"
	"cactus/internal/service/core"

	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type registerWorkerService interface {
	RegisterWorker(
		ctx context.Context,
		arg core.RegisterWorkerParams,
	) (dto.RegisteWorker, error)
}

// Отмена рассылки писем по UUID что был передан для идентификации сообщения
func RegisterWorker(s registerWorkerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req request.RegisterWorkerRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		errors, err := validation.ValidateStructure(&req)
		if err != nil {
			response.ResponseFailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ResponseValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", errors)
			return
		}

		// TODO проверить токен и вынести в middleware (токен не систем, а токен ИС, он где-то отдельно должен лежать, возможно для каждого worker свой токен)
		if req.Token != "decedb9c-3a96-4b0f-9638-f2276ec624dd" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403)
			return
		}

		workerUUID, err := uuid.Parse(req.WorkerUUID)
		if err != nil {
			response.ResponseFailJSON(w, "ошибка обработки UUID воркера.")
			return
		}

		created, err := s.RegisterWorker(ctx, core.RegisterWorkerParams{
			WorkerUUID:   workerUUID,
			Kind:         req.Kind,
			Type:         req.Type,
			ConfigSchema: req.ConfigSchema,
		})

		if err != nil {
			slog.Info("Ошибка регистрации воркера:", slog.Any("err", err), slog.Any("type", req.Type), slog.Any("kind", req.Kind))
			response.ResponseFailJSON(w, "Неудалось зарегистрировать воркер, проверьте запрос и попробуйте снова.")
			return
		}

		response.ResponseOKJSON(w, response.RegisterWorkerResponse{
			Created: created.Created,
			Id:      created.Id,
			Config:  created.Config,
		})
	}
}
