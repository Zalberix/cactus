package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	dto "github.com/zalberix/cactus/apps/core/internal/DTO"
	"github.com/zalberix/cactus/apps/core/internal/error/validation"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
	"github.com/zalberix/cactus/apps/core/internal/service/core"
	"github.com/zalberix/cactus/libs/pipeline"
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

		var req pipeline.RegisterWorkerRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		errors, err := validation.ValidateStructure(&req)
		if err != nil {
			response.FailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", errors)
			return
		}

		// TODO проверить токен и вынести в middleware (токен не систем, а токен ИС,
		// он где-то отдельно должен лежать, возможно для каждого worker свой токен)
		if req.Token != "decedb9c-3a96-4b0f-9638-f2276ec624dd" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403)
			return
		}

		workerUUID, err := uuid.Parse(req.WorkerUUID)
		if err != nil {
			response.FailJSON(w, "ошибка обработки UUID воркера.")
			return
		}

		created, err := s.RegisterWorker(ctx, core.RegisterWorkerParams{
			WorkerUUID:   workerUUID,
			Kind:         req.Kind,
			NameKind:     req.NameKind,
			Type:         req.Type,
			NameType:     req.NameType,
			ConfigSchema: req.ConfigSchema,
		})
		if err != nil {
			slog.Info(
				"Ошибка регистрации воркера:",
				slog.Any("err", err.Error()),
				slog.Any("type", req.Type),
				slog.Any("kind", req.Kind),
			)
			response.FailJSON(w, "Не удалось зарегистрировать воркер, проверьте запрос и попробуйте снова.")
			return
		}

		response.OKJSON(w, pipeline.RegisterWorkerResponse{
			Created: created.Created,
			ID:      created.ID,
			Config:  created.Config,
		})
	}
}
