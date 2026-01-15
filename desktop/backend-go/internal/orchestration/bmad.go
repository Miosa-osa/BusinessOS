package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// BMAD Method: Business → Development
// Used for new project/product development workflows

// BMADPhase represents a phase in the BMAD workflow
type BMADPhase string

const (
	BMADPhaseBusiness    BMADPhase = "business"
	BMADPhaseDevelopment BMADPhase = "development"
)

// BMADResult holds the result of a BMAD workflow execution
type BMADResult struct {
	WorkflowID  uuid.UUID
	Phases      map[BMADPhase]*BMADPhaseResult
	Success     bool
	Duration    time.Duration
	FinalOutput string
}

// BMADPhaseResult holds the result of a single BMAD phase
type BMADPhaseResult struct {
	Phase      BMADPhase
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Output     string
	AgentsUsed []string
	Success    bool
	Error      error
}

// BMADOrchestrator orchestrates BMAD method workflows
type BMADOrchestrator struct {
	osaClient *osa.Client
	registry  *agents.AgentRegistryV2
}

// NewBMADOrchestrator creates a new BMAD orchestrator
func NewBMADOrchestrator(osaClient *osa.Client, registry *agents.AgentRegistryV2) *BMADOrchestrator {
	return &BMADOrchestrator{
		osaClient: osaClient,
		registry:  registry,
	}
}

// ExecuteBMAD runs a full BMAD workflow
func (b *BMADOrchestrator) ExecuteBMAD(
	ctx context.Context,
	input agents.AgentInput,
	userID string,
	userName string,
) (*BMADResult, <-chan streaming.StreamEvent, <-chan error) {

	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	workflowID := uuid.New()
	result := &BMADResult{
		WorkflowID: workflowID,
		Phases:     make(map[BMADPhase]*BMADPhaseResult),
	}

	go func() {
		defer close(events)
		defer close(errs)

		startTime := time.Now()

		// Phase 1: BUSINESS
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "💼 BMAD Phase 1: Business - Defining requirements and business goals...",
		}

		businessResult := b.executeBusiness(ctx, input, userID, userName, events)
		result.Phases[BMADPhaseBusiness] = businessResult

		if !businessResult.Success {
			result.Success = false
			errs <- fmt.Errorf("business phase failed: %w", businessResult.Error)
			return
		}

		// Phase 2: DEVELOPMENT
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "⚙️ BMAD Phase 2: Development - Implementing the system...",
		}

		devResult := b.executeDevelopment(ctx, input, businessResult.Output, events)
		result.Phases[BMADPhaseDevelopment] = devResult

		if !devResult.Success {
			result.Success = false
			errs <- fmt.Errorf("development phase failed: %w", devResult.Error)
			return
		}

		// Success!
		result.Success = true
		result.FinalOutput = devResult.Output
		result.Duration = time.Since(startTime)

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: fmt.Sprintf("\n\n✨ **BMAD Workflow Complete**\n\nTotal Duration: %v\nPhases: %d completed\n\n",
				result.Duration,
				len(result.Phases)),
		}

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: result.FinalOutput,
		}

		events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
	}()

	return result, events, errs
}

// executeBusiness runs the Business phase
func (b *BMADOrchestrator) executeBusiness(
	ctx context.Context,
	input agents.AgentInput,
	userID string,
	userName string,
	events chan<- streaming.StreamEvent,
) *BMADPhaseResult {

	result := &BMADPhaseResult{
		Phase:     BMADPhaseBusiness,
		StartTime: time.Now(),
	}

	// Use Analyst Agent for business analysis
	analystAgent := b.registry.GetAgent(
		agents.AgentTypeV2Analyst,
		userID,
		userName,
		&input.ConversationID,
		input.Context,
	)

	// Add business-specific prompt
	businessPrompt := fmt.Sprintf(`You are in BUSINESS ANALYSIS mode (BMAD Framework Phase 1).

Analyze this business requirement:
%s

Provide:
1. Business objectives
2. Target users/stakeholders
3. Success metrics
4. Value proposition
5. Market context
6. Constraints and assumptions

Be analytical and strategic.`, getLastUserMessage(input.Messages))

	businessInput := input
	businessInput.Messages = append(businessInput.Messages, services.ChatMessage{
		Role:    "system",
		Content: businessPrompt,
	})

	// Execute analyst agent
	agentEvents, agentErrs := analystAgent.Run(ctx, businessInput)

	// Collect output
	var output string
	for {
		select {
		case event, ok := <-agentEvents:
			if !ok {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				result.Output = output
				result.AgentsUsed = []string{"analyst"}
				result.Success = true
				slog.Info("BMAD Business phase completed",
					"duration", result.Duration,
					"output_length", len(output))
				return result
			}

			if event.Type == streaming.EventTypeToken {
				if content, ok := event.Data.(string); ok {
					output += content
				}
				events <- event
			}

		case err := <-agentErrs:
			if err != nil {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				result.Success = false
				result.Error = err
				slog.Error("BMAD Business phase failed", "error", err)
				return result
			}

		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			return result
		}
	}
}

// executeDevelopment runs the Development phase (implementation via OSA)
func (b *BMADOrchestrator) executeDevelopment(
	ctx context.Context,
	input agents.AgentInput,
	businessAnalysis string,
	events chan<- streaming.StreamEvent,
) *BMADPhaseResult {

	result := &BMADPhaseResult{
		Phase:      BMADPhaseDevelopment,
		StartTime:  time.Now(),
		AgentsUsed: []string{"osa-5"},
	}

	// Route to OSA for implementation
	if b.osaClient != nil {
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "🚀 Routing to OSA-5 for implementation...",
		}

		lastMsg := getLastUserMessage(input.Messages)

		osaReq := &osa.OrchestrateRequest{
			Input: fmt.Sprintf(`Build application based on BMAD business analysis:

BUSINESS REQUIREMENTS:
%s

USER REQUEST:
%s

Generate complete application with all components.`,
				truncate(businessAnalysis, 500),
				lastMsg),
		}

		osaResp, err := b.osaClient.Orchestrate(ctx, osaReq)
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("OSA orchestration failed: %w", err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}

		result.Output = fmt.Sprintf("✅ **OSA Orchestration Complete!**\n\n%s\n\nExecution Time: %dms",
			osaResp.Output,
			osaResp.ExecutionTime)
		result.Success = osaResp.Success
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: result.Output,
		}

		return result
	}

	// Fallback if OSA not available
	result.Output = "Development phase - OSA client not configured"
	result.Success = false
	result.Error = fmt.Errorf("OSA client not available")
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// truncate limits string length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
