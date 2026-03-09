package core

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zalberix/cactus/apps/core/internal/error/validation"
	"github.com/zalberix/cactus/apps/core/internal/http/request"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
)

type getStatusService interface {
	GetStatus(ctx context.Context, uuid string) (string, error)
}

// Получение статуса по UUID что был передан для идентификации рассылки
func GetStatus(s getStatusService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.GetStatusMessageRequest
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

		status, err := s.GetStatus(r.Context(), req.UUID)
		if err != nil {
			// TODO: добавить trace для ошибок
			response.FailJSON(w, "Внутренняя ошибка сервера. Пожалуйста, попробуйте позже.")
		}

		response.OKJSON(w, response.GetStatusMessageResponse{
			Status: status,
		})
	}
}
