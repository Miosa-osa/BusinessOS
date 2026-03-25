package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRuleLoader_LoadRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Use the actual config file
	configPath := "../../../../config/compliance-rules.yaml"

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("config file not found at %s, skipping integration test", configPath)
	}

	loader := NewRuleLoader(configPath, logger)
	rules, err := loader.LoadRules()

	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}

	if len(rules) == 0 {
		t.Errorf("expected rules to be loaded, got %d", len(rules))
	}

	// Verify specific rules
	expectedRules := map[string]bool{
		"soc2.cc6.1": false,
		"soc2.c1.1":  false,
		"soc2.i1.1":  false,
	}

	for _, rule := range rules {
		if _, ok := expectedRules[rule.ID]; ok {
			expectedRules[rule.ID] = true
		}
	}

	for ruleID, found := range expectedRules {
		if !found {
			t.Errorf("expected rule %s not found", ruleID)
		}
	}
}

func TestRuleLoader_ValidateRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "high",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			rule: Rule{
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			rule: Rule{
				ID:        "test.rule",
				Condition: "user.role != admin",
				Action:    "audit",
			},
			wantErr: true,
		},
		{
			name: "missing condition",
			rule: Rule{
				ID:     "test.rule",
				Title:  "Test Rule",
				Action: "audit",
			},
			wantErr: true,
		},
		{
			name: "missing action",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
			},
			wantErr: true,
		},
		{
			name: "invalid action",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "invalid_action",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "invalid_severity",
			},
			wantErr: true,
		},
		{
			name: "valid severity critical",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "critical",
			},
			wantErr: false,
		},
		{
			name: "valid severity high",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "high",
			},
			wantErr: false,
		},
		{
			name: "valid severity medium",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "medium",
			},
			wantErr: false,
		},
		{
			name: "valid severity low",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "audit",
				Severity:  "low",
			},
			wantErr: false,
		},
		{
			name: "valid action create_gap",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "create_gap",
			},
			wantErr: false,
		},
		{
			name: "valid action notify",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "notify",
			},
			wantErr: false,
		},
		{
			name: "valid action escalate",
			rule: Rule{
				ID:        "test.rule",
				Title:     "Test Rule",
				Condition: "user.role != admin",
				Action:    "escalate",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRule(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleLoader_GetRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Create a temporary YAML file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-rules.yaml")

	content := `rules:
  - id: "test.rule1"
    title: "Test Rule 1"
    condition: "user.role != admin"
    action: "audit"
    enabled: true
    severity: "high"
    framework: "SOC2"
  - id: "test.rule2"
    title: "Test Rule 2"
    condition: "data.encrypted == true"
    action: "create_gap"
    enabled: true
    severity: "critical"
    framework: "SOC2"
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := NewRuleLoader(configPath, logger)
	rules, err := loader.LoadRules()
	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}

	// Test GetRules
	retrievedRules := loader.GetRules()
	if len(retrievedRules) != 2 {
		t.Errorf("GetRules() should return 2 rules, got %d", len(retrievedRules))
	}

	if retrievedRules[0].ID != "test.rule1" {
		t.Errorf("expected first rule ID test.rule1, got %s", retrievedRules[0].ID)
	}

	if retrievedRules[1].ID != "test.rule2" {
		t.Errorf("expected second rule ID test.rule2, got %s", retrievedRules[1].ID)
	}
}

func TestRuleLoader_InvalidYAML(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-rules.yaml")

	// Write invalid YAML
	content := `rules:
  - id: "test.rule"
    invalid yaml: [
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := NewRuleLoader(configPath, logger)
	_, err := loader.LoadRules()

	if err == nil {
		t.Errorf("expected error for invalid YAML, got none")
	}
}

func TestRuleLoader_InvalidRule(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-rule.yaml")

	// Valid YAML, but invalid rule (missing required field)
	content := `rules:
  - title: "Missing ID"
    condition: "user.role != admin"
    action: "audit"
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := NewRuleLoader(configPath, logger)
	_, err := loader.LoadRules()

	if err == nil {
		t.Errorf("expected error for invalid rule, got none")
	}
}
