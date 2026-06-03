package timeapi

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler serves time endpoints for CI/CD and GitOps sample workloads.
type Handler struct {
	clock Clock
}

// NewHandler returns a handler that uses the system clock.
func NewHandler() *Handler {
	return &Handler{clock: RealClock{}}
}

// NewHandlerWithClock returns a handler with a custom clock (for tests).
func NewHandlerWithClock(clock Clock) *Handler {
	return &Handler{clock: clock}
}

type utcResponse struct {
	Datetime string `json:"datetime"`
	Timezone string `json:"timezone"`
}

type epochResponse struct {
	Datetime string `json:"datetime"`
	Epoch    int64  `json:"epoch"`
	Timezone string `json:"timezone"`
}

type tzResponse struct {
	Datetime string `json:"datetime"`
	Timezone string `json:"timezone"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NowUTC handles GET / — current time in UTC as ISO-8601.
func (h *Handler) NowUTC(w http.ResponseWriter, r *http.Request) {
	now := h.clock.Now().UTC()
	writeJSON(w, http.StatusOK, utcResponse{
		Datetime: now.Format(time.RFC3339),
		Timezone: "UTC",
	})
}

// Epoch handles GET /epoch — current UTC time as ISO-8601 and Unix epoch seconds.
func (h *Handler) Epoch(w http.ResponseWriter, r *http.Request) {
	now := h.clock.Now().UTC()
	writeJSON(w, http.StatusOK, epochResponse{
		Datetime: now.Format(time.RFC3339),
		Epoch:    now.Unix(),
		Timezone: "UTC",
	})
}

// Timezone handles GET /TZ/{tz...} — current time in the given IANA timezone.
func (h *Handler) Timezone(w http.ResponseWriter, r *http.Request) {
	tzName := r.PathValue("tz")
	if tzName == "" {
		writeError(w, http.StatusBadRequest, "timezone is required")
		return
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid timezone: "+tzName)
		return
	}

	now := h.clock.Now().In(loc)
	writeJSON(w, http.StatusOK, tzResponse{
		Datetime: now.Format(time.RFC3339),
		Timezone: tzName,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
