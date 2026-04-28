package architecture

import (
	"context"
	"errors"
	"testing"
)

func TestBuiltinArchitecturesRegistered(t *testing.T) {
	want := []string{
		"text_signal", "code_commit", "audio_transcript", "image_asset",
		"multimodal_media", "structured_record", "time_series_window",
	}
	for _, name := range want {
		if _, err := LookupArchitecture(name); err != nil {
			t.Errorf("architecture %s missing: %v", name, err)
		}
	}
}

func TestArchitectureLookupByIDAndName(t *testing.T) {
	byName, _ := LookupArchitecture("text_signal")
	byID, err := LookupArchitecture(byName.ID)
	if err != nil {
		t.Fatal(err)
	}
	if byID != byName {
		t.Fatal("ID and name lookup must return the same pointer")
	}
}

func TestArchitectureLookupNotFound(t *testing.T) {
	if _, err := LookupArchitecture("nope"); !errors.Is(err, ErrArchNotFound) {
		t.Fatalf("expected ErrArchNotFound, got %v", err)
	}
}

func TestRequiredFields(t *testing.T) {
	a, _ := LookupArchitecture("text_signal")
	got := a.RequiredFields()
	want := map[string]bool{"title": true, "body": true}
	for _, f := range got {
		delete(want, f)
	}
	if len(want) != 0 {
		t.Fatalf("missing required fields: %+v (got %+v)", want, got)
	}
}

func TestFieldShapeChecking(t *testing.T) {
	f := Field{Name: "x", Modality: ModalityImage, Dims: []Dim{3, 224, 224}}
	if !f.IsCompatibleShape([]int{3, 224, 224}) {
		t.Fatal("exact shape must match")
	}
	if f.IsCompatibleShape([]int{3, 224}) {
		t.Fatal("length mismatch must reject")
	}
	if f.IsCompatibleShape([]int{1, 224, 224}) {
		t.Fatal("dim mismatch must reject")
	}
	g := Field{Name: "x", Dims: []Dim{DimAny, 768}}
	if !g.IsCompatibleShape([]int{42, 768}) {
		t.Fatal("DimAny must accept any value")
	}
}

func TestProcessorsRegistered(t *testing.T) {
	for _, id := range []string{"text_embedder", "code_embedder", "image_embedder",
		"audio_embedder", "video_embedder", "ts_feature_extractor"} {
		if _, err := LookupProcessor(id); err != nil {
			t.Errorf("processor %s not registered: %v", id, err)
		}
	}
}

func TestProcessorStubReturnsNoop(t *testing.T) {
	p, _ := LookupProcessor("text_embedder")
	out, err := p.Process(context.Background(), &Field{Name: "body", Modality: ModalityText}, "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Kind != EmitNoop {
		t.Fatalf("stub should emit noop, got %v", out.Kind)
	}
	if out.Metadata["stub"] != true {
		t.Fatalf("stub metadata flag missing: %+v", out.Metadata)
	}
}

func TestDefaultProcessorForModality(t *testing.T) {
	p := DefaultProcessorFor(ModalityText)
	if p == nil || p.ID() != "text_embedder" {
		t.Fatalf("text default: %+v", p)
	}
	p = DefaultProcessorFor(ModalityImage)
	if p == nil || p.ID() != "image_embedder" {
		t.Fatalf("image default: %+v", p)
	}
}

// realProcessor is the kind of impl a tenant would register to replace
// a stub. We use it to exercise SwapProcessor end-to-end.
type realProcessor struct{}

func (realProcessor) ID() string                                              { return "text_embedder" }
func (realProcessor) Modality() Modality                                      { return ModalityText }
func (realProcessor) Emits() []EmitKind                                       { return []EmitKind{EmitEmbedding} }
func (realProcessor) Init(_ context.Context, _ map[string]any) (State, error) { return nil, nil }
func (realProcessor) Process(_ context.Context, _ *Field, _ any, _ State) (*Output, error) {
	return &Output{Kind: EmitEmbedding, Value: []float32{0.1, 0.2, 0.3}}, nil
}

func TestSwapProcessor(t *testing.T) {
	if err := SwapProcessor("text_embedder", realProcessor{}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Restore the stub so other tests see the original behaviour.
		_ = SwapProcessor("text_embedder", &stubProcessor{
			id: "text_embedder", modality: ModalityText, emits: []EmitKind{EmitEmbedding},
		})
	}()

	p, _ := LookupProcessor("text_embedder")
	out, err := p.Process(context.Background(), &Field{Name: "body", Modality: ModalityText}, "hi", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out.Kind != EmitEmbedding {
		t.Fatalf("expected real embedder output, got %v", out.Kind)
	}
}

func TestSwapProcessorIDMismatch(t *testing.T) {
	if err := SwapProcessor("text_embedder", &stubProcessor{id: "wrong"}); err == nil {
		t.Fatal("should reject id mismatch")
	}
}

func TestSwapProcessorUnknownID(t *testing.T) {
	if err := SwapProcessor("not-registered", realProcessor{}); !errors.Is(err, ErrProcessorNotFound) {
		t.Fatalf("expected ErrProcessorNotFound, got %v", err)
	}
}

func TestArchitectureToSpecRoundTrips(t *testing.T) {
	a, _ := LookupArchitecture("text_signal")
	spec := a.ToSpec()
	if spec["name"] != "text_signal" {
		t.Fatalf("spec name: %v", spec["name"])
	}
	if _, ok := spec["fields"]; !ok {
		t.Fatal("spec missing fields")
	}
}
