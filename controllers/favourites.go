// Package controllers provides HTTP handlers for favourites sync endpoints.
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// FavouritesController handles favourites sync endpoints.
type FavouritesController struct {
	favouritesService *services.FavouritesService
}

// NewFavouritesController creates a new favourites controller with initialized service.
func NewFavouritesController() *FavouritesController {
	return &FavouritesController{favouritesService: services.NewFavouritesService()}
}

// GetFavourites handles GET /resource/favourites.
// Returns all categories and favourites for the authenticated user.
func (c *FavouritesController) GetFavourites(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pkg, err := c.favouritesService.GetFavourites(userID)
	if err != nil {
		logger.Error("failed to get favourites", "error", err, "user_id", userID)
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, pkg)
}

// SyncFavourites handles POST /resource/favourites.
// Synchronizes client favourites with the server.
// Returns 204 No Content if client data was accepted (client is up-to-date).
// Returns 200 with merged data if there were server-side changes to sync.
func (c *FavouritesController) SyncFavourites(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var pkg models.FavouritesPackage
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, hasChanges, err := c.favouritesService.SyncFavourites(userID, &pkg)
	if err != nil {
		logger.Error("failed to sync favourites", "error", err, "user_id", userID)
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !hasChanges {
		// Client data accepted, no changes to return
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return merged server state
	utils.WriteJSON(w, http.StatusOK, result)
}
