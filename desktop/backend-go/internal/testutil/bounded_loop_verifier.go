package testutil

import (
	"testing"
)

// AssertMaxIterations enforces WvdA liveness.
// Fails if loop exceeds max_iterations boundary.
//
// Usage:
//
//	count := AssertMaxIterations(t, 100, func() int {
//		i := 0
//		for item := range items {
//			_ = process(item)
//			i++
//		}
//		return i
//	})
//	assert.True(t, count <= 100)
func AssertMaxIterations(t *testing.T, maxIterations int, operation func() int) int {
	t.Helper()

	count := operation()

	if count > maxIterations {
		t.Fatalf("loop exceeded max iterations: %d > %d", count, maxIterations)
	}

	return count
}

// IterationCounter tracks loop iterations and enforces bounds.
type IterationCounter struct {
	count    int
	maxCount int
}

// NewIterationCounter creates bounded loop tracker.
//
// Usage:
//
//	counter := NewIterationCounter(100)
//	for item := range items {
//		if !counter.Increment() {
//			t.Fatal("exceeded max iterations")
//		}
//		process(item)
//	}
func NewIterationCounter(maxCount int) *IterationCounter {
	return &IterationCounter{
		count:    0,
		maxCount: maxCount,
	}
}

// Increment adds one iteration and returns false if limit exceeded.
func (c *IterationCounter) Increment() bool {
	c.count++
	return c.count <= c.maxCount
}

// Count returns current iteration count.
func (c *IterationCounter) Count() int {
	return c.count
}

// IsExhausted returns true if max iterations reached.
func (c *IterationCounter) IsExhausted() bool {
	return c.count >= c.maxCount
}

// AssertRecursionDepth validates recursive function depth.
//
// Usage:
//
//	depth := 0
//	var traverse func(node *Tree)
//	traverse = func(node *Tree) {
//		depth++
//		AssertRecursionDepth(t, 1000, depth)
//		if node.Left != nil {
//			traverse(node.Left)
//		}
//		depth--
//	}
func AssertRecursionDepth(t *testing.T, maxDepth int, currentDepth int) {
	t.Helper()

	if currentDepth > maxDepth {
		t.Fatalf("recursion exceeded max depth: %d > %d", currentDepth, maxDepth)
	}
}
