package observability

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentA2AHandler wraps A2A handler functions with tracing.
// Usage in handlers/a2a.go:
//   ctx, span := observability.StartA2ASpan(c, "a2a.discover_agent")
//   defer span.End()
func StartA2ASpan(c *gin.Context, operation string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(c.Request.Context(), operation)
	span.SetAttributes(attribute.String("a2a.operation", operation))
	return ctx, span
}

// StartCRMSpan creates a span for CRM operations.
// Usage in handlers/crm_deals.go:
//   ctx, span := observability.StartCRMSpan(c, "crm.list_deals", "pipelineID", pipelineID)
//   defer span.End()
func StartCRMSpan(c *gin.Context, operation string, attributes ...string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(c.Request.Context(), operation)
	span.SetAttributes(attribute.String("crm.operation", operation))

	// Add additional attributes in key-value pairs
	for i := 0; i < len(attributes); i += 2 {
		if i+1 < len(attributes) {
			span.SetAttributes(attribute.String("crm."+attributes[i], attributes[i+1]))
		}
	}

	return ctx, span
}

// StartProjectSpan creates a span for project operations.
// Usage in handlers/projects_update.go:
//   ctx, span := observability.StartProjectSpan(c, "project.create", "userID", userID)
//   defer span.End()
func StartProjectSpan(c *gin.Context, operation string, attributes ...string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(c.Request.Context(), operation)
	span.SetAttributes(attribute.String("project.operation", operation))

	// Add additional attributes in key-value pairs
	for i := 0; i < len(attributes); i += 2 {
		if i+1 < len(attributes) {
			span.SetAttributes(attribute.String("project."+attributes[i], attributes[i+1]))
		}
	}

	return ctx, span
}

// StartAuditSpan creates a span for audit trail operations.
// Usage in handlers/audit_handlers.go:
//   ctx, span := observability.StartAuditSpan(c, "audit.create_entry", "entityType", "Deal")
//   defer span.End()
func StartAuditSpan(c *gin.Context, operation string, attributes ...string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(c.Request.Context(), operation)
	span.SetAttributes(attribute.String("audit.operation", operation))

	// Add additional attributes in key-value pairs
	for i := 0; i < len(attributes); i += 2 {
		if i+1 < len(attributes) {
			span.SetAttributes(attribute.String("audit."+attributes[i], attributes[i+1]))
		}
	}

	return ctx, span
}

// RecordDatabaseQuery records database operation details in a span.
func RecordDatabaseQuery(span trace.Span, query string, duration int64) {
	span.SetAttributes(
		attribute.String("db.operation", query),
		attribute.Int64("db.duration_ms", duration),
	)
}

// RecordHTTPError records HTTP error details in a span.
func RecordHTTPError(span trace.Span, statusCode int, errorMsg string) {
	span.SetAttributes(
		attribute.Int("http.error_code", statusCode),
		attribute.String("http.error_message", errorMsg),
	)
	RecordError(span, fmt.Errorf("HTTP %d: %s", statusCode, errorMsg))
}
