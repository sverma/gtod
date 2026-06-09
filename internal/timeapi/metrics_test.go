package timeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordingMetrics struct {
	timezoneErrors []string
	deprecated     []string
}

func (m *recordingMetrics) RecordTimezoneLookupError(reason string) {
	m.timezoneErrors = append(m.timezoneErrors, reason)
}

func (m *recordingMetrics) RecordDeprecatedRoute(route string) {
	m.deprecated = append(m.deprecated, route)
}

func TestMetricsDeprecatedRoute(t *testing.T) {
	rec := &recordingMetrics{}
	h := newTestHandler().WithMetrics(rec)

	req := httptest.NewRequest(http.MethodGet, "/epoch", nil)
	h.Epoch(httptest.NewRecorder(), req)

	if len(rec.deprecated) != 1 || rec.deprecated[0] != "/epoch" {
		t.Fatalf("deprecated = %v, want [/epoch]", rec.deprecated)
	}
}

func TestMetricsTimezoneError(t *testing.T) {
	rec := &recordingMetrics{}
	h := newTestHandler().WithMetrics(rec)

	req := httptest.NewRequest(http.MethodGet, "/time?tz=Bad/Zone", nil)
	h.Time(httptest.NewRecorder(), req)

	if len(rec.timezoneErrors) != 1 || rec.timezoneErrors[0] != "invalid_tz" {
		t.Fatalf("timezoneErrors = %v, want [invalid_tz]", rec.timezoneErrors)
	}
}
