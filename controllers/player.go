package controllers

import (
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type PlayerController struct {
	playerService *services.PlayerService
}

func NewPlayerController() *PlayerController {
	return &PlayerController{playerService: services.NewPlayerService()}
}

func (c *PlayerController) List(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": c.playerService.List()})
}
