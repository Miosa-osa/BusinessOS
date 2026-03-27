package semconv

const (
	// BosGatewayDiscoverSpan is the span name for "bos.gateway.discover".
	//
	// Process model discovery via pm4py-rust gateway.
	// Kind: client
	// Stability: development
	BosGatewayDiscoverSpan = "bos.gateway.discover"
	// BosGatewayConformanceSpan is the span name for "bos.gateway.conformance".
	//
	// Conformance check via pm4py-rust gateway.
	// Kind: client
	// Stability: development
	BosGatewayConformanceSpan = "bos.gateway.conformance"
	// BosGatewayStatisticsSpan is the span name for "bos.gateway.statistics".
	//
	// Statistics extraction via pm4py-rust gateway.
	// Kind: client
	// Stability: development
	BosGatewayStatisticsSpan = "bos.gateway.statistics"

	// bos_audit_config_change is the span name for "bos.audit.config_change".
	//
	// Audit record for configuration change operation.
	// Kind: server
	// Stability: development
	BosAuditConfigChangeSpan = "bos.audit.config_change"
	// bos_audit_permission_grant is the span name for "bos.audit.permission_grant".
	//
	// Audit record for permission grant or revocation.
	// Kind: server
	// Stability: development
	BosAuditPermissionGrantSpan = "bos.audit.permission_grant"
	// bos_audit_record is the span name for "bos.audit.record".
	//
	// Recording of a compliance audit trail entry.
	// Kind: internal
	// Stability: development
	BosAuditRecordSpan = "bos.audit.record"
	// bos_compliance_check is the span name for "bos.compliance.check".
	//
	// Evaluation of a single compliance rule against current workspace state.
	// Kind: internal
	// Stability: development
	BosComplianceCheckSpan = "bos.compliance.check"
	// bos_compliance_evaluate is the span name for "bos.compliance.evaluate".
	//
	// Evaluation of a compliance control against current system state.
	// Kind: internal
	// Stability: development
	BosComplianceEvaluateSpan = "bos.compliance.evaluate"
	// bos_decision_record is the span name for "bos.decision.record".
	//
	// Recording of an architectural or operational decision in BusinessOS.
	// Kind: internal
	// Stability: development
	BosDecisionRecordSpan = "bos.decision.record"
	// bos_gap_detect is the span name for "bos.gap.detect".
	//
	// Detection and classification of a compliance gap.
	// Kind: internal
	// Stability: development
	BosGapDetectSpan = "bos.gap.detect"
	// bos_workspace_operation is the span name for "bos.workspace.operation".
	//
	// An operation performed against a BusinessOS workspace (create, update, query).
	// Kind: internal
	// Stability: development
	BosWorkspaceOperationSpan = "bos.workspace.operation"
)