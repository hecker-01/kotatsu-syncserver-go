// Package routes organizes HTTP route registration for all API endpoints.
// Each domain area (auth, users, games, etc.) has its own route setup function.
package routes

import "github.com/go-chi/chi/v5"

// RegisterAPIRoutes mounts all domain-specific route groups under the /api prefix.
func RegisterAPIRoutes(r chi.Router) {
	r.Route("/auth", AuthRoutes)
	r.Route("/users", UserRoutes)
	r.Route("/games", GameRoutes)
	r.Route("/history", HistoryRoutes)
	r.Route("/player", PlayerRoutes)
}
