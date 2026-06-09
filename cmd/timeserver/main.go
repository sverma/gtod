package main

import (
	"log"
	"net/http"
	"os"

	"gtod/internal/observability"
	"gtod/internal/timeapi"
	"gtod/internal/weatherclient"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	metrics := observability.New(observability.BuildInfoFromEnv())
	weather := weatherclient.NewFromEnv()
	h := timeapi.NewHandler().WithMetrics(metrics).WithWeatherClient(weather)
	if weather.Enabled() {
		log.Printf("weather integration enabled: %s", os.Getenv("WEATHER_SERVICE_URL"))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", metrics.Instrument(observability.RouteMetrics, func(w http.ResponseWriter, r *http.Request) {
		metrics.MetricsHandler().ServeHTTP(w, r)
	}))
	mux.HandleFunc("GET /health", metrics.Instrument(observability.RouteHealth, observability.HealthHandler))
	mux.HandleFunc("GET /ready", metrics.Instrument(observability.RouteReady, observability.ReadyHandler))
	mux.HandleFunc("GET /time/difference", metrics.Instrument(observability.RouteTimeDifference, h.TimeDifference))
	mux.HandleFunc("GET /time", metrics.Instrument(observability.RouteTime, h.Time))
	// Legacy routes (deprecated; see Deprecation and Link response headers).
	mux.HandleFunc("GET /{$}", metrics.Instrument(observability.RouteRoot, h.NowUTC))
	mux.HandleFunc("GET /epoch", metrics.Instrument(observability.RouteEpoch, h.Epoch))
	mux.HandleFunc("GET /TZ/{tz...}", metrics.Instrument(observability.RouteTZ, h.Timezone))

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
