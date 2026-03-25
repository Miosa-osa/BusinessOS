// Code generated from semconv/model/conversation/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 10

package semconv

import "go.opentelemetry.io/otel/attribute"

// Conversation Attributes

const (
	// ConversationIDKey is the OTel attribute key for conversation.id.
	// Unique identifier for the conversation session.
	ConversationIDKey = attribute.Key("conversation.id")
	// ConversationTurnCountKey is the OTel attribute key for conversation.turn_count.
	// Number of turns (user+assistant exchanges) in the conversation.
	ConversationTurnCountKey = attribute.Key("conversation.turn_count")
	// ConversationModelKey is the OTel attribute key for conversation.model.
	// LLM model used for the conversation.
	ConversationModelKey = attribute.Key("conversation.model")
	// ConversationContextTokensKey is the OTel attribute key for conversation.context_tokens.
	// Number of tokens in the current conversation context window.
	ConversationContextTokensKey = attribute.Key("conversation.context_tokens")
	// ConversationToolCallsKey is the OTel attribute key for conversation.tool_calls.
	// Number of tool calls made in the conversation.
	ConversationToolCallsKey = attribute.Key("conversation.tool_calls")
	// ConversationPhaseKey is the OTel attribute key for conversation.phase.
	// Current phase of the conversation lifecycle.
	ConversationPhaseKey = attribute.Key("conversation.phase")
	// ConversationUserIDKey is the OTel attribute key for conversation.user_id.
	// Identifier of the user participating in the conversation.
	ConversationUserIDKey = attribute.Key("conversation.user_id")
)

// ConversationID returns an attribute KeyValue for conversation.id.
func ConversationID(val string) attribute.KeyValue {
	return ConversationIDKey.String(val)
}

// ConversationTurnCount returns an attribute KeyValue for conversation.turn_count.
func ConversationTurnCount(val int) attribute.KeyValue {
	return ConversationTurnCountKey.Int(val)
}

// ConversationModel returns an attribute KeyValue for conversation.model.
func ConversationModel(val string) attribute.KeyValue {
	return ConversationModelKey.String(val)
}

// ConversationContextTokens returns an attribute KeyValue for conversation.context_tokens.
func ConversationContextTokens(val int) attribute.KeyValue {
	return ConversationContextTokensKey.Int(val)
}

// ConversationToolCalls returns an attribute KeyValue for conversation.tool_calls.
func ConversationToolCalls(val int) attribute.KeyValue {
	return ConversationToolCallsKey.Int(val)
}

// ConversationPhase returns an attribute KeyValue for conversation.phase.
func ConversationPhase(val string) attribute.KeyValue {
	return ConversationPhaseKey.String(val)
}

// ConversationUserID returns an attribute KeyValue for conversation.user_id.
func ConversationUserID(val string) attribute.KeyValue {
	return ConversationUserIDKey.String(val)
}

const (
	ConversationPhaseInit     = "init"
	ConversationPhaseActive   = "active"
	ConversationPhaseWaiting  = "waiting"
	ConversationPhaseComplete = "complete"
	ConversationPhaseError    = "error"
)
