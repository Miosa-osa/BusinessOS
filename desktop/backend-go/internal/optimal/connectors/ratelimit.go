package connectors

import (
	"context"
	"sync"
	"time"
)

// Token-bucket rate limiter, one bucket per adapter/endpoint key. Ported from
// Connectors.RateLimit (Elixir). The bucket refills at Rate tokens per second
// up to Burst capacity.
type bucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64
	burst  int
}

var (
	bucketsMu sync.Mutex
	buckets   = map[string]*bucket{}
)

func getBucket(spec *BucketSpec) *bucket {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()
	if b, ok := buckets[spec.Key]; ok {
		return b
	}
	b := &bucket{tokens: float64(spec.Burst), last: time.Now(), rate: spec.Rate, burst: spec.Burst}
	buckets[spec.Key] = b
	return b
}

// waitForBucket blocks until one token is available or ctx expires.
func waitForBucket(ctx context.Context, spec *BucketSpec) error {
	b := getBucket(spec)
	for {
		b.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(b.last).Seconds()
		b.tokens = min(float64(b.burst), b.tokens+elapsed*b.rate)
		b.last = now
		if b.tokens >= 1 {
			b.tokens--
			b.mu.Unlock()
			return nil
		}
		// Sleep long enough for exactly one token to accrue.
		deficit := 1 - b.tokens
		wait := time.Duration(deficit/b.rate*float64(time.Second)) + time.Millisecond
		b.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
