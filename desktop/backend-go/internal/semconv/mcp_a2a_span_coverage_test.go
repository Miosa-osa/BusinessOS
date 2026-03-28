// Package semconv — MCP and A2A span name coverage regression gate.
//
// Armstrong rule: if a constant is renamed or removed in semconv YAML and
// the Go constants are regenerated, the compile error here is the first
// signal that dead semconv has accumulated.
//
// Run with: cd BusinessOS/desktop/backend-go && go test ./internal/semconv/... -run MCP
//           cd BusinessOS/desktop/backend-go && go test ./internal/semconv/... -run A2A
package semconv

import (
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// MCP named equality tests
// ─────────────────────────────────────────────────────────────────────────────

func TestMcpCallSpanMatchesSemconv(t *testing.T) {
	if McpCallSpan != "mcp.call" {
		t.Errorf("McpCallSpan = %q, want %q", McpCallSpan, "mcp.call")
	}
}

func TestMcpToolExecuteSpanMatchesSemconv(t *testing.T) {
	if McpToolExecuteSpan != "mcp.tool_execute" {
		t.Errorf("McpToolExecuteSpan = %q, want %q", McpToolExecuteSpan, "mcp.tool_execute")
	}
}

func TestMcpConnectionEstablishSpanMatchesSemconv(t *testing.T) {
	if McpConnectionEstablishSpan != "mcp.connection.establish" {
		t.Errorf("McpConnectionEstablishSpan = %q, want %q", McpConnectionEstablishSpan, "mcp.connection.establish")
	}
}

func TestMcpRegistryDiscoverSpanMatchesSemconv(t *testing.T) {
	if McpRegistryDiscoverSpan != "mcp.registry.discover" {
		t.Errorf("McpRegistryDiscoverSpan = %q, want %q", McpRegistryDiscoverSpan, "mcp.registry.discover")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MCP bulk format: all 17 constants non-empty, dotted, mcp. prefix
// ─────────────────────────────────────────────────────────────────────────────

func TestAllMcpSpanConstantsAreNonEmptyDottedStringsWithMcpPrefix(t *testing.T) {
	spans := []string{
		McpCallSpan,
		McpConnectionEstablishSpan,
		McpConnectionPoolAcquireSpan,
		McpRegistryDiscoverSpan,
		McpResourceReadSpan,
		McpServerHealthCheckSpan,
		McpServerMetricsCollectSpan,
		McpToolAnalyticsRecordSpan,
		McpToolCacheLookupSpan,
		McpToolComposeSpan,
		McpToolDeprecateSpan,
		McpToolRetrySpan,
		McpToolTimeoutSpan,
		McpToolValidateSpan,
		McpToolVersionCheckSpan,
		McpToolExecuteSpan,
		McpTransportConnectSpan,
	}
	if len(spans) != 17 {
		t.Errorf("expected 17 MCP span constants, got %d", len(spans))
	}
	for _, span := range spans {
		if span == "" {
			t.Errorf("MCP span name must not be empty")
		}
		if !strings.Contains(span, ".") {
			t.Errorf("MCP span name must contain a dot (dotted namespace): %q", span)
		}
		if !strings.HasPrefix(span, "mcp.") {
			t.Errorf("MCP span name must start with 'mcp.': %q", span)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// A2A named equality tests
// ─────────────────────────────────────────────────────────────────────────────

func TestA2aCallSpanMatchesSemconv(t *testing.T) {
	if A2aCallSpan != "a2a.call" {
		t.Errorf("A2aCallSpan = %q, want %q", A2aCallSpan, "a2a.call")
	}
}

func TestA2aTaskDelegateSpanMatchesSemconv(t *testing.T) {
	if A2aTaskDelegateSpan != "a2a.task.delegate" {
		t.Errorf("A2aTaskDelegateSpan = %q, want %q", A2aTaskDelegateSpan, "a2a.task.delegate")
	}
}

func TestA2aNegotiateSpanMatchesSemconv(t *testing.T) {
	if A2aNegotiateSpan != "a2a.negotiate" {
		t.Errorf("A2aNegotiateSpan = %q, want %q", A2aNegotiateSpan, "a2a.negotiate")
	}
}

func TestA2aCreateDealSpanMatchesSemconv(t *testing.T) {
	if A2aCreateDealSpan != "a2a.create_deal" {
		t.Errorf("A2aCreateDealSpan = %q, want %q", A2aCreateDealSpan, "a2a.create_deal")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// A2A bulk format: all 29 constants non-empty, dotted, a2a. prefix
// ─────────────────────────────────────────────────────────────────────────────

func TestAllA2aSpanConstantsAreNonEmptyDottedStringsWithA2aPrefix(t *testing.T) {
	spans := []string{
		A2aAuctionRunSpan,
		A2aBidEvaluateSpan,
		A2aCallSpan,
		A2aCapabilityMatchSpan,
		A2aCapabilityNegotiateSpan,
		A2aCapabilityRegisterSpan,
		A2aContractAmendSpan,
		A2aContractDisputeSpan,
		A2aContractExecuteSpan,
		A2aContractNegotiateSpan,
		A2aCreateDealSpan,
		A2aDealStatusTransitionSpan,
		A2aDisputeResolveSpan,
		A2aEscrowCreateSpan,
		A2aEscrowReleaseSpan,
		A2aKnowledgeTransferSpan,
		A2aMessageBatchSpan,
		A2aMessageRouteSpan,
		A2aNegotiateSpan,
		A2aNegotiationStateTransitionSpan,
		A2aPenaltyApplySpan,
		A2aProtocolNegotiateSpan,
		A2aReputationDecaySpan,
		A2aReputationUpdateSpan,
		A2aSlaCheckSpan,
		A2aSloEvaluateSpan,
		A2aTaskDelegateSpan,
		A2aTrustEvaluateSpan,
		A2aTrustFederateSpan,
	}
	if len(spans) != 29 {
		t.Errorf("expected 29 A2A span constants, got %d", len(spans))
	}
	for _, span := range spans {
		if span == "" {
			t.Errorf("A2A span name must not be empty")
		}
		if !strings.Contains(span, ".") {
			t.Errorf("A2A span name must contain a dot (dotted namespace): %q", span)
		}
		if !strings.HasPrefix(span, "a2a.") {
			t.Errorf("A2A span name must start with 'a2a.': %q", span)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cross-domain uniqueness: MCP + A2A span names must not overlap
// ─────────────────────────────────────────────────────────────────────────────

func TestMcpAndA2aSpanNamesAreUnique(t *testing.T) {
	mcp := []string{
		McpCallSpan,
		McpToolExecuteSpan,
		McpConnectionEstablishSpan,
		McpRegistryDiscoverSpan,
	}
	a2a := []string{
		A2aCallSpan,
		A2aTaskDelegateSpan,
		A2aNegotiateSpan,
		A2aCapabilityRegisterSpan,
	}

	seen := make(map[string]bool)
	all := append(mcp, a2a...) //nolint:gocritic
	for _, span := range all {
		if seen[span] {
			t.Errorf("duplicate span name across MCP and A2A domains: %q", span)
		}
		seen[span] = true
	}
}
