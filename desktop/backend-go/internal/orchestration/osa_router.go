package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/services"
)

// OSAIntentType represents the type of OSA intent detected
type OSAIntentType string

const (
	OSAIntentNone            OSAIntentType = "none"              // No OSA intent, handle with BusinessOS
	OSAIntentAppGeneration   OSAIntentType = "app_generation"   // Generate a full-stack application
	OSAIntentModuleCreation  OSAIntentType = "module_creation"  // Create a new BusinessOS module
	OSAIntentCodeGeneration  OSAIntentType = "code_generation"  // Generate specific code/script
	OSAIntentWorkspaceDesign OSAIntentType = "workspace_design" // 3D workspace design/materialization
)

// OSAIntent represents a detected OSA-specific intent
type OSAIntent struct {
	Type       OSAIntentType
	Confidence float64
	Reasoning  string
	ShouldRoute bool // Whether to route to OSA-5 system
}

// OSARouter determines when to route requests to OSA-5 vs BusinessOS
type OSARouter struct {
	osaClient *osa.Client
	llm       services.LLMService
}

// NewOSARouter creates a new OSA router
func NewOSARouter(osaClient *osa.Client, llm services.LLMService) *OSARouter {
	return &OSARouter{
		osaClient: osaClient,
		llm:       llm,
	}
}

// ClassifyOSAIntent determines if a request should be routed to OSA
// Uses pattern matching + optional LLM confirmation for edge cases
func (r *OSARouter) ClassifyOSAIntent(ctx context.Context, messages []services.ChatMessage) OSAIntent {
	if len(messages) == 0 {
		return OSAIntent{Type: OSAIntentNone, ShouldRoute: false}
	}

	// Get last user message
	lastMessage := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastMessage = strings.ToLower(messages[i].Content)
			break
		}
	}

	if lastMessage == "" {
		return OSAIntent{Type: OSAIntentNone, ShouldRoute: false}
	}

	slog.Debug("OSARouter: Classifying intent", "message", lastMessage)

	// High-confidence pattern matching
	intent := r.patternMatch(lastMessage)
	if intent.Confidence >= 0.8 {
		slog.Info("OSARouter: High confidence intent detected",
			"type", intent.Type,
			"confidence", intent.Confidence,
			"route_to_osa", intent.ShouldRoute)
		return intent
	}

	// Medium confidence - use LLM for confirmation (if available)
	if intent.Confidence >= 0.5 && r.llm != nil {
		slog.Debug("OSARouter: Using LLM for intent confirmation")
		confirmed := r.llmConfirmIntent(ctx, lastMessage, intent)
		if confirmed.Confidence > intent.Confidence {
			return confirmed
		}
	}

	slog.Debug("OSARouter: Low confidence, defaulting to BusinessOS",
		"type", intent.Type,
		"confidence", intent.Confidence)
	return OSAIntent{Type: OSAIntentNone, ShouldRoute: false, Confidence: intent.Confidence}
}

// patternMatch uses pattern matching to detect OSA intents
func (r *OSARouter) patternMatch(message string) OSAIntent {
	// App Generation patterns - HIGH confidence
	appGenPatterns := []string{
		"build a", "build an", "build me",
		"create a", "create an", "create me",
		"generate a", "generate an", "generate me",
		"make a", "make an", "make me",
	}
	appGenKeywords := []string{
		"app", "application", "tool", "system", "platform",
		"full-stack", "full stack", "web app", "web application",
		"mini-app", "mini app", "microservice",
	}

	for _, pattern := range appGenPatterns {
		if strings.Contains(message, pattern) {
			for _, keyword := range appGenKeywords {
				if strings.Contains(message, keyword) {
					return OSAIntent{
						Type:       OSAIntentAppGeneration,
						Confidence: 0.9,
						Reasoning:  fmt.Sprintf("Detected '%s' + '%s'", pattern, keyword),
						ShouldRoute: true,
					}
				}
			}
		}
	}

	// Module Creation patterns - HIGH confidence
	modulePatterns := []string{
		"add a module", "create a module", "build a module",
		"new module for", "add module for",
		"i need a module", "i need to track",
		"add tracking for", "track my",
	}

	for _, pattern := range modulePatterns {
		if strings.Contains(message, pattern) {
			return OSAIntent{
				Type:       OSAIntentModuleCreation,
				Confidence: 0.85,
				Reasoning:  fmt.Sprintf("Detected module pattern: '%s'", pattern),
				ShouldRoute: true,
			}
		}
	}

	// Code Generation patterns - MEDIUM confidence
	codeGenPatterns := []string{
		"write code", "generate code", "write a script",
		"create a script", "write a function", "generate a function",
		"code for", "script for", "function for",
	}
	codeGenKeywords := []string{
		"parser", "api", "endpoint", "integration",
		"automation", "workflow", "pipeline",
	}

	for _, pattern := range codeGenPatterns {
		if strings.Contains(message, pattern) {
			for _, keyword := range codeGenKeywords {
				if strings.Contains(message, keyword) {
					return OSAIntent{
						Type:       OSAIntentCodeGeneration,
						Confidence: 0.7,
						Reasoning:  fmt.Sprintf("Detected code gen: '%s' + '%s'", pattern, keyword),
						ShouldRoute: true,
					}
				}
			}
		}
	}

	// Workspace Design patterns - MEDIUM confidence
	workspacePatterns := []string{
		"materialize", "3d workspace", "workspace layout",
		"arrange workspace", "build workspace",
		"design workspace", "visualize workspace",
	}

	for _, pattern := range workspacePatterns {
		if strings.Contains(message, pattern) {
			return OSAIntent{
				Type:       OSAIntentWorkspaceDesign,
				Confidence: 0.75,
				Reasoning:  fmt.Sprintf("Detected workspace pattern: '%s'", pattern),
				ShouldRoute: true,
			}
		}
	}

	// No OSA intent detected
	return OSAIntent{
		Type:       OSAIntentNone,
		Confidence: 0.0,
		Reasoning:  "No OSA patterns matched",
		ShouldRoute: false,
	}
}

// llmConfirmIntent uses LLM to confirm intent when pattern matching is uncertain
func (r *OSARouter) llmConfirmIntent(ctx context.Context, message string, prelimIntent OSAIntent) OSAIntent {
	systemPrompt := `You are an intent classifier for OSA (Operating System Agent).

Your job: Determine if a user request should be routed to OSA-5 (app generation system) or BusinessOS (business operations).

OSA-5 handles:
- Full-stack application generation (Express.js + React)
- New module creation for BusinessOS
- Code generation (parsers, scripts, integrations)
- 3D workspace design and materialization

BusinessOS handles:
- Business document creation (proposals, SOPs, reports)
- Project management and task operations
- Client/CRM operations
- Data analysis and metrics
- General business advice and strategy

Respond with ONLY a JSON object:
{
  "intent_type": "app_generation" | "module_creation" | "code_generation" | "workspace_design" | "none",
  "confidence": 0.0-1.0,
  "reasoning": "brief explanation"
}`

	userPrompt := fmt.Sprintf(`User request: "%s"

Preliminary classification: %s (confidence: %.2f)
Reason: %s

Should this be routed to OSA-5?`, message, prelimIntent.Type, prelimIntent.Confidence, prelimIntent.Reasoning)

	messages := []services.ChatMessage{
		{Role: "user", Content: userPrompt},
	}

	// Use fast model for classification
	response, err := r.llm.ChatComplete(ctx, messages, systemPrompt)
	if err != nil {
		slog.Error("OSARouter: LLM confirmation failed", "error", err)
		return prelimIntent
	}

	// Parse LLM response (simplified - in production, use proper JSON parsing)
	// For now, just boost confidence if LLM agrees
	lowerResp := strings.ToLower(response)
	if strings.Contains(lowerResp, string(prelimIntent.Type)) {
		return OSAIntent{
			Type:       prelimIntent.Type,
			Confidence: 0.9,
			Reasoning:  "LLM confirmed: " + prelimIntent.Reasoning,
			ShouldRoute: true,
		}
	}

	// LLM disagrees, reduce confidence
	return OSAIntent{
		Type:       OSAIntentNone,
		Confidence: 0.3,
		Reasoning:  "LLM disagreed with pattern match",
		ShouldRoute: false,
	}
}

// ShouldRouteToOSA is a convenience method to check if routing is needed
func (r *OSARouter) ShouldRouteToOSA(ctx context.Context, messages []services.ChatMessage) bool {
	intent := r.ClassifyOSAIntent(ctx, messages)
	return intent.ShouldRoute
}
