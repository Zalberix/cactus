package middleware

import (
	"cactus/internal/pkg/contextkeys"
	"context"
	"net/http"
)

type EmailService interface {
	GetSystemIdByToken(ctx context.Context, token string) (int32, error)
}

func CheckDomainToken(s EmailService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Token-Domain") // TODO в env
			ctx := r.Context()
			systemID, err := s.GetSystemIdByToken(ctx, token)
			if err != nil {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(401)
				return
			}
			if systemID == 0 {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(401)
				return
			}

			ctx = context.WithValue(ctx, contextkeys.SystemIDKey, systemID)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
