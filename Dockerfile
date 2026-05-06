# ── Build stage ──────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=docker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION}" -o /app/server ./cmd/server

# ── Runtime stage ────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl
RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/internal/database/migrations ./migrations

USER appuser

EXPOSE 8081

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/ready || exit 1

CMD ["./server"]
