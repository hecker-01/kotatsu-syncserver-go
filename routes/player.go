package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

func PlayerRoutes(r chi.Router) {
	controller := controllers.NewPlayerController()
	r.Use(middleware.RequireAuth)
	r.Get("/", controller.List)
}
