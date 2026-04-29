// Package services — text_embedder_processor.go
//
// Closes the architecture-layer embedding gap: the Wave-8 architecture
// package shipped six processor stubs (text/code/image/audio/video/ts) that
// all return EmitNoop, by design — the engine doesn't carry its own model
// client; tenant code wires real implementations via SwapProcessor.
//
// This file provides the BusinessOS-native text embedder. It implements
// architecture.Processor on top of the existing EmbeddingService (Ollama
// nomic-embed-text 768d). One-line bootstrap:
//
//	architecture.SwapProcessor("text_embedder",
//	    services.NewTextEmbedderProcessor(embeddingService))
//
// After that any architecture-driven ingest path (text_signal, code_commit's
// `message` field, audio_transcript's `transcript` field, etc.) produces
// real vectors instead of EmitNoop.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rhl/businessos-backend/internal/optimal/architecture"
)

// TextEmbedderProcessor wraps EmbeddingService.GenerateEmbedding so it
// satisfies architecture.Processor. The wrapper is deliberately tiny —
// the heavy lifting (caching, retries, model selection) already lives in
// EmbeddingService and we don't re-implement any of it here.
type TextEmbedderProcessor struct {
	embed *EmbeddingService
}

// NewTextEmbedderProcessor returns a processor wired to the given service.
// The service must be non-nil and pass HealthCheck for embeddings to
// actually produce vectors; if it doesn't, Process returns an error and
// the calling pipeline records it without crashing.
func NewTextEmbedderProcessor(svc *EmbeddingService) *TextEmbedderProcessor {
	return &TextEmbedderProcessor{embed: svc}
}

// ID matches the stub it replaces — SwapProcessor enforces equality.
func (p *TextEmbedderProcessor) ID() string { return "text_embedder" }

// Modality is text. Code, audio_transcript, multimodal_media, etc. all
// route their textual fields through the same embedder.
func (p *TextEmbedderProcessor) Modality() architecture.Modality {
	return architecture.ModalityText
}

// Emits is just the embedding kind — keep the contract narrow so callers
// can reason about routing. Summary/entities/etc. stay separate.
func (p *TextEmbedderProcessor) Emits() []architecture.EmitKind {
	return []architecture.EmitKind{architecture.EmitEmbedding}
}

// Init is a no-op. The EmbeddingService is constructed once at app start
// and shared across every Process call; we don't need per-field state.
func (p *TextEmbedderProcessor) Init(_ context.Context, _ map[string]any) (architecture.State, error) {
	return nil, nil
}

// Process generates a 768-dimensional embedding for the field's text value.
// Empty / whitespace-only inputs short-circuit to EmitNoop so the upstream
// engine doesn't waste an API round-trip on nothing.
//
// Errors are surfaced verbatim — the pipeline's per-field error handler
// decides whether to retry, skip, or escalate. We don't swallow.
func (p *TextEmbedderProcessor) Process(ctx context.Context, field *architecture.Field, value any, _ architecture.State) (*architecture.Output, error) {
	if p.embed == nil {
		return nil, errors.New("text_embedder: embedding service not wired")
	}
	if field == nil {
		return nil, errors.New("text_embedder: nil field")
	}
	text, err := stringFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("text_embedder: field %q: %w", field.Name, err)
	}
	if strings.TrimSpace(text) == "" {
		return &architecture.Output{
			Kind:     architecture.EmitNoop,
			Metadata: map[string]any{"reason": "empty input", "field": field.Name},
		}, nil
	}

	vec, err := p.embed.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("text_embedder: generate %q: %w", field.Name, err)
	}
	return &architecture.Output{
		Kind:  architecture.EmitEmbedding,
		Value: vec,
		Metadata: map[string]any{
			"field":      field.Name,
			"model":      "nomic-embed-text",
			"dimensions": len(vec),
		},
	}, nil
}

// stringFromValue normalizes the few shapes a Process caller might pass:
// the engine usually hands us a string, but architecture-driven ingest
// sometimes wraps content in a single-element []byte or fmt.Stringer.
// Anything else is rejected loudly so we don't silently embed `<nil>`.
func stringFromValue(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case fmt.Stringer:
		return x.String(), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("expected string/[]byte/Stringer, got %T", v)
	}
}

// Compile-time assertion that we satisfy the Processor contract. If
// architecture.Processor's signature ever drifts, this line breaks the
// build before we hit production with a runtime mismatch.
var _ architecture.Processor = (*TextEmbedderProcessor)(nil)

// CodeEmbedderProcessor reuses EmbeddingService — nomic-embed-text handles
// code reasonably well at our retrieval scale. When a code-specific model
// (e.g. CodeBERT, jina-embeddings-v2-base-code) lands, swap this struct's
// underlying call without changing the architecture/registry plumbing.
type CodeEmbedderProcessor struct {
	embed *EmbeddingService
}

// NewCodeEmbedderProcessor returns a processor wired to the given service.
func NewCodeEmbedderProcessor(svc *EmbeddingService) *CodeEmbedderProcessor {
	return &CodeEmbedderProcessor{embed: svc}
}

func (p *CodeEmbedderProcessor) ID() string                      { return "code_embedder" }
func (p *CodeEmbedderProcessor) Modality() architecture.Modality { return architecture.ModalityCode }
func (p *CodeEmbedderProcessor) Emits() []architecture.EmitKind {
	return []architecture.EmitKind{architecture.EmitEmbedding}
}
func (p *CodeEmbedderProcessor) Init(_ context.Context, _ map[string]any) (architecture.State, error) {
	return nil, nil
}

// Process generates an embedding for the code field. We deliberately don't
// strip syntax / normalize — nomic-embed-text learns enough surface
// features that comments, identifiers, and structure all contribute.
func (p *CodeEmbedderProcessor) Process(ctx context.Context, field *architecture.Field, value any, _ architecture.State) (*architecture.Output, error) {
	if p.embed == nil {
		return nil, errors.New("code_embedder: embedding service not wired")
	}
	if field == nil {
		return nil, errors.New("code_embedder: nil field")
	}
	src, err := stringFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("code_embedder: field %q: %w", field.Name, err)
	}
	if strings.TrimSpace(src) == "" {
		return &architecture.Output{
			Kind:     architecture.EmitNoop,
			Metadata: map[string]any{"reason": "empty input", "field": field.Name},
		}, nil
	}
	vec, err := p.embed.GenerateEmbedding(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("code_embedder: generate %q: %w", field.Name, err)
	}
	return &architecture.Output{
		Kind:  architecture.EmitEmbedding,
		Value: vec,
		Metadata: map[string]any{
			"field":      field.Name,
			"model":      "nomic-embed-text",
			"dimensions": len(vec),
			"input_kind": "code",
		},
	}, nil
}

var _ architecture.Processor = (*CodeEmbedderProcessor)(nil)
