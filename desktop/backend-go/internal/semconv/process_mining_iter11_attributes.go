// Code generated from semconv/model/process_mining/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 11

package semconv

import "go.opentelemetry.io/otel/attribute"

// Process mining replay and alignment quality attributes (iter11).

const (
	// ProcessMiningReplayPrecisionKey is the OTel attribute key for process_mining.replay.precision.
	// Token-based replay precision score [0.0, 1.0].
	ProcessMiningReplayPrecisionKey = attribute.Key("process_mining.replay.precision")
	// ProcessMiningReplayGeneralizationKey is the OTel attribute key for process_mining.replay.generalization.
	// Token-based replay generalization score [0.0, 1.0].
	ProcessMiningReplayGeneralizationKey = attribute.Key("process_mining.replay.generalization")
	// ProcessMiningReplaySimplicityKey is the OTel attribute key for process_mining.replay.simplicity.
	// Token-based replay simplicity score [0.0, 1.0].
	ProcessMiningReplaySimplicityKey = attribute.Key("process_mining.replay.simplicity")
	// ProcessMiningAlignmentCostKey is the OTel attribute key for process_mining.alignment.cost.
	// Total alignment cost for trace-to-model alignment.
	ProcessMiningAlignmentCostKey = attribute.Key("process_mining.alignment.cost")
	// ProcessMiningModelTypeKey is the OTel attribute key for process_mining.model.type.
	// Type of process model discovered or evaluated.
	ProcessMiningModelTypeKey = attribute.Key("process_mining.model.type")
)

// ProcessMiningReplayPrecision returns an attribute KeyValue for process_mining.replay.precision.
func ProcessMiningReplayPrecision(val float64) attribute.KeyValue {
	return ProcessMiningReplayPrecisionKey.Float64(val)
}

// ProcessMiningReplayGeneralization returns an attribute KeyValue for process_mining.replay.generalization.
func ProcessMiningReplayGeneralization(val float64) attribute.KeyValue {
	return ProcessMiningReplayGeneralizationKey.Float64(val)
}

// ProcessMiningReplaySimplicity returns an attribute KeyValue for process_mining.replay.simplicity.
func ProcessMiningReplaySimplicity(val float64) attribute.KeyValue {
	return ProcessMiningReplaySimplicityKey.Float64(val)
}

// ProcessMiningAlignmentCost returns an attribute KeyValue for process_mining.alignment.cost.
func ProcessMiningAlignmentCost(val float64) attribute.KeyValue {
	return ProcessMiningAlignmentCostKey.Float64(val)
}

// ProcessMiningModelType returns an attribute KeyValue for process_mining.model.type.
func ProcessMiningModelType(val string) attribute.KeyValue {
	return ProcessMiningModelTypeKey.String(val)
}
