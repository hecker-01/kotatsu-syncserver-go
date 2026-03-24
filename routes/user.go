package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/handlers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

func UserRoutes(r chi.Router) {
	r.Use(middleware.RequireAuth)
	r.Get("/me", handlers.Me)
}
