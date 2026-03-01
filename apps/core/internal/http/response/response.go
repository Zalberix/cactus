package response

import (
	"encoding/json"
	"net/http"
)

type Status bool

type responseOkAnswer[T any] struct {
	Success Status `json:"success"`
	Type    string `json:"type"`
	Data    T      `json:"data"`
}

type responseFailAnswer struct {
	Success Status `json:"success"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type responseAuthorizationErrorAnswer struct {
	Success Status `json:"success"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type responseValidationAnswer struct {
	Success Status            `json:"success"`
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func OKJSON[T any](w http.ResponseWriter, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonOK, _ := json.Marshal(responseOkAnswer[T]{
		Success: true,
		Type:    "data",
		Data:    data,
	})
	w.Write(jsonOK)
}

func UnauthorizedErrorJSON(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	jsonAuthorizationError, _ := json.Marshal(responseAuthorizationErrorAnswer{
		Success: false,
		Type:    "unauthorized",
		Message: message,
	})
	w.Write(jsonAuthorizationError)
}

func FailJSON(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	jsonFail, _ := json.Marshal(responseFailAnswer{
		Success: false,
		Type:    "fail",
		Message: message,
	})
	w.Write(jsonFail)
}

func ValidationJSON(w http.ResponseWriter, message string, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	jsonValidation, _ := json.Marshal(responseValidationAnswer{
		Success: false,
		Type:    "validation",
		Message: message,
		Errors:  errors,
	})
	w.Write(jsonValidation)
}
