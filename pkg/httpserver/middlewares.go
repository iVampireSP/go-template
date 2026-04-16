package httpserver

import "github.com/go-chi/chi/v5/middleware"

// DefaultMiddlewares returns standard middlewares for HTTP servers.
func DefaultMiddlewares() []Middleware {
	return []Middleware{
		middleware.Recoverer,
		RealIP(DefaultRealIPConfig()),
	}
}
