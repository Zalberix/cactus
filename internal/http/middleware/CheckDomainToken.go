package middleware

import (
	"cactus/internal/pkg/contextkeys"
	"cactus/internal/storage/db"
	"context"
	"net/http"
)

type coreService interface {
	GetTokenByPublicToken(ctx context.Context, token string) (db.Token, error)
}

func CheckDomainToken(s coreService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Token-Domain") // TODO в env
			ctx := r.Context()
			Token, err := s.GetTokenByPublicToken(ctx, token)
			if err != nil || Token.IDSystem == 0 {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(401)
				return
			}

			ctx = context.WithValue(ctx, contextkeys.SystemIDKey, Token.IDSystem)
			ctx = context.WithValue(ctx, contextkeys.KindIDKey, Token.IDKindWorker)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
