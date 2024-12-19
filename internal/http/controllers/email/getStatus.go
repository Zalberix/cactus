package email

import (
	validation "cactus/internal/error/validation"
	request "cactus/internal/http/request/email"
	response "cactus/internal/http/response/email"
	"context"
	"encoding/json"
	"net/http"
)

type getStatusService interface {
	GetStatus(ctx context.Context, uuid string) (string, error)
}

// Получение статуса по UUID что был передан для идентификации рассылки
func GetStatus(s getStatusService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.GetStatusRequest
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
		status, _ := s.GetStatus(r.Context(), req.UUID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(response.CreateResponseOK(response.GetStatusResponse{
			Status: status,
		}))
		w.Write(jsonResp)
	}
}
