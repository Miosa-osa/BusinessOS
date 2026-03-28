package semconv

const (
	// yawl_case is the span name for "yawl.case".
	//
	// Root span for a YAWL workflow case. One span per case_id. Encapsulates the full lifecycle from INSTANCE_CREATED to INSTANCE_COMPLETED or INSTANCE_CANCELLED. The yawl.case.id is the correlation key linking all task execution spans.

	// Kind: internal
	// Stability: development
	YawlCaseSpan = "yawl.case"
	// yawl_case_launch is the span name for "yawl.case.launch".
	//
	// Client span for launching a new YAWL case via the embedded server HTTP endpoint (POST /api/cases/launch). Emitted by CaseLifecycle GenServer.

	// Kind: client
	// Stability: development
	YawlCaseLaunchSpan = "yawl.case.launch"
	// yawl_task_execution is the span name for "yawl.task.execution".
	//
	// Span for a single YAWL task execution within a case. Child of span.yawl.case. Covers the full task lifecycle: TASK_ENABLED → TASK_STARTED (tokens consumed) → TASK_COMPLETED (tokens produced). The yawl.token.consumed and yawl.token.produced attributes record Petri net token flow.

	// Kind: internal
	// Stability: development
	YawlTaskExecutionSpan = "yawl.task.execution"
	// yawl_workitem_complete is the span name for "yawl.workitem.complete".
	//
	// Client span for completing (checking in) a YAWL work item via the embedded server (POST /api/cases/{id}/workitems/{wid}/complete). Emitted by CaseLifecycle GenServer.

	// Kind: client
	// Stability: development
	YawlWorkitemCompleteSpan = "yawl.workitem.complete"
)