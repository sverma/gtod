package timeapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type zoneSnapshot struct {
	Timezone         string `json:"timezone"`
	Datetime         string `json:"datetime"`
	UTCOffsetSeconds int    `json:"utc_offset_seconds"`
}

type timeDifferenceResponse struct {
	ReferenceInstant  string       `json:"reference_instant"`
	From              zoneSnapshot `json:"from"`
	To                zoneSnapshot `json:"to"`
	DifferenceSeconds int64        `json:"difference_seconds"`
	Difference        string       `json:"difference"`
}

// TimeDifference handles GET /time/difference — UTC offset gap between two IANA zones.
//
// Query parameters:
//   - from: source IANA timezone (required)
//   - to: target IANA timezone (required)
//   - at: optional reference instant (RFC3339); defaults to now
//
// difference_seconds = to.utc_offset_seconds - from.utc_offset_seconds (positive means to is ahead).
func (h *Handler) TimeDifference(w http.ResponseWriter, r *http.Request) {
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	at := r.URL.Query().Get("at")

	resp, errMsg, status := h.buildTimeDifference(from, to, at)
	if errMsg != "" {
		writeError(w, status, errMsg)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) buildTimeDifference(from, to, at string) (timeDifferenceResponse, string, int) {
	if from == "" || to == "" {
		return timeDifferenceResponse{}, "from and to are required", http.StatusBadRequest
	}

	instant, errMsg, status := parseReferenceInstant(at, h.clock)
	if errMsg != "" {
		return timeDifferenceResponse{}, errMsg, status
	}

	fromLoc, fromName, errMsg, status := resolveNamedLocation(from)
	if errMsg != "" {
		return timeDifferenceResponse{}, errMsg, status
	}

	toLoc, toName, errMsg, status := resolveNamedLocation(to)
	if errMsg != "" {
		return timeDifferenceResponse{}, errMsg, status
	}

	fromSnap := zoneAtInstant(instant, fromLoc, fromName)
	toSnap := zoneAtInstant(instant, toLoc, toName)
	diff := int64(toSnap.UTCOffsetSeconds - fromSnap.UTCOffsetSeconds)

	return timeDifferenceResponse{
		ReferenceInstant:  instant.UTC().Format(time.RFC3339),
		From:              fromSnap,
		To:                toSnap,
		DifferenceSeconds: diff,
		Difference:        formatDifference(diff),
	}, "", http.StatusOK
}

func parseReferenceInstant(at string, clock Clock) (time.Time, string, int) {
	at = strings.TrimSpace(at)
	if at == "" {
		return clock.Now().UTC(), "", http.StatusOK
	}

	instant, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return time.Time{}, "invalid at: " + at, http.StatusBadRequest
	}
	return instant.UTC(), "", http.StatusOK
}

func resolveNamedLocation(tz string) (*time.Location, string, string, int) {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return nil, "", "timezone is required", http.StatusBadRequest
	}
	if strings.EqualFold(tz, "UTC") {
		return time.UTC, "UTC", "", http.StatusOK
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, "", "invalid timezone: " + tz, http.StatusBadRequest
	}
	return loc, tz, "", http.StatusOK
}

func zoneAtInstant(instant time.Time, loc *time.Location, tzName string) zoneSnapshot {
	inLoc := instant.In(loc)
	_, offset := inLoc.Zone()
	return zoneSnapshot{
		Timezone:         tzName,
		Datetime:         inLoc.Format(time.RFC3339),
		UTCOffsetSeconds: offset,
	}
}

func formatDifference(seconds int64) string {
	if seconds == 0 {
		return "+0h"
	}

	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if minutes == 0 {
		return fmt.Sprintf("%s%dh", sign, hours)
	}
	return fmt.Sprintf("%s%dh%dm", sign, hours, minutes)
}
