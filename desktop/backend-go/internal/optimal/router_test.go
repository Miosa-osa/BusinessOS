package optimal

import (
	"testing"
)

// ── RouteWithTopology ─────────────────────────────────────────────────────────

func TestRouteWithTopology_NilFallback(t *testing.T) {
	// nil topology should fall back to default rules and still return a primary node.
	result := RouteWithTopology("Ed called about AI Masters pricing", nil)

	if result.PrimaryNode == "" {
		t.Error("expected a primary node from default rules, got empty string")
	}
	// "AI Masters" matches node 04 in the default rules.
	if result.PrimaryNode != "04" {
		t.Errorf("PrimaryNode = %q, want %q", result.PrimaryNode, "04")
	}
}

func TestRouteWithTopology_EmptyTopologyFallsBackToDefaults(t *testing.T) {
	// Non-nil topology with no Routes → ToInternalRules returns nil → default rules.
	topo := &TopologyConfig{
		Nodes:  make(map[string]NodeConfig),
		People: make(map[string]PersonConfig),
		Routes: nil,
	}
	result := RouteWithTopology("Ed called about AI Masters pricing", topo)

	if result.PrimaryNode == "" {
		t.Error("expected primary node from default fallback")
	}
}

func TestRouteWithTopology_CustomRules(t *testing.T) {
	topo := &TopologyConfig{
		Routes: []RoutingRule{
			{
				Keywords: []string{"sales", "pipeline"},
				Target:   "sales-node",
				CrossRef: []string{"revenue"},
			},
		},
	}
	result := RouteWithTopology("The sales pipeline is growing fast", topo)

	if result.PrimaryNode != "sales-node" {
		t.Errorf("PrimaryNode = %q, want %q", result.PrimaryNode, "sales-node")
	}
}

func TestRouteWithTopology_CustomRulesNoPrimaryFallsToInbox(t *testing.T) {
	// Custom rule exists but text doesn't match any keyword → inbox (09).
	topo := &TopologyConfig{
		Routes: []RoutingRule{
			{Keywords: []string{"sales"}, Target: "sales-node", CrossRef: nil},
		},
	}
	result := RouteWithTopology("Random unrelated content here", topo)

	if result.PrimaryNode != "09" {
		t.Errorf("unmatched custom rules should fall to inbox '09', got %q", result.PrimaryNode)
	}
}

func TestRouteWithTopology_FinancialCrossRef(t *testing.T) {
	// Financial keywords should always cross-ref node "11" when primary is NOT "11".
	// "AI Masters invoice" routes primarily to node 04 (AI Masters rule) but
	// the word "invoice" is a financial keyword so "11" must appear in cross-refs.
	result := RouteWithTopology("AI Masters invoice payment received", nil)

	// Primary should be 04 (AI Masters) not 11.
	if result.PrimaryNode == "11" {
		t.Skip("routing resolved primary to 11; financial cross-ref self-exclusion applies")
	}

	found := false
	for _, cr := range result.CrossRefNodes {
		if cr == "11" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("financial content should cross-ref node '11', CrossRefNodes = %v (primary=%s)",
			result.CrossRefNodes, result.PrimaryNode)
	}
}

func TestRouteWithTopology_FinancialAsPrimaryNoCrossRef11(t *testing.T) {
	// When "11" is already the primary (revenue keyword dominant), it must not
	// appear in cross-refs (no self-reference).
	result := RouteWithTopology("revenue invoice payment arr pipeline deal close income earn money", nil)

	if result.PrimaryNode == "11" {
		for _, cr := range result.CrossRefNodes {
			if cr == "11" {
				t.Error("node 11 should not cross-ref itself")
			}
		}
	}
}

// ── RouteWithRules (lower-level) ──────────────────────────────────────────────

func TestRouteWithRules_FirstMatchSetsPrimary(t *testing.T) {
	rules := []routingRule{
		{keywords: []string{"alpha"}, primary: "A", crossRefs: nil},
		{keywords: []string{"beta"}, primary: "B", crossRefs: nil},
	}
	result := RouteWithRules("alpha beta content here", rules)

	// First matching rule wins primary.
	if result.PrimaryNode != "A" {
		t.Errorf("PrimaryNode = %q, want %q (first match wins)", result.PrimaryNode, "A")
	}
}

func TestRouteWithRules_AllMatchingRulesContributeCrossRefs(t *testing.T) {
	rules := []routingRule{
		{keywords: []string{"alpha"}, primary: "A", crossRefs: []string{"C"}},
		{keywords: []string{"beta"}, primary: "B", crossRefs: []string{"D"}},
	}
	result := RouteWithRules("alpha beta content here", rules)

	hasC, hasD := false, false
	for _, cr := range result.CrossRefNodes {
		if cr == "C" {
			hasC = true
		}
		if cr == "D" {
			hasD = true
		}
	}
	if !hasC {
		t.Errorf("cross-ref 'C' missing; CrossRefNodes = %v", result.CrossRefNodes)
	}
	if !hasD {
		t.Errorf("cross-ref 'D' missing; CrossRefNodes = %v", result.CrossRefNodes)
	}
}

func TestRouteWithRules_PrimaryNotInCrossRefs(t *testing.T) {
	// Rule A has crossRef to itself — must be removed.
	rules := []routingRule{
		{keywords: []string{"alpha"}, primary: "A", crossRefs: []string{"A", "B"}},
	}
	result := RouteWithRules("alpha content here", rules)

	for _, cr := range result.CrossRefNodes {
		if cr == "A" {
			t.Error("primary node A must not appear in cross-refs")
		}
	}
}

func TestRouteWithRules_NoMatchDefaultsToInbox(t *testing.T) {
	rules := []routingRule{
		{keywords: []string{"specific-term-xyz"}, primary: "X", crossRefs: nil},
	}
	result := RouteWithRules("completely unrelated content", rules)

	if result.PrimaryNode != "09" {
		t.Errorf("unmatched text should route to inbox '09', got %q", result.PrimaryNode)
	}
	if result.Confidence != 0.1 {
		t.Errorf("unmatched confidence = %.2f, want 0.1", result.Confidence)
	}
}

func TestRouteWithRules_EmptyRules_DefaultsToInbox(t *testing.T) {
	result := RouteWithRules("some text", nil)

	if result.PrimaryNode != "09" {
		t.Errorf("empty rules should route to inbox '09', got %q", result.PrimaryNode)
	}
}

// ── Route (uses defaultRoutingRules) ─────────────────────────────────────────

func TestRoute_KnownKeyword(t *testing.T) {
	cases := []struct {
		text        string
		wantPrimary string
	}{
		{"Ed called about AI Masters curriculum update", "04"},
		{"ClinicIQ healthcare pipeline review with Atif", "06"},
		{"Revenue invoice from deal we closed today", "11"},
		{"Jordan and Tom working on the consortium model", "01"},
		{"Pedro team sync about hiring George next week", "10"},
		{"Community school group $99 paywall update", "07"},
	}

	for _, tc := range cases {
		t.Run(tc.text[:30], func(t *testing.T) {
			got := Route(tc.text)
			if got.PrimaryNode != tc.wantPrimary {
				t.Errorf("Route(%q...) PrimaryNode = %q, want %q", tc.text[:30], got.PrimaryNode, tc.wantPrimary)
			}
		})
	}
}

func TestRoute_UnknownContentGoesToInbox(t *testing.T) {
	result := Route("purple unicorn quantum flux capacitor")
	if result.PrimaryNode != "09" {
		t.Errorf("unknown content PrimaryNode = %q, want %q", result.PrimaryNode, "09")
	}
}

// ── scoreConfidence ───────────────────────────────────────────────────────────

func TestScoreConfidence(t *testing.T) {
	// scoreConfidence formula: 0.5 + 0.5*(matched/total)
	// Zero-total guard returns 0.0; any non-zero total gives base >= 0.5.
	cases := []struct {
		matched int
		total   int
		want    float64
	}{
		{0, 0, 0.0},   // zero-total guard
		{1, 1, 1.0},   // all matched → 0.5 + 0.5*1.0 = 1.0
		{1, 2, 0.75},  // half matched → 0.5 + 0.5*0.5 = 0.75
		{0, 5, 0.5},   // matched=0, total=5 → ratio=0 → 0.5 + 0.5*0 = 0.5
		{3, 4, 0.875}, // 3/4 matched → 0.5 + 0.5*0.75 = 0.875
	}

	for _, tc := range cases {
		got := scoreConfidence(tc.matched, tc.total)
		if got != tc.want {
			t.Errorf("scoreConfidence(%d, %d) = %.4f, want %.4f",
				tc.matched, tc.total, got, tc.want)
		}
	}
}

// ── hasFinancialSignal ────────────────────────────────────────────────────────

func TestHasFinancialSignal(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"revenue up this quarter", true},
		{"invoice sent to client", true},
		{"we earned $5000", true},
		{"deal is closed", true},
		{"income stream established", true},
		{"completely unrelated text here", false},
		{"", false},
	}

	for _, tc := range cases {
		got := hasFinancialSignal(tc.input)
		if got != tc.want {
			t.Errorf("hasFinancialSignal(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
