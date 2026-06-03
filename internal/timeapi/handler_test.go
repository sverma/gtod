package timeapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time {
	return c.t
}

var testInstant = time.Date(2026, 6, 3, 14, 30, 45, 0, time.UTC)

func newTestHandler() *Handler {
	return NewHandlerWithClock(fixedClock{t: testInstant})
}

func TestNowUTC(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.NowUTC(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body utcResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := "2026-06-03T14:30:45Z"
	if body.Datetime != want {
		t.Errorf("datetime = %q, want %q", body.Datetime, want)
	}
	if body.Timezone != "UTC" {
		t.Errorf("timezone = %q, want UTC", body.Timezone)
	}
}

func TestEpoch(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/epoch", nil)
	rec := httptest.NewRecorder()

	h.Epoch(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body epochResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	wantDatetime := "2026-06-03T14:30:45Z"
	if body.Datetime != wantDatetime {
		t.Errorf("datetime = %q, want %q", body.Datetime, wantDatetime)
	}
	if body.Epoch != testInstant.Unix() {
		t.Errorf("epoch = %d, want %d", body.Epoch, testInstant.Unix())
	}
	if body.Timezone != "UTC" {
		t.Errorf("timezone = %q, want UTC", body.Timezone)
	}
}

func TestTimezone(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/TZ/America/New_York", nil)
	req.SetPathValue("tz", "America/New_York")
	rec := httptest.NewRecorder()

	h.Timezone(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body tzResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// 14:30 UTC = 10:30 EDT (UTC-4) on 2026-06-03
	want := "2026-06-03T10:30:45-04:00"
	if body.Datetime != want {
		t.Errorf("datetime = %q, want %q", body.Datetime, want)
	}
	if body.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want America/New_York", body.Timezone)
	}
}

func TestTimezoneInvalid(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/TZ/Not/A/Zone", nil)
	req.SetPathValue("tz", "Not/A/Zone")
	rec := httptest.NewRecorder()

	h.Timezone(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTimezoneEmpty(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/TZ/", nil)
	rec := httptest.NewRecorder()

	h.Timezone(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
