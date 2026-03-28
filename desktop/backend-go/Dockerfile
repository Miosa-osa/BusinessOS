# =============================================================================
# BusinessOS Backend — Go server + optional OpenTelemetry Weaver (semconv MCP)
# =============================================================================

# Weaver CLI — official multi-arch image (source builds require a prebuilt UI tree).
FROM otel/weaver:v0.14.0 AS weaver

# ---------------------------------------------------------------------------
# Stage: Go build (static binary)
# ---------------------------------------------------------------------------
FROM golang:1.25-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
COPY internal/osasdk ./internal/osasdk/
RUN go mod download && go mod verify

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o /server ./cmd/server

# ---------------------------------------------------------------------------
# Stage: Runtime
# ---------------------------------------------------------------------------
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -r appgroup && useradd -r -g appgroup -d /app appuser

WORKDIR /app

COPY --from=builder /server /app/server
COPY --from=weaver /weaver/weaver /usr/local/bin/weaver

RUN chown -R appuser:appgroup /app && chmod +x /usr/local/bin/weaver

USER appuser

# Must match docker-compose.yml SERVER_PORT env and port mapping (8001).
ENV SERVER_PORT=8001

EXPOSE 8001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8001/healthz || exit 1

CMD ["/app/server"]
