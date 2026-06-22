package timeapi

import (
	"net/http"
	"strings"
	"time"
)

type zoneDatetime struct {
	Timezone string `json:"timezone"`
	Datetime string `json:"datetime"`
}

type timeConvertResponse struct {
	Instant string       `json:"instant"`
	From    zoneDatetime `json:"from"`
	To      zoneDatetime `json:"to"`
}

// TimeConvert handles GET /time/convert — express one instant in two IANA timezones.
//
// Query parameters:
//   - to: target IANA timezone (required)
//   - from: source IANA timezone (default UTC)
//   - at: instant to convert (RFC3339); defaults to now
func (h *Handler) TimeConvert(w http.ResponseWriter, r *http.Request) {
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	at := r.URL.Query().Get("at")

	resp, errMsg, status := h.buildTimeConvert(from, to, at)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) buildTimeConvert(from, to, at string) (timeConvertResponse, string, int) {
	if strings.TrimSpace(to) == "" {
		return timeConvertResponse{}, "to is required", http.StatusBadRequest
	}

	if strings.TrimSpace(from) == "" {
		from = "UTC"
	}

	instant, errMsg, status := parseReferenceInstant(at, h.clock)
	if errMsg != "" {
		return timeConvertResponse{}, errMsg, status
	}

	fromLoc, fromName, errMsg, status := resolveNamedLocation(from)
	if errMsg != "" {
		h.recordTimezoneErrorFromMessage(errMsg)
		return timeConvertResponse{}, errMsg, status
	}

	toLoc, toName, errMsg, status := resolveNamedLocation(to)
	if errMsg != "" {
		h.recordTimezoneErrorFromMessage(errMsg)
		return timeConvertResponse{}, errMsg, status
	}

	return timeConvertResponse{
		Instant: instant.UTC().Format(time.RFC3339),
		From: zoneDatetime{
			Timezone: fromName,
			Datetime: instant.In(fromLoc).Format(time.RFC3339),
		},
		To: zoneDatetime{
			Timezone: toName,
			Datetime: instant.In(toLoc).Format(time.RFC3339),
		},
	}, "", http.StatusOK
}
