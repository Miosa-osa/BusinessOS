// Code generated from semconv/model/a2a/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 11

package semconv

import "go.opentelemetry.io/otel/attribute"

// A2A SLA and retry tracking attributes (iter11).

const (
	// A2ASlaDeadlineMsKey is the OTel attribute key for a2a.sla.deadline_ms.
	// SLA deadline in milliseconds since epoch for the A2A operation.
	A2ASlaDeadlineMsKey = attribute.Key("a2a.sla.deadline_ms")
	// A2ASlaBreachKey is the OTel attribute key for a2a.sla.breach.
	// Whether the A2A SLA was breached (true/false).
	A2ASlaBreachKey = attribute.Key("a2a.sla.breach")
	// A2ASlaLatencyMsKey is the OTel attribute key for a2a.sla.latency_ms.
	// Actual latency of the A2A operation in milliseconds, compared against SLA.
	A2ASlaLatencyMsKey = attribute.Key("a2a.sla.latency_ms")
	// A2ARetryCountKey is the OTel attribute key for a2a.retry.count.
	// Number of times the A2A operation was retried before success or final failure.
	A2ARetryCountKey = attribute.Key("a2a.retry.count")
)

// A2ASlaDeadlineMs returns an attribute KeyValue for a2a.sla.deadline_ms.
func A2ASlaDeadlineMs(val int64) attribute.KeyValue {
	return A2ASlaDeadlineMsKey.Int64(val)
}

// A2ASlaBreach returns an attribute KeyValue for a2a.sla.breach.
func A2ASlaBreach(val bool) attribute.KeyValue {
	return A2ASlaBreachKey.Bool(val)
}

// A2ASlaLatencyMs returns an attribute KeyValue for a2a.sla.latency_ms.
func A2ASlaLatencyMs(val int64) attribute.KeyValue {
	return A2ASlaLatencyMsKey.Int64(val)
}

// A2ARetryCount returns an attribute KeyValue for a2a.retry.count.
func A2ARetryCount(val int) attribute.KeyValue {
	return A2ARetryCountKey.Int(val)
}
