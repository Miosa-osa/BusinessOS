package semconv

import "go.opentelemetry.io/otel/attribute"

// Signal Quality attributes (iter15)
const (
	SignalQualityScoreKey    = attribute.Key("signal.quality.score")
	SignalQualityDegradedKey = attribute.Key("signal.quality.degraded")
	SignalRetryCountKey      = attribute.Key("signal.retry.count")
)

func SignalQualityScore(val float64) attribute.KeyValue {
	return SignalQualityScoreKey.Float64(val)
}

func SignalQualityDegraded(val bool) attribute.KeyValue {
	return SignalQualityDegradedKey.Bool(val)
}

func SignalRetryCount(val int) attribute.KeyValue {
	return SignalRetryCountKey.Int(val)
}
