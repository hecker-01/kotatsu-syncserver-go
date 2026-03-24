package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
)

// GameRoutes configures /api/games endpoints (public game listing).
func GameRoutes(r chi.Router) {
	controller := controllers.NewGameController()
	r.Get("/", controller.List)
}
