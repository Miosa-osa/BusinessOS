package semconv

const (
	// process_mining_canopy_ingest is the span name for "process.mining.canopy.ingest".
	//
	// Span emitted when CSV event data is ingested from Canopy
	// Kind: consumer
	// Stability: development
	ProcessMiningCanopyIngestSpan = "process.mining.canopy.ingest"
	// process_mining_declare_check is the span name for "process.mining.declare.check".
	//
	// Span emitted when declare constraint conformance is checked
	// Kind: internal
	// Stability: development
	ProcessMiningDeclareCheckSpan = "process.mining.declare.check"
	// prediction_make and social_network_analyze: see process_span_names.go (full process mining catalog).
)