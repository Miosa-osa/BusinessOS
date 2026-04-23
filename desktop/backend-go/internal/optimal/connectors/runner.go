package connectors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// RunOptions tune a single Runner.Run cycle.
type RunOptions struct {
	// MaxBatches bounds how many Sync cycles Run will execute before
	// returning. 0 means "run until the adapter returns an empty cursor".
	MaxBatches int
	// InitialCursor overrides whatever persisted cursor the caller would
	// otherwise use. Empty string starts from the beginning.
	InitialCursor Cursor
	// OnBatch is invoked with every successful batch of signals. The
	// callback is where the caller hands signals to the ingest pipeline.
	// If it returns an error the run stops and surfaces that error.
	OnBatch func(ctx context.Context, signals []Signal) error
	// MaxRetries bounds transient retries (default 5).
	MaxRetries int
}

// RunResult summarizes a completed Run.
type RunResult struct {
	Kind        string
	Batches     int
	SignalCount int
	FinalCursor Cursor
}

// Run executes one sync cycle for the given connector kind. It resolves the
// adapter from the registry, initializes it with cfg, then loops Sync + OnBatch
// honoring the Behaviour error-category contract:
//
//	ErrRateLimited  → sleep RetryAfter then retry with same cursor
//	ErrAuthExpired  → return so the caller can refresh; retry-once is TBD
//	ErrTransient    → exponential backoff up to MaxRetries
//	ErrFatal        → return the error immediately
//
// This is intentionally conservative — refresh-and-retry on auth expiry
// will land in Phase 3 once the credential rotator GenServer is ported.
func Run(ctx context.Context, kind string, cfg Config, opts RunOptions) (*RunResult, error) {
	adapter, err := Lookup(kind)
	if err != nil {
		return nil, err
	}

	state, err := adapter.Init(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connectors/runner: init %s: %w", kind, err)
	}

	if opts.MaxRetries == 0 {
		opts.MaxRetries = 5
	}

	result := &RunResult{Kind: kind, FinalCursor: opts.InitialCursor}
	cursor := opts.InitialCursor
	attempt := 0

	for {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if opts.MaxBatches > 0 && result.Batches >= opts.MaxBatches {
			return result, nil
		}

		batch, err := adapter.Sync(ctx, state, cursor)
		if err != nil {
			var syncErr *SyncError
			if !errors.As(err, &syncErr) {
				return result, err
			}
			switch syncErr.Category {
			case ErrRateLimited:
				wait := syncErr.RetryAfter
				if wait == 0 {
					wait = 2 * time.Second
				}
				slog.Warn("connectors/runner: rate limited", "kind", kind, "wait", wait)
				if err := sleep(ctx, wait); err != nil {
					return result, err
				}
				continue
			case ErrTransient:
				attempt++
				if attempt > opts.MaxRetries {
					return result, fmt.Errorf("connectors/runner: transient retries exhausted: %w", err)
				}
				backoff := time.Duration(1<<attempt) * time.Second
				slog.Warn("connectors/runner: transient", "kind", kind, "attempt", attempt, "backoff", backoff)
				if err := sleep(ctx, backoff); err != nil {
					return result, err
				}
				continue
			case ErrAuthExpired, ErrFatal, ErrNotImplemented:
				return result, err
			default:
				return result, err
			}
		}
		attempt = 0

		if opts.OnBatch != nil && len(batch.Signals) > 0 {
			if err := opts.OnBatch(ctx, batch.Signals); err != nil {
				return result, fmt.Errorf("connectors/runner: onbatch: %w", err)
			}
		}
		result.Batches++
		result.SignalCount += len(batch.Signals)
		result.FinalCursor = batch.Next

		// Empty-cursor is how an adapter signals "caught up". Empty signals
		// with a non-nil cursor means "nothing new in this window, but more
		// windows remain" — keep looping.
		if batch.Next == "" && len(batch.Signals) == 0 {
			return result, nil
		}
		if batch.Next == cursor {
			// Defensive: a misbehaving adapter that returns the same cursor
			// without progress would otherwise loop forever.
			return result, nil
		}
		cursor = batch.Next
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
