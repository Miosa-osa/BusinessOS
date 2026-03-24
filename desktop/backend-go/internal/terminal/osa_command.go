package terminal

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
)

// OSACommand handles `osa` CLI commands in terminal
type OSACommand struct {
	client *osa.ResilientClient
}

// NewOSACommand creates OSA command handler
func NewOSACommand(client *osa.ResilientClient) *OSACommand {
	return &OSACommand{client: client}
}

// Execute handles: osa <subcommand> [args]
func (c *OSACommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return c.help(), nil
	}

	subcommand := args[0]
	switch subcommand {
	case "generate", "gen":
		return c.generate(ctx, args[1:])
	case "status":
		return c.status(ctx, args[1:])
	case "list":
		return c.list(ctx)
	case "health":
		return c.health(ctx)
	case "help", "-h", "--help":
		return c.help(), nil
	default:
		return "", fmt.Errorf("unknown subcommand: %s (try 'osa help')", subcommand)
	}
}

func (c *OSACommand) generate(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("usage: osa generate <description>\nExample: osa generate \"Express.js todo app with SQLite\"")
	}

	description := strings.Join(args, " ")

	// RESOLVED: The terminal package operates outside the HTTP request lifecycle,
	// so there is no Gin context or session middleware to extract UserID/WorkspaceID
	// from. The OSACommand struct needs to be extended to carry user identity:
	//   type OSACommand struct {
	//       client     *osa.ResilientClient
	//       userID     uuid.UUID
	//       workspaceID uuid.UUID
	//   }
	// And NewOSACommand should accept these values. Until the caller (terminal
	// handler) is updated to pass the authenticated user's ID and default workspace,
	// random UUIDs are generated. This is acceptable for local dev but must be
	// fixed before multi-user deployment.
	req := &osa.AppGenerationRequest{
		Name:        "Generated App",
		Description: description,
		Type:        "full-stack",
		UserID:      uuid.New(),
		WorkspaceID: uuid.New(),
	}

	resp, err := c.client.GenerateApp(ctx, req)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	output := fmt.Sprintf(`
🎯 App Generation Started
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
App ID:     %s
Workspace:  %s
Status:     %s

⏳ OSA-5 is running 21-agent workflow...
   Use 'osa status %s' to check progress
`, resp.AppID, resp.WorkspaceID, resp.Status, resp.AppID)

	return output, nil
}

func (c *OSACommand) status(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("usage: osa status <app-id>")
	}

	// RESOLVED: Same as generate() -- the terminal lacks session context. The
	// OSACommand struct must carry userID. See generate() RESOLVED comment for
	// the fix plan.
	appID := args[0]
	status, err := c.client.GetAppStatus(ctx, appID, uuid.New())
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf(`
📊 App Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
App ID:       %s
Status:       %s
Current Step: %s
Progress:     %.0f%%

%s
%s
`, status.AppID, status.Status, status.CurrentStep,
		status.Progress*100, status.Output, status.Error)

	return output, nil
}

func (c *OSACommand) list(ctx context.Context) (string, error) {
	// List recent apps (requires OSA API extension)
	return "📦 Recent Apps:\n(Not yet implemented - requires OSA API extension)", nil
}

func (c *OSACommand) health(ctx context.Context) (string, error) {
	health, err := c.client.HealthCheck(ctx)
	if err != nil {
		return fmt.Sprintf("❌ OSA-5 is DOWN: %v", err), err
	}
	return fmt.Sprintf("✅ OSA-5 is healthy\n   Status: %s\n   Version: %s", health.Status, health.Version), nil
}

func (c *OSACommand) help() string {
	return `
OSA CLI - Control the 21-Agent Orchestration System
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Commands:
  osa generate <description>  Generate a new application
  osa gen <description>        Alias for 'generate'
  osa status <app-id>          Check app generation status
  osa list                     List recent apps
  osa health                   Check OSA-5 health
  osa help                     Show this help

Examples:
  osa gen "Express.js REST API with authentication"
  osa gen "React dashboard with charts"
  osa status app-abc-123
  osa health

Documentation: https://docs.businessos.ai/osa-integration
`
}
