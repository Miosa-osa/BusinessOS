package db

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolScaling(t *testing.T) {
	p := NewAdaptivePool(&PoolConfig{
		MinSize: 10,
		MaxSize: 100,
		IdleTimeout: 5 * time.Minute,
	})
	defer p.Close()

	// Simulate load increase
	conns := make([]Connection, 0)
	for i := 0; i < 30; i++ {
		conn, err := p.Acquire(context.Background())
		if err != nil {
			t.Fatalf("failed to acquire connection: %v", err)
		}
		conns = append(conns, conn)
	}

	// Check pool size scaled up
	size := p.Size()
	if size < 30 {
		t.Errorf("expected pool size >= 30, got %d", size)
	}

	// Check utilization is high while holding connections
	util := p.Utilization()
	if util < 60 {
		t.Errorf("expected high utilization (>60%%), got %.2f%%", util)
	}

	// Release connections
	for _, conn := range conns {
		p.Release(conn)
	}
}

func TestConnectionReuse(t *testing.T) {
	p := NewAdaptivePool(&PoolConfig{
		MinSize: 5,
		MaxSize: 10,
		IdleTimeout: 1 * time.Minute,
	})
	defer p.Close()

	// Acquire 5 connections sequentially and track them
	acquired := make(map[string]bool)

	for i := 0; i < 5; i++ {
		conn, err := p.Acquire(context.Background())
		if err != nil {
			t.Fatalf("failed to acquire connection: %v", err)
		}
		acquired[conn.ID()] = true
		p.Release(conn)
	}

	// All 5 sequential acquisitions should have used the same 5 initial pool connections
	if len(acquired) > 5 {
		t.Errorf("expected at most 5 unique connections from reuse, got %d", len(acquired))
	}

	// Now acquire all at once to measure pool growth
	conns := make([]Connection, 0)
	for i := 0; i < 10; i++ {
		conn, err := p.Acquire(context.Background())
		if err != nil {
			t.Fatalf("failed to acquire connection: %v", err)
		}
		conns = append(conns, conn)
	}

	// Pool should have scaled to 10
	finalSize := p.Size()
	if finalSize < 10 {
		t.Errorf("expected pool to scale to 10, got %d", finalSize)
	}

	// Release all
	for _, conn := range conns {
		p.Release(conn)
	}
}

func TestTimeoutHandling(t *testing.T) {
	p := NewAdaptivePool(&PoolConfig{
		MinSize: 5,
		MaxSize: 10,
		IdleTimeout: 500 * time.Millisecond,
		AcquireTimeout: 100 * time.Millisecond,
	})
	defer p.Close()

	// Acquire all available connections
	conns := make([]Connection, 0)
	for i := 0; i < 10; i++ {
		conn, err := p.Acquire(context.Background())
		if err != nil {
			t.Fatalf("failed to acquire connection: %v", err)
		}
		conns = append(conns, conn)
	}

	// Try to acquire with timeout - should fail
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := p.Acquire(ctx)
	if err == nil || err != context.DeadlineExceeded {
		t.Errorf("expected context deadline exceeded, got %v", err)
	}

	// Release connections
	for _, conn := range conns {
		p.Release(conn)
	}
}

func TestConcurrentAccess(t *testing.T) {
	p := NewAdaptivePool(&PoolConfig{
		MinSize: 10,
		MaxSize: 100,
		IdleTimeout: 1 * time.Minute,
	})
	defer p.Close()

	successCount := atomic.Int64{}
	failureCount := atomic.Int64{}
	var wg sync.WaitGroup

	// Spawn 50 concurrent goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := p.Acquire(context.Background())
			if err != nil {
				failureCount.Add(1)
				return
			}

			successCount.Add(1)
			time.Sleep(10 * time.Millisecond)
			p.Release(conn)
		}()
	}

	wg.Wait()

	if failureCount.Load() > 5 {
		t.Errorf("expected < 5 failures, got %d", failureCount.Load())
	}
	if successCount.Load() < 45 {
		t.Errorf("expected >= 45 successes, got %d", successCount.Load())
	}
}
