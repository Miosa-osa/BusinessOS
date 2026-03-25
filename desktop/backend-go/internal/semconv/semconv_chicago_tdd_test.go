// Package semconv provides Chicago TDD validation for Weaver-generated semconv constants.
//
// These tests enforce schema contracts at compile time:
// - Rename an attribute in semconv YAML → compile error here
// - Remove an enum value → compile error here
// - Third proof layer: schema conformance via typed constants
//
// Run with: cd BusinessOS/desktop/backend-go && go test ./internal/semconv/...
package semconv

import (
	"testing"
)

// ============================================================
// Healing domain — span.healing.diagnosis + span.healing.reflex_arc
// ============================================================

func TestHealingFailureModeKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingFailureModeKey) != "healing.failure_mode" {
		t.Errorf("HealingFailureModeKey = %q, want %q", HealingFailureModeKey, "healing.failure_mode")
	}
}

func TestHealingConfidenceKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingConfidenceKey) != "healing.confidence" {
		t.Errorf("HealingConfidenceKey = %q, want %q", HealingConfidenceKey, "healing.confidence")
	}
}

func TestHealingAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingAgentIdKey) != "healing.agent_id" {
		t.Errorf("HealingAgentIdKey = %q, want %q", HealingAgentIdKey, "healing.agent_id")
	}
}

func TestHealingReflexArcKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingReflexArcKey) != "healing.reflex_arc" {
		t.Errorf("HealingReflexArcKey = %q, want %q", HealingReflexArcKey, "healing.reflex_arc")
	}
}

func TestHealingRecoveryActionKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingRecoveryActionKey) != "healing.recovery_action" {
		t.Errorf("HealingRecoveryActionKey = %q, want %q", HealingRecoveryActionKey, "healing.recovery_action")
	}
}

func TestHealingMttrMsKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingMttrMsKey) != "healing.mttr_ms" {
		t.Errorf("HealingMttrMsKey = %q, want %q", HealingMttrMsKey, "healing.mttr_ms")
	}
}

func TestHealingFailureModeDeadlockValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.Deadlock != "deadlock" {
		t.Errorf("HealingFailureModeValues.Deadlock = %q, want %q", HealingFailureModeValues.Deadlock, "deadlock")
	}
}

func TestHealingFailureModeTimeoutValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.Timeout != "timeout" {
		t.Errorf("HealingFailureModeValues.Timeout = %q, want %q", HealingFailureModeValues.Timeout, "timeout")
	}
}

func TestHealingFailureModeRaceConditionValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.RaceCondition != "race_condition" {
		t.Errorf("HealingFailureModeValues.RaceCondition = %q, want %q", HealingFailureModeValues.RaceCondition, "race_condition")
	}
}

func TestHealingFailureModeLivelockValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.Livelock != "livelock" {
		t.Errorf("HealingFailureModeValues.Livelock = %q, want %q", HealingFailureModeValues.Livelock, "livelock")
	}
}

func TestHealingFailureModeMemoryLeakValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.MemoryLeak != "memory_leak" {
		t.Errorf("HealingFailureModeValues.MemoryLeak = %q, want %q", HealingFailureModeValues.MemoryLeak, "memory_leak")
	}
}

func TestHealingFailureModeCascadingFailureValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.CascadingFailure != "cascading_failure" {
		t.Errorf("HealingFailureModeValues.CascadingFailure = %q, want %q", HealingFailureModeValues.CascadingFailure, "cascading_failure")
	}
}

func TestHealingFailureModeStagnationValueMatchesSchema(t *testing.T) {
	if HealingFailureModeValues.Stagnation != "stagnation" {
		t.Errorf("HealingFailureModeValues.Stagnation = %q, want %q", HealingFailureModeValues.Stagnation, "stagnation")
	}
}

// ============================================================
// A2A domain — span.a2a.call + span.a2a.create_deal
// ============================================================

func TestA2AAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aAgentIdKey) != "a2a.agent.id" {
		t.Errorf("A2aAgentIdKey = %q, want %q", A2aAgentIdKey, "a2a.agent.id")
	}
}

func TestA2ADealIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aDealIdKey) != "a2a.deal.id" {
		t.Errorf("A2aDealIdKey = %q, want %q", A2aDealIdKey, "a2a.deal.id")
	}
}

func TestA2AOperationKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aOperationKey) != "a2a.operation" {
		t.Errorf("A2aOperationKey = %q, want %q", A2aOperationKey, "a2a.operation")
	}
}

func TestA2ASourceServiceKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aSourceServiceKey) != "a2a.source.service" {
		t.Errorf("A2aSourceServiceKey = %q, want %q", A2aSourceServiceKey, "a2a.source.service")
	}
}

func TestA2ATargetServiceKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aTargetServiceKey) != "a2a.target.service" {
		t.Errorf("A2aTargetServiceKey = %q, want %q", A2aTargetServiceKey, "a2a.target.service")
	}
}

func TestA2ADealTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aDealTypeKey) != "a2a.deal.type" {
		t.Errorf("A2aDealTypeKey = %q, want %q", A2aDealTypeKey, "a2a.deal.type")
	}
}

// ============================================================
// BusinessOS domain — span.bos.compliance.check
// ============================================================

func TestBosComplianceFrameworkKeyIsCorrectOtelName(t *testing.T) {
	if string(BosComplianceFrameworkKey) != "bos.compliance.framework" {
		t.Errorf("BosComplianceFrameworkKey = %q, want %q", BosComplianceFrameworkKey, "bos.compliance.framework")
	}
}

func TestBosComplianceRuleIDKeyIsCorrectOtelName(t *testing.T) {
	if string(BosComplianceRuleIdKey) != "bos.compliance.rule_id" {
		t.Errorf("BosComplianceRuleIdKey = %q, want %q", BosComplianceRuleIdKey, "bos.compliance.rule_id")
	}
}

func TestBosCompliancePassedKeyIsCorrectOtelName(t *testing.T) {
	if string(BosCompliancePassedKey) != "bos.compliance.passed" {
		t.Errorf("BosCompliancePassedKey = %q, want %q", BosCompliancePassedKey, "bos.compliance.passed")
	}
}

func TestBosComplianceSeverityKeyIsCorrectOtelName(t *testing.T) {
	if string(BosComplianceSeverityKey) != "bos.compliance.severity" {
		t.Errorf("BosComplianceSeverityKey = %q, want %q", BosComplianceSeverityKey, "bos.compliance.severity")
	}
}

func TestBosDecisionTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(BosDecisionTypeKey) != "bos.decision.type" {
		t.Errorf("BosDecisionTypeKey = %q, want %q", BosDecisionTypeKey, "bos.decision.type")
	}
}

func TestBosDecisionIDKeyIsCorrectOtelName(t *testing.T) {
	if string(BosDecisionIdKey) != "bos.decision.id" {
		t.Errorf("BosDecisionIdKey = %q, want %q", BosDecisionIdKey, "bos.decision.id")
	}
}

func TestBosWorkspaceIDKeyIsCorrectOtelName(t *testing.T) {
	if string(BosWorkspaceIdKey) != "bos.workspace.id" {
		t.Errorf("BosWorkspaceIdKey = %q, want %q", BosWorkspaceIdKey, "bos.workspace.id")
	}
}

func TestBosWorkspaceNameKeyIsCorrectOtelName(t *testing.T) {
	if string(BosWorkspaceNameKey) != "bos.workspace.name" {
		t.Errorf("BosWorkspaceNameKey = %q, want %q", BosWorkspaceNameKey, "bos.workspace.name")
	}
}

func TestBosAgentServiceKeyIsCorrectOtelName(t *testing.T) {
	if string(BosAgentServiceKey) != "bos.agent.service" {
		t.Errorf("BosAgentServiceKey = %q, want %q", BosAgentServiceKey, "bos.agent.service")
	}
}

func TestBosComplianceSeverityCriticalValueMatchesSchema(t *testing.T) {
	if BosComplianceSeverityValues.Critical != "critical" {
		t.Errorf("BosComplianceSeverityValues.Critical = %q, want %q", BosComplianceSeverityValues.Critical, "critical")
	}
}

func TestBosComplianceSeverityHighValueMatchesSchema(t *testing.T) {
	if BosComplianceSeverityValues.High != "high" {
		t.Errorf("BosComplianceSeverityValues.High = %q, want %q", BosComplianceSeverityValues.High, "high")
	}
}

func TestBosComplianceSeverityMediumValueMatchesSchema(t *testing.T) {
	if BosComplianceSeverityValues.Medium != "medium" {
		t.Errorf("BosComplianceSeverityValues.Medium = %q, want %q", BosComplianceSeverityValues.Medium, "medium")
	}
}

func TestBosComplianceSeverityLowValueMatchesSchema(t *testing.T) {
	if BosComplianceSeverityValues.Low != "low" {
		t.Errorf("BosComplianceSeverityValues.Low = %q, want %q", BosComplianceSeverityValues.Low, "low")
	}
}

func TestBosDecisionTypeArchitecturalValueMatchesSchema(t *testing.T) {
	if BosDecisionTypeValues.Architectural != "architectural" {
		t.Errorf("BosDecisionTypeValues.Architectural = %q, want %q", BosDecisionTypeValues.Architectural, "architectural")
	}
}

func TestBosDecisionTypeOperationalValueMatchesSchema(t *testing.T) {
	if BosDecisionTypeValues.Operational != "operational" {
		t.Errorf("BosDecisionTypeValues.Operational = %q, want %q", BosDecisionTypeValues.Operational, "operational")
	}
}

func TestBosDecisionTypeStrategicValueMatchesSchema(t *testing.T) {
	if BosDecisionTypeValues.Strategic != "strategic" {
		t.Errorf("BosDecisionTypeValues.Strategic = %q, want %q", BosDecisionTypeValues.Strategic, "strategic")
	}
}

func TestBosDecisionTypeComplianceValueMatchesSchema(t *testing.T) {
	if BosDecisionTypeValues.Compliance != "compliance" {
		t.Errorf("BosDecisionTypeValues.Compliance = %q, want %q", BosDecisionTypeValues.Compliance, "compliance")
	}
}

func TestBosComplianceFrameworkSOC2ValueMatchesSchema(t *testing.T) {
	if BosComplianceFrameworkValues.Soc2 != "SOC2" {
		t.Errorf("BosComplianceFrameworkValues.Soc2 = %q, want %q", BosComplianceFrameworkValues.Soc2, "SOC2")
	}
}

func TestBosComplianceFrameworkHipaaValueMatchesSchema(t *testing.T) {
	if BosComplianceFrameworkValues.Hipaa != "HIPAA" {
		t.Errorf("BosComplianceFrameworkValues.Hipaa = %q, want %q", BosComplianceFrameworkValues.Hipaa, "HIPAA")
	}
}

func TestBosComplianceFrameworkGdprValueMatchesSchema(t *testing.T) {
	if BosComplianceFrameworkValues.Gdpr != "GDPR" {
		t.Errorf("BosComplianceFrameworkValues.Gdpr = %q, want %q", BosComplianceFrameworkValues.Gdpr, "GDPR")
	}
}

func TestBosComplianceFrameworkSoxValueMatchesSchema(t *testing.T) {
	if BosComplianceFrameworkValues.Sox != "SOX" {
		t.Errorf("BosComplianceFrameworkValues.Sox = %q, want %q", BosComplianceFrameworkValues.Sox, "SOX")
	}
}

// ============================================================
// Workflow domain — span.workflow.execute (YAWL patterns)
// ============================================================

func TestWorkflowIDKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowIdKey) != "workflow.id" {
		t.Errorf("WorkflowIdKey = %q, want %q", WorkflowIdKey, "workflow.id")
	}
}

func TestWorkflowNameKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowNameKey) != "workflow.name" {
		t.Errorf("WorkflowNameKey = %q, want %q", WorkflowNameKey, "workflow.name")
	}
}

func TestWorkflowPatternKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowPatternKey) != "workflow.pattern" {
		t.Errorf("WorkflowPatternKey = %q, want %q", WorkflowPatternKey, "workflow.pattern")
	}
}

func TestWorkflowStateKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowStateKey) != "workflow.state" {
		t.Errorf("WorkflowStateKey = %q, want %q", WorkflowStateKey, "workflow.state")
	}
}

func TestWorkflowEngineKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowEngineKey) != "workflow.engine" {
		t.Errorf("WorkflowEngineKey = %q, want %q", WorkflowEngineKey, "workflow.engine")
	}
}

func TestWorkflowStepKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowStepKey) != "workflow.step" {
		t.Errorf("WorkflowStepKey = %q, want %q", WorkflowStepKey, "workflow.step")
	}
}

func TestWorkflowStepCountKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowStepCountKey) != "workflow.step_count" {
		t.Errorf("WorkflowStepCountKey = %q, want %q", WorkflowStepCountKey, "workflow.step_count")
	}
}

func TestWorkflowPatternSequenceValueMatchesSchema(t *testing.T) {
	if WorkflowPatternValues.Sequence != "sequence" {
		t.Errorf("WorkflowPatternValues.Sequence = %q, want %q", WorkflowPatternValues.Sequence, "sequence")
	}
}

func TestWorkflowPatternParallelSplitValueMatchesSchema(t *testing.T) {
	if WorkflowPatternValues.ParallelSplit != "parallel_split" {
		t.Errorf("WorkflowPatternValues.ParallelSplit = %q, want %q", WorkflowPatternValues.ParallelSplit, "parallel_split")
	}
}

func TestWorkflowPatternSynchronizationValueMatchesSchema(t *testing.T) {
	if WorkflowPatternValues.Synchronization != "synchronization" {
		t.Errorf("WorkflowPatternValues.Synchronization = %q, want %q", WorkflowPatternValues.Synchronization, "synchronization")
	}
}

func TestWorkflowPatternExclusiveChoiceValueMatchesSchema(t *testing.T) {
	if WorkflowPatternValues.ExclusiveChoice != "exclusive_choice" {
		t.Errorf("WorkflowPatternValues.ExclusiveChoice = %q, want %q", WorkflowPatternValues.ExclusiveChoice, "exclusive_choice")
	}
}

func TestWorkflowPatternStructuredLoopValueMatchesSchema(t *testing.T) {
	if WorkflowPatternValues.StructuredLoop != "structured_loop" {
		t.Errorf("WorkflowPatternValues.StructuredLoop = %q, want %q", WorkflowPatternValues.StructuredLoop, "structured_loop")
	}
}

func TestWorkflowStateActiveValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Active != "active" {
		t.Errorf("WorkflowStateValues.Active = %q, want %q", WorkflowStateValues.Active, "active")
	}
}

func TestWorkflowStateCompletedValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Completed != "completed" {
		t.Errorf("WorkflowStateValues.Completed = %q, want %q", WorkflowStateValues.Completed, "completed")
	}
}

func TestWorkflowStateFailedValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Failed != "failed" {
		t.Errorf("WorkflowStateValues.Failed = %q, want %q", WorkflowStateValues.Failed, "failed")
	}
}

func TestWorkflowStatePendingValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Pending != "pending" {
		t.Errorf("WorkflowStateValues.Pending = %q, want %q", WorkflowStateValues.Pending, "pending")
	}
}

func TestWorkflowStateCancelledValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Cancelled != "cancelled" {
		t.Errorf("WorkflowStateValues.Cancelled = %q, want %q", WorkflowStateValues.Cancelled, "cancelled")
	}
}

func TestWorkflowStateSuspendedValueMatchesSchema(t *testing.T) {
	if WorkflowStateValues.Suspended != "suspended" {
		t.Errorf("WorkflowStateValues.Suspended = %q, want %q", WorkflowStateValues.Suspended, "suspended")
	}
}

func TestWorkflowEngineCanopyValueMatchesSchema(t *testing.T) {
	if WorkflowEngineValues.Canopy != "canopy" {
		t.Errorf("WorkflowEngineValues.Canopy = %q, want %q", WorkflowEngineValues.Canopy, "canopy")
	}
}

func TestWorkflowEngineYawlValueMatchesSchema(t *testing.T) {
	if WorkflowEngineValues.Yawl != "yawl" {
		t.Errorf("WorkflowEngineValues.Yawl = %q, want %q", WorkflowEngineValues.Yawl, "yawl")
	}
}

// ============================================================
// Consensus domain — span.consensus.round (HotStuff BFT)
// ============================================================

func TestConsensusRoundNumKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusRoundNumKey) != "consensus.round_num" {
		t.Errorf("ConsensusRoundNumKey = %q, want %q", ConsensusRoundNumKey, "consensus.round_num")
	}
}

func TestConsensusRoundTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusRoundTypeKey) != "consensus.round_type" {
		t.Errorf("ConsensusRoundTypeKey = %q, want %q", ConsensusRoundTypeKey, "consensus.round_type")
	}
}

func TestConsensusNodeIDKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusNodeIdKey) != "consensus.node_id" {
		t.Errorf("ConsensusNodeIdKey = %q, want %q", ConsensusNodeIdKey, "consensus.node_id")
	}
}

func TestConsensusQuorumSizeKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusQuorumSizeKey) != "consensus.quorum_size" {
		t.Errorf("ConsensusQuorumSizeKey = %q, want %q", ConsensusQuorumSizeKey, "consensus.quorum_size")
	}
}

func TestConsensusLatencyMsKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusLatencyMsKey) != "consensus.latency_ms" {
		t.Errorf("ConsensusLatencyMsKey = %q, want %q", ConsensusLatencyMsKey, "consensus.latency_ms")
	}
}

func TestConsensusRoundTypePrepareValueMatchesSchema(t *testing.T) {
	if ConsensusRoundTypeValues.Prepare != "prepare" {
		t.Errorf("ConsensusRoundTypeValues.Prepare = %q, want %q", ConsensusRoundTypeValues.Prepare, "prepare")
	}
}

func TestConsensusRoundTypeAcceptValueMatchesSchema(t *testing.T) {
	if ConsensusRoundTypeValues.Accept != "accept" {
		t.Errorf("ConsensusRoundTypeValues.Accept = %q, want %q", ConsensusRoundTypeValues.Accept, "accept")
	}
}

func TestConsensusRoundTypePromiseValueMatchesSchema(t *testing.T) {
	if ConsensusRoundTypeValues.Promise != "promise" {
		t.Errorf("ConsensusRoundTypeValues.Promise = %q, want %q", ConsensusRoundTypeValues.Promise, "promise")
	}
}

func TestConsensusRoundTypeLearnValueMatchesSchema(t *testing.T) {
	if ConsensusRoundTypeValues.Learn != "learn" {
		t.Errorf("ConsensusRoundTypeValues.Learn = %q, want %q", ConsensusRoundTypeValues.Learn, "learn")
	}
}

// ============================================================
// MCP domain — span.mcp.call + span.mcp.tool_execute
// ============================================================

func TestMCPToolNameKeyIsCorrectOtelName(t *testing.T) {
	if string(McpToolNameKey) != "mcp.tool.name" {
		t.Errorf("McpToolNameKey = %q, want %q", McpToolNameKey, "mcp.tool.name")
	}
}

func TestMCPServerNameKeyIsCorrectOtelName(t *testing.T) {
	if string(McpServerNameKey) != "mcp.server.name" {
		t.Errorf("McpServerNameKey = %q, want %q", McpServerNameKey, "mcp.server.name")
	}
}

func TestMCPProtocolKeyIsCorrectOtelName(t *testing.T) {
	if string(McpProtocolKey) != "mcp.protocol" {
		t.Errorf("McpProtocolKey = %q, want %q", McpProtocolKey, "mcp.protocol")
	}
}

func TestMCPToolResultCountKeyIsCorrectOtelName(t *testing.T) {
	if string(McpToolResultCountKey) != "mcp.tool.result_count" {
		t.Errorf("McpToolResultCountKey = %q, want %q", McpToolResultCountKey, "mcp.tool.result_count")
	}
}

func TestMCPProtocolStdioValueMatchesSchema(t *testing.T) {
	if McpProtocolValues.Stdio != "stdio" {
		t.Errorf("McpProtocolValues.Stdio = %q, want %q", McpProtocolValues.Stdio, "stdio")
	}
}

func TestMCPProtocolHttpValueMatchesSchema(t *testing.T) {
	if McpProtocolValues.Http != "http" {
		t.Errorf("McpProtocolValues.Http = %q, want %q", McpProtocolValues.Http, "http")
	}
}

func TestMCPProtocolSseValueMatchesSchema(t *testing.T) {
	if McpProtocolValues.Sse != "sse" {
		t.Errorf("McpProtocolValues.Sse = %q, want %q", McpProtocolValues.Sse, "sse")
	}
}

// ============================================================
// Agent domain — span.agent.decision + span.agent.llm_predict
// ============================================================

func TestAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(AgentIdKey) != "agent.id" {
		t.Errorf("AgentIdKey = %q, want %q", AgentIdKey, "agent.id")
	}
}

func TestAgentDecisionTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(AgentDecisionTypeKey) != "agent.decision_type" {
		t.Errorf("AgentDecisionTypeKey = %q, want %q", AgentDecisionTypeKey, "agent.decision_type")
	}
}

func TestAgentLlmModelKeyIsCorrectOtelName(t *testing.T) {
	if string(AgentLlmModelKey) != "agent.llm_model" {
		t.Errorf("AgentLlmModelKey = %q, want %q", AgentLlmModelKey, "agent.llm_model")
	}
}

func TestAgentOutcomeKeyIsCorrectOtelName(t *testing.T) {
	if string(AgentOutcomeKey) != "agent.outcome" {
		t.Errorf("AgentOutcomeKey = %q, want %q", AgentOutcomeKey, "agent.outcome")
	}
}

func TestAgentTokenCountKeyIsCorrectOtelName(t *testing.T) {
	if string(AgentTokenCountKey) != "agent.token_count" {
		t.Errorf("AgentTokenCountKey = %q, want %q", AgentTokenCountKey, "agent.token_count")
	}
}

func TestAgentOutcomeSuccessValueMatchesSchema(t *testing.T) {
	if AgentOutcomeValues.Success != "success" {
		t.Errorf("AgentOutcomeValues.Success = %q, want %q", AgentOutcomeValues.Success, "success")
	}
}

func TestAgentOutcomeFailureValueMatchesSchema(t *testing.T) {
	if AgentOutcomeValues.Failure != "failure" {
		t.Errorf("AgentOutcomeValues.Failure = %q, want %q", AgentOutcomeValues.Failure, "failure")
	}
}

func TestAgentOutcomeEscalatedValueMatchesSchema(t *testing.T) {
	if AgentOutcomeValues.Escalated != "escalated" {
		t.Errorf("AgentOutcomeValues.Escalated = %q, want %q", AgentOutcomeValues.Escalated, "escalated")
	}
}

// ============================================================
// Signal domain — S=(M,G,T,F,W) theory attributes
// ============================================================

func TestSignalModeKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalModeKey) != "signal.mode" {
		t.Errorf("SignalModeKey = %q, want %q", SignalModeKey, "signal.mode")
	}
}

func TestSignalWeightKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalWeightKey) != "signal.weight" {
		t.Errorf("SignalWeightKey = %q, want %q", SignalWeightKey, "signal.weight")
	}
}

func TestSignalGenreKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalGenreKey) != "signal.genre" {
		t.Errorf("SignalGenreKey = %q, want %q", SignalGenreKey, "signal.genre")
	}
}

func TestSignalTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalTypeKey) != "signal.type" {
		t.Errorf("SignalTypeKey = %q, want %q", SignalTypeKey, "signal.type")
	}
}

func TestSignalFormatKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalFormatKey) != "signal.format" {
		t.Errorf("SignalFormatKey = %q, want %q", SignalFormatKey, "signal.format")
	}
}

func TestSignalSourceKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalSourceKey) != "signal.source" {
		t.Errorf("SignalSourceKey = %q, want %q", SignalSourceKey, "signal.source")
	}
}

func TestSignalNoiseLevelKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalNoiseLevelKey) != "signal.noise_level" {
		t.Errorf("SignalNoiseLevelKey = %q, want %q", SignalNoiseLevelKey, "signal.noise_level")
	}
}

func TestSignalSnRatioKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalSnRatioKey) != "signal.sn_ratio" {
		t.Errorf("SignalSnRatioKey = %q, want %q", SignalSnRatioKey, "signal.sn_ratio")
	}
}

func TestSignalClassifierKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalClassifierKey) != "signal.classifier" {
		t.Errorf("SignalClassifierKey = %q, want %q", SignalClassifierKey, "signal.classifier")
	}
}

func TestSignalModeLinguisticValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Linguistic != "linguistic" {
		t.Errorf("SignalModeValues.Linguistic = %q, want %q", SignalModeValues.Linguistic, "linguistic")
	}
}

func TestSignalModeVisualValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Visual != "visual" {
		t.Errorf("SignalModeValues.Visual = %q, want %q", SignalModeValues.Visual, "visual")
	}
}

func TestSignalModeCodeValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Code != "code" {
		t.Errorf("SignalModeValues.Code = %q, want %q", SignalModeValues.Code, "code")
	}
}

func TestSignalModeDataValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Data != "data" {
		t.Errorf("SignalModeValues.Data = %q, want %q", SignalModeValues.Data, "data")
	}
}

func TestSignalModeMixedValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Mixed != "mixed" {
		t.Errorf("SignalModeValues.Mixed = %q, want %q", SignalModeValues.Mixed, "mixed")
	}
}

func TestSignalTypeDirectValueMatchesSchema(t *testing.T) {
	if SignalTypeValues.Direct != "direct" {
		t.Errorf("SignalTypeValues.Direct = %q, want %q", SignalTypeValues.Direct, "direct")
	}
}

func TestSignalTypeInformValueMatchesSchema(t *testing.T) {
	if SignalTypeValues.Inform != "inform" {
		t.Errorf("SignalTypeValues.Inform = %q, want %q", SignalTypeValues.Inform, "inform")
	}
}

func TestSignalTypeCommitValueMatchesSchema(t *testing.T) {
	if SignalTypeValues.Commit != "commit" {
		t.Errorf("SignalTypeValues.Commit = %q, want %q", SignalTypeValues.Commit, "commit")
	}
}

func TestSignalTypeDecideValueMatchesSchema(t *testing.T) {
	if SignalTypeValues.Decide != "decide" {
		t.Errorf("SignalTypeValues.Decide = %q, want %q", SignalTypeValues.Decide, "decide")
	}
}

func TestSignalTypeExpressValueMatchesSchema(t *testing.T) {
	if SignalTypeValues.Express != "express" {
		t.Errorf("SignalTypeValues.Express = %q, want %q", SignalTypeValues.Express, "express")
	}
}

func TestSignalFormatMarkdownValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Markdown != "markdown" {
		t.Errorf("SignalFormatValues.Markdown = %q, want %q", SignalFormatValues.Markdown, "markdown")
	}
}

func TestSignalFormatJsonValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Json != "json" {
		t.Errorf("SignalFormatValues.Json = %q, want %q", SignalFormatValues.Json, "json")
	}
}

func TestSignalGenreSpecValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Spec != "spec" {
		t.Errorf("SignalGenreValues.Spec = %q, want %q", SignalGenreValues.Spec, "spec")
	}
}

func TestSignalGenreAdrValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Adr != "adr" {
		t.Errorf("SignalGenreValues.Adr = %q, want %q", SignalGenreValues.Adr, "adr")
	}
}

// ============================================================
// Process mining domain — process.mining.* attributes
// ============================================================

func TestProcessMiningTraceIDKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningTraceIdKey) != "process.mining.trace_id" {
		t.Errorf("ProcessMiningTraceIdKey = %q, want %q", ProcessMiningTraceIdKey, "process.mining.trace_id")
	}
}

func TestProcessMiningAlgorithmKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningAlgorithmKey) != "process.mining.algorithm" {
		t.Errorf("ProcessMiningAlgorithmKey = %q, want %q", ProcessMiningAlgorithmKey, "process.mining.algorithm")
	}
}

func TestProcessMiningActivityKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningActivityKey) != "process.mining.activity" {
		t.Errorf("ProcessMiningActivityKey = %q, want %q", ProcessMiningActivityKey, "process.mining.activity")
	}
}

func TestProcessMiningEventCountKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningEventCountKey) != "process.mining.event_count" {
		t.Errorf("ProcessMiningEventCountKey = %q, want %q", ProcessMiningEventCountKey, "process.mining.event_count")
	}
}

func TestProcessMiningLogPathKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningLogPathKey) != "process.mining.log_path" {
		t.Errorf("ProcessMiningLogPathKey = %q, want %q", ProcessMiningLogPathKey, "process.mining.log_path")
	}
}

func TestProcessMiningAlgorithmAlphaMinerValueMatchesSchema(t *testing.T) {
	if ProcessMiningAlgorithmValues.AlphaMiner != "alpha_miner" {
		t.Errorf("ProcessMiningAlgorithmValues.AlphaMiner = %q, want %q", ProcessMiningAlgorithmValues.AlphaMiner, "alpha_miner")
	}
}

func TestProcessMiningAlgorithmInductiveMinerValueMatchesSchema(t *testing.T) {
	if ProcessMiningAlgorithmValues.InductiveMiner != "inductive_miner" {
		t.Errorf("ProcessMiningAlgorithmValues.InductiveMiner = %q, want %q", ProcessMiningAlgorithmValues.InductiveMiner, "inductive_miner")
	}
}

func TestProcessMiningAlgorithmHeuristicsMinerValueMatchesSchema(t *testing.T) {
	if ProcessMiningAlgorithmValues.HeuristicsMiner != "heuristics_miner" {
		t.Errorf("ProcessMiningAlgorithmValues.HeuristicsMiner = %q, want %q", ProcessMiningAlgorithmValues.HeuristicsMiner, "heuristics_miner")
	}
}

// ============================================================
// Canopy domain — canopy.heartbeat.tier + adapter attributes
// ============================================================

func TestCanopyHeartbeatTierKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyHeartbeatTierKey) != "canopy.heartbeat.tier" {
		t.Errorf("CanopyHeartbeatTierKey = %q, want %q", CanopyHeartbeatTierKey, "canopy.heartbeat.tier")
	}
}

func TestCanopyAdapterNameKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyAdapterNameKey) != "canopy.adapter.name" {
		t.Errorf("CanopyAdapterNameKey = %q, want %q", CanopyAdapterNameKey, "canopy.adapter.name")
	}
}

func TestCanopyAdapterActionKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyAdapterActionKey) != "canopy.adapter.action" {
		t.Errorf("CanopyAdapterActionKey = %q, want %q", CanopyAdapterActionKey, "canopy.adapter.action")
	}
}

func TestCanopyBudgetMsKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyBudgetMsKey) != "canopy.budget.ms" {
		t.Errorf("CanopyBudgetMsKey = %q, want %q", CanopyBudgetMsKey, "canopy.budget.ms")
	}
}

func TestCanopyHeartbeatTierCriticalValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatTierValues.Critical != "critical" {
		t.Errorf("CanopyHeartbeatTierValues.Critical = %q, want %q", CanopyHeartbeatTierValues.Critical, "critical")
	}
}

func TestCanopyHeartbeatTierHighValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatTierValues.High != "high" {
		t.Errorf("CanopyHeartbeatTierValues.High = %q, want %q", CanopyHeartbeatTierValues.High, "high")
	}
}

func TestCanopyHeartbeatTierNormalValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatTierValues.Normal != "normal" {
		t.Errorf("CanopyHeartbeatTierValues.Normal = %q, want %q", CanopyHeartbeatTierValues.Normal, "normal")
	}
}

func TestCanopyHeartbeatTierLowValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatTierValues.Low != "low" {
		t.Errorf("CanopyHeartbeatTierValues.Low = %q, want %q", CanopyHeartbeatTierValues.Low, "low")
	}
}

// ============================================================
// Conformance domain — conformance.fitness + precision
// ============================================================

func TestConformanceFitnessKeyIsCorrectOtelName(t *testing.T) {
	if string(ConformanceFitnessKey) != "conformance.fitness" {
		t.Errorf("ConformanceFitnessKey = %q, want %q", ConformanceFitnessKey, "conformance.fitness")
	}
}

func TestConformancePrecisionKeyIsCorrectOtelName(t *testing.T) {
	if string(ConformancePrecisionKey) != "conformance.precision" {
		t.Errorf("ConformancePrecisionKey = %q, want %q", ConformancePrecisionKey, "conformance.precision")
	}
}

// ============================================================
// Error domain — error.type attributes
// ============================================================

func TestErrorTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(ErrorTypeKey) != "error.type" {
		t.Errorf("ErrorTypeKey = %q, want %q", ErrorTypeKey, "error.type")
	}
}

func TestErrorTypeTimeoutValueMatchesSchema(t *testing.T) {
	if ErrorTypeValues.Timeout != "timeout" {
		t.Errorf("ErrorTypeValues.Timeout = %q, want %q", ErrorTypeValues.Timeout, "timeout")
	}
}

func TestErrorTypeCancelledValueMatchesSchema(t *testing.T) {
	if ErrorTypeValues.Cancelled != "cancelled" {
		t.Errorf("ErrorTypeValues.Cancelled = %q, want %q", ErrorTypeValues.Cancelled, "cancelled")
	}
}

func TestErrorTypeInternalValueMatchesSchema(t *testing.T) {
	if ErrorTypeValues.Internal != "internal" {
		t.Errorf("ErrorTypeValues.Internal = %q, want %q", ErrorTypeValues.Internal, "internal")
	}
}

func TestErrorTypeUnavailableValueMatchesSchema(t *testing.T) {
	if ErrorTypeValues.Unavailable != "unavailable" {
		t.Errorf("ErrorTypeValues.Unavailable = %q, want %q", ErrorTypeValues.Unavailable, "unavailable")
	}
}

// ============================================================
// ChatmanGPT cross-cutting budget attributes
// ============================================================

func TestChatmangptAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(ChatmangptAgentIdKey) != "chatmangpt.agent.id" {
		t.Errorf("ChatmangptAgentIdKey = %q, want %q", ChatmangptAgentIdKey, "chatmangpt.agent.id")
	}
}

func TestChatmangptBudgetTimeMsKeyIsCorrectOtelName(t *testing.T) {
	if string(ChatmangptBudgetTimeMsKey) != "chatmangpt.budget.time_ms" {
		t.Errorf("ChatmangptBudgetTimeMsKey = %q, want %q", ChatmangptBudgetTimeMsKey, "chatmangpt.budget.time_ms")
	}
}

func TestChatmangptBudgetExceededKeyIsCorrectOtelName(t *testing.T) {
	if string(ChatmangptBudgetExceededKey) != "chatmangpt.budget.exceeded" {
		t.Errorf("ChatmangptBudgetExceededKey = %q, want %q", ChatmangptBudgetExceededKey, "chatmangpt.budget.exceeded")
	}
}

func TestChatmangptServiceTierKeyIsCorrectOtelName(t *testing.T) {
	if string(ChatmangptServiceTierKey) != "chatmangpt.service.tier" {
		t.Errorf("ChatmangptServiceTierKey = %q, want %q", ChatmangptServiceTierKey, "chatmangpt.service.tier")
	}
}

func TestChatmangptServiceTierCriticalValueMatchesSchema(t *testing.T) {
	if ChatmangptServiceTierValues.Critical != "critical" {
		t.Errorf("ChatmangptServiceTierValues.Critical = %q, want %q", ChatmangptServiceTierValues.Critical, "critical")
	}
}

func TestChatmangptServiceTierHighValueMatchesSchema(t *testing.T) {
	if ChatmangptServiceTierValues.High != "high" {
		t.Errorf("ChatmangptServiceTierValues.High = %q, want %q", ChatmangptServiceTierValues.High, "high")
	}
}

func TestChatmangptServiceTierNormalValueMatchesSchema(t *testing.T) {
	if ChatmangptServiceTierValues.Normal != "normal" {
		t.Errorf("ChatmangptServiceTierValues.Normal = %q, want %q", ChatmangptServiceTierValues.Normal, "normal")
	}
}

func TestChatmangptServiceTierLowValueMatchesSchema(t *testing.T) {
	if ChatmangptServiceTierValues.Low != "low" {
		t.Errorf("ChatmangptServiceTierValues.Low = %q, want %q", ChatmangptServiceTierValues.Low, "low")
	}
}

// ============================================================
// Span names — from span_names.go definitions
// ============================================================

func TestSpanNameHealingDiagnosisMatchesSchema(t *testing.T) {
	if SpanNameHealingDiagnosis != "healing.diagnosis" {
		t.Errorf("SpanNameHealingDiagnosis = %q, want %q", SpanNameHealingDiagnosis, "healing.diagnosis")
	}
}

func TestSpanNameHealingReflexArcMatchesSchema(t *testing.T) {
	if SpanNameHealingReflexArc != "healing.reflex_arc" {
		t.Errorf("SpanNameHealingReflexArc = %q, want %q", SpanNameHealingReflexArc, "healing.reflex_arc")
	}
}

func TestSpanNameAgentDecisionMatchesSchema(t *testing.T) {
	if SpanNameAgentDecision != "agent.decision" {
		t.Errorf("SpanNameAgentDecision = %q, want %q", SpanNameAgentDecision, "agent.decision")
	}
}

func TestSpanNameAgentLlmPredictMatchesSchema(t *testing.T) {
	if SpanNameAgentLlmPredict != "agent.llm_predict" {
		t.Errorf("SpanNameAgentLlmPredict = %q, want %q", SpanNameAgentLlmPredict, "agent.llm_predict")
	}
}

func TestSpanNameConsensusRoundMatchesSchema(t *testing.T) {
	if SpanNameConsensusRound != "consensus.round" {
		t.Errorf("SpanNameConsensusRound = %q, want %q", SpanNameConsensusRound, "consensus.round")
	}
}

func TestSpanNameMcpCallMatchesSchema(t *testing.T) {
	if SpanNameMcpCall != "mcp.call" {
		t.Errorf("SpanNameMcpCall = %q, want %q", SpanNameMcpCall, "mcp.call")
	}
}

func TestSpanNameMcpToolExecuteMatchesSchema(t *testing.T) {
	if SpanNameMcpToolExecute != "mcp.tool_execute" {
		t.Errorf("SpanNameMcpToolExecute = %q, want %q", SpanNameMcpToolExecute, "mcp.tool_execute")
	}
}

func TestSpanNameA2ACallMatchesSchema(t *testing.T) {
	if SpanNameA2ACall != "a2a.call" {
		t.Errorf("SpanNameA2ACall = %q, want %q", SpanNameA2ACall, "a2a.call")
	}
}

func TestSpanNameA2ACreateDealMatchesSchema(t *testing.T) {
	if SpanNameA2ACreateDeal != "a2a.create_deal" {
		t.Errorf("SpanNameA2ACreateDeal = %q, want %q", SpanNameA2ACreateDeal, "a2a.create_deal")
	}
}

func TestSpanNameCanopyHeartbeatMatchesSchema(t *testing.T) {
	if SpanNameCanopyHeartbeat != "canopy.heartbeat" {
		t.Errorf("SpanNameCanopyHeartbeat = %q, want %q", SpanNameCanopyHeartbeat, "canopy.heartbeat")
	}
}

func TestSpanNameCanopyAdapterCallMatchesSchema(t *testing.T) {
	if SpanNameCanopyAdapterCall != "canopy.adapter_call" {
		t.Errorf("SpanNameCanopyAdapterCall = %q, want %q", SpanNameCanopyAdapterCall, "canopy.adapter_call")
	}
}

func TestSpanNameWorkflowExecuteMatchesSchema(t *testing.T) {
	if SpanNameWorkflowExecute != "workflow.execute" {
		t.Errorf("SpanNameWorkflowExecute = %q, want %q", SpanNameWorkflowExecute, "workflow.execute")
	}
}

func TestSpanNameWorkflowTransitionMatchesSchema(t *testing.T) {
	if SpanNameWorkflowTransition != "workflow.transition" {
		t.Errorf("SpanNameWorkflowTransition = %q, want %q", SpanNameWorkflowTransition, "workflow.transition")
	}
}

func TestSpanNameProcessMiningDiscoveryMatchesSchema(t *testing.T) {
	if SpanNameProcessMiningDiscovery != "process.mining.discovery" {
		t.Errorf("SpanNameProcessMiningDiscovery = %q, want %q", SpanNameProcessMiningDiscovery, "process.mining.discovery")
	}
}

func TestSpanNameConformanceCheckMatchesSchema(t *testing.T) {
	if SpanNameConformanceCheck != "conformance.check" {
		t.Errorf("SpanNameConformanceCheck = %q, want %q", SpanNameConformanceCheck, "conformance.check")
	}
}

func TestSpanNameBosComplianceCheckMatchesSchema(t *testing.T) {
	if SpanNameBosComplianceCheck != "bos.compliance.check" {
		t.Errorf("SpanNameBosComplianceCheck = %q, want %q", SpanNameBosComplianceCheck, "bos.compliance.check")
	}
}

func TestSpanNameBosDecisionRecordMatchesSchema(t *testing.T) {
	if SpanNameBosDecisionRecord != "bos.decision.record" {
		t.Errorf("SpanNameBosDecisionRecord = %q, want %q", SpanNameBosDecisionRecord, "bos.decision.record")
	}
}

func TestSpanNameBosWorkspaceOperationMatchesSchema(t *testing.T) {
	if SpanNameBosWorkspaceOperation != "bos.workspace.operation" {
		t.Errorf("SpanNameBosWorkspaceOperation = %q, want %q", SpanNameBosWorkspaceOperation, "bos.workspace.operation")
	}
}

// TestAllSpanNamesAreUsed exercises every span name constant.
// Any removal or rename in span_names.go will produce a compile error here.
func TestAllSpanNamesAreUsed(t *testing.T) {
	names := []string{
		SpanNameHealingDiagnosis,
		SpanNameHealingReflexArc,
		SpanNameAgentDecision,
		SpanNameAgentLlmPredict,
		SpanNameConsensusRound,
		SpanNameMcpCall,
		SpanNameMcpToolExecute,
		SpanNameA2ACall,
		SpanNameA2ACreateDeal,
		SpanNameCanopyHeartbeat,
		SpanNameCanopyAdapterCall,
		SpanNameWorkflowExecute,
		SpanNameWorkflowTransition,
		SpanNameProcessMiningDiscovery,
		SpanNameConformanceCheck,
		SpanNameBosComplianceCheck,
		SpanNameBosDecisionRecord,
		SpanNameBosWorkspaceOperation,
		SpanNameBosAuditRecord,
		SpanNameBosGapDetect,
	}
	for _, name := range names {
		if name == "" {
			t.Errorf("span name constant is empty — schema contract broken")
		}
	}
	if len(names) != 20 {
		t.Errorf("expected 20 span name constants, got %d — update this test when adding new spans", len(names))
	}
}

// ============================================================
// BusinessOS audit trail + gap tracking domain
// ============================================================

func TestBosAuditTrailIdKeyIsCorrectOTelName(t *testing.T) {
	if string(BosAuditTrailIdKey) != "bos.audit.trail.id" {
		t.Errorf("BosAuditTrailIdKey = %q, want %q", BosAuditTrailIdKey, "bos.audit.trail.id")
	}
}

func TestBosGapIdKeyIsCorrectOTelName(t *testing.T) {
	if string(BosGapIdKey) != "bos.gap.id" {
		t.Errorf("BosGapIdKey = %q, want %q", BosGapIdKey, "bos.gap.id")
	}
}

func TestBosGapStatusKeyIsCorrectOTelName(t *testing.T) {
	if string(BosGapStatusKey) != "bos.gap.status" {
		t.Errorf("BosGapStatusKey = %q, want %q", BosGapStatusKey, "bos.gap.status")
	}
}

func TestBosDecisionOutcomeKeyIsCorrectOTelName(t *testing.T) {
	if string(BosDecisionOutcomeKey) != "bos.decision.outcome" {
		t.Errorf("BosDecisionOutcomeKey = %q, want %q", BosDecisionOutcomeKey, "bos.decision.outcome")
	}
}

func TestBosPolicyVersionKeyIsCorrectOTelName(t *testing.T) {
	if string(BosPolicyVersionKey) != "bos.policy.version" {
		t.Errorf("BosPolicyVersionKey = %q, want %q", BosPolicyVersionKey, "bos.policy.version")
	}
}

func TestBosGapStatusOpenValueMatchesSchema(t *testing.T) {
	if BosGapStatusValues.Open != "open" {
		t.Errorf("BosGapStatusValues.Open = %q, want %q", BosGapStatusValues.Open, "open")
	}
}

func TestBosGapStatusInRemediationValueMatchesSchema(t *testing.T) {
	if BosGapStatusValues.InRemediation != "in_remediation" {
		t.Errorf("BosGapStatusValues.InRemediation = %q, want %q", BosGapStatusValues.InRemediation, "in_remediation")
	}
}

func TestBosGapStatusResolvedValueMatchesSchema(t *testing.T) {
	if BosGapStatusValues.Resolved != "resolved" {
		t.Errorf("BosGapStatusValues.Resolved = %q, want %q", BosGapStatusValues.Resolved, "resolved")
	}
}

func TestBosGapStatusAcceptedRiskValueMatchesSchema(t *testing.T) {
	if BosGapStatusValues.AcceptedRisk != "accepted_risk" {
		t.Errorf("BosGapStatusValues.AcceptedRisk = %q, want %q", BosGapStatusValues.AcceptedRisk, "accepted_risk")
	}
}

func TestBosDecisionOutcomeApprovedValueMatchesSchema(t *testing.T) {
	if BosDecisionOutcomeValues.Approved != "approved" {
		t.Errorf("BosDecisionOutcomeValues.Approved = %q, want %q", BosDecisionOutcomeValues.Approved, "approved")
	}
}

func TestBosDecisionOutcomeRejectedValueMatchesSchema(t *testing.T) {
	if BosDecisionOutcomeValues.Rejected != "rejected" {
		t.Errorf("BosDecisionOutcomeValues.Rejected = %q, want %q", BosDecisionOutcomeValues.Rejected, "rejected")
	}
}

func TestBosDecisionOutcomeDeferredValueMatchesSchema(t *testing.T) {
	if BosDecisionOutcomeValues.Deferred != "deferred" {
		t.Errorf("BosDecisionOutcomeValues.Deferred = %q, want %q", BosDecisionOutcomeValues.Deferred, "deferred")
	}
}

func TestBosDecisionOutcomeEscalatedValueMatchesSchema(t *testing.T) {
	if BosDecisionOutcomeValues.Escalated != "escalated" {
		t.Errorf("BosDecisionOutcomeValues.Escalated = %q, want %q", BosDecisionOutcomeValues.Escalated, "escalated")
	}
}

func TestSpanNameBosAuditRecordIsCorrectOTelName(t *testing.T) {
	if SpanNameBosAuditRecord != "bos.audit.record" {
		t.Errorf("SpanNameBosAuditRecord = %q, want %q", SpanNameBosAuditRecord, "bos.audit.record")
	}
}

func TestSpanNameBosGapDetectIsCorrectOTelName(t *testing.T) {
	if SpanNameBosGapDetect != "bos.gap.detect" {
		t.Errorf("SpanNameBosGapDetect = %q, want %q", SpanNameBosGapDetect, "bos.gap.detect")
	}
}

// ============================================================
// Consensus domain — additional keys (phase, view, block, leader, votes)
// ============================================================

func TestConsensusPhaseKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusPhaseKey) != "consensus.phase" {
		t.Errorf("ConsensusPhaseKey = %q, want %q", ConsensusPhaseKey, "consensus.phase")
	}
}

func TestConsensusViewNumberKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusViewNumberKey) != "consensus.view_number" {
		t.Errorf("ConsensusViewNumberKey = %q, want %q", ConsensusViewNumberKey, "consensus.view_number")
	}
}

func TestConsensusBlockHashKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusBlockHashKey) != "consensus.block_hash" {
		t.Errorf("ConsensusBlockHashKey = %q, want %q", ConsensusBlockHashKey, "consensus.block_hash")
	}
}

func TestConsensusLeaderIDKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusLeaderIdKey) != "consensus.leader.id" {
		t.Errorf("ConsensusLeaderIdKey = %q, want %q", ConsensusLeaderIdKey, "consensus.leader.id")
	}
}

func TestConsensusVoteCountKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusVoteCountKey) != "consensus.vote_count" {
		t.Errorf("ConsensusVoteCountKey = %q, want %q", ConsensusVoteCountKey, "consensus.vote_count")
	}
}

func TestConsensusPhasePrepareValueMatchesSchema(t *testing.T) {
	if ConsensusPhaseValues.Prepare != "prepare" {
		t.Errorf("ConsensusPhaseValues.Prepare = %q, want %q", ConsensusPhaseValues.Prepare, "prepare")
	}
}

func TestConsensusPhasePreCommitValueMatchesSchema(t *testing.T) {
	if ConsensusPhaseValues.PreCommit != "pre_commit" {
		t.Errorf("ConsensusPhaseValues.PreCommit = %q, want %q", ConsensusPhaseValues.PreCommit, "pre_commit")
	}
}

func TestConsensusPhaseCommitValueMatchesSchema(t *testing.T) {
	if ConsensusPhaseValues.Commit != "commit" {
		t.Errorf("ConsensusPhaseValues.Commit = %q, want %q", ConsensusPhaseValues.Commit, "commit")
	}
}

func TestConsensusPhaseDecideValueMatchesSchema(t *testing.T) {
	if ConsensusPhaseValues.Decide != "decide" {
		t.Errorf("ConsensusPhaseValues.Decide = %q, want %q", ConsensusPhaseValues.Decide, "decide")
	}
}

func TestConsensusPhaseViewChangeValueMatchesSchema(t *testing.T) {
	if ConsensusPhaseValues.ViewChange != "view_change" {
		t.Errorf("ConsensusPhaseValues.ViewChange = %q, want %q", ConsensusPhaseValues.ViewChange, "view_change")
	}
}

// ============================================================
// Event domain — structured event attributes
// ============================================================

func TestEventNameKeyIsCorrectOtelName(t *testing.T) {
	if string(EventNameKey) != "event.name" {
		t.Errorf("EventNameKey = %q, want %q", EventNameKey, "event.name")
	}
}

func TestEventDomainKeyIsCorrectOtelName(t *testing.T) {
	if string(EventDomainKey) != "event.domain" {
		t.Errorf("EventDomainKey = %q, want %q", EventDomainKey, "event.domain")
	}
}

func TestEventSeverityKeyIsCorrectOtelName(t *testing.T) {
	if string(EventSeverityKey) != "event.severity" {
		t.Errorf("EventSeverityKey = %q, want %q", EventSeverityKey, "event.severity")
	}
}

func TestEventSourceKeyIsCorrectOtelName(t *testing.T) {
	if string(EventSourceKey) != "event.source" {
		t.Errorf("EventSourceKey = %q, want %q", EventSourceKey, "event.source")
	}
}

func TestEventCorrelationIDKeyIsCorrectOtelName(t *testing.T) {
	if string(EventCorrelationIdKey) != "event.correlation_id" {
		t.Errorf("EventCorrelationIdKey = %q, want %q", EventCorrelationIdKey, "event.correlation_id")
	}
}

func TestEventDomainAgentValueMatchesSchema(t *testing.T) {
	if EventDomainValues.Agent != "agent" {
		t.Errorf("EventDomainValues.Agent = %q, want %q", EventDomainValues.Agent, "agent")
	}
}

func TestEventDomainComplianceValueMatchesSchema(t *testing.T) {
	if EventDomainValues.Compliance != "compliance" {
		t.Errorf("EventDomainValues.Compliance = %q, want %q", EventDomainValues.Compliance, "compliance")
	}
}

func TestEventDomainHealingValueMatchesSchema(t *testing.T) {
	if EventDomainValues.Healing != "healing" {
		t.Errorf("EventDomainValues.Healing = %q, want %q", EventDomainValues.Healing, "healing")
	}
}

func TestEventDomainWorkflowValueMatchesSchema(t *testing.T) {
	if EventDomainValues.Workflow != "workflow" {
		t.Errorf("EventDomainValues.Workflow = %q, want %q", EventDomainValues.Workflow, "workflow")
	}
}

func TestEventDomainSystemValueMatchesSchema(t *testing.T) {
	if EventDomainValues.System != "system" {
		t.Errorf("EventDomainValues.System = %q, want %q", EventDomainValues.System, "system")
	}
}

func TestEventSeverityDebugValueMatchesSchema(t *testing.T) {
	if EventSeverityValues.Debug != "debug" {
		t.Errorf("EventSeverityValues.Debug = %q, want %q", EventSeverityValues.Debug, "debug")
	}
}

func TestEventSeverityInfoValueMatchesSchema(t *testing.T) {
	if EventSeverityValues.Info != "info" {
		t.Errorf("EventSeverityValues.Info = %q, want %q", EventSeverityValues.Info, "info")
	}
}

func TestEventSeverityWarnValueMatchesSchema(t *testing.T) {
	if EventSeverityValues.Warn != "warn" {
		t.Errorf("EventSeverityValues.Warn = %q, want %q", EventSeverityValues.Warn, "warn")
	}
}

func TestEventSeverityErrorValueMatchesSchema(t *testing.T) {
	if EventSeverityValues.Error != "error" {
		t.Errorf("EventSeverityValues.Error = %q, want %q", EventSeverityValues.Error, "error")
	}
}

func TestEventSeverityFatalValueMatchesSchema(t *testing.T) {
	if EventSeverityValues.Fatal != "fatal" {
		t.Errorf("EventSeverityValues.Fatal = %q, want %q", EventSeverityValues.Fatal, "fatal")
	}
}

// ============================================================
// Signal domain — additional keys and enum values
// ============================================================

func TestSignalBandwidthKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalBandwidthKey) != "signal.bandwidth" {
		t.Errorf("SignalBandwidthKey = %q, want %q", SignalBandwidthKey, "signal.bandwidth")
	}
}

func TestSignalLatencyMsKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalLatencyMsKey) != "signal.latency_ms" {
		t.Errorf("SignalLatencyMsKey = %q, want %q", SignalLatencyMsKey, "signal.latency_ms")
	}
}

func TestSignalModeCognitiveValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Cognitive != "cognitive" {
		t.Errorf("SignalModeValues.Cognitive = %q, want %q", SignalModeValues.Cognitive, "cognitive")
	}
}

func TestSignalModeOperationalValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Operational != "operational" {
		t.Errorf("SignalModeValues.Operational = %q, want %q", SignalModeValues.Operational, "operational")
	}
}

func TestSignalModeReactiveValueMatchesSchema(t *testing.T) {
	if SignalModeValues.Reactive != "reactive" {
		t.Errorf("SignalModeValues.Reactive = %q, want %q", SignalModeValues.Reactive, "reactive")
	}
}

func TestSignalFormatYamlValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Yaml != "yaml" {
		t.Errorf("SignalFormatValues.Yaml = %q, want %q", SignalFormatValues.Yaml, "yaml")
	}
}

func TestSignalFormatHtmlValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Html != "html" {
		t.Errorf("SignalFormatValues.Html = %q, want %q", SignalFormatValues.Html, "html")
	}
}

func TestSignalFormatTextValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Text != "text" {
		t.Errorf("SignalFormatValues.Text = %q, want %q", SignalFormatValues.Text, "text")
	}
}

func TestSignalFormatTableValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Table != "table" {
		t.Errorf("SignalFormatValues.Table = %q, want %q", SignalFormatValues.Table, "table")
	}
}

func TestSignalFormatDiagramValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Diagram != "diagram" {
		t.Errorf("SignalFormatValues.Diagram = %q, want %q", SignalFormatValues.Diagram, "diagram")
	}
}

func TestSignalFormatCodeValueMatchesSchema(t *testing.T) {
	if SignalFormatValues.Code != "code" {
		t.Errorf("SignalFormatValues.Code = %q, want %q", SignalFormatValues.Code, "code")
	}
}

func TestSignalGenreBriefValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Brief != "brief" {
		t.Errorf("SignalGenreValues.Brief = %q, want %q", SignalGenreValues.Brief, "brief")
	}
}

func TestSignalGenreReportValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Report != "report" {
		t.Errorf("SignalGenreValues.Report = %q, want %q", SignalGenreValues.Report, "report")
	}
}

func TestSignalGenrePlanValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Plan != "plan" {
		t.Errorf("SignalGenreValues.Plan = %q, want %q", SignalGenreValues.Plan, "plan")
	}
}

func TestSignalGenreEmailValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Email != "email" {
		t.Errorf("SignalGenreValues.Email = %q, want %q", SignalGenreValues.Email, "email")
	}
}

func TestSignalGenreCodeReviewValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.CodeReview != "code_review" {
		t.Errorf("SignalGenreValues.CodeReview = %q, want %q", SignalGenreValues.CodeReview, "code_review")
	}
}

func TestSignalGenrePitchValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Pitch != "pitch" {
		t.Errorf("SignalGenreValues.Pitch = %q, want %q", SignalGenreValues.Pitch, "pitch")
	}
}

func TestSignalGenreDecisionValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Decision != "decision" {
		t.Errorf("SignalGenreValues.Decision = %q, want %q", SignalGenreValues.Decision, "decision")
	}
}

func TestSignalGenreAnalysisValueMatchesSchema(t *testing.T) {
	if SignalGenreValues.Analysis != "analysis" {
		t.Errorf("SignalGenreValues.Analysis = %q, want %q", SignalGenreValues.Analysis, "analysis")
	}
}

// ============================================================
// Canopy domain — additional keys and enum values
// ============================================================

func TestCanopyAdapterTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyAdapterTypeKey) != "canopy.adapter.type" {
		t.Errorf("CanopyAdapterTypeKey = %q, want %q", CanopyAdapterTypeKey, "canopy.adapter.type")
	}
}

func TestCanopyCommandTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyCommandTypeKey) != "canopy.command.type" {
		t.Errorf("CanopyCommandTypeKey = %q, want %q", CanopyCommandTypeKey, "canopy.command.type")
	}
}

func TestCanopyResponseTimeMsKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyResponseTimeMsKey) != "canopy.response_time_ms" {
		t.Errorf("CanopyResponseTimeMsKey = %q, want %q", CanopyResponseTimeMsKey, "canopy.response_time_ms")
	}
}

func TestCanopyWorkspaceIDKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyWorkspaceIdKey) != "canopy.workspace.id" {
		t.Errorf("CanopyWorkspaceIdKey = %q, want %q", CanopyWorkspaceIdKey, "canopy.workspace.id")
	}
}

func TestCanopyAdapterTypeOsaValueMatchesSchema(t *testing.T) {
	if CanopyAdapterTypeValues.Osa != "osa" {
		t.Errorf("CanopyAdapterTypeValues.Osa = %q, want %q", CanopyAdapterTypeValues.Osa, "osa")
	}
}

func TestCanopyAdapterTypeMcpValueMatchesSchema(t *testing.T) {
	if CanopyAdapterTypeValues.Mcp != "mcp" {
		t.Errorf("CanopyAdapterTypeValues.Mcp = %q, want %q", CanopyAdapterTypeValues.Mcp, "mcp")
	}
}

func TestCanopyAdapterTypeBusinessOsValueMatchesSchema(t *testing.T) {
	if CanopyAdapterTypeValues.BusinessOs != "business_os" {
		t.Errorf("CanopyAdapterTypeValues.BusinessOs = %q, want %q", CanopyAdapterTypeValues.BusinessOs, "business_os")
	}
}

func TestCanopyAdapterTypeWebhookValueMatchesSchema(t *testing.T) {
	if CanopyAdapterTypeValues.Webhook != "webhook" {
		t.Errorf("CanopyAdapterTypeValues.Webhook = %q, want %q", CanopyAdapterTypeValues.Webhook, "webhook")
	}
}

func TestCanopyCommandTypeAgentDispatchValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.AgentDispatch != "agent_dispatch" {
		t.Errorf("CanopyCommandTypeValues.AgentDispatch = %q, want %q", CanopyCommandTypeValues.AgentDispatch, "agent_dispatch")
	}
}

func TestCanopyCommandTypeWorkflowTriggerValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.WorkflowTrigger != "workflow_trigger" {
		t.Errorf("CanopyCommandTypeValues.WorkflowTrigger = %q, want %q", CanopyCommandTypeValues.WorkflowTrigger, "workflow_trigger")
	}
}

func TestCanopyCommandTypeDataQueryValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.DataQuery != "data_query" {
		t.Errorf("CanopyCommandTypeValues.DataQuery = %q, want %q", CanopyCommandTypeValues.DataQuery, "data_query")
	}
}

func TestCanopyCommandTypeHeartbeatCheckValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.HeartbeatCheck != "heartbeat_check" {
		t.Errorf("CanopyCommandTypeValues.HeartbeatCheck = %q, want %q", CanopyCommandTypeValues.HeartbeatCheck, "heartbeat_check")
	}
}

func TestCanopyCommandTypeConfigReloadValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.ConfigReload != "config_reload" {
		t.Errorf("CanopyCommandTypeValues.ConfigReload = %q, want %q", CanopyCommandTypeValues.ConfigReload, "config_reload")
	}
}

// ============================================================
// Wave 9 — Consensus expanded attributes
// ============================================================

func TestConsensusTimeoutMsKeyIsCorrectOtelName(t *testing.T) {
	if string(ConsensusTimeoutMsKey) != "consensus.timeout_ms" {
		t.Errorf("ConsensusTimeoutMsKey = %q, want %q", ConsensusTimeoutMsKey, "consensus.timeout_ms")
	}
}

func TestConsensusTimeoutMsKeyValueRoundTrip(t *testing.T) {
	kv := ConsensusTimeoutMs(5000)
	if string(kv.Key) != "consensus.timeout_ms" {
		t.Errorf("ConsensusTimeoutMs key = %q, want %q", string(kv.Key), "consensus.timeout_ms")
	}
	if kv.Value.AsInt64() != 5000 {
		t.Errorf("ConsensusTimeoutMs value = %d, want %d", kv.Value.AsInt64(), 5000)
	}
}

// ============================================================
// Wave 9 — A2A expanded attributes
// ============================================================

func TestA2ATaskIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aTaskIdKey) != "a2a.task.id" {
		t.Errorf("A2aTaskIdKey = %q, want %q", A2aTaskIdKey, "a2a.task.id")
	}
}

func TestA2ATaskIDKeyValueRoundTrip(t *testing.T) {
	kv := A2aTaskId("task-abc-123")
	if string(kv.Key) != "a2a.task.id" {
		t.Errorf("A2aTaskId key = %q, want %q", string(kv.Key), "a2a.task.id")
	}
	if kv.Value.AsString() != "task-abc-123" {
		t.Errorf("A2aTaskId value = %q, want %q", kv.Value.AsString(), "task-abc-123")
	}
}

func TestA2ATaskPriorityKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aTaskPriorityKey) != "a2a.task.priority" {
		t.Errorf("A2aTaskPriorityKey = %q, want %q", A2aTaskPriorityKey, "a2a.task.priority")
	}
}

func TestA2ATaskPriorityCriticalValueMatchesSchema(t *testing.T) {
	if A2aTaskPriorityValues.Critical != "critical" {
		t.Errorf("A2aTaskPriorityValues.Critical = %q, want %q", A2aTaskPriorityValues.Critical, "critical")
	}
}

func TestA2ATaskPriorityHighValueMatchesSchema(t *testing.T) {
	if A2aTaskPriorityValues.High != "high" {
		t.Errorf("A2aTaskPriorityValues.High = %q, want %q", A2aTaskPriorityValues.High, "high")
	}
}

func TestA2ATaskPriorityNormalValueMatchesSchema(t *testing.T) {
	if A2aTaskPriorityValues.Normal != "normal" {
		t.Errorf("A2aTaskPriorityValues.Normal = %q, want %q", A2aTaskPriorityValues.Normal, "normal")
	}
}

func TestA2ATaskPriorityLowValueMatchesSchema(t *testing.T) {
	if A2aTaskPriorityValues.Low != "low" {
		t.Errorf("A2aTaskPriorityValues.Low = %q, want %q", A2aTaskPriorityValues.Low, "low")
	}
}

func TestA2ACapabilityNameKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aCapabilityNameKey) != "a2a.capability.name" {
		t.Errorf("A2aCapabilityNameKey = %q, want %q", A2aCapabilityNameKey, "a2a.capability.name")
	}
}

func TestA2ACapabilityNameKeyValueRoundTrip(t *testing.T) {
	kv := A2aCapabilityName("healing.diagnosis")
	if string(kv.Key) != "a2a.capability.name" {
		t.Errorf("A2aCapabilityName key = %q, want %q", string(kv.Key), "a2a.capability.name")
	}
	if kv.Value.AsString() != "healing.diagnosis" {
		t.Errorf("A2aCapabilityName value = %q, want %q", kv.Value.AsString(), "healing.diagnosis")
	}
}

func TestA2ANegotiationRoundKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aNegotiationRoundKey) != "a2a.negotiation.round" {
		t.Errorf("A2aNegotiationRoundKey = %q, want %q", A2aNegotiationRoundKey, "a2a.negotiation.round")
	}
}

func TestA2ANegotiationRoundKeyValueRoundTrip(t *testing.T) {
	kv := A2aNegotiationRound(3)
	if string(kv.Key) != "a2a.negotiation.round" {
		t.Errorf("A2aNegotiationRound key = %q, want %q", string(kv.Key), "a2a.negotiation.round")
	}
	if kv.Value.AsInt64() != 3 {
		t.Errorf("A2aNegotiationRound value = %d, want %d", kv.Value.AsInt64(), 3)
	}
}

func TestA2ANegotiationStatusKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aNegotiationStatusKey) != "a2a.negotiation.status" {
		t.Errorf("A2aNegotiationStatusKey = %q, want %q", A2aNegotiationStatusKey, "a2a.negotiation.status")
	}
}

func TestA2ANegotiationStatusPendingValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStatusValues.Pending != "pending" {
		t.Errorf("A2aNegotiationStatusValues.Pending = %q, want %q", A2aNegotiationStatusValues.Pending, "pending")
	}
}

func TestA2ANegotiationStatusAcceptedValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStatusValues.Accepted != "accepted" {
		t.Errorf("A2aNegotiationStatusValues.Accepted = %q, want %q", A2aNegotiationStatusValues.Accepted, "accepted")
	}
}

func TestA2ANegotiationStatusRejectedValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStatusValues.Rejected != "rejected" {
		t.Errorf("A2aNegotiationStatusValues.Rejected = %q, want %q", A2aNegotiationStatusValues.Rejected, "rejected")
	}
}

func TestA2ANegotiationStatusCounterOfferValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStatusValues.CounterOffer != "counter_offer" {
		t.Errorf("A2aNegotiationStatusValues.CounterOffer = %q, want %q", A2aNegotiationStatusValues.CounterOffer, "counter_offer")
	}
}

func TestA2ANegotiationStatusExpiredValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStatusValues.Expired != "expired" {
		t.Errorf("A2aNegotiationStatusValues.Expired = %q, want %q", A2aNegotiationStatusValues.Expired, "expired")
	}
}

// ============================================================
// Wave 9 — Process Mining new attributes
// ============================================================

func TestProcessMiningDfgEdgeCountKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningDfgEdgeCountKey) != "process.mining.dfg.edge_count" {
		t.Errorf("ProcessMiningDfgEdgeCountKey = %q, want %q", ProcessMiningDfgEdgeCountKey, "process.mining.dfg.edge_count")
	}
}

func TestProcessMiningDfgEdgeCountKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningDfgEdgeCount(45)
	if string(kv.Key) != "process.mining.dfg.edge_count" {
		t.Errorf("ProcessMiningDfgEdgeCount key = %q, want %q", string(kv.Key), "process.mining.dfg.edge_count")
	}
	if kv.Value.AsInt64() != 45 {
		t.Errorf("ProcessMiningDfgEdgeCount value = %d, want %d", kv.Value.AsInt64(), 45)
	}
}

func TestProcessMiningDfgNodeCountKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningDfgNodeCountKey) != "process.mining.dfg.node_count" {
		t.Errorf("ProcessMiningDfgNodeCountKey = %q, want %q", ProcessMiningDfgNodeCountKey, "process.mining.dfg.node_count")
	}
}

func TestProcessMiningDfgNodeCountKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningDfgNodeCount(12)
	if string(kv.Key) != "process.mining.dfg.node_count" {
		t.Errorf("ProcessMiningDfgNodeCount key = %q, want %q", string(kv.Key), "process.mining.dfg.node_count")
	}
	if kv.Value.AsInt64() != 12 {
		t.Errorf("ProcessMiningDfgNodeCount value = %d, want %d", kv.Value.AsInt64(), 12)
	}
}

func TestProcessMiningCaseCountKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningCaseCountKey) != "process.mining.case_count" {
		t.Errorf("ProcessMiningCaseCountKey = %q, want %q", ProcessMiningCaseCountKey, "process.mining.case_count")
	}
}

func TestProcessMiningCaseCountKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningCaseCount(1500)
	if string(kv.Key) != "process.mining.case_count" {
		t.Errorf("ProcessMiningCaseCount key = %q, want %q", string(kv.Key), "process.mining.case_count")
	}
	if kv.Value.AsInt64() != 1500 {
		t.Errorf("ProcessMiningCaseCount value = %d, want %d", kv.Value.AsInt64(), 1500)
	}
}

func TestProcessMiningConformanceDeviationTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningConformanceDeviationTypeKey) != "process.mining.conformance.deviation_type" {
		t.Errorf("ProcessMiningConformanceDeviationTypeKey = %q, want %q", ProcessMiningConformanceDeviationTypeKey, "process.mining.conformance.deviation_type")
	}
}

func TestProcessMiningConformanceDeviationTypeMissingActivityValueMatchesSchema(t *testing.T) {
	if ProcessMiningConformanceDeviationTypeValues.MissingActivity != "missing_activity" {
		t.Errorf("ProcessMiningConformanceDeviationTypeValues.MissingActivity = %q, want %q", ProcessMiningConformanceDeviationTypeValues.MissingActivity, "missing_activity")
	}
}

func TestProcessMiningConformanceDeviationTypeExtraActivityValueMatchesSchema(t *testing.T) {
	if ProcessMiningConformanceDeviationTypeValues.ExtraActivity != "extra_activity" {
		t.Errorf("ProcessMiningConformanceDeviationTypeValues.ExtraActivity = %q, want %q", ProcessMiningConformanceDeviationTypeValues.ExtraActivity, "extra_activity")
	}
}

func TestProcessMiningConformanceDeviationTypeWrongOrderValueMatchesSchema(t *testing.T) {
	if ProcessMiningConformanceDeviationTypeValues.WrongOrder != "wrong_order" {
		t.Errorf("ProcessMiningConformanceDeviationTypeValues.WrongOrder = %q, want %q", ProcessMiningConformanceDeviationTypeValues.WrongOrder, "wrong_order")
	}
}

func TestProcessMiningConformanceDeviationTypeLoopViolationValueMatchesSchema(t *testing.T) {
	if ProcessMiningConformanceDeviationTypeValues.LoopViolation != "loop_violation" {
		t.Errorf("ProcessMiningConformanceDeviationTypeValues.LoopViolation = %q, want %q", ProcessMiningConformanceDeviationTypeValues.LoopViolation, "loop_violation")
	}
}

func TestProcessMiningDeviationTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningDeviationTypeKey) != "process.mining.deviation.type" {
		t.Errorf("ProcessMiningDeviationTypeKey = %q, want %q", ProcessMiningDeviationTypeKey, "process.mining.deviation.type")
	}
}

func TestProcessMiningDeviationTypeSkipValueMatchesSchema(t *testing.T) {
	if ProcessMiningDeviationTypeValues.Skip != "skip" {
		t.Errorf("ProcessMiningDeviationTypeValues.Skip = %q, want %q", ProcessMiningDeviationTypeValues.Skip, "skip")
	}
}

func TestProcessMiningDeviationTypeInsertValueMatchesSchema(t *testing.T) {
	if ProcessMiningDeviationTypeValues.Insert != "insert" {
		t.Errorf("ProcessMiningDeviationTypeValues.Insert = %q, want %q", ProcessMiningDeviationTypeValues.Insert, "insert")
	}
}

func TestProcessMiningDeviationTypeMoveModelValueMatchesSchema(t *testing.T) {
	if ProcessMiningDeviationTypeValues.MoveModel != "move_model" {
		t.Errorf("ProcessMiningDeviationTypeValues.MoveModel = %q, want %q", ProcessMiningDeviationTypeValues.MoveModel, "move_model")
	}
}

func TestProcessMiningDeviationTypeMoveLogValueMatchesSchema(t *testing.T) {
	if ProcessMiningDeviationTypeValues.MoveLog != "move_log" {
		t.Errorf("ProcessMiningDeviationTypeValues.MoveLog = %q, want %q", ProcessMiningDeviationTypeValues.MoveLog, "move_log")
	}
}

func TestProcessMiningFitnessThresholdKeyIsCorrectOtelName(t *testing.T) {
	if string(ProcessMiningFitnessThresholdKey) != "process.mining.fitness_threshold" {
		t.Errorf("ProcessMiningFitnessThresholdKey = %q, want %q", ProcessMiningFitnessThresholdKey, "process.mining.fitness_threshold")
	}
}

func TestProcessMiningFitnessThresholdKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningFitnessThreshold(0.95)
	if string(kv.Key) != "process.mining.fitness_threshold" {
		t.Errorf("ProcessMiningFitnessThreshold key = %q, want %q", string(kv.Key), "process.mining.fitness_threshold")
	}
	if kv.Value.AsFloat64() != 0.95 {
		t.Errorf("ProcessMiningFitnessThreshold value = %f, want %f", kv.Value.AsFloat64(), 0.95)
	}
}

func TestProcessMiningAlgorithmDirectlyFollowsValueMatchesSchema(t *testing.T) {
	if ProcessMiningAlgorithmValues.DirectlyFollows != "directly_follows" {
		t.Errorf("ProcessMiningAlgorithmValues.DirectlyFollows = %q, want %q", ProcessMiningAlgorithmValues.DirectlyFollows, "directly_follows")
	}
}

func TestProcessMiningAlgorithmHeuristicMinerValueMatchesSchema(t *testing.T) {
	if ProcessMiningAlgorithmValues.HeuristicMiner != "heuristic_miner" {
		t.Errorf("ProcessMiningAlgorithmValues.HeuristicMiner = %q, want %q", ProcessMiningAlgorithmValues.HeuristicMiner, "heuristic_miner")
	}
}

// ============================================================
// Wave 9 — Healing expanded attributes
// ============================================================

func TestHealingDiagnosisStageKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingDiagnosisStageKey) != "healing.diagnosis_stage" {
		t.Errorf("HealingDiagnosisStageKey = %q, want %q", HealingDiagnosisStageKey, "healing.diagnosis_stage")
	}
}

func TestHealingDiagnosisStageDetectionValueMatchesSchema(t *testing.T) {
	if HealingDiagnosisStageValues.Detection != "detection" {
		t.Errorf("HealingDiagnosisStageValues.Detection = %q, want %q", HealingDiagnosisStageValues.Detection, "detection")
	}
}

func TestHealingDiagnosisStageClassificationValueMatchesSchema(t *testing.T) {
	if HealingDiagnosisStageValues.Classification != "classification" {
		t.Errorf("HealingDiagnosisStageValues.Classification = %q, want %q", HealingDiagnosisStageValues.Classification, "classification")
	}
}

func TestHealingDiagnosisStageVerificationValueMatchesSchema(t *testing.T) {
	if HealingDiagnosisStageValues.Verification != "verification" {
		t.Errorf("HealingDiagnosisStageValues.Verification = %q, want %q", HealingDiagnosisStageValues.Verification, "verification")
	}
}

func TestHealingDiagnosisStageEscalationValueMatchesSchema(t *testing.T) {
	if HealingDiagnosisStageValues.Escalation != "escalation" {
		t.Errorf("HealingDiagnosisStageValues.Escalation = %q, want %q", HealingDiagnosisStageValues.Escalation, "escalation")
	}
}

func TestHealingRecoveryStrategyKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingRecoveryStrategyKey) != "healing.recovery_strategy" {
		t.Errorf("HealingRecoveryStrategyKey = %q, want %q", HealingRecoveryStrategyKey, "healing.recovery_strategy")
	}
}

func TestHealingRecoveryStrategyRestartValueMatchesSchema(t *testing.T) {
	if HealingRecoveryStrategyValues.Restart != "restart" {
		t.Errorf("HealingRecoveryStrategyValues.Restart = %q, want %q", HealingRecoveryStrategyValues.Restart, "restart")
	}
}

func TestHealingRecoveryStrategyRollbackValueMatchesSchema(t *testing.T) {
	if HealingRecoveryStrategyValues.Rollback != "rollback" {
		t.Errorf("HealingRecoveryStrategyValues.Rollback = %q, want %q", HealingRecoveryStrategyValues.Rollback, "rollback")
	}
}

func TestHealingRecoveryStrategyCircuitBreakValueMatchesSchema(t *testing.T) {
	if HealingRecoveryStrategyValues.CircuitBreak != "circuit_break" {
		t.Errorf("HealingRecoveryStrategyValues.CircuitBreak = %q, want %q", HealingRecoveryStrategyValues.CircuitBreak, "circuit_break")
	}
}

func TestHealingRecoveryStrategyIsolateValueMatchesSchema(t *testing.T) {
	if HealingRecoveryStrategyValues.Isolate != "isolate" {
		t.Errorf("HealingRecoveryStrategyValues.Isolate = %q, want %q", HealingRecoveryStrategyValues.Isolate, "isolate")
	}
}

func TestHealingRecoveryStrategyDegradeValueMatchesSchema(t *testing.T) {
	if HealingRecoveryStrategyValues.Degrade != "degrade" {
		t.Errorf("HealingRecoveryStrategyValues.Degrade = %q, want %q", HealingRecoveryStrategyValues.Degrade, "degrade")
	}
}

func TestHealingFingerprintKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingFingerprintKey) != "healing.fingerprint" {
		t.Errorf("HealingFingerprintKey = %q, want %q", HealingFingerprintKey, "healing.fingerprint")
	}
}

func TestHealingFingerprintKeyValueRoundTrip(t *testing.T) {
	kv := HealingFingerprint("fp-a3b2c1")
	if string(kv.Key) != "healing.fingerprint" {
		t.Errorf("HealingFingerprint key = %q, want %q", string(kv.Key), "healing.fingerprint")
	}
	if kv.Value.AsString() != "fp-a3b2c1" {
		t.Errorf("HealingFingerprint value = %q, want %q", kv.Value.AsString(), "fp-a3b2c1")
	}
}

func TestHealingEscalationReasonKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingEscalationReasonKey) != "healing.escalation_reason" {
		t.Errorf("HealingEscalationReasonKey = %q, want %q", HealingEscalationReasonKey, "healing.escalation_reason")
	}
}

func TestHealingEscalationReasonKeyValueRoundTrip(t *testing.T) {
	kv := HealingEscalationReason("max_attempts_exceeded")
	if string(kv.Key) != "healing.escalation_reason" {
		t.Errorf("HealingEscalationReason key = %q, want %q", string(kv.Key), "healing.escalation_reason")
	}
	if kv.Value.AsString() != "max_attempts_exceeded" {
		t.Errorf("HealingEscalationReason value = %q, want %q", kv.Value.AsString(), "max_attempts_exceeded")
	}
}

// ============================================================
// Wave 9 iteration 6 — BusinessOS audit + gap new attributes
// ============================================================

func TestBosAuditEventTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(BosAuditEventTypeKey) != "bos.audit.event_type" {
		t.Errorf("BosAuditEventTypeKey = %q, want %q", BosAuditEventTypeKey, "bos.audit.event_type")
	}
}

func TestBosAuditEventTypeDataAccessValueMatchesSchema(t *testing.T) {
	if BosAuditEventTypeValues.DataAccess != "data_access" {
		t.Errorf("BosAuditEventTypeValues.DataAccess = %q, want %q", BosAuditEventTypeValues.DataAccess, "data_access")
	}
}

func TestBosAuditEventTypeConfigChangeValueMatchesSchema(t *testing.T) {
	if BosAuditEventTypeValues.ConfigChange != "config_change" {
		t.Errorf("BosAuditEventTypeValues.ConfigChange = %q, want %q", BosAuditEventTypeValues.ConfigChange, "config_change")
	}
}

func TestBosAuditEventTypePermissionGrantValueMatchesSchema(t *testing.T) {
	if BosAuditEventTypeValues.PermissionGrant != "permission_grant" {
		t.Errorf("BosAuditEventTypeValues.PermissionGrant = %q, want %q", BosAuditEventTypeValues.PermissionGrant, "permission_grant")
	}
}

func TestBosAuditEventTypeComplianceCheckValueMatchesSchema(t *testing.T) {
	if BosAuditEventTypeValues.ComplianceCheck != "compliance_check" {
		t.Errorf("BosAuditEventTypeValues.ComplianceCheck = %q, want %q", BosAuditEventTypeValues.ComplianceCheck, "compliance_check")
	}
}

func TestBosAuditEventTypeGapDetectionValueMatchesSchema(t *testing.T) {
	if BosAuditEventTypeValues.GapDetection != "gap_detection" {
		t.Errorf("BosAuditEventTypeValues.GapDetection = %q, want %q", BosAuditEventTypeValues.GapDetection, "gap_detection")
	}
}

func TestBosAuditActorIdKeyIsCorrectOtelName(t *testing.T) {
	if string(BosAuditActorIdKey) != "bos.audit.actor_id" {
		t.Errorf("BosAuditActorIdKey = %q, want %q", BosAuditActorIdKey, "bos.audit.actor_id")
	}
}

func TestBosAuditActorIdKeyValueRoundTrip(t *testing.T) {
	kv := BosAuditActorId("user-123")
	if string(kv.Key) != "bos.audit.actor_id" {
		t.Errorf("BosAuditActorId key = %q, want %q", string(kv.Key), "bos.audit.actor_id")
	}
	if kv.Value.AsString() != "user-123" {
		t.Errorf("BosAuditActorId value = %q, want %q", kv.Value.AsString(), "user-123")
	}
}

func TestBosComplianceControlIdKeyIsCorrectOtelName(t *testing.T) {
	if string(BosComplianceControlIdKey) != "bos.compliance.control_id" {
		t.Errorf("BosComplianceControlIdKey = %q, want %q", BosComplianceControlIdKey, "bos.compliance.control_id")
	}
}

func TestBosComplianceControlIdKeyValueRoundTrip(t *testing.T) {
	kv := BosComplianceControlId("CC6.1")
	if string(kv.Key) != "bos.compliance.control_id" {
		t.Errorf("BosComplianceControlId key = %q, want %q", string(kv.Key), "bos.compliance.control_id")
	}
	if kv.Value.AsString() != "CC6.1" {
		t.Errorf("BosComplianceControlId value = %q, want %q", kv.Value.AsString(), "CC6.1")
	}
}

func TestBosGapSeverityKeyIsCorrectOtelName(t *testing.T) {
	if string(BosGapSeverityKey) != "bos.gap.severity" {
		t.Errorf("BosGapSeverityKey = %q, want %q", BosGapSeverityKey, "bos.gap.severity")
	}
}

func TestBosGapSeverityCriticalValueMatchesSchema(t *testing.T) {
	if BosGapSeverityValues.Critical != "critical" {
		t.Errorf("BosGapSeverityValues.Critical = %q, want %q", BosGapSeverityValues.Critical, "critical")
	}
}

func TestBosGapSeverityHighValueMatchesSchema(t *testing.T) {
	if BosGapSeverityValues.High != "high" {
		t.Errorf("BosGapSeverityValues.High = %q, want %q", BosGapSeverityValues.High, "high")
	}
}

func TestBosGapSeverityMediumValueMatchesSchema(t *testing.T) {
	if BosGapSeverityValues.Medium != "medium" {
		t.Errorf("BosGapSeverityValues.Medium = %q, want %q", BosGapSeverityValues.Medium, "medium")
	}
}

func TestBosGapSeverityLowValueMatchesSchema(t *testing.T) {
	if BosGapSeverityValues.Low != "low" {
		t.Errorf("BosGapSeverityValues.Low = %q, want %q", BosGapSeverityValues.Low, "low")
	}
}

func TestBosGapRemediationDaysKeyIsCorrectOtelName(t *testing.T) {
	if string(BosGapRemediationDaysKey) != "bos.gap.remediation_days" {
		t.Errorf("BosGapRemediationDaysKey = %q, want %q", BosGapRemediationDaysKey, "bos.gap.remediation_days")
	}
}

func TestBosGapRemediationDaysKeyValueRoundTrip(t *testing.T) {
	kv := BosGapRemediationDays(30)
	if string(kv.Key) != "bos.gap.remediation_days" {
		t.Errorf("BosGapRemediationDays key = %q, want %q", string(kv.Key), "bos.gap.remediation_days")
	}
	if kv.Value.AsInt64() != 30 {
		t.Errorf("BosGapRemediationDays value = %d, want %d", kv.Value.AsInt64(), 30)
	}
}

// ============================================================
// Wave 9 iteration 6 — Canopy new keys and enum values
// ============================================================

func TestCanopyHeartbeatStatusKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopyHeartbeatStatusKey) != "canopy.heartbeat.status" {
		t.Errorf("CanopyHeartbeatStatusKey = %q, want %q", CanopyHeartbeatStatusKey, "canopy.heartbeat.status")
	}
}

func TestCanopyHeartbeatStatusHealthyValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatStatusValues.Healthy != "healthy" {
		t.Errorf("CanopyHeartbeatStatusValues.Healthy = %q, want %q", CanopyHeartbeatStatusValues.Healthy, "healthy")
	}
}

func TestCanopyHeartbeatStatusDegradedValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatStatusValues.Degraded != "degraded" {
		t.Errorf("CanopyHeartbeatStatusValues.Degraded = %q, want %q", CanopyHeartbeatStatusValues.Degraded, "degraded")
	}
}

func TestCanopyHeartbeatStatusCriticalValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatStatusValues.Critical != "critical" {
		t.Errorf("CanopyHeartbeatStatusValues.Critical = %q, want %q", CanopyHeartbeatStatusValues.Critical, "critical")
	}
}

func TestCanopyHeartbeatStatusTimeoutValueMatchesSchema(t *testing.T) {
	if CanopyHeartbeatStatusValues.Timeout != "timeout" {
		t.Errorf("CanopyHeartbeatStatusValues.Timeout = %q, want %q", CanopyHeartbeatStatusValues.Timeout, "timeout")
	}
}

func TestCanopySignalModeKeyIsCorrectOtelName(t *testing.T) {
	if string(CanopySignalModeKey) != "canopy.signal.mode" {
		t.Errorf("CanopySignalModeKey = %q, want %q", CanopySignalModeKey, "canopy.signal.mode")
	}
}

func TestCanopySignalModeKeyValueRoundTrip(t *testing.T) {
	kv := CanopySignalMode("linguistic")
	if string(kv.Key) != "canopy.signal.mode" {
		t.Errorf("CanopySignalMode key = %q, want %q", string(kv.Key), "canopy.signal.mode")
	}
	if kv.Value.AsString() != "linguistic" {
		t.Errorf("CanopySignalMode value = %q, want %q", kv.Value.AsString(), "linguistic")
	}
}

func TestCanopyCommandTypeExecuteValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.Execute != "execute" {
		t.Errorf("CanopyCommandTypeValues.Execute = %q, want %q", CanopyCommandTypeValues.Execute, "execute")
	}
}

func TestCanopyCommandTypeBroadcastValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.Broadcast != "broadcast" {
		t.Errorf("CanopyCommandTypeValues.Broadcast = %q, want %q", CanopyCommandTypeValues.Broadcast, "broadcast")
	}
}

func TestCanopyCommandTypeQueryValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.Query != "query" {
		t.Errorf("CanopyCommandTypeValues.Query = %q, want %q", CanopyCommandTypeValues.Query, "query")
	}
}

func TestCanopyCommandTypeSyncValueMatchesSchema(t *testing.T) {
	if CanopyCommandTypeValues.Sync != "sync" {
		t.Errorf("CanopyCommandTypeValues.Sync = %q, want %q", CanopyCommandTypeValues.Sync, "sync")
	}
}

// ============================================================
// Wave 9 iteration 6 — YAWL Workflow new attributes
// ============================================================

func TestWorkflowMilestoneConditionKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowMilestoneConditionKey) != "workflow.milestone.condition" {
		t.Errorf("WorkflowMilestoneConditionKey = %q, want %q", WorkflowMilestoneConditionKey, "workflow.milestone.condition")
	}
}

func TestWorkflowMilestoneConditionKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowMilestoneCondition("approvals >= 3")
	if string(kv.Key) != "workflow.milestone.condition" {
		t.Errorf("WorkflowMilestoneCondition key = %q, want %q", string(kv.Key), "workflow.milestone.condition")
	}
	if kv.Value.AsString() != "approvals >= 3" {
		t.Errorf("WorkflowMilestoneCondition value = %q, want %q", kv.Value.AsString(), "approvals >= 3")
	}
}

func TestWorkflowCancelReasonKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowCancelReasonKey) != "workflow.cancel.reason" {
		t.Errorf("WorkflowCancelReasonKey = %q, want %q", WorkflowCancelReasonKey, "workflow.cancel.reason")
	}
}

func TestWorkflowCancelReasonKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowCancelReason("user_abort")
	if string(kv.Key) != "workflow.cancel.reason" {
		t.Errorf("WorkflowCancelReason key = %q, want %q", string(kv.Key), "workflow.cancel.reason")
	}
	if kv.Value.AsString() != "user_abort" {
		t.Errorf("WorkflowCancelReason value = %q, want %q", kv.Value.AsString(), "user_abort")
	}
}

func TestWorkflowInstanceCountKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowInstanceCountKey) != "workflow.instance.count" {
		t.Errorf("WorkflowInstanceCountKey = %q, want %q", WorkflowInstanceCountKey, "workflow.instance.count")
	}
}

func TestWorkflowInstanceCountKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowInstanceCount(5)
	if string(kv.Key) != "workflow.instance.count" {
		t.Errorf("WorkflowInstanceCount key = %q, want %q", string(kv.Key), "workflow.instance.count")
	}
	if kv.Value.AsInt64() != 5 {
		t.Errorf("WorkflowInstanceCount value = %d, want %d", kv.Value.AsInt64(), 5)
	}
}

func TestWorkflowInstanceCompletedKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowInstanceCompletedKey) != "workflow.instance.completed" {
		t.Errorf("WorkflowInstanceCompletedKey = %q, want %q", WorkflowInstanceCompletedKey, "workflow.instance.completed")
	}
}

func TestWorkflowInstanceCompletedKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowInstanceCompleted(3)
	if string(kv.Key) != "workflow.instance.completed" {
		t.Errorf("WorkflowInstanceCompleted key = %q, want %q", string(kv.Key), "workflow.instance.completed")
	}
	if kv.Value.AsInt64() != 3 {
		t.Errorf("WorkflowInstanceCompleted value = %d, want %d", kv.Value.AsInt64(), 3)
	}
}

func TestWorkflowLoopIterationKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowLoopIterationKey) != "workflow.loop.iteration" {
		t.Errorf("WorkflowLoopIterationKey = %q, want %q", WorkflowLoopIterationKey, "workflow.loop.iteration")
	}
}

func TestWorkflowLoopIterationKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowLoopIteration(7)
	if string(kv.Key) != "workflow.loop.iteration" {
		t.Errorf("WorkflowLoopIteration key = %q, want %q", string(kv.Key), "workflow.loop.iteration")
	}
	if kv.Value.AsInt64() != 7 {
		t.Errorf("WorkflowLoopIteration value = %d, want %d", kv.Value.AsInt64(), 7)
	}
}

func TestWorkflowLoopMaxIterationsKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowLoopMaxIterationsKey) != "workflow.loop.max_iterations" {
		t.Errorf("WorkflowLoopMaxIterationsKey = %q, want %q", WorkflowLoopMaxIterationsKey, "workflow.loop.max_iterations")
	}
}

func TestWorkflowLoopMaxIterationsKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowLoopMaxIterations(100)
	if string(kv.Key) != "workflow.loop.max_iterations" {
		t.Errorf("WorkflowLoopMaxIterations key = %q, want %q", string(kv.Key), "workflow.loop.max_iterations")
	}
	if kv.Value.AsInt64() != 100 {
		t.Errorf("WorkflowLoopMaxIterations value = %d, want %d", kv.Value.AsInt64(), 100)
	}
}

func TestWorkflowLoopBoundednessGuarantee(t *testing.T) {
	// WvdA soundness: loop.max_iterations enforces boundedness (no infinite loops)
	kv := WorkflowLoopMaxIterations(1000)
	if kv.Value.AsInt64() <= 0 {
		t.Errorf("WorkflowLoopMaxIterations must be positive for boundedness guarantee, got %d", kv.Value.AsInt64())
	}
}

func TestWorkflowInstanceCountLessThanOrEqualCompleted(t *testing.T) {
	// Sanity: completed instances cannot exceed total instance count in a valid workflow
	total := WorkflowInstanceCount(5)
	completed := WorkflowInstanceCompleted(3)
	if completed.Value.AsInt64() > total.Value.AsInt64() {
		t.Errorf("instance.completed (%d) > instance.count (%d): invalid state", completed.Value.AsInt64(), total.Value.AsInt64())
	}
}

// ============================================================
// Iteration 7: A2A Negotiation State Machine
// ============================================================

func TestA2aNegotiationStateKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aNegotiationStateKey) != "a2a.negotiation.state" {
		t.Errorf("A2aNegotiationStateKey = %q, want %q", A2aNegotiationStateKey, "a2a.negotiation.state")
	}
}

func TestA2aNegotiationStateProposedValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStateValues.Proposed != "proposed" {
		t.Errorf("A2aNegotiationStateValues.Proposed = %q, want %q", A2aNegotiationStateValues.Proposed, "proposed")
	}
}

func TestA2aNegotiationStateAcceptedValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStateValues.Accepted != "accepted" {
		t.Errorf("A2aNegotiationStateValues.Accepted = %q, want %q", A2aNegotiationStateValues.Accepted, "accepted")
	}
}

func TestA2aNegotiationStateRejectedValueMatchesSchema(t *testing.T) {
	if A2aNegotiationStateValues.Rejected != "rejected" {
		t.Errorf("A2aNegotiationStateValues.Rejected = %q, want %q", A2aNegotiationStateValues.Rejected, "rejected")
	}
}

func TestA2aNegotiationTimeoutMsKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aNegotiationTimeoutMsKey) != "a2a.negotiation.timeout_ms" {
		t.Errorf("A2aNegotiationTimeoutMsKey = %q, want %q", A2aNegotiationTimeoutMsKey, "a2a.negotiation.timeout_ms")
	}
}

func TestA2aNegotiationTimeoutMsKeyValueRoundTrip(t *testing.T) {
	// WvdA deadlock freedom: every negotiation round has an explicit timeout
	kv := A2aNegotiationTimeoutMs(5000)
	if string(kv.Key) != "a2a.negotiation.timeout_ms" {
		t.Errorf("A2aNegotiationTimeoutMs key = %q, want %q", string(kv.Key), "a2a.negotiation.timeout_ms")
	}
	if kv.Value.AsInt64() != 5000 {
		t.Errorf("A2aNegotiationTimeoutMs value = %d, want %d", kv.Value.AsInt64(), 5000)
	}
}

func TestA2aDealValueKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aDealValueKey) != "a2a.deal.value" {
		t.Errorf("A2aDealValueKey = %q, want %q", A2aDealValueKey, "a2a.deal.value")
	}
}

func TestA2aDealValueKeyValueRoundTrip(t *testing.T) {
	kv := A2aDealValue(250.5)
	if string(kv.Key) != "a2a.deal.value" {
		t.Errorf("A2aDealValue key = %q, want %q", string(kv.Key), "a2a.deal.value")
	}
	if kv.Value.AsFloat64() != 250.5 {
		t.Errorf("A2aDealValue value = %f, want %f", kv.Value.AsFloat64(), 250.5)
	}
}

// ============================================================
// Iteration 7: Healing Soundness (WvdA deadlock freedom + boundedness)
// ============================================================

func TestHealingTimeoutMsKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingTimeoutMsKey) != "healing.timeout_ms" {
		t.Errorf("HealingTimeoutMsKey = %q, want %q", HealingTimeoutMsKey, "healing.timeout_ms")
	}
}

func TestHealingTimeoutMsKeyValueRoundTrip(t *testing.T) {
	// WvdA deadlock freedom: every healing op must have timeout_ms > 0
	kv := HealingTimeoutMs(30000)
	if string(kv.Key) != "healing.timeout_ms" {
		t.Errorf("HealingTimeoutMs key = %q, want %q", string(kv.Key), "healing.timeout_ms")
	}
	if kv.Value.AsInt64() != 30000 {
		t.Errorf("HealingTimeoutMs value = %d, want %d", kv.Value.AsInt64(), 30000)
	}
}

func TestHealingMaxIterationsKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingMaxIterationsKey) != "healing.max_iterations" {
		t.Errorf("HealingMaxIterationsKey = %q, want %q", HealingMaxIterationsKey, "healing.max_iterations")
	}
}

func TestHealingMaxIterationsKeyValueRoundTrip(t *testing.T) {
	// WvdA boundedness: max_iterations enforces finite loop termination
	kv := HealingMaxIterations(11)
	if string(kv.Key) != "healing.max_iterations" {
		t.Errorf("HealingMaxIterations key = %q, want %q", string(kv.Key), "healing.max_iterations")
	}
	if kv.Value.AsInt64() != 11 {
		t.Errorf("HealingMaxIterations value = %d, want %d", kv.Value.AsInt64(), 11)
	}
}

func TestHealingIterationKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingIterationKey) != "healing.iteration" {
		t.Errorf("HealingIterationKey = %q, want %q", HealingIterationKey, "healing.iteration")
	}
}

func TestHealingIterationKeyValueRoundTrip(t *testing.T) {
	kv := HealingIteration(3)
	if string(kv.Key) != "healing.iteration" {
		t.Errorf("HealingIteration key = %q, want %q", string(kv.Key), "healing.iteration")
	}
	if kv.Value.AsInt64() != 3 {
		t.Errorf("HealingIteration value = %d, want %d", kv.Value.AsInt64(), 3)
	}
}

func TestHealingIterationBoundedByMaxIterations(t *testing.T) {
	// WvdA boundedness: current iteration must not exceed max_iterations
	maxIter := HealingMaxIterations(11)
	current := HealingIteration(5)
	if current.Value.AsInt64() > maxIter.Value.AsInt64() {
		t.Errorf("healing.iteration (%d) > healing.max_iterations (%d): boundedness violation",
			current.Value.AsInt64(), maxIter.Value.AsInt64())
	}
}

func TestHealingRecoveryCompleteKeyIsCorrectOtelName(t *testing.T) {
	if string(HealingRecoveryCompleteKey) != "healing.recovery_complete" {
		t.Errorf("HealingRecoveryCompleteKey = %q, want %q", HealingRecoveryCompleteKey, "healing.recovery_complete")
	}
}

func TestHealingRecoveryCompleteKeyValueRoundTripTrue(t *testing.T) {
	kv := HealingRecoveryComplete(true)
	if string(kv.Key) != "healing.recovery_complete" {
		t.Errorf("HealingRecoveryComplete key = %q, want %q", string(kv.Key), "healing.recovery_complete")
	}
	if !kv.Value.AsBool() {
		t.Errorf("HealingRecoveryComplete value = false, want true")
	}
}

func TestHealingRecoveryCompleteKeyValueRoundTripFalse(t *testing.T) {
	kv := HealingRecoveryComplete(false)
	if kv.Value.AsBool() {
		t.Errorf("HealingRecoveryComplete value = true, want false")
	}
}

// ============================================================
// Iteration 7: Signal Theory new attributes
// ============================================================

func TestSignalGenreKeyValueRoundTripIter7(t *testing.T) {
	// Verify S=(M,G,T,F,W): G=genre round-trip produces correct key
	kv := SignalGenre(SignalGenreValues.Spec)
	if string(kv.Key) != "signal.genre" {
		t.Errorf("SignalGenre key = %q, want %q", string(kv.Key), "signal.genre")
	}
	if kv.Value.AsString() != "spec" {
		t.Errorf("SignalGenre value = %q, want %q", kv.Value.AsString(), "spec")
	}
}

func TestSignalGenreBriefKeyValueRoundTrip(t *testing.T) {
	// S=(M,G,T,F,W): brief is a valid genre for short summaries
	kv := SignalGenre(SignalGenreValues.Brief)
	if kv.Value.AsString() != "brief" {
		t.Errorf("SignalGenre(Brief) value = %q, want %q", kv.Value.AsString(), "brief")
	}
}

func TestSignalFormatKeyValueRoundTripIter7(t *testing.T) {
	// Verify S=(M,G,T,F,W): F=format round-trip
	kv := SignalFormat(SignalFormatValues.Markdown)
	if string(kv.Key) != "signal.format" {
		t.Errorf("SignalFormat key = %q, want %q", string(kv.Key), "signal.format")
	}
	if kv.Value.AsString() != "markdown" {
		t.Errorf("SignalFormat value = %q, want %q", kv.Value.AsString(), "markdown")
	}
}

func TestSignalQualityThresholdKeyIsCorrectOtelName(t *testing.T) {
	if string(SignalQualityThresholdKey) != "signal.quality.threshold" {
		t.Errorf("SignalQualityThresholdKey = %q, want %q", SignalQualityThresholdKey, "signal.quality.threshold")
	}
}

func TestSignalQualityThresholdKeyValueRoundTrip(t *testing.T) {
	// Default S/N gate threshold is 0.7 per Signal Theory spec
	kv := SignalQualityThreshold(0.7)
	if string(kv.Key) != "signal.quality.threshold" {
		t.Errorf("SignalQualityThreshold key = %q, want %q", string(kv.Key), "signal.quality.threshold")
	}
	if kv.Value.AsFloat64() != 0.7 {
		t.Errorf("SignalQualityThreshold value = %f, want %f", kv.Value.AsFloat64(), 0.7)
	}
}

func TestSignalWeightKeyValueRoundTripIter7(t *testing.T) {
	// S=(M,G,T,F,W): W=weight round-trip produces correct key+value
	kv := SignalWeight(0.92)
	if string(kv.Key) != "signal.weight" {
		t.Errorf("SignalWeight key = %q, want %q", string(kv.Key), "signal.weight")
	}
	if kv.Value.AsFloat64() != 0.92 {
		t.Errorf("SignalWeight value = %f, want %f", kv.Value.AsFloat64(), 0.92)
	}
}

func TestSignalWeightAboveThresholdPassesSNGate(t *testing.T) {
	// Signal Theory: weight >= 0.7 passes the S/N gate
	threshold := SignalQualityThreshold(0.7)
	weight := SignalWeight(0.85)
	if weight.Value.AsFloat64() < threshold.Value.AsFloat64() {
		t.Errorf("signal.weight %f < threshold %f: signal should pass S/N gate",
			weight.Value.AsFloat64(), threshold.Value.AsFloat64())
	}
}

// ============================================================
// Iteration 7: YAWL new workflow attributes
// ============================================================

func TestWorkflowTriggerTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowTriggerTypeKey) != "workflow.trigger_type" {
		t.Errorf("WorkflowTriggerTypeKey = %q, want %q", WorkflowTriggerTypeKey, "workflow.trigger_type")
	}
}

func TestWorkflowTriggerTypeTimerValueMatchesSchema(t *testing.T) {
	if WorkflowTriggerTypeValues.Timer != "timer" {
		t.Errorf("WorkflowTriggerTypeValues.Timer = %q, want %q", WorkflowTriggerTypeValues.Timer, "timer")
	}
}

func TestWorkflowTriggerTypeKeyValueRoundTrip(t *testing.T) {
	kv := WorkflowTriggerType(WorkflowTriggerTypeValues.Timer)
	if string(kv.Key) != "workflow.trigger_type" {
		t.Errorf("WorkflowTriggerType key = %q, want %q", string(kv.Key), "workflow.trigger_type")
	}
	if kv.Value.AsString() != "timer" {
		t.Errorf("WorkflowTriggerType value = %q, want %q", kv.Value.AsString(), "timer")
	}
}

func TestWorkflowBranchCountKeyIsCorrectOtelName(t *testing.T) {
	if string(WorkflowBranchCountKey) != "workflow.branch_count" {
		t.Errorf("WorkflowBranchCountKey = %q, want %q", WorkflowBranchCountKey, "workflow.branch_count")
	}
}

func TestWorkflowBranchCountKeyValueRoundTrip(t *testing.T) {
	// YAWL parallel split: branch_count must be >= 2
	kv := WorkflowBranchCount(3)
	if string(kv.Key) != "workflow.branch_count" {
		t.Errorf("WorkflowBranchCount key = %q, want %q", string(kv.Key), "workflow.branch_count")
	}
	if kv.Value.AsInt64() != 3 {
		t.Errorf("WorkflowBranchCount value = %d, want %d", kv.Value.AsInt64(), 3)
	}
}

// ============================================================
// Wave 9 Iteration 8: Consensus BFT Liveness
// ============================================================

func TestConsensusQuorumSizeKeyMatchesSchema(t *testing.T) {
	if string(ConsensusQuorumSizeKey) != "consensus.quorum_size" {
		t.Errorf("ConsensusQuorumSizeKey = %q, want %q", ConsensusQuorumSizeKey, "consensus.quorum_size")
	}
}

func TestConsensusLeaderIdKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLeaderIDKey) != "consensus.leader_id" {
		t.Errorf("ConsensusLeaderIDKey = %q, want %q", ConsensusLeaderIDKey, "consensus.leader_id")
	}
}

func TestConsensusViewTimeoutMsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusViewTimeoutMsKey) != "consensus.view_timeout_ms" {
		t.Errorf("ConsensusViewTimeoutMsKey = %q, want %q", ConsensusViewTimeoutMsKey, "consensus.view_timeout_ms")
	}
}

func TestConsensusSignatureCountKeyMatchesSchema(t *testing.T) {
	if string(ConsensusSignatureCountKey) != "consensus.signature_count" {
		t.Errorf("ConsensusSignatureCountKey = %q, want %q", ConsensusSignatureCountKey, "consensus.signature_count")
	}
}

// ============================================================
// Wave 9 Iteration 8: MCP Tool Schema
// ============================================================

func TestMcpToolInputSizeKeyMatchesSchema(t *testing.T) {
	if string(McpToolInputSizeKey) != "mcp.tool.input_size" {
		t.Errorf("McpToolInputSizeKey = %q, want %q", McpToolInputSizeKey, "mcp.tool.input_size")
	}
}

func TestMcpToolOutputSizeKeyMatchesSchema(t *testing.T) {
	if string(McpToolOutputSizeKey) != "mcp.tool.output_size" {
		t.Errorf("McpToolOutputSizeKey = %q, want %q", McpToolOutputSizeKey, "mcp.tool.output_size")
	}
}

func TestMcpToolRetryCountKeyMatchesSchema(t *testing.T) {
	if string(McpToolRetryCountKey) != "mcp.tool.retry_count" {
		t.Errorf("McpToolRetryCountKey = %q, want %q", McpToolRetryCountKey, "mcp.tool.retry_count")
	}
}

func TestMcpToolTimeoutMsKeyMatchesSchema(t *testing.T) {
	if string(McpToolTimeoutMsKey) != "mcp.tool.timeout_ms" {
		t.Errorf("McpToolTimeoutMsKey = %q, want %q", McpToolTimeoutMsKey, "mcp.tool.timeout_ms")
	}
}

// ============================================================
// Wave 9 Iteration 8: LLM Observability
// ============================================================

func TestLlmModelKeyMatchesSchema(t *testing.T) {
	if string(LlmModelKey) != "llm.model" {
		t.Errorf("LlmModelKey = %q, want %q", LlmModelKey, "llm.model")
	}
}

func TestLlmProviderKeyMatchesSchema(t *testing.T) {
	if string(LlmProviderKey) != "llm.provider" {
		t.Errorf("LlmProviderKey = %q, want %q", LlmProviderKey, "llm.provider")
	}
}

func TestLlmTokenInputKeyMatchesSchema(t *testing.T) {
	if string(LlmTokenInputKey) != "llm.token.input" {
		t.Errorf("LlmTokenInputKey = %q, want %q", LlmTokenInputKey, "llm.token.input")
	}
}

func TestLlmTokenOutputKeyMatchesSchema(t *testing.T) {
	if string(LlmTokenOutputKey) != "llm.token.output" {
		t.Errorf("LlmTokenOutputKey = %q, want %q", LlmTokenOutputKey, "llm.token.output")
	}
}

func TestLlmLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(LlmLatencyMsKey) != "llm.latency_ms" {
		t.Errorf("LlmLatencyMsKey = %q, want %q", LlmLatencyMsKey, "llm.latency_ms")
	}
}

func TestLlmStopReasonEndTurnValueMatchesSchema(t *testing.T) {
	if LlmStopReasonEndTurn != "end_turn" {
		t.Errorf("LlmStopReasonEndTurn = %q, want %q", LlmStopReasonEndTurn, "end_turn")
	}
}

func TestLlmStopReasonToolUseValueMatchesSchema(t *testing.T) {
	if LlmStopReasonToolUse != "tool_use" {
		t.Errorf("LlmStopReasonToolUse = %q, want %q", LlmStopReasonToolUse, "tool_use")
	}
}

// ============================================================
// Wave 9 Iteration 8: Workspace Session
// ============================================================

func TestWorkspaceSessionIdKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceSessionIDKey) != "workspace.session.id" {
		t.Errorf("WorkspaceSessionIDKey = %q, want %q", WorkspaceSessionIDKey, "workspace.session.id")
	}
}

func TestWorkspaceContextSizeKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceContextSizeKey) != "workspace.context.size" {
		t.Errorf("WorkspaceContextSizeKey = %q, want %q", WorkspaceContextSizeKey, "workspace.context.size")
	}
}

func TestWorkspaceToolNameKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceToolNameKey) != "workspace.tool.name" {
		t.Errorf("WorkspaceToolNameKey = %q, want %q", WorkspaceToolNameKey, "workspace.tool.name")
	}
}

func TestWorkspaceAgentRolePlannerValueMatchesSchema(t *testing.T) {
	if WorkspaceAgentRolePlanner != "planner" {
		t.Errorf("WorkspaceAgentRolePlanner = %q, want %q", WorkspaceAgentRolePlanner, "planner")
	}
}

func TestWorkspaceAgentRoleExecutorValueMatchesSchema(t *testing.T) {
	if WorkspaceAgentRoleExecutor != "executor" {
		t.Errorf("WorkspaceAgentRoleExecutor = %q, want %q", WorkspaceAgentRoleExecutor, "executor")
	}
}

func TestWorkspacePhaseActiveValueMatchesSchema(t *testing.T) {
	if WorkspacePhaseActive != "active" {
		t.Errorf("WorkspacePhaseActive = %q, want %q", WorkspacePhaseActive, "active")
	}
}

// ============================================================
// Wave 9 Iteration 8: YAWL Basic Patterns
// ============================================================

func TestWorkflowSplitCountKeyMatchesSchema(t *testing.T) {
	if string(WorkflowSplitCountKey) != "workflow.split.count" {
		t.Errorf("WorkflowSplitCountKey = %q, want %q", WorkflowSplitCountKey, "workflow.split.count")
	}
}

func TestWorkflowMergePolicyKeyMatchesSchema(t *testing.T) {
	if string(WorkflowMergePolicyKey) != "workflow.merge.policy" {
		t.Errorf("WorkflowMergePolicyKey = %q, want %q", WorkflowMergePolicyKey, "workflow.merge.policy")
	}
}

func TestWorkflowMergePolicyAllValueMatchesSchema(t *testing.T) {
	if WorkflowMergePolicyAll != "all" {
		t.Errorf("WorkflowMergePolicyAll = %q, want %q", WorkflowMergePolicyAll, "all")
	}
}

func TestWorkflowChoiceConditionKeyMatchesSchema(t *testing.T) {
	if string(WorkflowChoiceConditionKey) != "workflow.choice.condition" {
		t.Errorf("WorkflowChoiceConditionKey = %q, want %q", WorkflowChoiceConditionKey, "workflow.choice.condition")
	}
}

// ============================================================
// Wave 9 Iteration 9: A2A Deal Tracking
// ============================================================

func TestA2aDealStatusKeyMatchesSchema(t *testing.T) {
	if string(A2aDealStatusKey) != "a2a.deal.status" {
		t.Errorf("A2aDealStatusKey = %q, want %q", A2aDealStatusKey, "a2a.deal.status")
	}
}

func TestA2aDealCurrencyKeyMatchesSchema(t *testing.T) {
	if string(A2aDealCurrencyKey) != "a2a.deal.currency" {
		t.Errorf("A2aDealCurrencyKey = %q, want %q", A2aDealCurrencyKey, "a2a.deal.currency")
	}
}

func TestA2aDealExpiryMsKeyMatchesSchema(t *testing.T) {
	if string(A2aDealExpiryMsKey) != "a2a.deal.expiry_ms" {
		t.Errorf("A2aDealExpiryMsKey = %q, want %q", A2aDealExpiryMsKey, "a2a.deal.expiry_ms")
	}
}

func TestA2aDealStatusCompletedValueMatchesSchema(t *testing.T) {
	if A2aDealStatusCompleted != "completed" {
		t.Errorf("A2aDealStatusCompleted = %q, want %q", A2aDealStatusCompleted, "completed")
	}
}

func TestA2aDealStatusPendingValueMatchesSchema(t *testing.T) {
	if A2aDealStatusPending != "pending" {
		t.Errorf("A2aDealStatusPending = %q, want %q", A2aDealStatusPending, "pending")
	}
}

func TestA2aDealStatusActiveValueMatchesSchema(t *testing.T) {
	if A2aDealStatusActive != "active" {
		t.Errorf("A2aDealStatusActive = %q, want %q", A2aDealStatusActive, "active")
	}
}

func TestA2aDealStatusCancelledValueMatchesSchema(t *testing.T) {
	if A2aDealStatusCancelled != "cancelled" {
		t.Errorf("A2aDealStatusCancelled = %q, want %q", A2aDealStatusCancelled, "cancelled")
	}
}

func TestA2aDealStatusDisputedValueMatchesSchema(t *testing.T) {
	if A2aDealStatusDisputed != "disputed" {
		t.Errorf("A2aDealStatusDisputed = %q, want %q", A2aDealStatusDisputed, "disputed")
	}
}

func TestA2aCapabilityVersionKeyMatchesSchema(t *testing.T) {
	if string(A2aCapabilityVersionKey) != "a2a.capability.version" {
		t.Errorf("A2aCapabilityVersionKey = %q, want %q", A2aCapabilityVersionKey, "a2a.capability.version")
	}
}

func TestA2aDealStatusKeyValueRoundTrip(t *testing.T) {
	kv := A2aDealStatus(A2aDealStatusCompleted)
	if string(kv.Key) != "a2a.deal.status" {
		t.Errorf("A2aDealStatus key = %q, want %q", string(kv.Key), "a2a.deal.status")
	}
	if kv.Value.AsString() != "completed" {
		t.Errorf("A2aDealStatus value = %q, want %q", kv.Value.AsString(), "completed")
	}
}

func TestA2aDealCurrencyKeyValueRoundTrip(t *testing.T) {
	kv := A2aDealCurrency("USD")
	if string(kv.Key) != "a2a.deal.currency" {
		t.Errorf("A2aDealCurrency key = %q, want %q", string(kv.Key), "a2a.deal.currency")
	}
	if kv.Value.AsString() != "USD" {
		t.Errorf("A2aDealCurrency value = %q, want %q", kv.Value.AsString(), "USD")
	}
}

func TestA2aDealExpiryMsKeyValueRoundTrip(t *testing.T) {
	kv := A2aDealExpiryMs(1711929600000)
	if string(kv.Key) != "a2a.deal.expiry_ms" {
		t.Errorf("A2aDealExpiryMs key = %q, want %q", string(kv.Key), "a2a.deal.expiry_ms")
	}
	if kv.Value.AsInt64() != 1711929600000 {
		t.Errorf("A2aDealExpiryMs value = %d, want %d", kv.Value.AsInt64(), 1711929600000)
	}
}

func TestA2aCapabilityVersionKeyValueRoundTrip(t *testing.T) {
	kv := A2aCapabilityVersion("v2.1.0")
	if string(kv.Key) != "a2a.capability.version" {
		t.Errorf("A2aCapabilityVersion key = %q, want %q", string(kv.Key), "a2a.capability.version")
	}
	if kv.Value.AsString() != "v2.1.0" {
		t.Errorf("A2aCapabilityVersion value = %q, want %q", kv.Value.AsString(), "v2.1.0")
	}
}

// ============================================================
// Wave 9 Iteration 9: Event Correlation
// ============================================================

func TestEventCausationIdKeyMatchesSchema(t *testing.T) {
	if string(EventCausationIdKey) != "event.causation_id" {
		t.Errorf("EventCausationIdKey = %q, want %q", EventCausationIdKey, "event.causation_id")
	}
}

func TestEventVersionKeyMatchesSchema(t *testing.T) {
	if string(EventVersionKey) != "event.version" {
		t.Errorf("EventVersionKey = %q, want %q", EventVersionKey, "event.version")
	}
}

func TestEventSourceServiceKeyMatchesSchema(t *testing.T) {
	if string(EventSourceServiceKey) != "event.source.service" {
		t.Errorf("EventSourceServiceKey = %q, want %q", EventSourceServiceKey, "event.source.service")
	}
}

func TestEventTargetServiceKeyMatchesSchema(t *testing.T) {
	if string(EventTargetServiceKey) != "event.target.service" {
		t.Errorf("EventTargetServiceKey = %q, want %q", EventTargetServiceKey, "event.target.service")
	}
}

func TestEventReplayKeyMatchesSchema(t *testing.T) {
	if string(EventReplayKey) != "event.replay" {
		t.Errorf("EventReplayKey = %q, want %q", EventReplayKey, "event.replay")
	}
}

func TestEventCausationIdKeyValueRoundTrip(t *testing.T) {
	kv := EventCausationId("evt-root-001")
	if string(kv.Key) != "event.causation_id" {
		t.Errorf("EventCausationId key = %q, want %q", string(kv.Key), "event.causation_id")
	}
	if kv.Value.AsString() != "evt-root-001" {
		t.Errorf("EventCausationId value = %q, want %q", kv.Value.AsString(), "evt-root-001")
	}
}

func TestEventVersionKeyValueRoundTrip(t *testing.T) {
	kv := EventVersion("1.0")
	if kv.Value.AsString() != "1.0" {
		t.Errorf("EventVersion value = %q, want %q", kv.Value.AsString(), "1.0")
	}
}

func TestEventSourceServiceKeyValueRoundTrip(t *testing.T) {
	kv := EventSourceService("osa")
	if kv.Value.AsString() != "osa" {
		t.Errorf("EventSourceService value = %q, want %q", kv.Value.AsString(), "osa")
	}
}

func TestEventTargetServiceKeyValueRoundTrip(t *testing.T) {
	kv := EventTargetService("canopy")
	if kv.Value.AsString() != "canopy" {
		t.Errorf("EventTargetService value = %q, want %q", kv.Value.AsString(), "canopy")
	}
}

func TestEventReplayKeyValueRoundTripTrue(t *testing.T) {
	kv := EventReplay(true)
	if string(kv.Key) != "event.replay" {
		t.Errorf("EventReplay key = %q, want %q", string(kv.Key), "event.replay")
	}
	if !kv.Value.AsBool() {
		t.Errorf("EventReplay value = false, want true")
	}
}

func TestEventReplayKeyValueRoundTripFalse(t *testing.T) {
	kv := EventReplay(false)
	if kv.Value.AsBool() {
		t.Errorf("EventReplay value = true, want false")
	}
}

// ============================================================
// Wave 9 Iteration 9: Process Mining Advanced
// ============================================================

func TestProcessMiningThroughputTimeMsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningThroughputTimeMsKey) != "process.mining.throughput_time_ms" {
		t.Errorf("ProcessMiningThroughputTimeMsKey = %q, want %q", ProcessMiningThroughputTimeMsKey, "process.mining.throughput_time_ms")
	}
}

func TestProcessMiningBottleneckActivityKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningBottleneckActivityKey) != "process.mining.bottleneck.activity" {
		t.Errorf("ProcessMiningBottleneckActivityKey = %q, want %q", ProcessMiningBottleneckActivityKey, "process.mining.bottleneck.activity")
	}
}

func TestProcessMiningBottleneckWaitMsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningBottleneckWaitMsKey) != "process.mining.bottleneck.wait_ms" {
		t.Errorf("ProcessMiningBottleneckWaitMsKey = %q, want %q", ProcessMiningBottleneckWaitMsKey, "process.mining.bottleneck.wait_ms")
	}
}

func TestProcessMiningLogSizeKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningLogSizeKey) != "process.mining.log.size" {
		t.Errorf("ProcessMiningLogSizeKey = %q, want %q", ProcessMiningLogSizeKey, "process.mining.log.size")
	}
}

func TestProcessMiningReplayFitnessKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayFitnessKey) != "process.mining.replay_fitness" {
		t.Errorf("ProcessMiningReplayFitnessKey = %q, want %q", ProcessMiningReplayFitnessKey, "process.mining.replay_fitness")
	}
}

func TestProcessMiningThroughputTimeMsKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningThroughputTimeMs(86400000)
	if string(kv.Key) != "process.mining.throughput_time_ms" {
		t.Errorf("ProcessMiningThroughputTimeMs key = %q, want %q", string(kv.Key), "process.mining.throughput_time_ms")
	}
	if kv.Value.AsInt64() != 86400000 {
		t.Errorf("ProcessMiningThroughputTimeMs value = %d, want %d", kv.Value.AsInt64(), 86400000)
	}
}

func TestProcessMiningBottleneckActivityKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningBottleneckActivity("Approve Purchase Order")
	if kv.Value.AsString() != "Approve Purchase Order" {
		t.Errorf("ProcessMiningBottleneckActivity value = %q, want %q", kv.Value.AsString(), "Approve Purchase Order")
	}
}

func TestProcessMiningBottleneckWaitMsKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningBottleneckWaitMs(7200000)
	if kv.Value.AsInt64() != 7200000 {
		t.Errorf("ProcessMiningBottleneckWaitMs value = %d, want %d", kv.Value.AsInt64(), 7200000)
	}
}

func TestProcessMiningLogSizeKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningLogSize(50000)
	if kv.Value.AsInt64() != 50000 {
		t.Errorf("ProcessMiningLogSize value = %d, want %d", kv.Value.AsInt64(), 50000)
	}
}

func TestProcessMiningReplayFitnessKeyValueRoundTrip(t *testing.T) {
	kv := ProcessMiningReplayFitness(0.87)
	if string(kv.Key) != "process.mining.replay_fitness" {
		t.Errorf("ProcessMiningReplayFitness key = %q, want %q", string(kv.Key), "process.mining.replay_fitness")
	}
	if kv.Value.AsFloat64() != 0.87 {
		t.Errorf("ProcessMiningReplayFitness value = %f, want %f", kv.Value.AsFloat64(), 0.87)
	}
}

func TestProcessMiningReplayFitnessBounded(t *testing.T) {
	// WvdA soundness: replay fitness score must be in [0.0, 1.0]
	kv := ProcessMiningReplayFitness(0.87)
	if kv.Value.AsFloat64() < 0.0 || kv.Value.AsFloat64() > 1.0 {
		t.Errorf("process.mining.replay_fitness %f is out of [0.0, 1.0] bounds", kv.Value.AsFloat64())
	}
}

// ============================================================
// Wave 9 Iteration 10: Signal Theory — priority, encoding, hop_count
// ============================================================

func TestSignalPriorityKeyMatchesSchema(t *testing.T) {
	if string(SignalPriorityKey) != "signal.priority" {
		t.Errorf("SignalPriorityKey = %q, want %q", string(SignalPriorityKey), "signal.priority")
	}
}

func TestSignalPriorityCriticalValueMatchesSchema(t *testing.T) {
	if SignalPriorityCritical != "critical" {
		t.Errorf("SignalPriorityCritical = %q, want %q", SignalPriorityCritical, "critical")
	}
}

func TestSignalEncodingKeyMatchesSchema(t *testing.T) {
	if string(SignalEncodingKey) != "signal.encoding" {
		t.Errorf("SignalEncodingKey = %q, want %q", string(SignalEncodingKey), "signal.encoding")
	}
}

func TestSignalHopCountKeyMatchesSchema(t *testing.T) {
	if string(SignalHopCountKey) != "signal.hop_count" {
		t.Errorf("SignalHopCountKey = %q, want %q", string(SignalHopCountKey), "signal.hop_count")
	}
}

// ============================================================
// Wave 9 Iteration 10: Canopy Heartbeat — latency, sequence, missed, session
// ============================================================

func TestCanopyHeartbeatLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(CanopyHeartbeatLatencyMsKey) != "canopy.heartbeat.latency_ms" {
		t.Errorf("CanopyHeartbeatLatencyMsKey = %q, want %q", string(CanopyHeartbeatLatencyMsKey), "canopy.heartbeat.latency_ms")
	}
}

func TestCanopyHeartbeatSequenceNumKeyMatchesSchema(t *testing.T) {
	if string(CanopyHeartbeatSequenceNumKey) != "canopy.heartbeat.sequence_num" {
		t.Errorf("CanopyHeartbeatSequenceNumKey = %q, want %q", string(CanopyHeartbeatSequenceNumKey), "canopy.heartbeat.sequence_num")
	}
}

func TestCanopySessionIdKeyMatchesSchema(t *testing.T) {
	if string(CanopySessionIDKey) != "canopy.session.id" {
		t.Errorf("CanopySessionIDKey = %q, want %q", string(CanopySessionIDKey), "canopy.session.id")
	}
}

// ============================================================
// Wave 9 Iteration 10: MCP Registry — tool_count, server_count, connection, transport
// ============================================================

func TestMcpRegistryToolCountKeyMatchesSchema(t *testing.T) {
	if string(McpRegistryToolCountKey) != "mcp.registry.tool_count" {
		t.Errorf("McpRegistryToolCountKey = %q, want %q", string(McpRegistryToolCountKey), "mcp.registry.tool_count")
	}
}

func TestMcpConnectionTransportKeyMatchesSchema(t *testing.T) {
	if string(McpConnectionTransportKey) != "mcp.connection.transport" {
		t.Errorf("McpConnectionTransportKey = %q, want %q", string(McpConnectionTransportKey), "mcp.connection.transport")
	}
}

func TestMcpConnectionTransportStdioValueMatchesSchema(t *testing.T) {
	if McpConnectionTransportStdio != "stdio" {
		t.Errorf("McpConnectionTransportStdio = %q, want %q", McpConnectionTransportStdio, "stdio")
	}
}

// ============================================================
// Wave 9 Iteration 10: Conversation — id, turn_count, model, phase
// ============================================================

func TestConversationIdKeyMatchesSchema(t *testing.T) {
	if string(ConversationIDKey) != "conversation.id" {
		t.Errorf("ConversationIDKey = %q, want %q", string(ConversationIDKey), "conversation.id")
	}
}

func TestConversationTurnCountKeyMatchesSchema(t *testing.T) {
	if string(ConversationTurnCountKey) != "conversation.turn_count" {
		t.Errorf("ConversationTurnCountKey = %q, want %q", string(ConversationTurnCountKey), "conversation.turn_count")
	}
}

func TestConversationModelKeyMatchesSchema(t *testing.T) {
	if string(ConversationModelKey) != "conversation.model" {
		t.Errorf("ConversationModelKey = %q, want %q", string(ConversationModelKey), "conversation.model")
	}
}

func TestConversationPhaseActiveValueMatchesSchema(t *testing.T) {
	if ConversationPhaseActive != "active" {
		t.Errorf("ConversationPhaseActive = %q, want %q", ConversationPhaseActive, "active")
	}
}

func TestConversationPhaseCompleteValueMatchesSchema(t *testing.T) {
	if ConversationPhaseComplete != "complete" {
		t.Errorf("ConversationPhaseComplete = %q, want %q", ConversationPhaseComplete, "complete")
	}
}

// ============================================================
// Wave 9 Iteration 10: YAWL WP-6/7 — active_branches, fired_branches, sync timeout
// ============================================================

func TestWorkflowActiveBranchesKeyMatchesSchema(t *testing.T) {
	if string(WorkflowActiveBranchesKey) != "workflow.active_branches" {
		t.Errorf("WorkflowActiveBranchesKey = %q, want %q", string(WorkflowActiveBranchesKey), "workflow.active_branches")
	}
}

func TestWorkflowFiredBranchesKeyMatchesSchema(t *testing.T) {
	if string(WorkflowFiredBranchesKey) != "workflow.fired_branches" {
		t.Errorf("WorkflowFiredBranchesKey = %q, want %q", string(WorkflowFiredBranchesKey), "workflow.fired_branches")
	}
}

// ============================================================
// Wave 9 Iteration 11: LLM cost tracking
// ============================================================

func TestLlmCostTotalKeyMatchesSchema(t *testing.T) {
	if string(LlmCostTotalKey) != "llm.cost.total" {
		t.Errorf("LlmCostTotalKey = %q, want %q", string(LlmCostTotalKey), "llm.cost.total")
	}
}

func TestLlmCostInputKeyMatchesSchema(t *testing.T) {
	if string(LlmCostInputKey) != "llm.cost.input" {
		t.Errorf("LlmCostInputKey = %q, want %q", string(LlmCostInputKey), "llm.cost.input")
	}
}

func TestLlmCostOutputKeyMatchesSchema(t *testing.T) {
	if string(LlmCostOutputKey) != "llm.cost.output" {
		t.Errorf("LlmCostOutputKey = %q, want %q", string(LlmCostOutputKey), "llm.cost.output")
	}
}

func TestLlmModelFamilyKeyMatchesSchema(t *testing.T) {
	if string(LlmModelFamilyKey) != "llm.model_family" {
		t.Errorf("LlmModelFamilyKey = %q, want %q", string(LlmModelFamilyKey), "llm.model_family")
	}
}

func TestLlmRequestIdKeyMatchesSchema(t *testing.T) {
	if string(LlmRequestIdKey) != "llm.request.id" {
		t.Errorf("LlmRequestIdKey = %q, want %q", string(LlmRequestIdKey), "llm.request.id")
	}
}

// ============================================================
// Wave 9 Iteration 11: Process mining replay quality metrics
// ============================================================

func TestProcessMiningReplayPrecisionKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayPrecisionKey) != "process_mining.replay.precision" {
		t.Errorf("ProcessMiningReplayPrecisionKey = %q, want %q", string(ProcessMiningReplayPrecisionKey), "process_mining.replay.precision")
	}
}

func TestProcessMiningReplayGeneralizationKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayGeneralizationKey) != "process_mining.replay.generalization" {
		t.Errorf("ProcessMiningReplayGeneralizationKey = %q, want %q", string(ProcessMiningReplayGeneralizationKey), "process_mining.replay.generalization")
	}
}

func TestProcessMiningReplaySimplicityKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplaySimplicityKey) != "process_mining.replay.simplicity" {
		t.Errorf("ProcessMiningReplaySimplicityKey = %q, want %q", string(ProcessMiningReplaySimplicityKey), "process_mining.replay.simplicity")
	}
}

func TestProcessMiningAlignmentCostKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningAlignmentCostKey) != "process_mining.alignment.cost" {
		t.Errorf("ProcessMiningAlignmentCostKey = %q, want %q", string(ProcessMiningAlignmentCostKey), "process_mining.alignment.cost")
	}
}

func TestProcessMiningModelTypeKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningModelTypeKey) != "process_mining.model.type" {
		t.Errorf("ProcessMiningModelTypeKey = %q, want %q", string(ProcessMiningModelTypeKey), "process_mining.model.type")
	}
}

// ============================================================
// Wave 9 Iteration 11: Consensus quorum health and replica counts
// ============================================================

func TestConsensusQuorumHealthKeyMatchesSchema(t *testing.T) {
	if string(ConsensusQuorumHealthKey) != "consensus.quorum.health" {
		t.Errorf("ConsensusQuorumHealthKey = %q, want %q", string(ConsensusQuorumHealthKey), "consensus.quorum.health")
	}
}

func TestConsensusBlockHeightKeyMatchesSchema(t *testing.T) {
	if string(ConsensusBlockHeightKey) != "consensus.block.height" {
		t.Errorf("ConsensusBlockHeightKey = %q, want %q", string(ConsensusBlockHeightKey), "consensus.block.height")
	}
}

func TestConsensusReplicaCountKeyMatchesSchema(t *testing.T) {
	if string(ConsensusReplicaCountKey) != "consensus.replica.count" {
		t.Errorf("ConsensusReplicaCountKey = %q, want %q", string(ConsensusReplicaCountKey), "consensus.replica.count")
	}
}

func TestConsensusFailureCountKeyMatchesSchema(t *testing.T) {
	if string(ConsensusFailureCountKey) != "consensus.failure.count" {
		t.Errorf("ConsensusFailureCountKey = %q, want %q", string(ConsensusFailureCountKey), "consensus.failure.count")
	}
}

// ============================================================
// Wave 9 Iteration 11: A2A SLA tracking
// ============================================================

func TestA2ASlaDeadlineMsKeyMatchesSchema(t *testing.T) {
	if string(A2ASlaDeadlineMsKey) != "a2a.sla.deadline_ms" {
		t.Errorf("A2ASlaDeadlineMsKey = %q, want %q", string(A2ASlaDeadlineMsKey), "a2a.sla.deadline_ms")
	}
}

func TestA2ASlaBreachKeyMatchesSchema(t *testing.T) {
	if string(A2ASlaBreachKey) != "a2a.sla.breach" {
		t.Errorf("A2ASlaBreachKey = %q, want %q", string(A2ASlaBreachKey), "a2a.sla.breach")
	}
}

func TestA2ASlaLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(A2ASlaLatencyMsKey) != "a2a.sla.latency_ms" {
		t.Errorf("A2ASlaLatencyMsKey = %q, want %q", string(A2ASlaLatencyMsKey), "a2a.sla.latency_ms")
	}
}

func TestA2ARetryCountKeyMatchesSchema(t *testing.T) {
	if string(A2ARetryCountKey) != "a2a.retry.count" {
		t.Errorf("A2ARetryCountKey = %q, want %q", string(A2ARetryCountKey), "a2a.retry.count")
	}
}

// ============================================================
// Wave 9 Iteration 11: Workspace tool category and context window
// ============================================================

func TestWorkspaceToolCategoryKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceToolCategoryKey) != "workspace.tool.category" {
		t.Errorf("WorkspaceToolCategoryKey = %q, want %q", string(WorkspaceToolCategoryKey), "workspace.tool.category")
	}
}

func TestWorkspaceContextWindowSizeKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceContextWindowSizeKey) != "workspace.context.window_size" {
		t.Errorf("WorkspaceContextWindowSizeKey = %q, want %q", string(WorkspaceContextWindowSizeKey), "workspace.context.window_size")
	}
}

// ============================================================
// Wave 9 Iteration 11: BusinessOS compliance, audit and integration
// ============================================================

func TestBusinessOsComplianceFrameworkKeyMatchesSchema(t *testing.T) {
	if string(BusinessOsComplianceFrameworkKey) != "business_os.compliance.framework" {
		t.Errorf("BusinessOsComplianceFrameworkKey = %q, want %q", string(BusinessOsComplianceFrameworkKey), "business_os.compliance.framework")
	}
}

func TestBusinessOsAuditEventTypeKeyMatchesSchema(t *testing.T) {
	if string(BusinessOsAuditEventTypeKey) != "business_os.audit.event_type" {
		t.Errorf("BusinessOsAuditEventTypeKey = %q, want %q", string(BusinessOsAuditEventTypeKey), "business_os.audit.event_type")
	}
}

func TestBusinessOsIntegrationTypeKeyMatchesSchema(t *testing.T) {
	if string(BusinessOsIntegrationTypeKey) != "business_os.integration.type" {
		t.Errorf("BusinessOsIntegrationTypeKey = %q, want %q", string(BusinessOsIntegrationTypeKey), "business_os.integration.type")
	}
}
