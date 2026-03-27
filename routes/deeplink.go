// Package routes provides route registration for deep link endpoints.
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
)

// DeeplinkRoutes registers routes for the /deeplink group.
func DeeplinkRoutes(r chi.Router) {
	c := controllers.NewDeeplinkController()
	r.Get("/reset-password", c.ResetPasswordDeeplink)
}
