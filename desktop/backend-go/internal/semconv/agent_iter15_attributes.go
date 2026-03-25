package semconv

import "go.opentelemetry.io/otel/attribute"

// Agent Memory Federation attributes (iter15)
const (
	AgentMemoryFederationIDKey        = attribute.Key("agent.memory.federation_id")
	AgentMemoryFederationPeerCountKey = attribute.Key("agent.memory.federation.peer_count")
	AgentMemorySyncLatencyMsKey       = attribute.Key("agent.memory.sync.latency_ms")
	AgentMemoryFederationVersionKey   = attribute.Key("agent.memory.federation.version")
)

func AgentMemoryFederationID(val string) attribute.KeyValue {
	return AgentMemoryFederationIDKey.String(val)
}

func AgentMemoryFederationPeerCount(val int) attribute.KeyValue {
	return AgentMemoryFederationPeerCountKey.Int(val)
}

func AgentMemorySyncLatencyMs(val int) attribute.KeyValue {
	return AgentMemorySyncLatencyMsKey.Int(val)
}

func AgentMemoryFederationVersion(val int) attribute.KeyValue {
	return AgentMemoryFederationVersionKey.Int(val)
}
