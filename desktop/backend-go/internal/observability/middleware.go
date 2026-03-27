package observability

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// correlationIDContextKey is the key used to store X-Correlation-ID in context.
type correlationIDContextKey struct{}

var tracer = otel.Tracer("businessos")

// TracingMiddleware is a Gin middleware that creates a span for each request.
// It extracts the upstream W3C traceparent header (if any) to continue a
// distributed trace, reads or generates an X-Correlation-ID, and stores it
// on the span using the chatmangpt.run.correlation_id semconv attribute.
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract upstream trace context (W3C traceparent / tracestate / baggage)
		// so that incoming spans are linked to the same distributed trace.
		propagatedCtx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		ctx, span := tracer.Start(propagatedCtx, c.Request.URL.Path)
		defer span.End()

		// Add attributes to the span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.host", c.Request.Host),
		)

		// Extract X-Correlation-ID and attach it as a semconv attribute.
		// Downstream handlers and outbound requests should propagate this header.
		correlationID := c.Request.Header.Get("X-Correlation-ID")
		if correlationID != "" {
			span.SetAttributes(attribute.String("chatmangpt.run.correlation_id", correlationID))
			// Store in context so gateway handlers can forward it to pm4py-rust.
			ctx = context.WithValue(ctx, correlationIDContextKey{}, correlationID)
		}

		// Update the request context with the new span (and correlation_id if set)
		c.Request = c.Request.WithContext(ctx)

		// Continue with the request
		c.Next()

		// Update span with response status
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)

		// Mark span as error if status is 4xx or 5xx
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP "+string(rune(c.Writer.Status())))
		}
	}
}

// CorrelationIDFromContext retrieves the X-Correlation-ID stored by TracingMiddleware.
// Returns empty string if none was set.
func CorrelationIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDContextKey{}).(string); ok {
		return v
	}
	return ""
}

// StartSpan creates a new span with the given name in the request context.
// Use this in handler functions to instrument specific operations.
func StartSpan(c *gin.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(c.Request.Context(), name)
}

// RecordError records an error in the current span.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddSpanAttribute adds a key-value attribute to a span.
func AddSpanAttribute(span trace.Span, key string, value interface{}) {
	span.SetAttributes(attribute.String(key, toString(value)))
}

// Helper to convert interface{} to string for attributes
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
