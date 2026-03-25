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
	if string(HEALING_FAILURE_MODEKey) != "healing.failure_mode" {
		t.Errorf("HEALING_FAILURE_MODEKey = %q, want %q", HEALING_FAILURE_MODEKey, "healing.failure_mode")
	}
}

func TestHealingConfidenceKeyIsCorrectOtelName(t *testing.T) {
	if string(HEALING_CONFIDENCEKey) != "healing.confidence" {
		t.Errorf("HEALING_CONFIDENCEKey = %q, want %q", HEALING_CONFIDENCEKey, "healing.confidence")
	}
}

func TestHealingAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(HEALING_AGENT_IDKey) != "healing.agent_id" {
		t.Errorf("HEALING_AGENT_IDKey = %q, want %q", HEALING_AGENT_IDKey, "healing.agent_id")
	}
}

func TestHealingReflexxArcKeyIsCorrectOtelName(t *testing.T) {
	if string(HEALING_REFLEX_ARCKey) != "healing.reflex_arc" {
		t.Errorf("HEALING_REFLEX_ARCKey = %q, want %q", HEALING_REFLEX_ARCKey, "healing.reflex_arc")
	}
}

func TestHealingRecoveryActionKeyIsCorrectOtelName(t *testing.T) {
	if string(HEALING_RECOVERY_ACTIONKey) != "healing.recovery_action" {
		t.Errorf("HEALING_RECOVERY_ACTIONKey = %q, want %q", HEALING_RECOVERY_ACTIONKey, "healing.recovery_action")
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

// ============================================================
// A2A domain — span.a2a.call + span.a2a.create_deal
// ============================================================

func TestA2AAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2A_AGENT_IDKey) != "a2a.agent.id" {
		t.Errorf("A2A_AGENT_IDKey = %q, want %q", A2A_AGENT_IDKey, "a2a.agent.id")
	}
}

func TestA2ADealIDKeyIsCorrectOtelName(t *testing.T) {
	if string(A2A_DEAL_IDKey) != "a2a.deal.id" {
		t.Errorf("A2A_DEAL_IDKey = %q, want %q", A2A_DEAL_IDKey, "a2a.deal.id")
	}
}

func TestA2AOperationKeyIsCorrectOtelName(t *testing.T) {
	if string(A2A_OPERATIONKey) != "a2a.operation" {
		t.Errorf("A2A_OPERATIONKey = %q, want %q", A2A_OPERATIONKey, "a2a.operation")
	}
}

func TestA2ASourceServiceKeyIsCorrectOtelName(t *testing.T) {
	if string(A2A_SOURCE_SERVICEKey) != "a2a.source.service" {
		t.Errorf("A2A_SOURCE_SERVICEKey = %q, want %q", A2A_SOURCE_SERVICEKey, "a2a.source.service")
	}
}

// ============================================================
// BusinessOS domain — span.bos.compliance.check (new!)
// ============================================================

func TestBosComplianceFrameworkKeyIsCorrectOtelName(t *testing.T) {
	if string(BOS_COMPLIANCE_FRAMEWORKKey) != "bos.compliance.framework" {
		t.Errorf("BOS_COMPLIANCE_FRAMEWORKKey = %q, want %q", BOS_COMPLIANCE_FRAMEWORKKey, "bos.compliance.framework")
	}
}

func TestBosComplianceRuleIDKeyIsCorrectOtelName(t *testing.T) {
	if string(BOS_COMPLIANCE_RULE_IDKey) != "bos.compliance.rule_id" {
		t.Errorf("BOS_COMPLIANCE_RULE_IDKey = %q, want %q", BOS_COMPLIANCE_RULE_IDKey, "bos.compliance.rule_id")
	}
}

func TestBosCompliancePassedKeyIsCorrectOtelName(t *testing.T) {
	if string(BOS_COMPLIANCE_PASSEDKey) != "bos.compliance.passed" {
		t.Errorf("BOS_COMPLIANCE_PASSEDKey = %q, want %q", BOS_COMPLIANCE_PASSEDKey, "bos.compliance.passed")
	}
}

func TestBosComplianceSeverityKeyIsCorrectOtelName(t *testing.T) {
	if string(BOS_COMPLIANCE_SEVERITYKey) != "bos.compliance.severity" {
		t.Errorf("BOS_COMPLIANCE_SEVERITYKey = %q, want %q", BOS_COMPLIANCE_SEVERITYKey, "bos.compliance.severity")
	}
}

func TestBosDecisionTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(BOS_DECISION_TYPEKey) != "bos.decision.type" {
		t.Errorf("BOS_DECISION_TYPEKey = %q, want %q", BOS_DECISION_TYPEKey, "bos.decision.type")
	}
}

func TestBosComplianceSeverityCriticalValueMatchesSchema(t *testing.T) {
	if BosComplianceSeverityValues.Critical != "critical" {
		t.Errorf("BosComplianceSeverityValues.Critical = %q, want %q", BosComplianceSeverityValues.Critical, "critical")
	}
}

func TestBosDecisionTypeArchitecturalValueMatchesSchema(t *testing.T) {
	if BosDecisionTypeValues.Architectural != "architectural" {
		t.Errorf("BosDecisionTypeValues.Architectural = %q, want %q", BosDecisionTypeValues.Architectural, "architectural")
	}
}

func TestBosComplianceFrameworkSOC2ValueMatchesSchema(t *testing.T) {
	if BosComplianceFrameworkValues.Soc2 != "SOC2" {
		t.Errorf("BosComplianceFrameworkValues.Soc2 = %q, want %q", BosComplianceFrameworkValues.Soc2, "SOC2")
	}
}

// ============================================================
// Workflow domain — span.workflow.execute (new YAWL patterns)
// ============================================================

func TestWorkflowIDKeyIsCorrectOtelName(t *testing.T) {
	if string(WORKFLOW_IDKey) != "workflow.id" {
		t.Errorf("WORKFLOW_IDKey = %q, want %q", WORKFLOW_IDKey, "workflow.id")
	}
}

func TestWorkflowNameKeyIsCorrectOtelName(t *testing.T) {
	if string(WORKFLOW_NAMEKey) != "workflow.name" {
		t.Errorf("WORKFLOW_NAMEKey = %q, want %q", WORKFLOW_NAMEKey, "workflow.name")
	}
}

func TestWorkflowPatternKeyIsCorrectOtelName(t *testing.T) {
	if string(WORKFLOW_PATTERNKey) != "workflow.pattern" {
		t.Errorf("WORKFLOW_PATTERNKey = %q, want %q", WORKFLOW_PATTERNKey, "workflow.pattern")
	}
}

func TestWorkflowStateKeyIsCorrectOtelName(t *testing.T) {
	if string(WORKFLOW_STATEKey) != "workflow.state" {
		t.Errorf("WORKFLOW_STATEKey = %q, want %q", WORKFLOW_STATEKey, "workflow.state")
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

// ============================================================
// Consensus domain — span.consensus.round (HotStuff BFT)
// ============================================================

func TestConsensusRoundNumKeyIsCorrectOtelName(t *testing.T) {
	if string(CONSENSUS_ROUND_NUMKey) != "consensus.round_num" {
		t.Errorf("CONSENSUS_ROUND_NUMKey = %q, want %q", CONSENSUS_ROUND_NUMKey, "consensus.round_num")
	}
}

func TestConsensusRoundTypeKeyIsCorrectOtelName(t *testing.T) {
	if string(CONSENSUS_ROUND_TYPEKey) != "consensus.round_type" {
		t.Errorf("CONSENSUS_ROUND_TYPEKey = %q, want %q", CONSENSUS_ROUND_TYPEKey, "consensus.round_type")
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

// ============================================================
// MCP domain
// ============================================================

func TestMCPToolNameKeyIsCorrectOtelName(t *testing.T) {
	if string(MCP_TOOL_NAMEKey) != "mcp.tool.name" {
		t.Errorf("MCP_TOOL_NAMEKey = %q, want %q", MCP_TOOL_NAMEKey, "mcp.tool.name")
	}
}

func TestMCPProtocolStdioValueMatchesSchema(t *testing.T) {
	if McpProtocolValues.Stdio != "stdio" {
		t.Errorf("McpProtocolValues.Stdio = %q, want %q", McpProtocolValues.Stdio, "stdio")
	}
}

// ============================================================
// Agent domain
// ============================================================

func TestAgentIDKeyIsCorrectOtelName(t *testing.T) {
	if string(AGENT_IDKey) != "agent.id" {
		t.Errorf("AGENT_IDKey = %q, want %q", AGENT_IDKey, "agent.id")
	}
}

func TestAgentOutcomeSuccessValueMatchesSchema(t *testing.T) {
	if AgentOutcomeValues.Success != "success" {
		t.Errorf("AgentOutcomeValues.Success = %q, want %q", AgentOutcomeValues.Success, "success")
	}
}
