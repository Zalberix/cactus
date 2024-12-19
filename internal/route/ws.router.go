package route

import (
	"cactus/internal/http/controllers/chat"
	chat_service "cactus/internal/service/chat"

	"github.com/go-chi/chi/v5"
)

func addRouteWS(
	r *chi.Mux,
	chatService *chat_service.Service,
) *chi.Mux {
	// TODO: вынести выше, тут запуск сервера WS
	go chatService.Listen()

	// r.Use(middleware.CheckDomainToken(emailService))

	r.Handle("/chat/{chatID}", chat.Chat(chatService))

	return r
}
