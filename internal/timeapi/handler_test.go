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

func TestTimeDefaultUTC(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	assertOK(t, rec)
	assertNoDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)

	want := "2026-06-03T14:30:45Z"
	if body.Datetime != want {
		t.Errorf("datetime = %q, want %q", body.Datetime, want)
	}
	if body.Timezone != "UTC" {
		t.Errorf("timezone = %q, want UTC", body.Timezone)
	}
	if body.Epoch != nil {
		t.Error("expected epoch to be omitted")
	}
}

func TestTimeFormatUnix(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?format=unix", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	assertOK(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)

	if body.Epoch == nil || *body.Epoch != testInstant.Unix() {
		t.Errorf("epoch = %v, want %d", body.Epoch, testInstant.Unix())
	}
}

func TestTimeFormatEpochAlias(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?format=epoch", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	assertOK(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Epoch == nil {
		t.Fatal("expected epoch field")
	}
}

func TestTimeTZ(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?tz=America/New_York", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	assertOK(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)

	want := "2026-06-03T10:30:45-04:00"
	if body.Datetime != want {
		t.Errorf("datetime = %q, want %q", body.Datetime, want)
	}
	if body.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want America/New_York", body.Timezone)
	}
}

func TestTimeFormatUnixAndTZ(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?format=unix&tz=Europe/London", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	assertOK(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)

	want := "2026-06-03T15:30:45+01:00"
	if body.Datetime != want {
		t.Errorf("datetime = %q, want %q", body.Datetime, want)
	}
	if body.Epoch == nil {
		t.Fatal("expected epoch field")
	}
}

func TestTimeInvalidTZ(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?tz=Not/A/Zone", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTimeInvalidFormat(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?format=csv", nil)
	rec := httptest.NewRecorder()

	h.Time(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNowUTCDeprecated(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.NowUTC(rec, req)

	assertOK(t, rec)
	assertDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Datetime != "2026-06-03T14:30:45Z" {
		t.Errorf("datetime = %q", body.Datetime)
	}
}

func TestEpochDeprecated(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/epoch", nil)
	rec := httptest.NewRecorder()

	h.Epoch(rec, req)

	assertOK(t, rec)
	assertDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Epoch == nil || *body.Epoch != testInstant.Unix() {
		t.Errorf("epoch = %v, want %d", body.Epoch, testInstant.Unix())
	}
}

func TestTimezoneDeprecated(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/TZ/America/New_York", nil)
	req.SetPathValue("tz", "America/New_York")
	rec := httptest.NewRecorder()

	h.Timezone(rec, req)

	assertOK(t, rec)
	assertDeprecation(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
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
	assertDeprecation(t, rec)
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

func assertOK(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func assertDeprecation(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Header().Get("Deprecation") != "true" {
		t.Error("expected Deprecation: true header on legacy route")
	}
	if rec.Header().Get("Link") == "" {
		t.Error("expected Link successor-version header on legacy route")
	}
}

func assertNoDeprecation(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Header().Get("Deprecation") != "" {
		t.Error("did not expect Deprecation header on /time")
	}
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, body any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
