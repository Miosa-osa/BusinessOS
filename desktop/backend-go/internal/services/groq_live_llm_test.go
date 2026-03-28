//go:build live_llm
// +build live_llm

package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/config"
	semconv "github.com/rhl/businessos-backend/internal/semconv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestGroqLiveChatComplete_WithInferenceSpan calls the real Groq API when
// GROQ_API_KEY is set and records an llm.inference span (OTEL proof).
func TestGroqLiveChatComplete_WithInferenceSpan(t *testing.T) {
	key := os.Getenv("GROQ_API_KEY")
	if key == "" {
		t.Skip("Skipping: GROQ_API_KEY not set")
	}
	model := os.Getenv("GROQ_MODEL")
	if model == "" {
		model = "openai/gpt-oss-20b"
	}

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
		_ = tp.Shutdown(context.Background())
	})

	cfg := &config.Config{
		GroqAPIKey: key,
		GroqModel:  model,
	}
	groq := NewGroqService(cfg, model)

	tr := otel.Tracer("businessos")
	ctx, span := tr.Start(context.Background(), semconv.LlmInferenceSpan)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	out, err := groq.ChatComplete(ctx, []ChatMessage{
		{Role: "user", Content: "Reply with exactly: OK"},
	}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, out)

	span.End()
	require.NoError(t, tp.ForceFlush(context.Background()))

	spans := rec.Ended()
	var found bool
	for _, s := range spans {
		if s.Name() == semconv.LlmInferenceSpan {
			found = true
			break
		}
	}
	assert.True(t, found, "expected ended span %q", semconv.LlmInferenceSpan)
}
