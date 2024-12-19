package email

import (
	validation "cactus/internal/error/validation"
	request "cactus/internal/http/request/email"
	response "cactus/internal/http/response/email"
	"context"
	"encoding/json"
	"net/http"
)

type abortEmailService interface {
	AbortEmail(ctx context.Context, uuid string) (string, error)
}

// TODO: вынести в message так как его задача отменаять сообщения и не важно какой тип у собщения
// Отмена рассылки писем по UUID что был передан для идентификации сообщения
func Abort(s abortEmailService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req request.AbortRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		errors := validation.ValidateStructure(&req)
		if errors != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(403)
			jsonErrors, _ := json.Marshal(response.CreateResponseError("Ошибка валидации, проверьте отправляемые поля", errors))
			w.Write(jsonErrors)
			return
		}

		// TODO обработать ошибку
		status, _ := s.AbortEmail(r.Context(), req.UUID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(response.CreateResponseOK(response.AbortResponse{
			Status: status,
		}))
		w.Write(jsonResp)
	}
}
