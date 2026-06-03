package main

import (
	"log"
	"net/http"
	"os"

	"gtod/internal/timeapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := timeapi.NewHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.NowUTC)
	mux.HandleFunc("GET /epoch", h.Epoch)
	// {tz...} captures multi-segment IANA names such as Europe/London.
	mux.HandleFunc("GET /TZ/{tz...}", h.Timezone)

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
