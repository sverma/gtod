# gtod — Go time sample service

A minimal HTTP service for exercising CI/CD pipelines and GitOps delivery. It exposes the current time in several formats.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Current date/time in UTC as ISO-8601 (RFC3339) |
| `GET` | `/epoch` | Current UTC date/time (ISO-8601) and Unix epoch seconds |
| `GET` | `/TZ/{tz...}` | Current date/time in the given [IANA timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). Use a path with slashes, e.g. `/TZ/Europe/London` |

All successful responses are JSON with `Content-Type: application/json`.

### Examples

```bash
curl -s http://localhost:8080/
# {"datetime":"2026-06-03T14:30:45Z","timezone":"UTC"}

curl -s http://localhost:8080/epoch
# {"datetime":"2026-06-03T14:30:45Z","epoch":1748958645,"timezone":"UTC"}

curl -s http://localhost:8080/TZ/Asia/Tokyo
# {"datetime":"2026-06-03T23:30:45+09:00","timezone":"Asia/Tokyo"}
```

Invalid timezones return `400` with `{"error":"invalid timezone: ..."}`.

## Requirements

- Go 1.22 or newer (uses `net/http` path patterns and `PathValue`)

## Build

From the repository root:

```bash
go build -o bin/timeserver ./cmd/timeserver
```

## Run

```bash
# Default port 8080
./bin/timeserver

# Or without building
go run ./cmd/timeserver

# Custom port
PORT=3000 go run ./cmd/timeserver
```

## Test

Run all unit tests with coverage:

```bash
go test ./... -v -cover
```

Run tests once (CI-friendly, no verbose output):

```bash
go test ./... -cover -count=1
```

Run tests for the handler package only:

```bash
go test ./internal/timeapi/... -v
```

## Project layout

```
.
├── cmd/timeserver/     # HTTP server entrypoint
├── internal/timeapi/   # Handlers and clock abstraction
├── go.mod
└── README.md
```

## CI/CD notes

Typical pipeline steps:

1. `go mod download`
2. `go test ./... -cover -count=1`
3. `go build -o timeserver ./cmd/timeserver`
4. Build and push a container image, then deploy via your GitOps tool (Argo CD, Flux, etc.)

The service listens on `PORT` (default `8080`) and has no external dependencies beyond the Go standard library.
