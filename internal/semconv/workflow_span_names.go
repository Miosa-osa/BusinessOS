package semconv

const (
	// workflow_cancel_region is the span name for "workflow.cancel_region".
	//
	// Cancellation of a workflow region — all in-flight activities in region halted.
	// Kind: internal
	// Stability: development
	WorkflowCancelRegionSpan = "workflow.cancel_region"
	// workflow_critical_section is the span name for "workflow.critical_section".
	//
	// Critical section execution — ensures atomic sequential execution of enclosed activities.
	// Kind: internal
	// Stability: development
	WorkflowCriticalSectionSpan = "workflow.critical_section"
	// workflow_deferred_choice is the span name for "workflow.deferred_choice".
	//
	// Deferred exclusive choice — decision deferred until first branch fires.
	// Kind: internal
	// Stability: development
	WorkflowDeferredChoiceSpan = "workflow.deferred_choice"
	// workflow_discriminator is the span name for "workflow.discriminator".
	//
	// N-out-of-M join evaluation — fires when N of M branches complete.
	// Kind: internal
	// Stability: development
	WorkflowDiscriminatorSpan = "workflow.discriminator"
	// workflow_exclusive_choice is the span name for "workflow.exclusive_choice".
	//
	// Exclusive choice pattern (WP-4) — XOR split, exactly one branch is selected based on condition.
	// Kind: internal
	// Stability: development
	WorkflowExclusiveChoiceSpan = "workflow.exclusive_choice"
	// workflow_execute is the span name for "workflow.execute".
	//
	// Execution of a single workflow step or activity in the YAWL workflow engine.
	// Kind: internal
	// Stability: development
	WorkflowExecuteSpan = "workflow.execute"
	// workflow_interleaved_routing is the span name for "workflow.interleaved_routing".
	//
	// Interleaved routing execution — activities in a set run one at a time in arbitrary order.
	// Kind: internal
	// Stability: development
	WorkflowInterleavedRoutingSpan = "workflow.interleaved_routing"
	// workflow_milestone is the span name for "workflow.milestone".
	//
	// Milestone gate check — execution blocked until milestone condition met.
	// Kind: internal
	// Stability: development
	WorkflowMilestoneSpan = "workflow.milestone"
	// workflow_multi_choice is the span name for "workflow.multi_choice".
	//
	// Multi-choice pattern (WP-6) — one or more branches selected based on runtime conditions.
	// Kind: internal
	// Stability: development
	WorkflowMultiChoiceSpan = "workflow.multi_choice"
	// workflow_multi_instance is the span name for "workflow.multi_instance".
	//
	// Multi-instance activity execution — N parallel instances of same activity.
	// Kind: internal
	// Stability: development
	WorkflowMultiInstanceSpan = "workflow.multi_instance"
	// workflow_parallel_split is the span name for "workflow.parallel_split".
	//
	// Parallel split pattern (WP-2) — single thread of control splits into N concurrent branches.
	// Kind: internal
	// Stability: development
	WorkflowParallelSplitSpan = "workflow.parallel_split"
	// workflow_persistent_trigger is the span name for "workflow.persistent_trigger".
	//
	// Persistent trigger activation — trigger that persists in the environment until explicitly consumed.
	// Kind: producer
	// Stability: development
	WorkflowPersistentTriggerSpan = "workflow.persistent_trigger"
	// workflow_sequence is the span name for "workflow.sequence".
	//
	// Sequence pattern (WP-1) — activities execute in strict serial order.
	// Kind: internal
	// Stability: development
	WorkflowSequenceSpan = "workflow.sequence"
	// workflow_simple_merge is the span name for "workflow.simple_merge".
	//
	// Simple merge pattern (WP-5) — merges two or more alternative branches without synchronization.
	// Kind: internal
	// Stability: development
	WorkflowSimpleMergeSpan = "workflow.simple_merge"
	// workflow_structured_loop is the span name for "workflow.structured_loop".
	//
	// Structured loop iteration — while-do execution with bounded iteration count.
	// Kind: internal
	// Stability: development
	WorkflowStructuredLoopSpan = "workflow.structured_loop"
	// workflow_structured_sync_merge is the span name for "workflow.structured_sync_merge".
	//
	// Structured synchronizing merge (WP-7) — merges branches, waiting for all that were activated.
	// Kind: internal
	// Stability: development
	WorkflowStructuredSyncMergeSpan = "workflow.structured_sync_merge"
	// workflow_synchronization is the span name for "workflow.synchronization".
	//
	// Synchronization pattern (WP-3) — waits for ALL concurrent branches to complete before merging.
	// Kind: internal
	// Stability: development
	WorkflowSynchronizationSpan = "workflow.synchronization"
	// workflow_transition is the span name for "workflow.transition".
	//
	// State transition within a workflow — moving from one state to another.
	// Kind: internal
	// Stability: development
	WorkflowTransitionSpan = "workflow.transition"
)