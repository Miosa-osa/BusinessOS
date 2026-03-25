package services

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestComplianceService_EvaluateAuditEvent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	// Set up test rules
	testRules := []Rule{
		{
			ID:        "test.rule1",
			Title:     "User Role Check",
			Condition: "user.role != admin",
			Action:    "audit",
			Enabled:   true,
			Severity:  "high",
			Framework: "SOC2",
		},
	}

	svc.ruleEngine.SetRules(testRules)

	ctx := context.Background()

	// Create test audit entry
	entry := AuditEntry{
		ID:        "aud1",
		SessionID: "sess1",
		Timestamp: time.Now(),
		Action:    "user_login",
		Actor:     "user@example.com",
	}

	// Evaluate audit event
	err := svc.EvaluateAuditEvent(ctx, entry, "user")
	if err != nil {
		t.Fatalf("EvaluateAuditEvent failed: %v", err)
	}
}

func TestComplianceService_GetRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	testRules := []Rule{
		{
			ID:        "test.rule1",
			Title:     "Test Rule",
			Condition: "user.role != admin",
			Action:    "audit",
			Enabled:   true,
		},
		{
			ID:        "test.rule2",
			Title:     "Test Rule 2",
			Condition: "data.encrypted == true",
			Action:    "audit",
			Enabled:   true,
		},
	}

	svc.ruleEngine.SetRules(testRules)

	rules := svc.GetRules()
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}

	if rules[0].ID != "test.rule1" {
		t.Errorf("expected rule ID test.rule1, got %s", rules[0].ID)
	}
}

func TestComplianceService_AddComplianceGap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	ctx := context.Background()

	gap := ComplianceGap{
		ID:          "gap1",
		Framework:   "SOC2",
		Control:     "CC6.1",
		Description: "Test gap",
		Severity:    "critical",
		Status:      "open",
	}

	err := svc.addComplianceGap(ctx, gap)
	if err != nil {
		t.Fatalf("addComplianceGap failed: %v", err)
	}

	// Verify gap was added
	gaps := svc.gaps["SOC2"]
	if len(gaps) != 1 {
		t.Errorf("expected 1 gap, got %d", len(gaps))
	}

	if gaps[0].ID != "gap1" {
		t.Errorf("expected gap ID gap1, got %s", gaps[0].ID)
	}
}

func TestComplianceService_RuleEngineIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	// Track actions
	gapsCreated := 0
	notificationsSent := 0

	svc.ruleEngine.SetGapHandler(func(ctx context.Context, gap ComplianceGap) error {
		gapsCreated++
		return nil
	})

	svc.ruleEngine.SetNotifyHandler(func(ctx context.Context, ruleID, message string) error {
		notificationsSent++
		return nil
	})

	// Set up rules with different actions
	testRules := []Rule{
		{
			ID:        "create_gap_rule",
			Title:     "Create Gap Rule",
			Condition: "user.role != admin",
			Action:    "create_gap",
			Enabled:   true,
			Framework: "SOC2",
		},
		{
			ID:        "notify_rule",
			Title:     "Notify Rule",
			Condition: "user.role != admin",
			Action:    "notify",
			Enabled:   true,
		},
	}

	svc.ruleEngine.SetRules(testRules)

	ctx := context.Background()

	// Create audit entry that triggers both rules
	entry := AuditEntry{
		ID:        "aud1",
		SessionID: "sess1",
		Timestamp: time.Now(),
		Action:    "user_login",
		Actor:     "user@example.com",
	}

	// Evaluate - this should trigger both create_gap and notify actions
	err := svc.EvaluateAuditEvent(ctx, entry, "user")
	if err != nil {
		t.Fatalf("EvaluateAuditEvent failed: %v", err)
	}

	// Verify actions were taken
	if gapsCreated == 0 {
		t.Errorf("expected gap to be created, but none were created")
	}

	if notificationsSent == 0 {
		t.Errorf("expected notification to be sent, but none were sent")
	}
}

func TestComplianceService_RuleFiltering_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	// Track actions
	actionsExecuted := 0

	svc.ruleEngine.SetGapHandler(func(ctx context.Context, gap ComplianceGap) error {
		actionsExecuted++
		return nil
	})

	// Set up one enabled and one disabled rule with same condition
	testRules := []Rule{
		{
			ID:        "enabled_rule",
			Title:     "Enabled Rule",
			Condition: "user.role != admin",
			Action:    "create_gap",
			Enabled:   true,
			Framework: "SOC2",
		},
		{
			ID:        "disabled_rule",
			Title:     "Disabled Rule",
			Condition: "user.role != admin",
			Action:    "create_gap",
			Enabled:   false,
			Framework: "SOC2",
		},
	}

	svc.ruleEngine.SetRules(testRules)

	ctx := context.Background()
	entry := AuditEntry{
		ID:        "aud1",
		SessionID: "sess1",
		Timestamp: time.Now(),
		Action:    "user_login",
		Actor:     "user@example.com",
	}

	svc.EvaluateAuditEvent(ctx, entry, "user")

	// Only the enabled rule should execute its action
	if actionsExecuted != 1 {
		t.Errorf("expected 1 action executed (only enabled rule), got %d", actionsExecuted)
	}
}

func TestComplianceService_MultipleFrameworks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	ctx := context.Background()

	// Add gaps from different frameworks
	gaps := []ComplianceGap{
		{ID: "gap1", Framework: "SOC2", Control: "CC6.1", Severity: "critical", Status: "open"},
		{ID: "gap2", Framework: "HIPAA", Control: "164.308", Severity: "high", Status: "open"},
		{ID: "gap3", Framework: "SOC2", Control: "A1.1", Severity: "high", Status: "open"},
	}

	for _, gap := range gaps {
		err := svc.addComplianceGap(ctx, gap)
		if err != nil {
			t.Fatalf("addComplianceGap failed: %v", err)
		}
	}

	// Verify gaps are organized by framework
	svc.mu.RLock()
	soc2Gaps := svc.gaps["SOC2"]
	hipaaGaps := svc.gaps["HIPAA"]
	svc.mu.RUnlock()

	if len(soc2Gaps) != 2 {
		t.Errorf("expected 2 SOC2 gaps, got %d", len(soc2Gaps))
	}

	if len(hipaaGaps) != 1 {
		t.Errorf("expected 1 HIPAA gap, got %d", len(hipaaGaps))
	}
}

func TestComplianceService_ConcurrentEvaluation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewComplianceService("http://localhost:8089", logger)

	testRules := []Rule{
		{
			ID:        "test.rule",
			Title:     "Test Rule",
			Condition: "user.role != admin",
			Action:    "audit",
			Enabled:   true,
		},
	}

	svc.ruleEngine.SetRules(testRules)

	ctx := context.Background()

	// Run concurrent evaluations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			entry := AuditEntry{
				ID:        "aud" + string(rune(id)),
				SessionID: "sess1",
				Timestamp: time.Now(),
				Action:    "user_login",
				Actor:     "user@example.com",
			}

			err := svc.EvaluateAuditEvent(ctx, entry, "user")
			if err != nil {
				t.Errorf("EvaluateAuditEvent failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
