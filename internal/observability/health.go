package observability

import (
	"encoding/json"
	"net/http"
)

type probeResponse struct {
	Status string `json:"status"`
}

// HealthHandler handles GET /health (liveness).
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeProbe(w, http.StatusOK, "ok")
}

// ReadyHandler handles GET /ready (readiness).
func ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	writeProbe(w, http.StatusOK, "ready")
}

func writeProbe(w http.ResponseWriter, status int, value string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(probeResponse{Status: value})
}
