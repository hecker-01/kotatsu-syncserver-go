package routes

import "github.com/go-chi/chi/v5"

func RegisterAPIRoutes(r chi.Router) {
	r.Route("/auth", AuthRoutes)
	r.Route("/users", UserRoutes)
	r.Route("/games", GameRoutes)
	r.Route("/history", HistoryRoutes)
	r.Route("/player", PlayerRoutes)
}
