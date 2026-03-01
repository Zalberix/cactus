package core

import "net/http"

type getLinkService interface{}

// Получение ссылки на страницу просмотра рассылки в реальном времени
func GetLink(_ getLinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// TODO: Реализовать.
		w.Write([]byte("Тест GetLink прошел успешно"))
	}
}
