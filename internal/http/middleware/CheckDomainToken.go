package middleware

import (
	"context"
	"net/http"

	dto "cactus/internal/DTO"
	"cactus/internal/http/response"
	"cactus/internal/pkg/contextkeys"
	"cactus/internal/storage/db"
)

type coreService interface {
	GetTokenByPublicToken(ctx context.Context, token string) (db.Token, error)
	GetKindWokerByID(ctx context.Context, id int32) (dto.KindWorker, error)
}

func CheckDomainToken(s coreService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Token-Domain") // TODO в env
			ctx := r.Context()
			Token, err := s.GetTokenByPublicToken(ctx, token)
			if err != nil || Token.IDSystem == 0 {
				response.UnauthorizedErrorJSON(rw, "Доступ запрещен")
				return
			}

			pluginSlug := r.PathValue("slug")

			kindWorker, err := s.GetKindWokerByID(ctx, Token.IDKindWorker)
			if err != nil {
				response.UnauthorizedErrorJSON(rw, "Доступ запрещен")
				return
			}

			if pluginSlug != kindWorker.Slug {
				response.UnauthorizedErrorJSON(rw, "Доступ запрещен")
				return
			}

			ctx = context.WithValue(ctx, contextkeys.SystemIDKey, Token.IDSystem)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
