package controllers

import (
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// GameController handles game-related endpoints.
type GameController struct {
	gameService *services.GameService
}

// NewGameController creates a new game controller with initialized service.
func NewGameController() *GameController {
	return &GameController{gameService: services.NewGameService()}
}

// List handles GET /api/games. Returns a list of available games.
func (c *GameController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.gameService.List()})
}
