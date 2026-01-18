# BusinessOS Agent System Architecture

> **Version:** V2 Architecture (Current)
> **Status:** Production
> **Last Updated:** January 2026

---

## Table of Contents

1. [Overview](#overview)
2. [Agent V2 Architecture](#agent-v2-architecture)
3. [Agent Types](#agent-types)
4. [Agent Lifecycle](#agent-lifecycle)
5. [Tool Integration](#tool-integration)
6. [Streaming Responses (SSE)](#streaming-responses-sse)
7. [Context Management](#context-management)
8. [Agent Registry & Dispatch](#agent-registry--dispatch)
9. [Intent Routing System](#intent-routing-system)
10. [Chain of Thought (COT) Orchestration](#chain-of-thought-cot-orchestration)
11. [Creating Custom Agents](#creating-custom-agents)
12. [Best Practices](#best-practices)
13. [Recent Changes & Improvements](#recent-changes--improvements)

---

## Overview

The BusinessOS Agent System is a modular, intelligent AI agent architecture that powers the core functionality of the platform. It uses specialized agents for different business domains, with an orchestrator that routes requests and coordinates multi-agent workflows.

### Key Features

- **Specialized Agents**: Domain-specific agents for documents, projects, tasks, clients, and analysis
- **Intelligent Routing**: Multi-layer intent classification with regex patterns, semantic signals, and LLM fallback
- **Tool Calling**: Agents can execute business logic through a structured tool system
- **Streaming Responses**: Real-time SSE (Server-Sent Events) for smooth UX
- **Chain of Thought**: Advanced orchestration for multi-agent workflows
- **Context Awareness**: Tiered context system with role-based permissions and memory hierarchy

---

## Agent V2 Architecture

### Core Interface

All agents implement the `AgentV2` interface:

```go
type AgentV2 interface {
    // Identity
    Type() AgentTypeV2
    Name() string
    Description() string

    // Configuration
    GetSystemPrompt() string
    GetContextRequirements() ContextRequirements

    // Execution - returns streaming events
    Run(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error)
    RunWithTools(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error)

    // Options
    SetModel(model string)
    SetOptions(opts services.LLMOptions)
    SetCustomSystemPrompt(prompt string)
    SetFocusModePrompt(prompt string)
    SetOutputStylePrompt(prompt string)
    SetRoleContextPrompt(prompt string)    // Feature 1: Role-based permissions
    SetMemoryContext(context string)       // Memory Hierarchy: Workspace memory
    SetSkillsPrompt(prompt string)         // Agent Skills System
}
```

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER REQUEST                            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR AGENT                           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  SmartIntentRouter (Multi-Layer Classification)         │   │
│  │  • Layer 1: Regex Patterns (fast, high precision)       │   │
│  │  • Layer 2: Semantic Signals (nuanced detection)        │   │
│  │  • Layer 3: Context Boosting (user selections, history) │   │
│  │  • Layer 4: LLM Classifier (ambiguous cases)            │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
              ┌──────────────┴───────────────┐
              │    Agent Dispatch Strategy    │
              └──────────────┬───────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   ┌─────────┐        ┌─────────┐         ┌─────────┐
   │ Direct  │        │Delegate │         │Multi-   │
   │(Orch.)  │        │(Single  │         │Agent/   │
   │         │        │Agent)   │         │Sequence │
   └─────────┘        └────┬────┘         └────┬────┘
                           │                   │
        ┌──────────────────┴───────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                      SPECIALIST AGENTS                          │
├─────────────────────────────────────────────────────────────────┤
│  📄 Document  │  📊 Project  │  ✅ Task  │  👥 Client  │  📈 Analyst │
│  Specialist   │  Specialist  │Specialist │ Specialist  │ Specialist  │
├─────────────────────────────────────────────────────────────────┤
│                         Tool Registry                           │
│  • Read Tools: get_*, list_*, search_*, query_*                │
│  • Write Tools: create_*, update_*, log_*                      │
│  • Context Tools: tree_search, browse_tree, load_context       │
│  • Search Tools: web_search                                    │
└─────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                    STREAMING RESPONSE (SSE)                     │
│  • EventTypeToken: Content chunks                              │
│  • EventTypeThinking: Internal reasoning                       │
│  • EventTypeArtifact: Structured outputs (JSON/XML)            │
│  • EventTypeToolCall: Tool execution logs                      │
│  • EventTypeDone: Completion signal                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Agent Types

### 1. Orchestrator Agent

**Type:** `AgentTypeV2Orchestrator`
**Role:** Primary interface, routes requests, coordinates multi-agent workflows

**Enabled Tools:**
- `search_documents`, `get_project`, `get_task`, `get_client`
- `create_task`, `create_project`, `create_client`
- `create_artifact`, `log_activity`
- `tree_search`, `browse_tree`, `load_context` (knowledge base)

**Context Requirements:**
```go
ContextRequirements{
    NeedsProjects:    true,
    NeedsTasks:       true,
    NeedsClients:     true,
    NeedsKnowledge:   true,
    MaxContextTokens: 10000,
}
```

**When to Use:**
- General queries
- Low-confidence intent classification
- Multi-domain requests requiring coordination

---

### 2. Document Agent

**Type:** `AgentTypeV2Document`
**Role:** Creates formal business documents (proposals, SOPs, reports, frameworks)

**Enabled Tools:**
- `create_artifact`, `search_documents`
- `get_project`, `get_client`
- `log_activity`
- `tree_search`, `browse_tree`, `load_context`

**Context Requirements:**
```go
ContextRequirements{
    NeedsProjects:    true,
    NeedsKnowledge:   true,
    NeedsClients:     true,
    MaxContextTokens: 10000,
    PrioritySections: []string{"project_details", "selected_documents", "client_info"},
}
```

**Activation Patterns:**
- "Create a proposal for..."
- "Write a formal SOP for..."
- "Draft a report on..."
- "Generate a business framework for..."

---

### 3. Project Agent

**Type:** `AgentTypeV2Project`
**Role:** Project management, planning, task coordination

**Enabled Tools:**
- `create_project`, `update_project`, `get_project`, `list_projects`
- `create_task`, `bulk_create_tasks`, `assign_task`
- `get_team_capacity`, `search_documents`
- `create_artifact`, `log_activity`
- `tree_search`, `browse_tree`, `load_context`

**Context Requirements:**
```go
ContextRequirements{
    NeedsProjects: true,
    NeedsTasks:    true,
    NeedsTeam:     true,
    NeedsClients:  true,
    MaxContextTokens: 8000,
}
```

**Activation Patterns:**
- "Create a new project for..."
- "Plan my sprint tasks"
- "What should I work on next?"
- "Break down this task into subtasks"

---

### 4. Task Agent

**Type:** `AgentTypeV2Task`
**Role:** Task management, prioritization, scheduling, dependencies

**Enabled Tools:**
- `create_task`, `update_task`, `get_task`, `list_tasks`
- `bulk_create_tasks`, `move_task`, `assign_task`
- `get_team_capacity`, `get_project`
- `log_activity`
- `tree_search`, `browse_tree`, `load_context`

**Context Requirements:**
```go
ContextRequirements{
    NeedsProjects: true,
    NeedsTasks:    true,
    NeedsTeam:     true,
    MaxContextTokens: 8000,
}
```

**Activation Patterns:**
- "Prioritize my tasks"
- "Schedule this task for tomorrow"
- "Mark task as done"
- "Assign this to John"

---

### 5. Client Agent

**Type:** `AgentTypeV2Client`
**Role:** Client relationship management, CRM, pipeline tracking

**Enabled Tools:**
- `create_client`, `update_client`, `get_client`
- `log_client_interaction`, `update_client_pipeline`
- `search_documents`, `get_project`
- `create_artifact`, `log_activity`
- `tree_search`, `browse_tree`, `load_context`

**Context Requirements:**
```go
ContextRequirements{
    NeedsClients:   true,
    NeedsProjects:  true,
    NeedsKnowledge: true,
    MaxContextTokens: 6000,
}
```

**Activation Patterns:**
- "Add a new client..."
- "Move client to pipeline stage..."
- "Log a meeting with client..."
- "Follow up with prospect..."

---

### 6. Analyst Agent

**Type:** `AgentTypeV2Analyst`
**Role:** Data analysis, metrics, insights, research

**Enabled Tools:**
- `query_metrics`, `get_team_capacity`
- `list_projects`, `list_tasks`, `get_project`
- `search_documents`, `create_artifact`
- `log_activity`
- `tree_search`, `browse_tree`, `load_context`

**Context Requirements:**
```go
ContextRequirements{
    NeedsProjects: true,
    NeedsTasks:    true,
    NeedsClients:  true,
    NeedsTeam:     true,
    MaxContextTokens: 8000,
}
```

**Activation Patterns:**
- "Analyze our project performance"
- "How are we doing on revenue?"
- "Compare Q3 vs Q4 metrics"
- "What's working and what's not?"
- Research queries: "How does X work?", "Explain Y"

---

## Agent Lifecycle

### 1. Initialization

```go
// Create agent registry
registry := agents.NewAgentRegistryV2(
    pool,
    cfg,
    embeddingService,
    promptPersonalizer,
)

// Get a specialized agent
agent := registry.GetAgent(
    agents.AgentTypeV2Document,
    userID,
    userName,
    &conversationID,
    tieredContext,
)
```

### 2. Configuration

```go
// Set LLM options
agent.SetOptions(services.LLMOptions{
    Temperature:       0.7,
    MaxTokens:         4000,
    ThinkingEnabled:   true,
    TopP:              0.9,
})

// Inject context modifiers
agent.SetFocusModePrompt("Focus on creating formal business documents...")
agent.SetRoleContextPrompt(roleContext)     // Role-based permissions
agent.SetMemoryContext(memoryContext)       // Workspace memory
```

### 3. Execution

```go
// Prepare input
input := agents.AgentInput{
    Messages:       chatMessages,
    Context:        tieredContext,
    Selections:     userSelections,
    FocusMode:      "write",
    ConversationID: conversationID,
    UserID:         userID,
    UserName:       userName,
    RoleContext:    roleContext,
    MemoryContext:  memoryContext,
}

// Run agent (streaming)
events, errs := agent.Run(ctx, input)

// Process streaming events
for event := range events {
    switch event.Type {
    case streaming.EventTypeToken:
        // Send content to client
    case streaming.EventTypeThinking:
        // Show thinking process
    case streaming.EventTypeArtifact:
        // Handle structured output
    case streaming.EventTypeDone:
        // Completion
    }
}
```

---

## Tool Integration

### Tool Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    AgentToolRegistry                            │
├─────────────────────────────────────────────────────────────────┤
│  registerTools()                                                │
│  ├─ Read Tools:   get_*, list_*, search_*, query_*             │
│  ├─ Write Tools:  create_*, update_*, log_*                    │
│  ├─ Context Tools: tree_search, browse_tree, load_context      │
│  └─ Search Tools:  web_search                                  │
├─────────────────────────────────────────────────────────────────┤
│  GetTool(name) → AgentTool                                      │
│  ExecuteTool(ctx, name, input) → (result, error)               │
│  GetToolDefinitions() → []ToolDefinition                        │
└─────────────────────────────────────────────────────────────────┘
```

### Tool Interface

Every tool implements:

```go
type AgentTool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```

### Example Tool Implementation

```go
type CreateTaskTool struct {
    pool   *pgxpool.Pool
    userID string
}

func (t *CreateTaskTool) Name() string {
    return "create_task"
}

func (t *CreateTaskTool) Description() string {
    return "Create a new task with title, description, and optional project assignment"
}

func (t *CreateTaskTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "title": map[string]interface{}{
                "type":        "string",
                "description": "Task title",
            },
            "description": map[string]interface{}{
                "type":        "string",
                "description": "Task description",
            },
            "project_id": map[string]interface{}{
                "type":        "string",
                "description": "Project UUID (optional)",
            },
            "priority": map[string]interface{}{
                "type": "string",
                "enum": []string{"low", "medium", "high", "critical"},
            },
        },
        "required": []string{"title"},
    }
}

func (t *CreateTaskTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
    var params struct {
        Title       string `json:"title"`
        Description string `json:"description"`
        ProjectID   string `json:"project_id"`
        Priority    string `json:"priority"`
    }
    if err := json.Unmarshal(input, &params); err != nil {
        return "", err
    }

    // Execute business logic
    taskID, err := createTaskInDB(ctx, t.pool, t.userID, params)
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("Task created successfully with ID: %s", taskID), nil
}
```

### Tool Categories

**Read Tools** (Query data without side effects):
- `get_project`, `get_task`, `get_client`
- `list_tasks`, `list_projects`
- `search_documents`, `query_metrics`
- `get_team_capacity`
- `tree_search`, `browse_tree`, `load_context`

**Write Tools** (Modify data):
- `create_task`, `update_task`, `assign_task`, `move_task`
- `create_project`, `update_project`
- `create_client`, `update_client`
- `log_client_interaction`, `update_client_pipeline`
- `create_artifact`, `log_activity`
- `bulk_create_tasks`

**Search Tools**:
- `web_search` (external knowledge)

---

## Streaming Responses (SSE)

### Event Types

```go
const (
    EventTypeToken     = "token"      // Content chunk
    EventTypeThinking  = "thinking"   // Internal reasoning
    EventTypeArtifact  = "artifact"   // Structured output
    EventTypeToolCall  = "tool_call"  // Tool execution
    EventTypeDone      = "done"       // Completion
    EventTypeError     = "error"      // Error occurred
)
```

### Streaming Flow

```go
// Agent emits streaming events
events := make(chan streaming.StreamEvent, 100)

// Event producer (agent)
events <- streaming.StreamEvent{
    Type: streaming.EventTypeThinking,
    Data: "Analyzing project requirements...",
}

events <- streaming.StreamEvent{
    Type: streaming.EventTypeToken,
    Data: "Based on the project scope, ",
}

events <- streaming.StreamEvent{
    Type: streaming.EventTypeArtifact,
    Data: map[string]interface{}{
        "type": "document",
        "title": "Project Proposal",
        "content": "...",
    },
}

events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
close(events)
```

### Artifact Detection

The system automatically detects structured outputs (XML/JSON artifacts) in the response stream:

```go
detector := streaming.NewArtifactDetector()

for chunk := range llmChunks {
    events := detector.ProcessChunk(chunk)
    for _, event := range events {
        // Automatically emits:
        // - EventTypeToken for regular content
        // - EventTypeArtifact for detected XML/JSON blocks
    }
}
```

---

## Context Management

### Tiered Context System

BusinessOS uses a three-tier context system to provide agents with relevant information:

```go
type TieredContext struct {
    Level1 Tier1Context  // Current workspace context
    Level2 Tier2Context  // Recent history
    Level3 Tier3Context  // Broader workspace context
}
```

**Level 1: Current Context** (Highest Priority)
- Selected project
- Selected documents
- Current node
- Linked client
- Active focus mode

**Level 2: Recent History**
- Recent tasks
- Recent projects
- Recent client interactions

**Level 3: Workspace Context**
- User profile
- Team information
- Workspace settings

### Context Requirements

Each agent declares what context it needs:

```go
type ContextRequirements struct {
    NeedsProjects    bool
    NeedsTasks       bool
    NeedsClients     bool
    NeedsTeam        bool
    NeedsKnowledge   bool
    NeedsMetrics     bool
    NeedsFullHistory bool
    MaxContextTokens int
    PrioritySections []string
}
```

### Context Injection Flow

```
User Request
    │
    ▼
Context Builder
    ├─ Extract user selections
    ├─ Query Level 1 (selected entities)
    ├─ Query Level 2 (recent history)
    ├─ Query Level 3 (workspace data)
    └─ Format for agent
    │
    ▼
Agent
    ├─ Filter by ContextRequirements
    ├─ Apply token budget (MaxContextTokens)
    ├─ Prioritize sections (PrioritySections)
    └─ Build final prompt
```

---

## Agent Registry & Dispatch

### AgentRegistryV2

Manages agent lifecycle and creation:

```go
type AgentRegistryV2 struct {
    pool               *pgxpool.Pool
    config             *config.Config
    embeddingService   *services.EmbeddingService
    promptPersonalizer *services.PromptPersonalizer
}
```

**Methods:**
- `GetAgent(agentType, userID, userName, conversationID, tieredContext) → AgentV2`
- `GetAgentForFocusModeV2(focusMode) → AgentTypeV2`

### Focus Mode Mapping

```go
func GetAgentForFocusModeV2(focusMode string) AgentTypeV2 {
    switch focusMode {
    case "write":
        return AgentTypeV2Document
    case "analyze", "research":
        return AgentTypeV2Analyst
    case "plan", "build":
        return AgentTypeV2Project
    default:
        return AgentTypeV2Orchestrator
    }
}
```

---

## Intent Routing System

### SmartIntentRouter

Multi-layer classification for intelligent routing:

```
┌─────────────────────────────────────────────────────────────────┐
│                   SmartIntentRouter                             │
├─────────────────────────────────────────────────────────────────┤
│  Layer 1: Pattern Matching (Fast, High Precision)              │
│  ├─ Regex patterns with weights                                │
│  ├─ MustMatch patterns (100% confidence)                       │
│  └─ Example: "create a proposal" → Document Agent (1.0)        │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2: Semantic Signals (Nuanced Detection)                 │
│  ├─ Keyword indicators with weights                            │
│  ├─ Category classification                                    │
│  └─ Example: "formal", "deadline" → boost scores               │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3: Context Boosting (User State Awareness)              │
│  ├─ Selected project → boost Project Agent                     │
│  ├─ Selected client → boost Client Agent                       │
│  └─ Conversation history analysis                              │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4: LLM Classifier (Ambiguous Cases)                     │
│  ├─ Only if confidence < 0.7                                   │
│  ├─ Fast model with 5s timeout                                 │
│  └─ JSON response with agent, confidence, reasoning            │
└─────────────────────────────────────────────────────────────────┘
```

### Classification Process

```go
intent := router.ClassifyIntent(ctx, messages, tieredContext)

type Intent struct {
    Category       string      // "document", "project", "client", etc.
    ShouldDelegate bool        // true if routing to specialist
    TargetAgent    AgentTypeV2 // Target agent type
    Confidence     float64     // 0.0 - 1.0
    Reasoning      string      // Why this agent was selected
}
```

### Pattern Examples

**Document Agent Patterns:**
```go
{Pattern: regexp.MustCompile(`(?i)(create|write|draft|generate)\s+(a\s+)?(formal\s+)?(proposal|sop|report|framework)`),
 Weight: 1.0,
 MustMatch: true}
```

**Project Agent Patterns:**
```go
{Pattern: regexp.MustCompile(`(?i)(what('s|s)?|which)\s+(should\s+i|to)\s+(work\s+on|do\s+next)`),
 Weight: 0.9,
 MustMatch: true}
```

**Client Agent Patterns:**
```go
{Pattern: regexp.MustCompile(`(?i)(move|update|change)\s+(client|lead|prospect)\s+(to|in|through)\s+pipeline`),
 Weight: 1.0,
 MustMatch: true}
```

### Semantic Signals

```go
// Document signals
{Indicator: "formal", Weight: 0.3, Category: "tone"}
{Indicator: "deliverable", Weight: 0.4, Category: "output"}

// Project signals
{Indicator: "deadline", Weight: 0.4, Category: "time"}
{Indicator: "priority", Weight: 0.4, Category: "organization"}

// Client signals
{Indicator: "pipeline", Weight: 0.5, Category: "crm"}
{Indicator: "relationship", Weight: 0.4, Category: "crm"}
```

---

## Chain of Thought (COT) Orchestration

### Overview

Advanced orchestration system that tracks multi-agent reasoning chains and enables complex workflows.

### COT Execution Strategies

```
┌─────────────────────────────────────────────────────────────────┐
│                      COT Strategies                             │
├─────────────────────────────────────────────────────────────────┤
│  1. DIRECT                                                      │
│     Orchestrator handles → No delegation                        │
│     Use: General queries, low confidence                        │
│                                                                 │
│  2. DELEGATE                                                    │
│     Orchestrator → Single Specialist Agent                      │
│     Use: Clear single-domain requests                           │
│                                                                 │
│  3. MULTI-AGENT (Parallel)                                      │
│     Orchestrator → Multiple Agents (concurrent)                 │
│     Use: Tasks spanning multiple domains                        │
│                                                                 │
│  4. SEQUENTIAL (Multi-Hop)                                      │
│     Agent A → Agent B → Agent C → Synthesis                     │
│     Use: "then", "after", "next" keywords                       │
└─────────────────────────────────────────────────────────────────┘
```

### COT Data Structures

```go
type ChainOfThought struct {
    ID             string
    UserMessage    string
    Steps          []*ThoughtStep
    FinalOutput    string
    TotalDuration  time.Duration
    AgentsInvolved []AgentTypeV2
    Status         string  // "planning", "executing", "synthesizing", "completed"
}

type ThoughtStep struct {
    ID          string
    Agent       AgentTypeV2
    Action      string      // "analyze", "delegate", "execute", "synthesize"
    Input       string
    Output      string
    Reasoning   string
    Confidence  float64
    Duration    time.Duration
    Children    []string    // IDs of child steps (parallel)
    Status      string      // "pending", "running", "completed", "failed"
}
```

### COT Execution Flow

```
User Request
    │
    ▼
┌─────────────────┐
│ Step 1: Analyze │  Classify intent, determine strategy
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Step 2: Plan    │  Create ExecutionPlan with steps
└────────┬────────┘
         │
         ▼
┌──────────────────────────────────────────┐
│ Step 3: Execute (based on strategy)     │
├──────────────────────────────────────────┤
│  Direct:     Orch handles                │
│  Delegate:   Single specialist           │
│  Multi:      Parallel agents (WaitGroup) │
│  Sequential: Chain with history context  │
└────────┬─────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│ Step 4: Synthesize  │  (if multi-agent)
└────────┬────────┘
         │
         ▼
    Final Output
```

### Multi-Hop Sequential Example

```go
// User: "Research competitors, then create a competitive analysis document"

ExecutionPlan:
  Strategy: "sequential"
  Steps:
    1. Analyst Agent  → Research competitors
       ├─ Output: "Found 5 key competitors..."
       └─ Chain history captured

    2. Document Agent → Create analysis document
       ├─ Input: Original request + Analyst's output
       └─ Uses chain history as context

Final: Synthesis of all outputs
```

### Chain History Format

```go
type ChainHistory struct {
    Steps []ChainStep
}

type ChainStep struct {
    Order     int
    Agent     AgentTypeV2
    Task      string
    Input     string
    Output    string
    Reasoning string
}

// Formatted as context for next agent:
func (ch *ChainHistory) FormatAsContext() string {
    return `
    ## Previous Agent Chain History

    ### Step 1: @analyst Agent
    **Output:**
    [previous agent's output...]

    ### Step 2: @document Agent
    **Output:**
    [previous agent's output...]

    ## Your Task
    Based on the above context, provide your contribution.
    `
}
```

---

## Creating Custom Agents

### Step 1: Define Agent Structure

```go
package myagent

import (
    "context"
    "github.com/rhl/businessos-backend/internal/agents"
    "github.com/rhl/businessos-backend/internal/streaming"
)

type MyCustomAgent struct {
    *agents.BaseAgentV2
}
```

### Step 2: Implement Constructor

```go
func New(ctx *agents.AgentContextV2) *MyCustomAgent {
    systemPrompt := "You are a specialist in..."

    base := agents.NewBaseAgentV2(agents.BaseAgentV2Config{
        Pool:           ctx.Pool,
        Config:         ctx.Config,
        UserID:         ctx.UserID,
        UserName:       ctx.UserName,
        ConversationID: ctx.ConversationID,
        AgentType:      agents.AgentTypeV2Custom, // Define custom type
        AgentName:      "My Custom Agent",
        Description:    "Handles custom domain logic",
        SystemPrompt:   systemPrompt,
        ContextReqs: agents.ContextRequirements{
            NeedsProjects:    true,
            MaxContextTokens: 8000,
        },
        EnabledTools: []string{
            "get_project",
            "create_task",
            "my_custom_tool",
        },
    })

    return &MyCustomAgent{
        BaseAgentV2: base,
    }
}
```

### Step 3: Implement AgentV2 Interface

```go
func (a *MyCustomAgent) Type() agents.AgentTypeV2 {
    return agents.AgentTypeV2Custom
}

func (a *MyCustomAgent) Run(ctx context.Context, input agents.AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
    // Option 1: Use base implementation
    return a.BaseAgentV2.Run(ctx, input)

    // Option 2: Custom implementation with pre/post processing
    events := make(chan streaming.StreamEvent, 100)
    errs := make(chan error, 1)

    go func() {
        defer close(events)
        defer close(errs)

        // Pre-process input
        modifiedInput := a.preprocess(input)

        // Call base agent
        baseEvents, baseErrs := a.BaseAgentV2.Run(ctx, modifiedInput)

        // Forward events with optional transformation
        for event := range baseEvents {
            events <- a.transformEvent(event)
        }

        // Handle errors
        if err := <-baseErrs; err != nil {
            errs <- err
        }
    }()

    return events, errs
}
```

### Step 4: Create Custom Tool (Optional)

```go
type MyCustomTool struct {
    pool   *pgxpool.Pool
    userID string
}

func (t *MyCustomTool) Name() string {
    return "my_custom_tool"
}

func (t *MyCustomTool) Description() string {
    return "Does something specific to my domain"
}

func (t *MyCustomTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
    var params struct {
        Param1 string `json:"param1"`
    }
    if err := json.Unmarshal(input, &params); err != nil {
        return "", err
    }

    // Execute custom logic
    result := performCustomOperation(ctx, t.pool, t.userID, params.Param1)

    return result, nil
}
```

### Step 5: Register Agent

```go
// In agent_v2.go or custom registry
func (r *AgentRegistryV2) GetAgent(
    agentType AgentTypeV2,
    userID string,
    userName string,
    conversationID *uuid.UUID,
    tieredContext *services.TieredContext,
) AgentV2 {
    ctx := &AgentContextV2{...}

    switch agentType {
    case AgentTypeV2Custom:
        return myagent.New(ctx)
    // ... other agents
    }
}
```

### Step 6: Add Intent Routing (Optional)

```go
// In intent_router_v2.go
func (r *SmartIntentRouter) initializePatterns() {
    r.patterns[AgentTypeV2Custom] = []*IntentPattern{
        {
            Pattern:     regexp.MustCompile(`(?i)(do custom thing|handle special case)`),
            Weight:      1.0,
            MustMatch:   true,
            Description: "Custom agent activation",
        },
    }
}
```

---

## Best Practices

### 1. Agent Design

**DO:**
- Keep agents focused on a single domain
- Use descriptive system prompts that clearly define the agent's role
- Declare explicit ContextRequirements
- Enable only necessary tools
- Return meaningful errors with context

**DON'T:**
- Create overlapping agents (leads to routing confusion)
- Over-enable tools (security and performance risk)
- Ignore context budget (can exceed LLM limits)
- Block on I/O in agent methods (use goroutines)

### 2. Tool Implementation

**DO:**
- Validate input thoroughly
- Use parameterized SQL queries (prevent injection)
- Return structured, parseable output
- Log tool execution for debugging
- Handle partial failures gracefully

**DON'T:**
- Execute arbitrary code from LLM
- Return sensitive data without filtering
- Assume input is valid
- Use blocking operations without timeouts

### 3. Streaming

**DO:**
- Use buffered channels (100+ capacity)
- Close channels in defer blocks
- Send EventTypeDone before closing
- Handle context cancellation
- Flush artifact detector at end

**DON'T:**
- Send to closed channels
- Forget to send done event
- Block indefinitely on channel sends
- Ignore errors from error channel

### 4. Context Management

**DO:**
- Request minimal required context
- Set reasonable MaxContextTokens
- Use PrioritySections for critical data
- Cache context when possible
- Filter sensitive information

**DON'T:**
- Load entire database into context
- Ignore token budgets
- Hardcode context in prompts
- Expose user credentials

### 5. Error Handling

**DO:**
- Return specific error messages
- Log errors with context (slog)
- Gracefully degrade on failures
- Retry transient errors
- Report errors through error channel

**DON'T:**
- Panic in agent code
- Return empty errors
- Swallow errors silently
- Expose stack traces to users

---

## Recent Changes & Improvements

### January 2026

**Agent V2 Architecture Stabilization:**
- Unified all agents under BaseAgentV2
- Standardized tool registry across agents
- Improved streaming event handling
- Added artifact detection for structured outputs

**SmartIntentRouter Enhancements:**
- Multi-layer classification (patterns + signals + context + LLM)
- Portuguese language support in patterns
- LLM fallback with 5s timeout
- Context-aware boosting from user selections

**Chain of Thought (COT) System:**
- Multi-agent orchestration with planning
- Parallel execution with WaitGroups
- Sequential multi-hop with chain history
- Execution strategy auto-detection
- COT summary tracking

**Tool System Improvements:**
- Context tools with embedding support (tree_search, browse_tree, load_context)
- Web search integration
- Bulk operations (bulk_create_tasks)
- Improved error messages
- Tool execution logging

**Context & Memory:**
- Role-based permission injection (SetRoleContextPrompt)
- Workspace memory hierarchy (SetMemoryContext)
- Agent skills system (SetSkillsPrompt)
- Prompt personalization service integration
- Tiered context with priority sections

### December 2025

**Initial V2 Launch:**
- Migrated from V1 delegation pattern to V2 streaming
- Introduced specialized agents (Document, Project, Task, Client, Analyst)
- Built tool calling infrastructure
- Implemented SSE streaming
- Created AgentRegistryV2

---

## Migration Guide (V1 → V2)

### Old Pattern (V1)

```go
agent := agents.NewOrchestratorAgent(pool, cfg, userID, &conversationID, model)
chunks, errs := agent.Run(ctx, messages)

for chunk := range chunks {
    fmt.Print(chunk)  // Raw string chunks
}
```

### New Pattern (V2)

```go
registry := agents.NewAgentRegistryV2(pool, cfg, embeddingService, promptPersonalizer)
agent := registry.GetAgent(agents.AgentTypeV2Orchestrator, userID, userName, &conversationID, tieredContext)

input := agents.AgentInput{
    Messages:  messages,
    Context:   tieredContext,
    UserID:    userID,
    UserName:  userName,
}

events, errs := agent.Run(ctx, input)

for event := range events {
    switch event.Type {
    case streaming.EventTypeToken:
        fmt.Print(event.Data)  // Content chunk
    case streaming.EventTypeArtifact:
        saveArtifact(event.Data)  // Structured output
    case streaming.EventTypeDone:
        // Completion
    }
}
```

---

## Troubleshooting

### Agent Not Routing Correctly

**Symptoms:** Wrong agent handles request
**Solutions:**
1. Check SmartIntentRouter patterns for your use case
2. Verify focus mode mapping (GetAgentForFocusModeV2)
3. Increase LLM classifier timeout if using
4. Add specific pattern with MustMatch: true

### Tools Not Working

**Symptoms:** "tool not enabled" errors
**Solutions:**
1. Verify tool is in agent's EnabledTools list
2. Check tool is registered in AgentToolRegistry
3. Confirm userID is set correctly
4. Check database permissions

### Streaming Stops

**Symptoms:** Response cuts off mid-stream
**Solutions:**
1. Check context cancellation
2. Verify channel buffer size (100+)
3. Check for panics in goroutines
4. Ensure EventTypeDone is sent
5. Check LLM service timeout

### Context Too Large

**Symptoms:** LLM errors, slow responses
**Solutions:**
1. Reduce MaxContextTokens
2. Use PrioritySections to filter
3. Limit NeedsFullHistory to false
4. Implement context summarization

---

## Additional Resources

- **Agent Prompts:** `/internal/prompts/agents/`
- **Tool Definitions:** `/internal/tools/agent_tools.go`
- **Streaming System:** `/internal/streaming/`
- **Context System:** `/internal/services/context.go`
- **Tests:** `/internal/agents/*_test.go`

---

**Questions?** Contact the BusinessOS team or check the main README.
