# gtod — Go time sample service

A minimal HTTP service for exercising CI/CD pipelines and GitOps delivery. It exposes the current time in several formats.

## Primary API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/time` | Current time (default: UTC, ISO-8601 / RFC3339) |

### Query parameters

| Parameter | Default | Values | Description |
|-----------|---------|--------|-------------|
| `tz` | `UTC` | IANA timezone name | e.g. `Europe/London`, `America/New_York` |
| `format` | `iso` | `iso`, `rfc3339`, `unix`, `epoch` | `unix` or `epoch` adds `"epoch"` (Unix seconds) to the JSON |

All successful responses use the same JSON shape:

```json
{"datetime":"2026-06-03T14:30:45Z","timezone":"UTC"}
```

With `format=unix` (or `format=epoch`):

```json
{"datetime":"2026-06-03T14:30:45Z","timezone":"UTC","epoch":1748958645}
```

### Examples

```bash
curl -s http://localhost:8080/time
# {"datetime":"2026-06-03T14:30:45Z","timezone":"UTC"}

curl -s "http://localhost:8080/time?format=unix"
# {"datetime":"...","timezone":"UTC","epoch":1748958645}

curl -s "http://localhost:8080/time?tz=Asia/Tokyo"
# {"datetime":"2026-06-03T23:30:45+09:00","timezone":"Asia/Tokyo"}

curl -s "http://localhost:8080/time?format=unix&tz=Europe/London"
# {"datetime":"...","timezone":"Europe/London","epoch":...}
```

Invalid `tz` or `format` values return `400` with `{"error":"..."}`.

## Deprecated routes (backward compatible)

These routes still work but return `Deprecation: true` and `Link: </time>; rel="successor-version"`. Prefer `/time` with query parameters.

| Legacy path | Replacement |
|-------------|-------------|
| `GET /` | `GET /time` |
| `GET /epoch` | `GET /time?format=unix` |
| `GET /TZ/{tz...}` | `GET /time?tz={tz}` (e.g. `/time?tz=Europe/London`) |

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
