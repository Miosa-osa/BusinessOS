package services

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

// SOX ITG-1: Change Management - Segregation of Duties
func TestSOX_ITG1_ChangeManagement_SegregationOfDuties(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
		desc      string
	}{
		{
			name:      "self-approved change (segregation violation)",
			condition: "change.requires_approval == true AND change.approved_by == change.made_by",
			ctx: RuleEvaluationContext{
				ChangeRequiresApproval: true,
				ChangeApprovedBy:       "alice",
				ChangeMadeBy:           "alice",
			},
			expected: true,
			desc:     "Same person made and approved change - critical violation",
		},
		{
			name:      "properly segregated change (should pass)",
			condition: "change.requires_approval == true AND change.approved_by == change.made_by",
			ctx: RuleEvaluationContext{
				ChangeRequiresApproval: true,
				ChangeApprovedBy:       "bob",
				ChangeMadeBy:           "alice",
			},
			expected: false,
			desc:     "Different people made and approved - proper segregation",
		},
		{
			name:      "non-critical change (no approval required)",
			condition: "change.requires_approval == true AND change.approved_by == change.made_by",
			ctx: RuleEvaluationContext{
				ChangeRequiresApproval: false,
				ChangeApprovedBy:       "alice",
				ChangeMadeBy:           "alice",
			},
			expected: false,
			desc:     "Non-critical change doesn't require approval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, msg := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("%s: expected %v, got %v. Message: %s", tt.desc, tt.expected, matched, msg)
			}
		})
	}
}

// SOX SA-1: System Availability - 99.9% Uptime SLA
func TestSOX_SA1_SystemAvailability_Uptime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
		desc      string
	}{
		{
			name:      "below SLA threshold (99.85%)",
			condition: "system.measured_uptime < 99.9",
			ctx: RuleEvaluationContext{
				SystemMeasuredUptime: 99.85,
			},
			expected: true,
			desc:     "Uptime below 99.9% SLA triggers critical gap",
		},
		{
			name:      "meets SLA threshold (99.95%)",
			condition: "system.measured_uptime < 99.9",
			ctx: RuleEvaluationContext{
				SystemMeasuredUptime: 99.95,
			},
			expected: false,
			desc:     "Uptime above 99.9% SLA passes",
		},
		{
			name:      "perfect uptime (100%)",
			condition: "system.measured_uptime < 99.9",
			ctx: RuleEvaluationContext{
				SystemMeasuredUptime: 100.0,
			},
			expected: false,
			desc:     "Perfect uptime passes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, msg := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("%s: expected %v, got %v. Message: %s", tt.desc, tt.expected, matched, msg)
			}
		})
	}
}

// SOX AL-1: Access Logging - 7-Year Retention
func TestSOX_AL1_AccessLogging_Retention(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
		desc      string
	}{
		{
			name:      "financial data not logged",
			condition: "financial_data.access_logged == false OR audit_log.retention_days < 2555",
			ctx: RuleEvaluationContext{
				FinancialDataAccessLogged: false,
				AuditLogRetentionDays:     2555,
			},
			expected: true,
			desc:     "Unlogged access to financial data triggers gap",
		},
		{
			name:      "insufficient retention (5 years instead of 7)",
			condition: "financial_data.access_logged == false OR audit_log.retention_days < 2555",
			ctx: RuleEvaluationContext{
				FinancialDataAccessLogged: true,
				AuditLogRetentionDays:     1825, // ~5 years
			},
			expected: true,
			desc:     "Retention below 7 years (2555 days) triggers gap",
		},
		{
			name:      "compliant: logged and retained properly",
			condition: "financial_data.access_logged == false OR audit_log.retention_days < 2555",
			ctx: RuleEvaluationContext{
				FinancialDataAccessLogged: true,
				AuditLogRetentionDays:     2555,
			},
			expected: false,
			desc:     "Properly logged and retained access passes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, msg := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("%s: expected %v, got %v. Message: %s", tt.desc, tt.expected, matched, msg)
			}
		})
	}
}

// SOX CM-1: Configuration Management - Production Changes
func TestSOX_CM1_ConfigurationManagement_ProductionChanges(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
		desc      string
	}{
		{
			name:      "undocumented production change",
			condition: "production_change.documented == false",
			ctx: RuleEvaluationContext{
				ProductionChangeDocumented: false,
			},
			expected: true,
			desc:     "Undocumented production change is high severity gap",
		},
		{
			name:      "properly documented change",
			condition: "production_change.documented == false",
			ctx: RuleEvaluationContext{
				ProductionChangeDocumented: true,
			},
			expected: false,
			desc:     "Documented production change passes",
		},
		{
			name:      "documented != false (double negative)",
			condition: "production_change.documented != false",
			ctx: RuleEvaluationContext{
				ProductionChangeDocumented: true,
			},
			expected: true,
			desc:     "Double negative check: documented != false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, msg := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("%s: expected %v, got %v. Message: %s", tt.desc, tt.expected, matched, msg)
			}
		})
	}
}

// SOX FDI-1: Financial Data Integrity - Checksums
func TestSOX_FDI1_FinancialDataIntegrity_Checksums(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
		desc      string
	}{
		{
			name:      "missing checksum",
			condition: "financial_record.has_checksum == false OR checksum.verified == false",
			ctx: RuleEvaluationContext{
				FinancialRecordHasChecksum: false,
				ChecksumVerified:           true,
			},
			expected: true,
			desc:     "Missing checksum triggers critical gap",
		},
		{
			name:      "checksum present but unverified",
			condition: "financial_record.has_checksum == false OR checksum.verified == false",
			ctx: RuleEvaluationContext{
				FinancialRecordHasChecksum: true,
				ChecksumVerified:           false,
			},
			expected: true,
			desc:     "Unverified checksum triggers critical gap",
		},
		{
			name:      "compliant: checksum present and verified",
			condition: "financial_record.has_checksum == false OR checksum.verified == false",
			ctx: RuleEvaluationContext{
				FinancialRecordHasChecksum: true,
				ChecksumVerified:           true,
			},
			expected: false,
			desc:     "Verified checksum passes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, msg := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("%s: expected %v, got %v. Message: %s", tt.desc, tt.expected, matched, msg)
			}
		})
	}
}

// Integration test: All SOX rules evaluated against a scenario
func TestSOX_IntegrationScenario_ComplianceAudit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	// Scenario: Financial system audit in progress
	ctx := RuleEvaluationContext{
		EventID:                    "audit-001",
		SessionID:                  "session-123",
		Timestamp:                  time.Now(),
		Action:                     "audit_check",
		Actor:                      "auditor-team",
		ChangeRequiresApproval:     true,
		ChangeApprovedBy:           "bob",
		ChangeMadeBy:               "alice",
		SystemMeasuredUptime:       99.95,
		FinancialDataAccessLogged:  true,
		AuditLogRetentionDays:      2555,
		ProductionChangeDocumented: true,
		FinancialRecordHasChecksum: true,
		ChecksumVerified:           true,
	}

	// All rules should pass (no violations)
	rules := []Rule{
		{
			ID:        "sox.itg.1",
			Title:     "Change management enforces segregation of duties",
			Condition: "change.requires_approval == true AND change.approved_by == change.made_by",
			Severity:  "critical",
			Framework: "SOX",
			Enabled:   true,
		},
		{
			ID:        "sox.sa.1",
			Title:     "Financial systems must maintain 99.9% uptime",
			Condition: "system.measured_uptime < 99.9",
			Severity:  "critical",
			Framework: "SOX",
			Enabled:   true,
		},
		{
			ID:        "sox.al.1",
			Title:     "All access to financial data must be logged for 7 years",
			Condition: "financial_data.access_logged == false OR audit_log.retention_days < 2555",
			Severity:  "high",
			Framework: "SOX",
			Enabled:   true,
		},
		{
			ID:        "sox.cm.1",
			Title:     "All production changes must be documented and approved",
			Condition: "production_change.documented == false",
			Severity:  "high",
			Framework: "SOX",
			Enabled:   true,
		},
		{
			ID:        "sox.fdi.1",
			Title:     "Financial records must have verified checksums/hashes",
			Condition: "financial_record.has_checksum == false OR checksum.verified == false",
			Severity:  "critical",
			Framework: "SOX",
			Enabled:   true,
		},
	}

	engine.SetRules(rules)
	results := engine.EvaluateAll(context.Background(), ctx)

	// Verify all rules evaluated and none matched
	if len(results) != 5 {
		t.Errorf("Expected 5 rule results, got %d", len(results))
	}

	for _, result := range results {
		if result.Matched {
			t.Errorf("Rule %s should not have matched in compliant scenario: %s", result.RuleID, result.Message)
		}
	}
}

// Integration test: Violation scenario
func TestSOX_IntegrationScenario_ViolationDetection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	// Scenario: Financial system has multiple violations
	ctx := RuleEvaluationContext{
		EventID:                    "audit-002",
		SessionID:                  "session-124",
		Timestamp:                  time.Now(),
		Action:                     "audit_check",
		Actor:                      "auditor-team",
		ChangeRequiresApproval:     true,
		ChangeApprovedBy:           "alice", // VIOLATION: same as who made it
		ChangeMadeBy:               "alice",
		SystemMeasuredUptime:       99.85, // VIOLATION: below 99.9%
		FinancialDataAccessLogged:  false, // VIOLATION: access not logged
		AuditLogRetentionDays:      1825,   // VIOLATION: only 5 years, not 7
		ProductionChangeDocumented: false, // VIOLATION: undocumented
		FinancialRecordHasChecksum: false, // VIOLATION: missing checksum
		ChecksumVerified:           false,
	}

	rules := []Rule{
		{
			ID:        "sox.itg.1",
			Title:     "Change management enforces segregation of duties",
			Condition: "change.requires_approval == true AND change.approved_by == change.made_by",
			Enabled:   true,
			Framework: "SOX",
		},
		{
			ID:        "sox.sa.1",
			Title:     "Financial systems must maintain 99.9% uptime",
			Condition: "system.measured_uptime < 99.9",
			Enabled:   true,
			Framework: "SOX",
		},
		{
			ID:        "sox.al.1",
			Title:     "All access to financial data must be logged for 7 years",
			Condition: "financial_data.access_logged == false OR audit_log.retention_days < 2555",
			Enabled:   true,
			Framework: "SOX",
		},
		{
			ID:        "sox.cm.1",
			Title:     "All production changes must be documented and approved",
			Condition: "production_change.documented == false",
			Enabled:   true,
			Framework: "SOX",
		},
		{
			ID:        "sox.fdi.1",
			Title:     "Financial records must have verified checksums/hashes",
			Condition: "financial_record.has_checksum == false OR checksum.verified == false",
			Enabled:   true,
			Framework: "SOX",
		},
	}

	engine.SetRules(rules)
	results := engine.EvaluateAll(context.Background(), ctx)

	// Verify all 5 rules triggered (matched)
	matchCount := 0
	for _, result := range results {
		if result.Matched {
			matchCount++
		}
	}

	if matchCount != 5 {
		t.Errorf("Expected 5 rule violations, got %d", matchCount)
	}
}
