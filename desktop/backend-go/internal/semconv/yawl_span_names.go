package semconv

const (
	// yawl_case is the span name for "yawl.case".
	//
	// Root span for a YAWL workflow case. One span per case_id. Encapsulates the full lifecycle from INSTANCE_CREATED to INSTANCE_COMPLETED or INSTANCE_CANCELLED. The yawl.case.id is the correlation key linking all task execution spans.

	// Kind: internal
	// Stability: development
	YawlCaseSpan = "yawl.case"
	// yawl_task_execution is the span name for "yawl.task.execution".
	//
	// Span for a single YAWL task execution within a case. Child of span.yawl.case. Covers the full task lifecycle: TASK_ENABLED → TASK_STARTED (tokens consumed) → TASK_COMPLETED (tokens produced). The yawl.token.consumed and yawl.token.produced attributes record Petri net token flow.

	// Kind: internal
	// Stability: development
	YawlTaskExecutionSpan = "yawl.task.execution"
)