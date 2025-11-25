package handlers

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_, err := w.Write([]byte(`{"status":"ok"}`))
	if err != nil {
		// At this point we already tried to write to the client.
		// Connection is likely broke; nothing more to do.
		// We intentionally ignore the error here.
		return
	}
}
