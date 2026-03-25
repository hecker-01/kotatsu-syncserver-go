package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// HealthRoutes configures the /api/health endpoint.
func HealthRoutes(r chi.Router) {
	r.Get("/", GetHealth)
}

// GetHealth handles GET /api/health. Returns application health information.
func GetHealth(w http.ResponseWriter, r *http.Request) {
	data := utils.GetHealthData()
	data["status"] = "healthy"
	w.Header().Set("Content-Type", "application/json")
	utils.WriteJSON(w, http.StatusOK, data)
}
