package core

import (
	"context"
	"encoding/json"
	"net/http"

	"cactus/internal/error/validation"
	"cactus/internal/http/request"
	"cactus/internal/http/response"
)

type abortEmailService interface {
	AbortMessage(ctx context.Context, uuid string) (string, error)
}

// Отмена рассылки писем по UUID что был передан для идентификации сообщения
func AbortMessage(s abortEmailService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.AbortMessageRequest
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

		// TODO обработать ошибку
		status, _ := s.AbortMessage(r.Context(), req.UUID)
		response.OKJSON(w, response.AbortMessageResponse{
			Status: status,
		})
	}
}
