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
	// process_mining_prediction_make is the span name for "process.mining.prediction.make".
	//
	// Span emitted when predictive analytics (next activity, remaining time, outcome) is computed
	// Kind: internal
	// Stability: development
	ProcessMiningPredictionMakeSpan = "process.mining.prediction.make"
	// process_mining_social_network_analyze is the span name for "process.mining.social_network.analyze".
	//
	// Span emitted when organizational/social network analysis is performed
	// Kind: internal
	// Stability: development
	ProcessMiningSocialNetworkAnalyzeSpan = "process.mining.social_network.analyze"
)