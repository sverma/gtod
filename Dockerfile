FROM golang:1.25.9-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /build 
COPY go.mod ./
RUN go mod download 
COPY . . 
RUN go build -o timeserver ./cmd/timeserver
FROM alpine:latest
RUN apk add --no-cache tzdata
COPY --from=builder /build/timeserver /usr/local/bin/timeserver
EXPOSE 8080
CMD ["timeserver"]