package optimal

import (
	"strings"
)

// RoutingResult is the output of a signal routing decision.
type RoutingResult struct {
	PrimaryNode   string   `json:"primary_node"`
	CrossRefNodes []string `json:"cross_ref_nodes"`
	Confidence    float64  `json:"confidence"`
}

// routingRule maps a set of trigger keywords to a primary node and optional
// cross-reference nodes. Keywords are matched case-insensitively against the
// full signal text. The first rule whose keywords produce any match wins the
// primary node; cross-refs accumulate from all matching rules.
type routingRule struct {
	keywords  []string
	primary   string
	crossRefs []string
}

// defaultRoutingRules is the canonical routing table from CLAUDE.md.
// Load order matters: rules are checked top-down; the first match sets the
// primary node, but all rules that match contribute their cross-refs.
var defaultRoutingRules = []routingRule{
	{
		keywords:  []string{"cliniciq", "clinic", "healthcare", "hbai", "ryan cole", "colt", "atif", "laura"},
		primary:   "06",
		crossRefs: []string{"03"},
	},
	{
		keywords:  []string{"ai masters", "ed honour", "robert potter", "adam", "curriculum", "beginner course", "advanced course"},
		primary:   "04",
		crossRefs: nil,
	},
	{
		keywords:  []string{"platform", "miosa", "skyscraper", "business os", "compute", "firecracker", "vm", "pedram", "middleware", "licensing"},
		primary:   "02",
		crossRefs: nil,
	},
	{
		keywords:  []string{"agency accelerants", "aa school", "bennett", "len", "liam", "jared", "d2d"},
		primary:   "06",
		crossRefs: nil,
	},
	{
		keywords:  []string{"community", "school group", "99/mo", "$99", "108 members", "paywall"},
		primary:   "07",
		crossRefs: nil,
	},
	{
		keywords:  []string{"content", "mosaic effect", "podcast", "sukhpreet", "ikram", "contentos", "tejas"},
		primary:   "08",
		crossRefs: nil,
	},
	{
		keywords:  []string{"os accelerator", "accelerator course"},
		primary:   "12",
		crossRefs: nil,
	},
	{
		keywords:  []string{"ahmed", "os architect", "youtube channel"},
		primary:   "05",
		crossRefs: nil,
	},
	{
		keywords:  []string{"revenue", "money", "invoice", "payment", "arr", "pipeline", "deal", "close", "earn", "income"},
		primary:   "11",
		crossRefs: nil,
	},
	{
		keywords:  []string{"pedro", "abdul", "nejd", "javaris", "team sync", "hiring", "george", "sean"},
		primary:   "10",
		crossRefs: nil,
	},
	{
		keywords:  []string{"jordan", "tom", "consortium", "political economy"},
		primary:   "01",
		crossRefs: []string{"02"},
	},
	{
		keywords:  []string{"personal", "health", "travel", "austin", "bangkok"},
		primary:   "01",
		crossRefs: nil,
	},
	// Financial signals always cross-reference money-revenue regardless of primary.
	// This is handled separately in the Route function below.
}

// financialKeywords are triggers that always add 11 as a cross-ref node.
var financialKeywords = []string{
	"revenue", "money", "invoice", "payment", "$", "arr", "pipeline",
	"deal closed", "close", "income", "earn", "profit", "loss",
}

// Route determines which node(s) a signal should be stored in, using the
// default routing table. To override rules, use RouteWithRules or
// RouteWithTopology.
func Route(text string) RoutingResult {
	return RouteWithRules(text, defaultRoutingRules)
}

// RouteWithTopology routes text using rules loaded from a TopologyConfig.
// When topo is nil or contains no routes, it falls back to defaultRoutingRules
// so callers without a configured topology.yaml work without changes.
func RouteWithTopology(text string, topo *TopologyConfig) RoutingResult {
	if topo != nil {
		if rules := topo.ToInternalRules(); rules != nil {
			return RouteWithRules(text, rules)
		}
	}
	return RouteWithRules(text, defaultRoutingRules)
}

// RouteWithRules evaluates rules against text and returns a RoutingResult.
// The first rule with any keyword match sets the primary node. All matching
// rules contribute their cross-refs. Financial content always cross-refs node 11.
func RouteWithRules(text string, rules []routingRule) RoutingResult {
	lower := strings.ToLower(text)

	primary := ""
	confidence := 0.0
	crossSet := make(map[string]struct{})

	for _, rule := range rules {
		matchCount := 0
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {
				matchCount++
			}
		}
		if matchCount == 0 {
			continue
		}

		// First match wins the primary.
		if primary == "" {
			primary = rule.primary
			confidence = scoreConfidence(matchCount, len(rule.keywords))
		}

		// All matches contribute cross-refs.
		for _, cr := range rule.crossRefs {
			crossSet[cr] = struct{}{}
		}
	}

	// If primary is set, remove it from cross-refs (no self-reference).
	delete(crossSet, primary)

	// Financial content always cross-refs node 11, unless 11 is already primary.
	if primary != "11" && hasFinancialSignal(lower) {
		crossSet["11"] = struct{}{}
	}

	// Default: unclassified signals go to inbox node 09.
	if primary == "" {
		primary = "09"
		confidence = 0.1
	}

	crossRefs := make([]string, 0, len(crossSet))
	for k := range crossSet {
		crossRefs = append(crossRefs, k)
	}

	return RoutingResult{
		PrimaryNode:   primary,
		CrossRefNodes: crossRefs,
		Confidence:    confidence,
	}
}

// scoreConfidence returns a 0.0–1.0 confidence value based on how many of a
// rule's keywords were matched versus total keywords in the rule.
func scoreConfidence(matched, total int) float64 {
	if total == 0 {
		return 0.0
	}
	ratio := float64(matched) / float64(total)
	// Base 0.5 for any match; scale the remainder by match ratio.
	return 0.5 + 0.5*ratio
}

// hasFinancialSignal reports whether text contains any financial keyword.
func hasFinancialSignal(lower string) bool {
	for _, kw := range financialKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
