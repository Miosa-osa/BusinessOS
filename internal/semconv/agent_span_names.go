package semconv

const (
	// agent_capability_catalog is the span name for "agent.capability.catalog".
	//
	// Agent capability catalog operation — registering or querying the catalog of agent capabilities.
	// Kind: internal
	// Stability: development
	AgentCapabilityCatalogSpan = "agent.capability.catalog"
	// agent_coordinate is the span name for "agent.coordinate".
	//
	// Agent coordination operation — dispatching tasks to sub-agents in a topology.
	// Kind: client
	// Stability: development
	AgentCoordinateSpan = "agent.coordinate"
	// agent_decision is the span name for "agent.decision".
	//
	// An autonomous decision made by an agent — action selection with confidence scoring.
	// Kind: internal
	// Stability: development
	AgentDecisionSpan = "agent.decision"
	// agent_execution_graph is the span name for "agent.execution.graph".
	//
	// Execution of an agent execution graph — traversing a DAG of agent steps to completion.
	// Kind: internal
	// Stability: development
	AgentExecutionGraphSpan = "agent.execution.graph"
	// agent_handoff is the span name for "agent.handoff".
	//
	// Agent handoff — transfers control and state to another agent based on capability, load, or priority.
	// Kind: producer
	// Stability: development
	AgentHandoffSpan = "agent.handoff"
	// agent_llm_predict is the span name for "agent.llm_predict".
	//
	// LLM inference call made by an OSA agent.
	// Kind: client
	// Stability: development
	AgentLlmPredictSpan = "agent.llm_predict"
	// agent_loop is the span name for "agent.loop".
	//
	// One iteration of the agent's main reasoning and action loop.
	// Kind: internal
	// Stability: development
	AgentLoopSpan = "agent.loop"
	// agent_memory_federate is the span name for "agent.memory.federate".
	//
	// Synchronizing agent memory state with a federated memory pool shared across agents.
	// Kind: client
	// Stability: development
	AgentMemoryFederateSpan = "agent.memory.federate"
	// agent_memory_update is the span name for "agent.memory.update".
	//
	// Agent memory update — writing new information to agent working memory.
	// Kind: internal
	// Stability: development
	AgentMemoryUpdateSpan = "agent.memory.update"
	// agent_pipeline_execute is the span name for "agent.pipeline.execute".
	//
	// Execution of an agent pipeline stage — processes data through a defined transformation.
	// Kind: internal
	// Stability: development
	AgentPipelineExecuteSpan = "agent.pipeline.execute"
	// agent_reasoning_trace is the span name for "agent.reasoning.trace".
	//
	// Agent reasoning trace — records the chain-of-thought steps an agent takes to reach a decision.
	// Kind: internal
	// Stability: development
	AgentReasoningTraceSpan = "agent.reasoning.trace"
	// agent_spawn is the span name for "agent.spawn".
	//
	// Agent spawning — creating a new child agent under the current supervision tree.
	// Kind: internal
	// Stability: development
	AgentSpawnSpan = "agent.spawn"
	// agent_spawn_profile is the span name for "agent.spawn.profile".
	//
	// Agent spawn profiling — observing the performance characteristics of a child agent spawn operation.
	// Kind: internal
	// Stability: development
	AgentSpawnProfileSpan = "agent.spawn.profile"
	// agent_workflow_checkpoint is the span name for "agent.workflow.checkpoint".
	//
	// Agent workflow checkpoint — capturing workflow state to enable resumption after interruption.
	// Kind: internal
	// Stability: development
	AgentWorkflowCheckpointSpan = "agent.workflow.checkpoint"
)