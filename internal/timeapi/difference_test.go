package timeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTimeDifference(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=Europe/London&to=Asia/Tokyo", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	assertOK(t, rec)

	var body timeDifferenceResponse
	decodeBody(t, rec, &body)

	if body.ReferenceInstant != "2026-06-03T14:30:45Z" {
		t.Errorf("reference_instant = %q, want 2026-06-03T14:30:45Z", body.ReferenceInstant)
	}
	if body.From.Timezone != "Europe/London" || body.From.UTCOffsetSeconds != 3600 {
		t.Errorf("from = %+v, want Europe/London offset 3600", body.From)
	}
	if body.To.Timezone != "Asia/Tokyo" || body.To.UTCOffsetSeconds != 32400 {
		t.Errorf("to = %+v, want Asia/Tokyo offset 32400", body.To)
	}
	if body.DifferenceSeconds != 28800 {
		t.Errorf("difference_seconds = %d, want 28800", body.DifferenceSeconds)
	}
	if body.Difference != "+8h" {
		t.Errorf("difference = %q, want +8h", body.Difference)
	}
}

func TestTimeDifferenceWithAt(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/time/difference?from=America/New_York&to=Europe/London&at=2026-01-15T12:00:00Z", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	assertOK(t, rec)

	var body timeDifferenceResponse
	decodeBody(t, rec, &body)

	// EST (UTC-5) vs GMT (UTC+0) in January → London 5 hours ahead.
	if body.DifferenceSeconds != 18000 {
		t.Errorf("difference_seconds = %d, want 18000", body.DifferenceSeconds)
	}
	if body.Difference != "+5h" {
		t.Errorf("difference = %q, want +5h", body.Difference)
	}
}

func TestTimeDifferenceSameZone(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=UTC&to=UTC", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	assertOK(t, rec)

	var body timeDifferenceResponse
	decodeBody(t, rec, &body)
	if body.DifferenceSeconds != 0 || body.Difference != "+0h" {
		t.Errorf("difference = %d %q, want 0 +0h", body.DifferenceSeconds, body.Difference)
	}
}

func TestTimeDifferenceMissingParams(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=UTC", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTimeDifferenceInvalidTimezone(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=Bad/Zone&to=UTC", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTimeDifferenceInvalidAt(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=UTC&to=UTC&at=not-a-date", nil)
	rec := httptest.NewRecorder()

	h.TimeDifference(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFormatDifference(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{0, "+0h"},
		{28800, "+8h"},
		{-18000, "-5h"},
		{19800, "+5h30m"},
		{-9000, "-2h30m"},
	}
	for _, tc := range tests {
		if got := formatDifference(tc.seconds); got != tc.want {
			t.Errorf("formatDifference(%d) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}
