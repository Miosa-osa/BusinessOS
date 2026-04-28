package architecture

import "context"

// EmitKind names what a processor emits — the engine routes the output to
// the matching downstream table (embeddings, entities, classifications, …).
type EmitKind string

const (
	EmitEmbedding      EmitKind = "embedding"
	EmitClassification EmitKind = "classification"
	EmitEntities       EmitKind = "entities"
	EmitSummary        EmitKind = "summary"
	EmitTranscription  EmitKind = "transcription"
	EmitOCR            EmitKind = "ocr"
	EmitTags           EmitKind = "tags"
	EmitFeatures       EmitKind = "features"
	EmitScore          EmitKind = "score"
	EmitAnomaly        EmitKind = "anomaly"
	EmitCaption        EmitKind = "caption"
	EmitSegmentation   EmitKind = "segmentation"
	EmitNoop           EmitKind = "noop"
)

// Output is the canonical processor return value.
type Output struct {
	Kind     EmitKind
	Value    any
	Metadata map[string]any
}

// Processor is the contract every concrete processor implements. The
// engine doesn't care which category a processor belongs to — it cares
// that the processor can be initialized, can process a field, and can
// declare what it emits.
type Processor interface {
	// ID is the stable atom the architecture refers to ("text_embedder").
	ID() string
	// Modality reports what kind of field this processor consumes.
	// Returning ModalityText means "texty fields only"; returning
	// "" means "any modality" (matches Elixir's `:any`).
	Modality() Modality
	// Emits lists every EmitKind this processor can produce. A single
	// processor may emit multiple kinds (e.g. an LLM that returns both
	// embedding + summary).
	Emits() []EmitKind
	// Init hydrates runtime state from config. Optional — the registry
	// passes nil-config to processors that don't override.
	Init(ctx context.Context, cfg map[string]any) (State, error)
	// Process applies the processor to a field value, producing one
	// Output per call. Returning EmitNoop is fine — the engine simply
	// skips persisting the row.
	Process(ctx context.Context, field *Field, value any, state State) (*Output, error)
}

// State is the processor-private runtime handle (a model client, a tokenizer,
// or nothing). Threaded through Process by the caller.
type State any

// AcceptsModality reports whether p can consume fields of the given modality.
// "" / "any" matches anything.
func AcceptsModality(p Processor, m Modality) bool {
	pm := p.Modality()
	return pm == "" || pm == "any" || pm == m
}
