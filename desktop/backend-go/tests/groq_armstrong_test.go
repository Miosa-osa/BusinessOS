//go:build integration

package tests

// Chicago TDD Armstrong Fault-Tolerance Tests — Real Groq API
//
// Five tests prove Armstrong principles hold for GroqService:
//   1. Let-It-Crash: invalid key returns explicit error (not swallowed)
//   2. Observability: OTEL span propagates with provider+model attributes
//   3. Channel Safety: streaming cancellation produces no goroutine leak
//   4. Timeout Budget: HTTP client has bounded timeout [30s, 180s]
//   5. Concurrent Safety: 5 goroutines share one GroqService safely (-race)
//
// Run: GROQ_API_KEY=gsk_... go test -tags=integration -race ./tests/... -run TestGroqArmstrong -v -timeout 120s

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/rhl/businessos-backend/internal/config"
	bossemconv "github.com/rhl/businessos-backend/internal/semconv"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/sync/errgroup"
)

// findAttr looks up an attribute by key in a slice of KeyValues.
func findAttr(attrs []attribute.KeyValue, key attribute.Key) (attribute.Value, bool) {
	for _, kv := range attrs {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return attribute.Value{}, false
}

// ── Test 1: Let-It-Crash — invalid key must return explicit error ─────────────

// TestGroqArmstrong_LetItCrash_InvalidKey proves that GroqService propagates
// Groq API auth errors rather than swallowing them (Armstrong: fail fast).
//
// No GROQ_API_KEY required — this test deliberately uses an invalid key.
func TestGroqArmstrong_LetItCrash_InvalidKey(t *testing.T) {
	cfg := &config.Config{
		GroqAPIKey: "gsk_INVALID_KEY_FOR_ARMSTRONG_TEST_DO_NOT_USE",
		GroqModel:  "openai/gpt-oss-20b",
	}
	svc := services.NewGroqService(cfg, "openai/gpt-oss-20b")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err := svc.ChatComplete(ctx, []services.ChatMessage{
		{Role: "user", Content: "ping"},
	}, "")

	// Armstrong: error MUST be returned — not nil, not swallowed
	require.Error(t, err,
		"Expected authentication error from Groq API with invalid key, got nil — "+
			"GroqService is swallowing errors (Armstrong violation)")

	assert.Empty(t, content,
		"Content must be empty when auth fails, got: %q", content)

	// Error must indicate authentication failure (HTTP 401) or connection
	errMsg := err.Error()
	isAuthErr := containsAny(errMsg, "401", "Unauthorized", "unauthorized", "invalid_api_key")
	isNetErr := containsAny(errMsg, "connection", "dial", "network", "EOF")
	assert.True(t, isAuthErr || isNetErr,
		"Expected 401 auth error or network error, got: %q", errMsg)

	t.Logf("Armstrong Let-It-Crash confirmed — error: %v", err)
}

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// ── Test 2: OTEL span propagation ─────────────────────────────────────────────

// TestGroqArmstrong_OTELSpanPropagate proves that a real Groq call can be
// wrapped in an OTEL span with correct llm.provider and llm.model attributes.
func TestGroqArmstrong_OTELSpanPropagate(t *testing.T) {
	apiKey := getGroqAPIKey(t) // skips if GROQ_API_KEY not set
	model := getGroqModel()

	// Install in-process span recorder
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		_ = tp.Shutdown(context.Background())
	})

	cfg := &config.Config{GroqAPIKey: apiKey, GroqModel: model}
	svc := services.NewGroqService(cfg, model)
	tr := otel.Tracer("businessos.armstrong.test")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start span, decorate with LLM semantic convention attributes
	ctx, span := tr.Start(ctx, bossemconv.LlmInferenceSpan)
	span.SetAttributes(
		bossemconv.ServiceName("businessos"),
		bossemconv.LlmProvider("groq"),
		bossemconv.LlmModel(model),
	)

	content, err := svc.ChatComplete(ctx, []services.ChatMessage{
		{Role: "user", Content: "Reply with exactly: OK"},
	}, "")
	span.End()

	// Force flush before inspecting
	require.NoError(t, tp.ForceFlush(context.Background()))

	// Real call must succeed
	require.NoError(t, err, "Real Groq call failed — check GROQ_API_KEY")
	assert.NotEmpty(t, content, "Groq returned empty content")

	// Find our span in recorded spans
	ended := rec.Ended()
	var foundSpan *sdktrace.ReadOnlySpan
	for i := range ended {
		s := ended[i]
		if s.Name() == bossemconv.LlmInferenceSpan {
			foundSpan = &s
			break
		}
	}
	require.NotNil(t, foundSpan,
		"Expected span named %q in recorded spans, got %d spans: %v",
		bossemconv.LlmInferenceSpan, len(ended), spanNames(ended))

	attrs := (*foundSpan).Attributes()

	// Verify service.name attribute
	svcVal, ok := findAttr(attrs, bossemconv.ServiceNameKey)
	assert.True(t, ok, "span missing service.name attribute")
	assert.Equal(t, "businessos", svcVal.AsString())

	// Verify llm.provider attribute
	provVal, ok := findAttr(attrs, bossemconv.LlmProviderKey)
	assert.True(t, ok, "span missing llm.provider attribute")
	assert.Equal(t, "groq", provVal.AsString())

	// Verify llm.model attribute
	modelVal, ok := findAttr(attrs, bossemconv.LlmModelKey)
	assert.True(t, ok, "span missing llm.model attribute")
	assert.Equal(t, model, modelVal.AsString())

	t.Logf("OTEL span confirmed: name=%s provider=%s model=%s",
		(*foundSpan).Name(), provVal.AsString(), modelVal.AsString())
}

// spanNames extracts names from a slice of ReadOnlySpans for error messages.
func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	names := make([]string, len(spans))
	for i, s := range spans {
		names[i] = s.Name()
	}
	return names
}

// ── Test 3: Stream clean close — no goroutine leak ────────────────────────────

// TestGroqArmstrong_StreamCleanClose proves that cancelling a streaming Groq
// call mid-stream does not leak goroutines (Armstrong: no orphaned processes).
func TestGroqArmstrong_StreamCleanClose(t *testing.T) {
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	cfg := &config.Config{GroqAPIKey: apiKey, GroqModel: model}
	svc := services.NewGroqService(cfg, model)

	baseline := runtime.NumGoroutine()

	// Use a cancellable context so we control termination precisely
	ctx, cancel := context.WithCancel(context.Background())
	// Safety deadline so test can't hang
	ctx, deadline := context.WithTimeout(ctx, 30*time.Second)
	defer deadline()

	chunks, errs := svc.StreamChat(ctx, []services.ChatMessage{
		{Role: "user", Content: "Count slowly from 1 to 100, one number per line."},
	}, "")

	// Consume at least one chunk to prove stream started
	var firstChunk string
	select {
	case c, ok := <-chunks:
		if ok {
			firstChunk = c
		}
	case err := <-errs:
		// If error before first chunk, skip goroutine leak check (network issue)
		t.Skipf("Stream errored before first chunk: %v", err)
	case <-time.After(15 * time.Second):
		t.Fatal("Timed out waiting for first streaming chunk from Groq")
	}

	// Mid-stream cancellation — the Armstrong moment
	cancel()

	// Drain both channels to unblock the goroutine
	for range chunks {
	}
	for range errs {
	}

	// Give the Go scheduler time to clean up
	time.Sleep(300 * time.Millisecond)

	afterCancel := runtime.NumGoroutine()

	assert.NotEmpty(t, firstChunk, "Stream should have produced at least one chunk before cancel")
	assert.LessOrEqual(t, afterCancel, baseline+2,
		"Goroutine leak: before=%d after=%d (tolerance=2) — streaming goroutine not cleaned up",
		baseline, afterCancel)

	t.Logf("Goroutine count: baseline=%d after_cancel=%d first_chunk=%q",
		baseline, afterCancel, firstChunk[:min(len(firstChunk), 30)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Test 4: Timeout budget — HTTP client must have bounded timeout ────────────

// TestGroqArmstrong_TimeoutBudget proves the GroqService HTTP client has a
// bounded timeout — not zero (infinite) and not too aggressive.
//
// Armstrong: every operation has a resource budget. An infinite HTTP timeout
// is a resource leak waiting to happen under network partition.
//
// No API key needed — structural inspection only.
func TestGroqArmstrong_TimeoutBudget(t *testing.T) {
	cfg := &config.Config{
		GroqAPIKey: "any-key-timeout-test",
		GroqModel:  "openai/gpt-oss-20b",
	}
	svc := services.NewGroqService(cfg, "openai/gpt-oss-20b")

	// Access unexported "client" field via reflect + unsafe
	// This is the correct Go idiom for structural validation from outside-package tests
	val := reflect.ValueOf(svc).Elem()
	clientField := val.FieldByName("client")
	require.True(t, clientField.IsValid(), "GroqService has no 'client' field — struct changed?")

	clientPtr := (*http.Client)(unsafe.Pointer(clientField.UnsafeAddr()))
	require.NotNil(t, clientPtr, "GroqService.client is nil — HTTP client not initialized")

	timeout := clientPtr.Timeout

	// Armstrong budget constraints:
	assert.NotZero(t, timeout,
		"HTTP client timeout must not be zero (infinite) — "+
			"GroqService will hang under network partition")

	assert.GreaterOrEqual(t, timeout, 30*time.Second,
		"HTTP timeout %v is too aggressive — LLM responses can take 10-20s for complex prompts", timeout)

	assert.LessOrEqual(t, timeout, 180*time.Second,
		"HTTP timeout %v is too permissive — will hold connections open too long under failure", timeout)

	t.Logf("HTTP client timeout: %v (within Armstrong budget [30s, 180s])", timeout)
}

// ── Test 5: Concurrent safety — 5 goroutines, one service, -race ─────────────

// TestGroqArmstrong_ConcurrentSafety proves that GroqService is safe for
// concurrent use by 5 goroutines without data races.
//
// Armstrong: no shared mutable state — http.Client is safe for concurrent use
// per Go docs; this test forces the race detector to observe that invariant.
//
// Run with: go test -tags=integration -race ./tests/... -run TestGroqArmstrong_ConcurrentSafety
func TestGroqArmstrong_ConcurrentSafety(t *testing.T) {
	// Note: free-tier Groq may return 429 on highly concurrent calls.
	// Use a paid key for reliable concurrent testing.
	apiKey := getGroqAPIKey(t)
	model := getGroqModel()

	cfg := &config.Config{GroqAPIKey: apiKey, GroqModel: model}

	// ONE shared GroqService instance across all goroutines — this is the race trigger
	svc := services.NewGroqService(cfg, model)

	const workers = 5
	results := make([]string, workers)
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	for i := 0; i < workers; i++ {
		i := i // capture loop variable
		g.Go(func() error {
			content, err := svc.ChatComplete(ctx, []services.ChatMessage{
				{
					Role: "user",
					Content: fmt.Sprintf(
						"Worker %d: What is %d + %d? Single number only.",
						i, i*10, i*5,
					),
				},
			}, "")
			if err != nil {
				return fmt.Errorf("worker %d: %w", i, err)
			}
			mu.Lock()
			results[i] = content
			mu.Unlock()
			return nil
		})
	}

	err := g.Wait()
	require.NoError(t, err,
		"All %d concurrent Groq calls must succeed — got error: %v "+
			"(if 429: use paid Groq key for concurrent testing)", workers, err)

	for i, r := range results {
		assert.NotEmpty(t, r, "Worker %d returned empty result", i)
	}

	t.Logf("Concurrent results (%d workers): %v", workers, results)
}
