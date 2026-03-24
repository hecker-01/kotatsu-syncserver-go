package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/handlers"
)

func AuthRoutes(r chi.Router) {
	r.Post("/register", handlers.Register)
	r.Post("/login", handlers.Login)
}
