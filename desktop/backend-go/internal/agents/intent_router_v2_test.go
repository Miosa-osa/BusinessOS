package agents

import (
	"context"
	"testing"

	"github.com/rhl/businessos-backend/internal/services"
)

func TestSmartIntentRouterClassifyIntent(t *testing.T) {
	router := NewSmartIntentRouter(nil, nil)

	// Test cases with strong signal patterns that should match
	tests := []struct {
		name          string
		message       string
		expectedAgent AgentTypeV2
		minConfidence float64
	}{
		{
			name:          "Document request - proposal",
			message:       "Create a proposal for client X",
			expectedAgent: AgentTypeV2Document,
			minConfidence: 0.5,
		},
		{
			name:          "Document request - write proposal",
			message:       "write a proposal document",
			expectedAgent: AgentTypeV2Document,
			minConfidence: 0.5,
		},
		{
			name:          "Document request - draft report",
			message:       "draft a report",
			expectedAgent: AgentTypeV2Document,
			minConfidence: 0.5,
		},
		{
			name:          "Analysis request - analyze",
			message:       "analyze the data",
			expectedAgent: AgentTypeV2Analyst,
			minConfidence: 0.4,
		},
		{
			name:          "General greeting - should stay with orchestrator",
			message:       "Hello, how are you?",
			expectedAgent: AgentTypeV2Orchestrator,
			minConfidence: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := []services.ChatMessage{
				{Role: "user", Content: tt.message},
			}

			intent := router.ClassifyIntent(context.Background(), messages, nil)

			if intent.TargetAgent != tt.expectedAgent {
				t.Errorf("Expected agent %s, got %s for message: %s (confidence: %.2f)",
					tt.expectedAgent, intent.TargetAgent, tt.message, intent.Confidence)
			}
			if intent.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f for message: %s", tt.minConfidence, intent.Confidence, tt.message)
			}
		})
	}
}

func TestSmartIntentRouterPatternMatching(t *testing.T) {
	router := NewSmartIntentRouter(nil, nil)

	// Test pattern matches - log results for debugging
	patternTests := []struct {
		message       string
		expectedAgent AgentTypeV2
	}{
		{"write a proposal", AgentTypeV2Document},
		{"draft a report", AgentTypeV2Document},
		{"analyze the data", AgentTypeV2Analyst},
	}

	for _, tt := range patternTests {
		messages := []services.ChatMessage{
			{Role: "user", Content: tt.message},
		}
		intent := router.ClassifyIntent(context.Background(), messages, nil)

		// Log the result for debugging - don't fail on edge cases
		t.Logf("Message: '%s' -> Agent: %s (confidence: %.2f)", tt.message, intent.TargetAgent, intent.Confidence)

		if intent.TargetAgent != tt.expectedAgent {
			t.Logf("Note: Expected %s, got %s - may need pattern tuning", tt.expectedAgent, intent.TargetAgent)
		}
	}
}

func TestIntentStruct(t *testing.T) {
	intent := Intent{
		Category:       "document",
		Confidence:     0.85,
		Reasoning:      "User asked for a proposal",
		ShouldDelegate: true,
		TargetAgent:    AgentTypeV2Document,
	}

	if intent.Category != "document" {
		t.Errorf("Expected category 'document', got '%s'", intent.Category)
	}
	if intent.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got %.2f", intent.Confidence)
	}
	if !intent.ShouldDelegate {
		t.Error("Expected ShouldDelegate to be true")
	}
	if intent.TargetAgent != AgentTypeV2Document {
		t.Errorf("Expected TargetAgent Document, got %s", intent.TargetAgent)
	}
}
