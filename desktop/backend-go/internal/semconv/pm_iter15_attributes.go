package semconv

import "go.opentelemetry.io/otel/attribute"

// Process Mining Replay attributes (iter15)
const (
	ProcessMiningReplayEnabledTransitionsKey = attribute.Key("process.mining.replay.enabled_transitions")
	ProcessMiningReplayMissingTokensKey      = attribute.Key("process.mining.replay.missing_tokens")
	ProcessMiningReplayConsumedTokensKey     = attribute.Key("process.mining.replay.consumed_tokens")
	ProcessMiningCaseVariantIDKey            = attribute.Key("process.mining.case.variant_id")
)

func ProcessMiningReplayEnabledTransitions(val int) attribute.KeyValue {
	return ProcessMiningReplayEnabledTransitionsKey.Int(val)
}

func ProcessMiningReplayMissingTokens(val int) attribute.KeyValue {
	return ProcessMiningReplayMissingTokensKey.Int(val)
}

func ProcessMiningReplayConsumedTokens(val int) attribute.KeyValue {
	return ProcessMiningReplayConsumedTokensKey.Int(val)
}

func ProcessMiningCaseVariantID(val string) attribute.KeyValue {
	return ProcessMiningCaseVariantIDKey.String(val)
}
