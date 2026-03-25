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
	EventID                   string
	SessionID                 string
	Timestamp                 time.Time
	Action                    string
	Actor                     string
	Details                   map[string]interface{}
	UserRole                  string
	DataType                  string
	Encrypted                 bool
	Uptime                    float64
	SignatureValid            bool
	DataClassification        string // HIPAA: "phi", "general", etc.
	DataContainsPHI           bool   // HIPAA: whether data contains Protected Health Information
	TransmissionProtocol      string // HIPAA: "https", "http", etc.
	AuditLogMissingPHIEntries bool   // HIPAA: whether PHI access audit log has gaps
	DataRetentionDays         int    // HIPAA: data retention days
	MessageContainsPHI        bool   // HIPAA: whether transmitted message contains PHI
	// GDPR-specific fields
	DataSubjectRequestPending     bool   // GDPR: pending data subject request exists
	DaysElapsedSinceRequest       int    // GDPR: days since request (for 30-day deadline)
	DataProcessingRequiresConsent bool   // GDPR: does this processing require consent
	UserConsentGiven              bool   // GDPR: has user given consent
	ProcessorDPASigned            bool   // GDPR: is Data Processing Agreement signed
	ProcessorHandlesData          bool   // GDPR: does processor handle personal data
	DataCollectedFieldCount       int    // GDPR: number of fields collected
	DataNeededFieldCount          int    // GDPR: number of fields actually needed
	DataContainsPII               bool   // GDPR: data contains personally identifiable information
	DataLocation                  string // GDPR: where data is stored ("eu", "us", "other")
	OrgRegion                     string // GDPR: organization region ("eu", "us", "other")
	// SOX-specific fields
	ChangeRequiresApproval     bool    // SOX: whether change needs approval
	ChangeApprovedBy           string  // SOX: who approved the change
	ChangeMadeBy               string  // SOX: who made the change
	SystemMeasuredUptime       float64 // SOX: measured system uptime percentage
	FinancialDataAccessLogged  bool    // SOX: whether access to financial data is logged
	AuditLogRetentionDays      int     // SOX: audit log retention days
	ProductionChangeDocumented bool    // SOX: whether production change is documented
	FinancialRecordHasChecksum bool    // SOX: whether financial record has checksum
	ChecksumVerified           bool    // SOX: whether checksum is verified
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
//   - "data.classification == phi"
//   - "data.contains_phi == true"
//   - "transmission.protocol != https"
//   - "audit_log.missing_phi_access_entries == true"
//   - "data.retention_days > 2555"
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
		// HIPAA: data.classification == phi
		{
			regex: regexp.MustCompile(`^data\.classification\s*(==|!=)\s*(\w+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data\.classification\s*(==|!=)\s*(\w+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid data.classification condition"
				}
				op := matches[1]
				expectedClass := matches[2]
				if op == "!=" {
					if ctx.DataClassification != expectedClass {
						return true, fmt.Sprintf("data.classification=%s (not %s)", ctx.DataClassification, expectedClass)
					}
				} else if op == "==" {
					if ctx.DataClassification == expectedClass {
						return true, fmt.Sprintf("data.classification=%s", ctx.DataClassification)
					}
				}
				return false, fmt.Sprintf("data.classification check failed: classification=%s, expected=%s", ctx.DataClassification, expectedClass)
			},
		},
		// HIPAA: data.contains_phi == true/false
		{
			regex: regexp.MustCompile(`^data\.contains_phi\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data\.contains_phi\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid data.contains_phi condition"
				}
				op := matches[1]
				expectedVal := matches[2] == "true"
				if op == "==" {
					if ctx.DataContainsPHI == expectedVal {
						return true, fmt.Sprintf("data.contains_phi=%v", ctx.DataContainsPHI)
					}
				} else if op == "!=" {
					if ctx.DataContainsPHI != expectedVal {
						return true, fmt.Sprintf("data.contains_phi=%v (not %v)", ctx.DataContainsPHI, expectedVal)
					}
				}
				return false, fmt.Sprintf("data.contains_phi check failed: contains_phi=%v, expected=%v", ctx.DataContainsPHI, expectedVal)
			},
		},
		// HIPAA: transmission.protocol != https
		{
			regex: regexp.MustCompile(`^transmission\.protocol\s*(!=|==)\s*(\w+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^transmission\.protocol\s*(!=|==)\s*(\w+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid transmission.protocol condition"
				}
				op := matches[1]
				expectedProto := matches[2]
				if op == "!=" {
					if ctx.TransmissionProtocol != expectedProto {
						return true, fmt.Sprintf("transmission.protocol=%s (not %s)", ctx.TransmissionProtocol, expectedProto)
					}
				} else if op == "==" {
					if ctx.TransmissionProtocol == expectedProto {
						return true, fmt.Sprintf("transmission.protocol=%s", ctx.TransmissionProtocol)
					}
				}
				return false, fmt.Sprintf("transmission.protocol check failed: protocol=%s, expected=%s", ctx.TransmissionProtocol, expectedProto)
			},
		},
		// HIPAA: audit_log.missing_phi_access_entries == true/false
		{
			regex: regexp.MustCompile(`^audit_log\.missing_phi_access_entries\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^audit_log\.missing_phi_access_entries\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid audit_log.missing_phi_access_entries condition"
				}
				op := matches[1]
				expectedVal := matches[2] == "true"
				if op == "==" {
					if ctx.AuditLogMissingPHIEntries == expectedVal {
						return true, fmt.Sprintf("audit_log.missing_phi_access_entries=%v", ctx.AuditLogMissingPHIEntries)
					}
				} else if op == "!=" {
					if ctx.AuditLogMissingPHIEntries != expectedVal {
						return true, fmt.Sprintf("audit_log.missing_phi_access_entries=%v (not %v)", ctx.AuditLogMissingPHIEntries, expectedVal)
					}
				}
				return false, fmt.Sprintf("audit_log.missing_phi_access_entries check failed: missing=%v, expected=%v", ctx.AuditLogMissingPHIEntries, expectedVal)
			},
		},
		// HIPAA: data.retention_days > 2555
		{
			regex: regexp.MustCompile(`^data\.retention_days\s*(<|>|<=|>=)\s*(\d+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data\.retention_days\s*(<|>|<=|>=)\s*(\d+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid data.retention_days condition"
				}
				op := matches[1]
				var threshold int
				fmt.Sscanf(matches[2], "%d", &threshold)

				matched := false
				switch op {
				case "<":
					matched = ctx.DataRetentionDays < threshold
				case ">":
					matched = ctx.DataRetentionDays > threshold
				case "<=":
					matched = ctx.DataRetentionDays <= threshold
				case ">=":
					matched = ctx.DataRetentionDays >= threshold
				}

				if matched {
					return true, fmt.Sprintf("data.retention_days=%d %s %d", ctx.DataRetentionDays, op, threshold)
				}
				return false, fmt.Sprintf("data.retention_days check failed: %d %s %d", ctx.DataRetentionDays, op, threshold)
			},
		},
		// GDPR: data_subject_request.pending == true AND days_elapsed > 30
		{
			regex: regexp.MustCompile(`^data_subject_request\.pending\s*(==|!=)\s*(true|false)\s+AND\s+days_elapsed\s*(>|<|>=|<=)\s*(\d+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data_subject_request\.pending\s*(==|!=)\s*(true|false)\s+AND\s+days_elapsed\s*(>|<|>=|<=)\s*(\d+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 5 {
					return false, "invalid data_subject_request condition"
				}
				pendingOp := matches[1]
				pendingVal := matches[2] == "true"
				daysOp := matches[3]
				var dayThreshold int
				fmt.Sscanf(matches[4], "%d", &dayThreshold)

				pendingMatches := false
				if pendingOp == "==" {
					pendingMatches = ctx.DataSubjectRequestPending == pendingVal
				} else if pendingOp == "!=" {
					pendingMatches = ctx.DataSubjectRequestPending != pendingVal
				}

				if !pendingMatches {
					return false, fmt.Sprintf("data_subject_request.pending=%v (expected %v)", ctx.DataSubjectRequestPending, pendingVal)
				}

				daysMatches := false
				switch daysOp {
				case ">":
					daysMatches = ctx.DaysElapsedSinceRequest > dayThreshold
				case "<":
					daysMatches = ctx.DaysElapsedSinceRequest < dayThreshold
				case ">=":
					daysMatches = ctx.DaysElapsedSinceRequest >= dayThreshold
				case "<=":
					daysMatches = ctx.DaysElapsedSinceRequest <= dayThreshold
				}

				if daysMatches && pendingMatches {
					return true, fmt.Sprintf("data_subject_request.pending=%v AND days_elapsed=%d %s %d", ctx.DataSubjectRequestPending, ctx.DaysElapsedSinceRequest, daysOp, dayThreshold)
				}
				return false, fmt.Sprintf("data_subject_request check failed: pending=%v, days=%d", ctx.DataSubjectRequestPending, ctx.DaysElapsedSinceRequest)
			},
		},
		// GDPR: data_processing.requires_consent == true AND user.consent_given != true
		{
			regex: regexp.MustCompile(`^data_processing\.requires_consent\s*(==|!=)\s*(true|false)\s+AND\s+user\.consent_given\s*(!=|==)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data_processing\.requires_consent\s*(==|!=)\s*(true|false)\s+AND\s+user\.consent_given\s*(!=|==)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 5 {
					return false, "invalid data_processing.requires_consent condition"
				}
				requiresOp := matches[1]
				requiresVal := matches[2] == "true"
				consentOp := matches[3]
				consentVal := matches[4] == "true"

				requiresMatches := false
				if requiresOp == "==" {
					requiresMatches = ctx.DataProcessingRequiresConsent == requiresVal
				} else if requiresOp == "!=" {
					requiresMatches = ctx.DataProcessingRequiresConsent != requiresVal
				}

				if !requiresMatches {
					return false, fmt.Sprintf("data_processing.requires_consent=%v (expected %v)", ctx.DataProcessingRequiresConsent, requiresVal)
				}

				consentMatches := false
				if consentOp == "==" {
					consentMatches = ctx.UserConsentGiven == consentVal
				} else if consentOp == "!=" {
					consentMatches = ctx.UserConsentGiven != consentVal
				}

				if requiresMatches && consentMatches {
					return true, fmt.Sprintf("data_processing.requires_consent=%v AND user.consent_given=%v", ctx.DataProcessingRequiresConsent, ctx.UserConsentGiven)
				}
				return false, fmt.Sprintf("data_processing check failed: requires_consent=%v, consent_given=%v", ctx.DataProcessingRequiresConsent, ctx.UserConsentGiven)
			},
		},
		// GDPR: processor.dpa_signed != true AND processor.handles_data == true
		{
			regex: regexp.MustCompile(`^processor\.dpa_signed\s*(!=|==)\s*(true|false)\s+AND\s+processor\.handles_data\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^processor\.dpa_signed\s*(!=|==)\s*(true|false)\s+AND\s+processor\.handles_data\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 5 {
					return false, "invalid processor.dpa_signed condition"
				}
				dpaOp := matches[1]
				dpaVal := matches[2] == "true"
				handlesOp := matches[3]
				handlesVal := matches[4] == "true"

				dpaMatches := false
				if dpaOp == "==" {
					dpaMatches = ctx.ProcessorDPASigned == dpaVal
				} else if dpaOp == "!=" {
					dpaMatches = ctx.ProcessorDPASigned != dpaVal
				}

				handlesMatches := false
				if handlesOp == "==" {
					handlesMatches = ctx.ProcessorHandlesData == handlesVal
				} else if handlesOp == "!=" {
					handlesMatches = ctx.ProcessorHandlesData != handlesVal
				}

				if dpaMatches && handlesMatches {
					return true, fmt.Sprintf("processor.dpa_signed=%v AND processor.handles_data=%v", ctx.ProcessorDPASigned, ctx.ProcessorHandlesData)
				}
				return false, fmt.Sprintf("processor check failed: dpa_signed=%v, handles_data=%v", ctx.ProcessorDPASigned, ctx.ProcessorHandlesData)
			},
		},
		// GDPR: data_collected.field_count > data_needed.field_count
		{
			regex: regexp.MustCompile(`^data_collected\.field_count\s*(>|<|>=|<=)\s*data_needed\.field_count$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data_collected\.field_count\s*(>|<|>=|<=)\s*data_needed\.field_count$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 2 {
					return false, "invalid data_collected.field_count condition"
				}
				op := matches[1]

				matched := false
				switch op {
				case ">":
					matched = ctx.DataCollectedFieldCount > ctx.DataNeededFieldCount
				case "<":
					matched = ctx.DataCollectedFieldCount < ctx.DataNeededFieldCount
				case ">=":
					matched = ctx.DataCollectedFieldCount >= ctx.DataNeededFieldCount
				case "<=":
					matched = ctx.DataCollectedFieldCount <= ctx.DataNeededFieldCount
				}

				if matched {
					return true, fmt.Sprintf("data_collected.field_count=%d %s data_needed.field_count=%d", ctx.DataCollectedFieldCount, op, ctx.DataNeededFieldCount)
				}
				return false, fmt.Sprintf("data_collected check failed: collected=%d, needed=%d", ctx.DataCollectedFieldCount, ctx.DataNeededFieldCount)
			},
		},
		// GDPR: data.contains_pii == true AND data.location != eu AND org.region == eu
		{
			regex: regexp.MustCompile(`^data\.contains_pii\s*(==|!=)\s*(true|false)\s+AND\s+data\.location\s*(!=|==)\s*(\w+)\s+AND\s+org\.region\s*(==|!=)\s*(\w+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^data\.contains_pii\s*(==|!=)\s*(true|false)\s+AND\s+data\.location\s*(!=|==)\s*(\w+)\s+AND\s+org\.region\s*(==|!=)\s*(\w+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 7 {
					return false, "invalid data.contains_pii condition"
				}
				piiOp := matches[1]
				piiVal := matches[2] == "true"
				locOp := matches[3]
				expectedLoc := matches[4]
				regionOp := matches[5]
				expectedRegion := matches[6]

				piiMatches := false
				if piiOp == "==" {
					piiMatches = ctx.DataContainsPII == piiVal
				} else if piiOp == "!=" {
					piiMatches = ctx.DataContainsPII != piiVal
				}

				if !piiMatches {
					return false, fmt.Sprintf("data.contains_pii=%v (expected %v)", ctx.DataContainsPII, piiVal)
				}

				locMatches := false
				if locOp == "==" {
					locMatches = ctx.DataLocation == expectedLoc
				} else if locOp == "!=" {
					locMatches = ctx.DataLocation != expectedLoc
				}

				if !locMatches {
					return false, fmt.Sprintf("data.location=%s (expected %s)", ctx.DataLocation, expectedLoc)
				}

				regionMatches := false
				if regionOp == "==" {
					regionMatches = ctx.OrgRegion == expectedRegion
				} else if regionOp == "!=" {
					regionMatches = ctx.OrgRegion != expectedRegion
				}

				if piiMatches && locMatches && regionMatches {
					return true, fmt.Sprintf("data.contains_pii=%v AND data.location=%s AND org.region=%s", ctx.DataContainsPII, ctx.DataLocation, ctx.OrgRegion)
				}
				return false, fmt.Sprintf("data location check failed: contains_pii=%v, location=%s, org_region=%s", ctx.DataContainsPII, ctx.DataLocation, ctx.OrgRegion)
			},
		},
		// SOX: change.requires_approval == true AND change.approved_by == change.made_by
		{
			regex: regexp.MustCompile(`^change\.requires_approval\s*(==|!=)\s*(true|false)\s+AND\s+change\.approved_by\s*(==|!=)\s*change\.made_by$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^change\.requires_approval\s*(==|!=)\s*(true|false)\s+AND\s+change\.approved_by\s*(==|!=)\s*change\.made_by$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 4 {
					return false, "invalid change.requires_approval condition"
				}
				approvalOp := matches[1]
				approvalVal := matches[2] == "true"
				comparisonOp := matches[3]

				approvalMatches := false
				if approvalOp == "==" {
					approvalMatches = ctx.ChangeRequiresApproval == approvalVal
				} else if approvalOp == "!=" {
					approvalMatches = ctx.ChangeRequiresApproval != approvalVal
				}

				if !approvalMatches {
					return false, fmt.Sprintf("change.requires_approval=%v (expected %v)", ctx.ChangeRequiresApproval, approvalVal)
				}

				// Check if approved_by == made_by (segregation of duties violation)
				comparisonMatches := false
				if comparisonOp == "==" {
					comparisonMatches = ctx.ChangeApprovedBy == ctx.ChangeMadeBy
				} else if comparisonOp == "!=" {
					comparisonMatches = ctx.ChangeApprovedBy != ctx.ChangeMadeBy
				}

				if approvalMatches && comparisonMatches {
					return true, fmt.Sprintf("change.requires_approval=%v AND approved_by=%s %s made_by=%s", ctx.ChangeRequiresApproval, ctx.ChangeApprovedBy, comparisonOp, ctx.ChangeMadeBy)
				}
				return false, fmt.Sprintf("change segregation check failed: requires_approval=%v, approved_by=%s, made_by=%s", ctx.ChangeRequiresApproval, ctx.ChangeApprovedBy, ctx.ChangeMadeBy)
			},
		},
		// SOX: system.measured_uptime < 99.9
		{
			regex: regexp.MustCompile(`^system\.measured_uptime\s*(<|>|<=|>=)\s*(\d+(?:\.\d+)?)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^system\.measured_uptime\s*(<|>|<=|>=)\s*(\d+(?:\.\d+)?)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid system.measured_uptime condition"
				}
				op := matches[1]
				var threshold float64
				fmt.Sscanf(matches[2], "%f", &threshold)

				matched := false
				switch op {
				case "<":
					matched = ctx.SystemMeasuredUptime < threshold
				case ">":
					matched = ctx.SystemMeasuredUptime > threshold
				case "<=":
					matched = ctx.SystemMeasuredUptime <= threshold
				case ">=":
					matched = ctx.SystemMeasuredUptime >= threshold
				}

				if matched {
					return true, fmt.Sprintf("system.measured_uptime=%f %s %f", ctx.SystemMeasuredUptime, op, threshold)
				}
				return false, fmt.Sprintf("system.measured_uptime check failed: %f %s %f", ctx.SystemMeasuredUptime, op, threshold)
			},
		},
		// SOX: financial_data.access_logged == false OR audit_log.retention_days < 2555
		{
			regex: regexp.MustCompile(`^financial_data\.access_logged\s*(==|!=)\s*(true|false)\s+OR\s+audit_log\.retention_days\s*(<|>|<=|>=)\s*(\d+)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^financial_data\.access_logged\s*(==|!=)\s*(true|false)\s+OR\s+audit_log\.retention_days\s*(<|>|<=|>=)\s*(\d+)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 5 {
					return false, "invalid financial_data.access_logged condition"
				}
				loggedOp := matches[1]
				loggedVal := matches[2] == "true"
				retentionOp := matches[3]
				var retentionThreshold int
				fmt.Sscanf(matches[4], "%d", &retentionThreshold)

				loggedMatches := false
				if loggedOp == "==" {
					loggedMatches = ctx.FinancialDataAccessLogged == loggedVal
				} else if loggedOp == "!=" {
					loggedMatches = ctx.FinancialDataAccessLogged != loggedVal
				}

				retentionMatches := false
				switch retentionOp {
				case "<":
					retentionMatches = ctx.AuditLogRetentionDays < retentionThreshold
				case ">":
					retentionMatches = ctx.AuditLogRetentionDays > retentionThreshold
				case "<=":
					retentionMatches = ctx.AuditLogRetentionDays <= retentionThreshold
				case ">=":
					retentionMatches = ctx.AuditLogRetentionDays >= retentionThreshold
				}

				// OR condition: either one being true triggers the gap
				if loggedMatches || retentionMatches {
					return true, fmt.Sprintf("financial_data.access_logged=%v OR audit_log.retention_days=%d %s %d", ctx.FinancialDataAccessLogged, ctx.AuditLogRetentionDays, retentionOp, retentionThreshold)
				}
				return false, fmt.Sprintf("financial data access check failed: logged=%v, retention_days=%d", ctx.FinancialDataAccessLogged, ctx.AuditLogRetentionDays)
			},
		},
		// SOX: production_change.documented == false
		{
			regex: regexp.MustCompile(`^production_change\.documented\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^production_change\.documented\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 3 {
					return false, "invalid production_change.documented condition"
				}
				op := matches[1]
				expectedVal := matches[2] == "true"
				if op == "==" {
					if ctx.ProductionChangeDocumented == expectedVal {
						return true, fmt.Sprintf("production_change.documented=%v", ctx.ProductionChangeDocumented)
					}
				} else if op == "!=" {
					if ctx.ProductionChangeDocumented != expectedVal {
						return true, fmt.Sprintf("production_change.documented=%v (not %v)", ctx.ProductionChangeDocumented, expectedVal)
					}
				}
				return false, fmt.Sprintf("production_change.documented check failed: documented=%v, expected=%v", ctx.ProductionChangeDocumented, expectedVal)
			},
		},
		// SOX: financial_record.has_checksum == false OR checksum.verified == false
		{
			regex: regexp.MustCompile(`^financial_record\.has_checksum\s*(==|!=)\s*(true|false)\s+OR\s+checksum\.verified\s*(==|!=)\s*(true|false)$`),
			handler: func(cond string, ctx RuleEvaluationContext) (bool, string) {
				re := regexp.MustCompile(`^financial_record\.has_checksum\s*(==|!=)\s*(true|false)\s+OR\s+checksum\.verified\s*(==|!=)\s*(true|false)$`)
				matches := re.FindStringSubmatch(cond)
				if len(matches) < 5 {
					return false, "invalid financial_record.has_checksum condition"
				}
				checksumOp := matches[1]
				checksumVal := matches[2] == "true"
				verifiedOp := matches[3]
				verifiedVal := matches[4] == "true"

				checksumMatches := false
				if checksumOp == "==" {
					checksumMatches = ctx.FinancialRecordHasChecksum == checksumVal
				} else if checksumOp == "!=" {
					checksumMatches = ctx.FinancialRecordHasChecksum != checksumVal
				}

				verifiedMatches := false
				if verifiedOp == "==" {
					verifiedMatches = ctx.ChecksumVerified == verifiedVal
				} else if verifiedOp == "!=" {
					verifiedMatches = ctx.ChecksumVerified != verifiedVal
				}

				// OR condition: either one being true triggers the gap
				if checksumMatches || verifiedMatches {
					return true, fmt.Sprintf("financial_record.has_checksum=%v OR checksum.verified=%v", ctx.FinancialRecordHasChecksum, ctx.ChecksumVerified)
				}
				return false, fmt.Sprintf("financial_record checksum check failed: has_checksum=%v, verified=%v", ctx.FinancialRecordHasChecksum, ctx.ChecksumVerified)
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
