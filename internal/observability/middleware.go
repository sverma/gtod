package observability

import (
	"net/http"
	"time"
)

// Route name constants for low-cardinality metric labels.
const (
	RouteTime           = "/time"
	RouteTimeDifference = "/time/difference"
	RouteRoot           = "/"
	RouteEpoch          = "/epoch"
	RouteTZ             = "/TZ/{tz...}"
	RouteMetrics        = "/metrics"
	RouteHealth         = "/health"
	RouteReady          = "/ready"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Instrument wraps a handler with RED metrics for the given route template.
func (c *Collector) Instrument(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.requestsInFlight.Inc()
		defer c.requestsInFlight.Dec()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)

		elapsed := time.Since(start).Seconds()
		statusClass := classifyStatus(rec.status)

		c.requestsTotal.WithLabelValues(r.Method, route, statusClass).Inc()
		c.requestDuration.WithLabelValues(r.Method, route).Observe(elapsed)

		if rec.status >= 400 {
			errorType := "client"
			if rec.status >= 500 {
				errorType = "server"
			}
			c.requestErrorsTotal.WithLabelValues(r.Method, route, errorType).Inc()
		}
	}
}

func classifyStatus(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	default:
		return "2xx"
	}
}
