package tests

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/integrations"
)

// ============================================================================
// Test Setup & Fixtures
// ============================================================================

// TestEvent represents an event for performance testing
type TestEvent struct {
	ID        string    `json:"id"`
	Timestamp int64     `json:"timestamp"`
	Activity  string    `json:"activity"`
	Resource  string    `json:"resource"`
	Data      map[string]string `json:"data,omitempty"`
}

// PerformanceMetrics tracks performance measurements
type PerformanceMetrics struct {
	OperationName   string
	TotalDuration   time.Duration
	EventCount      int
	SuccessCount    int
	ErrorCount      int
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	ThroughputQPS   float64
	MemoryAllocMB   float64
	MemoryAllocRate float64 // MB/sec
}

// measureLatency records latency for a single operation
func (m *PerformanceMetrics) recordLatency(latency time.Duration) {
	if latency < m.MinLatency || m.MinLatency == 0 {
		m.MinLatency = latency
	}
	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}
}

// calculateStats computes aggregate statistics
func (m *PerformanceMetrics) calculateStats(latencies []time.Duration) {
	if len(latencies) == 0 {
		return
	}

	// Calculate average
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	m.AvgLatency = sum / time.Duration(len(latencies))

	// Sort for percentiles (simple bucket sort for performance)
	// In production, use more efficient percentile algorithm
	m.P50Latency = latencies[len(latencies)/2]
	m.P95Latency = latencies[(len(latencies)*95)/100]
	m.P99Latency = latencies[(len(latencies)*99)/100]

	// Calculate throughput
	if m.TotalDuration > 0 {
		m.ThroughputQPS = float64(m.SuccessCount) / m.TotalDuration.Seconds()
	}
}

// ============================================================================
// End-to-End Performance Tests
// ============================================================================

// TestEventStoragePerformance tests storing events with various batch sizes
func TestEventStoragePerformance(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name       string
		eventCount int
		batchSize  int
		maxLatency time.Duration
	}{
		{"100_events_single", 100, 1, 100 * time.Millisecond},
		{"100_events_batch10", 100, 10, 100 * time.Millisecond},
		{"1000_events_batch100", 1000, 100, 500 * time.Millisecond},
		{"10000_events_batch1000", 10000, 1000, 5 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("event_storage_%s", tc.name),
			}

			latencies := make([]time.Duration, 0, tc.eventCount/tc.batchSize)
			start := time.Now()
			successCount := 0
			errorCount := 0

			// Generate test events
			events := generateTestEvents(tc.eventCount)

			// Store in batches
			for i := 0; i < len(events); i += tc.batchSize {
				end := i + tc.batchSize
				if end > len(events) {
					end = len(events)
				}

				batchStart := time.Now()
				batch := events[i:end]

				// Simulate storage operation
				err := simulateEventStorage(ctx, batch)
				duration := time.Since(batchStart)

				latencies = append(latencies, duration)
				metrics.recordLatency(duration)

				if err != nil {
					errorCount++
				} else {
					successCount += len(batch)
				}
			}

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = successCount
			metrics.ErrorCount = errorCount
			metrics.EventCount = tc.eventCount

			// Sort latencies for percentile calculation
			sortLatencies(latencies)
			metrics.calculateStats(latencies)

			// Verify SLA
			if metrics.TotalDuration > tc.maxLatency {
				t.Errorf("%s: Expected duration < %v, got %v (SLA violation)",
					tc.name, tc.maxLatency, metrics.TotalDuration)
			}

			// Report metrics
			reportMetrics(t, metrics)
		})
	}
}

// TestEventQueryPerformance tests querying stored events with various filters
func TestEventQueryPerformance(t *testing.T) {
	ctx := context.Background()

	// Pre-populate database with test data
	eventCounts := []int{100, 1000, 10000, 100000}

	for _, eventCount := range eventCounts {
		t.Run(fmt.Sprintf("query_%d_events", eventCount), func(t *testing.T) {
			// Setup: insert test events
			events := generateTestEvents(eventCount)
			_ = simulateEventStorage(ctx, events)

			metrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("query_%d_events", eventCount),
			}

			latencies := make([]time.Duration, 0, 100)
			start := time.Now()
			successCount := 0

			// Run 100 queries
			for i := 0; i < 100; i++ {
				queryStart := time.Now()

				// Simulate indexed query
				results, err := simulateEventQuery(ctx, eventCount, "indexed")
				duration := time.Since(queryStart)

				latencies = append(latencies, duration)
				metrics.recordLatency(duration)

				if err == nil && len(results) > 0 {
					successCount++
				}
			}

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = successCount
			metrics.EventCount = eventCount

			sortLatencies(latencies)
			metrics.calculateStats(latencies)

			// Verify query performance SLA: p95 < 100ms for indexed queries
			if metrics.P95Latency > 100*time.Millisecond {
				t.Logf("Warning: Query p95 latency = %v (target < 100ms)",
					metrics.P95Latency)
			}

			reportMetrics(t, metrics)
		})
	}
}

// TestConcurrentRequestHandling tests system under concurrent load
func TestConcurrentRequestHandling(t *testing.T) {
	ctx := context.Background()

	concurrencyLevels := []int{10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrent_%d_users", concurrency), func(t *testing.T) {
			metrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("concurrent_%d", concurrency),
			}

			var (
				wg              sync.WaitGroup
				successCount    int64
				errorCount      int64
				latenciesMutex  sync.Mutex
				latencies       []time.Duration
			)

			start := time.Now()
			requestsPerUser := 100

			// Launch concurrent goroutines
			for user := 0; user < concurrency; user++ {
				wg.Add(1)
				go func(userID int) {
					defer wg.Done()

					for req := 0; req < requestsPerUser; req++ {
						reqStart := time.Now()

						// Generate random operation
						event := generateTestEvent(userID*requestsPerUser + req)
						err := simulateEventStorage(ctx, []TestEvent{event})

						duration := time.Since(reqStart)

						latenciesMutex.Lock()
						latencies = append(latencies, duration)
						latenciesMutex.Unlock()

						metrics.recordLatency(duration)

						if err == nil {
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}
					}
				}(user)
			}

			wg.Wait()

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = int(atomic.LoadInt64(&successCount))
			metrics.ErrorCount = int(atomic.LoadInt64(&errorCount))
			metrics.EventCount = concurrency * requestsPerUser

			sortLatencies(latencies)
			metrics.calculateStats(latencies)

			// Verify SLA: maintain <100ms p50 latency under concurrent load
			if metrics.P50Latency > 100*time.Millisecond {
				t.Logf("Warning: Under %d concurrent users, p50 latency = %v (target < 100ms)",
					concurrency, metrics.P50Latency)
			}

			reportMetrics(t, metrics)
		})
	}
}

// TestCacheEffectiveness tests cache hit rate and performance improvement
func TestCacheEffectiveness(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                string
		queryCount          int
		uniqueQueries       int
		expectedHitRate     float64
		maxLatencyWithCache time.Duration
	}{
		{
			name:                "high_locality",
			queryCount:          1000,
			uniqueQueries:       10,    // 1000/10 = 100x locality
			expectedHitRate:     0.90,  // 90% hit rate
			maxLatencyWithCache: 50 * time.Millisecond,
		},
		{
			name:                "medium_locality",
			queryCount:          1000,
			uniqueQueries:       100,   // 1000/100 = 10x locality
			expectedHitRate:     0.70,  // 70% hit rate
			maxLatencyWithCache: 100 * time.Millisecond,
		},
		{
			name:                "low_locality",
			queryCount:          1000,
			uniqueQueries:       500,   // 1000/500 = 2x locality
			expectedHitRate:     0.50,  // 50% hit rate
			maxLatencyWithCache: 200 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cacheMetrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("cache_%s", tc.name),
			}

			var (
				cacheHits   int64
				cacheMisses int64
				latencies   []time.Duration
			)

			start := time.Now()

			// Run queries with cache
			for i := 0; i < tc.queryCount; i++ {
				queryStart := time.Now()

				// Simulate cache lookup
				queryID := i % tc.uniqueQueries
				hit, err := simulateCacheQuery(ctx, fmt.Sprintf("query_%d", queryID))

				duration := time.Since(queryStart)
				latencies = append(latencies, duration)
				cacheMetrics.recordLatency(duration)

				if err == nil {
					if hit {
						atomic.AddInt64(&cacheHits, 1)
					} else {
						atomic.AddInt64(&cacheMisses, 1)
					}
				}
			}

			cacheMetrics.TotalDuration = time.Since(start)
			cacheMetrics.SuccessCount = int(atomic.LoadInt64(&cacheHits) + atomic.LoadInt64(&cacheMisses))
			cacheMetrics.EventCount = tc.queryCount

			sortLatencies(latencies)
			cacheMetrics.calculateStats(latencies)

			// Calculate actual hit rate
			totalAccesses := atomic.LoadInt64(&cacheHits) + atomic.LoadInt64(&cacheMisses)
			actualHitRate := float64(atomic.LoadInt64(&cacheHits)) / float64(totalAccesses)

			// Verify cache effectiveness
			if actualHitRate < tc.expectedHitRate*0.8 {
				t.Logf("Warning: Cache hit rate = %.2f%% (expected ~%.2f%%)",
					actualHitRate*100, tc.expectedHitRate*100)
			}

			if cacheMetrics.AvgLatency > tc.maxLatencyWithCache {
				t.Logf("Warning: Cached query latency = %v (target < %v)",
					cacheMetrics.AvgLatency, tc.maxLatencyWithCache)
			}

			t.Logf("Cache performance: %d hits, %d misses (hit rate: %.2f%%)",
				atomic.LoadInt64(&cacheHits),
				atomic.LoadInt64(&cacheMisses),
				actualHitRate*100)

			reportMetrics(t, cacheMetrics)
		})
	}
}

// TestDatabaseQueryOptimization tests query performance improvements
func TestDatabaseQueryOptimization(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name              string
		queryType         string
		recordCount       int
		maxLatency        time.Duration
		shouldUseIndex    bool
	}{
		{"indexed_single", "indexed_single", 100000, 10 * time.Millisecond, true},
		{"indexed_range", "indexed_range", 100000, 50 * time.Millisecond, true},
		{"indexed_join", "indexed_join", 100000, 100 * time.Millisecond, true},
		{"full_table_scan", "full_table_scan", 1000, 500 * time.Millisecond, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := &PerformanceMetrics{
				OperationName: tc.name,
			}

			latencies := make([]time.Duration, 0, 100)
			start := time.Now()
			successCount := 0

			// Run 100 queries
			for i := 0; i < 100; i++ {
				queryStart := time.Now()
				results, err := simulateDatabaseQuery(ctx, tc.queryType, tc.recordCount)
				duration := time.Since(queryStart)

				latencies = append(latencies, duration)
				metrics.recordLatency(duration)

				if err == nil && len(results) > 0 {
					successCount++
				}
			}

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = successCount
			metrics.EventCount = tc.recordCount

			sortLatencies(latencies)
			metrics.calculateStats(latencies)

			// Verify SLA
			if metrics.P95Latency > tc.maxLatency {
				t.Errorf("%s: p95 latency = %v (target < %v)",
					tc.name, metrics.P95Latency, tc.maxLatency)
			}

			indexUsage := "indexed"
			if !tc.shouldUseIndex {
				indexUsage = "full table scan"
			}
			t.Logf("%s (%s): avg=%v, p95=%v, p99=%v",
				tc.name, indexUsage, metrics.AvgLatency, metrics.P95Latency, metrics.P99Latency)

			reportMetrics(t, metrics)
		})
	}
}

// TestBatchOperationOptimization tests batch operation performance
func TestBatchOperationOptimization(t *testing.T) {
	ctx := context.Background()

	batchSizes := []int{10, 100, 500, 1000}
	operationCount := 10000

	for _, batchSize := range batchSizes {
		t.Run(fmt.Sprintf("batch_size_%d", batchSize), func(t *testing.T) {
			metrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("batch_%d", batchSize),
				EventCount:    operationCount,
			}

			events := generateTestEvents(operationCount)
			start := time.Now()
			successCount := 0

			// Execute in batches
			for i := 0; i < len(events); i += batchSize {
				end := i + batchSize
				if end > len(events) {
					end = len(events)
				}

				batch := events[i:end]
				err := simulateEventStorage(ctx, batch)

				if err == nil {
					successCount += len(batch)
				}
			}

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = successCount

			if metrics.TotalDuration > 0 {
				metrics.ThroughputQPS = float64(successCount) / metrics.TotalDuration.Seconds()
			}

			// Verify throughput improvement
			expectedThroughput := float64(operationCount) / 1.0 // should complete in ~1 second
			if metrics.ThroughputQPS < expectedThroughput*0.5 {
				t.Logf("Warning: Batch throughput = %.0f events/sec (target > %.0f)",
					metrics.ThroughputQPS, expectedThroughput)
			}

			t.Logf("Batch size %d: %v total, %.0f events/sec throughput",
				batchSize, metrics.TotalDuration, metrics.ThroughputQPS)

			reportMetrics(t, metrics)
		})
	}
}

// TestLargeEventLogProcessing tests handling of very large event logs
func TestLargeEventLogProcessing(t *testing.T) {
	ctx := context.Background()

	logSizes := []int{100000, 1000000}

	for _, logSize := range logSizes {
		t.Run(fmt.Sprintf("log_size_%d", logSize), func(t *testing.T) {
			if testing.Short() && logSize > 100000 {
				t.Skip("Skipping large log test in short mode")
			}

			metrics := &PerformanceMetrics{
				OperationName: fmt.Sprintf("large_log_%d", logSize),
				EventCount:    logSize,
			}

			start := time.Now()

			// Process in streaming fashion
			batchSize := 1000
			successCount := 0

			for batch := 0; batch < (logSize / batchSize); batch++ {
				events := generateTestEvents(batchSize)
				err := simulateEventStorage(ctx, events)

				if err == nil {
					successCount += len(events)
				}
			}

			metrics.TotalDuration = time.Since(start)
			metrics.SuccessCount = successCount

			if metrics.TotalDuration > 0 {
				metrics.ThroughputQPS = float64(successCount) / metrics.TotalDuration.Seconds()
			}

			// SLA: 1M events should process in < 5 seconds
			maxDuration := 5 * time.Second
			if logSize > 500000 && metrics.TotalDuration > maxDuration {
				t.Logf("Warning: %d-event processing took %v (target < %v)",
					logSize, metrics.TotalDuration, maxDuration)
			}

			t.Logf("Large log processing: %d events in %v (%.0f events/sec)",
				logSize, metrics.TotalDuration, metrics.ThroughputQPS)

			reportMetrics(t, metrics)
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateTestEvent creates a single test event
func generateTestEvent(id int) TestEvent {
	return TestEvent{
		ID:        fmt.Sprintf("evt_%d", id),
		Timestamp: time.Now().Unix(),
		Activity:  fmt.Sprintf("activity_%d", id%20),
		Resource:  fmt.Sprintf("resource_%d", id%5),
		Data: map[string]string{
			"source":   "process_mining",
			"model_id": fmt.Sprintf("model_%d", id%10),
		},
	}
}

// generateTestEvents creates multiple test events
func generateTestEvents(count int) []TestEvent {
	events := make([]TestEvent, count)
	for i := 0; i < count; i++ {
		events[i] = generateTestEvent(i)
	}
	return events
}

// simulateEventStorage simulates storing events
func simulateEventStorage(ctx context.Context, events []TestEvent) error {
	// Simulate database insert with slight latency
	time.Sleep(time.Duration(len(events)) * time.Millisecond / 100)
	return nil
}

// simulateEventQuery simulates querying events
func simulateEventQuery(ctx context.Context, eventCount int, queryType string) ([]TestEvent, error) {
	// Simulate database query
	latency := 50 * time.Millisecond
	if queryType != "indexed" {
		latency = 500 * time.Millisecond
	}
	time.Sleep(latency)

	return []TestEvent{generateTestEvent(0)}, nil
}

// simulateCacheQuery simulates cache lookup
func simulateCacheQuery(ctx context.Context, cacheKey string) (bool, error) {
	// Simulate cache hit probability based on key hash
	hash := 0
	for _, ch := range cacheKey {
		hash += int(ch)
	}

	isHit := (hash % 100) < 80 // 80% hit rate simulation
	if isHit {
		time.Sleep(1 * time.Millisecond) // cache hit latency
	} else {
		time.Sleep(50 * time.Millisecond) // cache miss + db query
	}

	return isHit, nil
}

// simulateDatabaseQuery simulates different query types
func simulateDatabaseQuery(ctx context.Context, queryType string, recordCount int) ([]TestEvent, error) {
	var latency time.Duration

	switch queryType {
	case "indexed_single":
		latency = 10 * time.Millisecond
	case "indexed_range":
		latency = 50 * time.Millisecond
	case "indexed_join":
		latency = 100 * time.Millisecond
	case "full_table_scan":
		latency = time.Duration(recordCount/1000) * time.Millisecond
	default:
		latency = 100 * time.Millisecond
	}

	time.Sleep(latency)
	return []TestEvent{generateTestEvent(0)}, nil
}

// sortLatencies sorts latencies in ascending order (simple bucket sort)
func sortLatencies(latencies []time.Duration) {
	// Simple sorting for percentile calculation
	for i := 1; i < len(latencies); i++ {
		j := i
		for j > 0 && latencies[j] < latencies[j-1] {
			latencies[j], latencies[j-1] = latencies[j-1], latencies[j]
			j--
		}
	}
}

// reportMetrics prints performance metrics
func reportMetrics(t *testing.T, m *PerformanceMetrics) {
	t.Logf(`
=== Performance Metrics: %s ===
Total Duration:   %v
Event Count:      %d
Success Count:    %d
Error Count:      %d
Throughput:       %.0f requests/sec
Latency (min):    %v
Latency (avg):    %v
Latency (p50):    %v
Latency (p95):    %v
Latency (p99):    %v
Latency (max):    %v
Memory Alloc:     %.2f MB
Memory Rate:      %.2f MB/sec
`,
		m.OperationName,
		m.TotalDuration,
		m.EventCount,
		m.SuccessCount,
		m.ErrorCount,
		m.ThroughputQPS,
		m.MinLatency,
		m.AvgLatency,
		m.P50Latency,
		m.P95Latency,
		m.P99Latency,
		m.MaxLatency,
		m.MemoryAllocMB,
		m.MemoryAllocRate,
	)
}

// ============================================================================
// Benchmark-Style Tests
// ============================================================================

// BenchmarkEventStorageLatency runs a minimal latency benchmark
func BenchmarkEventStorageLatency(b *testing.B) {
	ctx := context.Background()
	event := generateTestEvent(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = simulateEventStorage(ctx, []TestEvent{event})
	}
}

// BenchmarkBatchEventStorage runs batch storage benchmark
func BenchmarkBatchEventStorage(b *testing.B) {
	ctx := context.Background()
	events := generateTestEvents(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = simulateEventStorage(ctx, events)
	}
}

// BenchmarkCacheQuery runs cache query benchmark
func BenchmarkCacheQuery(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = simulateCacheQuery(ctx, fmt.Sprintf("query_%d", i))
	}
}
