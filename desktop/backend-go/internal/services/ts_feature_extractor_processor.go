// ts_feature_extractor_processor.go — pure-Go time-series feature
// extraction. No external service: the math is light enough to run inline,
// and a "processor" that needs no infra is the right shape for Phase 1
// retrieval over telemetry windows.
//
// We emit EmitFeatures with a fixed keyset so downstream indexing has a
// stable schema:
//
//	{ count, mean, stdev, min, max, slope, last_z }
//
// `last_z` is the z-score of the last sample relative to the rest of the
// window — a quick anomaly indicator. A separate processor that emits
// EmitAnomaly with a calibrated threshold can layer on top later.
package services

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rhl/businessos-backend/internal/optimal/architecture"
)

// TSFeatureExtractorProcessor implements architecture.Processor for fixed-
// duration numeric windows. Input is []float64 (or anything we can coerce);
// output is a map[string]float64 with the canonical features above.
type TSFeatureExtractorProcessor struct{}

// NewTSFeatureExtractorProcessor returns a stateless processor.
func NewTSFeatureExtractorProcessor() *TSFeatureExtractorProcessor {
	return &TSFeatureExtractorProcessor{}
}

func (p *TSFeatureExtractorProcessor) ID() string { return "ts_feature_extractor" }
func (p *TSFeatureExtractorProcessor) Modality() architecture.Modality {
	return architecture.ModalityTimeSeries
}

func (p *TSFeatureExtractorProcessor) Emits() []architecture.EmitKind {
	return []architecture.EmitKind{architecture.EmitFeatures}
}

func (p *TSFeatureExtractorProcessor) Init(_ context.Context, _ map[string]any) (architecture.State, error) {
	return nil, nil
}

// Process computes features for the window. NaN samples are dropped;
// inf samples are dropped. This is intentional — telemetry pipelines
// occasionally produce dirty values and we'd rather under-report than
// poison every downstream feature with a single garbage point.
func (p *TSFeatureExtractorProcessor) Process(_ context.Context, field *architecture.Field, value any, _ architecture.State) (*architecture.Output, error) {
	if field == nil {
		return nil, errors.New("ts_feature_extractor: nil field")
	}
	samples, err := samplesFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("ts_feature_extractor: field %q: %w", field.Name, err)
	}
	if len(samples) == 0 {
		return &architecture.Output{
			Kind:     architecture.EmitNoop,
			Metadata: map[string]any{"reason": "empty window", "field": field.Name},
		}, nil
	}

	features := computeFeatures(samples)
	return &architecture.Output{
		Kind:  architecture.EmitFeatures,
		Value: features,
		Metadata: map[string]any{
			"field":       field.Name,
			"window_size": len(samples),
			"keyset":      []string{"count", "mean", "stdev", "min", "max", "slope", "last_z"},
		},
	}, nil
}

// samplesFromValue accepts the few numeric-window shapes upstream callers
// might emit. Strings / maps are rejected loudly — a TS window must be
// numeric to be meaningful.
func samplesFromValue(v any) ([]float64, error) {
	switch x := v.(type) {
	case []float64:
		return filterFinite(x), nil
	case []float32:
		out := make([]float64, len(x))
		for i, n := range x {
			out[i] = float64(n)
		}
		return filterFinite(out), nil
	case []int:
		out := make([]float64, len(x))
		for i, n := range x {
			out[i] = float64(n)
		}
		return out, nil
	case []any:
		out := make([]float64, 0, len(x))
		for i, n := range x {
			f, err := coerceNumber(n)
			if err != nil {
				return nil, fmt.Errorf("sample[%d]: %w", i, err)
			}
			if math.IsNaN(f) || math.IsInf(f, 0) {
				continue
			}
			out = append(out, f)
		}
		return out, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected []float64/[]float32/[]int/[]any, got %T", v)
	}
}

func filterFinite(in []float64) []float64 {
	out := in[:0]
	for _, x := range in {
		if !math.IsNaN(x) && !math.IsInf(x, 0) {
			out = append(out, x)
		}
	}
	return out
}

func coerceNumber(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	default:
		return 0, fmt.Errorf("non-numeric sample: %T", v)
	}
}

// computeFeatures runs the canonical feature set in a single pass where
// possible. Slope is via least-squares against integer indices [0..N-1]
// — meaningful when samples are evenly spaced; for irregular series the
// caller should provide explicit timestamps (future enhancement).
func computeFeatures(s []float64) map[string]float64 {
	n := float64(len(s))
	if n == 0 {
		return nil
	}
	var sum, mn, mx float64
	mn = math.Inf(1)
	mx = math.Inf(-1)
	for _, v := range s {
		sum += v
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	mean := sum / n

	// Stdev (population) — second pass to keep arithmetic stable; for
	// retrieval features the cost is negligible at our window sizes.
	var sqDiff float64
	for _, v := range s {
		d := v - mean
		sqDiff += d * d
	}
	stdev := math.Sqrt(sqDiff / n)

	// Slope via least-squares on (i, v_i).
	var sumX, sumY, sumXY, sumXX float64
	for i, v := range s {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumXX += x * x
	}
	denom := n*sumXX - sumX*sumX
	slope := 0.0
	if denom != 0 {
		slope = (n*sumXY - sumX*sumY) / denom
	}

	// Z-score of the last sample.
	lastZ := 0.0
	if stdev > 0 {
		lastZ = (s[len(s)-1] - mean) / stdev
	}

	return map[string]float64{
		"count":  n,
		"mean":   mean,
		"stdev":  stdev,
		"min":    mn,
		"max":    mx,
		"slope":  slope,
		"last_z": lastZ,
	}
}

var _ architecture.Processor = (*TSFeatureExtractorProcessor)(nil)
