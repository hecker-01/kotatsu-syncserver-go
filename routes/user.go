package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

// UserRoutes configures /api/users endpoints. All routes require authentication.
func UserRoutes(r chi.Router) {
	controller := controllers.NewUserController()

	r.Use(middleware.RequireAuth)
	r.Get("/me", controller.Me)
}
