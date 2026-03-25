package services

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GDPR DS-1: Data Subject Rights Tests
// ---------------------------------------------------------------------------

func TestGDPR_DS1_DataSubjectRightsPending_NotExceeded(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.ds.1",
		Title:     "Data subject access requests must be fulfilled",
		Condition: "data_subject_request.pending == true AND days_elapsed > 30",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Request is pending but only 15 days elapsed - should NOT match
	ctx := RuleEvaluationContext{
		EventID:                   "evt-001",
		DataSubjectRequestPending: true,
		DaysElapsedSinceRequest:   15,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when days_elapsed <= 30")
}

func TestGDPR_DS1_DataSubjectRightsOverdue(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.ds.1",
		Title:     "Data subject access requests must be fulfilled",
		Condition: "data_subject_request.pending == true AND days_elapsed > 30",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Request is pending and 35 days elapsed - SHOULD match
	ctx := RuleEvaluationContext{
		EventID:                   "evt-002",
		DataSubjectRequestPending: true,
		DaysElapsedSinceRequest:   35,
	}

	result := engine.evaluate(rule, ctx)
	assert.True(t, result.Matched, "rule should match when pending=true AND days_elapsed > 30")
	assert.Equal(t, "escalate", result.Action)
}

func TestGDPR_DS1_NoPendingRequest(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.ds.1",
		Title:     "Data subject access requests must be fulfilled",
		Condition: "data_subject_request.pending == true AND days_elapsed > 30",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// No pending request - should NOT match even if days_elapsed > 30
	ctx := RuleEvaluationContext{
		EventID:                   "evt-003",
		DataSubjectRequestPending: false,
		DaysElapsedSinceRequest:   45,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when pending=false")
}

// ---------------------------------------------------------------------------
// GDPR CM-1: Consent Management Tests
// ---------------------------------------------------------------------------

func TestGDPR_CM1_ConsentRequired_NoConsent(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.cm.1",
		Title:     "Explicit consent required for personal data processing",
		Condition: "data_processing.requires_consent == true AND user.consent_given != true",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Processing requires consent but user hasn't given it - SHOULD match
	ctx := RuleEvaluationContext{
		EventID:                       "evt-004",
		DataProcessingRequiresConsent: true,
		UserConsentGiven:              false,
	}

	result := engine.evaluate(rule, ctx)
	assert.True(t, result.Matched, "rule should match when consent required but not given")
	assert.Equal(t, "create_gap", result.Action)
}

func TestGDPR_CM1_ConsentRequired_WithConsent(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.cm.1",
		Title:     "Explicit consent required for personal data processing",
		Condition: "data_processing.requires_consent == true AND user.consent_given != true",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Processing requires consent and user has given it - should NOT match
	ctx := RuleEvaluationContext{
		EventID:                       "evt-005",
		DataProcessingRequiresConsent: true,
		UserConsentGiven:              true,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when consent required and given")
}

func TestGDPR_CM1_ConsentNotRequired(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.cm.1",
		Title:     "Explicit consent required for personal data processing",
		Condition: "data_processing.requires_consent == true AND user.consent_given != true",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Processing doesn't require consent - should NOT match
	ctx := RuleEvaluationContext{
		EventID:                       "evt-006",
		DataProcessingRequiresConsent: false,
		UserConsentGiven:              false,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when consent not required")
}

// ---------------------------------------------------------------------------
// GDPR DPA-1: Data Processing Agreement Tests
// ---------------------------------------------------------------------------

func TestGDPR_DPA1_NoDPA_WithDataHandling(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dpa.1",
		Title:     "Data Processing Agreement with all sub-processors",
		Condition: "processor.dpa_signed != true AND processor.handles_data == true",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Processor handles data but no DPA signed - SHOULD match
	ctx := RuleEvaluationContext{
		EventID:              "evt-007",
		ProcessorDPASigned:   false,
		ProcessorHandlesData: true,
	}

	result := engine.evaluate(rule, ctx)
	assert.True(t, result.Matched, "rule should match when processor handles data but no DPA")
	assert.Equal(t, "escalate", result.Action)
}

func TestGDPR_DPA1_DPASigned(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dpa.1",
		Title:     "Data Processing Agreement with all sub-processors",
		Condition: "processor.dpa_signed != true AND processor.handles_data == true",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// DPA is signed - should NOT match
	ctx := RuleEvaluationContext{
		EventID:              "evt-008",
		ProcessorDPASigned:   true,
		ProcessorHandlesData: true,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when DPA is signed")
}

func TestGDPR_DPA1_NoDataHandling(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dpa.1",
		Title:     "Data Processing Agreement with all sub-processors",
		Condition: "processor.dpa_signed != true AND processor.handles_data == true",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Processor doesn't handle data - should NOT match even without DPA
	ctx := RuleEvaluationContext{
		EventID:              "evt-009",
		ProcessorDPASigned:   false,
		ProcessorHandlesData: false,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when processor doesn't handle data")
}

// ---------------------------------------------------------------------------
// GDPR DM-1: Data Minimization Tests
// ---------------------------------------------------------------------------

func TestGDPR_DM1_ExcessiveDataCollection(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dm.1",
		Title:     "Only collect personal data actually needed",
		Condition: "data_collected.field_count > data_needed.field_count",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "medium",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Collecting 15 fields but only need 8 - SHOULD match
	ctx := RuleEvaluationContext{
		EventID:                 "evt-010",
		DataCollectedFieldCount: 15,
		DataNeededFieldCount:    8,
	}

	result := engine.evaluate(rule, ctx)
	assert.True(t, result.Matched, "rule should match when data_collected > data_needed")
	assert.Equal(t, "create_gap", result.Action)
}

func TestGDPR_DM1_AppropriateDataCollection(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dm.1",
		Title:     "Only collect personal data actually needed",
		Condition: "data_collected.field_count > data_needed.field_count",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "medium",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Collecting 8 fields and need 8 - should NOT match
	ctx := RuleEvaluationContext{
		EventID:                 "evt-011",
		DataCollectedFieldCount: 8,
		DataNeededFieldCount:    8,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when data_collected <= data_needed")
}

func TestGDPR_DM1_LessDataCollected(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dm.1",
		Title:     "Only collect personal data actually needed",
		Condition: "data_collected.field_count > data_needed.field_count",
		Action:    "create_gap",
		Enabled:   true,
		Severity:  "medium",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// Collecting 5 fields but need 10 - should NOT match
	ctx := RuleEvaluationContext{
		EventID:                 "evt-012",
		DataCollectedFieldCount: 5,
		DataNeededFieldCount:    10,
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when data_collected < data_needed")
}

// ---------------------------------------------------------------------------
// GDPR DR-1: Data Residency Tests
// ---------------------------------------------------------------------------

func TestGDPR_DR1_EUDataOutsideEU(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dr.1",
		Title:     "EU personal data must reside in EU (for EU-based organizations)",
		Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// EU org storing EU resident data in US - SHOULD match (CRITICAL)
	ctx := RuleEvaluationContext{
		EventID:         "evt-013",
		DataContainsPII: true,
		DataLocation:    "us",
		OrgRegion:       "eu",
	}

	result := engine.evaluate(rule, ctx)
	assert.True(t, result.Matched, "rule should match when EU org stores PII outside EU")
	assert.Equal(t, "escalate", result.Action)
}

func TestGDPR_DR1_EUDataInEU(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dr.1",
		Title:     "EU personal data must reside in EU (for EU-based organizations)",
		Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// EU org storing EU resident data in EU - should NOT match
	ctx := RuleEvaluationContext{
		EventID:         "evt-014",
		DataContainsPII: true,
		DataLocation:    "eu",
		OrgRegion:       "eu",
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when data is in EU")
}

func TestGDPR_DR1_USDataOutsideEU(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dr.1",
		Title:     "EU personal data must reside in EU (for EU-based organizations)",
		Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// EU org storing non-PII data - should NOT match
	ctx := RuleEvaluationContext{
		EventID:         "evt-015",
		DataContainsPII: false,
		DataLocation:    "us",
		OrgRegion:       "eu",
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when data doesn't contain PII")
}

func TestGDPR_DR1_USorgOutsideEU(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dr.1",
		Title:     "EU personal data must reside in EU (for EU-based organizations)",
		Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	// US org storing PII outside EU - should NOT match (org is not EU)
	ctx := RuleEvaluationContext{
		EventID:         "evt-016",
		DataContainsPII: true,
		DataLocation:    "us",
		OrgRegion:       "us",
	}

	result := engine.evaluate(rule, ctx)
	assert.False(t, result.Matched, "rule should not match when org is not EU-based")
}

// ---------------------------------------------------------------------------
// Full GDPR Test Suite with Multiple Rules
// ---------------------------------------------------------------------------

func TestGDPR_AllRulesLoad(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	gdprRules := []Rule{
		{
			ID:        "gdpr.ds.1",
			Title:     "Data subject access requests must be fulfilled",
			Condition: "data_subject_request.pending == true AND days_elapsed > 30",
			Action:    "escalate",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
		{
			ID:        "gdpr.cm.1",
			Title:     "Explicit consent required for personal data processing",
			Condition: "data_processing.requires_consent == true AND user.consent_given != true",
			Action:    "create_gap",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
		{
			ID:        "gdpr.dpa.1",
			Title:     "Data Processing Agreement with all sub-processors",
			Condition: "processor.dpa_signed != true AND processor.handles_data == true",
			Action:    "escalate",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
		{
			ID:        "gdpr.dm.1",
			Title:     "Only collect personal data actually needed",
			Condition: "data_collected.field_count > data_needed.field_count",
			Action:    "create_gap",
			Enabled:   true,
			Severity:  "medium",
			Framework: "GDPR",
		},
		{
			ID:        "gdpr.dr.1",
			Title:     "EU personal data must reside in EU (for EU-based organizations)",
			Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
			Action:    "escalate",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
	}

	engine.SetRules(gdprRules)

	// Verify all rules were loaded
	loaded := engine.GetRules()
	assert.Len(t, loaded, 5, "all 5 GDPR rules should load")

	// Verify each rule loaded with correct metadata
	for _, rule := range loaded {
		assert.NotEmpty(t, rule.ID)
		assert.NotEmpty(t, rule.Title)
		assert.NotEmpty(t, rule.Condition)
		assert.NotEmpty(t, rule.Action)
		assert.True(t, rule.Enabled)
		assert.Equal(t, "GDPR", rule.Framework)
	}
}

func TestGDPR_CompoundEvaluation(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rules := []Rule{
		{
			ID:        "gdpr.ds.1",
			Title:     "Data subject access requests must be fulfilled",
			Condition: "data_subject_request.pending == true AND days_elapsed > 30",
			Action:    "escalate",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
		{
			ID:        "gdpr.dpa.1",
			Title:     "Data Processing Agreement with all sub-processors",
			Condition: "processor.dpa_signed != true AND processor.handles_data == true",
			Action:    "escalate",
			Enabled:   true,
			Severity:  "critical",
			Framework: "GDPR",
		},
	}

	engine.SetRules(rules)

	// Scenario: both DS-1 and DPA-1 violations
	ctx := RuleEvaluationContext{
		EventID:                   "evt-017",
		DataSubjectRequestPending: true,
		DaysElapsedSinceRequest:   45,
		ProcessorDPASigned:        false,
		ProcessorHandlesData:      true,
	}

	results := engine.EvaluateAll(context.Background(), ctx)
	require.Len(t, results, 2, "should evaluate 2 rules")

	// Both should match
	ds1Match := false
	dpa1Match := false
	for _, r := range results {
		if r.RuleID == "gdpr.ds.1" {
			ds1Match = r.Matched
		}
		if r.RuleID == "gdpr.dpa.1" {
			dpa1Match = r.Matched
		}
	}

	assert.True(t, ds1Match, "DS-1 rule should match")
	assert.True(t, dpa1Match, "DPA-1 rule should match")
}

func TestGDPR_DisabledRuleNotEvaluated(t *testing.T) {
	logger := slog.Default()
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.cm.1",
		Title:     "Explicit consent required for personal data processing",
		Condition: "data_processing.requires_consent == true AND user.consent_given != true",
		Action:    "create_gap",
		Enabled:   false, // DISABLED
		Severity:  "critical",
		Framework: "GDPR",
	}

	engine.SetRules([]Rule{rule})

	// Even though condition is true, rule should NOT evaluate because enabled=false
	ctx := RuleEvaluationContext{
		EventID:                       "evt-018",
		DataProcessingRequiresConsent: true,
		UserConsentGiven:              false,
	}

	results := engine.EvaluateAll(context.Background(), ctx)
	assert.Len(t, results, 0, "disabled rule should not be evaluated")
}
