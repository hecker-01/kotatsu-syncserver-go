package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// MangaController handles manga-related endpoints.
type MangaController struct {
	mangaService *services.MangaService
}

// NewMangaController creates a new manga controller with initialized service.
func NewMangaController() *MangaController {
	return &MangaController{mangaService: services.NewMangaService()}
}

// ListManga handles GET /api/manga?offset={offset}&limit={limit}
// Returns a paginated list of manga.
func (c *MangaController) ListManga(w http.ResponseWriter, r *http.Request) {
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		utils.WriteError(w, http.StatusBadRequest, `Parameter "offset" is missing or invalid`)
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, `Parameter "offset" is missing or invalid`)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		utils.WriteError(w, http.StatusBadRequest, `Parameter "limit" is missing or invalid`)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, `Parameter "limit" is missing or invalid`)
		return
	}

	mangaList, err := c.mangaService.ListManga(offset, limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, mangaList)
}

// GetManga handles GET /api/manga/{id}
// Returns a single manga by ID or 404 if not found.
func (c *MangaController) GetManga(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Not Found")
		return
	}

	manga, err := c.mangaService.GetManga(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if manga == nil {
		utils.WriteError(w, http.StatusNotFound, "Not Found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, manga)
}
