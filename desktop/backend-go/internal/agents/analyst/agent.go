package analyst

import (
	"context"

	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// AnalystAgent handles data analysis and insights
type AnalystAgent struct {
	*agents.BaseAgentV2
}

// New creates a new AnalystAgent
func New(ctx *agents.AgentContextV2) *AnalystAgent {
	systemPrompt := prompts.DefaultComposer.ComposeForAnalysis(prompts_agents.AnalystAgentPrompt)

	base := agents.NewBaseAgentV2(agents.BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      agents.AgentTypeV2Analyst,
		AgentName:      "Business Analyst",
		Description:    "Analyzes data, metrics, trends, and provides business insights",
		SystemPrompt:   systemPrompt,
		ContextReqs: agents.ContextRequirements{
			NeedsProjects:    true,
			NeedsMetrics:     true,
			MaxContextTokens: 8000,
			PrioritySections: []string{"metrics_data", "historical_trends", "kpis"},
		},
		EnabledTools: []string{
			"query_metrics", "get_team_capacity",
			"list_projects", "list_tasks", "get_project",
			"search_documents",
			"create_artifact", "log_activity",
		},
	})

	return &AnalystAgent{
		BaseAgentV2: base,
	}
}

// Type returns the agent type
func (a *AnalystAgent) Type() agents.AgentTypeV2 {
	return agents.AgentTypeV2Analyst
}

// Run executes the analyst agent - delegates to base implementation
func (a *AnalystAgent) Run(ctx context.Context, input agents.AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	return a.BaseAgentV2.Run(ctx, input)
}
