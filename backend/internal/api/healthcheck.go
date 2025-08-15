package api

import (
	"encoding/json"
	"net/http"
)

// HealthCheckHandler handles the health check endpoint
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"status": "ok"}
	json.NewEncoder(w).Encode(response)
}
