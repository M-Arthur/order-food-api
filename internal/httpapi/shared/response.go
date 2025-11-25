package shared

import (
	"encoding/json"
	"net/http"
)

// Helper to write JSON response
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
	// if err := json.NewEncoder(w).Encode(v); err != nil {
	// 	// TODO: log the error message
	// }
}

// Helper for simple error payloads.
type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Message: msg})
}
