package routes

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

// AuthRoutes configures /api/auth endpoints (register, login) with strict rate limiting.
func AuthRoutes(r chi.Router) {
	controller := controllers.NewAuthController()
	authLimiter := middleware.NewRateLimiter(5, 5*time.Minute)

	r.With(authLimiter).Post("/register", controller.Register)
	r.With(authLimiter).Post("/login", controller.Login)
}
