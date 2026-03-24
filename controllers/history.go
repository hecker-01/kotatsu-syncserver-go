package controllers

import (
	"net/http"

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

// List handles GET /api/history. Returns history data for the authenticated user.
func (c *HistoryController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.historyService.List()})
}
