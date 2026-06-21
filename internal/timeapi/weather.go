package timeapi

import (
	"context"
	"time"

	"gtod/internal/weatherclient"
)

// Weather status values included when WEATHER_SERVICE_URL is configured.
const (
	WeatherStatusOK          = "ok"
	WeatherStatusUnavailable = "unavailable"
)

type weatherLookup struct {
	forecast *weatherclient.Forecast
	status   string
}

func (h *Handler) fetchWeather(ctx context.Context, tz string, at time.Time) weatherLookup {
	if h.weather == nil || !h.weather.Enabled() {
		return weatherLookup{}
	}

	forecast, err := h.weather.Lookup(ctx, tz, at.UTC().Format(time.RFC3339))
	if err != nil {
		return weatherLookup{status: WeatherStatusUnavailable}
	}
	return weatherLookup{forecast: &forecast, status: WeatherStatusOK}
}
