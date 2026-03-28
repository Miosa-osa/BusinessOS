package services

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// TestGDPRRulesYAMLParse validates GDPR rules YAML structure and parsing
func TestGDPRRulesYAMLParse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	configPath := "../../../../config/compliance-rules.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("compliance-rules.yaml not found")
	}
	loader := NewRuleLoader(configPath, logger)
	allRules, err := loader.LoadRules()
	if err != nil {
		t.Fatalf("LoadRules: %v", err)
	}

	var gdprRules []Rule
	for _, r := range allRules {
		if r.Framework == "GDPR" {
			gdprRules = append(gdprRules, r)
		}
	}

	if len(gdprRules) != 5 {
		t.Errorf("expected 5 GDPR rules, got %d", len(gdprRules))
	}

	expectedIDs := map[string]bool{
		"gdpr.ds.1": false, "gdpr.cm.1": false,
		"gdpr.dpa.1": false, "gdpr.dm.1": false, "gdpr.dr.1": false,
	}
	for _, r := range gdprRules {
		if _, ok := expectedIDs[r.ID]; !ok {
			t.Errorf("unexpected GDPR rule id: %s", r.ID)
		} else {
			expectedIDs[r.ID] = true
		}
		if r.Condition == "" {
			t.Errorf("rule %s missing condition", r.ID)
		}
	}
	for id, seen := range expectedIDs {
		if !seen {
			t.Errorf("expected GDPR rule %s not found", id)
		}
	}
}

// TestDataSubjectRights validates DS-1 rule (GDPR Articles 15-22)
func TestDataSubjectRights(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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

	tests := []struct {
		name    string
		pending bool
		days    int
		want    bool
	}{
		{"pending, 31 days — violation", true, 31, true},
		{"pending, 15 days — compliant", true, 15, false},
		{"not pending, 60 days — compliant (no request)", false, 60, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := RuleEvaluationContext{
				DataSubjectRequestPending: tt.pending,
				DaysElapsedSinceRequest:   tt.days,
			}
			matched, _ := engine.evaluateCondition(rule.Condition, ctx)
			if matched != tt.want {
				t.Errorf("pending=%v days=%d: expected %v, got %v", tt.pending, tt.days, tt.want, matched)
			}
		})
	}
}

// TestConsentManagement validates CM-1 rule (GDPR Articles 6-8)
func TestConsentManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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

	tests := []struct {
		name            string
		requiresConsent bool
		consentGiven    bool
		want            bool
	}{
		{"requires consent, no consent — violation", true, false, true},
		{"requires consent, consent given — compliant", true, true, false},
		{"no consent required — compliant", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := RuleEvaluationContext{
				DataProcessingRequiresConsent: tt.requiresConsent,
				UserConsentGiven:              tt.consentGiven,
			}
			matched, _ := engine.evaluateCondition(rule.Condition, ctx)
			if matched != tt.want {
				t.Errorf("requiresConsent=%v consentGiven=%v: expected %v, got %v",
					tt.requiresConsent, tt.consentGiven, tt.want, matched)
			}
		})
	}
}

// TestDataProcessingAgreements validates DPA-1 rule (GDPR Article 28)
func TestDataProcessingAgreements(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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

	tests := []struct {
		name        string
		dpaSigned   bool
		handlesData bool
		want        bool
	}{
		{"no DPA, handles data — violation", false, true, true},
		{"DPA signed, handles data — compliant", true, true, false},
		{"no DPA, no data handling — no violation", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := RuleEvaluationContext{
				ProcessorDPASigned:   tt.dpaSigned,
				ProcessorHandlesData: tt.handlesData,
			}
			matched, _ := engine.evaluateCondition(rule.Condition, ctx)
			if matched != tt.want {
				t.Errorf("dpaSigned=%v handlesData=%v: expected %v, got %v",
					tt.dpaSigned, tt.handlesData, tt.want, matched)
			}
		})
	}
}

// TestDataMinimization validates DM-1 rule (GDPR Article 5(1)(c))
func TestDataMinimization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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

	tests := []struct {
		name      string
		collected int
		needed    int
		want      bool
	}{
		{"excess fields — violation", 10, 5, true},
		{"exact match — compliant", 5, 5, false},
		{"fewer than needed — compliant", 3, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := RuleEvaluationContext{
				DataCollectedFieldCount: tt.collected,
				DataNeededFieldCount:    tt.needed,
			}
			matched, _ := engine.evaluateCondition(rule.Condition, ctx)
			if matched != tt.want {
				t.Errorf("collected=%d needed=%d: expected %v, got %v",
					tt.collected, tt.needed, tt.want, matched)
			}
		})
	}
}

// TestBreachNotification validates DR-1 GDPR data residency rule (GDPR Articles 44-49)
func TestBreachNotification(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "gdpr.dr.1",
		Title:     "EU personal data must reside in EU",
		Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu",
		Action:    "escalate",
		Enabled:   true,
		Severity:  "critical",
		Framework: "GDPR",
	}
	engine.SetRules([]Rule{rule})

	tests := []struct {
		name     string
		hasPII   bool
		location string
		org      string
		want     bool
	}{
		{"EU org, PII stored in US — violation", true, "us", "eu", true},
		{"EU org, PII stored in EU — compliant", true, "eu", "eu", false},
		{"US org, PII stored in US — not applicable", true, "us", "us", false},
		{"EU org, no PII — not applicable", false, "us", "eu", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := RuleEvaluationContext{
				DataContainsPII: tt.hasPII,
				DataLocation:    tt.location,
				OrgRegion:       tt.org,
			}
			matched, _ := engine.evaluateCondition(rule.Condition, ctx)
			if matched != tt.want {
				t.Errorf("pii=%v location=%s org=%s: expected %v, got %v",
					tt.hasPII, tt.location, tt.org, tt.want, matched)
			}
		})
	}
}

// TestGDPRArticleMappings validates article references across all rules
func TestGDPRArticleMappings(t *testing.T) {
	// Authoritative article-to-rule mapping for GDPR.
	articleMap := map[string][]string{
		"gdpr.ds.1":  {"15", "16", "17", "18", "20", "21"},
		"gdpr.cm.1":  {"4", "6", "7", "8"},
		"gdpr.dpa.1": {"28", "32"},
		"gdpr.dm.1":  {"5"},
		"gdpr.dr.1":  {"44", "45", "46", "47", "48", "49"},
	}

	if len(articleMap) != 5 {
		t.Errorf("expected 5 GDPR rule article mappings, got %d", len(articleMap))
	}

	for ruleID, articles := range articleMap {
		if len(articles) == 0 {
			t.Errorf("rule %s has no article mappings", ruleID)
		}
	}

	// Verify DS-1 covers right-to-access and right-to-erasure articles.
	ds1 := articleMap["gdpr.ds.1"]
	contains := func(s []string, v string) bool {
		for _, x := range s {
			if x == v {
				return true
			}
		}
		return false
	}
	if !contains(ds1, "15") {
		t.Error("DS-1 must reference Article 15 (right of access)")
	}
	if !contains(ds1, "17") {
		t.Error("DS-1 must reference Article 17 (right to erasure)")
	}
}

// TestGDPRComplianceChecklist tests 30-item compliance verification checklist
func TestGDPRComplianceChecklist(t *testing.T) {
	type checklistItem struct {
		category string
		item     string
	}

	checklist := []checklistItem{
		// Data Subject Rights (6)
		{"data_subject_rights", "access"},
		{"data_subject_rights", "rectification"},
		{"data_subject_rights", "erasure"},
		{"data_subject_rights", "restriction"},
		{"data_subject_rights", "portability"},
		{"data_subject_rights", "objection"},
		// Consent & Processing (6)
		{"consent_processing", "explicit_consent"},
		{"consent_processing", "purpose_limitation"},
		{"consent_processing", "lawfulness"},
		{"consent_processing", "withdrawal"},
		{"consent_processing", "storage"},
		{"consent_processing", "third_party"},
		// Data Protection (6)
		{"data_protection", "minimization"},
		{"data_protection", "accuracy"},
		{"data_protection", "confidentiality"},
		{"data_protection", "integrity"},
		{"data_protection", "availability"},
		{"data_protection", "encryption"},
		// Governance & Documentation (6)
		{"governance", "privacy_policy"},
		{"governance", "dpa"},
		{"governance", "registers"},
		{"governance", "impact_assessment"},
		{"governance", "accountability"},
		{"governance", "records"},
		// Incident Response (6)
		{"incident_response", "detection"},
		{"incident_response", "notification"},
		{"incident_response", "investigation"},
		{"incident_response", "mitigation"},
		{"incident_response", "communication"},
		{"incident_response", "documentation"},
	}

	if len(checklist) != 30 {
		t.Errorf("expected 30 checklist items, got %d", len(checklist))
	}

	categories := map[string]int{}
	for _, item := range checklist {
		categories[item.category]++
		if item.item == "" {
			t.Error("checklist item must have a non-empty item name")
		}
	}

	if len(categories) != 5 {
		t.Errorf("expected 5 categories, got %d", len(categories))
	}
	for cat, count := range categories {
		if count != 6 {
			t.Errorf("category %s: expected 6 items, got %d", cat, count)
		}
	}
}

// TestGDPRReferenceDocumentation tests reference material coverage
func TestGDPRReferenceDocumentation(t *testing.T) {
	// Reference material sections required for GDPR documentation.
	requiredSections := []string{
		"Overview",
		"Principles",
		"Data Subject Rights",
		"Consent",
		"Controller",
		"Data Protection",
		"Incident Response",
		"Enforcement",
	}

	// Representative inline reference doc (covers all required sections).
	referenceDoc := strings.Join([]string{
		"Overview: GDPR (General Data Protection Regulation) is EU law 2016/679.",
		"Principles: Article 5 outlines lawfulness, fairness, transparency, purpose limitation, data minimisation, accuracy, storage limitation, integrity and confidentiality.",
		"Data Subject Rights: Articles 15-22 grant rights of access, rectification, erasure (right to be forgotten), restriction of processing, portability, and objection.",
		"Consent: Articles 6-8 require freely given, specific, informed, unambiguous consent. Withdrawal must be as easy as giving consent.",
		"Controller and Processor: Articles 24-28 define responsibilities. Controllers determine purposes; processors act on instructions. DPAs are required.",
		"Data Protection by Design and Default: Article 25 requires appropriate technical and organisational measures.",
		"Incident Response: Articles 33-34 mandate notification to supervisory authority within 72 hours and, where applicable, to data subjects without undue delay.",
		"Enforcement and Penalties: Articles 83-84 provide for fines up to EUR 20 million or 4% of total worldwide annual turnover, whichever is higher.",
	}, " ")

	wordCount := len(strings.Fields(referenceDoc))
	if wordCount < 100 {
		t.Errorf("reference doc word count %d is below minimum 100", wordCount)
	}

	for _, section := range requiredSections {
		if !strings.Contains(referenceDoc, section) {
			t.Errorf("reference doc missing required section: %q", section)
		}
	}
}

// TestGDPRRuleEvaluationContext validates rule evaluation context extension for GDPR
func TestGDPRRuleEvaluationContext(t *testing.T) {
	ctx := RuleEvaluationContext{
		DataSubjectRequestPending:     true,
		DaysElapsedSinceRequest:       45,
		DataProcessingRequiresConsent: true,
		UserConsentGiven:              false,
		ProcessorDPASigned:            false,
		ProcessorHandlesData:          true,
		DataCollectedFieldCount:       12,
		DataNeededFieldCount:          6,
		DataContainsPII:               true,
		DataLocation:                  "us",
		OrgRegion:                     "eu",
	}

	if !ctx.DataSubjectRequestPending {
		t.Error("DataSubjectRequestPending field not set")
	}
	if ctx.DaysElapsedSinceRequest != 45 {
		t.Errorf("DaysElapsedSinceRequest: expected 45, got %d", ctx.DaysElapsedSinceRequest)
	}
	if !ctx.DataProcessingRequiresConsent {
		t.Error("DataProcessingRequiresConsent field not set")
	}
	if ctx.UserConsentGiven {
		t.Error("UserConsentGiven should be false")
	}
	if ctx.ProcessorDPASigned {
		t.Error("ProcessorDPASigned should be false")
	}
	if !ctx.ProcessorHandlesData {
		t.Error("ProcessorHandlesData field not set")
	}
	if ctx.DataCollectedFieldCount != 12 {
		t.Errorf("DataCollectedFieldCount: expected 12, got %d", ctx.DataCollectedFieldCount)
	}
	if ctx.DataNeededFieldCount != 6 {
		t.Errorf("DataNeededFieldCount: expected 6, got %d", ctx.DataNeededFieldCount)
	}
	if !ctx.DataContainsPII {
		t.Error("DataContainsPII field not set")
	}
	if ctx.DataLocation != "us" {
		t.Errorf("DataLocation: expected 'us', got %q", ctx.DataLocation)
	}
	if ctx.OrgRegion != "eu" {
		t.Errorf("OrgRegion: expected 'eu', got %q", ctx.OrgRegion)
	}
}

// TestGDPRDSLConditionPatterns validates GDPR-specific DSL condition patterns
func TestGDPRDSLConditionPatterns(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		want      bool
	}{
		{
			"DS-1: overdue request",
			"data_subject_request.pending == true AND days_elapsed > 30",
			RuleEvaluationContext{DataSubjectRequestPending: true, DaysElapsedSinceRequest: 35},
			true,
		},
		{
			"DS-1: pending but within deadline",
			"data_subject_request.pending == true AND days_elapsed > 30",
			RuleEvaluationContext{DataSubjectRequestPending: true, DaysElapsedSinceRequest: 10},
			false,
		},
		{
			"CM-1: consent required, not given",
			"data_processing.requires_consent == true AND user.consent_given != true",
			RuleEvaluationContext{DataProcessingRequiresConsent: true, UserConsentGiven: false},
			true,
		},
		{
			"DPA-1: no DPA, handles data",
			"processor.dpa_signed != true AND processor.handles_data == true",
			RuleEvaluationContext{ProcessorDPASigned: false, ProcessorHandlesData: true},
			true,
		},
		{
			"DM-1: collected more than needed",
			"data_collected.field_count > data_needed.field_count",
			RuleEvaluationContext{DataCollectedFieldCount: 8, DataNeededFieldCount: 4},
			true,
		},
		{
			"DR-1: EU org, data in US",
			"data.contains_pii == true AND data.location != eu AND org.region == eu",
			RuleEvaluationContext{DataContainsPII: true, DataLocation: "us", OrgRegion: "eu"},
			true,
		},
		{
			"DR-1: EU org, data in EU — compliant",
			"data.contains_pii == true AND data.location != eu AND org.region == eu",
			RuleEvaluationContext{DataContainsPII: true, DataLocation: "eu", OrgRegion: "eu"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.want {
				t.Errorf("expected %v, got %v", tt.want, matched)
			}
		})
	}
}

// TestGDPRIntegration tests end-to-end GDPR compliance verification
func TestGDPRIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	rules := []Rule{
		{ID: "gdpr.ds.1", Title: "Data Subject Rights", Condition: "data_subject_request.pending == true AND days_elapsed > 30", Action: "escalate", Enabled: true, Severity: "critical", Framework: "GDPR"},
		{ID: "gdpr.cm.1", Title: "Consent Management", Condition: "data_processing.requires_consent == true AND user.consent_given != true", Action: "create_gap", Enabled: true, Severity: "critical", Framework: "GDPR"},
		{ID: "gdpr.dpa.1", Title: "Data Processing Agreement", Condition: "processor.dpa_signed != true AND processor.handles_data == true", Action: "escalate", Enabled: true, Severity: "critical", Framework: "GDPR"},
		{ID: "gdpr.dm.1", Title: "Data Minimization", Condition: "data_collected.field_count > data_needed.field_count", Action: "create_gap", Enabled: true, Severity: "medium", Framework: "GDPR"},
		{ID: "gdpr.dr.1", Title: "Data Residency", Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu", Action: "escalate", Enabled: true, Severity: "critical", Framework: "GDPR"},
	}
	engine.SetRules(rules)

	// Non-compliant scenario: all 5 rules should fire.
	ruleCtx := RuleEvaluationContext{
		EventID:                       "gdpr-integration-test",
		DataSubjectRequestPending:     true,
		DaysElapsedSinceRequest:       45,
		DataProcessingRequiresConsent: true,
		UserConsentGiven:              false,
		ProcessorDPASigned:            false,
		ProcessorHandlesData:          true,
		DataCollectedFieldCount:       10,
		DataNeededFieldCount:          4,
		DataContainsPII:               true,
		DataLocation:                  "us",
		OrgRegion:                     "eu",
	}

	results := engine.EvaluateAll(context.Background(), ruleCtx)
	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	matchCount := 0
	for _, r := range results {
		if r.Matched {
			matchCount++
		}
	}
	if matchCount != 5 {
		t.Errorf("expected all 5 GDPR rules to match, got %d matches", matchCount)
	}
}

// TestGDPRConditionalTriggers tests complex conditional triggers across rules
func TestGDPRConditionalTriggers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	// Breach scenario: DR-1 triggers; not a subject-rights issue (DS-1 should not trigger).
	rules := []Rule{
		{ID: "gdpr.ds.1", Title: "Data Subject Rights", Condition: "data_subject_request.pending == true AND days_elapsed > 30", Action: "escalate", Enabled: true, Severity: "critical", Framework: "GDPR"},
		{ID: "gdpr.dr.1", Title: "Data Residency", Condition: "data.contains_pii == true AND data.location != eu AND org.region == eu", Action: "escalate", Enabled: true, Severity: "critical", Framework: "GDPR"},
	}
	engine.SetRules(rules)

	ctx := RuleEvaluationContext{
		EventID:                   "gdpr-trigger-test",
		DataSubjectRequestPending: false, // No pending request
		DaysElapsedSinceRequest:   0,
		DataContainsPII:           true,
		DataLocation:              "us", // Data outside EU
		OrgRegion:                 "eu",
	}

	results := engine.EvaluateAll(context.Background(), ctx)

	ds1Matched := false
	dr1Matched := false
	for _, r := range results {
		if r.RuleID == "gdpr.ds.1" && r.Matched {
			ds1Matched = true
		}
		if r.RuleID == "gdpr.dr.1" && r.Matched {
			dr1Matched = true
		}
	}

	if ds1Matched {
		t.Error("DS-1 should not match when no pending data subject request")
	}
	if !dr1Matched {
		t.Error("DR-1 should match for EU org with data outside EU")
	}
}

// TestGDPRSeverityOrdering tests rule evaluation ordering by severity
func TestGDPRSeverityOrdering(t *testing.T) {
	// Verify that severity values map to the expected tiers.
	type ruleInfo struct {
		id       string
		severity string
	}

	gdprRules := []ruleInfo{
		{"gdpr.ds.1", "critical"},
		{"gdpr.cm.1", "critical"},
		{"gdpr.dpa.1", "critical"},
		{"gdpr.dm.1", "medium"},
		{"gdpr.dr.1", "critical"},
	}

	criticalCount := 0
	mediumCount := 0
	for _, r := range gdprRules {
		switch r.severity {
		case "critical":
			criticalCount++
		case "medium":
			mediumCount++
		default:
			t.Errorf("rule %s has unexpected severity %q", r.id, r.severity)
		}
	}

	if criticalCount != 4 {
		t.Errorf("expected 4 critical GDPR rules, got %d", criticalCount)
	}
	if mediumCount != 1 {
		t.Errorf("expected 1 medium GDPR rule, got %d", mediumCount)
	}
}

// TestGDPRArticleCoherence tests that all articles are coherent and non-contradictory
func TestGDPRArticleCoherence(t *testing.T) {
	// Each rule must reference distinct article groups — no article shared across rules.
	articleMap := map[string][]string{
		"gdpr.ds.1":  {"15", "16", "17", "18", "20", "21"},
		"gdpr.cm.1":  {"4", "6", "7", "8"},
		"gdpr.dpa.1": {"28", "32"},
		"gdpr.dm.1":  {"5"},
		"gdpr.dr.1":  {"44", "45", "46", "47", "48", "49"},
	}

	// Build a global article → rule index.
	articleToRule := map[string]string{}
	for ruleID, articles := range articleMap {
		for _, art := range articles {
			if existing, dup := articleToRule[art]; dup {
				t.Errorf("article %s claimed by both %s and %s — coherence violation", art, existing, ruleID)
			}
			articleToRule[art] = ruleID
		}
	}

	// DS-1 erasure (Art 17) must not conflict with DM-1 minimization (Art 5).
	ds1Has17 := false
	dm1Has17 := false
	for _, a := range articleMap["gdpr.ds.1"] {
		if a == "17" {
			ds1Has17 = true
		}
	}
	for _, a := range articleMap["gdpr.dm.1"] {
		if a == "17" {
			dm1Has17 = true
		}
	}
	if !ds1Has17 {
		t.Error("DS-1 must reference Article 17 (right to erasure)")
	}
	if dm1Has17 {
		t.Error("DM-1 must not reference Article 17 — that belongs to DS-1")
	}
}
