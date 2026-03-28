package concurrency

import (
	"context"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/observability"
)

// TestNewSemaphore verifies semaphore initialization
func TestNewSemaphore(t *testing.T) {
	tel := observability.New()

	tests := []struct {
		name      string
		maxSlots  int
		expectCap int
	}{
		{
			name:      "default 200 slots",
			maxSlots:  200,
			expectCap: 200,
		},
		{
			name:      "10 slots",
			maxSlots:  10,
			expectCap: 10,
		},
		{
			name:      "invalid slots defaults to 200",
			maxSlots:  -1,
			expectCap: 200,
		},
		{
			name:      "zero slots defaults to 200",
			maxSlots:  0,
			expectCap: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := New(tt.maxSlots, tel)

			if sem.Available() != tt.expectCap {
				t.Errorf("Expected %d available slots, got %d", tt.expectCap, sem.Available())
			}

			if sem.maxSlots != int32(tt.expectCap) {
				t.Errorf("Expected maxSlots %d, got %d", tt.expectCap, sem.maxSlots)
			}
		})
	}
}

// TestAcquireRelease verifies basic acquire/release cycle
func TestAcquireRelease(t *testing.T) {
	tel := observability.New()
	sem := New(10, tel)

	initialAvailable := sem.Available()

	// Acquire a slot
	ctx := context.Background()
	err := sem.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// Verify slot was taken
	if sem.Available() != initialAvailable-1 {
		t.Errorf("Expected %d available after acquire, got %d", initialAvailable-1, sem.Available())
	}

	// Release the slot
	sem.Release()

	// Verify slot was returned
	if sem.Available() != initialAvailable {
		t.Errorf("Expected %d available after release, got %d", initialAvailable, sem.Available())
	}
}

// TestAcquireAllSlots verifies semaphore exhausts all slots
func TestAcquireAllSlots(t *testing.T) {
	tel := observability.New()
	sem := New(5, tel)
	ctx := context.Background()

	// Acquire all slots
	acquired := make([]struct{}, 5)
	for i := 0; i < 5; i++ {
		err := sem.Acquire(ctx)
		if err != nil {
			t.Fatalf("Acquire %d failed: %v", i, err)
		}
		acquired[i] = struct{}{}
	}

	// Verify no slots available
	if sem.Available() != 0 {
		t.Errorf("Expected 0 available slots, got %d", sem.Available())
	}

	// Try to acquire one more (should timeout)
	// Use short timeout for test
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx)
	if err == nil {
		t.Error("Expected timeout error when acquiring from full semaphore, got nil")
	}

	// Release all slots
	for range acquired {
		sem.Release()
	}

	// Verify all slots available
	if sem.Available() != 5 {
		t.Errorf("Expected 5 available slots after releasing all, got %d", sem.Available())
	}
}

// TestConcurrentAcquire verifies concurrent access works correctly
func TestConcurrentAcquire(t *testing.T) {
	tel := observability.New()
	sem := New(10, tel)
	ctx := context.Background()

	// Spawn 20 goroutines trying to acquire 10 slots
	done := make(chan bool, 20)
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			err := sem.Acquire(ctx)
			if err != nil {
				errors <- err
				done <- false
				return
			}

			// Hold the slot briefly
			time.Sleep(10 * time.Millisecond)

			sem.Release()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	successCount := 0
	timeoutCount := 0
	for i := 0; i < 20; i++ {
		select {
		case <-done:
			successCount++
		case err := <-errors:
			if err == context.DeadlineExceeded || err == context.Canceled {
				timeoutCount++
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}

	// All should succeed (with retries via context timeout)
	if successCount != 20 {
		t.Errorf("Expected 20 successful acquires, got %d", successCount)
	}

	// Verify all slots released
	if sem.Available() != 10 {
		t.Errorf("Expected 10 available slots after all goroutines, got %d", sem.Available())
	}
}

// TestUtilization verifies utilization calculation
func TestUtilization(t *testing.T) {
	tel := observability.New()
	sem := New(100, tel)

	// Empty semaphore: 0% utilization
	if sem.Utilization() != 0.0 {
		t.Errorf("Expected 0%% utilization, got %.2f%%", sem.Utilization())
	}

	ctx := context.Background()
	// Acquire 50 slots
	for i := 0; i < 50; i++ {
		sem.Acquire(ctx)
	}

	// 50% utilization
	util := sem.Utilization()
	if util < 49.9 || util > 50.1 {
		t.Errorf("Expected ~50%% utilization, got %.2f%%", util)
	}

	// Release all
	for i := 0; i < 50; i++ {
		sem.Release()
	}

	// Back to 0%
	if sem.Utilization() != 0.0 {
		t.Errorf("Expected 0%% utilization after release, got %.2f%%", sem.Utilization())
	}
}

// TestRejectionRate verifies rejection rate calculation
func TestRejectionRate(t *testing.T) {
	tel := observability.New()
	sem := New(5, tel)
	ctx := context.Background()

	// No requests yet: 0% rejection
	if sem.RejectionRate() != 0.0 {
		t.Errorf("Expected 0%% rejection rate, got %.2f%%", sem.RejectionRate())
	}

	// Acquire all slots
	for i := 0; i < 5; i++ {
		sem.Acquire(ctx)
	}

	// Try to acquire one more (will timeout and be rejected)
	ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	_ = sem.Acquire(ctxTimeout)
	cancel()

	// 1 rejection out of 6 requests = ~16.67%
	rejRate := sem.RejectionRate()
	if rejRate < 16.0 || rejRate > 17.0 {
		t.Errorf("Expected ~16.67%% rejection rate, got %.2f%%", rejRate)
	}
}

// TestGetStats verifies stats snapshot
func TestGetStats(t *testing.T) {
	tel := observability.New()
	sem := New(10, tel)
	ctx := context.Background()

	// Acquire 3 slots
	for i := 0; i < 3; i++ {
		sem.Acquire(ctx)
	}

	stats := sem.GetStats()

	if stats.MaxSlots != 10 {
		t.Errorf("Expected MaxSlots 10, got %d", stats.MaxSlots)
	}

	if stats.Available != 7 {
		t.Errorf("Expected Available 7, got %d", stats.Available)
	}

	util := stats.Utilization
	if util < 29.9 || util > 30.1 {
		t.Errorf("Expected ~30%% utilization, got %.2f%%", util)
	}

	if stats.TotalRequests != 3 {
		t.Errorf("Expected TotalRequests 3, got %d", stats.TotalRequests)
	}

	if stats.TotalAcquired != 3 {
		t.Errorf("Expected TotalAcquired 3, got %d", stats.TotalAcquired)
	}
}

// TestWvdACompliance verifies timeout compliance (WvdA soundness)
func TestWvdACompliance(t *testing.T) {
	tel := observability.New()
	sem := New(1, tel)

	// Acquire the only slot
	ctx := context.Background()
	err := sem.Acquire(ctx)
	if err != nil {
		t.Fatalf("First acquire failed: %v", err)
	}

	// Second acquire should timeout within 5s (WvdA requirement)
	start := time.Now()
	ctxTimeout, cancel := context.WithTimeout(ctx, 6*time.Second) // Slightly longer than 5s to allow timeout
	defer cancel()

	err = sem.Acquire(ctxTimeout)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Verify timeout occurred within expected time (5s acquire timeout + small margin)
	if elapsed < 5*time.Second || elapsed > 6*time.Second {
		t.Errorf("Timeout took %v, expected ~5s", elapsed)
	}

	sem.Release()
}

// BenchmarkAcquireRelease benchmarks acquire/release performance
func BenchmarkAcquireRelease(b *testing.B) {
	tel := observability.New()
	sem := New(1000, tel)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sem.Acquire(ctx)
			sem.Release()
		}
	})
}
