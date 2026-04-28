package architecture

// Built-in architectures, ported from
// lib/optimal_engine/architecture/architectures/*.ex
//
// Each is registered via init() so importing this package side-effect-loads
// all 7 schemas. Tenants can override or extend by calling RegisterArchitecture
// with custom definitions at startup.

func init() {
	for _, build := range []func() *Architecture{
		textSignal, codeCommit, audioTranscript, imageAsset,
		multimodalMedia, structuredRecord, timeSeriesWindow,
	} {
		RegisterArchitecture(build())
	}
}

// textSignal is the engine's default architecture: free-text content with
// metadata. Matches notes / transcripts / specs / briefs / chat messages.
func textSignal() *Architecture {
	a, _ := New("text_signal", ModalityText,
		WithDescription("Free-text signal (note, transcript, doc, message)"),
		WithGranularity("document", "section", "paragraph", "sentence"),
		WithFields(
			Field{Name: "title", Modality: ModalityText, Required: true,
				Description: "Short display title"},
			Field{Name: "body", Modality: ModalityText, Required: true,
				Processor: "text_embedder", Description: "Primary textual content"},
			Field{Name: "genre", Modality: ModalityStructured,
				Description: "Speech-act genre (brief, spec, note, transcript, …)"},
			Field{Name: "authored_at", Modality: ModalityStructured,
				Description: "ISO-8601 timestamp"},
		),
	)
	return a
}

// codeCommit captures a git commit. Decomposes at commit / file / hunk / line.
func codeCommit() *Architecture {
	a, _ := New("code_commit", ModalityCode,
		WithDescription("A git commit and its file/hunk/line breakdown"),
		WithGranularity("commit", "file", "hunk", "line"),
		WithFields(
			Field{Name: "sha", Modality: ModalityStructured, Required: true},
			Field{Name: "message", Modality: ModalityText, Required: true,
				Processor: "text_embedder"},
			Field{Name: "diff", Modality: ModalityCode, Required: true,
				Processor: "code_embedder", Description: "Unified diff body"},
			Field{Name: "author", Modality: ModalityStructured},
			Field{Name: "authored_at", Modality: ModalityStructured},
		),
	)
	return a
}

// audioTranscript is what comes out of a Zoom / Teams / Riverside session.
// We treat it as text for indexing but keep the underlying audio reference
// in metadata so callers can replay the source.
func audioTranscript() *Architecture {
	a, _ := New("audio_transcript", ModalityAudio,
		WithDescription("Transcript of an audio recording (call, meeting, podcast)"),
		WithGranularity("recording", "speaker_turn", "sentence"),
		WithFields(
			Field{Name: "transcript", Modality: ModalityText, Required: true,
				Processor: "text_embedder", Description: "Plain-text transcript"},
			Field{Name: "audio_uri", Modality: ModalityAudio,
				Description: "Pointer to the source audio (s3://, file://)"},
			Field{Name: "duration_sec", Modality: ModalityStructured},
			Field{Name: "speakers", Modality: ModalityStructured,
				Description: "List of {label, display_name}"},
			Field{Name: "recorded_at", Modality: ModalityStructured},
		),
	)
	return a
}

// imageAsset is a still image with optional caption + OCR fields.
func imageAsset() *Architecture {
	a, _ := New("image_asset", ModalityImage,
		WithDescription("Still image (PNG / JPEG / SVG) with caption + OCR"),
		WithGranularity("image"),
		WithFields(
			Field{Name: "image", Modality: ModalityImage, Required: true,
				Processor: "image_embedder", Dims: []Dim{3, 224, 224}},
			Field{Name: "caption", Modality: ModalityText,
				Processor: "text_embedder"},
			Field{Name: "ocr_text", Modality: ModalityText,
				Processor: "text_embedder", Description: "OCR'd text from the image"},
			Field{Name: "captured_at", Modality: ModalityStructured},
		),
	)
	return a
}

// multimodalMedia is the catch-all for content that mixes modalities — a
// social post with text + image + video, a slide deck, a PDF with figures.
func multimodalMedia() *Architecture {
	a, _ := New("multimodal_media", ModalityText,
		WithDescription("Mixed-modality media (post, slide, PDF, gallery)"),
		WithGranularity("media", "block"),
		WithFields(
			Field{Name: "title", Modality: ModalityText, Required: true},
			Field{Name: "body", Modality: ModalityText,
				Processor: "text_embedder"},
			Field{Name: "image_refs", Modality: ModalityStructured,
				Description: "List of image URIs"},
			Field{Name: "video_refs", Modality: ModalityStructured,
				Description: "List of video URIs"},
			Field{Name: "audio_refs", Modality: ModalityStructured,
				Description: "List of audio URIs"},
		),
	)
	return a
}

// structuredRecord is a typed key-value blob — a CRM contact, an HR record,
// an invoice. The body lives entirely in the structured field.
func structuredRecord() *Architecture {
	a, _ := New("structured_record", ModalityStructured,
		WithDescription("Typed key-value record (CRM, HR, invoice, ticket meta)"),
		WithGranularity("record"),
		WithFields(
			Field{Name: "schema_id", Modality: ModalityStructured, Required: true,
				Description: "Caller-defined sub-schema id"},
			Field{Name: "data", Modality: ModalityStructured, Required: true,
				Description: "The actual key-value payload"},
			Field{Name: "summary", Modality: ModalityText,
				Processor:   "text_embedder",
				Description: "Auto-generated text summary for retrieval"},
			Field{Name: "created_at", Modality: ModalityStructured},
		),
	)
	return a
}

// timeSeriesWindow is a fixed-duration slice of a numeric stream. The
// engine indexes the extracted features (mean, slope, anomaly score) for
// retrieval since the raw samples are too dense to embed directly.
func timeSeriesWindow() *Architecture {
	a, _ := New("time_series_window", ModalityTimeSeries,
		WithDescription("Fixed-duration window of a numeric time-series stream"),
		WithGranularity("window"),
		WithFields(
			Field{Name: "stream_id", Modality: ModalityStructured, Required: true},
			Field{Name: "samples", Modality: ModalityTimeSeries, Required: true,
				Processor: "ts_feature_extractor",
				Dims:      []Dim{DimAny}, Description: "Raw numeric samples"},
			Field{Name: "start_at", Modality: ModalityStructured, Required: true},
			Field{Name: "end_at", Modality: ModalityStructured, Required: true},
			Field{Name: "sample_rate_hz", Modality: ModalityStructured},
		),
	)
	return a
}
