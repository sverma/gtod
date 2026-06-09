package timeapi

import "gtod/internal/weatherclient"

// MetricsRecorder records application-level metrics for time handlers.
type MetricsRecorder interface {
	RecordTimezoneLookupError(reason string)
	RecordDeprecatedRoute(route string)
}

type noopMetricsRecorder struct{}

func (noopMetricsRecorder) RecordTimezoneLookupError(string) {}
func (noopMetricsRecorder) RecordDeprecatedRoute(string)     {}

// WithMetrics attaches a metrics recorder to the handler.
func (h *Handler) WithMetrics(recorder MetricsRecorder) *Handler {
	h.metrics = recorder
	return h
}

// WithWeatherClient attaches a weather service client to the handler.
func (h *Handler) WithWeatherClient(client weatherclient.Client) *Handler {
	h.weather = client
	return h
}
