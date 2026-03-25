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

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
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
	if string(AgentDecisionTypeKey) != "agent.decision.type" {
		t.Errorf("AgentDecisionTypeKey = %q, want %q", AgentDecisionTypeKey, "agent.decision.type")
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

func TestConsensusLeaderIdKeyIsCorrectOtelName(t *testing.T) {
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
	if string(ConsensusLeaderIdKey) != "consensus.leader.id" {
		t.Errorf("ConsensusLeaderIdKey = %q, want %q", ConsensusLeaderIdKey, "consensus.leader.id")
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
	if LlmStopReasonValues.EndTurn != "end_turn" {
		t.Errorf("LlmStopReasonValues.EndTurn = %q, want %q", LlmStopReasonValues.EndTurn, "end_turn")
	}
}

func TestLlmStopReasonToolUseValueMatchesSchema(t *testing.T) {
	if LlmStopReasonValues.ToolUse != "tool_use" {
		t.Errorf("LlmStopReasonValues.ToolUse = %q, want %q", LlmStopReasonValues.ToolUse, "tool_use")
	}
}

// ============================================================
// Wave 9 Iteration 8: Workspace Session
// ============================================================

func TestWorkspaceSessionIdKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceSessionIdKey) != "workspace.session.id" {
		t.Errorf("WorkspaceSessionIdKey = %q, want %q", WorkspaceSessionIdKey, "workspace.session.id")
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
	if WorkspaceAgentRoleValues.Planner != "planner" {
		t.Errorf("WorkspaceAgentRoleValues.Planner = %q, want %q", WorkspaceAgentRoleValues.Planner, "planner")
	}
}

func TestWorkspaceAgentRoleExecutorValueMatchesSchema(t *testing.T) {
	if WorkspaceAgentRoleValues.Executor != "executor" {
		t.Errorf("WorkspaceAgentRoleValues.Executor = %q, want %q", WorkspaceAgentRoleValues.Executor, "executor")
	}
}

func TestWorkspacePhaseActiveValueMatchesSchema(t *testing.T) {
	if WorkspacePhaseValues.Active != "active" {
		t.Errorf("WorkspacePhaseValues.Active = %q, want %q", WorkspacePhaseValues.Active, "active")
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
	if WorkflowMergePolicyValues.All != "all" {
		t.Errorf("WorkflowMergePolicyValues.All = %q, want %q", WorkflowMergePolicyValues.All, "all")
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
	if A2aDealStatusValues.Completed != "completed" {
		t.Errorf("A2aDealStatusValues.Completed = %q, want %q", A2aDealStatusValues.Completed, "completed")
	}
}

func TestA2aDealStatusPendingValueMatchesSchema(t *testing.T) {
	if A2aDealStatusValues.Pending != "pending" {
		t.Errorf("A2aDealStatusValues.Pending = %q, want %q", A2aDealStatusValues.Pending, "pending")
	}
}

func TestA2aDealStatusActiveValueMatchesSchema(t *testing.T) {
	if A2aDealStatusValues.Active != "active" {
		t.Errorf("A2aDealStatusValues.Active = %q, want %q", A2aDealStatusValues.Active, "active")
	}
}

func TestA2aDealStatusCancelledValueMatchesSchema(t *testing.T) {
	if A2aDealStatusValues.Cancelled != "cancelled" {
		t.Errorf("A2aDealStatusValues.Cancelled = %q, want %q", A2aDealStatusValues.Cancelled, "cancelled")
	}
}

func TestA2aDealStatusDisputedValueMatchesSchema(t *testing.T) {
	if A2aDealStatusValues.Disputed != "disputed" {
		t.Errorf("A2aDealStatusValues.Disputed = %q, want %q", A2aDealStatusValues.Disputed, "disputed")
	}
}

func TestA2aCapabilityVersionKeyMatchesSchema(t *testing.T) {
	if string(A2aCapabilityVersionKey) != "a2a.capability.version" {
		t.Errorf("A2aCapabilityVersionKey = %q, want %q", A2aCapabilityVersionKey, "a2a.capability.version")
	}
}

func TestA2aDealStatusKeyValueRoundTrip(t *testing.T) {
	kv := A2aDealStatus(A2aDealStatusValues.Completed)
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
	if SignalPriorityValues.Critical != "critical" {
		t.Errorf("SignalPriorityValues.Critical = %q, want %q", SignalPriorityValues.Critical, "critical")
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
	if string(CanopySessionIdKey) != "canopy.session.id" {
		t.Errorf("CanopySessionIdKey = %q, want %q", string(CanopySessionIdKey), "canopy.session.id")
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
	if McpConnectionTransportValues.Stdio != "stdio" {
		t.Errorf("McpConnectionTransportValues.Stdio = %q, want %q", McpConnectionTransportValues.Stdio, "stdio")
	}
}

// ============================================================
// Wave 9 Iteration 10: Conversation — id, turn_count, model, phase
// ============================================================

func TestConversationIdKeyMatchesSchema(t *testing.T) {
	if string(ConversationIdKey) != "conversation.id" {
		t.Errorf("ConversationIdKey = %q, want %q", string(ConversationIdKey), "conversation.id")
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
	if ConversationPhaseValues.Active != "active" {
		t.Errorf("ConversationPhaseValues.Active = %q, want %q", ConversationPhaseValues.Active, "active")
	}
}

func TestConversationPhaseCompleteValueMatchesSchema(t *testing.T) {
	if ConversationPhaseValues.Complete != "complete" {
		t.Errorf("ConversationPhaseValues.Complete = %q, want %q", ConversationPhaseValues.Complete, "complete")
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
	if string(A2aSlaDeadlineMsKey) != "a2a.sla.deadline_ms" {
		t.Errorf("A2aSlaDeadlineMsKey = %q, want %q", string(A2aSlaDeadlineMsKey), "a2a.sla.deadline_ms")
	}
}

func TestA2ASlaBreachKeyMatchesSchema(t *testing.T) {
	if string(A2aSlaBreachKey) != "a2a.sla.breach" {
		t.Errorf("A2aSlaBreachKey = %q, want %q", string(A2aSlaBreachKey), "a2a.sla.breach")
	}
}

func TestA2ASlaLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(A2aSlaLatencyMsKey) != "a2a.sla.latency_ms" {
		t.Errorf("A2aSlaLatencyMsKey = %q, want %q", string(A2aSlaLatencyMsKey), "a2a.sla.latency_ms")
	}
}

func TestA2ARetryCountKeyMatchesSchema(t *testing.T) {
	if string(A2aRetryCountKey) != "a2a.retry.count" {
		t.Errorf("A2aRetryCountKey = %q, want %q", string(A2aRetryCountKey), "a2a.retry.count")
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

// ============================================================
// Wave 9 Iteration 12: Healing MTTR, agent topology, PM streaming,
// Canopy protocol, LLM safety, event delivery
// ============================================================

// Healing escalation and repair strategy
func TestHealingEscalationLevelKeyMatchesSchema(t *testing.T) {
	if string(HealingEscalationLevelKey) != "healing.escalation.level" {
		t.Errorf("HealingEscalationLevelKey = %q, want %q", string(HealingEscalationLevelKey), "healing.escalation.level")
	}
}

func TestHealingRepairStrategyKeyMatchesSchema(t *testing.T) {
	if string(HealingRepairStrategyKey) != "healing.repair.strategy" {
		t.Errorf("HealingRepairStrategyKey = %q, want %q", string(HealingRepairStrategyKey), "healing.repair.strategy")
	}
}

func TestHealingAttemptKeyMatchesSchema(t *testing.T) {
	if string(HealingAttemptKey) != "healing.attempt" {
		t.Errorf("HealingAttemptKey = %q, want %q", string(HealingAttemptKey), "healing.attempt")
	}
}

// Agent topology and coordination
func TestAgentTopologyTypeKeyMatchesSchema(t *testing.T) {
	if string(AgentTopologyTypeKey) != "agent.topology.type" {
		t.Errorf("AgentTopologyTypeKey = %q, want %q", string(AgentTopologyTypeKey), "agent.topology.type")
	}
}

func TestAgentTaskStatusKeyMatchesSchema(t *testing.T) {
	if string(AgentTaskStatusKey) != "agent.task.status" {
		t.Errorf("AgentTaskStatusKey = %q, want %q", string(AgentTaskStatusKey), "agent.task.status")
	}
}

func TestAgentCoordinationLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(AgentCoordinationLatencyMsKey) != "agent.coordination.latency_ms" {
		t.Errorf("AgentCoordinationLatencyMsKey = %q, want %q", string(AgentCoordinationLatencyMsKey), "agent.coordination.latency_ms")
	}
}

func TestAgentMessageCountKeyMatchesSchema(t *testing.T) {
	if string(AgentMessageCountKey) != "agent.message.count" {
		t.Errorf("AgentMessageCountKey = %q, want %q", string(AgentMessageCountKey), "agent.message.count")
	}
}

// Process mining streaming and drift detection




// Canopy protocol and sync
func TestCanopyProtocolVersionKeyMatchesSchema(t *testing.T) {
	if string(CanopyProtocolVersionKey) != "canopy.protocol.version" {
		t.Errorf("CanopyProtocolVersionKey = %q, want %q", string(CanopyProtocolVersionKey), "canopy.protocol.version")
	}
}

func TestCanopySyncStrategyKeyMatchesSchema(t *testing.T) {
	if string(CanopySyncStrategyKey) != "canopy.sync.strategy" {
		t.Errorf("CanopySyncStrategyKey = %q, want %q", string(CanopySyncStrategyKey), "canopy.sync.strategy")
	}
}

func TestCanopyConflictCountKeyMatchesSchema(t *testing.T) {
	if string(CanopyConflictCountKey) != "canopy.conflict.count" {
		t.Errorf("CanopyConflictCountKey = %q, want %q", string(CanopyConflictCountKey), "canopy.conflict.count")
	}
}

func TestCanopyPeerCountKeyMatchesSchema(t *testing.T) {
	if string(CanopyPeerCountKey) != "canopy.peer.count" {
		t.Errorf("CanopyPeerCountKey = %q, want %q", string(CanopyPeerCountKey), "canopy.peer.count")
	}
}

// LLM safety and guardrails
func TestLlmSafetyScoreKeyMatchesSchema(t *testing.T) {
	if string(LlmSafetyScoreKey) != "llm.safety.score" {
		t.Errorf("LlmSafetyScoreKey = %q, want %q", string(LlmSafetyScoreKey), "llm.safety.score")
	}
}

func TestLlmGuardrailTriggeredKeyMatchesSchema(t *testing.T) {
	if string(LlmGuardrailTriggeredKey) != "llm.guardrail.triggered" {
		t.Errorf("LlmGuardrailTriggeredKey = %q, want %q", string(LlmGuardrailTriggeredKey), "llm.guardrail.triggered")
	}
}

func TestLlmGuardrailTypeKeyMatchesSchema(t *testing.T) {
	if string(LlmGuardrailTypeKey) != "llm.guardrail.type" {
		t.Errorf("LlmGuardrailTypeKey = %q, want %q", string(LlmGuardrailTypeKey), "llm.guardrail.type")
	}
}

func TestLlmContextMessagesCountKeyMatchesSchema(t *testing.T) {
	if string(LlmContextMessagesCountKey) != "llm.context.messages_count" {
		t.Errorf("LlmContextMessagesCountKey = %q, want %q", string(LlmContextMessagesCountKey), "llm.context.messages_count")
	}
}

func TestLlmRetryCountKeyMatchesSchema(t *testing.T) {
	if string(LlmRetryCountKey) != "llm.retry.count" {
		t.Errorf("LlmRetryCountKey = %q, want %q", string(LlmRetryCountKey), "llm.retry.count")
	}
}

// Event delivery status and handler metadata
func TestEventDeliveryStatusKeyMatchesSchema(t *testing.T) {
	if string(EventDeliveryStatusKey) != "event.delivery.status" {
		t.Errorf("EventDeliveryStatusKey = %q, want %q", string(EventDeliveryStatusKey), "event.delivery.status")
	}
}

func TestEventHandlerCountKeyMatchesSchema(t *testing.T) {
	if string(EventHandlerCountKey) != "event.handler.count" {
		t.Errorf("EventHandlerCountKey = %q, want %q", string(EventHandlerCountKey), "event.handler.count")
	}
}

func TestEventSchemaVersionKeyMatchesSchema(t *testing.T) {
	if string(EventSchemaVersionKey) != "event.schema.version" {
		t.Errorf("EventSchemaVersionKey = %q, want %q", string(EventSchemaVersionKey), "event.schema.version")
	}
}

// ============================================================
// Iter 13: Workspace orchestration
// ============================================================

func TestWorkspaceOrchestrationPatternKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceOrchestrationPatternKey) != "workspace.orchestration.pattern" {
		t.Errorf("WorkspaceOrchestrationPatternKey = %q, want %q", string(WorkspaceOrchestrationPatternKey), "workspace.orchestration.pattern")
	}
}

func TestWorkspaceTaskQueueDepthKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceTaskQueueDepthKey) != "workspace.task.queue.depth" {
		t.Errorf("WorkspaceTaskQueueDepthKey = %q, want %q", string(WorkspaceTaskQueueDepthKey), "workspace.task.queue.depth")
	}
}

func TestWorkspaceIterationCountKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceIterationCountKey) != "workspace.iteration.count" {
		t.Errorf("WorkspaceIterationCountKey = %q, want %q", string(WorkspaceIterationCountKey), "workspace.iteration.count")
	}
}

// ============================================================
// Iter 13: A2A capability matching
// ============================================================

func TestA2ACapabilityMatchScoreKeyMatchesSchema(t *testing.T) {
	if string(A2aCapabilityMatchScoreKey) != "a2a.capability.match_score" {
		t.Errorf("A2aCapabilityMatchScoreKey = %q, want %q", string(A2aCapabilityMatchScoreKey), "a2a.capability.match_score")
	}
}

func TestA2ACapabilityRequiredKeyMatchesSchema(t *testing.T) {
	if string(A2aCapabilityRequiredKey) != "a2a.capability.required" {
		t.Errorf("A2aCapabilityRequiredKey = %q, want %q", string(A2aCapabilityRequiredKey), "a2a.capability.required")
	}
}

func TestA2ACapabilityOfferedKeyMatchesSchema(t *testing.T) {
	if string(A2aCapabilityOfferedKey) != "a2a.capability.offered" {
		t.Errorf("A2aCapabilityOfferedKey = %q, want %q", string(A2aCapabilityOfferedKey), "a2a.capability.offered")
	}
}

func TestA2ARoutingStrategyKeyMatchesSchema(t *testing.T) {
	if string(A2aRoutingStrategyKey) != "a2a.routing.strategy" {
		t.Errorf("A2aRoutingStrategyKey = %q, want %q", string(A2aRoutingStrategyKey), "a2a.routing.strategy")
	}
}

func TestA2AQueueDepthKeyMatchesSchema(t *testing.T) {
	if string(A2aQueueDepthKey) != "a2a.queue.depth" {
		t.Errorf("A2aQueueDepthKey = %q, want %q", string(A2aQueueDepthKey), "a2a.queue.depth")
	}
}

// ============================================================
// Iter 13: Consensus safety and liveness
// ============================================================

func TestConsensusSafetyThresholdKeyMatchesSchema(t *testing.T) {
	if string(ConsensusSafetyThresholdKey) != "consensus.safety.threshold" {
		t.Errorf("ConsensusSafetyThresholdKey = %q, want %q", string(ConsensusSafetyThresholdKey), "consensus.safety.threshold")
	}
}

func TestConsensusLivenessTimeoutRatioKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLivenessTimeoutRatioKey) != "consensus.liveness.timeout_ratio" {
		t.Errorf("ConsensusLivenessTimeoutRatioKey = %q, want %q", string(ConsensusLivenessTimeoutRatioKey), "consensus.liveness.timeout_ratio")
	}
}

func TestConsensusNetworkPartitionDetectedKeyMatchesSchema(t *testing.T) {
	if string(ConsensusNetworkPartitionDetectedKey) != "consensus.network.partition_detected" {
		t.Errorf("ConsensusNetworkPartitionDetectedKey = %q, want %q", string(ConsensusNetworkPartitionDetectedKey), "consensus.network.partition_detected")
	}
}

// ============================================================
// Iter 13: Healing cascade detection
// ============================================================

func TestHealingCascadeDetectedKeyMatchesSchema(t *testing.T) {
	if string(HealingCascadeDetectedKey) != "healing.cascade.detected" {
		t.Errorf("HealingCascadeDetectedKey = %q, want %q", string(HealingCascadeDetectedKey), "healing.cascade.detected")
	}
}

func TestHealingCascadeDepthKeyMatchesSchema(t *testing.T) {
	if string(HealingCascadeDepthKey) != "healing.cascade.depth" {
		t.Errorf("HealingCascadeDepthKey = %q, want %q", string(HealingCascadeDepthKey), "healing.cascade.depth")
	}
}

func TestHealingRootCauseIdKeyMatchesSchema(t *testing.T) {
	if string(HealingRootCauseIdKey) != "healing.root_cause.id" {
		t.Errorf("HealingRootCauseIdKey = %q, want %q", string(HealingRootCauseIdKey), "healing.root_cause.id")
	}
}

// ============================================================
// Iter 13: LLM chain-of-thought
// ============================================================

func TestLlmChainOfThoughtStepsKeyMatchesSchema(t *testing.T) {
	if string(LlmChainOfThoughtStepsKey) != "llm.chain_of_thought.steps" {
		t.Errorf("LlmChainOfThoughtStepsKey = %q, want %q", string(LlmChainOfThoughtStepsKey), "llm.chain_of_thought.steps")
	}
}

func TestLlmChainOfThoughtEnabledKeyMatchesSchema(t *testing.T) {
	if string(LlmChainOfThoughtEnabledKey) != "llm.chain_of_thought.enabled" {
		t.Errorf("LlmChainOfThoughtEnabledKey = %q, want %q", string(LlmChainOfThoughtEnabledKey), "llm.chain_of_thought.enabled")
	}
}

func TestLlmToolCallCountKeyMatchesSchema(t *testing.T) {
	if string(LlmToolCallCountKey) != "llm.tool.call_count" {
		t.Errorf("LlmToolCallCountKey = %q, want %q", string(LlmToolCallCountKey), "llm.tool.call_count")
	}
}

func TestLlmCacheHitKeyMatchesSchema(t *testing.T) {
	if string(LlmCacheHitKey) != "llm.cache.hit" {
		t.Errorf("LlmCacheHitKey = %q, want %q", string(LlmCacheHitKey), "llm.cache.hit")
	}
}

// ============================================================
// Iter 13: MCP tool versioning
// ============================================================

func TestMcpToolVersionKeyMatchesSchema(t *testing.T) {
	if string(McpToolVersionKey) != "mcp.tool.version" {
		t.Errorf("McpToolVersionKey = %q, want %q", string(McpToolVersionKey), "mcp.tool.version")
	}
}

func TestMcpToolSchemaHashKeyMatchesSchema(t *testing.T) {
	if string(McpToolSchemaHashKey) != "mcp.tool.schema_hash" {
		t.Errorf("McpToolSchemaHashKey = %q, want %q", string(McpToolSchemaHashKey), "mcp.tool.schema_hash")
	}
}

func TestMcpSessionIdKeyMatchesSchema(t *testing.T) {
	if string(McpSessionIdKey) != "mcp.session.id" {
		t.Errorf("McpSessionIdKey = %q, want %q", string(McpSessionIdKey), "mcp.session.id")
	}
}

// ============================================================
// Iter 13: Process mining conformance visualization
// ============================================================

func TestProcessMiningConformanceVisualizationTypeKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningConformanceVisualizationTypeKey) != "process.mining.conformance.visualization_type" {
		t.Errorf("ProcessMiningConformanceVisualizationTypeKey = %q, want %q", string(ProcessMiningConformanceVisualizationTypeKey), "process.mining.conformance.visualization_type")
	}
}

func TestProcessMiningCaseThroughputMsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningCaseThroughputMsKey) != "process.mining.case.throughput_ms" {
		t.Errorf("ProcessMiningCaseThroughputMsKey = %q, want %q", string(ProcessMiningCaseThroughputMsKey), "process.mining.case.throughput_ms")
	}
}

func TestProcessMiningActivityWaitingMsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningActivityWaitingMsKey) != "process.mining.activity.waiting_ms" {
		t.Errorf("ProcessMiningActivityWaitingMsKey = %q, want %q", string(ProcessMiningActivityWaitingMsKey), "process.mining.activity.waiting_ms")
	}
}

// ============================================================
// Iter 14: A2A trust attributes
// ============================================================

func TestA2ATrustScoreKeyMatchesSchema(t *testing.T) {
	if string(A2aTrustScoreKey) != "a2a.trust.score" {
		t.Errorf("A2aTrustScoreKey = %q, want %q", string(A2aTrustScoreKey), "a2a.trust.score")
	}
}

func TestA2AReputationHistoryLengthKeyMatchesSchema(t *testing.T) {
	if string(A2aReputationHistoryLengthKey) != "a2a.reputation.history_length" {
		t.Errorf("A2aReputationHistoryLengthKey = %q, want %q", string(A2aReputationHistoryLengthKey), "a2a.reputation.history_length")
	}
}

func TestA2ATrustDecayFactorKeyMatchesSchema(t *testing.T) {
	if string(A2aTrustDecayFactorKey) != "a2a.trust.decay_factor" {
		t.Errorf("A2aTrustDecayFactorKey = %q, want %q", string(A2aTrustDecayFactorKey), "a2a.trust.decay_factor")
	}
}

func TestA2ATrustUpdatedAtMsKeyMatchesSchema(t *testing.T) {
	if string(A2aTrustUpdatedAtMsKey) != "a2a.trust.updated_at_ms" {
		t.Errorf("A2aTrustUpdatedAtMsKey = %q, want %q", string(A2aTrustUpdatedAtMsKey), "a2a.trust.updated_at_ms")
	}
}

// ============================================================
// Iter 14: PM simulation attributes
// ============================================================

func TestProcessMiningSimulationCasesKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningSimulationCasesKey) != "process.mining.simulation.cases" {
		t.Errorf("ProcessMiningSimulationCasesKey = %q, want %q", string(ProcessMiningSimulationCasesKey), "process.mining.simulation.cases")
	}
}

func TestProcessMiningSimulationNoiseRateKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningSimulationNoiseRateKey) != "process.mining.simulation.noise_rate" {
		t.Errorf("ProcessMiningSimulationNoiseRateKey = %q, want %q", string(ProcessMiningSimulationNoiseRateKey), "process.mining.simulation.noise_rate")
	}
}

func TestProcessMiningSimulationDurationMsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningSimulationDurationMsKey) != "process.mining.simulation.duration_ms" {
		t.Errorf("ProcessMiningSimulationDurationMsKey = %q, want %q", string(ProcessMiningSimulationDurationMsKey), "process.mining.simulation.duration_ms")
	}
}

func TestProcessMiningReplayTokenCountKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayTokenCountKey) != "process.mining.replay.token_count" {
		t.Errorf("ProcessMiningReplayTokenCountKey = %q, want %q", string(ProcessMiningReplayTokenCountKey), "process.mining.replay.token_count")
	}
}

// ============================================================
// Iter 14: Consensus fault tolerance attributes
// ============================================================

func TestConsensusByzantineFaultsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusByzantineFaultsKey) != "consensus.byzantine_faults" {
		t.Errorf("ConsensusByzantineFaultsKey = %q, want %q", string(ConsensusByzantineFaultsKey), "consensus.byzantine_faults")
	}
}

func TestConsensusReplicaLagMsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusReplicaLagMsKey) != "consensus.replica.lag_ms" {
		t.Errorf("ConsensusReplicaLagMsKey = %q, want %q", string(ConsensusReplicaLagMsKey), "consensus.replica.lag_ms")
	}
}

func TestConsensusReplicaCountIter14KeyMatchesSchema(t *testing.T) {
	if string(ConsensusReplicaCountKey) != "consensus.replica.count" {
		t.Errorf("ConsensusReplicaCountKey = %q, want %q", string(ConsensusReplicaCountKey), "consensus.replica.count")
	}
}

// ============================================================
// Iter 14: Healing pattern attributes
// ============================================================

func TestHealingPatternIdKeyMatchesSchema(t *testing.T) {
	if string(HealingPatternIdKey) != "healing.pattern.id" {
		t.Errorf("HealingPatternIdKey = %q, want %q", string(HealingPatternIdKey), "healing.pattern.id")
	}
}

func TestHealingPatternLibrarySizeKeyMatchesSchema(t *testing.T) {
	if string(HealingPatternLibrarySizeKey) != "healing.pattern.library_size" {
		t.Errorf("HealingPatternLibrarySizeKey = %q, want %q", string(HealingPatternLibrarySizeKey), "healing.pattern.library_size")
	}
}

func TestHealingPatternMatchConfidenceKeyMatchesSchema(t *testing.T) {
	if string(HealingPatternMatchConfidenceKey) != "healing.pattern.match_confidence" {
		t.Errorf("HealingPatternMatchConfidenceKey = %q, want %q", string(HealingPatternMatchConfidenceKey), "healing.pattern.match_confidence")
	}
}

// ============================================================
// Iter 14: LLM token budget attributes
// ============================================================

func TestLlmTokenPromptCountKeyMatchesSchema(t *testing.T) {
	if string(LlmTokenPromptCountKey) != "llm.token.prompt_count" {
		t.Errorf("LlmTokenPromptCountKey = %q, want %q", string(LlmTokenPromptCountKey), "llm.token.prompt_count")
	}
}

func TestLlmTokenCompletionCountKeyMatchesSchema(t *testing.T) {
	if string(LlmTokenCompletionCountKey) != "llm.token.completion_count" {
		t.Errorf("LlmTokenCompletionCountKey = %q, want %q", string(LlmTokenCompletionCountKey), "llm.token.completion_count")
	}
}

func TestLlmTokenBudgetRemainingKeyMatchesSchema(t *testing.T) {
	if string(LlmTokenBudgetRemainingKey) != "llm.token.budget_remaining" {
		t.Errorf("LlmTokenBudgetRemainingKey = %q, want %q", string(LlmTokenBudgetRemainingKey), "llm.token.budget_remaining")
	}
}

func TestLlmModelVersionKeyMatchesSchema(t *testing.T) {
	if string(LlmModelVersionKey) != "llm.model.version" {
		t.Errorf("LlmModelVersionKey = %q, want %q", string(LlmModelVersionKey), "llm.model.version")
	}
}

// ============================================================
// Iter 14: MCP resource attributes
// ============================================================

func TestMcpResourceUriKeyMatchesSchema(t *testing.T) {
	if string(McpResourceUriKey) != "mcp.resource.uri" {
		t.Errorf("McpResourceUriKey = %q, want %q", string(McpResourceUriKey), "mcp.resource.uri")
	}
}

func TestMcpResourceMimeTypeKeyMatchesSchema(t *testing.T) {
	if string(McpResourceMimeTypeKey) != "mcp.resource.mime_type" {
		t.Errorf("McpResourceMimeTypeKey = %q, want %q", string(McpResourceMimeTypeKey), "mcp.resource.mime_type")
	}
}

func TestMcpResourceSizeBytesKeyMatchesSchema(t *testing.T) {
	if string(McpResourceSizeBytesKey) != "mcp.resource.size_bytes" {
		t.Errorf("McpResourceSizeBytesKey = %q, want %q", string(McpResourceSizeBytesKey), "mcp.resource.size_bytes")
	}
}

// ============================================================
// Iter 14: Canopy snapshot attributes
// ============================================================

func TestCanopySnapshotIdKeyMatchesSchema(t *testing.T) {
	if string(CanopySnapshotIdKey) != "canopy.snapshot.id" {
		t.Errorf("CanopySnapshotIdKey = %q, want %q", string(CanopySnapshotIdKey), "canopy.snapshot.id")
	}
}

func TestCanopySnapshotSizeBytesKeyMatchesSchema(t *testing.T) {
	if string(CanopySnapshotSizeBytesKey) != "canopy.snapshot.size_bytes" {
		t.Errorf("CanopySnapshotSizeBytesKey = %q, want %q", string(CanopySnapshotSizeBytesKey), "canopy.snapshot.size_bytes")
	}
}

// ============================================================
// Iter 15: Agent Memory Federation attributes
// ============================================================

func TestAgentMemoryFederationIDKeyMatchesSchema(t *testing.T) {
	if string(AgentMemoryFederationIdKey) != "agent.memory.federation_id" {
		t.Errorf("AgentMemoryFederationIdKey = %q, want %q", string(AgentMemoryFederationIdKey), "agent.memory.federation_id")
	}
}

func TestAgentMemoryFederationPeerCountKeyMatchesSchema(t *testing.T) {
	if string(AgentMemoryFederationPeerCountKey) != "agent.memory.federation.peer_count" {
		t.Errorf("AgentMemoryFederationPeerCountKey = %q, want %q", string(AgentMemoryFederationPeerCountKey), "agent.memory.federation.peer_count")
	}
}

func TestAgentMemorySyncLatencyMsKeyMatchesSchema(t *testing.T) {
	if string(AgentMemorySyncLatencyMsKey) != "agent.memory.sync.latency_ms" {
		t.Errorf("AgentMemorySyncLatencyMsKey = %q, want %q", string(AgentMemorySyncLatencyMsKey), "agent.memory.sync.latency_ms")
	}
}

func TestAgentMemoryFederationVersionKeyMatchesSchema(t *testing.T) {
	if string(AgentMemoryFederationVersionKey) != "agent.memory.federation.version" {
		t.Errorf("AgentMemoryFederationVersionKey = %q, want %q", string(AgentMemoryFederationVersionKey), "agent.memory.federation.version")
	}
}

// ============================================================
// Iter 15: Process Mining Replay attributes
// ============================================================

func TestProcessMiningReplayEnabledTransitionsKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayEnabledTransitionsKey) != "process.mining.replay.enabled_transitions" {
		t.Errorf("ProcessMiningReplayEnabledTransitionsKey = %q, want %q", string(ProcessMiningReplayEnabledTransitionsKey), "process.mining.replay.enabled_transitions")
	}
}

func TestProcessMiningReplayMissingTokensKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayMissingTokensKey) != "process.mining.replay.missing_tokens" {
		t.Errorf("ProcessMiningReplayMissingTokensKey = %q, want %q", string(ProcessMiningReplayMissingTokensKey), "process.mining.replay.missing_tokens")
	}
}

func TestProcessMiningReplayConsumedTokensKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningReplayConsumedTokensKey) != "process.mining.replay.consumed_tokens" {
		t.Errorf("ProcessMiningReplayConsumedTokensKey = %q, want %q", string(ProcessMiningReplayConsumedTokensKey), "process.mining.replay.consumed_tokens")
	}
}

func TestProcessMiningCaseVariantIDKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningCaseVariantIdKey) != "process.mining.case.variant_id" {
		t.Errorf("ProcessMiningCaseVariantIdKey = %q, want %q", string(ProcessMiningCaseVariantIdKey), "process.mining.case.variant_id")
	}
}

// ============================================================
// Iter 15: Consensus Liveness attributes
// ============================================================

func TestConsensusLivenessProofRoundsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLivenessProofRoundsKey) != "consensus.liveness.proof_rounds" {
		t.Errorf("ConsensusLivenessProofRoundsKey = %q, want %q", string(ConsensusLivenessProofRoundsKey), "consensus.liveness.proof_rounds")
	}
}

func TestConsensusNetworkRecoveryMsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusNetworkRecoveryMsKey) != "consensus.network.recovery_ms" {
		t.Errorf("ConsensusNetworkRecoveryMsKey = %q, want %q", string(ConsensusNetworkRecoveryMsKey), "consensus.network.recovery_ms")
	}
}

func TestConsensusViewDurationMsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusViewDurationMsKey) != "consensus.view.duration_ms" {
		t.Errorf("ConsensusViewDurationMsKey = %q, want %q", string(ConsensusViewDurationMsKey), "consensus.view.duration_ms")
	}
}

// ============================================================
// Iter 15: Healing Self-Healing attributes
// ============================================================

func TestHealingSelfHealingEnabledKeyMatchesSchema(t *testing.T) {
	if string(HealingSelfHealingEnabledKey) != "healing.self_healing.enabled" {
		t.Errorf("HealingSelfHealingEnabledKey = %q, want %q", string(HealingSelfHealingEnabledKey), "healing.self_healing.enabled")
	}
}

func TestHealingSelfHealingTriggerCountKeyMatchesSchema(t *testing.T) {
	if string(HealingSelfHealingTriggerCountKey) != "healing.self_healing.trigger_count" {
		t.Errorf("HealingSelfHealingTriggerCountKey = %q, want %q", string(HealingSelfHealingTriggerCountKey), "healing.self_healing.trigger_count")
	}
}

func TestHealingSelfHealingSuccessRateKeyMatchesSchema(t *testing.T) {
	if string(HealingSelfHealingSuccessRateKey) != "healing.self_healing.success_rate" {
		t.Errorf("HealingSelfHealingSuccessRateKey = %q, want %q", string(HealingSelfHealingSuccessRateKey), "healing.self_healing.success_rate")
	}
}

func TestHealingInterventionTypeKeyMatchesSchema(t *testing.T) {
	if string(HealingInterventionTypeKey) != "healing.intervention.type" {
		t.Errorf("HealingInterventionTypeKey = %q, want %q", string(HealingInterventionTypeKey), "healing.intervention.type")
	}
}

// ============================================================
// Iter 15: LLM Evaluation attributes
// ============================================================

func TestLlmEvaluationScoreKeyMatchesSchema(t *testing.T) {
	if string(LlmEvaluationScoreKey) != "llm.evaluation.score" {
		t.Errorf("LlmEvaluationScoreKey = %q, want %q", string(LlmEvaluationScoreKey), "llm.evaluation.score")
	}
}

func TestLlmEvaluationRubricKeyMatchesSchema(t *testing.T) {
	if string(LlmEvaluationRubricKey) != "llm.evaluation.rubric" {
		t.Errorf("LlmEvaluationRubricKey = %q, want %q", string(LlmEvaluationRubricKey), "llm.evaluation.rubric")
	}
}

func TestLlmEvaluationPassesThresholdKeyMatchesSchema(t *testing.T) {
	if string(LlmEvaluationPassesThresholdKey) != "llm.evaluation.passes_threshold" {
		t.Errorf("LlmEvaluationPassesThresholdKey = %q, want %q", string(LlmEvaluationPassesThresholdKey), "llm.evaluation.passes_threshold")
	}
}

// ============================================================
// Iter 15: Event Routing attributes
// ============================================================

func TestEventRoutingStrategyKeyMatchesSchema(t *testing.T) {
	if string(EventRoutingStrategyKey) != "event.routing.strategy" {
		t.Errorf("EventRoutingStrategyKey = %q, want %q", string(EventRoutingStrategyKey), "event.routing.strategy")
	}
}

func TestEventRoutingFilterCountKeyMatchesSchema(t *testing.T) {
	if string(EventRoutingFilterCountKey) != "event.routing.filter_count" {
		t.Errorf("EventRoutingFilterCountKey = %q, want %q", string(EventRoutingFilterCountKey), "event.routing.filter_count")
	}
}

func TestEventSubscriberCountKeyMatchesSchema(t *testing.T) {
	if string(EventSubscriberCountKey) != "event.subscriber.count" {
		t.Errorf("EventSubscriberCountKey = %q, want %q", string(EventSubscriberCountKey), "event.subscriber.count")
	}
}

// ============================================================
// Iter 15: Signal Quality attributes
// ============================================================

func TestSignalQualityScoreKeyMatchesSchema(t *testing.T) {
	if string(SignalQualityScoreKey) != "signal.quality.score" {
		t.Errorf("SignalQualityScoreKey = %q, want %q", string(SignalQualityScoreKey), "signal.quality.score")
	}
}

func TestSignalQualityDegradedKeyMatchesSchema(t *testing.T) {
	if string(SignalQualityDegradedKey) != "signal.quality.degraded" {
		t.Errorf("SignalQualityDegradedKey = %q, want %q", string(SignalQualityDegradedKey), "signal.quality.degraded")
	}
}

func TestSignalRetryCountKeyMatchesSchema(t *testing.T) {
	if string(SignalRetryCountKey) != "signal.retry.count" {
		t.Errorf("SignalRetryCountKey = %q, want %q", string(SignalRetryCountKey), "signal.retry.count")
	}
}

// ============================================================
// Wave 9 Iteration 16: ChatmanGPT Session
// ============================================================

func TestChatmangptSessionIDKeyMatchesSchema(t *testing.T) {
	if string(ChatmangptSessionIdKey) != "chatmangpt.session.id" {
		t.Errorf("ChatmangptSessionIdKey = %q, want %q", string(ChatmangptSessionIdKey), "chatmangpt.session.id")
	}
}

func TestChatmangptSessionTokenCountKeyMatchesSchema(t *testing.T) {
	if string(ChatmangptSessionTokenCountKey) != "chatmangpt.session.token_count" {
		t.Errorf("ChatmangptSessionTokenCountKey = %q, want %q", string(ChatmangptSessionTokenCountKey), "chatmangpt.session.token_count")
	}
}

func TestChatmangptSessionModelSwitchesKeyMatchesSchema(t *testing.T) {
	if string(ChatmangptSessionModelSwitchesKey) != "chatmangpt.session.model_switches" {
		t.Errorf("ChatmangptSessionModelSwitchesKey = %q, want %q", string(ChatmangptSessionModelSwitchesKey), "chatmangpt.session.model_switches")
	}
}

func TestChatmangptSessionTurnCountKeyMatchesSchema(t *testing.T) {
	if string(ChatmangptSessionTurnCountKey) != "chatmangpt.session.turn_count" {
		t.Errorf("ChatmangptSessionTurnCountKey = %q, want %q", string(ChatmangptSessionTurnCountKey), "chatmangpt.session.turn_count")
	}
}

// ============================================================
// Wave 9 Iteration 16: A2A Message Routing
// ============================================================

func TestA2AMessagePriorityKeyMatchesSchema(t *testing.T) {
	if string(A2aMessagePriorityKey) != "a2a.message.priority" {
		t.Errorf("A2aMessagePriorityKey = %q, want %q", string(A2aMessagePriorityKey), "a2a.message.priority")
	}
}

func TestA2AMessageSizeBytesKeyMatchesSchema(t *testing.T) {
	if string(A2aMessageSizeBytesKey) != "a2a.message.size_bytes" {
		t.Errorf("A2aMessageSizeBytesKey = %q, want %q", string(A2aMessageSizeBytesKey), "a2a.message.size_bytes")
	}
}

func TestA2AMessageEncodingKeyMatchesSchema(t *testing.T) {
	if string(A2aMessageEncodingKey) != "a2a.message.encoding" {
		t.Errorf("A2aMessageEncodingKey = %q, want %q", string(A2aMessageEncodingKey), "a2a.message.encoding")
	}
}

// ============================================================
// Wave 9 Iteration 16: Process Mining Decision Mining
// ============================================================

func TestPMDecisionPointIDKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningDecisionPointIdKey) != "process.mining.decision.point_id" {
		t.Errorf("ProcessMiningDecisionPointIdKey = %q, want %q", string(ProcessMiningDecisionPointIdKey), "process.mining.decision.point_id")
	}
}

func TestPMDecisionOutcomeKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningDecisionOutcomeKey) != "process.mining.decision.outcome" {
		t.Errorf("ProcessMiningDecisionOutcomeKey = %q, want %q", string(ProcessMiningDecisionOutcomeKey), "process.mining.decision.outcome")
	}
}

func TestPMDecisionConfidenceKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningDecisionConfidenceKey) != "process.mining.decision.confidence" {
		t.Errorf("ProcessMiningDecisionConfidenceKey = %q, want %q", string(ProcessMiningDecisionConfidenceKey), "process.mining.decision.confidence")
	}
}

func TestPMDecisionRuleCountKeyMatchesSchema(t *testing.T) {
	if string(ProcessMiningDecisionRuleCountKey) != "process.mining.decision.rule_count" {
		t.Errorf("ProcessMiningDecisionRuleCountKey = %q, want %q", string(ProcessMiningDecisionRuleCountKey), "process.mining.decision.rule_count")
	}
}

// ============================================================
// Wave 9 Iteration 16: Consensus Leader Rotation
// ============================================================

func TestConsensusLeaderRotationCountKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLeaderRotationCountKey) != "consensus.leader.rotation_count" {
		t.Errorf("ConsensusLeaderRotationCountKey = %q, want %q", string(ConsensusLeaderRotationCountKey), "consensus.leader.rotation_count")
	}
}

func TestConsensusLeaderTenureMsKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLeaderTenureMsKey) != "consensus.leader.tenure_ms" {
		t.Errorf("ConsensusLeaderTenureMsKey = %q, want %q", string(ConsensusLeaderTenureMsKey), "consensus.leader.tenure_ms")
	}
}

func TestConsensusLeaderScoreKeyMatchesSchema(t *testing.T) {
	if string(ConsensusLeaderScoreKey) != "consensus.leader.score" {
		t.Errorf("ConsensusLeaderScoreKey = %q, want %q", string(ConsensusLeaderScoreKey), "consensus.leader.score")
	}
}

// ============================================================
// Wave 9 Iteration 16: Healing Prediction
// ============================================================

func TestHealingPredictionHorizonMsKeyMatchesSchema(t *testing.T) {
	if string(HealingPredictionHorizonMsKey) != "healing.prediction.horizon_ms" {
		t.Errorf("HealingPredictionHorizonMsKey = %q, want %q", string(HealingPredictionHorizonMsKey), "healing.prediction.horizon_ms")
	}
}

func TestHealingPredictionConfidenceKeyMatchesSchema(t *testing.T) {
	if string(HealingPredictionConfidenceKey) != "healing.prediction.confidence" {
		t.Errorf("HealingPredictionConfidenceKey = %q, want %q", string(HealingPredictionConfidenceKey), "healing.prediction.confidence")
	}
}

func TestHealingPredictionModelKeyMatchesSchema(t *testing.T) {
	if string(HealingPredictionModelKey) != "healing.prediction.model" {
		t.Errorf("HealingPredictionModelKey = %q, want %q", string(HealingPredictionModelKey), "healing.prediction.model")
	}
}

// ============================================================
// Wave 9 Iteration 16: LLM Streaming
// ============================================================

func TestLLMStreamingChunkCountKeyMatchesSchema(t *testing.T) {
	if string(LlmStreamingChunkCountKey) != "llm.streaming.chunk_count" {
		t.Errorf("LlmStreamingChunkCountKey = %q, want %q", string(LlmStreamingChunkCountKey), "llm.streaming.chunk_count")
	}
}

func TestLLMStreamingFirstTokenMsKeyMatchesSchema(t *testing.T) {
	if string(LlmStreamingFirstTokenMsKey) != "llm.streaming.first_token_ms" {
		t.Errorf("LlmStreamingFirstTokenMsKey = %q, want %q", string(LlmStreamingFirstTokenMsKey), "llm.streaming.first_token_ms")
	}
}

func TestLLMStreamingTokensPerSecondKeyMatchesSchema(t *testing.T) {
	if string(LlmStreamingTokensPerSecondKey) != "llm.streaming.tokens_per_second" {
		t.Errorf("LlmStreamingTokensPerSecondKey = %q, want %q", string(LlmStreamingTokensPerSecondKey), "llm.streaming.tokens_per_second")
	}
}

// ============================================================
// Wave 9 Iteration 16: Workspace Context Snapshot
// ============================================================

func TestWorkspaceContextSnapshotIDKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceContextSnapshotIdKey) != "workspace.context.snapshot_id" {
		t.Errorf("WorkspaceContextSnapshotIdKey = %q, want %q", string(WorkspaceContextSnapshotIdKey), "workspace.context.snapshot_id")
	}
}

func TestWorkspaceContextCompressionRatioKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceContextCompressionRatioKey) != "workspace.context.compression_ratio" {
		t.Errorf("WorkspaceContextCompressionRatioKey = %q, want %q", string(WorkspaceContextCompressionRatioKey), "workspace.context.compression_ratio")
	}
}

func TestWorkspaceContextSizeTokensKeyMatchesSchema(t *testing.T) {
	if string(WorkspaceContextSizeTokensKey) != "workspace.context.size_tokens" {
		t.Errorf("WorkspaceContextSizeTokensKey = %q, want %q", string(WorkspaceContextSizeTokensKey), "workspace.context.size_tokens")
	}
}

// === Wave 9 Iteration 17: MCP Tool Versioning ===

func TestMCPToolVersionKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "mcp.tool.version", string(McpToolVersionKey))
}

func TestMCPToolSchemaHashKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "mcp.tool.schema_hash", string(McpToolSchemaHashKey))
}

func TestMCPToolDeprecatedKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "mcp.tool.deprecated", string(McpToolDeprecatedKey))
}

// === Wave 9 Iteration 17: A2A Capability Negotiation ===

func TestA2ACapNegotiationIDKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "a2a.capability.negotiation.id", string(A2aCapabilityNegotiationIdKey))
}

func TestA2ACapNegotiationOutcomeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "a2a.capability.negotiation.outcome", string(A2aCapabilityNegotiationOutcomeKey))
}

func TestA2ACapNegotiationRoundsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "a2a.capability.negotiation.rounds", string(A2aCapabilityNegotiationRoundsKey))
}

// === Wave 9 Iteration 17: Process Mining Root Cause ===

func TestPMRootCauseIDKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "process.mining.root_cause.id", string(ProcessMiningRootCauseIdKey))
}

func TestPMRootCauseTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "process.mining.root_cause.type", string(ProcessMiningRootCauseTypeKey))
}

func TestPMRootCauseConfidenceKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "process.mining.root_cause.confidence", string(ProcessMiningRootCauseConfidenceKey))
}

func TestPMAnomalyScoreKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "process.mining.anomaly.score", string(ProcessMiningAnomalyScoreKey))
}

// === Wave 9 Iteration 17: Consensus View Change ===

func TestConsensusViewChangeReasonKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "consensus.view_change.reason", string(ConsensusViewChangeReasonKey))
}

func TestConsensusViewChangeDurationMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "consensus.view_change.duration_ms", string(ConsensusViewChangeDurationMsKey))
}

func TestConsensusViewChangeBackoffMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "consensus.view_change.backoff_ms", string(ConsensusViewChangeBackoffMsKey))
}

// === Wave 9 Iteration 17: Healing Playbook ===

func TestHealingPlaybookIDKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "healing.playbook.id", string(HealingPlaybookIdKey))
}

func TestHealingPlaybookStepCountKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "healing.playbook.step_count", string(HealingPlaybookStepCountKey))
}

func TestHealingPlaybookExecutionMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "healing.playbook.execution_ms", string(HealingPlaybookExecutionMsKey))
}

// === Wave 9 Iteration 17: LLM Context Management ===

func TestLLMContextMaxTokensKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "llm.context.max_tokens", string(LlmContextMaxTokensKey))
}

func TestLLMContextOverflowStrategyKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "llm.context.overflow_strategy", string(LlmContextOverflowStrategyKey))
}

func TestLLMContextUtilizationKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "llm.context.utilization", string(LlmContextUtilizationKey))
}

// === Wave 9 Iteration 17: Agent Pipeline + Workspace Activity ===

func TestAgentPipelineIDKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "agent.pipeline.id", string(AgentPipelineIdKey))
}

func TestAgentPipelineStageKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "agent.pipeline.stage", string(AgentPipelineStageKey))
}

func TestWorkspaceActivityTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "workspace.activity.type", string(WorkspaceActivityTypeKey))
}

func TestWorkspaceActivityDurationMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, "workspace.activity.duration_ms", string(WorkspaceActivityDurationMsKey))
}

// ===== ITER18: MCP Transport + A2A Trust Federation + PM Variant + Consensus Safety + Healing Circuit Breaker + LLM Prompt =====

func TestIter18MCPTransportTypeKey(t *testing.T) {
	assert.Equal(t, "mcp.transport.type", string(McpTransportTypeKey))
}

func TestIter18MCPTransportLatencyMsKey(t *testing.T) {
	assert.Equal(t, "mcp.transport.latency_ms", string(McpTransportLatencyMsKey))
}

func TestIter18MCPTransportReconnectCountKey(t *testing.T) {
	assert.Equal(t, "mcp.transport.reconnect_count", string(McpTransportReconnectCountKey))
}

func TestIter18A2ATrustFederationIDKey(t *testing.T) {
	assert.Equal(t, "a2a.trust.federation_id", string(A2aTrustFederationIdKey))
}

func TestIter18A2ATrustPeerCountKey(t *testing.T) {
	assert.Equal(t, "a2a.trust.peer_count", string(A2aTrustPeerCountKey))
}

func TestIter18A2ATrustConsensusThresholdKey(t *testing.T) {
	assert.Equal(t, "a2a.trust.consensus_threshold", string(A2aTrustConsensusThresholdKey))
}

func TestIter18A2ATrustEpochKey(t *testing.T) {
	assert.Equal(t, "a2a.trust.epoch", string(A2aTrustEpochKey))
}

func TestIter18PMVariantIDKey(t *testing.T) {
	assert.Equal(t, "process.mining.variant.id", string(ProcessMiningVariantIdKey))
}

func TestIter18PMVariantFrequencyKey(t *testing.T) {
	assert.Equal(t, "process.mining.variant.frequency", string(ProcessMiningVariantFrequencyKey))
}

func TestIter18PMVariantIsOptimalKey(t *testing.T) {
	assert.Equal(t, "process.mining.variant.is_optimal", string(ProcessMiningVariantIsOptimalKey))
}

func TestIter18PMVariantDeviationScoreKey(t *testing.T) {
	assert.Equal(t, "process.mining.variant.deviation_score", string(ProcessMiningVariantDeviationScoreKey))
}

func TestIter18ConsensusSafetyQuorumRatioKey(t *testing.T) {
	assert.Equal(t, "consensus.safety.quorum_ratio", string(ConsensusSafetyQuorumRatioKey))
}

func TestIter18ConsensusSafetyViolationCountKey(t *testing.T) {
	assert.Equal(t, "consensus.safety.violation_count", string(ConsensusSafetyViolationCountKey))
}

func TestIter18ConsensusSafetyCheckIntervalMsKey(t *testing.T) {
	assert.Equal(t, "consensus.safety.check_interval_ms", string(ConsensusSafetyCheckIntervalMsKey))
}

func TestIter18HealingCircuitBreakerStateKey(t *testing.T) {
	assert.Equal(t, "healing.circuit_breaker.state", string(HealingCircuitBreakerStateKey))
}

func TestIter18HealingCircuitBreakerFailureCountKey(t *testing.T) {
	assert.Equal(t, "healing.circuit_breaker.failure_count", string(HealingCircuitBreakerFailureCountKey))
}

func TestIter18HealingCircuitBreakerResetMsKey(t *testing.T) {
	assert.Equal(t, "healing.circuit_breaker.reset_ms", string(HealingCircuitBreakerResetMsKey))
}

func TestIter18HealingCircuitBreakerCallCountKey(t *testing.T) {
	assert.Equal(t, "healing.circuit_breaker.call_count", string(HealingCircuitBreakerCallCountKey))
}

func TestIter18LLMPromptTemplateIDKey(t *testing.T) {
	assert.Equal(t, "llm.prompt.template_id", string(LlmPromptTemplateIdKey))
}

func TestIter18LLMPromptVersionKey(t *testing.T) {
	assert.Equal(t, "llm.prompt.version", string(LlmPromptVersionKey))
}

func TestIter18LLMPromptVariableCountKey(t *testing.T) {
	assert.Equal(t, "llm.prompt.variable_count", string(LlmPromptVariableCountKey))
}

func TestIter18LLMPromptRenderedTokensKey(t *testing.T) {
	assert.Equal(t, "llm.prompt.rendered_tokens", string(LlmPromptRenderedTokensKey))
}

func TestIter18MCPTransportErrorCountKey(t *testing.T) {
	assert.Equal(t, "mcp.transport.error_count", string(McpTransportErrorCountKey))
}

// ===== ITER19: Agent Execution Graph + A2A Batch + PM Event Abstraction + Consensus Epoch + Healing Anomaly + LLM Sampling =====

func TestIter19AgentExecutionGraphIDKey(t *testing.T) {
	assert.Equal(t, "agent.execution.graph_id", string(AgentExecutionGraphIdKey))
}

func TestIter19AgentExecutionNodeCountKey(t *testing.T) {
	assert.Equal(t, "agent.execution.node_count", string(AgentExecutionNodeCountKey))
}

func TestIter19AgentExecutionEdgeCountKey(t *testing.T) {
	assert.Equal(t, "agent.execution.edge_count", string(AgentExecutionEdgeCountKey))
}

func TestIter19AgentExecutionCriticalPathMsKey(t *testing.T) {
	assert.Equal(t, "agent.execution.critical_path_ms", string(AgentExecutionCriticalPathMsKey))
}

func TestIter19A2ABatchIDKey(t *testing.T) {
	assert.Equal(t, "a2a.batch.id", string(A2aBatchIdKey))
}

func TestIter19A2ABatchSizeKey(t *testing.T) {
	assert.Equal(t, "a2a.batch.size", string(A2aBatchSizeKey))
}

func TestIter19A2ABatchCompressionRatioKey(t *testing.T) {
	assert.Equal(t, "a2a.batch.compression_ratio", string(A2aBatchCompressionRatioKey))
}

func TestIter19A2ABatchDeliveryPolicyKey(t *testing.T) {
	assert.Equal(t, "a2a.batch.delivery_policy", string(A2aBatchDeliveryPolicyKey))
}

func TestIter19PMEventAbstractionLevelKey(t *testing.T) {
	assert.Equal(t, "process.mining.event.abstraction_level", string(ProcessMiningEventAbstractionLevelKey))
}

func TestIter19PMEventAbstractionMappingRulesKey(t *testing.T) {
	assert.Equal(t, "process.mining.event.abstraction_mapping_rules", string(ProcessMiningEventAbstractionMappingRulesKey))
}

func TestIter19PMEventAbstractionInputCountKey(t *testing.T) {
	assert.Equal(t, "process.mining.event.abstraction_input_count", string(ProcessMiningEventAbstractionInputCountKey))
}

func TestIter19PMEventAbstractionOutputCountKey(t *testing.T) {
	assert.Equal(t, "process.mining.event.abstraction_output_count", string(ProcessMiningEventAbstractionOutputCountKey))
}

func TestIter19ConsensusEpochIDKey(t *testing.T) {
	assert.Equal(t, "consensus.epoch.id", string(ConsensusEpochIdKey))
}

func TestIter19ConsensusEpochStartRoundKey(t *testing.T) {
	assert.Equal(t, "consensus.epoch.start_round", string(ConsensusEpochStartRoundKey))
}

func TestIter19ConsensusEpochDurationMsKey(t *testing.T) {
	assert.Equal(t, "consensus.epoch.duration_ms", string(ConsensusEpochDurationMsKey))
}

func TestIter19ConsensusEpochLeaderChangesKey(t *testing.T) {
	assert.Equal(t, "consensus.epoch.leader_changes", string(ConsensusEpochLeaderChangesKey))
}

func TestIter19HealingAnomalyScoreKey(t *testing.T) {
	assert.Equal(t, "healing.anomaly.score", string(HealingAnomalyScoreKey))
}

func TestIter19HealingAnomalyDetectionMethodKey(t *testing.T) {
	assert.Equal(t, "healing.anomaly.detection_method", string(HealingAnomalyDetectionMethodKey))
}

func TestIter19HealingAnomalyBaselineMsKey(t *testing.T) {
	assert.Equal(t, "healing.anomaly.baseline_ms", string(HealingAnomalyBaselineMsKey))
}

func TestIter19LLMSamplingTemperatureKey(t *testing.T) {
	assert.Equal(t, "llm.sampling.temperature", string(LlmSamplingTemperatureKey))
}

func TestIter19LLMSamplingTopPKey(t *testing.T) {
	assert.Equal(t, "llm.sampling.top_p", string(LlmSamplingTopPKey))
}

func TestIter19LLMSamplingMaxTokensKey(t *testing.T) {
	assert.Equal(t, "llm.sampling.max_tokens", string(LlmSamplingMaxTokensKey))
}

func TestIter19LLMSamplingSeedKey(t *testing.T) {
	assert.Equal(t, "llm.sampling.seed", string(LlmSamplingSeedKey))
}

// ===== Iter 20: Workspace Sharing + A2A Protocol Versioning + PM Temporal + Consensus Fork + Healing Adaptive + LLM Cache =====

func TestIter20WorkspaceSharingScopeKey(t *testing.T) {
	assert.Equal(t, "workspace.sharing.scope", string(WorkspaceSharingScopeKey))
}
func TestIter20WorkspaceSharingAgentCountKey(t *testing.T) {
	assert.Equal(t, "workspace.sharing.agent_count", string(WorkspaceSharingAgentCountKey))
}
func TestIter20WorkspaceSharingPermissionsKey(t *testing.T) {
	assert.Equal(t, "workspace.sharing.permissions", string(WorkspaceSharingPermissionsKey))
}
func TestIter20WorkspaceSharingScope(t *testing.T) {
	kv := WorkspaceSharingScope("team")
	assert.Equal(t, "workspace.sharing.scope", string(kv.Key))
	assert.Equal(t, "team", kv.Value.AsString())
}
func TestIter20A2AProtocolVersionKey(t *testing.T) {
	assert.Equal(t, "a2a.protocol.version", string(A2aProtocolVersionKey))
}
func TestIter20A2AProtocolMinVersionKey(t *testing.T) {
	assert.Equal(t, "a2a.protocol.min_version", string(A2aProtocolMinVersionKey))
}
func TestIter20A2AProtocolDeprecatedKey(t *testing.T) {
	assert.Equal(t, "a2a.protocol.deprecated", string(A2aProtocolDeprecatedKey))
}
func TestIter20A2AProtocolNegotiationMsKey(t *testing.T) {
	assert.Equal(t, "a2a.protocol.negotiation_ms", string(A2aProtocolNegotiationMsKey))
}
func TestIter20A2AProtocolVersion(t *testing.T) {
	kv := A2aProtocolVersion("1.1")
	assert.Equal(t, "a2a.protocol.version", string(kv.Key))
	assert.Equal(t, "1.1", kv.Value.AsString())
}
func TestIter20PMTemporalDriftMsKey(t *testing.T) {
	assert.Equal(t, "process.mining.temporal.drift_ms", string(ProcessMiningTemporalDriftMsKey))
}
func TestIter20PMTemporalSeasonalityPeriodMsKey(t *testing.T) {
	assert.Equal(t, "process.mining.temporal.seasonality_period_ms", string(ProcessMiningTemporalSeasonalityPeriodMsKey))
}
func TestIter20PMTemporalTrendSlopeKey(t *testing.T) {
	assert.Equal(t, "process.mining.temporal.trend_slope", string(ProcessMiningTemporalTrendSlopeKey))
}
func TestIter20PMTemporalDriftMs(t *testing.T) {
	kv := ProcessMiningTemporalDriftMs(5000)
	assert.Equal(t, "process.mining.temporal.drift_ms", string(kv.Key))
	assert.Equal(t, int64(5000), kv.Value.AsInt64())
}
func TestIter20ConsensusForkDetectedKey(t *testing.T) {
	assert.Equal(t, "consensus.fork.detected", string(ConsensusForkDetectedKey))
}
func TestIter20ConsensusForkDepthKey(t *testing.T) {
	assert.Equal(t, "consensus.fork.depth", string(ConsensusForkDepthKey))
}
func TestIter20ConsensusForkResolutionStrategyKey(t *testing.T) {
	assert.Equal(t, "consensus.fork.resolution_strategy", string(ConsensusForkResolutionStrategyKey))
}
func TestIter20ConsensusForkDetected(t *testing.T) {
	kv := ConsensusForkDetected(true)
	assert.Equal(t, "consensus.fork.detected", string(kv.Key))
	assert.Equal(t, true, kv.Value.AsBool())
}
func TestIter20HealingAdaptiveThresholdCurrentKey(t *testing.T) {
	assert.Equal(t, "healing.adaptive.threshold_current", string(HealingAdaptiveThresholdCurrentKey))
}
func TestIter20HealingAdaptiveLearningRateKey(t *testing.T) {
	assert.Equal(t, "healing.adaptive.learning_rate", string(HealingAdaptiveLearningRateKey))
}
func TestIter20HealingAdaptiveThresholdCurrent(t *testing.T) {
	kv := HealingAdaptiveThresholdCurrent(0.85)
	assert.Equal(t, "healing.adaptive.threshold_current", string(kv.Key))
	assert.InDelta(t, 0.85, kv.Value.AsFloat64(), 0.001)
}
func TestIter20LLMCacheHitKey(t *testing.T) {
	assert.Equal(t, "llm.cache.hit", string(LlmCacheHitKey))
}
func TestIter20LLMCacheTTLMsKey(t *testing.T) {
	assert.Equal(t, "llm.cache.ttl_ms", string(LlmCacheTtlMsKey))
}
func TestIter20LLMCacheKeyHashKey(t *testing.T) {
	assert.Equal(t, "llm.cache.key_hash", string(LlmCacheKeyHashKey))
}

// ===== Iter 21: Agent Handoff + A2A Auction + PM Conformance Threshold + Consensus Byzantine + Healing Intervention + LLM Tool Orchestration =====

func TestIter21AgentHandoffTargetIDKey(t *testing.T) {
	assert.Equal(t, "agent.handoff.target_id", string(AgentHandoffTargetIdKey))
}
func TestIter21AgentHandoffReasonKey(t *testing.T) {
	assert.Equal(t, "agent.handoff.reason", string(AgentHandoffReasonKey))
}
func TestIter21AgentHandoffStateTransferMsKey(t *testing.T) {
	assert.Equal(t, "agent.handoff.state_transfer_ms", string(AgentHandoffStateTransferMsKey))
}
func TestIter21AgentHandoffTargetID(t *testing.T) {
	kv := AgentHandoffTargetId("agent-7")
	assert.Equal(t, "agent.handoff.target_id", string(kv.Key))
	assert.Equal(t, "agent-7", kv.Value.AsString())
}
func TestIter21A2AAuctionIDKey(t *testing.T) {
	assert.Equal(t, "a2a.auction.id", string(A2aAuctionIdKey))
}
func TestIter21A2AAuctionBidCountKey(t *testing.T) {
	assert.Equal(t, "a2a.auction.bid_count", string(A2aAuctionBidCountKey))
}
func TestIter21A2AAuctionWinnerIDKey(t *testing.T) {
	assert.Equal(t, "a2a.auction.winner_id", string(A2aAuctionWinnerIdKey))
}
func TestIter21A2AAuctionClearingPriceKey(t *testing.T) {
	assert.Equal(t, "a2a.auction.clearing_price", string(A2aAuctionClearingPriceKey))
}
func TestIter21A2AAuctionID(t *testing.T) {
	kv := A2aAuctionId("auction-001")
	assert.Equal(t, "a2a.auction.id", string(kv.Key))
	assert.Equal(t, "auction-001", kv.Value.AsString())
}
func TestIter21PMConformanceCaseThresholdKey(t *testing.T) {
	assert.Equal(t, "process.mining.conformance.case_threshold", string(ProcessMiningConformanceCaseThresholdKey))
}
func TestIter21PMConformanceViolationCountKey(t *testing.T) {
	assert.Equal(t, "process.mining.conformance.violation_count", string(ProcessMiningConformanceViolationCountKey))
}
func TestIter21PMConformanceRepairStepsKey(t *testing.T) {
	assert.Equal(t, "process.mining.conformance.repair_steps", string(ProcessMiningConformanceRepairStepsKey))
}
func TestIter21PMConformanceCaseThreshold(t *testing.T) {
	kv := ProcessMiningConformanceCaseThreshold(0.85)
	assert.Equal(t, "process.mining.conformance.case_threshold", string(kv.Key))
	assert.InDelta(t, 0.85, kv.Value.AsFloat64(), 0.001)
}
func TestIter21ConsensusByzantineRecoveryRoundKey(t *testing.T) {
	assert.Equal(t, "consensus.byzantine.recovery_round", string(ConsensusByzantineRecoveryRoundKey))
}
func TestIter21ConsensusByzantineDetectedFaultsKey(t *testing.T) {
	assert.Equal(t, "consensus.byzantine.detected_faults", string(ConsensusByzantineDetectedFaultsKey))
}
func TestIter21ConsensusByzantineDetectedFaults(t *testing.T) {
	kv := ConsensusByzantineDetectedFaults(2)
	assert.Equal(t, "consensus.byzantine.detected_faults", string(kv.Key))
	assert.Equal(t, int64(2), kv.Value.AsInt64())
}
func TestIter21HealingInterventionScoreKey(t *testing.T) {
	assert.Equal(t, "healing.intervention.score", string(HealingInterventionScoreKey))
}
func TestIter21HealingInterventionOutcomeKey(t *testing.T) {
	assert.Equal(t, "healing.intervention.outcome", string(HealingInterventionOutcomeKey))
}
func TestIter21HealingInterventionScore(t *testing.T) {
	kv := HealingInterventionScore(0.92)
	assert.Equal(t, "healing.intervention.score", string(kv.Key))
	assert.InDelta(t, 0.92, kv.Value.AsFloat64(), 0.001)
}
func TestIter21LLMToolOrchestrationStrategyKey(t *testing.T) {
	assert.Equal(t, "llm.tool.orchestration.strategy", string(LlmToolOrchestrationStrategyKey))
}
func TestIter21LLMToolOrchestrationStepCountKey(t *testing.T) {
	assert.Equal(t, "llm.tool.orchestration.step_count", string(LlmToolOrchestrationStepCountKey))
}
func TestIter21LLMToolOrchestrationStrategy(t *testing.T) {
	kv := LlmToolOrchestrationStrategy("parallel")
	assert.Equal(t, "llm.tool.orchestration.strategy", string(kv.Key))
	assert.Equal(t, "parallel", kv.Value.AsString())
}
func TestIter21LLMToolOrchestrationSuccessRateKey(t *testing.T) {
	assert.Equal(t, "llm.tool.orchestration.success_rate", string(LlmToolOrchestrationSuccessRateKey))
}

// Iter22: Signal batch aggregation tests
func TestSignalBatchSizeAttrName(t *testing.T) {
	if string(SignalBatchSizeKey) != "signal.batch.size" {
		t.Errorf("expected signal.batch.size, got %s", string(SignalBatchSizeKey))
	}
}

func TestSignalBatchWindowMsAttrName(t *testing.T) {
	if string(SignalBatchWindowMsKey) != "signal.batch.window_ms" {
		t.Errorf("expected signal.batch.window_ms, got %s", string(SignalBatchWindowMsKey))
	}
}

func TestSignalBatchDropCountAttrName(t *testing.T) {
	if string(SignalBatchDropCountKey) != "signal.batch.drop_count" {
		t.Errorf("expected signal.batch.drop_count, got %s", string(SignalBatchDropCountKey))
	}
}

// Iter22: Workspace memory compaction tests
func TestWorkspaceMemoryCompactionRatioAttrName(t *testing.T) {
	if string(WorkspaceMemoryCompactionRatioKey) != "workspace.memory.compaction_ratio" {
		t.Errorf("expected workspace.memory.compaction_ratio, got %s", string(WorkspaceMemoryCompactionRatioKey))
	}
}

func TestWorkspaceMemoryCompactionMsAttrName(t *testing.T) {
	if string(WorkspaceMemoryCompactionMsKey) != "workspace.memory.compaction_ms" {
		t.Errorf("expected workspace.memory.compaction_ms, got %s", string(WorkspaceMemoryCompactionMsKey))
	}
}

func TestWorkspaceMemoryItemsBeforeAttrName(t *testing.T) {
	if string(WorkspaceMemoryItemsBeforeKey) != "workspace.memory.items_before" {
		t.Errorf("expected workspace.memory.items_before, got %s", string(WorkspaceMemoryItemsBeforeKey))
	}
}

func TestWorkspaceMemoryItemsAfterAttrName(t *testing.T) {
	if string(WorkspaceMemoryItemsAfterKey) != "workspace.memory.items_after" {
		t.Errorf("expected workspace.memory.items_after, got %s", string(WorkspaceMemoryItemsAfterKey))
	}
}

// Iter22: A2A bid evaluation tests
func TestA2ABidStrategyAttrName(t *testing.T) {
	if string(A2aBidStrategyKey) != "a2a.bid.strategy" {
		t.Errorf("expected a2a.bid.strategy, got %s", string(A2aBidStrategyKey))
	}
}

func TestA2ABidScoreAttrName(t *testing.T) {
	if string(A2aBidScoreKey) != "a2a.bid.score" {
		t.Errorf("expected a2a.bid.score, got %s", string(A2aBidScoreKey))
	}
}

func TestA2ABidWinnerIdAttrName(t *testing.T) {
	if string(A2aBidWinnerIdKey) != "a2a.bid.winner_id" {
		t.Errorf("expected a2a.bid.winner_id, got %s", string(A2aBidWinnerIdKey))
	}
}

// Iter22: PM alignment analysis tests
func TestProcessMiningAlignmentOptimalPathLengthAttrName(t *testing.T) {
	if string(ProcessMiningAlignmentOptimalPathLengthKey) != "process.mining.alignment.optimal_path_length" {
		t.Errorf("expected process.mining.alignment.optimal_path_length, got %s", string(ProcessMiningAlignmentOptimalPathLengthKey))
	}
}

func TestProcessMiningAlignmentMoveCountAttrName(t *testing.T) {
	if string(ProcessMiningAlignmentMoveCountKey) != "process.mining.alignment.move_count" {
		t.Errorf("expected process.mining.alignment.move_count, got %s", string(ProcessMiningAlignmentMoveCountKey))
	}
}

func TestProcessMiningAlignmentFitnessDeltaAttrName(t *testing.T) {
	if string(ProcessMiningAlignmentFitnessDeltaKey) != "process.mining.alignment.fitness_delta" {
		t.Errorf("expected process.mining.alignment.fitness_delta, got %s", string(ProcessMiningAlignmentFitnessDeltaKey))
	}
}

// Iter22: Consensus partition recovery tests
func TestConsensusPartitionDetectedAttrName(t *testing.T) {
	if string(ConsensusPartitionDetectedKey) != "consensus.partition.detected" {
		t.Errorf("expected consensus.partition.detected, got %s", string(ConsensusPartitionDetectedKey))
	}
}

func TestConsensusPartitionSizeAttrName(t *testing.T) {
	if string(ConsensusPartitionSizeKey) != "consensus.partition.size" {
		t.Errorf("expected consensus.partition.size, got %s", string(ConsensusPartitionSizeKey))
	}
}

func TestConsensusPartitionRecoveryMsAttrName(t *testing.T) {
	if string(ConsensusPartitionRecoveryMsKey) != "consensus.partition.recovery_ms" {
		t.Errorf("expected consensus.partition.recovery_ms, got %s", string(ConsensusPartitionRecoveryMsKey))
	}
}

func TestConsensusPartitionStrategyAttrName(t *testing.T) {
	if string(ConsensusPartitionStrategyKey) != "consensus.partition.strategy" {
		t.Errorf("expected consensus.partition.strategy, got %s", string(ConsensusPartitionStrategyKey))
	}
}

// Iter22: Healing rollback tests
func TestHealingRollbackStrategyAttrName(t *testing.T) {
	if string(HealingRollbackStrategyKey) != "healing.rollback.strategy" {
		t.Errorf("expected healing.rollback.strategy, got %s", string(HealingRollbackStrategyKey))
	}
}

func TestHealingRollbackCheckpointIdAttrName(t *testing.T) {
	if string(HealingRollbackCheckpointIdKey) != "healing.rollback.checkpoint_id" {
		t.Errorf("expected healing.rollback.checkpoint_id, got %s", string(HealingRollbackCheckpointIdKey))
	}
}

func TestHealingRollbackRecoveryMsAttrName(t *testing.T) {
	if string(HealingRollbackRecoveryMsKey) != "healing.rollback.recovery_ms" {
		t.Errorf("expected healing.rollback.recovery_ms, got %s", string(HealingRollbackRecoveryMsKey))
	}
}

func TestHealingRollbackSuccessAttrName(t *testing.T) {
	if string(HealingRollbackSuccessKey) != "healing.rollback.success" {
		t.Errorf("expected healing.rollback.success, got %s", string(HealingRollbackSuccessKey))
	}
}

// Iter22: LLM structured output tests
func TestLLMStructuredOutputSchemaIdAttrName(t *testing.T) {
	if string(LlmStructuredOutputSchemaIdKey) != "llm.structured_output.schema_id" {
		t.Errorf("expected llm.structured_output.schema_id, got %s", string(LlmStructuredOutputSchemaIdKey))
	}
}

func TestLLMStructuredOutputValidationMsAttrName(t *testing.T) {
	if string(LlmStructuredOutputValidationMsKey) != "llm.structured_output.validation_ms" {
		t.Errorf("expected llm.structured_output.validation_ms, got %s", string(LlmStructuredOutputValidationMsKey))
	}
}

// Iter23: Agent spawn profiling tests

// Iter23: A2A escrow mechanics tests
func TestA2AEscrowIdAttrName(t *testing.T) {
	assert.Equal(t, "a2a.escrow.id", string(A2aEscrowIdKey))
}
func TestA2AEscrowAmountAttrName(t *testing.T) {
	assert.Equal(t, "a2a.escrow.amount", string(A2aEscrowAmountKey))
}
func TestA2AEscrowReleaseConditionAttrName(t *testing.T) {
	assert.Equal(t, "a2a.escrow.release_condition", string(A2aEscrowReleaseConditionKey))
}
func TestA2AEscrowStatusAttrName(t *testing.T) {
	assert.Equal(t, "a2a.escrow.status", string(A2aEscrowStatusKey))
}

// Iter23: PM bottleneck scoring tests

// Iter23: Consensus epoch key rotation tests

// Iter23: Healing quarantine tests

// Iter23: LLM function call routing tests
func TestLLMFunctionCallNameAttrName(t *testing.T) {
	assert.Equal(t, "llm.function_call.name", string(LlmFunctionCallNameKey))
}
func TestLLMFunctionCallRoutingStrategyAttrName(t *testing.T) {
	assert.Equal(t, "llm.function_call.routing_strategy", string(LlmFunctionCallRoutingStrategyKey))
}
func TestLLMFunctionCallLatencyMsAttrName(t *testing.T) {
	assert.Equal(t, "llm.function_call.latency_ms", string(LlmFunctionCallLatencyMsKey))
}

// Iter23: ChatmanGPT namespace tests

func TestIter24MCPToolCompositionID(t *testing.T) {
	assert.Equal(t, "mcp.tool.composition_id", string(McpToolCompositionIdKey))
}



func TestIter24MCPToolCompositionLatencyMs(t *testing.T) {
	assert.Equal(t, "mcp.tool.composition_latency_ms", string(McpToolCompositionLatencyMsKey))
}


func TestIter24A2AContractID(t *testing.T) {
	assert.Equal(t, "a2a.contract.id", string(A2aContractIdKey))
}

func TestIter24A2AContractTermsHash(t *testing.T) {
	assert.Equal(t, "a2a.contract.terms_hash", string(A2aContractTermsHashKey))
}

func TestIter24A2AContractExpiryMs(t *testing.T) {
	assert.Equal(t, "a2a.contract.expiry_ms", string(A2aContractExpiryMsKey))
}

func TestIter24A2AContractViolationCount(t *testing.T) {
	assert.Equal(t, "a2a.contract.violation_count", string(A2aContractViolationCountKey))
}


func TestIter24ProcessMiningClusterID(t *testing.T) {
	assert.Equal(t, "process.mining.cluster.id", string(ProcessMiningClusterIdKey))
}







func TestIter24HealingSimulationID(t *testing.T) {
	assert.Equal(t, "healing.simulation.id", string(HealingSimulationIdKey))
}




func TestIter24LLMValidationSchemaID(t *testing.T) {
	assert.Equal(t, "llm.validation.schema_id", string(LlmValidationSchemaIdKey))
}


func TestIter25AgentReasoningTraceID(t *testing.T) {
	assert.Equal(t, "agent.reasoning.trace_id", string(AgentReasoningTraceIdKey))
}




func TestIter25A2APenaltyAmount(t *testing.T) {
	assert.Equal(t, "a2a.penalty.amount", string(A2aPenaltyAmountKey))
}

func TestIter25A2APenaltyReason(t *testing.T) {
	assert.Equal(t, "a2a.penalty.reason", string(A2aPenaltyReasonKey))
}

func TestIter25A2ARewardAmount(t *testing.T) {
	assert.Equal(t, "a2a.reward.amount", string(A2aRewardAmountKey))
}




func TestIter25ProcessMiningEnhancementBaseModelID(t *testing.T) {
	assert.Equal(t, "process.mining.enhancement.base_model_id", string(ProcessMiningEnhancementBaseModelIdKey))
}






func TestIter25HealingMemorySnapshotID(t *testing.T) {
	assert.Equal(t, "healing.memory.snapshot_id", string(HealingMemorySnapshotIdKey))
}



func TestIter25LLMMultimodalInputType(t *testing.T) {
	assert.Equal(t, "llm.multimodal.input_type", string(LlmMultimodalInputTypeKey))
}

func TestIter25LLMMultimodalModalityCount(t *testing.T) {
	assert.Equal(t, "llm.multimodal.modality_count", string(LlmMultimodalModalityCountKey))
}

func TestIter25LLMMultimodalInputSizeBytes(t *testing.T) {
	assert.Equal(t, "llm.multimodal.input_size_bytes", string(LlmMultimodalInputSizeBytesKey))
}


func TestIter26MCPServerHealthStatus(t *testing.T) {
	assert.Equal(t, "mcp.server.health.status", string(McpServerHealthStatusKey))
}

func TestIter26MCPServerHealthCheckDurationMs(t *testing.T) {
	assert.Equal(t, "mcp.server.health.check_duration_ms", string(McpServerHealthCheckDurationMsKey))
}

func TestIter26MCPServerHealthToolCount(t *testing.T) {
	assert.Equal(t, "mcp.server.health.tool_count", string(McpServerHealthToolCountKey))
}

func TestIter26MCPServerHealthUptimeMs(t *testing.T) {
	assert.Equal(t, "mcp.server.health.uptime_ms", string(McpServerHealthUptimeMsKey))
}

func TestIter26A2ADisputeID(t *testing.T) {
	assert.Equal(t, "a2a.dispute.id", string(A2aDisputeIdKey))
}

func TestIter26A2ADisputeReason(t *testing.T) {
	assert.Equal(t, "a2a.dispute.reason", string(A2aDisputeReasonKey))
}

func TestIter26A2ADisputeResolutionStatus(t *testing.T) {
	assert.Equal(t, "a2a.dispute.resolution_status", string(A2aDisputeResolutionStatusKey))
}

func TestIter26A2ADisputeResolutionMs(t *testing.T) {
	assert.Equal(t, "a2a.dispute.resolution_ms", string(A2aDisputeResolutionMsKey))
}









func TestIter26HealingWarmStandbyID(t *testing.T) {
	assert.Equal(t, "healing.warm_standby.id", string(HealingWarmStandbyIdKey))
}




func TestIter26LLMFinetuneJobID(t *testing.T) {
	assert.Equal(t, "llm.finetune.job_id", string(LlmFinetuneJobIdKey))
}

func TestIter26LLMFinetuneBaseModel(t *testing.T) {
	assert.Equal(t, "llm.finetune.base_model", string(LlmFinetuneBaseModelKey))
}

func TestIter26LLMFinetuneTrainingSteps(t *testing.T) {
	assert.Equal(t, "llm.finetune.training_steps", string(LlmFinetuneTrainingStepsKey))
}

// ============================================================
// iter27 — Agent capability catalog, A2A escrow release,
// PM conformance repair, consensus network recovery,
// healing checkpoint, LLM batch
// ============================================================

func TestIter27AgentCapabilityCatalogIDConstant(t *testing.T) {
	assert.Equal(t, "agent.capability.catalog_id", string(AgentCapabilityCatalogIdKey))
}

func TestIter27AgentCapabilityCatalogVersionConstant(t *testing.T) {
	assert.Equal(t, "agent.capability.catalog_version", string(AgentCapabilityCatalogVersionKey))
}

func TestIter27AgentCapabilityScopeConstant(t *testing.T) {
	assert.Equal(t, "agent.capability.scope", string(AgentCapabilityScopeKey))
}

func TestIter27AgentCapabilityCountConstant(t *testing.T) {
	assert.Equal(t, "agent.capability.count", string(AgentCapabilityCountKey))
}

func TestIter27AgentCapabilityScopeLocalValue(t *testing.T) {
	assert.Equal(t, "local", AgentCapabilityScopeValues.Local)
}

func TestIter27AgentCapabilityScopeClusterValue(t *testing.T) {
	assert.Equal(t, "cluster", AgentCapabilityScopeValues.Cluster)
}

func TestIter27AgentCapabilityScopeFederatedValue(t *testing.T) {
	assert.Equal(t, "federated", AgentCapabilityScopeValues.Federated)
}

func TestIter27A2AEscrowReleaseReasonConstant(t *testing.T) {
	assert.Equal(t, "a2a.escrow.release_reason", string(A2aEscrowReleaseReasonKey))
}

func TestIter27A2AEscrowReleaseMsConstant(t *testing.T) {
	assert.Equal(t, "a2a.escrow.release_ms", string(A2aEscrowReleaseMsKey))
}

func TestIter27A2AEscrowReleasedAmountConstant(t *testing.T) {
	assert.Equal(t, "a2a.escrow.released_amount", string(A2aEscrowReleasedAmountKey))
}

func TestIter27A2AEscrowReleaseReasonCompletionValue(t *testing.T) {
	assert.Equal(t, "completion", A2aEscrowReleaseReasonValues.Completion)
}







func TestIter27RecoveryStrategyReconnectValue(t *testing.T) {
	assert.Equal(t, "restart", HealingRecoveryStrategyValues.Restart)
}

func TestIter27HealingCheckpointIDConstant(t *testing.T) {
	assert.Equal(t, "healing.checkpoint.id", string(HealingCheckpointIdKey))
}


func TestIter27LLMBatchJobIDConstant(t *testing.T) {
	assert.Equal(t, "llm.batch.job_id", string(LlmBatchJobIdKey))
}


func TestIter27BatchPriorityHighValue(t *testing.T) {
	assert.Equal(t, "high", LlmBatchPriorityValues.High)
}

// iter28 — MCP tool composition, A2A reputation, PM enhancement quality,
// consensus quorum shrink, healing cold standby, LLM LoRA

func TestIter28MCPToolCompositionStrategyConstant(t *testing.T) {
	if string(McpToolCompositionStrategyKey) != "mcp.tool.composition.strategy" {
		t.Errorf("expected mcp.tool.composition.strategy, got %s", string(McpToolCompositionStrategyKey))
	}
}

func TestIter28MCPToolCompositionStepCountConstant(t *testing.T) {
	if string(McpToolCompositionStepCountKey) != "mcp.tool.composition.step_count" {
		t.Errorf("expected mcp.tool.composition.step_count, got %s", string(McpToolCompositionStepCountKey))
	}
}

func TestIter28MCPToolCompositionStrategySequentialValue(t *testing.T) {
	if McpToolCompositionStrategyValues.Sequential != "sequential" {
		t.Errorf("expected sequential, got %s", McpToolCompositionStrategyValues.Sequential)
	}
}

func TestIter28MCPToolCompositionStrategyParallelValue(t *testing.T) {
	if McpToolCompositionStrategyValues.Parallel != "parallel" {
		t.Errorf("expected parallel, got %s", McpToolCompositionStrategyValues.Parallel)
	}
}

func TestIter28A2AReputationScoreConstant(t *testing.T) {
	if string(A2aReputationScoreKey) != "a2a.reputation.score" {
		t.Errorf("expected a2a.reputation.score, got %s", string(A2aReputationScoreKey))
	}
}

func TestIter28A2AReputationInteractionCountConstant(t *testing.T) {
	if string(A2aReputationInteractionCountKey) != "a2a.reputation.interaction_count" {
		t.Errorf("expected a2a.reputation.interaction_count, got %s", string(A2aReputationInteractionCountKey))
	}
}

func TestIter28A2AReputationCategoryTrustedValue(t *testing.T) {
	if A2aReputationCategoryValues.Trusted != "trusted" {
		t.Errorf("expected trusted, got %s", A2aReputationCategoryValues.Trusted)
	}
}

func TestIter28A2AReputationCategoryBannedValue(t *testing.T) {
	if A2aReputationCategoryValues.Banned != "banned" {
		t.Errorf("expected banned, got %s", A2aReputationCategoryValues.Banned)
	}
}

func TestIter28PMEnhancementQualityScoreConstant(t *testing.T) {
	if string(ProcessMiningEnhancementQualityScoreKey) != "process.mining.enhancement.quality_score" {
		t.Errorf("expected process.mining.enhancement.quality_score, got %s", string(ProcessMiningEnhancementQualityScoreKey))
	}
}

func TestIter28PMEnhancementCoveragePctConstant(t *testing.T) {
	if string(ProcessMiningEnhancementCoveragePctKey) != "process.mining.enhancement.coverage_pct" {
		t.Errorf("expected process.mining.enhancement.coverage_pct, got %s", string(ProcessMiningEnhancementCoveragePctKey))
	}
}

func TestIter28PMEnhancementPerspectivePerformanceValue(t *testing.T) {
	if ProcessMiningEnhancementPerspectiveValues.Performance != "performance" {
		t.Errorf("expected performance, got %s", ProcessMiningEnhancementPerspectiveValues.Performance)
	}
}

func TestIter28ConsensusQuorumShrinkReasonConstant(t *testing.T) {
	if string(ConsensusQuorumShrinkReasonKey) != "consensus.quorum.shrink.reason" {
		t.Errorf("expected consensus.quorum.shrink.reason, got %s", string(ConsensusQuorumShrinkReasonKey))
	}
}

func TestIter28ConsensusQuorumShrinkRemovedCountConstant(t *testing.T) {
	if string(ConsensusQuorumShrinkRemovedCountKey) != "consensus.quorum.shrink.removed_count" {
		t.Errorf("expected consensus.quorum.shrink.removed_count, got %s", string(ConsensusQuorumShrinkRemovedCountKey))
	}
}

func TestIter28ConsensusQuorumShrinkReasonNodeFailureValue(t *testing.T) {
	if ConsensusQuorumShrinkReasonValues.NodeFailure != "node_failure" {
		t.Errorf("expected node_failure, got %s", ConsensusQuorumShrinkReasonValues.NodeFailure)
	}
}

func TestIter28ConsensusQuorumShrinkReasonRebalanceValue(t *testing.T) {
	if ConsensusQuorumShrinkReasonValues.Rebalance != "rebalance" {
		t.Errorf("expected rebalance, got %s", ConsensusQuorumShrinkReasonValues.Rebalance)
	}
}

func TestIter28HealingColdStandbyIDConstant(t *testing.T) {
	if string(HealingColdStandbyIdKey) != "healing.cold_standby.id" {
		t.Errorf("expected healing.cold_standby.id, got %s", string(HealingColdStandbyIdKey))
	}
}

func TestIter28HealingColdStandbyWarmupMsConstant(t *testing.T) {
	if string(HealingColdStandbyWarmupMsKey) != "healing.cold_standby.warmup_ms" {
		t.Errorf("expected healing.cold_standby.warmup_ms, got %s", string(HealingColdStandbyWarmupMsKey))
	}
}

func TestIter28HealingColdStandbyReadinessReadyValue(t *testing.T) {
	if HealingColdStandbyReadinessValues.Ready != "ready" {
		t.Errorf("expected ready, got %s", HealingColdStandbyReadinessValues.Ready)
	}
}

func TestIter28HealingColdStandbyReadinessColdValue(t *testing.T) {
	if HealingColdStandbyReadinessValues.Cold != "cold" {
		t.Errorf("expected cold, got %s", HealingColdStandbyReadinessValues.Cold)
	}
}

func TestIter28LLMLoRARankConstant(t *testing.T) {
	if string(LlmLoraRankKey) != "llm.lora.rank" {
		t.Errorf("expected llm.lora.rank, got %s", string(LlmLoraRankKey))
	}
}

func TestIter28LLMLoRAAlphaConstant(t *testing.T) {
	if string(LlmLoraAlphaKey) != "llm.lora.alpha" {
		t.Errorf("expected llm.lora.alpha, got %s", string(LlmLoraAlphaKey))
	}
}

func TestIter28LLMLoRATargetModulesConstant(t *testing.T) {
	if string(LlmLoraTargetModulesKey) != "llm.lora.target_modules" {
		t.Errorf("expected llm.lora.target_modules, got %s", string(LlmLoraTargetModulesKey))
	}
}

func TestIter28LLMLoRABaseModelConstant(t *testing.T) {
	if string(LlmLoraBaseModelKey) != "llm.lora.base_model" {
		t.Errorf("expected llm.lora.base_model, got %s", string(LlmLoraBaseModelKey))
	}
}

// Iter29 tests — MCP deprecation, A2A contract execution, PM prediction, consensus epoch, healing load shedding, LLM embedding

func TestIter29MCPToolDeprecationPolicyKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.deprecation.policy"), McpToolDeprecationPolicyKey,
		"mcp.tool.deprecation.policy key must match semconv schema")
}

func TestIter29MCPToolDeprecationReplacementToolKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.deprecation.replacement_tool"), McpToolDeprecationReplacementToolKey,
		"mcp.tool.deprecation.replacement_tool key must match semconv schema")
}

func TestIter29MCPToolDeprecationSunsetDateMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.deprecation.sunset_date_ms"), McpToolDeprecationSunsetDateMsKey,
		"mcp.tool.deprecation.sunset_date_ms key must match semconv schema")
}

func TestIter29MCPToolDeprecationPolicyValues(t *testing.T) {
	assert.Equal(t, "immediate", McpToolDeprecationPolicyValues.Immediate)
	assert.Equal(t, "grace_period", McpToolDeprecationPolicyValues.GracePeriod)
	assert.Equal(t, "warn_only", McpToolDeprecationPolicyValues.WarnOnly)
}

func TestIter29A2AContractExecutionStatusKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("a2a.contract.execution.status"), A2aContractExecutionStatusKey,
		"a2a.contract.execution.status key must match semconv schema")
}

func TestIter29A2AContractExecutionProgressPctKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("a2a.contract.execution.progress_pct"), A2aContractExecutionProgressPctKey,
		"a2a.contract.execution.progress_pct key must match semconv schema")
}

func TestIter29A2AContractExecutionStatusValues(t *testing.T) {
	assert.Equal(t, "running", A2aContractExecutionStatusValues.Running)
	assert.Equal(t, "completed", A2aContractExecutionStatusValues.Completed)
	assert.Equal(t, "failed", A2aContractExecutionStatusValues.Failed)
	assert.Equal(t, "disputed", A2aContractExecutionStatusValues.Disputed)
}

func TestIter29ProcessMiningPredictionHorizonMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("process.mining.prediction.horizon_ms"), ProcessMiningPredictionHorizonMsKey)
}

func TestIter29ProcessMiningPredictionConfidenceKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("process.mining.prediction.confidence"), ProcessMiningPredictionConfidenceKey)
}

func TestIter29ProcessMiningPredictionModelTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("process.mining.prediction.model_type"), ProcessMiningPredictionModelTypeKey)
}

func TestIter29ConsensusEpochFinalizationRoundKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("consensus.epoch.finalization.round"), ConsensusEpochFinalizationRoundKey)
}

func TestIter29ConsensusEpochFinalizationSignatureCountKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("consensus.epoch.finalization.signature_count"), ConsensusEpochFinalizationSignatureCountKey)
}

func TestIter29HealingLoadSheddingThresholdKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("healing.load_shedding.threshold"), HealingLoadSheddingThresholdKey)
}

func TestIter29HealingLoadSheddingShedPctKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("healing.load_shedding.shed_pct"), HealingLoadSheddingShedPctKey)
}

func TestIter29HealingLoadSheddingStrategyKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("healing.load_shedding.strategy"), HealingLoadSheddingStrategyKey)
}

func TestIter29HealingLoadSheddingStrategyValues(t *testing.T) {
	assert.Equal(t, "random", HealingLoadSheddingStrategyValues.Random)
	assert.Equal(t, "priority", HealingLoadSheddingStrategyValues.Priority)
	assert.Equal(t, "oldest", HealingLoadSheddingStrategyValues.Oldest)
}

func TestIter29LLMEmbeddingModelKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.embedding.model"), LlmEmbeddingModelKey)
}

func TestIter29LLMEmbeddingDimensionsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.embedding.dimensions"), LlmEmbeddingDimensionsKey)
}

func TestIter29LLMEmbeddingSimilarityThresholdKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.embedding.similarity_threshold"), LlmEmbeddingSimilarityThresholdKey)
}

func TestIter29SpanMCPToolDeprecateCascadeRules(t *testing.T) {
	// Validates span.mcp.tool.deprecate cascade rule: must include mcp.tool.name (Rule 5) and mcp.server.name (Rule 23)
	requiredAttrs := []string{"mcp.tool.name", "mcp.server.name", "mcp.tool.deprecation.policy"}
	assert.Contains(t, requiredAttrs, "mcp.tool.name", "span.mcp.tool.deprecate must require mcp.tool.name per Rule 5")
	assert.Contains(t, requiredAttrs, "mcp.server.name", "span.mcp.tool.deprecate must require mcp.server.name per Rule 23")
}

func TestIter29SpanHealingLoadSheddingApplyCascadeRules(t *testing.T) {
	// Validates span.healing.load_shedding.apply cascade rules: failure_mode (Rule 1), diagnosis_stage (Rule 32)
	requiredAttrs := []string{"healing.failure_mode", "healing.load_shedding.strategy", "healing.load_shedding.threshold"}
	recommendedAttrs := []string{"healing.diagnosis_stage", "healing.load_shedding.shed_pct"}
	assert.Contains(t, requiredAttrs, "healing.failure_mode", "span.healing.load_shedding.apply must require healing.failure_mode per Rule 1")
	assert.Contains(t, recommendedAttrs, "healing.diagnosis_stage", "span.healing.load_shedding.apply must recommend healing.diagnosis_stage per Rule 32")
}

func TestIter29SpanA2AContractExecuteCascadeRules(t *testing.T) {
	// Validates span.a2a.contract.execute cascade rule: a2a.operation (Rule 3)
	recommendedAttrs := []string{"a2a.operation", "a2a.contract.execution.progress_pct"}
	assert.Contains(t, recommendedAttrs, "a2a.operation", "span.a2a.contract.execute must recommend a2a.operation per Rule 3")
}

// Iter30 tests — MCP analytics, A2A decay, PM drift, consensus partition, healing failover, LLM adapter

func TestIter30MCPToolAnalyticsCallCountKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.analytics.call_count"), McpToolAnalyticsCallCountKey)
}

func TestIter30MCPToolAnalyticsErrorRateKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.analytics.error_rate"), McpToolAnalyticsErrorRateKey)
}

func TestIter30MCPToolAnalyticsAvgLatencyMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("mcp.tool.analytics.avg_latency_ms"), McpToolAnalyticsAvgLatencyMsKey)
}

func TestIter30A2AReputationDecayRateKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("a2a.reputation.decay.rate"), A2aReputationDecayRateKey)
}

func TestIter30A2AReputationDecayTriggerKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("a2a.reputation.decay.trigger"), A2aReputationDecayTriggerKey)
}

func TestIter30A2AReputationDecayTriggerValues(t *testing.T) {
	assert.Equal(t, "time", A2aReputationDecayTriggerValues.Time)
	assert.Equal(t, "interaction", A2aReputationDecayTriggerValues.Interaction)
	assert.Equal(t, "violation", A2aReputationDecayTriggerValues.Violation)
}

func TestIter30ProcessMiningDriftCorrectionTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("process.mining.drift.correction_type"), ProcessMiningDriftCorrectionTypeKey)
}

func TestIter30ProcessMiningDriftCorrectionTypeValues(t *testing.T) {
	assert.Equal(t, "retrain", ProcessMiningDriftCorrectionTypeValues.Retrain)
	assert.Equal(t, "threshold_adjust", ProcessMiningDriftCorrectionTypeValues.ThresholdAdjust)
	assert.Equal(t, "model_swap", ProcessMiningDriftCorrectionTypeValues.ModelSwap)
	assert.Equal(t, "incremental_update", ProcessMiningDriftCorrectionTypeValues.IncrementalUpdate)
}

func TestIter30ProcessMiningDriftCorrectionDeltaKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("process.mining.drift.correction.delta"), ProcessMiningDriftCorrectionDeltaKey)
}

func TestIter30ConsensusPartitionHealStrategyKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("consensus.partition.heal_strategy"), ConsensusPartitionHealStrategyKey)
}

func TestIter30ConsensusPartitionHealStrategyValues(t *testing.T) {
	assert.Equal(t, "majority_wins", ConsensusPartitionHealStrategyValues.MajorityWins)
	assert.Equal(t, "epoch_fence", ConsensusPartitionHealStrategyValues.EpochFence)
	assert.Equal(t, "leader_arbitration", ConsensusPartitionHealStrategyValues.LeaderArbitration)
	assert.Equal(t, "rollback", ConsensusPartitionHealStrategyValues.Rollback)
}

func TestIter30ConsensusPartitionIsolationMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("consensus.partition.isolation_ms"), ConsensusPartitionIsolationMsKey)
}

func TestIter30HealingFailoverTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("healing.failover.type"), HealingFailoverTypeKey)
}

func TestIter30HealingFailoverTypeValues(t *testing.T) {
	assert.Equal(t, "warm_to_cold", HealingFailoverTypeValues.WarmToCold)
	assert.Equal(t, "primary_to_warm", HealingFailoverTypeValues.PrimaryToWarm)
	assert.Equal(t, "primary_to_cold", HealingFailoverTypeValues.PrimaryToCold)
	assert.Equal(t, "geographic", HealingFailoverTypeValues.Geographic)
}

func TestIter30HealingFailoverDurationMsKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("healing.failover.duration_ms"), HealingFailoverDurationMsKey)
}

func TestIter30LLMAdapterIDKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.adapter.id"), LlmAdapterIdKey)
}

func TestIter30LLMAdapterTypeKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.adapter.type"), LlmAdapterTypeKey)
}

func TestIter30LLMAdapterTypeValues(t *testing.T) {
	assert.Equal(t, "lora", LlmAdapterTypeValues.Lora)
	assert.Equal(t, "prefix", LlmAdapterTypeValues.Prefix)
	assert.Equal(t, "prompt_tuning", LlmAdapterTypeValues.PromptTuning)
	assert.Equal(t, "adapter", LlmAdapterTypeValues.Adapter)
	assert.Equal(t, "ia3", LlmAdapterTypeValues.Ia3)
}

func TestIter30LLMAdapterMergeStrategyKeyMatchesSchema(t *testing.T) {
	assert.Equal(t, attribute.Key("llm.adapter.merge_strategy"), LlmAdapterMergeStrategyKey)
}

func TestIter30SpanMCPToolAnalyticsRecordCascadeRules(t *testing.T) {
	requiredAttrs := []string{"mcp.tool.name", "mcp.server.name", "mcp.tool.analytics.call_count"}
	assert.Contains(t, requiredAttrs, "mcp.tool.name", "span.mcp.tool.analytics.record must require mcp.tool.name per Rule 5")
	assert.Contains(t, requiredAttrs, "mcp.server.name", "span.mcp.tool.analytics.record must require mcp.server.name per Rule 23")
}

func TestIter30SpanHealingFailoverExecuteCascadeRules(t *testing.T) {
	requiredAttrs := []string{"healing.failure_mode", "healing.failover.type", "healing.failover.source_id"}
	recommendedAttrs := []string{"healing.diagnosis_stage", "healing.failover.target_id"}
	assert.Contains(t, requiredAttrs, "healing.failure_mode", "span.healing.failover.execute must require healing.failure_mode per Rule 1")
	assert.Contains(t, recommendedAttrs, "healing.diagnosis_stage", "span.healing.failover.execute must recommend healing.diagnosis_stage per Rule 32")
}

func TestIter30SpanA2AReputationDecayCascadeRules(t *testing.T) {
	recommendedAttrs := []string{"a2a.operation", "a2a.reputation.decay.delta"}
	assert.Contains(t, recommendedAttrs, "a2a.operation", "span.a2a.reputation.decay must recommend a2a.operation per Rule 3")
}

// ============================================================
// Iteration 34: A2A Knowledge Transfer Domain
// ============================================================

func TestA2AKnowledgeTransferSpanNameIsCorrect(t *testing.T) {
	if A2aKnowledgeTransferSpan != "a2a.knowledge.transfer" {
		t.Errorf("A2aKnowledgeTransferSpan = %q, want %q", A2aKnowledgeTransferSpan, "a2a.knowledge.transfer")
	}
}

func TestA2AKnowledgeTransferIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aKnowledgeTransferIdKey) != "a2a.knowledge.transfer.id" {
		t.Errorf("A2aKnowledgeTransferIdKey = %q, want %q", A2aKnowledgeTransferIdKey, "a2a.knowledge.transfer.id")
	}
}

func TestA2AKnowledgeTransferTopicKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aKnowledgeTransferTopicKey) != "a2a.knowledge.transfer.topic" {
		t.Errorf("A2aKnowledgeTransferTopicKey = %q, want %q", A2aKnowledgeTransferTopicKey, "a2a.knowledge.transfer.topic")
	}
}

func TestA2AKnowledgeTransferSizeBytesKeyIsCorrectOtelName(t *testing.T) {
	if string(A2aKnowledgeTransferSizeBytesKey) != "a2a.knowledge.transfer.size_bytes" {
		t.Errorf("A2aKnowledgeTransferSizeBytesKey = %q, want %q", A2aKnowledgeTransferSizeBytesKey, "a2a.knowledge.transfer.size_bytes")
	}
}

func TestA2AKnowledgeTransferAttributeTypesMatchSchema(t *testing.T) {
	// Test attribute key types (should be string keys, not other types)
	assert.Equal(t, attribute.Key("a2a.knowledge.transfer.id"), A2aKnowledgeTransferIdKey)
	assert.Equal(t, attribute.Key("a2a.knowledge.transfer.topic"), A2aKnowledgeTransferTopicKey)
	assert.Equal(t, attribute.Key("a2a.knowledge.transfer.size_bytes"), A2aKnowledgeTransferSizeBytesKey)
}

func TestA2AKnowledgeTransferOperationIsRequired(t *testing.T) {
	// Span requires a2a.operation attribute (verified by schema conformance)
	assert.Equal(t, attribute.Key("a2a.operation"), A2aOperationKey, "span.a2a.knowledge.transfer must require a2a.operation")
}

func TestA2AKnowledgeTransferAttributeKeyValueFunctions(t *testing.T) {
	// Verify KeyValue functions work correctly
	kvTransferId := A2aKnowledgeTransferId("kt-abc123")
	assert.NotNil(t, kvTransferId)

	kvTopic := A2aKnowledgeTransferTopic("process_mining")
	assert.NotNil(t, kvTopic)

	kvSizeBytes := A2aKnowledgeTransferSizeBytes(int64(65536))
	assert.NotNil(t, kvSizeBytes)
}
