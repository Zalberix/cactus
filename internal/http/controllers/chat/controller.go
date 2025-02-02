package chat

import (
	"log/slog"
	"net/http"

	chat_service "cactus/internal/service/chat"
)

// Подключение к сокету.
func Chat(s *chat_service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()

		ws, err := s.Upgrader().Upgrade(w, r, nil)
		if err != nil {
			slog.Error(err.Error())
		}
		defer ws.Close()

		client := chat_service.NewClient(ws, s)
		s.Add(client)
		client.Listen()
	}
}
