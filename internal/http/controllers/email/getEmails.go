package email

import (
	dto "cactus/internal/DTO"
	validation "cactus/internal/error/validation"
	request "cactus/internal/http/request/message"
	response "cactus/internal/http/response/email"
	"cactus/internal/schema"
	"context"
	"encoding/json"
	"net/http"
)

type getEmailsService interface {
	GetEmails(context.Context, int) ([]dto.Message[schema.Email], error)
}

// Возвращает список всех сообщений определенной системы и процесса email
func GetEmails(s getEmailsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.GetMessageByRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		// Переписать валидатор, так как сейчас мы пишем в структуре валидацию и приходится туда сюда прыгать.
		errors := validation.ValidateStructure(&req)
		if errors != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403) // TODO код не 403 (forbiden) а другой.
			jsonErrors, _ := json.Marshal(response.CreateResponseError("Ошибка валидации, проверьте отправляемые поля", errors))
			w.Write(jsonErrors)
			return
		}

		// TODO обработать ошибку
		// TODO: контекст объеденить с контекстом приложения, мне важно чтобы при завершении приложения и ответы вернулись.
		emails, _ := s.GetEmails(r.Context(), req.ClientId)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(response.CreateResponseOK(response.GetEmailsResponse{
			Emails: emails,
		}))
		w.Write(jsonResp)
	}
}
