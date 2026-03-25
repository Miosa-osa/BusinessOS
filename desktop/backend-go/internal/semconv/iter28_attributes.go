package semconv

// iter28 — MCP tool composition, A2A reputation, PM enhancement quality,
// consensus quorum shrink, healing cold standby, LLM LoRA

const (
	// MCP Tool Composition
	MCPToolCompositionStrategyKey       = "mcp.tool.composition.strategy"
	MCPToolCompositionStepCountKey      = "mcp.tool.composition.step_count"
	MCPToolCompositionTimeoutMsKey      = "mcp.tool.composition.timeout_ms"
	MCPToolCompositionCompletedStepsKey = "mcp.tool.composition.completed_steps"

	// MCP Tool Composition Strategy Values
	MCPToolCompositionStrategySequential = "sequential"
	MCPToolCompositionStrategyParallel   = "parallel"
	MCPToolCompositionStrategyFallback   = "fallback"
	MCPToolCompositionStrategyPipeline   = "pipeline"

	// A2A Reputation
	A2AReputationScoreKey            = "a2a.reputation.score"
	A2AReputationInteractionCountKey = "a2a.reputation.interaction_count"
	A2AReputationDecayFactorKey      = "a2a.reputation.decay_factor"
	A2AReputationCategoryKey         = "a2a.reputation.category"

	// A2A Reputation Category Values
	A2AReputationCategoryTrusted   = "trusted"
	A2AReputationCategoryNeutral   = "neutral"
	A2AReputationCategoryProbation = "probation"
	A2AReputationCategoryBanned    = "banned"

	// PM Enhancement Quality
	PMEnhancementQualityScoreKey = "process.mining.enhancement.quality_score"
	PMEnhancementCoveragePctKey  = "process.mining.enhancement.coverage_pct"
	PMEnhancementPerspectiveKey  = "process.mining.enhancement.perspective"
	PMEnhancementModelIDKey      = "process.mining.enhancement.model_id"

	// PM Enhancement Perspective Values
	PMEnhancementPerspectivePerformance  = "performance"
	PMEnhancementPerspectiveConformance  = "conformance"
	PMEnhancementPerspectiveOrganization = "organizational"
	PMEnhancementPerspectiveDecision     = "decision"

	// Consensus Quorum Shrink
	ConsensusQuorumShrinkReasonKey       = "consensus.quorum.shrink.reason"
	ConsensusQuorumShrinkRemovedCountKey = "consensus.quorum.shrink.removed_count"
	ConsensusQuorumShrinkNewSizeKey      = "consensus.quorum.shrink.new_size"
	ConsensusQuorumShrinkSafetyMarginKey = "consensus.quorum.shrink.safety_margin"

	// Consensus Quorum Shrink Reason Values
	ConsensusQuorumShrinkReasonNodeFailure  = "node_failure"
	ConsensusQuorumShrinkReasonConfigChange = "config_change"
	ConsensusQuorumShrinkReasonRebalance    = "rebalance"
	ConsensusQuorumShrinkReasonDecommission = "decommission"

	// Healing Cold Standby
	HealingColdStandbyIDKey        = "healing.cold_standby.id"
	HealingColdStandbyWarmupMsKey  = "healing.cold_standby.warmup_ms"
	HealingColdStandbyReadinessKey = "healing.cold_standby.readiness"
	HealingColdStandbyDataLagMsKey = "healing.cold_standby.data_lag_ms"

	// Healing Cold Standby Readiness Values
	HealingColdStandbyReadinessCold    = "cold"
	HealingColdStandbyReadinessWarming = "warming"
	HealingColdStandbyReadinessReady   = "ready"
	HealingColdStandbyReadinessFailed  = "failed"

	// LLM LoRA
	LLMLoRARankKey            = "llm.lora.rank"
	LLMLoRAlphaKey            = "llm.lora.alpha"
	LLMLoRATargetModulesKey   = "llm.lora.target_modules"
	LLMLoRATrainableParamsKey = "llm.lora.trainable_params"
	LLMLoRABaseModelKey       = "llm.lora.base_model"
)
