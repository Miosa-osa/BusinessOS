package client

import (
	"context"

	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ClientAgent handles client relationship management
type ClientAgent struct {
	*agents.BaseAgentV2
}

// New creates a new ClientAgent
func New(ctx *agents.AgentContextV2) *ClientAgent {
	systemPrompt := prompts.DefaultComposer.ComposeForClient(prompts_agents.ClientAgentPrompt)

	base := agents.NewBaseAgentV2(agents.BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      agents.AgentTypeV2Client,
		AgentName:      "Client Relationship Manager",
		Description:    "Manages client relationships, pipeline, and communications",
		SystemPrompt:   systemPrompt,
		ContextReqs: agents.ContextRequirements{
			NeedsClients:     true,
			NeedsProjects:    true,
			MaxContextTokens: 6000,
			PrioritySections: []string{"client_details", "pipeline_status", "interaction_history"},
		},
		EnabledTools: []string{
			"get_client", "create_client", "update_client",
			"log_client_interaction", "update_client_pipeline",
			"search_documents", "get_project",
			"create_artifact", "log_activity",
		},
	})

	return &ClientAgent{
		BaseAgentV2: base,
	}
}

// Type returns the agent type
func (a *ClientAgent) Type() agents.AgentTypeV2 {
	return agents.AgentTypeV2Client
}

// Run executes the client agent - delegates to base implementation
func (a *ClientAgent) Run(ctx context.Context, input agents.AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	return a.BaseAgentV2.Run(ctx, input)
}
