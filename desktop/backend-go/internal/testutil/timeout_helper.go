package testutil

import (
	"context"
	"testing"
	"time"
)

// AssertCompletesWithTimeout enforces WvdA deadlock-freedom.
// Fails fast if operation exceeds timeout or timeout is missing.
//
// Usage:
//
//	AssertCompletesWithTimeout(t, 5*time.Second, func(ctx context.Context) {
//		result, _ := service.QueryWithContext(ctx)
//		assert.NotNil(t, result)
//	})
func AssertCompletesWithTimeout(t *testing.T, timeout time.Duration, operation func(context.Context)) {
	t.Helper()

	if timeout == 0 {
		t.Fatal("timeout is required (WvdA deadlock-freedom constraint)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	done := make(chan struct{})
	go func() {
		operation(ctx)
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(start)
		if elapsed > timeout {
			t.Fatalf("operation exceeded timeout: %v > %v", elapsed, timeout)
		}
	case <-ctx.Done():
		t.Fatalf("operation timed out after %v", timeout)
	}
}

// AssertCompletesWithinDuration verifies operation duration is bounded.
//
// Usage:
//
//	AssertCompletesWithinDuration(t, 100*time.Millisecond, func() {
//		result := fastOperation()
//		assert.NotNil(t, result)
//	})
func AssertCompletesWithinDuration(t *testing.T, maxDuration time.Duration, operation func()) {
	t.Helper()

	start := time.Now()
	operation()
	elapsed := time.Since(start)

	if elapsed > maxDuration {
		t.Fatalf("operation exceeded duration: %v > %v", elapsed, maxDuration)
	}
}
