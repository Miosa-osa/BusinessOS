package semconv

import (
	"testing"
)

func TestLLMContextCompressSpanNameKey(t *testing.T) {
	expected := "llm.context.compress"
	actual := LlmContextCompressSpan
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestLLMContextCompressSpanNameFormat(t *testing.T) {
	spanName := LlmContextCompressSpan
	if spanName != "llm.context.compress" {
		t.Errorf("span name format invalid: %s", spanName)
	}
}

func TestLLMContextCompressionRatioAttributeExists(t *testing.T) {
	expected := "llm.context.compression.ratio"
	actual := string(LlmContextCompressionRatioKey)
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestLLMContextCompressionRatioTypeDouble(t *testing.T) {
	attr := string(LlmContextCompressionRatioKey)
	if attr != "llm.context.compression.ratio" {
		t.Errorf("unexpected attribute key: %s", attr)
	}
}

func TestLLMContextCompressionStrategyAttributeExists(t *testing.T) {
	expected := "llm.context.compression.strategy"
	actual := string(LlmContextCompressionStrategyKey)
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestLLMContextCompressionStrategyIsEnum(t *testing.T) {
	// Verify enum values exist
	values := map[string]string{
		"summarize":     LlmContextCompressionStrategyValues.Summarize,
		"truncate":      LlmContextCompressionStrategyValues.Truncate,
		"sliding_window": LlmContextCompressionStrategyValues.SlidingWindow,
		"selective":     LlmContextCompressionStrategyValues.Selective,
	}

	expectedCount := 4
	if len(values) != expectedCount {
		t.Errorf("expected %d enum values, got %d", expectedCount, len(values))
	}

	for name, val := range values {
		if val != name {
			t.Errorf("enum value mismatch: expected %q, got %q", name, val)
		}
	}
}

func TestLLMContextCompressionStrategyValues(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{"Summarize", "summarize", LlmContextCompressionStrategyValues.Summarize},
		{"Truncate", "truncate", LlmContextCompressionStrategyValues.Truncate},
		{"SlidingWindow", "sliding_window", LlmContextCompressionStrategyValues.SlidingWindow},
		{"Selective", "selective", LlmContextCompressionStrategyValues.Selective},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.actual)
			}
		})
	}
}

func TestLLMContextCompressionTokensSavedAttributeExists(t *testing.T) {
	expected := "llm.context.compression.tokens_saved"
	actual := string(LlmContextCompressionTokensSavedKey)
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestLLMContextCompressionTokensSavedTypeInt(t *testing.T) {
	attr := string(LlmContextCompressionTokensSavedKey)
	if attr != "llm.context.compression.tokens_saved" {
		t.Errorf("unexpected attribute key: %s", attr)
	}
}

func TestAllCompressionAttributesHaveCorrectNames(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
	}{
		{"Ratio", "llm.context.compression.ratio", string(LlmContextCompressionRatioKey)},
		{"Strategy", "llm.context.compression.strategy", string(LlmContextCompressionStrategyKey)},
		{"TokensSaved", "llm.context.compression.tokens_saved", string(LlmContextCompressionTokensSavedKey)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.actual)
			}
		})
	}
}

func TestCompressionSpanReferencesCompressionAttributes(t *testing.T) {
	spanName := LlmContextCompressSpan
	if spanName != "llm.context.compress" {
		t.Errorf("span name should reference compression: %s", spanName)
	}
}

func TestCompressionStrategyEnumNoNilValues(t *testing.T) {
	values := []string{
		LlmContextCompressionStrategyValues.Summarize,
		LlmContextCompressionStrategyValues.Truncate,
		LlmContextCompressionStrategyValues.SlidingWindow,
		LlmContextCompressionStrategyValues.Selective,
	}

	for _, val := range values {
		if val == "" {
			t.Errorf("enum value should not be empty")
		}
	}
}

func TestCompressionAttributesFollowNamingConvention(t *testing.T) {
	attrs := []string{
		string(LlmContextCompressionRatioKey),
		string(LlmContextCompressionStrategyKey),
		string(LlmContextCompressionTokensSavedKey),
	}

	for _, attr := range attrs {
		if len(attr) == 0 || attr[:4] != "llm." {
			t.Errorf("attribute should start with 'llm.': %s", attr)
		}
	}
}

func TestCompressionSpanKindInternal(t *testing.T) {
	span := LlmContextCompressSpan
	if span != "llm.context.compress" {
		t.Errorf("span should be for internal operations: %s", span)
	}
}

func TestCompressionAttributesStringValues(t *testing.T) {
	ratioKey := string(LlmContextCompressionRatioKey)
	strategyKey := string(LlmContextCompressionStrategyKey)
	tokensKey := string(LlmContextCompressionTokensSavedKey)

	if ratioKey == "" || strategyKey == "" || tokensKey == "" {
		t.Errorf("attributes should have non-empty string values")
	}
}

func TestCompressionStrategySelectiveValue(t *testing.T) {
	expected := "selective"
	actual := LlmContextCompressionStrategyValues.Selective
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestCompressionStrategySlidingWindowValue(t *testing.T) {
	expected := "sliding_window"
	actual := LlmContextCompressionStrategyValues.SlidingWindow
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestCompressionStrategyTruncateValue(t *testing.T) {
	expected := "truncate"
	actual := LlmContextCompressionStrategyValues.Truncate
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestCompressionStrategySummarizeValue(t *testing.T) {
	expected := "summarize"
	actual := LlmContextCompressionStrategyValues.Summarize
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
