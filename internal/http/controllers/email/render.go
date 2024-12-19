package email

import "net/http"

type renderService interface {
}

// Генерирует разметку для просмотра как будет выглядеть сообщение, если email
func Render(s renderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
