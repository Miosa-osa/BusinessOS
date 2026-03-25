// Code generated from semconv/model/llm/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 8

package semconv

import "go.opentelemetry.io/otel/attribute"

// LLM Attributes

const (
	// LlmModelKey is the OTel attribute key for llm.model.
	// The LLM model identifier used for the request.
	LlmModelKey = attribute.Key("llm.model")
	// LlmProviderKey is the OTel attribute key for llm.provider.
	// The LLM provider (e.g. anthropic, openai, google).
	LlmProviderKey = attribute.Key("llm.provider")
	// LlmTokenInputKey is the OTel attribute key for llm.token.input.
	// Number of input tokens consumed by the LLM request.
	LlmTokenInputKey = attribute.Key("llm.token.input")
	// LlmTokenOutputKey is the OTel attribute key for llm.token.output.
	// Number of output tokens produced by the LLM response.
	LlmTokenOutputKey = attribute.Key("llm.token.output")
	// LlmLatencyMsKey is the OTel attribute key for llm.latency_ms.
	// End-to-end latency of the LLM request in milliseconds.
	LlmLatencyMsKey = attribute.Key("llm.latency_ms")
	// LlmTemperatureKey is the OTel attribute key for llm.temperature.
	// Sampling temperature used for the LLM request.
	LlmTemperatureKey = attribute.Key("llm.temperature")
	// LlmStopReasonKey is the OTel attribute key for llm.stop_reason.
	// The reason the LLM stopped generating tokens.
	LlmStopReasonKey = attribute.Key("llm.stop_reason")
)

// LlmModel returns an attribute KeyValue for llm.model.
func LlmModel(val string) attribute.KeyValue { return LlmModelKey.String(val) }

// LlmProvider returns an attribute KeyValue for llm.provider.
func LlmProvider(val string) attribute.KeyValue { return LlmProviderKey.String(val) }

// LlmTokenInput returns an attribute KeyValue for llm.token.input.
func LlmTokenInput(val int) attribute.KeyValue { return LlmTokenInputKey.Int(val) }

// LlmTokenOutput returns an attribute KeyValue for llm.token.output.
func LlmTokenOutput(val int) attribute.KeyValue { return LlmTokenOutputKey.Int(val) }

// LlmLatencyMs returns an attribute KeyValue for llm.latency_ms.
func LlmLatencyMs(val int) attribute.KeyValue { return LlmLatencyMsKey.Int(val) }

// LlmTemperature returns an attribute KeyValue for llm.temperature.
func LlmTemperature(val float64) attribute.KeyValue { return LlmTemperatureKey.Float64(val) }

// LlmStopReason returns an attribute KeyValue for llm.stop_reason.
func LlmStopReason(val string) attribute.KeyValue { return LlmStopReasonKey.String(val) }

// LlmStopReason* constants are the known enum values for llm.stop_reason.
const (
	LlmStopReasonMaxTokens  = "max_tokens"
	LlmStopReasonStopSeq    = "stop_sequence"
	LlmStopReasonLength     = "length"
	LlmStopReasonEndTurn    = "end_turn"
	LlmStopReasonToolUse    = "tool_use"
)
