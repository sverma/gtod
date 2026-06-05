package timeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMux() *http.ServeMux {
	h := newTestHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /time", h.Time)
	mux.HandleFunc("GET /{$}", h.NowUTC)
	mux.HandleFunc("GET /epoch", h.Epoch)
	mux.HandleFunc("GET /TZ/{tz...}", h.Timezone)
	return mux
}

func TestMuxTimeQueryTZ(t *testing.T) {
	mux := newTestMux()
	req := httptest.NewRequest(http.MethodGet, "/time?tz=Europe/London", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	assertNoDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Timezone != "Europe/London" {
		t.Errorf("timezone = %q, want Europe/London", body.Timezone)
	}
}

func TestMuxTimezoneMultiSegmentDeprecated(t *testing.T) {
	mux := newTestMux()
	req := httptest.NewRequest(http.MethodGet, "/TZ/Europe/London", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /TZ/Europe/London status = %d, want %d; body: %s",
			rec.Code, http.StatusOK, rec.Body.String())
	}
	assertDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Timezone != "Europe/London" {
		t.Errorf("timezone = %q, want Europe/London", body.Timezone)
	}
}
