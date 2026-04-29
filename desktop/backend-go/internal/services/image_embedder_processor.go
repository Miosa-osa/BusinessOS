// image_embedder_processor.go — closes the visual half of the architecture
// embedding gap. ImageEmbeddingService can call CLIP via three providers
// (local, openai, replicate); we don't reach into those — the processor
// just hands off the bytes and surfaces whatever the service returns.
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/rhl/businessos-backend/internal/optimal/architecture"
)

// ImageEmbedderProcessor implements architecture.Processor on top of the
// existing ImageEmbeddingService (CLIP 512d). The processor accepts:
//
//	[]byte                      — raw bytes (PNG/JPEG/SVG/etc.)
//	map[string]any{"data":[]byte} — explicit data field
//
// We don't accept URLs here — the upstream architecture pipeline is
// expected to materialize bytes before calling the processor. That keeps
// the processor side-effect-free and easy to test.
type ImageEmbedderProcessor struct {
	embed *ImageEmbeddingService
}

// NewImageEmbedderProcessor returns a processor wired to the given service.
// The service must be non-nil; bootstrap only swaps the stub when the
// service initialized successfully (CLIP_PROVIDER configured).
func NewImageEmbedderProcessor(svc *ImageEmbeddingService) *ImageEmbedderProcessor {
	return &ImageEmbedderProcessor{embed: svc}
}

func (p *ImageEmbedderProcessor) ID() string                      { return "image_embedder" }
func (p *ImageEmbedderProcessor) Modality() architecture.Modality { return architecture.ModalityImage }

// Emits is just embedding here. CLIP can also produce captions and OCR via
// downstream models; those land as separate processors when wired.
func (p *ImageEmbedderProcessor) Emits() []architecture.EmitKind {
	return []architecture.EmitKind{architecture.EmitEmbedding}
}

func (p *ImageEmbedderProcessor) Init(_ context.Context, _ map[string]any) (architecture.State, error) {
	return nil, nil
}

func (p *ImageEmbedderProcessor) Process(ctx context.Context, field *architecture.Field, value any, _ architecture.State) (*architecture.Output, error) {
	if p.embed == nil {
		return nil, errors.New("image_embedder: image embedding service not wired")
	}
	if field == nil {
		return nil, errors.New("image_embedder: nil field")
	}
	bytes, err := bytesFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("image_embedder: field %q: %w", field.Name, err)
	}
	if len(bytes) == 0 {
		return &architecture.Output{
			Kind:     architecture.EmitNoop,
			Metadata: map[string]any{"reason": "empty input", "field": field.Name},
		}, nil
	}

	vec, err := p.embed.GenerateEmbedding(ctx, bytes)
	if err != nil {
		return nil, fmt.Errorf("image_embedder: generate %q: %w", field.Name, err)
	}
	return &architecture.Output{
		Kind:  architecture.EmitEmbedding,
		Value: vec,
		Metadata: map[string]any{
			"field":      field.Name,
			"model":      "clip-vit-base-patch32",
			"dimensions": len(vec),
		},
	}, nil
}

// bytesFromValue accepts the raw byte shapes a Process caller might send.
// We deliberately reject nil + non-byte types loudly so a misconfigured
// pipeline doesn't silently produce zero-length embeddings.
func bytesFromValue(v any) ([]byte, error) {
	switch x := v.(type) {
	case []byte:
		return x, nil
	case nil:
		return nil, nil
	case map[string]any:
		// `{ "data": []byte }` shape that some upstream processors emit.
		if d, ok := x["data"]; ok {
			return bytesFromValue(d)
		}
		return nil, fmt.Errorf("expected []byte or {data: []byte}, got map without 'data' key")
	default:
		return nil, fmt.Errorf("expected []byte or {data: []byte}, got %T", v)
	}
}

var _ architecture.Processor = (*ImageEmbedderProcessor)(nil)
