package weatherclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientLookup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/weather" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("tz") != "Europe/London" {
			t.Fatalf("tz = %q", r.URL.Query().Get("tz"))
		}
		if r.URL.Query().Get("at") != "2026-06-03T14:30:45Z" {
			t.Fatalf("at = %q", r.URL.Query().Get("at"))
		}
		_ = json.NewEncoder(w).Encode(Forecast{
			Timezone: "Europe/London",
			Location: "London",
			Provider: "stub",
		})
	}))
	defer server.Close()

	client := &HTTPClient{BaseURL: server.URL}
	forecast, err := client.Lookup(context.Background(), "Europe/London", "2026-06-03T14:30:45Z")
	if err != nil {
		t.Fatal(err)
	}
	if forecast.Location != "London" {
		t.Errorf("location = %q", forecast.Location)
	}
}

func TestNoopClientNotEnabled(t *testing.T) {
	client := noopClient{}
	if client.Enabled() {
		t.Fatal("expected noop client to be disabled")
	}
}

func TestNewFromEnvNoop(t *testing.T) {
	t.Setenv("WEATHER_SERVICE_URL", "")
	client := NewFromEnv()
	if client.Enabled() {
		t.Fatal("expected disabled client without env var")
	}
}
