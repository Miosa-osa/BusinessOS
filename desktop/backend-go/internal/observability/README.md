# OpenTelemetry Instrumentation for BusinessOS

This package provides distributed tracing capabilities via OpenTelemetry (OTEL).

## Overview

The observability module exports traces to an OTLP (OpenTelemetry Protocol) collector via HTTP, enabling:
- **Distributed tracing** across service boundaries
- **Latency analysis** for request paths
- **Error tracking** with automatic span status codes
- **Request correlation** for debugging multi-service flows

## Configuration

### Environment Variables

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317  # Default OTLP receiver (gRPC format)
```

In production, point this to your observability backend (Jaeger, Grafana Tempo, etc.):
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.monitoring:4317
```

## Usage

### Global Middleware (Automatic)

All HTTP requests are automatically traced via `TracingMiddleware()`, which is registered early in the middleware chain. Each request gets a span with:
- HTTP method and path
- Response status code
- Error markers for 4xx/5xx responses

### Handler-Level Instrumentation

Wrap specific operations with spans:

```go
// In handlers/a2a.go
func (h *A2AHandler) DiscoverAgent(c *gin.Context) {
    ctx, span := observability.StartA2ASpan(c, "a2a.discover_agent")
    defer span.End()

    // Your existing handler logic
    card, err := h.a2aClient.DiscoverAgent(ctx, req.AgentURL)
    if err != nil {
        observability.RecordError(span, err)
        return
    }

    c.JSON(http.StatusOK, card)
}
```

### Helper Functions

- `StartA2ASpan()` — A2A agent communication spans
- `StartCRMSpan()` — CRM operations (deals, contacts, etc.)
- `StartProjectSpan()` — Project management operations
- `StartAuditSpan()` — Audit trail operations
- `RecordError()` — Record exceptions in spans
- `RecordDatabaseQuery()` — Track database operations
- `RecordHTTPError()` — Track HTTP errors

### Example: CRM Deals Handler

```go
// Before:
func (h *CRMHandler) ListCRMDeals(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    if user == nil {
        utils.RespondUnauthorized(c, slog.Default())
        return
    }
    // ... rest of handler
}

// After:
func (h *CRMHandler) ListCRMDeals(c *gin.Context) {
    ctx, span := observability.StartCRMSpan(c, "crm.list_deals",
        "user_id", user.ID,
        "limit", strconv.Itoa(pg.Limit),
    )
    defer span.End()

    user := middleware.GetCurrentUser(c)
    if user == nil {
        utils.RespondUnauthorized(c, slog.Default())
        return
    }
    // ... rest of handler (use ctx from span)
}
```

## Span Attributes

Spans include structured attributes for filtering and analysis:

- **HTTP Layer**: `http.method`, `http.url`, `http.status_code`, `http.error_code`
- **A2A Layer**: `a2a.operation`, `a2a.agent_url`
- **CRM Layer**: `crm.operation`, `crm.pipeline_id`, `crm.deal_id`
- **Project Layer**: `project.operation`, `project.user_id`
- **Audit Layer**: `audit.operation`, `audit.entity_type`
- **Database Layer**: `db.operation`, `db.duration_ms`

## Integration with Observability Stack

### Local Development (Jaeger)

```bash
# Start Jaeger all-in-one
docker run -d \
  -p 4317:4317/udp \
  -p 16686:16686 \
  jaegertracing/all-in-one

# Start BusinessOS with OTEL enabled
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
go run cmd/server/main.go

# View traces at http://localhost:16686
```

### Production (Grafana Tempo)

```bash
# Deploy Tempo + Loki stack
# Configure collector endpoint in deployment
export OTEL_EXPORTER_OTLP_ENDPOINT=tempo-distributor.monitoring:4317

# Traces appear in Grafana Explore with service.name=businessos
```

## Testing

The tracer provider can be tested via:

```bash
cd /Users/sac/chatmangpt/BusinessOS/desktop/backend-go
go build ./internal/observability
```

## Files

- `tracer.go` — Tracer provider initialization and shutdown
- `middleware.go` — Gin middleware for automatic request tracing
- `handlers.go` — Helper functions for common handler patterns
- `README.md` — This documentation

## References

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [OTEL Protocol Specification](https://opentelemetry.io/docs/reference/specification/protocol/)
- [Jaeger Deployment](https://www.jaegertracing.io/docs/latest/deployment/)
- [Grafana Tempo](https://grafana.com/docs/tempo/latest/)
