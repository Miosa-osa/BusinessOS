package handlers

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// NotificationSeedHandler handles seeding test notifications (DEV ONLY)
type NotificationSeedHandler struct {
	pool *pgxpool.Pool
	svc  *services.NotificationService
}

func NewNotificationSeedHandler(pool *pgxpool.Pool, svc *services.NotificationService) *NotificationSeedHandler {
	return &NotificationSeedHandler{pool: pool, svc: svc}
}

// IsDevMode checks if we're running in development mode
func IsDevMode() bool {
	env := os.Getenv("ENVIRONMENT")
	ginMode := os.Getenv("GIN_MODE")
	return env != "production" && ginMode != "release"
}

// SeedNotifications creates test notifications for the current user
// POST /api/dev/notifications/seed
func (h *NotificationSeedHandler) SeedNotifications(c *gin.Context) {
	if !IsDevMode() {
		utils.RespondForbidden(c, slog.Default(), "seed endpoint only available in development mode")
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	ctx := c.Request.Context()

	// Generate seed notifications
	notifications := generateSeedNotifications(user.ID)

	created := 0
	for _, input := range notifications {
		_, err := h.svc.Create(ctx, input)
		if err != nil {
			// Log but continue - some might fail due to batching
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Seed notifications created",
		"count":   created,
	})
}

// ClearSeedNotifications removes all notifications for current user
// DELETE /api/dev/notifications/seed
func (h *NotificationSeedHandler) ClearSeedNotifications(c *gin.Context) {
	if !IsDevMode() {
		utils.RespondForbidden(c, slog.Default(), "seed endpoint only available in development mode")
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	ctx := c.Request.Context()

	// Delete all notifications for this user
	result, err := h.pool.Exec(ctx, `DELETE FROM notifications WHERE user_id = $1`, user.ID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "clear notifications", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notifications cleared",
		"deleted": result.RowsAffected(),
	})
}

// generateSeedNotifications creates a diverse set of test notifications
func generateSeedNotifications(userID string) []services.CreateInput {
	now := time.Now()

	// Sample entity IDs (fake but valid UUIDs)
	taskID1 := uuid.New()
	taskID2 := uuid.New()
	taskID3 := uuid.New()
	projectID1 := uuid.New()
	projectID2 := uuid.New()
	clientID := uuid.New()

	// Sample senders
	senders := []struct {
		ID   string
		Name string
	}{
		{"sender-001", "Sarah Chen"},
		{"sender-002", "Mike Johnson"},
		{"sender-003", "Emily Davis"},
		{"sender-004", "Alex Thompson"},
		{"sender-005", "System"},
	}

	// getSender available for future use
	_ = func(idx int) (string, string) {
		s := senders[idx%len(senders)]
		return s.ID, s.Name
	}

	notifications := []services.CreateInput{
		// ============ TODAY - URGENT/HIGH PRIORITY ============
		{
			UserID:     userID,
			Type:       services.NotifTaskOverdue,
			Title:      "Task overdue: Q4 Financial Report",
			Body:       "This task was due yesterday and needs immediate attention.",
			EntityType: "task",
			EntityID:   &taskID1,
			SenderID:   senders[0].ID,
			SenderName: senders[0].Name,
			Priority:   services.PriorityUrgent,
			Metadata:   map[string]interface{}{"due_date": now.Add(-24 * time.Hour).Format(time.RFC3339), "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifTaskDueToday,
			Title:      "Task due today: Review PR #142",
			Body:       "Code review for the notifications feature is due today.",
			EntityType: "task",
			EntityID:   &taskID2,
			SenderID:   senders[1].ID,
			SenderName: senders[1].Name,
			Priority:   services.PriorityHigh,
			Metadata:   map[string]interface{}{"due_date": now.Format(time.RFC3339), "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifIntegrationSyncFailed,
			Title:      "Google Calendar sync failed",
			Body:       "Unable to sync calendar events. Please reconnect your account.",
			EntityType: "integration",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityHigh,
			Metadata:   map[string]interface{}{"integration": "google_calendar", "error": "token_expired", "seeded": true, "seed_time": "today"},
		},

		// ============ TODAY - NORMAL PRIORITY ============
		{
			UserID:     userID,
			Type:       services.NotifTaskAssigned,
			Title:      "You were assigned: Update API documentation",
			Body:       "Sarah Chen assigned you to this task in Project Alpha.",
			EntityType: "task",
			EntityID:   &taskID3,
			SenderID:   senders[0].ID,
			SenderName: senders[0].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"project_name": "Project Alpha", "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifMentionComment,
			Title:      "@you in: Design Review Discussion",
			Body:       "Mike Johnson mentioned you: \"@you what do you think about this approach?\"",
			EntityType: "comment",
			EntityID:   &taskID1,
			SenderID:   senders[1].ID,
			SenderName: senders[1].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"comment_preview": "what do you think about this approach?", "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifProjectAdded,
			Title:      "Added to project: Mobile App Redesign",
			Body:       "You've been added as a contributor to this project.",
			EntityType: "project",
			EntityID:   &projectID1,
			SenderID:   senders[2].ID,
			SenderName: senders[2].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"role": "contributor", "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifTaskComment,
			Title:      "New comment on: Update API documentation",
			Body:       "Emily Davis commented: \"I've added some notes to the shared doc.\"",
			EntityType: "task",
			EntityID:   &taskID3,
			SenderID:   senders[2].ID,
			SenderName: senders[2].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"comment_preview": "I've added some notes to the shared doc.", "seeded": true, "seed_time": "today"},
		},
		{
			UserID:     userID,
			Type:       services.NotifClientMeetingScheduled,
			Title:      "Meeting scheduled with Acme Corp",
			Body:       "Tomorrow at 2:00 PM - Quarterly business review",
			EntityType: "client",
			EntityID:   &clientID,
			SenderID:   senders[3].ID,
			SenderName: senders[3].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"meeting_time": now.Add(24 * time.Hour).Format(time.RFC3339), "client_name": "Acme Corp", "seeded": true, "seed_time": "today"},
		},

		// ============ YESTERDAY ============
		{
			UserID:     userID,
			Type:       services.NotifTaskCompleted,
			Title:      "Task completed: Setup CI/CD pipeline",
			Body:       "Alex Thompson marked this task as complete.",
			EntityType: "task",
			SenderID:   senders[3].ID,
			SenderName: senders[3].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"seeded": true, "seed_time": "yesterday", "created_offset": "-20h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifTeamMemberJoined,
			Title:      "New team member: Jordan Park",
			Body:       "Jordan Park joined the Engineering team.",
			EntityType: "team",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityLow,
			Metadata:   map[string]interface{}{"team_name": "Engineering", "seeded": true, "seed_time": "yesterday", "created_offset": "-22h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifMentionTask,
			Title:      "@you in task: Database Migration Plan",
			Body:       "Sarah Chen mentioned you in a task description.",
			EntityType: "task",
			SenderID:   senders[0].ID,
			SenderName: senders[0].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"seeded": true, "seed_time": "yesterday", "created_offset": "-26h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifProjectStatusChanged,
			Title:      "Project status: Mobile App Redesign → In Progress",
			Body:       "Project moved from Planning to In Progress.",
			EntityType: "project",
			EntityID:   &projectID1,
			SenderID:   senders[2].ID,
			SenderName: senders[2].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"old_status": "planning", "new_status": "in_progress", "seeded": true, "seed_time": "yesterday", "created_offset": "-28h"},
		},

		// ============ THIS WEEK ============
		{
			UserID:     userID,
			Type:       services.NotifTaskDueSoon,
			Title:      "Task due soon: Prepare presentation slides",
			Body:       "This task is due in 3 days.",
			EntityType: "task",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityHigh,
			Metadata:   map[string]interface{}{"days_until_due": 3, "seeded": true, "seed_time": "this_week", "created_offset": "-72h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifClientDealUpdate,
			Title:      "Deal update: Acme Corp - Enterprise Plan",
			Body:       "Deal value updated to $50,000. Stage: Negotiation.",
			EntityType: "client",
			EntityID:   &clientID,
			SenderID:   senders[1].ID,
			SenderName: senders[1].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"deal_value": 50000, "stage": "negotiation", "seeded": true, "seed_time": "this_week", "created_offset": "-96h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifIntegrationConnected,
			Title:      "Slack connected successfully",
			Body:       "Your Slack workspace is now connected.",
			EntityType: "integration",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityLow,
			Metadata:   map[string]interface{}{"integration": "slack", "seeded": true, "seed_time": "this_week", "created_offset": "-100h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifProjectCompleted,
			Title:      "Project completed: Website Refresh",
			Body:       "Congratulations! The Website Refresh project is now complete.",
			EntityType: "project",
			EntityID:   &projectID2,
			SenderID:   senders[2].ID,
			SenderName: senders[2].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"seeded": true, "seed_time": "this_week", "created_offset": "-120h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifChatArtifactReady,
			Title:      "AI artifact ready: Market Analysis Report",
			Body:       "Your requested analysis is ready to view.",
			EntityType: "chat",
			SenderID:   "ai-assistant",
			SenderName: "AI Assistant",
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"artifact_type": "report", "seeded": true, "seed_time": "this_week", "created_offset": "-130h"},
		},

		// ============ EARLIER (OLDER) ============
		{
			UserID:     userID,
			Type:       services.NotifSystemWelcome,
			Title:      "Welcome to Business OS!",
			Body:       "Get started by creating your first project or exploring the dashboard.",
			EntityType: "system",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityLow,
			Metadata:   map[string]interface{}{"seeded": true, "seed_time": "earlier", "created_offset": "-240h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifSystemFeatureAnnouncement,
			Title:      "New feature: AI-powered task suggestions",
			Body:       "Try our new AI feature that suggests task breakdowns automatically.",
			EntityType: "system",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityLow,
			Metadata:   map[string]interface{}{"feature": "ai_task_suggestions", "seeded": true, "seed_time": "earlier", "created_offset": "-336h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifDailyLogReminder,
			Title:      "Don't forget your daily log",
			Body:       "Take a moment to log your progress for today.",
			EntityType: "dailylog",
			SenderID:   "system",
			SenderName: "System",
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"seeded": true, "seed_time": "earlier", "created_offset": "-360h"},
		},
		{
			UserID:     userID,
			Type:       services.NotifTeamRoleChanged,
			Title:      "Your role updated: Project Lead",
			Body:       "You've been promoted to Project Lead for the Engineering team.",
			EntityType: "team",
			SenderID:   senders[0].ID,
			SenderName: senders[0].Name,
			Priority:   services.PriorityNormal,
			Metadata:   map[string]interface{}{"old_role": "member", "new_role": "project_lead", "seeded": true, "seed_time": "earlier", "created_offset": "-400h"},
		},
	}

	// Randomly mark some as read (about 40%)
	rand.Seed(time.Now().UnixNano())
	for i := range notifications {
		if rand.Float32() < 0.4 {
			if notifications[i].Metadata == nil {
				notifications[i].Metadata = make(map[string]interface{})
			}
			notifications[i].Metadata["pre_read"] = true
		}
	}

	return notifications
}

// CreateNotificationWithTimestamp creates a notification with a specific timestamp
// This bypasses the normal service to allow backdated notifications
func (h *NotificationSeedHandler) createWithTimestamp(ctx context.Context, input services.CreateInput, createdAt time.Time) error {
	// Get offset from metadata if present
	if offset, ok := input.Metadata["created_offset"].(string); ok {
		duration, err := time.ParseDuration(offset)
		if err == nil {
			createdAt = time.Now().Add(duration)
		}
	}

	// Direct SQL insert to allow custom created_at
	query := `
		INSERT INTO notifications (
			user_id, workspace_id, type, title, body, entity_type, entity_id,
			sender_id, sender_name, priority, metadata, is_read, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	isRead := false
	if preRead, ok := input.Metadata["pre_read"].(bool); ok && preRead {
		isRead = true
	}

	_, err := h.pool.Exec(ctx, query,
		input.UserID,
		input.WorkspaceID,
		input.Type,
		input.Title,
		input.Body,
		input.EntityType,
		input.EntityID,
		input.SenderID,
		input.SenderName,
		input.Priority,
		input.Metadata,
		isRead,
		createdAt,
	)
	return err
}

// SeedNotificationsWithTimestamps creates test notifications with varied timestamps
// POST /api/dev/notifications/seed-full
func (h *NotificationSeedHandler) SeedNotificationsWithTimestamps(c *gin.Context) {
	if !IsDevMode() {
		utils.RespondForbidden(c, slog.Default(), "seed endpoint only available in development mode")
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	ctx := c.Request.Context()
	now := time.Now()

	// Generate seed notifications
	notifications := generateSeedNotifications(user.ID)

	created := 0
	for _, input := range notifications {
		createdAt := now

		// Parse offset from metadata
		if offset, ok := input.Metadata["created_offset"].(string); ok {
			duration, err := time.ParseDuration(offset)
			if err == nil {
				createdAt = now.Add(duration)
			}
		}

		err := h.createWithTimestamp(ctx, input, createdAt)
		if err != nil {
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Seed notifications created with timestamps",
		"count":   created,
	})
}
