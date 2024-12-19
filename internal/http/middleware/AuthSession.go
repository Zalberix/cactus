package middleware

import (
	"net/http"
)

func AuthSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

			next.ServeHTTP(rw, r)
		})
	}
}
