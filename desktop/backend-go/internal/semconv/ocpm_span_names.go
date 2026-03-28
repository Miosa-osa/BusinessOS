package semconv

const (
	// ocpm_conformance_check is the span name for "ocpm.conformance.check".
	//
	// Object-Centric token replay conformance check
	// Kind: internal
	// Stability: development
	OcpmConformanceCheckSpan = "ocpm.conformance.check"
	// ocpm_discovery_dfg is the span name for "ocpm.discovery.dfg".
	//
	// Object-Centric DFG discovery from an OCEL 2.0 log
	// Kind: internal
	// Stability: development
	OcpmDiscoveryDfgSpan = "ocpm.discovery.dfg"
	// ocpm_discovery_petri_net is the span name for "ocpm.discovery.petri_net".
	//
	// Object-Centric Petri Net discovery from an OCEL 2.0 log
	// Kind: internal
	// Stability: development
	OcpmDiscoveryPetriNetSpan = "ocpm.discovery.petri_net"
	// ocpm_llm_query is the span name for "ocpm.llm.query".
	//
	// OCEL-grounded LLM query — RAG over real process data (Connection 4)
	// Kind: client
	// Stability: development
	OcpmLlmQuerySpan = "ocpm.llm.query"
	// ocpm_ocel_ingest is the span name for "ocpm.ocel.ingest".
	//
	// OCEL 2.0 log ingestion — parse and load into ObjectCentricEventLog
	// Kind: internal
	// Stability: development
	OcpmOcelIngestSpan = "ocpm.ocel.ingest"
	// ocpm_performance_bottleneck is the span name for "ocpm.performance.bottleneck".
	//
	// Object-Centric bottleneck detection — top-N edges by severity score
	// Kind: internal
	// Stability: development
	OcpmPerformanceBottleneckSpan = "ocpm.performance.bottleneck"
	// ocpm_performance_throughput is the span name for "ocpm.performance.throughput".
	//
	// Object-Centric throughput computation — end-to-end duration per object type
	// Kind: internal
	// Stability: development
	OcpmPerformanceThroughputSpan = "ocpm.performance.throughput"
)