// Package webhooks handles incoming webhooks from integration providers.
package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// Handler processes incoming webhooks from various providers.
type Handler struct {
	pool   *pgxpool.Pool
	config *config.Config
}

// NewHandler creates a new webhook handler.
func NewHandler(pool *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		pool:   pool,
		config: cfg,
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

		// Notion webhooks (when available)
		webhooks.POST("/notion", h.NotionWebhook)
	}
}

// ============================================================================
// Google Calendar Webhooks
// ============================================================================

// GoogleCalendarWebhook handles Google Calendar push notifications.
func (h *Handler) GoogleCalendarWebhook(c *gin.Context) {
	// Google sends push notifications for calendar changes
	// Headers contain the channel info
	channelID := c.GetHeader("X-Goog-Channel-ID")
	resourceID := c.GetHeader("X-Goog-Resource-ID")
	resourceState := c.GetHeader("X-Goog-Resource-State")
	channelToken := c.GetHeader("X-Goog-Channel-Token")

	log.Printf("[Webhook] Google Calendar: channel=%s resource=%s state=%s token=%s",
		channelID, resourceID, resourceState, channelToken)

	// Handle different resource states
	switch resourceState {
	case "sync":
		// Initial sync message - acknowledge and return
		c.Status(http.StatusOK)
		return
	case "exists":
		// Resource exists/updated - queue a sync
		h.queueCalendarSync(channelID, resourceID, channelToken)
	case "not_exists":
		// Resource deleted
		log.Printf("[Webhook] Calendar resource deleted: %s", resourceID)
	}

	c.Status(http.StatusOK)
}

func (h *Handler) queueCalendarSync(channelID, resourceID, token string) {
	// Parse token to get user_id (we store user_id in the token)
	userID := token

	// Queue a sync job via background_jobs table
	ctx := context.TODO()
	query := `
		INSERT INTO background_jobs (
			job_type,
			payload,
			status,
			priority,
			max_attempts,
			created_at,
			updated_at
		) VALUES (
			'calendar_sync',
			jsonb_build_object(
				'user_id', $1,
				'channel_id', $2,
				'resource_id', $3,
				'provider', 'google'
			),
			'pending',
			1,
			3,
			NOW(),
			NOW()
		)
	`

	_, err := h.pool.Exec(ctx, query, userID, channelID, resourceID)
	if err != nil {
		log.Printf("[Webhook] Failed to queue calendar sync for user %s: %v", userID, err)
		return
	}

	log.Printf("[Webhook] Successfully queued calendar sync job for user %s (channel: %s)", userID, channelID)
}

// ============================================================================
// Slack Events API
// ============================================================================

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
	if !h.verifySlackSignature(body, timestamp, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Parse the event
	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	// Handle different event types
	switch event.Type {
	case "event_callback":
		h.handleSlackEventCallback(event)
	}

	c.Status(http.StatusOK)
}

// SlackEvent represents a Slack event.
type SlackEvent struct {
	Type       string          `json:"type"`
	TeamID     string          `json:"team_id"`
	Event      json.RawMessage `json:"event"`
	EventID    string          `json:"event_id"`
	EventTime  int64           `json:"event_time"`
	APIAppID   string          `json:"api_app_id"`
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

func (h *Handler) handleSlackEventCallback(event SlackEvent) {
	// Parse the inner event
	var innerEvent struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(event.Event, &innerEvent); err != nil {
		log.Printf("[Webhook] Failed to parse Slack inner event: %v", err)
		return
	}

	log.Printf("[Webhook] Slack event: type=%s team=%s", innerEvent.Type, event.TeamID)

	switch innerEvent.Type {
	case "message":
		var msgEvent SlackMessageEvent
		if err := json.Unmarshal(event.Event, &msgEvent); err == nil {
			h.handleSlackMessage(event.TeamID, msgEvent)
		}
	case "channel_created":
		log.Printf("[Webhook] New Slack channel created in team %s", event.TeamID)
	case "channel_deleted":
		log.Printf("[Webhook] Slack channel deleted in team %s", event.TeamID)
	case "member_joined_channel":
		log.Printf("[Webhook] Member joined channel in team %s", event.TeamID)
	}
}

func (h *Handler) handleSlackMessage(teamID string, msg SlackMessageEvent) {
	log.Printf("[Webhook] Slack message in channel %s from %s: %s",
		msg.Channel, msg.User, truncateString(msg.Text, 50))

	// TODO: Save message to database or trigger notification
}

func (h *Handler) verifySlackSignature(body []byte, timestamp, signature string) bool {
	// Get signing secret from config
	signingSecret := h.config.WebhookSigningSecret

	if signingSecret == "" {
		// Skip verification if no secret configured (development mode)
		return true
	}

	// Create the signature base string
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))

	// Calculate HMAC
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(baseString))
	expectedSig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ============================================================================
// Linear Webhooks
// ============================================================================

// LinearWebhook handles Linear webhooks.
func (h *Handler) LinearWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Verify signature if configured
	signature := c.GetHeader("Linear-Signature")
	if signature != "" && h.config.LinearWebhookSecret != "" {
		if !h.verifyLinearSignature(body, signature) {
			log.Printf("[Webhook] Linear: invalid signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	var event LinearWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	log.Printf("[Webhook] Linear: action=%s type=%s", event.Action, event.Type)

	// Handle different webhook types
	switch event.Type {
	case "Issue":
		h.handleLinearIssueEvent(event)
	case "Project":
		h.handleLinearProjectEvent(event)
	case "Comment":
		h.handleLinearCommentEvent(event)
	}

	c.Status(http.StatusOK)
}

// LinearWebhookEvent represents a Linear webhook event.
type LinearWebhookEvent struct {
	Action         string          `json:"action"` // create, update, remove
	Type           string          `json:"type"`   // Issue, Project, Comment, etc.
	Data           json.RawMessage `json:"data"`
	OrganizationID string          `json:"organizationId"`
	WebhookID      string          `json:"webhookId"`
	CreatedAt      time.Time       `json:"createdAt"`
}

func (h *Handler) handleLinearIssueEvent(event LinearWebhookEvent) {
	var issue struct {
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
		Title      string `json:"title"`
		State      struct {
			Name string `json:"name"`
		} `json:"state"`
	}
	if err := json.Unmarshal(event.Data, &issue); err != nil {
		log.Printf("[Webhook] Failed to parse Linear issue: %v", err)
		return
	}

	log.Printf("[Webhook] Linear issue %s: %s (%s) - %s",
		event.Action, issue.Identifier, issue.Title, issue.State.Name)

	// TODO: Update local database
}

func (h *Handler) handleLinearProjectEvent(event LinearWebhookEvent) {
	var project struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(event.Data, &project); err != nil {
		log.Printf("[Webhook] Failed to parse Linear project: %v", err)
		return
	}

	log.Printf("[Webhook] Linear project %s: %s (%s)",
		event.Action, project.Name, project.State)
}

func (h *Handler) handleLinearCommentEvent(event LinearWebhookEvent) {
	var comment struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(event.Data, &comment); err != nil {
		log.Printf("[Webhook] Failed to parse Linear comment: %v", err)
		return
	}

	log.Printf("[Webhook] Linear comment %s: %s",
		event.Action, truncateString(comment.Body, 50))
}

// verifyLinearSignature verifies the Linear webhook signature using HMAC-SHA256.
// Linear uses the format: HMAC-SHA256(webhook_secret, request_body)
func (h *Handler) verifyLinearSignature(body []byte, signature string) bool {
	secret := h.config.LinearWebhookSecret

	if secret == "" {
		// Skip verification if no secret configured (development mode)
		return true
	}

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Linear signatures are hex-encoded HMAC-SHA256
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ============================================================================
// HubSpot Webhooks
// ============================================================================

// HubSpotWebhook handles HubSpot webhooks.
func (h *Handler) HubSpotWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// HubSpot sends batches of events
	var events []HubSpotWebhookEvent
	if err := json.Unmarshal(body, &events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid events"})
		return
	}

	for _, event := range events {
		log.Printf("[Webhook] HubSpot: type=%s objectType=%s objectId=%d",
			event.SubscriptionType, event.ObjectType, event.ObjectID)

		switch event.ObjectType {
		case "contact":
			h.handleHubSpotContactEvent(event)
		case "company":
			h.handleHubSpotCompanyEvent(event)
		case "deal":
			h.handleHubSpotDealEvent(event)
		}
	}

	c.Status(http.StatusOK)
}

// HubSpotWebhookEvent represents a HubSpot webhook event.
type HubSpotWebhookEvent struct {
	ObjectID         int64  `json:"objectId"`
	ObjectType       string `json:"objectTypeId"` // contact, company, deal
	ChangeSource     string `json:"changeSource"`
	EventID          int64  `json:"eventId"`
	SubscriptionID   int64  `json:"subscriptionId"`
	SubscriptionType string `json:"subscriptionType"` // contact.creation, contact.propertyChange, etc.
	PortalID         int64  `json:"portalId"`
	OccurredAt       int64  `json:"occurredAt"`
}

func (h *Handler) handleHubSpotContactEvent(event HubSpotWebhookEvent) {
	log.Printf("[Webhook] HubSpot contact event: id=%d type=%s",
		event.ObjectID, event.SubscriptionType)

	// TODO: Fetch updated contact and sync to local database
}

func (h *Handler) handleHubSpotCompanyEvent(event HubSpotWebhookEvent) {
	log.Printf("[Webhook] HubSpot company event: id=%d type=%s",
		event.ObjectID, event.SubscriptionType)
}

func (h *Handler) handleHubSpotDealEvent(event HubSpotWebhookEvent) {
	log.Printf("[Webhook] HubSpot deal event: id=%d type=%s",
		event.ObjectID, event.SubscriptionType)
}

// ============================================================================
// Notion Webhooks (Placeholder)
// ============================================================================

// NotionWebhook handles Notion webhooks.
// Note: Notion webhooks are limited and require specific setup.
func (h *Handler) NotionWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	log.Printf("[Webhook] Notion: %s", truncateString(string(body), 100))

	// TODO: Parse and handle Notion webhook events when they become available

	c.Status(http.StatusOK)
}

// ============================================================================
// Helper Functions
// ============================================================================

func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
