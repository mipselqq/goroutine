package handler

import (
	"encoding/json"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(payload)
	_ = err // TODO: log
}

func respondWithError(w http.ResponseWriter, code int, message error) {
	respondWithJSON(w, code, map[string]string{"error": message.Error()})
}
