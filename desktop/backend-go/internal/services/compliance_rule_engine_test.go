package services

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRuleEngine_EvaluateCondition_UserRole(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
	}{
		{
			name:      "user.role != admin (should match when not admin)",
			condition: "user.role != admin",
			ctx: RuleEvaluationContext{
				UserRole: "user",
			},
			expected: true,
		},
		{
			name:      "user.role != admin (should not match when admin)",
			condition: "user.role != admin",
			ctx: RuleEvaluationContext{
				UserRole: "admin",
			},
			expected: false,
		},
		{
			name:      "user.role == admin (should match when admin)",
			condition: "user.role == admin",
			ctx: RuleEvaluationContext{
				UserRole: "admin",
			},
			expected: true,
		},
		{
			name:      "user.role == guest (should match when guest)",
			condition: "user.role == guest",
			ctx: RuleEvaluationContext{
				UserRole: "guest",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestRuleEngine_EvaluateCondition_DataEncryption(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
	}{
		{
			name:      "data.encrypted == true (should match when encrypted)",
			condition: "data.encrypted == true",
			ctx: RuleEvaluationContext{
				Encrypted: true,
			},
			expected: true,
		},
		{
			name:      "data.encrypted == true (should not match when not encrypted)",
			condition: "data.encrypted == true",
			ctx: RuleEvaluationContext{
				Encrypted: false,
			},
			expected: false,
		},
		{
			name:      "data.encrypted != true (should match when not encrypted)",
			condition: "data.encrypted != true",
			ctx: RuleEvaluationContext{
				Encrypted: false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestRuleEngine_EvaluateCondition_ServiceUptime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
	}{
		{
			name:      "service.uptime < 99.9 (should match for 99.5)",
			condition: "service.uptime < 99.9",
			ctx: RuleEvaluationContext{
				Uptime: 99.5,
			},
			expected: true,
		},
		{
			name:      "service.uptime < 99.9 (should not match for 99.95)",
			condition: "service.uptime < 99.9",
			ctx: RuleEvaluationContext{
				Uptime: 99.95,
			},
			expected: false,
		},
		{
			name:      "service.uptime >= 99.9 (should match for 99.95)",
			condition: "service.uptime >= 99.9",
			ctx: RuleEvaluationContext{
				Uptime: 99.95,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestRuleEngine_EvaluateCondition_SignatureValidity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	tests := []struct {
		name      string
		condition string
		ctx       RuleEvaluationContext
		expected  bool
	}{
		{
			name:      "audit_entry.signature_valid == true (should match when valid)",
			condition: "audit_entry.signature_valid == true",
			ctx: RuleEvaluationContext{
				SignatureValid: true,
			},
			expected: true,
		},
		{
			name:      "audit_entry.signature_valid == true (should not match when invalid)",
			condition: "audit_entry.signature_valid == true",
			ctx: RuleEvaluationContext{
				SignatureValid: false,
			},
			expected: false,
		},
		{
			name:      "audit_entry.signature_valid != true (should match when invalid)",
			condition: "audit_entry.signature_valid != true",
			ctx: RuleEvaluationContext{
				SignatureValid: false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := engine.evaluateCondition(tt.condition, tt.ctx)
			if matched != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestRuleEngine_EvaluateAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	rules := []Rule{
		{
			ID:        "rule1",
			Title:     "Test Rule 1",
			Condition: "user.role != admin",
			Action:    "audit",
			Enabled:   true,
			Severity:  "high",
			Framework: "SOC2",
		},
		{
			ID:        "rule2",
			Title:     "Test Rule 2",
			Condition: "data.encrypted == true",
			Action:    "audit",
			Enabled:   true,
			Severity:  "critical",
			Framework: "SOC2",
		},
		{
			ID:        "rule3",
			Title:     "Test Rule 3 (disabled)",
			Condition: "user.role == admin",
			Action:    "audit",
			Enabled:   false,
			Severity:  "low",
			Framework: "SOC2",
		},
	}

	engine.SetRules(rules)

	ctx := context.Background()
	ruleCtx := RuleEvaluationContext{
		EventID:   "evt1",
		UserRole:  "user",
		Encrypted: true,
		Timestamp: time.Now(),
	}

	results := engine.EvaluateAll(ctx, ruleCtx)

	// We should have results for rule1 (matches) and rule2 (matches), rule3 is disabled
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify rule1 matched
	if !results[0].Matched {
		t.Errorf("rule1 should have matched (user.role != admin for role=user)")
	}

	// Verify rule2 matched
	if !results[1].Matched {
		t.Errorf("rule2 should have matched (data.encrypted == true)")
	}
}

func TestRuleEngine_Caching(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	rule := Rule{
		ID:        "rule1",
		Title:     "Test Rule",
		Condition: "user.role != admin",
		Action:    "audit",
		Enabled:   true,
	}
	engine.SetRules([]Rule{rule})

	ctx := context.Background()
	ruleCtx := RuleEvaluationContext{
		EventID:   "evt1",
		UserRole:  "user",
		Timestamp: time.Now(),
	}

	// First evaluation
	results1 := engine.EvaluateAll(ctx, ruleCtx)
	if len(results1) != 1 {
		t.Errorf("expected 1 result, got %d", len(results1))
	}

	// Second evaluation (should come from cache)
	results2 := engine.EvaluateAll(ctx, ruleCtx)
	if len(results2) != 1 {
		t.Errorf("expected 1 result, got %d", len(results2))
	}

	// Both should have same result
	if results1[0].RuleID != results2[0].RuleID {
		t.Errorf("cached result should match first evaluation")
	}
}

func TestRuleEngine_CacheClearExpired(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)
	engine.cacheExpiry = 1 * time.Millisecond

	rule := Rule{
		ID:        "rule1",
		Title:     "Test Rule",
		Condition: "user.role != admin",
		Action:    "audit",
		Enabled:   true,
	}
	engine.SetRules([]Rule{rule})

	ctx := context.Background()
	ruleCtx := RuleEvaluationContext{
		EventID:   "evt1",
		UserRole:  "user",
		Timestamp: time.Now(),
	}

	// First evaluation (caches result)
	engine.EvaluateAll(ctx, ruleCtx)

	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)

	// Clear expired entries
	engine.ClearExpiredCache()

	// Verify cache is empty
	engine.mu.RLock()
	if len(engine.cache) != 0 {
		t.Errorf("expected empty cache after clearing expired entries, got %d", len(engine.cache))
	}
	engine.mu.RUnlock()
}

func TestRuleEngine_InvalidCondition(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	matched, msg := engine.evaluateCondition("invalid.condition format", RuleEvaluationContext{})
	if matched {
		t.Errorf("expected no match for invalid condition")
	}
	if msg == "" {
		t.Errorf("expected error message for invalid condition")
	}
}

func TestRuleEngine_MultipleActions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	engine := NewRuleEngine(logger)

	actionsCalled := make(map[string]int)

	engine.SetGapHandler(func(ctx context.Context, gap ComplianceGap) error {
		actionsCalled["gap"]++
		return nil
	})

	engine.SetNotifyHandler(func(ctx context.Context, ruleID, message string) error {
		actionsCalled["notify"]++
		return nil
	})

	rules := []Rule{
		{
			ID:        "gap_rule",
			Title:     "Gap Rule",
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

	engine.SetRules(rules)

	ctx := context.Background()
	ruleCtx := RuleEvaluationContext{
		EventID:   "evt1",
		UserRole:  "user",
		Timestamp: time.Now(),
	}

	engine.EvaluateAll(ctx, ruleCtx)

	if actionsCalled["gap"] != 1 {
		t.Errorf("expected gap action to be called once, got %d", actionsCalled["gap"])
	}

	if actionsCalled["notify"] != 1 {
		t.Errorf("expected notify action to be called once, got %d", actionsCalled["notify"])
	}
}
