package controllers

import (
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type HistoryController struct {
	historyService *services.HistoryService
}

func NewHistoryController() *HistoryController {
	return &HistoryController{historyService: services.NewHistoryService()}
}

func (c *HistoryController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.historyService.List()})
}
