package semconv

import "go.opentelemetry.io/otel/attribute"

// Healing Self-Healing attributes (iter15)
const (
	HealingSelfHealingEnabledKey      = attribute.Key("healing.self_healing.enabled")
	HealingSelfHealingTriggerCountKey = attribute.Key("healing.self_healing.trigger_count")
	HealingSelfHealingSuccessRateKey  = attribute.Key("healing.self_healing.success_rate")
	HealingInterventionTypeKey        = attribute.Key("healing.intervention.type")
)

func HealingSelfHealingEnabled(val bool) attribute.KeyValue {
	return HealingSelfHealingEnabledKey.Bool(val)
}

func HealingSelfHealingTriggerCount(val int) attribute.KeyValue {
	return HealingSelfHealingTriggerCountKey.Int(val)
}

func HealingSelfHealingSuccessRate(val float64) attribute.KeyValue {
	return HealingSelfHealingSuccessRateKey.Float64(val)
}

func HealingInterventionType(val string) attribute.KeyValue {
	return HealingInterventionTypeKey.String(val)
}
