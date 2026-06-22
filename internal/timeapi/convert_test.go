package timeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTimeConvert(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/time/convert?from=UTC&to=Europe/London&at=2026-06-03T14:30:45Z", nil)
	rec := httptest.NewRecorder()

	h.TimeConvert(rec, req)

	assertOK(t, rec)

	var body timeConvertResponse
	decodeBody(t, rec, &body)

	if body.Instant != "2026-06-03T14:30:45Z" {
		t.Errorf("instant = %q", body.Instant)
	}
	if body.From.Datetime != "2026-06-03T14:30:45Z" {
		t.Errorf("from.datetime = %q", body.From.Datetime)
	}
	if body.To.Datetime != "2026-06-03T15:30:45+01:00" {
		t.Errorf("to.datetime = %q", body.To.Datetime)
	}
}

func TestTimeConvertDefaultFromUTC(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/time/convert?to=Asia/Tokyo&at=2026-06-03T14:30:45Z", nil)
	rec := httptest.NewRecorder()

	h.TimeConvert(rec, req)

	assertOK(t, rec)

	var body timeConvertResponse
	decodeBody(t, rec, &body)
	if body.From.Timezone != "UTC" {
		t.Errorf("from.timezone = %q, want UTC", body.From.Timezone)
	}
	if body.To.Datetime != "2026-06-03T23:30:45+09:00" {
		t.Errorf("to.datetime = %q", body.To.Datetime)
	}
}

func TestTimeConvertMissingTo(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/convert?from=UTC", nil)
	rec := httptest.NewRecorder()

	h.TimeConvert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTimeConvertInvalidTimezone(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time/convert?to=Bad/Zone", nil)
	rec := httptest.NewRecorder()

	h.TimeConvert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMuxTimeConvert(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /time/convert", h.TimeConvert)
	mux.HandleFunc("GET /time", h.Time)

	req := httptest.NewRequest(http.MethodGet,
		"/time/convert?to=UTC&at=2026-06-03T14:30:45Z", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
