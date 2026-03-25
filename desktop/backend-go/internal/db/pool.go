package db

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Connection interface for pool-managed connections
type Connection interface {
	ID() string
	Close() error
	IsValid() bool
}

// PoolConfig holds configuration for adaptive connection pool
type PoolConfig struct {
	MinSize        int
	MaxSize        int
	IdleTimeout    time.Duration
	AcquireTimeout time.Duration
}

// PoolStats holds pool statistics
type PoolStats struct {
	Active     int
	Idle       int
	Total      int
	Recycled   int64
	AvgWaitMs  float64
	WaitTimes  []int64
}

// AdaptivePool manages database connections with adaptive sizing
type AdaptivePool struct {
	config      *PoolConfig
	available   chan Connection
	allConns    map[string]Connection
	mutex       sync.RWMutex
	closed      bool
	active      atomic.Int64
	recycled    atomic.Int64
	waitTimes   []int64
	waitMutex   sync.Mutex
	idleTimer   *time.Ticker
	lastResize  time.Time
}

// NewAdaptivePool creates a new adaptive connection pool
func NewAdaptivePool(config *PoolConfig) *AdaptivePool {
	if config.MinSize <= 0 {
		config.MinSize = 10
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.AcquireTimeout == 0 {
		config.AcquireTimeout = 30 * time.Second
	}

	pool := &AdaptivePool{
		config:     config,
		available:  make(chan Connection, config.MaxSize),
		allConns:   make(map[string]Connection),
		waitTimes:  make([]int64, 0),
		lastResize: time.Now(),
	}

	// Initialize with minimum pool size
	for i := 0; i < config.MinSize; i++ {
		conn := createMockConnection(fmt.Sprintf("conn-%d", i))
		pool.allConns[conn.ID()] = conn
		pool.available <- conn
	}

	// Start idle timeout handler
	pool.idleTimer = time.NewTicker(1 * time.Minute)
	go pool.handleIdleTimeout()

	return pool
}

// Acquire gets a connection from the pool
func (p *AdaptivePool) Acquire(ctx context.Context) (Connection, error) {
	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		return nil, fmt.Errorf("pool is closed")
	}
	p.mutex.RUnlock()

	start := time.Now()

	// Try non-blocking first
	select {
	case conn := <-p.available:
		elapsed := time.Since(start).Milliseconds()
		p.recordWaitTime(elapsed)
		if conn.IsValid() {
			p.active.Add(1)
			return conn, nil
		}
		// Invalid connection, discard and try again
		return p.Acquire(ctx)
	default:
		// Pool exhausted, try to scale up
		if p.canScale() {
			elapsed := time.Since(start).Milliseconds()
			p.recordWaitTime(elapsed)
			conn := createMockConnection(fmt.Sprintf("conn-%d", len(p.allConns)))
			p.mutex.Lock()
			p.allConns[conn.ID()] = conn
			p.mutex.Unlock()
			p.active.Add(1)
			return conn, nil
		}

		// Wait with timeout
		select {
		case conn := <-p.available:
			elapsed := time.Since(start).Milliseconds()
			p.recordWaitTime(elapsed)
			if conn.IsValid() {
				p.active.Add(1)
				return conn, nil
			}
			return p.Acquire(ctx)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(p.config.AcquireTimeout):
			return nil, context.DeadlineExceeded
		}
	}
}

// Release returns a connection to the pool
func (p *AdaptivePool) Release(conn Connection) {
	if conn == nil {
		return
	}

	p.active.Add(-1)

	p.mutex.RLock()
	if p.closed {
		p.mutex.RUnlock()
		conn.Close()
		return
	}
	p.mutex.RUnlock()

	if !conn.IsValid() {
		conn.Close()
		return
	}

	// Try to return to pool
	select {
	case p.available <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection and record recycling
		conn.Close()
		p.recycled.Add(1)
	}
}

// Utilization returns percentage of pool in use
func (p *AdaptivePool) Utilization() float64 {
	p.mutex.RLock()
	total := len(p.allConns)
	p.mutex.RUnlock()

	active := p.active.Load()
	if total == 0 {
		return 0
	}
	return float64(active) / float64(total) * 100
}

// Size returns current pool size
func (p *AdaptivePool) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.allConns)
}

// Stats returns pool statistics
func (p *AdaptivePool) Stats() PoolStats {
	p.mutex.RLock()
	total := len(p.allConns)
	active := int(p.active.Load())
	p.mutex.RUnlock()

	idle := total - active
	if idle < 0 {
		idle = 0
	}

	avgWait := 0.0
	p.waitMutex.Lock()
	if len(p.waitTimes) > 0 {
		sum := int64(0)
		for _, t := range p.waitTimes {
			sum += t
		}
		avgWait = float64(sum) / float64(len(p.waitTimes))
	}
	p.waitMutex.Unlock()

	return PoolStats{
		Active:    active,
		Idle:      idle,
		Total:     total,
		Recycled:  p.recycled.Load(),
		AvgWaitMs: avgWait,
	}
}

// Close closes all connections in the pool
func (p *AdaptivePool) Close() error {
	p.mutex.Lock()
	if p.closed {
		p.mutex.Unlock()
		return nil
	}
	p.closed = true
	p.mutex.Unlock()

	if p.idleTimer != nil {
		p.idleTimer.Stop()
	}

	close(p.available)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.allConns {
		conn.Close()
	}
	p.allConns = make(map[string]Connection)

	return nil
}

// Private helpers

func (p *AdaptivePool) canScale() bool {
	p.mutex.RLock()
	size := len(p.allConns)
	p.mutex.RUnlock()
	return size < p.config.MaxSize
}

func (p *AdaptivePool) recordWaitTime(ms int64) {
	p.waitMutex.Lock()
	defer p.waitMutex.Unlock()

	p.waitTimes = append(p.waitTimes, ms)

	// Keep only last 1000 measurements
	if len(p.waitTimes) > 1000 {
		p.waitTimes = p.waitTimes[len(p.waitTimes)-1000:]
	}
}

func (p *AdaptivePool) handleIdleTimeout() {
	for range p.idleTimer.C {
		if time.Since(p.lastResize) < 5*time.Minute {
			continue
		}

		util := p.Utilization()

		// Scale down if underutilized
		if util < 30 {
			p.scaleDown()
		}

		p.lastResize = time.Now()
	}
}

func (p *AdaptivePool) scaleDown() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	currentSize := len(p.allConns)
	targetSize := p.config.MinSize

	if currentSize <= targetSize {
		return
	}

	// Remove excess connections
	toRemove := currentSize - targetSize
	removed := 0

	for id, conn := range p.allConns {
		if removed >= toRemove {
			break
		}

		select {
		case <-p.available:
			conn.Close()
			delete(p.allConns, id)
			removed++
		default:
			// Connection in use, skip
		}
	}
}

// Mock connection for testing
type mockConnection struct {
	id string
}

func createMockConnection(id string) Connection {
	return &mockConnection{id: id}
}

func (m *mockConnection) ID() string {
	return m.id
}

func (m *mockConnection) Close() error {
	return nil
}

func (m *mockConnection) IsValid() bool {
	return true
}
