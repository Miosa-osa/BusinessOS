package project

import (
	"context"

	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ProjectAgent handles project management and planning tasks
type ProjectAgent struct {
	*agents.BaseAgentV2
}

// New creates a new ProjectAgent
func New(ctx *agents.AgentContextV2) *ProjectAgent {
	systemPrompt := prompts.DefaultComposer.ComposeForProject(prompts_agents.ProjectAgentPrompt)

	base := agents.NewBaseAgentV2(agents.BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      agents.AgentTypeV2Project,
		AgentName:      "Project Manager",
		Description:    "Manages projects, tasks, timelines, and team coordination",
		SystemPrompt:   systemPrompt,
		ContextReqs: agents.ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsTeam:        true,
			MaxContextTokens: 8000,
			PrioritySections: []string{"project_details", "active_tasks", "team_members", "milestones"},
		},
		EnabledTools: []string{
			"create_project", "update_project", "get_project", "list_projects",
			"create_task", "bulk_create_tasks", "assign_task",
			"get_team_capacity", "search_documents",
			"create_artifact", "log_activity",
		},
	})

	return &ProjectAgent{
		BaseAgentV2: base,
	}
}

// Type returns the agent type
func (a *ProjectAgent) Type() agents.AgentTypeV2 {
	return agents.AgentTypeV2Project
}

// Run executes the project agent - delegates to base implementation
func (a *ProjectAgent) Run(ctx context.Context, input agents.AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	return a.BaseAgentV2.Run(ctx, input)
}
