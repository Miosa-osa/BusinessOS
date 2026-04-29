// audio_embedder_processor.go — wraps the local WhisperService so audio
// fields produce a transcript (downstream stages may then re-route the
// transcript through text_embedder for retrieval).
//
// We emit EmitTranscription, not EmitEmbedding. This matches the Elixir
// processor architecture: an audio "embedder" really produces a transcript
// first; vector embedding is a downstream concern. Keeping the kinds
// distinct lets the pipeline route each emission to the right table.
package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Miosa-osa/OptimalEngine-go/architecture"
)

// AudioEmbedderProcessor turns audio bytes into a transcript via Whisper.
// The "embedder" name is kept for symmetry with the architecture-layer
// stub; what we actually emit is EmitTranscription.
type AudioEmbedderProcessor struct {
	whisper *WhisperService
}

// NewAudioEmbedderProcessor returns a processor wired to WhisperService.
// Caller (bootstrap) only swaps when whisper.IsAvailable() reports true,
// so we don't have to defend against nil whisper here.
func NewAudioEmbedderProcessor(svc *WhisperService) *AudioEmbedderProcessor {
	return &AudioEmbedderProcessor{whisper: svc}
}

func (p *AudioEmbedderProcessor) ID() string                      { return "audio_embedder" }
func (p *AudioEmbedderProcessor) Modality() architecture.Modality { return architecture.ModalityAudio }

func (p *AudioEmbedderProcessor) Emits() []architecture.EmitKind {
	return []architecture.EmitKind{architecture.EmitTranscription}
}

func (p *AudioEmbedderProcessor) Init(_ context.Context, _ map[string]any) (architecture.State, error) {
	return nil, nil
}

// Process accepts raw audio bytes and returns the transcript text. The
// upstream architecture caller is responsible for materializing bytes
// from a URI when needed; the processor doesn't fetch.
//
// Format detection: we default to "wav" when the field metadata doesn't
// carry one. Whisper's CLI accepts most formats anyway.
func (p *AudioEmbedderProcessor) Process(ctx context.Context, field *architecture.Field, value any, _ architecture.State) (*architecture.Output, error) {
	if p.whisper == nil {
		return nil, errors.New("audio_embedder: whisper service not wired")
	}
	if !p.whisper.IsAvailable() {
		return nil, errors.New("audio_embedder: whisper unavailable (model/binary missing)")
	}
	if field == nil {
		return nil, errors.New("audio_embedder: nil field")
	}
	raw, err := bytesFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("audio_embedder: field %q: %w", field.Name, err)
	}
	if len(raw) == 0 {
		return &architecture.Output{
			Kind:     architecture.EmitNoop,
			Metadata: map[string]any{"reason": "empty input", "field": field.Name},
		}, nil
	}

	format := "wav"
	if f, ok := field.Description, true; ok && f != "" {
		// We can't easily put per-field hints in architecture.Field today;
		// future: read field.Metadata["audio_format"] when available.
		_ = f
	}

	var reader io.Reader = bytes.NewReader(raw)
	transcription, err := p.whisper.Transcribe(ctx, reader, format)
	if err != nil {
		return nil, fmt.Errorf("audio_embedder: transcribe %q: %w", field.Name, err)
	}
	return &architecture.Output{
		Kind:  architecture.EmitTranscription,
		Value: transcription.Text,
		Metadata: map[string]any{
			"field":           field.Name,
			"engine":          "whisper.cpp",
			"audio_format":    format,
			"audio_byte_size": len(raw),
		},
	}, nil
}

var _ architecture.Processor = (*AudioEmbedderProcessor)(nil)
