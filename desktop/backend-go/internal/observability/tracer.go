package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

// InitTracer initializes the OpenTelemetry tracer provider and sets it globally.
// It exports traces to an OTEL collector via HTTP (OTLP protocol).
// Returns the TracerProvider for graceful shutdown on application exit.
func InitTracer(ctx context.Context, otelEndpoint string) (*trace.TracerProvider, error) {
	// Create HTTP exporter to send traces to the OpenTelemetry collector
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otelEndpoint),
	)
	if err != nil {
		slog.Error("failed to create OTLP exporter", "error", err)
		return nil, err
	}

	// Create resource describing this service
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("businessos"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		slog.Error("failed to create resource", "error", err)
		return nil, err
	}

	// Create TracerProvider with batching exporter
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	// Register W3C TraceContext + Baggage propagators so that traceparent headers
	// are extracted from inbound requests and injected into outbound requests.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("OpenTelemetry tracer initialized", "endpoint", otelEndpoint)

	return tp, nil
}

// ShutdownTracer gracefully shuts down the tracer provider and flushes any pending spans.
func ShutdownTracer(ctx context.Context, tp *trace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	if err := tp.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown tracer provider", "error", err)
		return err
	}
	slog.Info("OpenTelemetry tracer shutdown complete")
	return nil
}
