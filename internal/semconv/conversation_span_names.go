package semconv

const (
	// conversation_compress is the span name for "conversation.compress".
	//
	// Context compression — summarizing or truncating conversation history to fit context window.
	// Kind: internal
	// Stability: development
	ConversationCompressSpan = "conversation.compress"
	// conversation_start is the span name for "conversation.start".
	//
	// Conversation session initialization — first turn, context loaded.
	// Kind: internal
	// Stability: development
	ConversationStartSpan = "conversation.start"
	// conversation_turn is the span name for "conversation.turn".
	//
	// Single conversation turn — user message received, assistant response generated.
	// Kind: internal
	// Stability: development
	ConversationTurnSpan = "conversation.turn"
)