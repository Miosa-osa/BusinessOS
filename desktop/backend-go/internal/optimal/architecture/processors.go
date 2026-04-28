package architecture

import (
	"context"
	"errors"
	"fmt"
)

// Processor stubs ported from lib/optimal_engine/architecture/processors/*.ex.
//
// The Elixir source's processors are themselves thin wrappers — the real
// model calls live in OptimalEngine.LLM, OptimalEngine.Vision, etc. Here we
// register each processor as a no-op that returns EmitNoop by default.
// Callers (or tenant code) can SwapProcessor() with a concrete impl that
// calls Anthropic / Ollama / OpenAI / a vision model / a TS feature lib.
//
// Why no-ops: BusinessOS already has its own LLM + provider plumbing in
// internal/handlers/optimal.go and internal/integrations/providers/. The
// architecture layer is the schema seam, not a duplicate model client.

func init() {
	for _, p := range []Processor{
		&stubProcessor{id: "text_embedder", modality: ModalityText, emits: []EmitKind{EmitEmbedding}},
		&stubProcessor{id: "code_embedder", modality: ModalityCode, emits: []EmitKind{EmitEmbedding}},
		&stubProcessor{id: "image_embedder", modality: ModalityImage, emits: []EmitKind{EmitEmbedding, EmitCaption, EmitOCR}},
		&stubProcessor{id: "audio_embedder", modality: ModalityAudio, emits: []EmitKind{EmitEmbedding, EmitTranscription}},
		&stubProcessor{id: "video_embedder", modality: ModalityVideo, emits: []EmitKind{EmitEmbedding, EmitCaption}},
		&stubProcessor{id: "ts_feature_extractor", modality: ModalityTimeSeries, emits: []EmitKind{EmitFeatures, EmitAnomaly}},
	} {
		RegisterProcessor(p)
	}
}

// stubProcessor returns EmitNoop with a metadata flag so callers can detect
// "processor stub still in place" in dev. SwapProcessor replaces it with
// a real impl when the concrete layer is wired.
type stubProcessor struct {
	id       string
	modality Modality
	emits    []EmitKind
}

func (s *stubProcessor) ID() string                                              { return s.id }
func (s *stubProcessor) Modality() Modality                                      { return s.modality }
func (s *stubProcessor) Emits() []EmitKind                                       { return s.emits }
func (s *stubProcessor) Init(_ context.Context, _ map[string]any) (State, error) { return nil, nil }

func (s *stubProcessor) Process(_ context.Context, field *Field, _ any, _ State) (*Output, error) {
	if field == nil {
		return nil, errors.New("architecture: processor received nil field")
	}
	return &Output{
		Kind:  EmitNoop,
		Value: nil,
		Metadata: map[string]any{
			"stub":      true,
			"processor": s.id,
			"field":     field.Name,
			"note":      fmt.Sprintf("swap %s with a real Processor; see architecture/processors.go", s.id),
		},
	}, nil
}

// SwapProcessor replaces a registered processor by id. Returns
// ErrProcessorNotFound when the id slot doesn't exist (callers should
// ensure init() ran first), or an id-mismatch error when the new processor
// reports a different ID than the slot.
func SwapProcessor(id string, p Processor) error {
	if p == nil {
		return errors.New("architecture: cannot swap to nil processor")
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := procByID[id]; !ok {
		return ErrProcessorNotFound
	}
	if p.ID() != id {
		return fmt.Errorf("architecture: id mismatch — registry %q vs processor.ID %q", id, p.ID())
	}
	procByID[id] = p
	procByModality[p.Modality()] = p
	return nil
}
