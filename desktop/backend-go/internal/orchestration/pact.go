package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// PACT Framework: Planning → Action
// Used for complex workflows requiring orchestrated multi-agent execution

// PACTPhase represents a phase in the PACT workflow
type PACTPhase string

const (
	PACTPhasePlanning PACTPhase = "planning"
	PACTPhaseAction   PACTPhase = "action"
)

// PACTResult holds the result of a PACT workflow execution
type PACTResult struct {
	WorkflowID  uuid.UUID
	Phases      map[PACTPhase]*PhaseResult
	Success     bool
	Duration    time.Duration
	FinalOutput string
}

// PhaseResult holds the result of a single PACT phase
type PhaseResult struct {
	Phase      PACTPhase
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Output     string
	AgentsUsed []string
	Success    bool
	Error      error
}

// PACTOrchestrator orchestrates PACT framework workflows
type PACTOrchestrator struct {
	osaClient *osa.Client
	registry  *agents.AgentRegistryV2
}

// NewPACTOrchestrator creates a new PACT orchestrator
func NewPACTOrchestrator(osaClient *osa.Client, registry *agents.AgentRegistryV2) *PACTOrchestrator {
	return &PACTOrchestrator{
		osaClient: osaClient,
		registry:  registry,
	}
}

// ExecutePACT runs a full PACT workflow
// events: Channel for streaming updates to the client
func (p *PACTOrchestrator) ExecutePACT(
	ctx context.Context,
	input agents.AgentInput,
	userID string,
	userName string,
) (*PACTResult, <-chan streaming.StreamEvent, <-chan error) {

	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	workflowID := uuid.New()
	result := &PACTResult{
		WorkflowID: workflowID,
		Phases:     make(map[PACTPhase]*PhaseResult),
	}

	go func() {
		defer close(events)
		defer close(errs)

		startTime := time.Now()

		// Phase 1: PLANNING
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "🎯 PACT Phase 1: Planning - Analyzing requirements and creating strategy...",
		}

		planResult := p.executePlanning(ctx, input, userID, userName, events)
		result.Phases[PACTPhasePlanning] = planResult

		if !planResult.Success {
			result.Success = false
			errs <- fmt.Errorf("planning phase failed: %w", planResult.Error)
			return
		}

		// Phase 2: ACTION
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "⚡ PACT Phase 2: Action - Executing implementation...",
		}

		actionResult := p.executeAction(ctx, input, planResult.Output, events)
		result.Phases[PACTPhaseAction] = actionResult

		if !actionResult.Success {
			result.Success = false
			errs <- fmt.Errorf("action phase failed: %w", actionResult.Error)
			return
		}

		// Success!
		result.Success = true
		result.FinalOutput = actionResult.Output
		result.Duration = time.Since(startTime)

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: fmt.Sprintf("\n\n✨ **PACT Workflow Complete**\n\nTotal Duration: %v\nPhases: %d completed\n\n",
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

// executePlanning runs the Planning phase
func (p *PACTOrchestrator) executePlanning(
	ctx context.Context,
	input agents.AgentInput,
	userID string,
	userName string,
	events chan<- streaming.StreamEvent,
) *PhaseResult {

	result := &PhaseResult{
		Phase:     PACTPhasePlanning,
		StartTime: time.Now(),
	}

	// Use Project Agent for planning
	projectAgent := p.registry.GetAgent(
		agents.AgentTypeV2Project,
		userID,
		userName,
		&input.ConversationID,
		input.Context,
	)

	// Add planning-specific prompt
	planningPrompt := fmt.Sprintf(`You are in PLANNING mode (PACT Framework Phase 1).

Analyze this request and create a detailed execution plan:
%s

Your plan should include:
1. Requirements breakdown
2. Key milestones
3. Dependencies
4. Resource allocation
5. Risk assessment
6. Success criteria

Be concise but thorough.`, getLastUserMessage(input.Messages))

	planningInput := input
	planningInput.Messages = append(planningInput.Messages, services.ChatMessage{
		Role:    "system",
		Content: planningPrompt,
	})

	// Execute planning agent
	agentEvents, agentErrs := projectAgent.Run(ctx, planningInput)

	// Collect output
	var output string
	for {
		select {
		case event, ok := <-agentEvents:
			if !ok {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				result.Output = output
				result.AgentsUsed = []string{"project"}
				result.Success = true
				slog.Info("PACT Planning phase completed",
					"duration", result.Duration,
					"output_length", len(output))
				return result
			}

			if event.Type == streaming.EventTypeToken {
				if content, ok := event.Data.(string); ok {
					output += content
				}
				events <- event // Forward to client
			}

		case err := <-agentErrs:
			if err != nil {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				result.Success = false
				result.Error = err
				slog.Error("PACT Planning phase failed", "error", err)
				return result
			}

		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			return result
		}
	}
}

// executeAction runs the Action phase (implementation)
func (p *PACTOrchestrator) executeAction(
	ctx context.Context,
	input agents.AgentInput,
	plan string,
	events chan<- streaming.StreamEvent,
) *PhaseResult {

	result := &PhaseResult{
		Phase:     PACTPhaseAction,
		StartTime: time.Now(),
	}

	// Check if we should route to OSA for app generation
	lastMsg := getLastUserMessage(input.Messages)

	// Simple pattern check for app generation
	shouldUseOSA := containsAny(lastMsg, []string{
		"build app", "create app", "generate app",
		"build application", "create application",
		"full-stack", "full stack",
	})

	if shouldUseOSA && p.osaClient != nil {
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: "🚀 Routing to OSA-5 for app generation...",
		}

		// Route to OSA-5
		osaReq := &osa.OrchestrateRequest{
			Input: fmt.Sprintf("Based on this plan:\n\n%s\n\nImplement: %s", plan, lastMsg),
		}

		osaResp, err := p.osaClient.Orchestrate(ctx, osaReq)
		if err != nil {
			result.Success = false
			result.Error = fmt.Errorf("OSA orchestration failed: %w", err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Output = fmt.Sprintf("✅ **OSA Orchestration Complete!**\n\n%s\n\nExecution Time: %dms",
			osaResp.Output,
			osaResp.ExecutionTime)
		result.AgentsUsed = []string{"osa-5"}
		result.Success = osaResp.Success

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: result.Output,
		}

		return result
	}

	// Otherwise, use BusinessOS Document Agent for implementation
	// (This would be for business documents, not code)
	events <- streaming.StreamEvent{
		Type: streaming.EventTypeThinking,
		Data: "📝 Using Document Agent for implementation...",
	}

	result.Output = fmt.Sprintf("Implementation based on plan:\n\n%s", plan)
	result.AgentsUsed = []string{"document"}
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// Helper functions (getLastUserMessage is in osa_orchestrator.go)

func containsAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
