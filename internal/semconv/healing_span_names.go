package semconv

const (
	// healing_adaptive_adjust is the span name for "healing.adaptive.adjust".
	//
	// Adaptive threshold adjustment — updates the healing detection threshold based on observed system behavior.
	// Kind: internal
	// Stability: development
	HealingAdaptiveAdjustSpan = "healing.adaptive.adjust"
	// healing_anomaly_detect is the span name for "healing.anomaly.detect".
	//
	// Anomaly detection scan — identifies abnormal system behavior patterns for healing intervention.
	// Kind: internal
	// Stability: development
	HealingAnomalyDetectSpan = "healing.anomaly.detect"
	// healing_backpressure_apply is the span name for "healing.backpressure.apply".
	//
	// Backpressure application — managing healing request flow under system overload.
	// Kind: internal
	// Stability: development
	HealingBackpressureApplySpan = "healing.backpressure.apply"
	// healing_cascade_detect is the span name for "healing.cascade.detect".
	//
	// Detecting cascade failure pattern — identifying correlated failures and root cause.
	// Kind: internal
	// Stability: development
	HealingCascadeDetectSpan = "healing.cascade.detect"
	// healing_checkpoint_create is the span name for "healing.checkpoint.create".
	//
	// Healing checkpoint creation — capturing system state as a recovery checkpoint before risky operations.
	// Kind: internal
	// Stability: development
	HealingCheckpointCreateSpan = "healing.checkpoint.create"
	// healing_circuit_breaker_trip is the span name for "healing.circuit_breaker.trip".
	//
	// Circuit breaker state transition — healing subsystem trips open to prevent cascade failures.
	// Kind: internal
	// Stability: development
	HealingCircuitBreakerTripSpan = "healing.circuit_breaker.trip"
	// healing_cold_standby_promote is the span name for "healing.cold_standby.promote".
	//
	// Cold standby promotion — warming up and promoting a cold replica to primary during a healing failover.
	// Kind: internal
	// Stability: development
	HealingColdStandbyPromoteSpan = "healing.cold_standby.promote"
	// healing_diagnosis is the span name for "healing.diagnosis".
	//
	// Classifies a system failure into a known failure mode with a confidence score.
	// Kind: internal
	// Stability: development
	HealingDiagnosisSpan = "healing.diagnosis"
	// healing_escalation is the span name for "healing.escalation".
	//
	// Escalation to human operator when healing max attempts exceeded.
	// Kind: internal
	// Stability: development
	HealingEscalationSpan = "healing.escalation"
	// healing_failover_execute is the span name for "healing.failover.execute".
	//
	// Healing failover execution — transitioning service from a failing component to a standby replacement.
	// Kind: internal
	// Stability: development
	HealingFailoverExecuteSpan = "healing.failover.execute"
	// healing_fingerprint is the span name for "healing.fingerprint".
	//
	// Process fingerprinting — computes a failure signature for pattern matching.
	// Kind: internal
	// Stability: development
	HealingFingerprintSpan = "healing.fingerprint"
	// healing_intervention_score is the span name for "healing.intervention.score".
	//
	// Healing intervention scoring — evaluates the effectiveness of a completed healing intervention.
	// Kind: internal
	// Stability: development
	HealingInterventionScoreSpan = "healing.intervention.score"
	// healing_load_shedding_apply is the span name for "healing.load_shedding.apply".
	//
	// Load shedding application — intentionally dropping requests to protect the system under overload conditions.
	// Kind: internal
	// Stability: development
	HealingLoadSheddingApplySpan = "healing.load_shedding.apply"
	// healing_memory_snapshot is the span name for "healing.memory.snapshot".
	//
	// Memory snapshot — capturing the current system state to enable fast recovery during healing.
	// Kind: internal
	// Stability: development
	HealingMemorySnapshotSpan = "healing.memory.snapshot"
	// healing_mttr_measure is the span name for "healing.mttr.measure".
	//
	// Measuring MTTR for a completed healing cycle — from failure detection to full recovery.
	// Kind: internal
	// Stability: development
	HealingMttrMeasureSpan = "healing.mttr.measure"
	// healing_pattern_match is the span name for "healing.pattern.match".
	//
	// Matching a failure signature against the healing pattern library to identify recovery action.
	// Kind: internal
	// Stability: development
	HealingPatternMatchSpan = "healing.pattern.match"
	// healing_playbook_execute is the span name for "healing.playbook.execute".
	//
	// Execution of a healing recovery playbook — structured series of remediation steps.
	// Kind: internal
	// Stability: development
	HealingPlaybookExecuteSpan = "healing.playbook.execute"
	// healing_prediction_make is the span name for "healing.prediction.make".
	//
	// Predictive healing — forecasts failure probability within a time horizon using ML model.
	// Kind: internal
	// Stability: development
	HealingPredictionMakeSpan = "healing.prediction.make"
	// healing_quarantine_apply is the span name for "healing.quarantine.apply".
	//
	// Quarantine application — isolating a component to prevent cascade failures during healing.
	// Kind: internal
	// Stability: development
	HealingQuarantineApplySpan = "healing.quarantine.apply"
	// healing_rate_limit_enforce is the span name for "healing.rate_limit.enforce".
	//
	// Rate limit enforcement — throttling healing attempts to prevent cascade recovery storms.
	// Kind: internal
	// Stability: development
	HealingRateLimitEnforceSpan = "healing.rate_limit.enforce"
	// healing_recovery_simulate is the span name for "healing.recovery.simulate".
	//
	// Recovery simulation — running synthetic failure scenarios to validate healing playbooks and reflex arcs.
	// Kind: internal
	// Stability: development
	HealingRecoverySimulateSpan = "healing.recovery.simulate"
	// healing_recovery_loop is the span name for "healing.recovery_loop".
	//
	// Bounded recovery loop execution — WvdA liveness-bounded healing iteration.
	// Kind: internal
	// Stability: development
	HealingRecoveryLoopSpan = "healing.recovery_loop"
	// healing_reflex_arc is the span name for "healing.reflex_arc".
	//
	// Execution of a healing reflex arc — automated recovery action triggered by a detected failure pattern.
	// Kind: internal
	// Stability: development
	HealingReflexArcSpan = "healing.reflex_arc"
	// healing_retry_adaptive is the span name for "healing.retry.adaptive".
	//
	// Adaptive retry backoff execution — applying dynamic retry strategy during healing.
	// Kind: internal
	// Stability: development
	HealingRetryAdaptiveSpan = "healing.retry.adaptive"
	// healing_rollback_execute is the span name for "healing.rollback.execute".
	//
	// Rollback execution — reverting the system to a known-good checkpoint or snapshot after a healing failure.
	// Kind: internal
	// Stability: development
	HealingRollbackExecuteSpan = "healing.rollback.execute"
	// healing_self_healing_trigger is the span name for "healing.self_healing.trigger".
	//
	// Triggering an autonomous self-healing action in response to a detected failure.
	// Kind: internal
	// Stability: development
	HealingSelfHealingTriggerSpan = "healing.self_healing.trigger"
	// healing_surge_detect is the span name for "healing.surge.detect".
	//
	// Detecting a healing surge and applying mitigation strategy.
	// Kind: internal
	// Stability: development
	HealingSurgeDetectSpan = "healing.surge.detect"
	// healing_warm_standby_activate is the span name for "healing.warm_standby.activate".
	//
	// Warm standby activation — promoting a warm replica to primary during a healing failover event.
	// Kind: internal
	// Stability: development
	HealingWarmStandbyActivateSpan = "healing.warm_standby.activate"
)