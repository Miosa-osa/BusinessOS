package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"
)

// Rule represents a compliance rule with condition and action.
type Rule struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Condition string `json:"condition"` // simple DSL: "user.role != admin", "data.encrypted == true"
	Action    string `json:"action"`    // "create_gap" | "notify" | "escalate" | "audit"
	Enabled   bool   `json:"enabled"`
	Severity  string `json:"severity"`  // critical, high, medium, low
	Framework string `json:"framework"` // SOC2, HIPAA, GDPR, SOX
}

// RuleEvaluationContext holds context for rule evaluation.
type RuleEvaluationContext struct {
	EventID        string
	SessionID      string
	Timestamp      time.Time
	Action         string
	Actor          string
	Details        map[string]interface{}
	UserRole       string
	DataType       string
	Encrypted      bool
	Uptime         float64
	SignatureValid bool
}

// RuleResult represents the result of rule evaluation.
type RuleResult struct {
	RuleID    string
	Matched   bool
	Action    string
	Message   string
	Timestamp time.Time
}

// RuleEngine evaluates compliance rules against audit events.
type RuleEngine struct {
	mu            sync.RWMutex
	rules         []Rule
	cache         map[string]cacheEntry
	cacheExpiry   time.Duration
	logger        *slog.Logger
	notifyHandler func(context.Context, string, string) error
	gapHandler    func(context.Context, ComplianceGap) error
}

type cacheEntry struct {
	timestamp time.Time
	result    RuleResult
}

// NewRuleEngine creates a new rule engine with default settings.
func NewRuleEngine(logger *slog.Logger) *RuleEngine {
	return &RuleEngine{
		rules:       []Rule{},
		cache:       make(map[string]cacheEntry),
		cacheExpiry: 5 * time.Minute,
		logger:      logger,
	}
}

// SetRules replaces the rules in the engine.
func (re *RuleEngine) SetRules(rules []Rule) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.rules = rules
	re.logger.Info("rules loaded", "count", len(rules))
}

// GetRules returns the current rules.
func (re *RuleEngine) GetRules() []Rule {
	re.mu.RLock()
	defer re.mu.RUnlock()
	rules := make([]Rule, len(re.rules))
	copy(rules, re.rules)
	return rules
}

// SetNotifyHandler sets the callback for notify actions.
func (re *RuleEngine) SetNotifyHandler(handler func(context.Context, string, string) error) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.notifyHandler = handler
}

// SetGapHandler sets the callback for create_gap actions.
func (re *RuleEngine) SetGapHandler(handler func(context.Context, ComplianceGap) error) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.gapHandler = handler
}

// EvaluateAll evaluates all enabled rules against the context.
func (re *RuleEngine) EvaluateAll(ctx context.Context, ruleCtx RuleEvaluationContext) []RuleResult {
	re.mu.RLock()
	rules := make([]Rule, len(re.rules))
	copy(rules, re.rules)
	re.mu.RUnlock()

	results := []RuleResult{}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check cache
		cacheKey := fmt.Sprintf("%s:%s", rule.ID, ruleCtx.EventID)
		if cached, ok := re.getCachedResult(cacheKey); ok {
			results = append(results, cached)
			continue
		}

		// Evaluate
		result := re.evaluate(rule, ruleCtx)

		// Cache
		re.setCachedResult(cacheKey, result)

		// Dispatch action
		if result.Matched {
			re.dispatchAction(ctx, rule, result)
		}

		results = append(results, result)
	}

	return results
}

// evaluate checks if a rule condition matches the context.
func (re *RuleEngine) evaluate(rule Rule, ruleCtx RuleEvaluationContext) RuleResult {
	result := RuleResult{
		RuleID:    rule.ID,
		Matched:   false,
		Action:    rule.Action,
		Timestamp: time.Now(),
	}

	matched, msg := re.evaluateCondition(rule.Condition, ruleCtx)
	result.Matched = matched
	result.Message = msg

	if matched {
		re.logger.Info("rule matched",
			"rule_id", rule.ID,
			"rule_title", rule.Title,
			"action", rule.Action,
			"message", msg,
		)
	}

	return result
}

// evaluateCondition evaluates a rule condition DSL string.
// Supports:
//   - "user.role != admin"
//   - "data.encrypted == true"
//   - "service.uptime < 99.9"
//   - "audit_entry.signature_valid == false"
//   - "new_user.verified_at.IsZero() == true"
func (re *RuleEngine) evaluateCondition(condition string, ruleCtx RuleEvaluationContext) (bool, string) {
	// Simple condition parser - supports basic comparisons
	patterns := []struct {
		regex   *regexp.Regexp
		handler func(string, RuleEvaluationContext) (bool, string)
	}{
		{
			regex: regexp.MustCompile(`^user\.role\s*(!=|==)\s*(\w+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^user\.role\s*(!=|==)\s*(\w+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid user.role condition"
				}
				op := matches[1]
				expectedRole := matches[2]
				if op == "!=" {
					if ctx.UserRole != expectedRole {
						return true, fmt.Sprintf("user.role=%s (not %s)", ctx.UserRole, expectedRole)
					}
				} else if op == "==" {
					if ctx.UserRole == expectedRole {
						return true, fmt.Sprintf("user.role=%s", ctx.UserRole)
					}
				}
				return false, fmt.Sprintf("user.role check failed: role=%s, expected=%s", ctx.UserRole, expectedRole)
			},
		},
		{
			regex: regexp.MustCompile(`^data\.encrypted\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data\.encrypted\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid data.encrypted condition"
				}
				op := matches[1]
				expectedVal := matches[2] == "true"
				if op == "==" {
					if ctx.Encrypted == expectedVal {
						return true, fmt.Sprintf("data.encrypted=%v", ctx.Encrypted)
					}
				} else if op == "!=" {
					if ctx.Encrypted != expectedVal {
						return true, fmt.Sprintf("data.encrypted=%v (not %v)", ctx.Encrypted, expectedVal)
					}
				}
				return false, fmt.Sprintf("data.encrypted check failed: encrypted=%v, expected=%v", ctx.Encrypted, expectedVal)
			},
		},
		{
			regex: regexp.MustCompile(`^service\.uptime\s*(<|>|<=|>=)\s*(\d+(?:\.\d+)?)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^service\.uptime\s*(<|>|<=|>=)\s*(\d+(?:\.\d+)?)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid service.uptime condition"
				}
				op := matches[1]
				var threshold float64
				fmt.Sscanf(matches[2], "%f", &threshold)

				matched := false
				switch op {
				case "<":
					matched = ctx.Uptime < threshold
				case ">":
					matched = ctx.Uptime > threshold
				case "<=":
					matched = ctx.Uptime <= threshold
				case ">=":
					matched = ctx.Uptime >= threshold
				}

				if matched {
					return true, fmt.Sprintf("service.uptime=%f %s %f", ctx.Uptime, op, threshold)
				}
				return false, fmt.Sprintf("service.uptime check failed: %f %s %f", ctx.Uptime, op, threshold)
			},
		},
		{
			regex: regexp.MustCompile(`^audit_entry\.signature_valid\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^audit_entry\.signature_valid\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid audit_entry.signature_valid condition"
				}
				op := matches[1]
				expectedVal := matches[2] == "true"
				if op == "==" {
					if ctx.SignatureValid == expectedVal {
						return true, fmt.Sprintf("audit_entry.signature_valid=%v", ctx.SignatureValid)
					}
				} else if op == "!=" {
					if ctx.SignatureValid != expectedVal {
						return true, fmt.Sprintf("audit_entry.signature_valid=%v (not %v)", ctx.SignatureValid, expectedVal)
					}
				}
				return false, fmt.Sprintf("audit_entry.signature_valid check failed: valid=%v, expected=%v", ctx.SignatureValid, expectedVal)
			},
		},
	}

	for _, p := range patterns {
		if p.regex.MatchString(condition) {
			return p.handler(condition, ruleCtx)
		}
	}

	return false, fmt.Sprintf("unsupported condition format: %s", condition)
}

// dispatchAction performs the action associated with a rule match.
func (re *RuleEngine) dispatchAction(ctx context.Context, rule Rule, result RuleResult) {
	switch rule.Action {
	case "create_gap":
		re.createGap(ctx, rule)
	case "notify":
		re.notify(ctx, rule, result)
	case "escalate":
		re.escalate(ctx, rule, result)
	case "audit":
		re.audit(rule, result)
	default:
		re.logger.Warn("unknown rule action", "action", rule.Action)
	}
}

// createGap creates a compliance gap from a triggered rule.
func (re *RuleEngine) createGap(ctx context.Context, rule Rule) {
	if re.gapHandler == nil {
		re.logger.Warn("no gap handler configured, skipping create_gap")
		return
	}

	gap := ComplianceGap{
		ID:          fmt.Sprintf("gap-%d", time.Now().UnixNano()),
		Framework:   rule.Framework,
		Control:     rule.ID,
		Description: rule.Title,
		Severity:    rule.Severity,
		Status:      "open",
	}

	if err := re.gapHandler(ctx, gap); err != nil {
		re.logger.Error("failed to create gap", "error", err, "rule_id", rule.ID)
	}
}

// notify sends an alert for a triggered rule.
func (re *RuleEngine) notify(ctx context.Context, rule Rule, result RuleResult) {
	if re.notifyHandler == nil {
		re.logger.Warn("no notify handler configured, skipping notify")
		return
	}

	message := fmt.Sprintf("Compliance rule triggered: %s (%s) - %s",
		rule.Title, rule.ID, result.Message)

	if err := re.notifyHandler(ctx, rule.ID, message); err != nil {
		re.logger.Error("failed to send notification", "error", err, "rule_id", rule.ID)
	}
}

// escalate escalates a rule violation for manual review.
func (re *RuleEngine) escalate(ctx context.Context, rule Rule, result RuleResult) {
	re.logger.Error("compliance escalation required",
		"rule_id", rule.ID,
		"title", rule.Title,
		"severity", rule.Severity,
		"message", result.Message,
	)

	// Also notify for escalations
	re.notify(ctx, rule, result)
}

// audit logs a rule evaluation to the audit trail.
func (re *RuleEngine) audit(rule Rule, result RuleResult) {
	re.logger.Info("audit trail: rule evaluated",
		"rule_id", rule.ID,
		"matched", result.Matched,
		"action", rule.Action,
		"message", result.Message,
	)
}

// getCachedResult retrieves a cached rule evaluation if still valid.
func (re *RuleEngine) getCachedResult(key string) (RuleResult, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	entry, ok := re.cache[key]
	if !ok {
		return RuleResult{}, false
	}

	if time.Since(entry.timestamp) > re.cacheExpiry {
		return RuleResult{}, false
	}

	return entry.result, true
}

// setCachedResult caches a rule evaluation result.
func (re *RuleEngine) setCachedResult(key string, result RuleResult) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.cache[key] = cacheEntry{
		timestamp: time.Now(),
		result:    result,
	}
}

// ClearCache clears the evaluation cache.
func (re *RuleEngine) ClearCache() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.cache = make(map[string]cacheEntry)
}

// ClearExpiredCache removes expired entries from the cache.
func (re *RuleEngine) ClearExpiredCache() {
	re.mu.Lock()
	defer re.mu.Unlock()

	now := time.Now()
	for key, entry := range re.cache {
		if now.Sub(entry.timestamp) > re.cacheExpiry {
			delete(re.cache, key)
		}
	}
}
