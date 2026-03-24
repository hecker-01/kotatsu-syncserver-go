package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

// HistoryRoutes configures /api/history endpoints. All routes require authentication.
func HistoryRoutes(r chi.Router) {
	controller := controllers.NewHistoryController()
	r.Use(middleware.RequireAuth)
	r.Get("/", controller.List)
}
