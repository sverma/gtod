package observability

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandlerExposesREDMetrics(t *testing.T) {
	c := New(BuildInfo{Version: "test", GoVersion: "go1.24", GitCommit: "abc"})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", c.Instrument(RouteMetrics, func(w http.ResponseWriter, r *http.Request) {
		c.MetricsHandler().ServeHTTP(w, r)
	}))
	mux.HandleFunc("GET /ok", c.Instrument("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mux.HandleFunc("GET /bad", c.Instrument("/bad", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	okReq := httptest.NewRequest(http.MethodGet, "/ok", nil)
	okRec := httptest.NewRecorder()
	mux.ServeHTTP(okRec, okReq)

	badReq := httptest.NewRequest(http.MethodGet, "/bad", nil)
	badRec := httptest.NewRecorder()
	mux.ServeHTTP(badRec, badReq)

	c.RecordTimezoneLookupError("invalid_tz")
	c.RecordDeprecatedRoute("/epoch")

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	mux.ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", metricsRec.Code, http.StatusOK)
	}

	body, err := io.ReadAll(metricsRec.Body)
	if err != nil {
		t.Fatalf("read metrics body: %v", err)
	}
	output := string(body)

	for _, want := range []string{
		"http_requests_total",
		"http_request_errors_total",
		"http_request_duration_seconds",
		"http_requests_in_flight",
		"gtod_timezone_lookup_errors_total",
		"gtod_deprecated_route_requests_total",
		"gtod_build_info",
		"gtod_process_start_time_seconds",
		`route="/ok"`,
		`error_type="client"`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
}

func TestBuildInfoFromEnv(t *testing.T) {
	t.Setenv("VERSION", "1.2.3")
	t.Setenv("GIT_COMMIT", "deadbeef")

	info := BuildInfoFromEnv()
	if info.Version != "1.2.3" || info.GitCommit != "deadbeef" {
		t.Fatalf("info = %+v", info)
	}
	if info.GoVersion == "" {
		t.Fatal("expected go version")
	}
}
