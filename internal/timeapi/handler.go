package timeapi

import (
	"encoding/json"
	"net/http"
	"strings"
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

// timeResponse is the unified JSON body for GET /time and legacy routes.
type timeResponse struct {
	Datetime string `json:"datetime"`
	Timezone string `json:"timezone"`
	Epoch    *int64 `json:"epoch,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Time handles GET /time — primary API.
//
// Query parameters:
//   - tz: IANA timezone (default UTC), e.g. tz=Europe/London
//   - format: empty or "iso" for RFC3339 only; "unix" or "epoch" to include epoch seconds
func (h *Handler) Time(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	tz := r.URL.Query().Get("tz")

	resp, errMsg, status := h.buildTime(format, tz)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// NowUTC handles GET / (deprecated; use GET /time).
func (h *Handler) NowUTC(w http.ResponseWriter, r *http.Request) {
	h.writeLegacy(w, "", "")
}

// Epoch handles GET /epoch (deprecated; use GET /time?format=unix).
func (h *Handler) Epoch(w http.ResponseWriter, r *http.Request) {
	h.writeLegacy(w, "unix", "")
}

// Timezone handles GET /TZ/{tz...} (deprecated; use GET /time?tz=...).
func (h *Handler) Timezone(w http.ResponseWriter, r *http.Request) {
	tzName := r.PathValue("tz")
	if tzName == "" {
		setDeprecation(w)
		writeError(w, http.StatusBadRequest, "timezone is required")
		return
	}
	h.writeLegacy(w, "", tzName)
}

func (h *Handler) writeLegacy(w http.ResponseWriter, format, tz string) {
	setDeprecation(w)
	resp, errMsg, status := h.buildTime(format, tz)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) buildTime(format, tz string) (timeResponse, string, int) {
	includeEpoch, errMsg, status := parseFormat(format)
	if errMsg != "" {
		return timeResponse{}, errMsg, status
	}

	loc, tzName, errMsg, status := resolveLocation(tz)
	if errMsg != "" {
		return timeResponse{}, errMsg, status
	}

	now := h.clock.Now().In(loc)
	resp := timeResponse{
		Datetime: now.Format(time.RFC3339),
		Timezone: tzName,
	}
	if includeEpoch {
		epoch := now.Unix()
		resp.Epoch = &epoch
	}
	return resp, "", http.StatusOK
}

func parseFormat(format string) (includeEpoch bool, errMsg string, status int) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "iso", "rfc3339":
		return false, "", http.StatusOK
	case "unix", "epoch":
		return true, "", http.StatusOK
	default:
		return false, "invalid format: " + format + ` (use "iso" or "unix")`, http.StatusBadRequest
	}
}

func resolveLocation(tz string) (*time.Location, string, string, int) {
	tz = strings.TrimSpace(tz)
	if tz == "" || strings.EqualFold(tz, "UTC") {
		return time.UTC, "UTC", "", http.StatusOK
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, "", "invalid timezone: " + tz, http.StatusBadRequest
	}
	return loc, tz, "", http.StatusOK
}

func setDeprecation(w http.ResponseWriter) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Link", `</time>; rel="successor-version"`)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
