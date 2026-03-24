package handlers

import (
	"encoding/json"
	"net/http"
)

func Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID})
}
