package semconv

import "go.opentelemetry.io/otel/attribute"

// Consensus Liveness attributes (iter15)
const (
	ConsensusLivenessProofRoundsKey  = attribute.Key("consensus.liveness.proof_rounds")
	ConsensusNetworkRecoveryMsKey    = attribute.Key("consensus.network.recovery_ms")
	ConsensusViewDurationMsKey       = attribute.Key("consensus.view.duration_ms")
)

func ConsensusLivenessProofRounds(val int) attribute.KeyValue {
	return ConsensusLivenessProofRoundsKey.Int(val)
}

func ConsensusNetworkRecoveryMs(val int) attribute.KeyValue {
	return ConsensusNetworkRecoveryMsKey.Int(val)
}

func ConsensusViewDurationMs(val int) attribute.KeyValue {
	return ConsensusViewDurationMsKey.Int(val)
}
