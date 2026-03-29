package semconv

const (
	// process_mining_alignment_analyze is the span name for "process.mining.alignment.analyze".
	//
	// Alignment analysis — examining multiple alignment results to identify common deviation patterns and fitness trends.
	// Kind: internal
	// Stability: development
	ProcessMiningAlignmentAnalyzeSpan = "process.mining.alignment.analyze"
	// process_mining_bottleneck_analyze is the span name for "process.mining.bottleneck.analyze".
	//
	// Bottleneck analysis — scoring and ranking detected bottlenecks by severity and impact.
	// Kind: internal
	// Stability: development
	ProcessMiningBottleneckAnalyzeSpan = "process.mining.bottleneck.analyze"
	// process_mining_bottleneck_detection is the span name for "process.mining.bottleneck_detection".
	//
	// Bottleneck detection — identifying the activity with the highest average waiting time.
	// Kind: internal
	// Stability: development
	ProcessMiningBottleneckDetectionSpan = "process.mining.bottleneck_detection"
	// process_mining_case_cluster is the span name for "process.mining.case.cluster".
	//
	// Case clustering — grouping process cases by behavioral similarity using ML clustering algorithms.
	// Kind: internal
	// Stability: development
	ProcessMiningCaseClusterSpan = "process.mining.case.cluster"
	// process_mining_complexity_measure is the span name for "process.mining.complexity.measure".
	//
	// Process complexity measurement — computing complexity metrics for a discovered process model.
	// Kind: internal
	// Stability: development
	ProcessMiningComplexityMeasureSpan = "process.mining.complexity.measure"
	// process_mining_conformance_deviation is the span name for "process.mining.conformance.deviation".
	//
	// Detection of a single conformance deviation during trace alignment.
	// Kind: internal
	// Stability: development
	ProcessMiningConformanceDeviationSpan = "process.mining.conformance.deviation"
	// process_mining_conformance_repair is the span name for "process.mining.conformance.repair".
	//
	// Conformance repair — automatically repairing a non-conformant trace to align with the process model.
	// Kind: internal
	// Stability: development
	ProcessMiningConformanceRepairSpan = "process.mining.conformance.repair"
	// process_mining_conformance_threshold is the span name for "process.mining.conformance.threshold".
	//
	// Conformance threshold check — evaluates all cases against the defined conformance threshold and reports violations.
	// Kind: internal
	// Stability: development
	ProcessMiningConformanceThresholdSpan = "process.mining.conformance.threshold"
	// process_mining_conformance_visualize is the span name for "process.mining.conformance.visualize".
	//
	// Generating a conformance visualization — token replay, alignment diagram, or footprint matrix.
	// Kind: internal
	// Stability: development
	ProcessMiningConformanceVisualizeSpan = "process.mining.conformance.visualize"
	// process_mining_decision_mine is the span name for "process.mining.decision.mine".
	//
	// Mining decision rules from a process log — discovers conditions that determine process branching.
	// Kind: internal
	// Stability: development
	ProcessMiningDecisionMineSpan = "process.mining.decision.mine"
	// process_mining_deviation is the span name for "process.mining.deviation".
	//
	// Detection of a single conformance deviation during trace alignment.
	// Kind: internal
	// Stability: development
	ProcessMiningDeviationSpan = "process.mining.deviation"
	// process_mining_dfg is the span name for "process.mining.dfg".
	//
	// Computation of a Directly-Follows Graph (DFG) from an event log.
	// Kind: internal
	// Stability: development
	ProcessMiningDfgSpan = "process.mining.dfg"
	// process_mining_dfg_compute is the span name for "process.mining.dfg.compute".
	//
	// Computation of a Directly-Follows Graph from an event log.
	// Kind: internal
	// Stability: development
	ProcessMiningDfgComputeSpan = "process.mining.dfg.compute"
	// process_mining_discovery is the span name for "process.mining.discovery".
	//
	// Process model discovery run — applying a mining algorithm to an event log to produce a Petri net or BPMN model.
	// Kind: internal
	// Stability: development
	ProcessMiningDiscoverySpan = "process.mining.discovery"
	// process_mining_drift_correct is the span name for "process.mining.drift.correct".
	//
	// Process drift correction — applying model adaptation to address detected concept drift.
	// Kind: internal
	// Stability: development
	ProcessMiningDriftCorrectSpan = "process.mining.drift.correct"
	// process_mining_drift_detect is the span name for "process.mining.drift.detect".
	//
	// Detecting concept drift in a streaming process mining window.
	// Kind: internal
	// Stability: development
	ProcessMiningDriftDetectSpan = "process.mining.drift.detect"
	// process_mining_event_abstract is the span name for "process.mining.event.abstract".
	//
	// Event abstraction — mapping raw low-level events to higher-level process activities.
	// Kind: internal
	// Stability: development
	ProcessMiningEventAbstractSpan = "process.mining.event.abstract"
	// process_mining_hierarchy_build is the span name for "process.mining.hierarchy.build".
	//
	// Building a process hierarchy tree from process mining trace data.
	// Kind: internal
	// Stability: development
	ProcessMiningHierarchyBuildSpan = "process.mining.hierarchy.build"
	// process_mining_inductive_mine is the span name for "process.mining.inductive.mine".
	//
	// Inductive miner algorithm — discovers a process model by recursively partitioning the event log using cut semantics.
	// Kind: internal
	// Stability: development
	ProcessMiningInductiveMineSpan = "process.mining.inductive.mine"
	// process_mining_log_preprocess is the span name for "process.mining.log.preprocess".
	//
	// Preprocessing an event log — filtering, sorting, and preparing for mining or conformance.
	// Kind: internal
	// Stability: development
	ProcessMiningLogPreprocessSpan = "process.mining.log.preprocess"
	// process_mining_model_enhance is the span name for "process.mining.model.enhance".
	//
	// Process model enhancement — augmenting a discovered model with performance, conformance, or organizational perspectives.
	// Kind: internal
	// Stability: development
	ProcessMiningModelEnhanceSpan = "process.mining.model.enhance"
	// process_mining_model_quality is the span name for "process.mining.model.quality".
	//
	// Quality assessment of an enhanced process model — measures coverage, fitness improvement, and enhancement perspective.
	// Kind: internal
	// Stability: development
	ProcessMiningModelQualitySpan = "process.mining.model.quality"
	// process_mining_prediction_make is the span name for "process.mining.prediction.make".
	//
	// Process outcome prediction — forecasting future trace completion, bottlenecks, or deviations using a predictive model.
	// Kind: internal
	// Stability: development
	ProcessMiningPredictionMakeSpan = "process.mining.prediction.make"
	// process_mining_replay_alignment is the span name for "process.mining.replay.alignment".
	//
	// Alignment-based conformance checking — computing optimal alignments between log and model.
	// Kind: internal
	// Stability: development
	ProcessMiningReplayAlignmentSpan = "process.mining.replay.alignment"
	// process_mining_replay_check is the span name for "process.mining.replay.check".
	//
	// Token replay conformance check — replaying a trace against a Petri net model to measure fitness.
	// Kind: internal
	// Stability: development
	ProcessMiningReplayCheckSpan = "process.mining.replay.check"
	// process_mining_replay_compare is the span name for "process.mining.replay.compare".
	//
	// Replay comparison — comparing fitness scores between baseline and target process models.
	// Kind: internal
	// Stability: development
	ProcessMiningReplayCompareSpan = "process.mining.replay.compare"
	// process_mining_root_cause_analyze is the span name for "process.mining.root_cause.analyze".
	//
	// Root cause analysis of a process anomaly — identifies why a deviation occurred.
	// Kind: internal
	// Stability: development
	ProcessMiningRootCauseAnalyzeSpan = "process.mining.root_cause.analyze"
	// process_mining_simulation_run is the span name for "process.mining.simulation.run".
	//
	// Running a process simulation — generates synthetic event logs from a discovered model.
	// Kind: internal
	// Stability: development
	ProcessMiningSimulationRunSpan = "process.mining.simulation.run"
	// process_mining_social_network_analyze is the span name for "process.mining.social_network.analyze".
	//
	// Social network analysis of a process log — discovering collaboration patterns, handover-of-work, and resource roles.
	// Kind: internal
	// Stability: development
	ProcessMiningSocialNetworkAnalyzeSpan = "process.mining.social_network.analyze"
	// process_mining_streaming_ingest is the span name for "process.mining.streaming.ingest".
	//
	// Ingesting an event batch into the streaming process mining window.
	// Kind: consumer
	// Stability: development
	ProcessMiningStreamingIngestSpan = "process.mining.streaming.ingest"
	// process_mining_temporal_analyze is the span name for "process.mining.temporal.analyze".
	//
	// Temporal analysis of a process — detecting drift, seasonality, and trend patterns.
	// Kind: internal
	// Stability: development
	ProcessMiningTemporalAnalyzeSpan = "process.mining.temporal.analyze"
	// process_mining_variant_analyze is the span name for "process.mining.variant.analyze".
	//
	// Analysis of process variants — identifying distinct execution patterns and their frequencies in the event log.
	// Kind: internal
	// Stability: development
	ProcessMiningVariantAnalyzeSpan = "process.mining.variant.analyze"
	// process_mining_variant_analysis is the span name for "process.mining.variant_analysis".
	//
	// Process variant analysis — identifying and ranking unique execution paths in the event log.
	// Kind: internal
	// Stability: development
	ProcessMiningVariantAnalysisSpan = "process.mining.variant_analysis"
)