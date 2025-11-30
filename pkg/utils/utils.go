package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteErr(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

// Param retrieves a path parameter from Request. For net/http without routers,
// this is a stub; integrate with your router as needed.
func Param(r *http.Request, key string) string {
	// If using http.ServeMux with PathValue in Go 1.22+, prefer r.PathValue(key)
	if v := r.PathValue(key); v != "" {
		return v
	}
	return r.URL.Query().Get(key)
}
