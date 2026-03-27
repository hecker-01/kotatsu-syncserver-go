// Package routes organizes HTTP route registration for all API endpoints.
// Each domain area (auth, users, manga, etc.) has its own route setup function.
// Routes are mounted at the root level to match the Kotatsu API structure.
package routes

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/controllers"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

// Rate limiter configurations
var (
	// GlobalAPILimiter applies to most endpoints (100 requests per 5 minutes)
	GlobalAPILimiter = middleware.NewRateLimiter(100, 5*time.Minute)
	// AuthLimiter applies to /auth endpoint (5 requests per 5 minutes)
	AuthLimiter = middleware.NewRateLimiter(5, 5*time.Minute)
	// ForgotPasswordLimiter applies to /forgot-password endpoint (3 requests per 5 minutes)
	ForgotPasswordLimiter = middleware.NewRateLimiter(3, 5*time.Minute)
	// ResetPasswordLimiter applies to /reset-password endpoint (5 requests per 5 minutes)
	ResetPasswordLimiter = middleware.NewRateLimiter(5, 5*time.Minute)
)

// RegisterRoutes mounts all domain-specific routes at the root level,
// matching the Kotatsu API structure.
func RegisterRoutes(r chi.Router) {
	authController := controllers.NewAuthController()
	passwordController := controllers.NewPasswordController()
	userController := controllers.NewUserController()
	mangaController := controllers.NewMangaController()
	historyController := controllers.NewHistoryController()
	favouritesController := controllers.NewFavouritesController()

	// Auth endpoints at root level with specific rate limiters
	r.With(AuthLimiter).Post("/auth", authController.Auth)
	r.With(ForgotPasswordLimiter).Post("/forgot-password", passwordController.ForgotPassword)
	r.With(ResetPasswordLimiter).Post("/reset-password", passwordController.ResetPassword)

	// Deeplink routes (no rate limiter needed)
	r.Route("/deeplink", DeeplinkRoutes)

	// /me endpoint - requires auth and global rate limiter
	r.With(GlobalAPILimiter, middleware.RequireAuth).Get("/me", userController.Me)

	// /manga endpoints - public, with global rate limiter
	r.Route("/manga", func(r chi.Router) {
		r.Use(GlobalAPILimiter)
		r.Get("/", mangaController.ListManga)
		r.Get("/{id}", mangaController.GetManga)
	})

	// /resource endpoints - require auth and global rate limiter
	r.Route("/resource", func(r chi.Router) {
		r.Use(GlobalAPILimiter)
		r.Use(middleware.RequireAuth)

		// Favourites endpoints
		r.Get("/favourites", favouritesController.GetFavourites)
		r.Post("/favourites", favouritesController.SyncFavourites)

		// History endpoints
		r.Get("/history", historyController.GetHistory)
		r.Post("/history", historyController.SyncHistory)
	})
}
