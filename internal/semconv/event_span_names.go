package semconv

const (
	// event_correlate is the span name for "event.correlate".
	//
	// Event correlation — linking multiple events into a causal chain for distributed tracing.
	// Kind: internal
	// Stability: development
	EventCorrelateSpan = "event.correlate"
	// event_deliver is the span name for "event.deliver".
	//
	// Delivering an event to registered handlers in the event bus.
	// Kind: internal
	// Stability: development
	EventDeliverSpan = "event.deliver"
	// event_emit is the span name for "event.emit".
	//
	// Emission of a structured log event to the event bus.
	// Kind: producer
	// Stability: development
	EventEmitSpan = "event.emit"
	// event_process is the span name for "event.process".
	//
	// Processing of a received structured log event from the bus.
	// Kind: consumer
	// Stability: development
	EventProcessSpan = "event.process"
	// event_replay is the span name for "event.replay".
	//
	// Event replay — re-processing a previously emitted event for recovery or audit.
	// Kind: internal
	// Stability: development
	EventReplaySpan = "event.replay"
	// event_route is the span name for "event.route".
	//
	// Routing an event to subscribers based on routing strategy and filters.
	// Kind: internal
	// Stability: development
	EventRouteSpan = "event.route"
)