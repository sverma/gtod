package timeapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gtod/internal/weatherclient"
)

type stubWeatherClient struct {
	forecast weatherclient.Forecast
}

func (s stubWeatherClient) Enabled() bool { return true }

func (s stubWeatherClient) Lookup(_ context.Context, tz, _ string) (weatherclient.Forecast, error) {
	f := s.forecast
	f.Timezone = tz
	return f, nil
}

func TestTimeIncludesWeather(t *testing.T) {
	h := newTestHandler().WithWeatherClient(stubWeatherClient{
		forecast: weatherclient.Forecast{
			Location:   "London",
			Provider:   "stub",
			Conditions: weatherclient.Conditions{Summary: "Sunny"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/time?tz=Europe/London", nil)
	rec := httptest.NewRecorder()
	h.Time(rec, req)

	assertOK(t, rec)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Weather == nil {
		t.Fatal("expected weather in response")
	}
	if body.Weather.Location != "London" {
		t.Errorf("weather.location = %q", body.Weather.Location)
	}
}

func TestTimeDifferenceIncludesWeather(t *testing.T) {
	h := newTestHandler().WithWeatherClient(stubWeatherClient{
		forecast: weatherclient.Forecast{Provider: "stub"},
	})

	req := httptest.NewRequest(http.MethodGet, "/time/difference?from=Europe/London&to=Asia/Tokyo", nil)
	rec := httptest.NewRecorder()
	h.TimeDifference(rec, req)

	assertOK(t, rec)

	var body timeDifferenceResponse
	decodeBody(t, rec, &body)
	if body.From.Weather == nil || body.To.Weather == nil {
		t.Fatal("expected weather on from and to zones")
	}
	if body.From.Weather.Timezone != "Europe/London" {
		t.Errorf("from weather tz = %q", body.From.Weather.Timezone)
	}
	if body.To.Weather.Timezone != "Asia/Tokyo" {
		t.Errorf("to weather tz = %q", body.To.Weather.Timezone)
	}
}

func TestTimeOmitsWeatherWhenClientDisabled(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/time?tz=UTC", nil)
	rec := httptest.NewRecorder()
	h.Time(rec, req)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Weather != nil {
		t.Fatal("expected no weather without client")
	}
}

func TestTimeIntegrationWithWeatherServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(weatherclient.Forecast{
			Timezone: r.URL.Query().Get("tz"),
			Location: "Test City",
			Provider: "stub",
		})
	}))
	defer server.Close()

	h := newTestHandler().WithWeatherClient(&weatherclient.HTTPClient{BaseURL: server.URL})
	req := httptest.NewRequest(http.MethodGet, "/time?tz=Asia/Tokyo", nil)
	rec := httptest.NewRecorder()
	h.Time(rec, req)

	var body timeResponse
	decodeBody(t, rec, &body)
	if body.Weather == nil || body.Weather.Location != "Test City" {
		t.Fatalf("weather = %+v", body.Weather)
	}
}
