// Code generated from semconv/model/consensus/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 11

package semconv

import "go.opentelemetry.io/otel/attribute"

// Consensus quorum health and replica attributes (iter11).

const (
	// ConsensusQuorumHealthKey is the OTel attribute key for consensus.quorum.health.
	// Health status of the consensus quorum.
	ConsensusQuorumHealthKey = attribute.Key("consensus.quorum.health")
	// ConsensusBlockHeightKey is the OTel attribute key for consensus.block.height.
	// Current block height in the consensus chain.
	ConsensusBlockHeightKey = attribute.Key("consensus.block.height")
	// ConsensusReplicaCountKey is the OTel attribute key for consensus.replica.count.
	// Number of replicas participating in this consensus round.
	ConsensusReplicaCountKey = attribute.Key("consensus.replica.count")
	// ConsensusFailureCountKey is the OTel attribute key for consensus.failure.count.
	// Number of Byzantine failures tolerated (f in 2f+1 quorum formula).
	ConsensusFailureCountKey = attribute.Key("consensus.failure.count")
)

// ConsensusQuorumHealth returns an attribute KeyValue for consensus.quorum.health.
func ConsensusQuorumHealth(val string) attribute.KeyValue {
	return ConsensusQuorumHealthKey.String(val)
}

// ConsensusBlockHeight returns an attribute KeyValue for consensus.block.height.
func ConsensusBlockHeight(val int64) attribute.KeyValue {
	return ConsensusBlockHeightKey.Int64(val)
}

// ConsensusReplicaCount returns an attribute KeyValue for consensus.replica.count.
func ConsensusReplicaCount(val int) attribute.KeyValue {
	return ConsensusReplicaCountKey.Int(val)
}

// ConsensusFailureCount returns an attribute KeyValue for consensus.failure.count.
func ConsensusFailureCount(val int) attribute.KeyValue {
	return ConsensusFailureCountKey.Int(val)
}
