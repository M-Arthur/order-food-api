package shared

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

// Helper to write JSON response
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		zerolog.Ctx(r.Context()).Error().Err(err).Msg("failed to write JSON")
	}
}

// Helper for simple error payloads.
type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	WriteJSON(w, r, status, ErrorResponse{Message: msg})
}
