package services

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// RuleConfig represents the structure of compliance-rules.yaml.
type RuleConfig struct {
	Rules []Rule `yaml:"rules"`
}

// RuleLoader handles loading and managing compliance rules from YAML files.
type RuleLoader struct {
	mu       sync.RWMutex
	filePath string
	rules    []Rule
	logger   *slog.Logger
	lastMod  int64
}

// NewRuleLoader creates a new rule loader for the specified YAML file.
func NewRuleLoader(filePath string, logger *slog.Logger) *RuleLoader {
	return &RuleLoader{
		filePath: filePath,
		rules:    []Rule{},
		logger:   logger,
	}
}

// LoadRules loads rules from the YAML file.
func (rl *RuleLoader) LoadRules() ([]Rule, error) {
	data, err := os.ReadFile(rl.filePath)
	if err != nil {
		return nil, fmt.Errorf("read rules file: %w", err)
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal rules YAML: %w", err)
	}

	// Validate rules
	for i, rule := range config.Rules {
		if err := validateRule(rule); err != nil {
			return nil, fmt.Errorf("validate rule %d (%s): %w", i, rule.ID, err)
		}
	}

	rl.mu.Lock()
	rl.rules = config.Rules
	rl.mu.Unlock()

	rl.logger.Info("rules loaded from file",
		"path", rl.filePath,
		"count", len(config.Rules),
	)

	return config.Rules, nil
}

// GetRules returns the currently loaded rules.
func (rl *RuleLoader) GetRules() []Rule {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rules := make([]Rule, len(rl.rules))
	copy(rules, rl.rules)
	return rules
}

// ReloadIfChanged reloads rules from disk if the file has been modified.
func (rl *RuleLoader) ReloadIfChanged() error {
	info, err := os.Stat(rl.filePath)
	if err != nil {
		return fmt.Errorf("stat rules file: %w", err)
	}

	modTime := info.ModTime().Unix()

	rl.mu.RLock()
	lastMod := rl.lastMod
	rl.mu.RUnlock()

	if modTime == lastMod {
		return nil
	}

	rules, err := rl.LoadRules()
	if err != nil {
		return err
	}

	rl.mu.Lock()
	rl.lastMod = modTime
	rl.mu.Unlock()

	rl.logger.Info("rules reloaded from file", "path", rl.filePath, "count", len(rules))
	return nil
}

// validateRule checks that a rule has all required fields.
func validateRule(rule Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule must have an id")
	}
	if rule.Title == "" {
		return fmt.Errorf("rule must have a title")
	}
	if rule.Condition == "" {
		return fmt.Errorf("rule must have a condition")
	}
	if rule.Action == "" {
		return fmt.Errorf("rule must have an action")
	}

	validActions := map[string]bool{
		"create_gap": true,
		"notify":     true,
		"escalate":   true,
		"audit":      true,
	}
	if !validActions[rule.Action] {
		return fmt.Errorf("invalid action '%s' (must be create_gap, notify, escalate, or audit)", rule.Action)
	}

	validSeverities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}
	if rule.Severity != "" && !validSeverities[rule.Severity] {
		return fmt.Errorf("invalid severity '%s' (must be critical, high, medium, or low)", rule.Severity)
	}

	return nil
}
