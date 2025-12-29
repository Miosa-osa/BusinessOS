package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ========== CHAIN OF THOUGHT (COT) SYSTEM ==========

// ThoughtStep represents a single step in the chain of thought
type ThoughtStep struct {
	ID          string        `json:"id"`
	Agent       AgentTypeV2   `json:"agent"`
	Action      string        `json:"action"`     // "analyze", "delegate", "execute", "synthesize"
	Input       string        `json:"input"`      // What this step received
	Output      string        `json:"output"`     // What this step produced
	Reasoning   string        `json:"reasoning"`  // Why this step was taken
	Confidence  float64       `json:"confidence"` // 0-1 confidence in this step
	Duration    time.Duration `json:"duration"`   // How long this step took
	Children    []string      `json:"children"`   // IDs of child steps (for parallel execution)
	Status      string        `json:"status"`     // "pending", "running", "completed", "failed"
	Error       string        `json:"error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

// ChainOfThought tracks the full reasoning chain
type ChainOfThought struct {
	ID             string         `json:"id"`
	UserMessage    string         `json:"user_message"`
	Steps          []*ThoughtStep `json:"steps"`
	FinalOutput    string         `json:"final_output"`
	TotalDuration  time.Duration  `json:"total_duration"`
	AgentsInvolved []AgentTypeV2  `json:"agents_involved"`
	Status         string         `json:"status"` // "planning", "executing", "synthesizing", "completed"
	mu             sync.RWMutex
}

// NewChainOfThought creates a new COT tracker
func NewChainOfThought(userMessage string) *ChainOfThought {
	return &ChainOfThought{
		ID:          uuid.New().String(),
		UserMessage: userMessage,
		Steps:       make([]*ThoughtStep, 0),
		Status:      "planning",
	}
}

// AddStep adds a new step to the chain
func (cot *ChainOfThought) AddStep(step *ThoughtStep) {
	cot.mu.Lock()
	defer cot.mu.Unlock()
	step.ID = uuid.New().String()
	step.CreatedAt = time.Now()
	step.Status = "pending"
	cot.Steps = append(cot.Steps, step)
}

// UpdateStep updates an existing step
func (cot *ChainOfThought) UpdateStep(stepID string, output string, status string) {
	cot.mu.Lock()
	defer cot.mu.Unlock()
	for _, step := range cot.Steps {
		if step.ID == stepID {
			step.Output = output
			step.Status = status
			if status == "completed" || status == "failed" {
				now := time.Now()
				step.CompletedAt = &now
				step.Duration = now.Sub(step.CreatedAt)
			}
			break
		}
	}
}

// GetAgentsInvolved returns unique list of agents used
func (cot *ChainOfThought) GetAgentsInvolved() []AgentTypeV2 {
	cot.mu.RLock()
	defer cot.mu.RUnlock()
	agentMap := make(map[AgentTypeV2]bool)
	for _, step := range cot.Steps {
		agentMap[step.Agent] = true
	}
	agents := make([]AgentTypeV2, 0, len(agentMap))
	for agent := range agentMap {
		agents = append(agents, agent)
	}
	return agents
}

// ========== AGENT MESSAGE SYSTEM ==========

// AgentMessage represents a message passed between agents
type AgentMessage struct {
	ID        string            `json:"id"`
	From      AgentTypeV2       `json:"from"`
	To        AgentTypeV2       `json:"to"`
	Type      AgentMessageType  `json:"type"`
	Content   string            `json:"content"`
	Context   map[string]string `json:"context,omitempty"`
	RequestID string            `json:"request_id"` // Links related messages
	InReplyTo string            `json:"in_reply_to,omitempty"`
	Priority  int               `json:"priority"` // 1=low, 2=normal, 3=high, 4=urgent
	CreatedAt time.Time         `json:"created_at"`
}

// AgentMessageType defines the type of inter-agent message
type AgentMessageType string

const (
	MsgTypeRequest    AgentMessageType = "request"    // Asking another agent to do something
	MsgTypeResponse   AgentMessageType = "response"   // Response to a request
	MsgTypeDelegate   AgentMessageType = "delegate"   // Delegating a task
	MsgTypeInform     AgentMessageType = "inform"     // Sharing information
	MsgTypeQuery      AgentMessageType = "query"      // Asking for information
	MsgTypeConfirm    AgentMessageType = "confirm"    // Confirming receipt/action
	MsgTypeSynthesize AgentMessageType = "synthesize" // Request to combine outputs
)

// NewAgentMessage creates a new inter-agent message
func NewAgentMessage(from, to AgentTypeV2, msgType AgentMessageType, content string) *AgentMessage {
	return &AgentMessage{
		ID:        uuid.New().String(),
		From:      from,
		To:        to,
		Type:      msgType,
		Content:   content,
		Context:   make(map[string]string),
		Priority:  2, // Normal priority
		CreatedAt: time.Now(),
	}
}

// ========== ORCHESTRATOR COT ENGINE ==========

// OrchestratorCOT manages the chain of thought execution
type OrchestratorCOT struct {
	pool     *pgxpool.Pool
	config   *config.Config
	registry *AgentRegistryV2
	router   *SmartIntentRouter

	// Message bus for inter-agent communication
	messageBus chan *AgentMessage

	// Active chains
	activeChains map[string]*ChainOfThought
	chainsMu     sync.RWMutex
}

// NewOrchestratorCOT creates a new COT orchestrator
func NewOrchestratorCOT(pool *pgxpool.Pool, cfg *config.Config, registry *AgentRegistryV2) *OrchestratorCOT {
	return &OrchestratorCOT{
		pool:         pool,
		config:       cfg,
		registry:     registry,
		router:       NewSmartIntentRouter(pool, cfg),
		messageBus:   make(chan *AgentMessage, 100),
		activeChains: make(map[string]*ChainOfThought),
	}
}

// ExecutionPlan represents the orchestrator's plan for handling a request
type ExecutionPlan struct {
	Strategy     string        `json:"strategy"` // "direct", "delegate", "multi-agent", "sequential"
	PrimaryAgent AgentTypeV2   `json:"primary_agent"`
	Steps        []PlannedStep `json:"steps"`
	Reasoning    string        `json:"reasoning"`
	Confidence   float64       `json:"confidence"`
}

// PlannedStep represents a planned execution step
type PlannedStep struct {
	Order       int         `json:"order"`
	Agent       AgentTypeV2 `json:"agent"`
	Task        string      `json:"task"`
	DependsOn   []int       `json:"depends_on,omitempty"` // Order numbers of dependencies
	CanParallel bool        `json:"can_parallel"`
	Priority    int         `json:"priority"`
}

// ProcessWithCOT processes a user request with full chain of thought tracking
func (o *OrchestratorCOT) ProcessWithCOT(
	ctx context.Context,
	input AgentInput,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	llmOptions services.LLMOptions,
) (<-chan streaming.StreamEvent, <-chan error, *ChainOfThought) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	// Get user message
	var userMessage string
	for i := len(input.Messages) - 1; i >= 0; i-- {
		if strings.ToLower(input.Messages[i].Role) == "user" {
			userMessage = input.Messages[i].Content
			break
		}
	}

	// Create chain of thought
	cot := NewChainOfThought(userMessage)
	o.chainsMu.Lock()
	o.activeChains[cot.ID] = cot
	o.chainsMu.Unlock()

	go func() {
		defer close(events)
		defer close(errs)
		startTime := time.Now()

		// Step 1: Analysis - Understand the request
		analysisStep := &ThoughtStep{
			Agent:     AgentTypeV2Orchestrator,
			Action:    "analyze",
			Input:     userMessage,
			Reasoning: "Analyzing user request to determine intent and required agents",
		}
		cot.AddStep(analysisStep)
		analysisStep.Status = "running"

		// Send thinking event
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: "🧠 *Analyzing request...*\n\n",
		}

		// Classify intent using SmartIntentRouter
		intent := o.router.ClassifyIntent(ctx, input.Messages, input.Context)
		analysisStep.Output = fmt.Sprintf("Intent: %s (confidence: %.2f)", intent.TargetAgent, intent.Confidence)
		analysisStep.Confidence = intent.Confidence
		cot.UpdateStep(analysisStep.ID, analysisStep.Output, "completed")

		// Step 2: Planning - Decide execution strategy
		planStep := &ThoughtStep{
			Agent:     AgentTypeV2Orchestrator,
			Action:    "plan",
			Input:     analysisStep.Output,
			Reasoning: "Creating execution plan based on intent analysis",
		}
		cot.AddStep(planStep)
		planStep.Status = "running"

		plan := o.createExecutionPlan(ctx, userMessage, intent, input)
		planStep.Output = fmt.Sprintf("Strategy: %s, Primary: %s, Steps: %d",
			plan.Strategy, plan.PrimaryAgent, len(plan.Steps))
		planStep.Confidence = plan.Confidence
		cot.UpdateStep(planStep.ID, planStep.Output, "completed")

		// Send plan event
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: fmt.Sprintf("📋 *Plan: %s*\n\n", plan.Strategy),
		}

		cot.Status = "executing"

		// Step 3: Execute based on strategy
		switch plan.Strategy {
		case "direct":
			// Orchestrator handles directly
			o.executeDirectly(ctx, cot, input, llmOptions, events, errs)

		case "delegate":
			// Single agent delegation
			o.executeDelegation(ctx, cot, plan, input, userID, userName, conversationID, llmOptions, events, errs)

		case "multi-agent":
			// Multiple agents in parallel
			o.executeMultiAgent(ctx, cot, plan, input, userID, userName, conversationID, llmOptions, events, errs)

		case "sequential":
			// Agents in sequence
			o.executeSequential(ctx, cot, plan, input, userID, userName, conversationID, llmOptions, events, errs)
		}

		// Step 4: Synthesis (if multiple agents were involved)
		if len(cot.GetAgentsInvolved()) > 1 {
			cot.Status = "synthesizing"
			o.synthesizeResults(ctx, cot, events)
		}

		// Complete
		cot.Status = "completed"
		cot.TotalDuration = time.Since(startTime)
		cot.AgentsInvolved = cot.GetAgentsInvolved()

		// Send completion event with COT summary
		cotSummary := o.formatCOTSummary(cot)
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: cotSummary,
		}
	}()

	return events, errs, cot
}

// createExecutionPlan determines how to handle the request
func (o *OrchestratorCOT) createExecutionPlan(
	_ context.Context,
	userMessage string,
	intent Intent,
	_ AgentInput,
) *ExecutionPlan {
	plan := &ExecutionPlan{
		PrimaryAgent: intent.TargetAgent,
		Confidence:   intent.Confidence,
		Steps:        make([]PlannedStep, 0),
	}

	// Determine strategy based on intent and message complexity
	msgLower := strings.ToLower(userMessage)

	// Check for multi-agent indicators
	hasMultipleIntents := strings.Contains(msgLower, " and ") ||
		strings.Contains(msgLower, " then ") ||
		strings.Contains(msgLower, " also ")

	// Check for specific patterns
	needsAnalysis := strings.Contains(msgLower, "analyze") || strings.Contains(msgLower, "metrics")
	needsDocument := strings.Contains(msgLower, "document") || strings.Contains(msgLower, "report") || strings.Contains(msgLower, "proposal")
	needsProject := strings.Contains(msgLower, "project") || strings.Contains(msgLower, "plan")
	needsTask := strings.Contains(msgLower, "task") || strings.Contains(msgLower, "todo") || strings.Contains(msgLower, "prioritize")
	needsClient := strings.Contains(msgLower, "client") || strings.Contains(msgLower, "pipeline")

	// Count how many domains are involved
	domainCount := 0
	if needsAnalysis {
		domainCount++
	}
	if needsDocument {
		domainCount++
	}
	if needsProject {
		domainCount++
	}
	if needsTask {
		domainCount++
	}
	if needsClient {
		domainCount++
	}

	// Decide strategy
	if !intent.ShouldDelegate || intent.Confidence < 0.5 {
		plan.Strategy = "direct"
		plan.Reasoning = "Low confidence or general request - handling directly"
		plan.Steps = append(plan.Steps, PlannedStep{
			Order: 1,
			Agent: AgentTypeV2Orchestrator,
			Task:  userMessage,
		})
	} else if domainCount > 1 || hasMultipleIntents {
		// Multiple domains - need multi-agent or sequential
		if strings.Contains(msgLower, " then ") {
			plan.Strategy = "sequential"
			plan.Reasoning = "Sequential tasks detected - executing in order"
		} else {
			plan.Strategy = "multi-agent"
			plan.Reasoning = "Multiple domains detected - parallel execution"
		}

		// Add steps for each domain
		order := 1
		if needsAnalysis {
			plan.Steps = append(plan.Steps, PlannedStep{Order: order, Agent: AgentTypeV2Analyst, Task: "Analyze relevant data", CanParallel: true})
			order++
		}
		if needsProject {
			plan.Steps = append(plan.Steps, PlannedStep{Order: order, Agent: AgentTypeV2Project, Task: "Handle project aspects", CanParallel: true})
			order++
		}
		if needsTask {
			plan.Steps = append(plan.Steps, PlannedStep{Order: order, Agent: AgentTypeV2Task, Task: "Handle task management", CanParallel: true})
			order++
		}
		if needsClient {
			plan.Steps = append(plan.Steps, PlannedStep{Order: order, Agent: AgentTypeV2Client, Task: "Handle client aspects", CanParallel: true})
			order++
		}
		if needsDocument {
			// Documents usually come last (need other info first)
			plan.Steps = append(plan.Steps, PlannedStep{Order: order, Agent: AgentTypeV2Document, Task: "Create document", CanParallel: false, DependsOn: []int{1}})
		}
	} else {
		// Single domain - delegate
		plan.Strategy = "delegate"
		plan.Reasoning = fmt.Sprintf("Clear intent for %s agent", intent.TargetAgent)
		plan.Steps = append(plan.Steps, PlannedStep{
			Order: 1,
			Agent: intent.TargetAgent,
			Task:  userMessage,
		})
	}

	return plan
}

// executeDirectly handles the request directly via orchestrator
func (o *OrchestratorCOT) executeDirectly(
	ctx context.Context,
	cot *ChainOfThought,
	input AgentInput,
	llmOptions services.LLMOptions,
	events chan<- streaming.StreamEvent,
	errs chan<- error,
) {
	step := &ThoughtStep{
		Agent:     AgentTypeV2Orchestrator,
		Action:    "execute",
		Input:     cot.UserMessage,
		Reasoning: "Handling directly as general request",
	}
	cot.AddStep(step)
	step.Status = "running"

	// Get orchestrator agent and set LLM options
	agent := o.registry.GetAgent(AgentTypeV2Orchestrator, input.UserID, input.UserName, &input.ConversationID, input.Context)
	agent.SetOptions(llmOptions)
	agentEvents, agentErrs := agent.Run(ctx, input)

	var output strings.Builder
	for {
		select {
		case event, ok := <-agentEvents:
			if !ok {
				step.Output = output.String()
				cot.UpdateStep(step.ID, step.Output, "completed")
				cot.FinalOutput = output.String()
				return
			}
			events <- event
			if event.Type == streaming.EventTypeToken {
				if content, ok := event.Data.(string); ok {
					output.WriteString(content)
				}
			}
		case err := <-agentErrs:
			if err != nil {
				step.Error = err.Error()
				cot.UpdateStep(step.ID, "", "failed")
				errs <- err
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

// executeDelegation delegates to a single specialist agent
func (o *OrchestratorCOT) executeDelegation(
	ctx context.Context,
	cot *ChainOfThought,
	plan *ExecutionPlan,
	input AgentInput,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	llmOptions services.LLMOptions,
	events chan<- streaming.StreamEvent,
	errs chan<- error,
) {
	if len(plan.Steps) == 0 {
		return
	}

	targetAgent := plan.Steps[0].Agent

	// Create delegation message
	delegateMsg := NewAgentMessage(AgentTypeV2Orchestrator, targetAgent, MsgTypeDelegate, cot.UserMessage)

	// Log the delegation step
	delegateStep := &ThoughtStep{
		Agent:     AgentTypeV2Orchestrator,
		Action:    "delegate",
		Input:     cot.UserMessage,
		Output:    fmt.Sprintf("Delegating to %s agent", targetAgent),
		Reasoning: plan.Reasoning,
	}
	cot.AddStep(delegateStep)
	cot.UpdateStep(delegateStep.ID, delegateStep.Output, "completed")

	// Execute step
	execStep := &ThoughtStep{
		Agent:     targetAgent,
		Action:    "execute",
		Input:     delegateMsg.Content,
		Reasoning: "Executing delegated task from orchestrator",
	}
	cot.AddStep(execStep)
	execStep.Status = "running"

	// Send agent indicator
	events <- streaming.StreamEvent{
		Type: streaming.EventTypeToken,
		Data: fmt.Sprintf("🤖 **%s Agent**\n\n", targetAgent),
	}

	// Get and run the agent with LLM options
	agent := o.registry.GetAgent(targetAgent, userID, userName, conversationID, input.Context)
	agent.SetOptions(llmOptions)
	agentEvents, agentErrs := agent.Run(ctx, input)

	var output strings.Builder
	for {
		select {
		case event, ok := <-agentEvents:
			if !ok {
				execStep.Output = output.String()
				cot.UpdateStep(execStep.ID, execStep.Output, "completed")
				cot.FinalOutput = output.String()
				return
			}
			events <- event
			if event.Type == streaming.EventTypeToken {
				if content, ok := event.Data.(string); ok {
					output.WriteString(content)
				}
			}
		case err := <-agentErrs:
			if err != nil {
				execStep.Error = err.Error()
				cot.UpdateStep(execStep.ID, "", "failed")
				errs <- err
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

// executeMultiAgent runs multiple agents in parallel
func (o *OrchestratorCOT) executeMultiAgent(
	ctx context.Context,
	cot *ChainOfThought,
	plan *ExecutionPlan,
	input AgentInput,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	llmOptions services.LLMOptions,
	events chan<- streaming.StreamEvent,
	_ chan<- error, // errs - reserved for future error handling
) {
	events <- streaming.StreamEvent{
		Type: streaming.EventTypeToken,
		Data: fmt.Sprintf("🔄 **Multi-Agent Execution** (%d agents)\n\n", len(plan.Steps)),
	}

	var wg sync.WaitGroup
	results := make(chan AgentResult, len(plan.Steps))

	// Launch parallel agents
	for _, step := range plan.Steps {
		if !step.CanParallel {
			continue // Handle non-parallel steps later
		}

		wg.Add(1)
		go func(s PlannedStep) {
			defer wg.Done()

			// Create execution step in COT
			execStep := &ThoughtStep{
				Agent:     s.Agent,
				Action:    "execute",
				Input:     s.Task,
				Reasoning: "Parallel execution as part of multi-agent plan",
			}
			cot.AddStep(execStep)
			execStep.Status = "running"

			// Run agent with LLM options
			agent := o.registry.GetAgent(s.Agent, userID, userName, conversationID, input.Context)
			agent.SetOptions(llmOptions)
			agentEvents, agentErrs := agent.Run(ctx, input)

			var output strings.Builder
			for {
				select {
				case event, ok := <-agentEvents:
					if !ok {
						execStep.Output = output.String()
						cot.UpdateStep(execStep.ID, execStep.Output, "completed")
						results <- AgentResult{AgentType: s.Agent, Output: output.String()}
						return
					}
					if event.Type == streaming.EventTypeToken {
						if content, ok := event.Data.(string); ok {
							output.WriteString(content)
						}
					}
				case err := <-agentErrs:
					if err != nil {
						execStep.Error = err.Error()
						cot.UpdateStep(execStep.ID, "", "failed")
						results <- AgentResult{AgentType: s.Agent, Error: err}
					}
					return
				case <-ctx.Done():
					return
				}
			}
		}(step)
	}

	// Wait for all parallel agents
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allOutputs strings.Builder
	for result := range results {
		if result.Error != nil {
			continue
		}
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: fmt.Sprintf("\n---\n### %s Agent:\n%s\n", result.AgentType, truncateOutput(result.Output, 500)),
		}
		allOutputs.WriteString(result.Output)
		allOutputs.WriteString("\n\n")
	}

	cot.FinalOutput = allOutputs.String()
}

// executeSequential runs agents in sequence
func (o *OrchestratorCOT) executeSequential(
	ctx context.Context,
	cot *ChainOfThought,
	plan *ExecutionPlan,
	input AgentInput,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	llmOptions services.LLMOptions,
	events chan<- streaming.StreamEvent,
	errs chan<- error,
) {
	events <- streaming.StreamEvent{
		Type: streaming.EventTypeToken,
		Data: fmt.Sprintf("📝 **Sequential Execution** (%d steps)\n\n", len(plan.Steps)),
	}

	var previousOutput string

	for i, step := range plan.Steps {
		// Create execution step
		execStep := &ThoughtStep{
			Agent:     step.Agent,
			Action:    "execute",
			Input:     step.Task,
			Reasoning: fmt.Sprintf("Step %d of sequential execution", i+1),
		}
		cot.AddStep(execStep)
		execStep.Status = "running"

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: fmt.Sprintf("\n**Step %d: %s Agent**\n", i+1, step.Agent),
		}

		// Modify input to include previous output
		modifiedInput := input
		if previousOutput != "" {
			contextMsg := services.ChatMessage{
				Role:    "system",
				Content: "Previous step output:\n" + previousOutput,
			}
			modifiedInput.Messages = append([]services.ChatMessage{contextMsg}, modifiedInput.Messages...)
		}

		// Run agent with LLM options
		agent := o.registry.GetAgent(step.Agent, userID, userName, conversationID, input.Context)
		agent.SetOptions(llmOptions)
		agentEvents, agentErrs := agent.Run(ctx, modifiedInput)

		var output strings.Builder
		done := false
		for !done {
			select {
			case event, ok := <-agentEvents:
				if !ok {
					done = true
					break
				}
				events <- event
				if event.Type == streaming.EventTypeToken {
					if content, ok := event.Data.(string); ok {
						output.WriteString(content)
					}
				}
			case err := <-agentErrs:
				if err != nil {
					execStep.Error = err.Error()
					cot.UpdateStep(execStep.ID, "", "failed")
					errs <- err
					return
				}
				done = true
			case <-ctx.Done():
				return
			}
		}

		previousOutput = output.String()
		execStep.Output = previousOutput
		cot.UpdateStep(execStep.ID, execStep.Output, "completed")
	}

	cot.FinalOutput = previousOutput
}

// synthesizeResults combines outputs from multiple agents
func (o *OrchestratorCOT) synthesizeResults(
	_ context.Context,
	cot *ChainOfThought,
	_ chan<- streaming.StreamEvent,
) {
	synthStep := &ThoughtStep{
		Agent:     AgentTypeV2Orchestrator,
		Action:    "synthesize",
		Input:     "Multiple agent outputs",
		Reasoning: "Combining outputs from multiple specialists",
	}
	cot.AddStep(synthStep)
	synthStep.Status = "running"

	// Collect all outputs
	var outputs []string
	for _, step := range cot.Steps {
		if step.Action == "execute" && step.Status == "completed" && step.Output != "" {
			outputs = append(outputs, fmt.Sprintf("**%s**: %s", step.Agent, truncateOutput(step.Output, 300)))
		}
	}

	synthStep.Output = fmt.Sprintf("Synthesized %d agent outputs", len(outputs))
	cot.UpdateStep(synthStep.ID, synthStep.Output, "completed")
}

// formatCOTSummary creates a readable summary of the chain of thought
func (o *OrchestratorCOT) formatCOTSummary(cot *ChainOfThought) string {
	var sb strings.Builder
	sb.WriteString("\n\n---\n")
	sb.WriteString("### 🧠 Chain of Thought Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", cot.TotalDuration.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("**Agents:** %v\n", cot.AgentsInvolved))
	sb.WriteString(fmt.Sprintf("**Steps:** %d\n", len(cot.Steps)))
	return sb.String()
}

// AgentResult holds result from an agent execution
type AgentResult struct {
	AgentType AgentTypeV2
	Output    string
	Error     error
}

// truncateOutput truncates output for display
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetCOTAsJSON returns the chain of thought as JSON for debugging/logging
func (cot *ChainOfThought) GetCOTAsJSON() (string, error) {
	cot.mu.RLock()
	defer cot.mu.RUnlock()
	data, err := json.MarshalIndent(cot, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
