# ========================= 
# Builder Stage 
# =========================
FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /build 
COPY go.mod ./
RUN go mod download 
COPY . . 
RUN go build -o timeserver ./cmd/timeserver

# ========================= 
# Runtime Stage 
# =========================

FROM alpine:latest
RUN apk add --no-cache tzdata
COPY --from=builder /build/timeserver /usr/local/bin/timeserver

USER 65532:65532

EXPOSE 8080
CMD ["timeserver"]