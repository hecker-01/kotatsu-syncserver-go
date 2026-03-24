package handlers

import (
	"encoding/json"
	"net/http"
)

// Me is a legacy handler that returns the authenticated user's ID.
// New code should use controllers.UserController.Me instead.
func Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID})
}
