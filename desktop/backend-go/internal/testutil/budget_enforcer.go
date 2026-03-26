package testutil

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BudgetTier defines resource limits for different priority levels.
type BudgetTier string

const (
	Critical BudgetTier = "critical"
	High     BudgetTier = "high"
	Normal   BudgetTier = "normal"
	Low      BudgetTier = "low"
)

// Budget defines time and memory constraints.
type Budget struct {
	MaxDurationMs int
	MaxMemoryMB   int
}

// BudgetTiers maps tier names to resource limits.
var BudgetTiers = map[BudgetTier]Budget{
	Critical: {MaxDurationMs: 100, MaxMemoryMB: 50},
	High:     {MaxDurationMs: 500, MaxMemoryMB: 200},
	Normal:   {MaxDurationMs: 5000, MaxMemoryMB: 1000},
	Low:      {MaxDurationMs: 30000, MaxMemoryMB: 5000},
}

// AssertWithinBudget enforces Armstrong budget constraints.
//
// Usage:
//
//	AssertWithinBudget(t, Budget{MaxDurationMs: 100, MaxMemoryMB: 50}, func() {
//		result := fastOperation()
//		assert.NotNil(t, result)
//	})
func AssertWithinBudget(t *testing.T, budget Budget, operation func()) {
	t.Helper()

	if budget.MaxDurationMs == 0 {
		t.Fatal("MaxDurationMs is required (Armstrong budget constraint)")
	}

	startTime := time.Now()
	startMem := getMemoryMB()

	operation()

	elapsed := time.Since(startTime)
	memUsed := getMemoryMB() - startMem

	if elapsed.Milliseconds() > int64(budget.MaxDurationMs) {
		t.Fatalf("operation exceeded time budget: %vms > %dms",
			elapsed.Milliseconds(), budget.MaxDurationMs)
	}

	if budget.MaxMemoryMB > 0 && int(memUsed) > budget.MaxMemoryMB {
		t.Fatalf("operation exceeded memory budget: %dmb > %dmb",
			int(memUsed), budget.MaxMemoryMB)
	}
}

// AssertTierCompliant enforces tier-based resource limits.
//
// Usage:
//
//	AssertTierCompliant(t, Critical, func() {
//		result := criticalPathOperation()
//		assert.NotNil(t, result)
//	})
func AssertTierCompliant(t *testing.T, tier BudgetTier, operation func()) {
	t.Helper()

	budget, ok := BudgetTiers[tier]
	if !ok {
		t.Fatalf("unknown budget tier: %s", tier)
	}

	AssertWithinBudget(t, budget, operation)
}

// BudgetMonitor tracks resource usage across operations.
type BudgetMonitor struct {
	tier   BudgetTier
	ops    []OpMetrics
	budget Budget
}

// OpMetrics records operation resource usage.
type OpMetrics struct {
	Name       string
	DurationMs int64
	MemoryMB   int
}

// NewBudgetMonitor creates resource tracker for a tier.
func NewBudgetMonitor(tier BudgetTier) *BudgetMonitor {
	return &BudgetMonitor{
		tier:   tier,
		ops:    []OpMetrics{},
		budget: BudgetTiers[tier],
	}
}

// MeasureOperation tracks resource usage of an operation.
func (bm *BudgetMonitor) MeasureOperation(name string, operation func()) *OpMetrics {
	startTime := time.Now()
	startMem := getMemoryMB()

	operation()

	elapsed := time.Since(startTime)
	memUsed := int(getMemoryMB() - startMem)

	metrics := OpMetrics{
		Name:       name,
		DurationMs: elapsed.Milliseconds(),
		MemoryMB:   memUsed,
	}

	bm.ops = append(bm.ops, metrics)
	return &metrics
}

// AssertAllCompliant verifies all tracked operations respect tier limits.
func (bm *BudgetMonitor) AssertAllCompliant(t *testing.T) {
	t.Helper()

	for _, op := range bm.ops {
		if op.DurationMs > int64(bm.budget.MaxDurationMs) {
			t.Errorf("operation %q exceeded time budget: %dms > %dms",
				op.Name, op.DurationMs, bm.budget.MaxDurationMs)
		}
		if op.MemoryMB > bm.budget.MaxMemoryMB {
			t.Errorf("operation %q exceeded memory budget: %dmb > %dmb",
				op.Name, op.MemoryMB, bm.budget.MaxMemoryMB)
		}
	}

	if t.Failed() {
		t.FailNow()
	}
}

// Summary returns formatted metrics report.
func (bm *BudgetMonitor) Summary() string {
	summary := fmt.Sprintf("=== Budget Monitor: %s ===\n", bm.tier)
	for _, op := range bm.ops {
		summary += fmt.Sprintf(
			"%s: %dms, %dmb\n",
			op.Name, op.DurationMs, op.MemoryMB,
		)
	}
	return summary
}

func getMemoryMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1_000_000
}
