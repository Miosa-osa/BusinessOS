// Package webhooks handles incoming webhooks from integration providers.
package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rhl/businessos-backend/internal/services"
)

// Handler processes incoming webhooks from various providers.
type Handler struct {
	pool        *pgxpool.Pool
	syncService *services.SyncService
	verifier    *SignatureVerifier
	logger      *slog.Logger
}

// NewHandler creates a new webhook handler.
func NewHandler(pool *pgxpool.Pool, syncService *services.SyncService, secrets map[string]string, logger *slog.Logger) *Handler {
	return &Handler{
		pool:        pool,
		syncService: syncService,
		verifier:    NewSignatureVerifier(secrets),
		logger:      logger,
	}
}

// RegisterRoutes registers webhook routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	webhooks := r.Group("/webhooks")
	{
		// Google Calendar push notifications
		webhooks.POST("/google/calendar", h.GoogleCalendarWebhook)

		// Slack Events API
		webhooks.POST("/slack/events", h.SlackEventsWebhook)

		// Linear webhooks
		webhooks.POST("/linear", h.LinearWebhook)

		// HubSpot webhooks
		webhooks.POST("/hubspot", h.HubSpotWebhook)

		// Notion webhooks
		webhooks.POST("/notion", h.NotionWebhook)

		// ClickUp webhooks
		webhooks.POST("/clickup", h.ClickUpWebhook)

		// Airtable webhooks
		webhooks.POST("/airtable", h.AirtableWebhook)

		// Fathom webhooks
		webhooks.POST("/fathom", h.FathomWebhook)

		// Microsoft Graph webhooks
		webhooks.POST("/microsoft", h.MicrosoftWebhook)
	}
}

// =============================================================================
// GOOGLE CALENDAR WEBHOOKS
// =============================================================================

// GoogleCalendarWebhook handles Google Calendar push notifications.
func (h *Handler) GoogleCalendarWebhook(c *gin.Context) {
	channelID := c.GetHeader("X-Goog-Channel-ID")
	resourceID := c.GetHeader("X-Goog-Resource-ID")
	resourceState := c.GetHeader("X-Goog-Resource-State")
	channelToken := c.GetHeader("X-Goog-Channel-Token")

	h.logger.Info("Google Calendar webhook received",
		slog.String("channel_id", channelID),
		slog.String("resource_id", resourceID),
		slog.String("resource_state", resourceState),
	)

	// Verify channel token
	if !h.verifier.VerifyGoogleChannelToken(channelToken) {
		h.logger.Warn("Invalid Google Calendar channel token")
		c.Status(http.StatusUnauthorized)
		return
	}

	// Handle different resource states
	switch resourceState {
	case "sync":
		// Initial sync message - acknowledge and return
		c.Status(http.StatusOK)
		return
	case "exists":
		// Resource exists/updated - queue a sync
		go h.syncGoogleCalendarEvents(context.Background(), channelToken, resourceID)
	case "not_exists":
		// Resource deleted
		h.logger.Info("Calendar resource deleted", slog.String("resource_id", resourceID))
	}

	c.Status(http.StatusOK)
}

func (h *Handler) syncGoogleCalendarEvents(ctx context.Context, userIDStr, resourceID string) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error("Invalid user ID in channel token", slog.String("token", userIDStr), slog.Any("error", err))
		return
	}

	// Get sync token for incremental sync
	syncToken, err := h.syncService.GetSyncToken(ctx, userID, "google", "calendar")
	if err != nil {
		h.logger.Error("Failed to get sync token", slog.Any("error", err))
	}

	h.logger.Info("Syncing Google Calendar events",
		slog.String("user_id", userID.String()),
		slog.String("sync_token", syncToken),
	)

	// TODO: Implement actual Google Calendar API call to fetch events
	// This would use the Google Calendar API client with the user's OAuth token
	// For now, we log the intent and acknowledge the webhook
}

// =============================================================================
// SLACK EVENTS API
// =============================================================================

// SlackEventsWebhook handles Slack Events API webhooks.
func (h *Handler) SlackEventsWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// First, check if this is a URL verification challenge
	var challenge struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		c.JSON(http.StatusOK, gin.H{"challenge": challenge.Challenge})
		return
	}

	// Verify the request signature
	timestamp := c.GetHeader("X-Slack-Request-Timestamp")
	signature := c.GetHeader("X-Slack-Signature")
	if !h.verifier.VerifySlackSignature(body, timestamp, signature) {
		h.logger.Warn("Invalid Slack webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Parse the event
	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	h.logger.Info("Slack webhook received",
		slog.String("type", event.Type),
		slog.String("team_id", event.TeamID),
	)

	// Handle different event types
	switch event.Type {
	case "event_callback":
		go h.handleSlackEventCallback(context.Background(), event)
	}

	c.Status(http.StatusOK)
}

// SlackEvent represents a Slack event.
type SlackEvent struct {
	Type      string          `json:"type"`
	TeamID    string          `json:"team_id"`
	Event     json.RawMessage `json:"event"`
	EventID   string          `json:"event_id"`
	EventTime int64           `json:"event_time"`
	APIAppID  string          `json:"api_app_id"`
}

// SlackMessageEvent represents a message event.
type SlackMessageEvent struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	User        string `json:"user"`
	Text        string `json:"text"`
	TS          string `json:"ts"`
	ThreadTS    string `json:"thread_ts,omitempty"`
	ChannelType string `json:"channel_type"`
}

func (h *Handler) handleSlackEventCallback(ctx context.Context, event SlackEvent) {
	var innerEvent struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(event.Event, &innerEvent); err != nil {
		h.logger.Error("Failed to parse Slack inner event", slog.Any("error", err))
		return
	}

	h.logger.Info("Processing Slack event",
		slog.String("type", innerEvent.Type),
		slog.String("team_id", event.TeamID),
	)

	switch innerEvent.Type {
	case "message":
		var msgEvent SlackMessageEvent
		if err := json.Unmarshal(event.Event, &msgEvent); err == nil {
			h.syncSlackMessage(ctx, event.TeamID, msgEvent)
		}
	case "channel_created":
		h.logger.Info("New Slack channel created", slog.String("team_id", event.TeamID))
	case "channel_deleted":
		h.logger.Info("Slack channel deleted", slog.String("team_id", event.TeamID))
	case "member_joined_channel":
		h.logger.Info("Member joined Slack channel", slog.String("team_id", event.TeamID))
	}
}

func (h *Handler) syncSlackMessage(ctx context.Context, teamID string, msg SlackMessageEvent) {
	// Look up user by Slack team ID
	userID, err := h.syncService.GetUserIDByProviderTeam(ctx, "slack", teamID)
	if err != nil {
		h.logger.Error("Failed to find user for Slack team",
			slog.String("team_id", teamID),
			slog.Any("error", err),
		)
		return
	}

	// Parse timestamp to time
	var sentAt *time.Time
	if msg.TS != "" {
		// Slack timestamps are Unix timestamps with microseconds as decimal
		parts := strings.Split(msg.TS, ".")
		if len(parts) > 0 {
			var ts int64
			fmt.Sscanf(parts[0], "%d", &ts)
			t := time.Unix(ts, 0)
			sentAt = &t
		}
	}

	// Sync message to database
	message := services.SyncedMessage{
		UserID:        userID,
		Provider:      "slack",
		ExternalID:    msg.TS, // Slack uses timestamp as message ID
		ChannelID:     msg.Channel,
		ChannelType:   msg.ChannelType,
		SenderID:      msg.User,
		Content:       msg.Text,
		ThreadID:      msg.ThreadTS,
		IsThreadReply: msg.ThreadTS != "",
		SentAt:        sentAt,
		RawData: map[string]interface{}{
			"ts":           msg.TS,
			"channel":      msg.Channel,
			"user":         msg.User,
			"channel_type": msg.ChannelType,
		},
	}

	_, err = h.syncService.UpsertMessage(ctx, message)
	if err != nil {
		h.logger.Error("Failed to sync Slack message",
			slog.String("channel", msg.Channel),
			slog.Any("error", err),
		)
	}
}

// =============================================================================
// LINEAR WEBHOOKS
// =============================================================================

// LinearWebhook handles Linear webhooks.
func (h *Handler) LinearWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("Linear-Signature")
	if !h.verifier.VerifyLinearSignature(body, signature) {
		h.logger.Warn("Invalid Linear webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	var event LinearWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	h.logger.Info("Linear webhook received",
		slog.String("action", event.Action),
		slog.String("type", event.Type),
		slog.String("organization_id", event.OrganizationID),
	)

	// Handle different webhook types
	switch event.Type {
	case "Issue":
		go h.handleLinearIssueEvent(context.Background(), event)
	case "Project":
		go h.handleLinearProjectEvent(context.Background(), event)
	case "Comment":
		go h.handleLinearCommentEvent(context.Background(), event)
	}

	c.Status(http.StatusOK)
}

// LinearWebhookEvent represents a Linear webhook event.
type LinearWebhookEvent struct {
	Action           string          `json:"action"` // create, update, remove
	Type             string          `json:"type"`   // Issue, Project, Comment, etc.
	Data             json.RawMessage `json:"data"`
	OrganizationID   string          `json:"organizationId"`
	WebhookID        string          `json:"webhookId"`
	WebhookTimestamp int64           `json:"webhookTimestamp"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// LinearIssue represents a Linear issue from webhook data.
type LinearIssue struct {
	ID          string `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    int    `json:"priority"`
	Estimate    int    `json:"estimate"`
	DueDate     string `json:"dueDate"`
	State       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Assignee *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"assignee"`
	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Project *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Labels struct {
		Nodes []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"nodes"`
	} `json:"labels"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

func (h *Handler) handleLinearIssueEvent(ctx context.Context, event LinearWebhookEvent) {
	var issue LinearIssue
	if err := json.Unmarshal(event.Data, &issue); err != nil {
		h.logger.Error("Failed to parse Linear issue", slog.Any("error", err))
		return
	}

	h.logger.Info("Processing Linear issue",
		slog.String("action", event.Action),
		slog.String("identifier", issue.Identifier),
		slog.String("title", issue.Title),
		slog.String("state", issue.State.Name),
	)

	// Look up user by Linear organization
	userID, err := h.syncService.GetUserIDByProviderTeam(ctx, "linear", event.OrganizationID)
	if err != nil {
		h.logger.Error("Failed to find user for Linear organization",
			slog.String("organization_id", event.OrganizationID),
			slog.Any("error", err),
		)
		return
	}

	if event.Action == "remove" {
		// Delete the task
		if err := h.syncService.DeleteTask(ctx, userID, "linear", issue.ID); err != nil {
			h.logger.Error("Failed to delete Linear task", slog.Any("error", err))
		}
		return
	}

	// Map priority (Linear: 0=no priority, 1=urgent, 2=high, 3=normal, 4=low)
	priorityNames := map[int]string{0: "", 1: "urgent", 2: "high", 3: "normal", 4: "low"}
	priority := priorityNames[issue.Priority]

	// Map state type
	stateType := issue.State.Type
	if stateType == "" {
		// Infer from state name
		switch strings.ToLower(issue.State.Name) {
		case "backlog", "triage":
			stateType = "backlog"
		case "todo", "to do":
			stateType = "todo"
		case "in progress", "in review":
			stateType = "in_progress"
		case "done", "completed":
			stateType = "done"
		case "canceled", "cancelled":
			stateType = "canceled"
		}
	}

	// Parse due date
	var dueDate *time.Time
	if issue.DueDate != "" {
		if t, err := time.Parse("2006-01-02", issue.DueDate); err == nil {
			dueDate = &t
		}
	}

	// Build labels
	var labels []services.TaskLabel
	for _, l := range issue.Labels.Nodes {
		labels = append(labels, services.TaskLabel{
			ID:    l.ID,
			Name:  l.Name,
			Color: l.Color,
		})
	}

	// Build task
	task := services.SyncedTask{
		UserID:        userID,
		Provider:      "linear",
		ExternalID:    issue.ID,
		Identifier:    issue.Identifier,
		URL:           issue.URL,
		Title:         issue.Title,
		Description:   issue.Description,
		Status:        issue.State.Name,
		StatusType:    stateType,
		Priority:      priority,
		PriorityOrder: issue.Priority,
		TeamID:        issue.Team.ID,
		TeamName:      issue.Team.Name,
		DueDate:       dueDate,
		StartedAt:     issue.StartedAt,
		CompletedAt:   issue.CompletedAt,
		Labels:        labels,
		Estimate:      issue.Estimate,
		RawData: map[string]interface{}{
			"organization_id": event.OrganizationID,
			"webhook_id":      event.WebhookID,
		},
	}

	if issue.Assignee != nil {
		task.AssigneeID = issue.Assignee.ID
		task.AssigneeName = issue.Assignee.Name
		task.AssigneeEmail = issue.Assignee.Email
	}

	if issue.Project != nil {
		task.ProjectID = issue.Project.ID
		task.ProjectName = issue.Project.Name
	}

	_, err = h.syncService.UpsertTask(ctx, task)
	if err != nil {
		h.logger.Error("Failed to sync Linear issue",
			slog.String("identifier", issue.Identifier),
			slog.Any("error", err),
		)
	}
}

func (h *Handler) handleLinearProjectEvent(ctx context.Context, event LinearWebhookEvent) {
	var project struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(event.Data, &project); err != nil {
		h.logger.Error("Failed to parse Linear project", slog.Any("error", err))
		return
	}

	h.logger.Info("Linear project event",
		slog.String("action", event.Action),
		slog.String("name", project.Name),
		slog.String("state", project.State),
	)

	// TODO: Sync project to local database
}

func (h *Handler) handleLinearCommentEvent(ctx context.Context, event LinearWebhookEvent) {
	var comment struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(event.Data, &comment); err != nil {
		h.logger.Error("Failed to parse Linear comment", slog.Any("error", err))
		return
	}

	h.logger.Info("Linear comment event",
		slog.String("action", event.Action),
		slog.String("body", truncateString(comment.Body, 50)),
	)

	// TODO: Sync comment to local database
}

// =============================================================================
// HUBSPOT WEBHOOKS
// =============================================================================

// HubSpotWebhook handles HubSpot webhooks.
func (h *Handler) HubSpotWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature (v3)
	signature := c.GetHeader("X-HubSpot-Signature-v3")
	timestamp := c.GetHeader("X-HubSpot-Request-Timestamp")
	if !h.verifier.VerifyHubSpotSignature(body, signature, timestamp, c.Request.Method, c.Request.URL.Path) {
		h.logger.Warn("Invalid HubSpot webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// HubSpot sends batches of events
	var events []HubSpotWebhookEvent
	if err := json.Unmarshal(body, &events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid events"})
		return
	}

	h.logger.Info("HubSpot webhook received", slog.Int("event_count", len(events)))

	for _, event := range events {
		h.logger.Info("Processing HubSpot event",
			slog.String("type", event.SubscriptionType),
			slog.String("object_type", event.ObjectType),
			slog.Int64("object_id", event.ObjectID),
		)

		switch event.ObjectType {
		case "contact":
			go h.handleHubSpotContactEvent(context.Background(), event)
		case "company":
			go h.handleHubSpotCompanyEvent(context.Background(), event)
		case "deal":
			go h.handleHubSpotDealEvent(context.Background(), event)
		}
	}

	c.Status(http.StatusOK)
}

// HubSpotWebhookEvent represents a HubSpot webhook event.
type HubSpotWebhookEvent struct {
	ObjectID         int64  `json:"objectId"`
	ObjectType       string `json:"objectTypeId"`
	ChangeSource     string `json:"changeSource"`
	EventID          int64  `json:"eventId"`
	SubscriptionID   int64  `json:"subscriptionId"`
	SubscriptionType string `json:"subscriptionType"` // contact.creation, contact.propertyChange, etc.
	PortalID         int64  `json:"portalId"`
	OccurredAt       int64  `json:"occurredAt"`
	PropertyName     string `json:"propertyName"`
	PropertyValue    string `json:"propertyValue"`
}

func (h *Handler) handleHubSpotContactEvent(ctx context.Context, event HubSpotWebhookEvent) {
	h.logger.Info("Processing HubSpot contact event",
		slog.Int64("object_id", event.ObjectID),
		slog.String("type", event.SubscriptionType),
	)

	// Look up user by HubSpot portal ID
	userID, err := h.syncService.GetUserIDByProviderPortal(ctx, "hubspot", event.PortalID)
	if err != nil {
		h.logger.Error("Failed to find user for HubSpot portal",
			slog.Int64("portal_id", event.PortalID),
			slog.Any("error", err),
		)
		return
	}

	// TODO: Fetch full contact data from HubSpot API and sync
	// For now, we create a placeholder with the ID
	contact := services.SyncedContact{
		UserID:     userID,
		Provider:   "hubspot",
		ExternalID: fmt.Sprintf("%d", event.ObjectID),
		RawData: map[string]interface{}{
			"portal_id":         event.PortalID,
			"subscription_type": event.SubscriptionType,
			"property_name":     event.PropertyName,
			"property_value":    event.PropertyValue,
		},
	}

	_, err = h.syncService.UpsertContact(ctx, contact)
	if err != nil {
		h.logger.Error("Failed to sync HubSpot contact",
			slog.Int64("object_id", event.ObjectID),
			slog.Any("error", err),
		)
	}
}

func (h *Handler) handleHubSpotCompanyEvent(ctx context.Context, event HubSpotWebhookEvent) {
	h.logger.Info("HubSpot company event",
		slog.Int64("object_id", event.ObjectID),
		slog.String("type", event.SubscriptionType),
	)
	// TODO: Sync company to local database
}

func (h *Handler) handleHubSpotDealEvent(ctx context.Context, event HubSpotWebhookEvent) {
	h.logger.Info("HubSpot deal event",
		slog.Int64("object_id", event.ObjectID),
		slog.String("type", event.SubscriptionType),
	)
	// TODO: Sync deal to local database
}

// =============================================================================
// NOTION WEBHOOKS
// =============================================================================

// NotionWebhook handles Notion webhooks.
func (h *Handler) NotionWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Notion-Signature")
	if !h.verifier.VerifyNotionSignature(body, signature) {
		h.logger.Warn("Invalid Notion webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	h.logger.Info("Notion webhook received", slog.String("body_preview", truncateString(string(body), 100)))

	// TODO: Parse and handle Notion webhook events
	c.Status(http.StatusOK)
}

// =============================================================================
// CLICKUP WEBHOOKS
// =============================================================================

// ClickUpWebhook handles ClickUp webhooks.
func (h *Handler) ClickUpWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var event struct {
		Event     string          `json:"event"`
		WebhookID string          `json:"webhook_id"`
		TaskID    string          `json:"task_id"`
		HistoryID string          `json:"history_id"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	// Verify signature using webhook_id
	if !h.verifier.VerifyClickUpSignature(body, event.WebhookID) {
		h.logger.Warn("Invalid ClickUp webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	h.logger.Info("ClickUp webhook received",
		slog.String("event", event.Event),
		slog.String("task_id", event.TaskID),
	)

	// TODO: Handle ClickUp events
	c.Status(http.StatusOK)
}

// =============================================================================
// AIRTABLE WEBHOOKS
// =============================================================================

// AirtableWebhook handles Airtable webhooks.
func (h *Handler) AirtableWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Airtable-Content-MAC")
	timestamp := c.GetHeader("X-Airtable-Deliver-Time")
	if !h.verifier.VerifyAirtableSignature(body, signature, timestamp) {
		h.logger.Warn("Invalid Airtable webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	h.logger.Info("Airtable webhook received", slog.String("body_preview", truncateString(string(body), 100)))

	// TODO: Parse and handle Airtable webhook events
	c.Status(http.StatusOK)
}

// =============================================================================
// FATHOM WEBHOOKS
// =============================================================================

// FathomWebhook handles Fathom webhooks.
func (h *Handler) FathomWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Fathom-Signature")
	if !h.verifier.VerifyFathomSignature(body, signature) {
		h.logger.Warn("Invalid Fathom webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	var event struct {
		Type      string          `json:"type"`
		Data      json.RawMessage `json:"data"`
		MeetingID string          `json:"meeting_id"`
		Timestamp string          `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	h.logger.Info("Fathom webhook received",
		slog.String("type", event.Type),
		slog.String("meeting_id", event.MeetingID),
	)

	// Handle meeting completion events
	if event.Type == "meeting.completed" || event.Type == "meeting.processed" {
		go h.handleFathomMeetingEvent(context.Background(), event.Data)
	}

	c.Status(http.StatusOK)
}

func (h *Handler) handleFathomMeetingEvent(ctx context.Context, data json.RawMessage) {
	var meeting struct {
		ID              string   `json:"id"`
		Title           string   `json:"title"`
		StartTime       string   `json:"start_time"`
		EndTime         string   `json:"end_time"`
		DurationSeconds int      `json:"duration_seconds"`
		Summary         string   `json:"summary"`
		Transcript      string   `json:"transcript"`
		RecordingURL    string   `json:"recording_url"`
		Participants    []string `json:"participants"`
		ActionItems     []string `json:"action_items"`
		CalendarEventID string   `json:"calendar_event_id"`
		OrganizationID  string   `json:"organization_id"`
	}
	if err := json.Unmarshal(data, &meeting); err != nil {
		h.logger.Error("Failed to parse Fathom meeting", slog.Any("error", err))
		return
	}

	h.logger.Info("Processing Fathom meeting",
		slog.String("meeting_id", meeting.ID),
		slog.String("title", meeting.Title),
	)

	// Look up user by Fathom organization
	userID, err := h.syncService.GetUserIDByProviderTeam(ctx, "fathom", meeting.OrganizationID)
	if err != nil {
		h.logger.Error("Failed to find user for Fathom organization",
			slog.String("organization_id", meeting.OrganizationID),
			slog.Any("error", err),
		)
		return
	}

	// Parse times
	var startTime, endTime *time.Time
	if meeting.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, meeting.StartTime); err == nil {
			startTime = &t
		}
	}
	if meeting.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, meeting.EndTime); err == nil {
			endTime = &t
		}
	}

	// Build participants
	var participants []services.MeetingParticipant
	for _, p := range meeting.Participants {
		participants = append(participants, services.MeetingParticipant{Name: p})
	}

	// Build action items
	var actionItems []services.MeetingActionItem
	for _, item := range meeting.ActionItems {
		actionItems = append(actionItems, services.MeetingActionItem{Task: item})
	}

	// Sync meeting
	syncedMeeting := services.SyncedMeeting{
		UserID:           userID,
		Provider:         "fathom",
		ExternalID:       meeting.ID,
		Title:            meeting.Title,
		StartTime:        startTime,
		EndTime:          endTime,
		DurationSeconds:  meeting.DurationSeconds,
		Participants:     participants,
		ParticipantCount: len(participants),
		Transcript:       meeting.Transcript,
		Summary:          meeting.Summary,
		ActionItems:      actionItems,
		RecordingURL:     meeting.RecordingURL,
		CalendarEventID:  meeting.CalendarEventID,
		RawData: map[string]interface{}{
			"organization_id": meeting.OrganizationID,
		},
	}

	_, err = h.syncService.UpsertMeeting(ctx, syncedMeeting)
	if err != nil {
		h.logger.Error("Failed to sync Fathom meeting",
			slog.String("meeting_id", meeting.ID),
			slog.Any("error", err),
		)
	}
}

// =============================================================================
// MICROSOFT GRAPH WEBHOOKS
// =============================================================================

// MicrosoftWebhook handles Microsoft Graph webhooks.
func (h *Handler) MicrosoftWebhook(c *gin.Context) {
	// Handle validation request
	validationToken := c.Query("validationToken")
	if validationToken != "" {
		h.logger.Info("Microsoft webhook validation", slog.String("token", validationToken))
		c.String(http.StatusOK, validationToken)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var notification struct {
		Value []struct {
			SubscriptionID                 string `json:"subscriptionId"`
			ClientState                    string `json:"clientState"`
			ChangeType                     string `json:"changeType"`
			Resource                       string `json:"resource"`
			SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime"`
			ResourceData                   struct {
				ID string `json:"id"`
			} `json:"resourceData"`
		} `json:"value"`
	}
	if err := json.Unmarshal(body, &notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification"})
		return
	}

	for _, change := range notification.Value {
		// Verify client state
		if !h.verifier.VerifyMicrosoftClientState(change.ClientState) {
			h.logger.Warn("Invalid Microsoft clientState",
				slog.String("subscription_id", change.SubscriptionID),
			)
			continue
		}

		h.logger.Info("Microsoft webhook received",
			slog.String("subscription_id", change.SubscriptionID),
			slog.String("change_type", change.ChangeType),
			slog.String("resource", change.Resource),
		)

		// TODO: Fetch full resource data and sync
	}

	c.Status(http.StatusAccepted)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
