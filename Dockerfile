FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/call-notes-ai-service ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/bin/call-notes-ai-service .
COPY --from=builder /app/config ./config

USER appuser

EXPOSE 8080 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health/live || exit 1

ENTRYPOINT ["./call-notes-ai-service"]
