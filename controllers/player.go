package controllers

import (
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// PlayerController handles player-related endpoints.
type PlayerController struct {
	playerService *services.PlayerService
}

// NewPlayerController creates a new player controller with initialized service.
func NewPlayerController() *PlayerController {
	return &PlayerController{playerService: services.NewPlayerService()}
}

// List handles GET /api/player. Returns player data for the authenticated user.
func (c *PlayerController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.playerService.List()})
}
