package semconv

const (
	// signal_batch_aggregate is the span name for "signal.batch.aggregate".
	//
	// Batch aggregation of signals — collecting signals within a time window and processing them as a group.
	// Kind: internal
	// Stability: development
	SignalBatchAggregateSpan = "signal.batch.aggregate"
	// signal_classify is the span name for "signal.classify".
	//
	// Classifies a signal's mode, genre, and type according to Signal Theory S=(M,G,T,F,W).
	// Kind: internal
	// Stability: development
	SignalClassifySpan = "signal.classify"
	// signal_compress is the span name for "signal.compress".
	//
	// Compressing a signal payload before transmission — bandwidth optimization.
	// Kind: internal
	// Stability: development
	SignalCompressSpan = "signal.compress"
	// signal_decode is the span name for "signal.decode".
	//
	// Signal deserialization — decoding a received signal payload from its wire format.
	// Kind: internal
	// Stability: development
	SignalDecodeSpan = "signal.decode"
	// signal_encode is the span name for "signal.encode".
	//
	// Encoding of a signal using the S=(M,G,T,F,W) Signal Theory model.
	// Kind: internal
	// Stability: development
	SignalEncodeSpan = "signal.encode"
	// signal_filter is the span name for "signal.filter".
	//
	// Applies the S/N gate to filter noise — signals below the weight threshold are rejected.
	// Kind: internal
	// Stability: development
	SignalFilterSpan = "signal.filter"
	// signal_quality_assess is the span name for "signal.quality.assess".
	//
	// Assessing the composite quality of a signal against acceptance thresholds.
	// Kind: internal
	// Stability: development
	SignalQualityAssessSpan = "signal.quality.assess"
	// signal_route is the span name for "signal.route".
	//
	// Signal routing decision — determining which service or agent receives this signal.
	// Kind: internal
	// Stability: development
	SignalRouteSpan = "signal.route"
	// signal_sn_gate is the span name for "signal.sn_gate".
	//
	// Signal quality gate — filters signals below S/N ratio threshold.
	// Kind: internal
	// Stability: development
	SignalSnGateSpan = "signal.sn_gate"
)