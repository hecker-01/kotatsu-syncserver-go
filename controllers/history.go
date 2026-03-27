package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// HistoryController handles history-related endpoints.
type HistoryController struct {
	historyService *services.HistoryService
}

// NewHistoryController creates a new history controller with initialized service.
func NewHistoryController() *HistoryController {
	return &HistoryController{historyService: services.NewHistoryService()}
}

// GetHistory handles GET /resource/history. Returns all history for the authenticated user.
func (c *HistoryController) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pkg, err := c.historyService.GetHistory(userID)
	if err != nil {
		logger.Error("failed to get history", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, pkg)
}

// SyncHistory handles POST /resource/history. Syncs history with conflict resolution.
// Returns 204 if no changes needed, 200 with merged state if conflicts exist.
func (c *HistoryController) SyncHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var clientPkg models.HistoryPackage
	if err := json.NewDecoder(r.Body).Decode(&clientPkg); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	mergedPkg, hasChanges, err := c.historyService.SyncHistory(userID, &clientPkg)
	if err != nil {
		logger.Error("failed to sync history", "user_id", userID, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if !hasChanges {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	utils.WriteJSON(w, http.StatusOK, mergedPkg)
}
