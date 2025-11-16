# Build stage
FROM golang:1.23.0-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o otlp-log-parser-assignment ./cmd

RUN go test -v ./...

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

RUN addgroup -g 1000 otlp && \
    adduser -D -u 1000 -G otlp otlp

WORKDIR /app

COPY --from=builder /build/otlp-log-parser-assignment .

RUN chown -R otlp:otlp /app

USER otlp

EXPOSE 4317 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 4317 && nc -z localhost 9090 || exit 1

ENV PORT=4317 \
    METRICS_PORT=9090 \
    ATTRIBUTE_KEY=service.name \
    WINDOW_DURATION=10s \
    DEBUG=false

ENTRYPOINT ["/app/otlp-log-parser-assignment"]
CMD []
