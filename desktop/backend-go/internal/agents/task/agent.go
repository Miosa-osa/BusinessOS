package task

import (
	"context"

	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// TaskAgent handles task management and prioritization
type TaskAgent struct {
	*agents.BaseAgentV2
}

// New creates a new TaskAgent
func New(ctx *agents.AgentContextV2) *TaskAgent {
	systemPrompt := prompts.Compose(prompts_agents.TaskAgentPrompt)

	base := agents.NewBaseAgentV2(agents.BaseAgentV2Config{
		Pool:           ctx.Pool,
		Config:         ctx.Config,
		UserID:         ctx.UserID,
		UserName:       ctx.UserName,
		ConversationID: ctx.ConversationID,
		AgentType:      agents.AgentTypeV2Task,
		AgentName:      "Task Specialist",
		Description:    "Task management, prioritization, scheduling, and dependencies",
		SystemPrompt:   systemPrompt,
		ContextReqs: agents.ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsTeam:        true,
			MaxContextTokens: 8000,
			PrioritySections: []string{"active_tasks", "project_tasks", "team_capacity"},
		},
		EnabledTools: []string{
			"create_task", "update_task", "get_task", "list_tasks",
			"bulk_create_tasks", "move_task", "assign_task",
			"get_team_capacity", "get_project",
			"log_activity",
		},
	})

	return &TaskAgent{
		BaseAgentV2: base,
	}
}

// Type returns the agent type
func (a *TaskAgent) Type() agents.AgentTypeV2 {
	return agents.AgentTypeV2Task
}

// Run executes the task agent - delegates to base implementation
func (a *TaskAgent) Run(ctx context.Context, input agents.AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	return a.BaseAgentV2.Run(ctx, input)
}
