package core

import (
	"context"
	"encoding/json"
	"net/http"

	dto "cactus/apps/core/internal/DTO"
	"cactus/apps/core/internal/error/validation"
	"cactus/apps/core/internal/http/request"
	"cactus/apps/core/internal/http/response"
)

type getMessagesService interface {
	GetMessages(ctx context.Context, slug string, systemID int) ([]dto.Message, error)
}

func GetMessages(s getMessagesService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.GetMessageForRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		// Переписать валидатор, так как сейчас мы пишем в структуре валидацию и приходится туда сюда прыгать.
		errors, err := validation.ValidateStructure(&req)
		if err != nil {
			response.FailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if errors != nil {
			response.ValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", errors)
			return
		}

		slug := r.PathValue("slug")
		// s.GetMessages

		// TODO обработать ошибку
		// TODO: контекст объеденить с контекстом приложения, мне важно чтобы при завершении приложения и ответы вернулись.

		messages, _ := s.GetMessages(r.Context(), slug, req.ClientID)

		response.OKJSON(w, response.GetMessagesResponse{
			Messages: messages,
		})
	}
}
