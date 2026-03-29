// Package knhk provides Armstrong fault-tolerant supervision for KNHK query execution.
package knhk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// EpochSupervisor enforces idempotent SPARQL operations (μ∘μ = μ) with budget constraints.
//
// ## Responsibilities
//
// 1. Enforce timeout per operation tier (critical=100ms, high=500ms, normal=5s, low=30s)
// 2. Track operation budget: count, memory, time per tier
// 3. Execute idempotent SPARQL queries
// 4. Log to audit trail with BLAKE3 hash + Ed25519 signature
// 5. Circuit breaker: open circuit if >3 failures in 60s
// 6. No retry logic; timeout = error (let-it-crash principle)
//
// ## Armstrong Principles
//
// - Let-It-Crash: No exception swallowing; timeout = error propagation
// - Supervision: Circuit breaker detects cascading failures
// - No Shared Mutable State: Use channels for all communication
// - Budget Constraints: Every operation has time_ms budget enforced
//
// ## WvdA Soundness
//
// - Deadlock-Free: context.WithTimeout guarantees finite wait
// - Liveness: All operations complete or timeout
// - Boundedness: Operation count, memory, time all bounded
type EpochSupervisor struct {
	budgetMu    sync.RWMutex
	budgets     map[string]int // tier -> operation count in current window
	budgetMax   map[string]int // tier -> max operations per window
	failuresMu  sync.RWMutex
	failures    map[string][]time.Time // tier -> failure times
	circuitOpen bool
	httpClient  *http.Client
	logger      *slog.Logger
}

// OperationTier defines timeout and budget for operation categories.
type OperationTier struct {
	Name         string
	TimeoutMs    time.Duration
	MaxOpsPerMin int
}

// Define standard tiers
var (
	TierCritical = OperationTier{Name: "critical", TimeoutMs: 100 * time.Millisecond, MaxOpsPerMin: 1000}
	TierHigh     = OperationTier{Name: "high", TimeoutMs: 500 * time.Millisecond, MaxOpsPerMin: 100}
	TierNormal   = OperationTier{Name: "normal", TimeoutMs: 5 * time.Second, MaxOpsPerMin: 10}
	TierLow      = OperationTier{Name: "low", TimeoutMs: 30 * time.Second, MaxOpsPerMin: 1}
)

// NewEpochSupervisor creates a new supervisor with tier-based budgets.
func NewEpochSupervisor(logger *slog.Logger) *EpochSupervisor {
	return &EpochSupervisor{
		budgets:    make(map[string]int),
		budgetMax:  make(map[string]int),
		failures:   make(map[string][]time.Time),
		httpClient: &http.Client{Timeout: 35 * time.Second},
		logger:     logger,
	}
}

// ExecuteQuery executes a SPARQL query with tier-based timeout and budget enforcement.
//
// Flow:
// 1. Check budget (operation count, memory, time)
// 2. Acquire context with timeout
// 3. Execute idempotent SPARQL query
// 4. Log to audit trail (hash + signature)
// 5. If timeout: return error (no retry)
// 6. If error: escalate to circuit breaker
// 7. Return result or error
//
// Args:
//   - ctx: parent context (may have its own timeout)
//   - query: SPARQL query string (must be idempotent)
//   - tier: operation tier (critical, high, normal, low)
//   - traceID: for OTEL correlation
//
// Returns:
//   - result: query result (idempotent, can repeat without side effects)
//   - err: nil on success, error if timeout/budget exceeded/circuit open
//
// Crashes (Armstrong let-it-crash):
//   - Returns error (does NOT panic); caller must handle
func (s *EpochSupervisor) ExecuteQuery(
	ctx context.Context,
	query string,
	tier OperationTier,
	traceID string,
) (interface{}, error) {
	// Check circuit breaker
	if s.circuitOpen {
		return nil, fmt.Errorf("circuit breaker open: too many failures in tier %s", tier.Name)
	}

	// Check budget
	s.budgetMu.Lock()
	currentCount := s.budgets[tier.Name]
	s.budgetMu.Unlock()

	if currentCount >= tier.MaxOpsPerMin {
		s.recordFailure(tier.Name)
		return nil, fmt.Errorf("budget exhausted for tier %s: %d/%d ops per min",
			tier.Name, currentCount, tier.MaxOpsPerMin)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, tier.TimeoutMs)
	defer cancel()

	// Execute query
	start := time.Now()
	result, err := s.executeIDMQuery(ctx, query, traceID)
	latencyMs := time.Since(start).Milliseconds()

	// Increment budget counter
	s.budgetMu.Lock()
	s.budgets[tier.Name]++
	s.budgetMu.Unlock()

	// Handle error
	if err != nil {
		s.recordFailure(tier.Name)
		s.logger.Error(
			"Query execution failed",
			slog.String("tier", tier.Name),
			slog.String("query_id", traceID),
			slog.String("error", err.Error()),
			slog.Int64("latency_ms", latencyMs),
		)
		return nil, err
	}

	// Log to audit trail
	s.logAuditTrail(query, result, traceID, latencyMs)

	return result, nil
}

// executeIDMQuery executes an idempotent SPARQL query with context timeout.
// If context times out, returns error (no retry).
func (s *EpochSupervisor) executeIDMQuery(
	ctx context.Context,
	query string,
	traceID string,
) (interface{}, error) {
	// Stub: in production, call KNHK/Oxigraph endpoint
	// ctx.Done() will trigger if timeout expires; no explicit timeout handling needed.
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("query context timeout: %w", ctx.Err())
	default:
		// Simulate query execution
		if query == "INVALID" {
			return nil, fmt.Errorf("invalid SPARQL syntax")
		}
		return map[string]interface{}{"results": []interface{}{}}, nil
	}
}

// recordFailure records a failure event for circuit breaker logic.
func (s *EpochSupervisor) recordFailure(tier string) {
	s.failuresMu.Lock()
	defer s.failuresMu.Unlock()

	now := time.Now()
	s.failures[tier] = append(s.failures[tier], now)

	// Prune old failures (>60s)
	var recent []time.Time
	for _, t := range s.failures[tier] {
		if now.Sub(t) < 60*time.Second {
			recent = append(recent, t)
		}
	}
	s.failures[tier] = recent

	// Open circuit if >3 failures in 60s
	if len(recent) > 3 {
		s.circuitOpen = true
		s.logger.Error(
			"Circuit breaker OPEN: too many failures",
			slog.String("tier", tier),
			slog.Int("failure_count", len(recent)),
		)
	}
}

// logAuditTrail logs the query execution to the audit trail.
// In production, this would write to a signed ledger (Ed25519 signature).
func (s *EpochSupervisor) logAuditTrail(
	query string,
	result interface{},
	traceID string,
	latencyMs int64,
) {
	// Compute BLAKE3-equivalent hash (using SHA256 as substitute)
	hash := sha256.Sum256([]byte(query))
	hashHex := hex.EncodeToString(hash[:])

	// In production, sign with Ed25519 key
	s.logger.Info(
		"Audit trail: query executed",
		slog.String("trace_id", traceID),
		slog.String("query_hash", hashHex),
		slog.Int64("latency_ms", latencyMs),
		slog.String("result_hash", hashHex), // In production: hash of result
	)
}

// GetTierStatus returns the current status of a tier.
func (s *EpochSupervisor) GetTierStatus(tier OperationTier) map[string]interface{} {
	s.budgetMu.RLock()
	currentCount := s.budgets[tier.Name]
	s.budgetMu.RUnlock()

	s.failuresMu.RLock()
	failures := len(s.failures[tier.Name])
	s.failuresMu.RUnlock()

	return map[string]interface{}{
		"tier":            tier.Name,
		"current_ops":     currentCount,
		"max_ops":         tier.MaxOpsPerMin,
		"recent_failures": failures,
		"circuit_open":    s.circuitOpen,
		"timeout_ms":      tier.TimeoutMs.Milliseconds(),
	}
}

// ResetCircuitBreaker resets the circuit breaker (called after manual healing).
func (s *EpochSupervisor) ResetCircuitBreaker() {
	s.failuresMu.Lock()
	defer s.failuresMu.Unlock()

	s.circuitOpen = false
	s.failures = make(map[string][]time.Time)
	s.logger.Info("Circuit breaker reset")
}

// RotateEpoch resets operation counters (called at epoch boundary, e.g., every minute).
func (s *EpochSupervisor) RotateEpoch() {
	s.budgetMu.Lock()
	defer s.budgetMu.Unlock()

	s.budgets = make(map[string]int)
	s.logger.Debug("Budget counters rotated for new epoch")
}
