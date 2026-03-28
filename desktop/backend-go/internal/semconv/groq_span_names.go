package semconv

const (
	// groq_workflow_decision is the span name for "groq.workflow.decision".
	//
	// Span for a Groq LLM call that produces a YAWL workflow routing decision. Bridges the Groq response to a YAWL workflow action (launch_case, start_workitem, complete_workitem, checkpoint). The decision.wcp_pattern identifies which WCP pattern the LLM decision is targeting.

	// Kind: client
	// Stability: development
	GroqWorkflowDecisionSpan = "groq.workflow.decision"
)