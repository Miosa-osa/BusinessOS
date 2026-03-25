package semconv

import "go.opentelemetry.io/otel/attribute"

// LLM Evaluation attributes (iter15)
const (
	LlmEvaluationScoreKey            = attribute.Key("llm.evaluation.score")
	LlmEvaluationRubricKey           = attribute.Key("llm.evaluation.rubric")
	LlmEvaluationPassesThresholdKey  = attribute.Key("llm.evaluation.passes_threshold")
)

func LlmEvaluationScore(val float64) attribute.KeyValue {
	return LlmEvaluationScoreKey.Float64(val)
}

func LlmEvaluationRubric(val string) attribute.KeyValue {
	return LlmEvaluationRubricKey.String(val)
}

func LlmEvaluationPassesThreshold(val bool) attribute.KeyValue {
	return LlmEvaluationPassesThresholdKey.Bool(val)
}
