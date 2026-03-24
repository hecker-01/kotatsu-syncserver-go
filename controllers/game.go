package controllers

import (
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type GameController struct {
	gameService *services.GameService
}

func NewGameController() *GameController {
	return &GameController{gameService: services.NewGameService()}
}

func (c *GameController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.gameService.List()})
}
