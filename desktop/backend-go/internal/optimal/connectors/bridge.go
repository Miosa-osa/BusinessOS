package connectors

import (
	"context"
	"fmt"
	"strings"
)

// Bridge is the seam between the connector layer and the ingest pipeline.
// The caller supplies an IngestFunc; the bridge formats each Signal into the
// plain-text body the optimal.Ingest function already accepts, preserving
// title + path metadata so classify/route/topology all keep working.
//
// This deliberately avoids importing the optimal package directly to avoid
// a circular dep (connectors ← adapters, optimal → connectors at call-site).
type IngestFunc func(ctx context.Context, body, genre string) error

// IngestBatch walks a set of Signals, formats each into a frontmatter-prefixed
// body, and feeds it to the pipeline. It returns the count of successful
// ingestions and the first error encountered, if any.
func IngestBatch(ctx context.Context, signals []Signal, ingest IngestFunc) (int, error) {
	ok := 0
	for _, sig := range signals {
		body := formatSignalBody(sig)
		if err := ingest(ctx, body, sig.Genre); err != nil {
			return ok, fmt.Errorf("connectors/bridge: ingest %s: %w", sig.ID, err)
		}
		ok++
	}
	return ok, nil
}

// MakeOnBatch is a convenience that wraps IngestBatch as a RunOptions.OnBatch
// callback. Typical usage:
//
//	opts := connectors.RunOptions{OnBatch: connectors.MakeOnBatch(myIngest)}
func MakeOnBatch(ingest IngestFunc) func(ctx context.Context, signals []Signal) error {
	return func(ctx context.Context, signals []Signal) error {
		_, err := IngestBatch(ctx, signals, ingest)
		return err
	}
}

// formatSignalBody renders a Signal as a YAML-frontmatter + markdown document,
// the same shape the ingest pipeline parses via frontmatter.go. Keeping the
// formatter in this package means adapters don't need to know pipeline
// internals.
func formatSignalBody(sig Signal) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", sig.ID)
	if sig.Title != "" {
		fmt.Fprintf(&b, "title: %q\n", sig.Title)
	}
	if sig.Path != "" {
		fmt.Fprintf(&b, "path: %s\n", sig.Path)
	}
	if sig.Genre != "" {
		fmt.Fprintf(&b, "genre: %s\n", sig.Genre)
	}
	if !sig.ModifiedAt.IsZero() {
		fmt.Fprintf(&b, "modified_at: %s\n", sig.ModifiedAt.UTC().Format("2006-01-02T15:04:05Z"))
	}
	if len(sig.Entities) > 0 {
		fmt.Fprintf(&b, "entities: [%s]\n", strings.Join(quoteAll(sig.Entities), ", "))
	}
	b.WriteString("---\n\n")
	if sig.Title != "" {
		fmt.Fprintf(&b, "# %s\n\n", sig.Title)
	}
	b.WriteString(sig.Content)
	return b.String()
}

func quoteAll(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = fmt.Sprintf("%q", x)
	}
	return out
}
