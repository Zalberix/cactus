package router

import "net/http"

type Middleware func(next http.Handler) http.Handler

type ServerRouter struct {
	*http.ServeMux
	middleware []Middleware
}

func NewServerRouter() *ServerRouter {
	return &ServerRouter{
		ServeMux: http.NewServeMux(),
	}
}

func (r *ServerRouter) Use(middleware Middleware) {
	r.middleware = append(r.middleware, middleware)
}

func (r *ServerRouter) HandleFunc(pattern string, handler http.HandlerFunc) {
	finalHandler := http.Handler(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		finalHandler = r.middleware[i](finalHandler)
	}
	r.ServeMux.HandleFunc(pattern, finalHandler.ServeHTTP)
}

func (r *ServerRouter) Group(group func(ServerRouter)) {
	groupServerRouter := ServerRouter{
		ServeMux: r.ServeMux,
	}
	group(groupServerRouter)
}
