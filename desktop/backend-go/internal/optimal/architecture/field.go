// Package architecture ports OptimalEngine.Architecture (Elixir) to Go.
//
// A DataArchitecture describes the shape of one class of data point — a
// clinical visit, a sales call, a code commit — as a composition of typed
// Fields plus the processor bindings that turn each field into something
// retrievable. The engine is model-agnostic: an architecture doesn't say
// "use GPT-4", it says "apply a text-embedding processor to the `note`
// field". Swap the processor; the data-point shape is unchanged.
package architecture

// Modality is the kind of signal a field carries. Every modality maps to
// a default embedding family, decomposition strategy, and classifier preset.
type Modality string

const (
	ModalityText       Modality = "text"
	ModalityCode       Modality = "code"
	ModalityImage      Modality = "image"
	ModalityAudio      Modality = "audio"
	ModalityVideo      Modality = "video"
	ModalityTimeSeries Modality = "time_series"
	ModalityTable      Modality = "table"
	ModalityStructured Modality = "structured"
	ModalityGraph      Modality = "graph"
	ModalityTensor     Modality = "tensor"
	ModalityGeo        Modality = "geo"
	ModalityBinary     Modality = "binary"
)

// Modalities returns every recognized modality, sorted for stable output.
func Modalities() []Modality {
	return []Modality{
		ModalityText, ModalityCode, ModalityImage, ModalityAudio,
		ModalityVideo, ModalityTimeSeries, ModalityTable, ModalityStructured,
		ModalityGraph, ModalityTensor, ModalityGeo, ModalityBinary,
	}
}

// Dim is one dimension entry in a field's shape. -1 (or DimAny) means "any".
// Concrete sizes (768, 224, …) are the typical case.
type Dim int

// DimAny is the wildcard dimension, matching Elixir's `:any`.
const DimAny Dim = -1

// Field is one component of an Architecture.
//
// Shape (`Dims`) lets the engine reject mis-shaped inputs before they hit a
// model that can't handle them. A 768-d embedding → [768]. A 224×224 RGB
// image → [3, 224, 224]. A stream whose length varies → [DimAny, 768].
//
// `Processor` is the id of the preferred processor in the registry; a field
// without a hint falls back to the registry default for that modality.
type Field struct {
	Name        string
	Modality    Modality
	Dims        []Dim
	Required    bool
	Processor   string
	Description string
}

// IsCompatibleShape reports whether a concrete dim vector matches the field's
// declared shape. Concrete dims must equal the declared dim (when not DimAny).
// Length mismatch is always incompatible.
func (f *Field) IsCompatibleShape(actual []int) bool {
	if len(actual) != len(f.Dims) {
		return false
	}
	for i, declared := range f.Dims {
		if declared == DimAny {
			continue
		}
		if int(declared) != actual[i] {
			return false
		}
	}
	return true
}
