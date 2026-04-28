package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/rhl/businessos-backend/internal/optimal"
)

// Signal is the bridge to the signal-classification + failure-mode +
// CloudEvents tooling. The underlying classifier and failure-mode detector
// already exist in internal/optimal — this file is a thin facade so callers
// don't have to know which file holds which function.
type Signal struct{}

// NewSignal returns the bridge — stateless; package-level functions would
// work too but we keep the type so callers can mock during testing.
func NewSignal() *Signal { return &Signal{} }

// Classify wraps optimal.Classify with the same return shape so callers
// migrating from Elixir's Bridge.Signal.classify/1 get a 1:1 swap.
func (s *Signal) Classify(text string) optimal.SignalClassification {
	return optimal.Classify(text)
}

// Audit returns the 11 Signal-Theory failure modes detected against
// content. Empty list = clean signal. Mirrors Bridge.Signal.audit/1.
func (s *Signal) Audit(text string) []optimal.FailureMode {
	cl := optimal.Classify(text)
	return optimal.DetectFailureModes(cl, text, nil)
}

// MeasureSN returns the S/N ratio computed by the classifier. Always
// in [0, 1].
func (s *Signal) MeasureSN(text string) float64 {
	return optimal.Classify(text).SNRatio
}

// CloudEvent wraps a Signal as a CloudEvents v1.0.2 envelope map. The shape
// is the JSON-mode binding; callers serialize with encoding/json.
//
// Required fields (per CloudEvents spec): id, source, specversion, type.
// Optional: time, datacontenttype, subject, data.
func (s *Signal) CloudEvent(text, source string) map[string]any {
	cl := optimal.Classify(text)
	id := newEventID()
	if source == "" {
		source = "optimal-engine"
	}
	eventType := "optimal." + valueOr(cl.Genre, "note") + "." + valueOr(cl.Type, "inform")
	return map[string]any{
		"specversion":     "1.0.2",
		"id":              id,
		"source":          source,
		"type":            eventType,
		"time":            time.Now().UTC().Format(time.RFC3339),
		"datacontenttype": "application/json",
		"data": map[string]any{
			"text":      text,
			"mode":      cl.Mode,
			"genre":     cl.Genre,
			"type":      cl.Type, // signal-theory T dimension
			"format":    cl.Format,
			"structure": cl.Structure,
			"sn_ratio":  cl.SNRatio,
		},
	}
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func newEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Time-based fallback so callers always get a non-empty id.
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b)
}
