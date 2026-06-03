package timeapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxTimezoneMultiSegment(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /TZ/{tz...}", h.Timezone)

	req := httptest.NewRequest(http.MethodGet, "/TZ/Europe/London", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /TZ/Europe/London status = %d, want %d; body: %s",
			rec.Code, http.StatusOK, rec.Body.String())
	}

	var body tzResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Timezone != "Europe/London" {
		t.Errorf("timezone = %q, want Europe/London", body.Timezone)
	}
}
