package orchestration

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// OSAOrchestrator coordinates OSA routing and workflow execution
type OSAOrchestrator struct {
	osaClient *osa.Client
	osaRouter *OSARouter
	registry  *agents.AgentRegistryV2
	llm       services.LLMService
}

// NewOSAOrchestrator creates a new OSA orchestrator
func NewOSAOrchestrator(
	osaClient *osa.Client,
	registry *agents.AgentRegistryV2,
	llm services.LLMService,
) *OSAOrchestrator {
	return &OSAOrchestrator{
		osaClient: osaClient,
		osaRouter: NewOSARouter(osaClient, llm),
		registry:  registry,
		llm:       llm,
	}
}

// ProcessWithOSARouting determines routing and executes appropriate workflow
// Returns: (shouldContinue, events, errors)
// If shouldContinue is false, the caller should stop and use the returned events
func (o *OSAOrchestrator) ProcessWithOSARouting(
	ctx context.Context,
	input agents.AgentInput,
	userID string,
	userName string,
) (bool, <-chan streaming.StreamEvent, <-chan error) {

	// Check if OSA client is configured
	if o.osaClient == nil {
		slog.Debug("OSA client not configured, skipping OSA routing")
		return true, nil, nil // Continue with normal BusinessOS routing
	}

	// Classify intent
	intent := o.osaRouter.ClassifyOSAIntent(ctx, input.Messages)

	slog.Info("OSA intent classification",
		"type", intent.Type,
		"confidence", intent.Confidence,
		"should_route", intent.ShouldRoute)

	// If no OSA intent, continue with normal BusinessOS flow
	if !intent.ShouldRoute {
		return true, nil, nil
	}

	// OSA intent detected - determine workflow type
	return o.routeToOSAWorkflow(ctx, input, intent, userID, userName)
}

// routeToOSAWorkflow routes to appropriate OSA workflow
func (o *OSAOrchestrator) routeToOSAWorkflow(
	ctx context.Context,
	input agents.AgentInput,
	intent OSAIntent,
	userID string,
	userName string,
) (bool, <-chan streaming.StreamEvent, <-chan error) {

	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Send initial routing notification
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: fmt.Sprintf("🎯 Detected %s intent (confidence: %.0f%%) - Routing to OSA workflow...",
				intent.Type,
				intent.Confidence*100),
		}

		// Determine workflow based on intent type
		switch intent.Type {
		case OSAIntentAppGeneration:
			// Use PACT framework for app generation
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: "📋 Using PACT Framework (Planning → Action)",
			}

			pactOrch := NewPACTOrchestrator(o.osaClient, o.registry)
			_, pactEvents, pactErrs := pactOrch.ExecutePACT(ctx, input, userID, userName)
			o.forwardEvents(ctx, pactEvents, pactErrs, events, errs)

		case OSAIntentModuleCreation:
			// Use BMAD method for module/product creation
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: "🏗️ Using BMAD Method (Business → Development)",
			}

			bmadOrch := NewBMADOrchestrator(o.osaClient, o.registry)
			_, bmadEvents, bmadErrs := bmadOrch.ExecuteBMAD(ctx, input, userID, userName)
			o.forwardEvents(ctx, bmadEvents, bmadErrs, events, errs)

		case OSAIntentCodeGeneration:
			// Direct OSA orchestration for code generation
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: "💻 Routing to OSA for code generation...",
			}

			o.executeDirectOSA(ctx, input, events, errs)

		case OSAIntentWorkspaceDesign:
			// Workspace design (future implementation)
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeToken,
				Data: "🏗️ Workspace design is planned for future release. For now, using standard workflow.",
			}
			events <- streaming.StreamEvent{Type: streaming.EventTypeDone}

		default:
			// Fallback to normal BusinessOS
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeToken,
				Data: "Unrecognized OSA intent. Using standard BusinessOS workflow.",
			}
			events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
		}
	}()

	return false, events, errs // shouldContinue=false (OSA handled it)
}

// executeDirectOSA calls OSA API directly
func (o *OSAOrchestrator) executeDirectOSA(
	ctx context.Context,
	input agents.AgentInput,
	events chan<- streaming.StreamEvent,
	errs chan<- error,
) {
	lastMsg := getLastUserMessage(input.Messages)

	req := &osa.OrchestrateRequest{
		Input: lastMsg,
	}

	resp, err := o.osaClient.Orchestrate(ctx, req)
	if err != nil {
		errs <- fmt.Errorf("OSA orchestration failed: %w", err)
		return
	}

	// Send success result
	events <- streaming.StreamEvent{
		Type: streaming.EventTypeToken,
		Data: fmt.Sprintf("✅ **OSA Orchestration Complete!**\n\n%s\n\nExecution Time: %dms",
			resp.Output,
			resp.ExecutionTime),
	}

	events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
}

// forwardEvents forwards events and errors from child workflow to parent channels
func (o *OSAOrchestrator) forwardEvents(
	ctx context.Context,
	childEvents <-chan streaming.StreamEvent,
	childErrs <-chan error,
	parentEvents chan<- streaming.StreamEvent,
	parentErrs chan<- error,
) {
	for {
		select {
		case event, ok := <-childEvents:
			if !ok {
				return
			}
			parentEvents <- event

		case err := <-childErrs:
			if err != nil {
				parentErrs <- err
			}
			return

		case <-ctx.Done():
			parentErrs <- ctx.Err()
			return
		}
	}
}

// Helper function to get last user message
func getLastUserMessage(messages []services.ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}
