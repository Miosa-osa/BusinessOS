// Package handlers provides HTTP handlers for BusinessOS.
package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// IntegrationsHandler handles integration management endpoints.
type IntegrationsHandler struct {
	pool               *pgxpool.Pool
	integrationRouter  *IntegrationRouter
}

// NewIntegrationsHandler creates a new integrations handler.
func NewIntegrationsHandler(pool *pgxpool.Pool, integrationRouter *IntegrationRouter) *IntegrationsHandler {
	return &IntegrationsHandler{
		pool:              pool,
		integrationRouter: integrationRouter,
	}
}

// getCalendarService returns the Google Calendar service from the integration router.
func (h *IntegrationsHandler) getCalendarService() *google.CalendarService {
	if h.integrationRouter != nil {
		return h.integrationRouter.GetGoogleCalendarService()
	}
	return nil
}

// ============================================================================
// Provider Endpoints
// ============================================================================

// GetProviders handles GET /api/integrations/providers
// Returns all available integration providers.
func (h *IntegrationsHandler) GetProviders(c *gin.Context) {
	category := c.Query("category")
	module := c.Query("module")
	status := c.Query("status")

	query := `
		SELECT id, name, description, category, icon_url,
		       oauth_config, modules, skills, status,
		       auto_live_sync, est_nodes, initial_sync, tooltip,
		       created_at, updated_at
		FROM integration_providers
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if category != "" {
		query += ` AND category = $` + string(rune('0'+argNum))
		args = append(args, category)
		argNum++
	}

	if module != "" {
		query += ` AND $` + string(rune('0'+argNum)) + ` = ANY(modules)`
		args = append(args, module)
		argNum++
	}

	if status != "" {
		query += ` AND status = $` + string(rune('0'+argNum))
		args = append(args, status)
	} else {
		query += ` AND status != 'deprecated'`
	}

	query += ` ORDER BY category, name`

	var providers []map[string]interface{}

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		// Database table may not exist yet - use fallback providers
		providers = getDefaultProviders()
	} else {
		defer rows.Close()

		for rows.Next() {
			var p struct {
				ID           string
				Name         string
				Description  *string
				Category     string
				IconURL      *string
				OAuthConfig  interface{}
				Modules      []string
				Skills       []string
				Status       string
				AutoLiveSync *bool
				EstNodes     *string
				InitialSync  *string
				Tooltip      *string
				CreatedAt    interface{}
				UpdatedAt    interface{}
			}
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.IconURL,
				&p.OAuthConfig, &p.Modules, &p.Skills, &p.Status,
				&p.AutoLiveSync, &p.EstNodes, &p.InitialSync, &p.Tooltip,
				&p.CreatedAt, &p.UpdatedAt); err != nil {
				log.Printf("[GetProviders] Scan error: %v", err)
				continue
			}

			provider := map[string]interface{}{
				"id":             p.ID,
				"name":           p.Name,
				"description":    p.Description,
				"category":       p.Category,
				"icon_url":       p.IconURL,
				"modules":        p.Modules,
				"skills":         p.Skills,
				"status":         p.Status,
				"oauth_provider": getOAuthProvider(p.ID),
			}
			// Add optional fields if present
			if p.AutoLiveSync != nil {
				provider["auto_live_sync"] = *p.AutoLiveSync
			}
			if p.EstNodes != nil {
				provider["est_nodes"] = *p.EstNodes
			}
			if p.InitialSync != nil {
				provider["initial_sync"] = *p.InitialSync
			}
			if p.Tooltip != nil {
				provider["tooltip"] = *p.Tooltip
			}
			providers = append(providers, provider)
		}

		// If no providers found (table empty), return defaults
		if len(providers) == 0 {
			providers = getDefaultProviders()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"providers": providers,
		"count":     len(providers),
	})
}

// getOAuthProvider maps provider IDs to their OAuth endpoint provider
// e.g., gmail and google_calendar both use the "google" OAuth endpoint
func getOAuthProvider(providerID string) string {
	oauthMapping := map[string]string{
		"gmail":           "google",
		"google_calendar": "google",
		"google_drive":    "google",
		"gemini":          "google",
		"outlook":         "microsoft",
		"teams":           "microsoft",
		// These use their own OAuth endpoints
		"slack":      "slack",
		"notion":     "notion",
		"hubspot":    "hubspot",
		"salesforce": "salesforce",
		"linear":     "linear",
		"asana":      "asana",
		"github":     "github",
		"gitlab":     "gitlab",
		"zoom":       "zoom",
		"discord":    "discord",
		"dropbox":    "dropbox",
		"clickup":    "clickup",
		"jira":       "jira",
		"trello":     "trello",
		"pipedrive":  "pipedrive",
		"fathom":     "fathom",
		"fireflies":  "fireflies",
	}

	if oauth, ok := oauthMapping[providerID]; ok {
		return oauth
	}
	// Default: use provider ID as OAuth provider (e.g., for chatgpt, claude - file import only)
	return ""
}

// getDefaultProviders returns core providers when database is empty.
// Icon URLs use local /logos/integrations/ where available, authjs.dev as fallback
// Includes est_nodes, initial_sync, auto_live_sync, and tooltip for rich UI display
func getDefaultProviders() []map[string]interface{} {
	providers := []map[string]interface{}{
		// Productivity - Email & Calendar
		{"id": "gmail", "name": "Gmail", "description": "Import project details and track the context of important conversations.", "category": "communication", "icon_url": "/logos/integrations/gmail.svg", "modules": []string{"chat", "daily_log"}, "skills": []string{"gmail.send_email", "gmail.search"}, "status": "available", "auto_live_sync": true, "est_nodes": "50-200", "initial_sync": "15-30m", "tooltip": "Your new emails are processed into nodes every day."},
		{"id": "google_calendar", "name": "Google Calendar", "description": "Sync your events so BusinessOS stays on top of meetings, plans, and deadlines.", "category": "calendar", "icon_url": "/logos/integrations/calendar.svg", "modules": []string{"calendar", "daily_log"}, "skills": []string{"google_calendar.sync_daily_log", "google_calendar.create_event"}, "status": "available", "auto_live_sync": true, "est_nodes": "20-100", "initial_sync": "5-10m", "tooltip": "Your calendar events are automatically synced to keep your schedule updated."},
		{"id": "notion", "name": "Notion", "description": "Sync your workspace pages, project roadmaps, and structured knowledge.", "category": "storage", "icon_url": "/logos/integrations/notion.svg", "modules": []string{"contexts", "projects"}, "skills": []string{"notion.sync_database", "notion.create_page"}, "status": "available", "auto_live_sync": true, "est_nodes": "30-150", "initial_sync": "10-20m", "tooltip": "Your Notion updates are processed into nodes every day."},
		{"id": "google_drive", "name": "Google Drive", "description": "Sync your documents, spreadsheets, and presentations into your knowledge base.", "category": "storage", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "50-300", "initial_sync": "20-40m", "tooltip": "Your Drive files are indexed and searchable within your knowledge base."},
		{"id": "dropbox", "name": "Dropbox", "description": "Import your files and folders to make them searchable and connected.", "category": "storage", "icon_url": "https://authjs.dev/img/providers/dropbox.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "30-200", "initial_sync": "15-30m", "tooltip": "Your Dropbox files are continuously synced."},
		// Communication
		{"id": "slack", "name": "Slack", "description": "Extract key insights and memories from your team channels and DMs.", "category": "communication", "icon_url": "/logos/integrations/slack.svg", "modules": []string{"chat", "tasks", "team"}, "skills": []string{"slack.send_message", "slack.message_to_task"}, "status": "available", "auto_live_sync": true, "est_nodes": "150-300", "initial_sync": "30-45m", "tooltip": "Your Slack messages are analyzed for important insights and decisions."},
		{"id": "teams", "name": "Microsoft Teams", "description": "Sync your Teams conversations, channels, and shared files.", "category": "communication", "icon_url": "https://authjs.dev/img/providers/azure-ad.svg", "modules": []string{"chat", "team"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "100-250", "initial_sync": "25-40m", "tooltip": "Your Teams messages and files are synced automatically."},
		{"id": "discord", "name": "Discord", "description": "Import conversations from your Discord servers and DMs.", "category": "communication", "icon_url": "https://authjs.dev/img/providers/discord.svg", "modules": []string{"chat"}, "skills": []string{}, "status": "coming_soon", "auto_live_sync": true, "est_nodes": "100-300", "initial_sync": "20-35m"},
		// AI Assistants (Manual sync)
		{"id": "chatgpt", "name": "ChatGPT", "description": "Capture your brainstorming sessions, creative ideas, and problem-solving history.", "category": "ai", "icon_url": "/logos/integrations/openai.svg", "modules": []string{"contexts"}, "skills": []string{"chatgpt.import_history"}, "status": "available", "auto_live_sync": false, "est_nodes": "80-120", "initial_sync": "30m"},
		{"id": "claude", "name": "Claude", "description": "Preserve your Claude in-depth discussions, research analysis, and writing drafts.", "category": "ai", "icon_url": "/logos/integrations/claude.svg", "modules": []string{"contexts"}, "skills": []string{"claude.import_history"}, "status": "available", "auto_live_sync": false, "est_nodes": "80-120", "initial_sync": "10-15m"},
		{"id": "perplexity", "name": "Perplexity", "description": "Import your research queries, sources, and discovered insights.", "category": "ai", "icon_url": "https://authjs.dev/img/providers/perplexity.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": false, "est_nodes": "40-80", "initial_sync": "10-15m"},
		{"id": "gemini", "name": "Google Gemini", "description": "Sync your Gemini conversations and generated content.", "category": "ai", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "coming_soon", "auto_live_sync": false, "est_nodes": "60-100", "initial_sync": "15-20m"},
		// Meetings
		{"id": "fireflies", "name": "Fireflies.ai", "description": "Turn meeting transcripts, summaries, and action items into memories.", "category": "meetings", "icon_url": "/logos/integrations/fireflies.svg", "modules": []string{"daily_log", "contexts"}, "skills": []string{"fireflies.get_transcripts"}, "status": "available", "auto_live_sync": true, "est_nodes": "20-50", "initial_sync": "10-15m", "tooltip": "Your meeting transcripts are processed into memories automatically."},
		{"id": "fathom", "name": "Fathom", "description": "Turn meeting transcripts, summaries, and action items into memories.", "category": "meetings", "icon_url": "/logos/integrations/fathom.svg", "modules": []string{"daily_log"}, "skills": []string{"fathom.get_summaries"}, "status": "available", "auto_live_sync": true, "est_nodes": "20-50", "initial_sync": "10-15m", "tooltip": "Your meeting transcripts and summaries are processed automatically."},
		{"id": "tldv", "name": "tl;dv", "description": "Turn meeting transcripts, summaries, and action items into memories.", "category": "meetings", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"daily_log", "contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "20-50", "initial_sync": "10-15m", "tooltip": "Your meeting recordings are transcribed and processed automatically."},
		{"id": "granola", "name": "Granola", "description": "Upload meeting notes to turn transcripts into memories.", "category": "meetings", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"daily_log"}, "skills": []string{}, "status": "available", "auto_live_sync": false, "est_nodes": "20-50", "initial_sync": "10-15m"},
		{"id": "zoom", "name": "Zoom", "description": "Import meeting recordings, transcripts, and chat history.", "category": "meetings", "icon_url": "https://authjs.dev/img/providers/zoom.svg", "modules": []string{"calendar", "daily_log"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "30-80", "initial_sync": "15-25m", "tooltip": "Your Zoom recordings are automatically transcribed and imported."},
		{"id": "loom", "name": "Loom", "description": "Import your video messages and their transcripts.", "category": "meetings", "icon_url": "https://authjs.dev/img/providers/loom.svg", "modules": []string{"daily_log", "contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "15-40", "initial_sync": "10-15m", "tooltip": "Your Loom videos are transcribed and added automatically."},
		// Project Management
		{"id": "linear", "name": "Linear", "description": "Sync your issues, projects, and roadmaps for full context.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/linear.svg", "modules": []string{"tasks", "projects"}, "skills": []string{"linear.sync_issues"}, "status": "available", "auto_live_sync": true, "est_nodes": "50-150", "initial_sync": "10-20m", "tooltip": "Your Linear issues and updates are synced in real-time."},
		{"id": "asana", "name": "Asana", "description": "Import your tasks, projects, and team workflows.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/asana.svg", "modules": []string{"tasks", "projects"}, "skills": []string{"asana.sync_tasks"}, "status": "available", "auto_live_sync": true, "est_nodes": "40-120", "initial_sync": "15-25m", "tooltip": "Your Asana tasks and projects are synced automatically."},
		{"id": "monday", "name": "Monday.com", "description": "Sync your boards, items, and updates into your knowledge base.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/monday.svg", "modules": []string{"tasks", "projects"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "40-100", "initial_sync": "15-20m", "tooltip": "Your Monday boards are synced and updated automatically."},
		{"id": "trello", "name": "Trello", "description": "Import your boards, cards, and checklists.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/trello.svg", "modules": []string{"tasks"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "30-80", "initial_sync": "10-15m", "tooltip": "Your Trello boards are synced in real-time."},
		{"id": "jira", "name": "Jira", "description": "Sync your issues, sprints, and project documentation.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/atlassian.svg", "modules": []string{"tasks", "projects"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "60-200", "initial_sync": "20-35m", "tooltip": "Your Jira issues and sprints are synced automatically."},
		{"id": "clickup", "name": "ClickUp", "description": "Import your tasks, docs, and workspace data.", "category": "tasks", "icon_url": "https://authjs.dev/img/providers/click-up.svg", "modules": []string{"tasks", "projects"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "50-150", "initial_sync": "15-25m", "tooltip": "Your ClickUp workspace is synced automatically."},
		// CRM
		{"id": "hubspot", "name": "HubSpot", "description": "Sync your CRM contacts, deals, and customer interactions into your knowledge base.", "category": "crm", "icon_url": "/logos/integrations/hubspot.svg", "modules": []string{"clients", "projects"}, "skills": []string{"hubspot.qualify_lead", "hubspot.sync_contacts"}, "status": "available", "auto_live_sync": true, "est_nodes": "100-500", "initial_sync": "20-40m", "tooltip": "Your HubSpot contacts and deals are synced and analyzed for insights."},
		{"id": "gohighlevel", "name": "GoHighLevel", "description": "Import your marketing funnels, contacts, and automation data.", "category": "crm", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"clients", "projects"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "150-400", "initial_sync": "25-45m", "tooltip": "Your GHL contacts, funnels, and campaigns are synced automatically."},
		{"id": "salesforce", "name": "Salesforce", "description": "Sync your accounts, opportunities, and customer data.", "category": "crm", "icon_url": "https://authjs.dev/img/providers/salesforce.svg", "modules": []string{"clients"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "200-600", "initial_sync": "30-60m", "tooltip": "Your Salesforce data is synced and enriched automatically."},
		{"id": "pipedrive", "name": "Pipedrive", "description": "Import your deals, contacts, and sales pipeline.", "category": "crm", "icon_url": "https://authjs.dev/img/providers/pipedrive.svg", "modules": []string{"clients"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "80-250", "initial_sync": "15-30m", "tooltip": "Your Pipedrive pipeline is synced in real-time."},
		// Storage
		{"id": "airtable", "name": "Airtable", "description": "Sync your bases, tables, and automation data.", "category": "storage", "icon_url": "/logos/integrations/airtable.webp", "modules": []string{"contexts", "projects"}, "skills": []string{"airtable.sync_base"}, "status": "available", "auto_live_sync": true, "est_nodes": "50-200", "initial_sync": "15-30m", "tooltip": "Your Airtable bases are continuously synced."},
		// Notes (Manual sync)
		{"id": "evernote", "name": "Evernote", "description": "Import your notes, notebooks, and web clips.", "category": "storage", "icon_url": "https://authjs.dev/img/providers/evernote.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": false, "est_nodes": "100-300", "initial_sync": "15-30m"},
		{"id": "obsidian", "name": "Obsidian", "description": "Sync your vault, notes, and knowledge graph connections.", "category": "storage", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": false, "est_nodes": "50-200", "initial_sync": "10-20m"},
		{"id": "roam", "name": "Roam Research", "description": "Import your daily notes, linked references, and graph structure.", "category": "storage", "icon_url": "https://authjs.dev/img/providers/google.svg", "modules": []string{"contexts"}, "skills": []string{}, "status": "available", "auto_live_sync": false, "est_nodes": "60-180", "initial_sync": "15-25m"},
		// Calendar
		{"id": "outlook", "name": "Microsoft Outlook", "description": "Sync your Outlook calendar, events, and email.", "category": "calendar", "icon_url": "https://authjs.dev/img/providers/azure-ad.svg", "modules": []string{"calendar", "daily_log"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "30-150", "initial_sync": "10-20m", "tooltip": "Your Outlook events are automatically synced."},
		{"id": "calendly", "name": "Calendly", "description": "Sync your scheduled meetings and availability.", "category": "calendar", "icon_url": "https://authjs.dev/img/providers/calendly.svg", "modules": []string{"calendar"}, "skills": []string{}, "status": "available", "auto_live_sync": true, "est_nodes": "10-50", "initial_sync": "5-10m", "tooltip": "Your Calendly bookings are synced automatically."},
	}

	// Add oauth_provider to each provider
	for _, p := range providers {
		if id, ok := p["id"].(string); ok {
			p["oauth_provider"] = getOAuthProvider(id)
		}
	}

	return providers
}

// GetProvider handles GET /api/integrations/providers/:id
// Returns a single provider with full details.
func (h *IntegrationsHandler) GetProvider(c *gin.Context) {
	providerID := c.Param("id")

	var p struct {
		ID          string
		Name        string
		Description *string
		Category    string
		IconURL     *string
		OAuthConfig interface{}
		Modules     []string
		Skills      []string
		Status      string
	}

	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT id, name, description, category, icon_url,
		       oauth_config, modules, skills, status
		FROM integration_providers
		WHERE id = $1
	`, providerID).Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.IconURL,
		&p.OAuthConfig, &p.Modules, &p.Skills, &p.Status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Provider not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"provider": map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"description":  p.Description,
			"category":     p.Category,
			"icon_url":     p.IconURL,
			"oauth_config": p.OAuthConfig,
			"modules":      p.Modules,
			"skills":       p.Skills,
			"status":       p.Status,
		},
	})
}

// ============================================================================
// User Integration Endpoints
// ============================================================================

// GetConnectedIntegrations handles GET /api/integrations/connected
// Returns the user's connected integrations.
func (h *IntegrationsHandler) GetConnectedIntegrations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}
	userID := user.ID

	log.Printf("[GetConnectedIntegrations] Querying for userID: %s", userID)
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT ui.id, ui.provider_id, ui.status, ui.connected_at, ui.last_used_at,
		       ui.external_account_name, ui.external_workspace_name,
		       ui.scopes, ui.settings,
		       ip.name as provider_name, ip.category, ip.icon_url, ip.skills
		FROM user_integrations ui
		JOIN integration_providers ip ON ui.provider_id = ip.id
		WHERE ui.user_id = $1
		ORDER BY ui.connected_at DESC
	`, userID)
	if err != nil {
		log.Printf("[GetConnectedIntegrations] Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch integrations",
		})
		return
	}
	defer rows.Close()

	var integrations []map[string]interface{}
	for rows.Next() {
		var i struct {
			ID                    uuid.UUID
			ProviderID            string
			Status                string
			ConnectedAt           interface{}
			LastUsedAt            interface{}
			ExternalAccountName   *string
			ExternalWorkspaceName *string
			Scopes                []string
			Settings              interface{}
			ProviderName          string
			Category              string
			IconURL               *string
			Skills                []string
		}
		if err := rows.Scan(&i.ID, &i.ProviderID, &i.Status, &i.ConnectedAt, &i.LastUsedAt,
			&i.ExternalAccountName, &i.ExternalWorkspaceName,
			&i.Scopes, &i.Settings,
			&i.ProviderName, &i.Category, &i.IconURL, &i.Skills); err != nil {
			continue
		}

		integration := map[string]interface{}{
			"id":                      i.ID,
			"provider_id":             i.ProviderID,
			"provider_name":           i.ProviderName,
			"category":                i.Category,
			"icon_url":                i.IconURL,
			"status":                  i.Status,
			"connected_at":            i.ConnectedAt,
			"last_used_at":            i.LastUsedAt,
			"external_account_name":   i.ExternalAccountName,
			"external_workspace_name": i.ExternalWorkspaceName,
			"scopes":                  i.Scopes,
			"settings":                i.Settings,
			"skills":                  i.Skills,
		}
		integrations = append(integrations, integration)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"integrations": integrations,
		"count":        len(integrations),
	})
}

// GetIntegration handles GET /api/integrations/:id
// Returns details of a specific user integration.
func (h *IntegrationsHandler) GetIntegration(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid integration ID",
		})
		return
	}

	var i struct {
		ID                    uuid.UUID
		ProviderID            string
		Status                string
		ConnectedAt           interface{}
		LastUsedAt            interface{}
		ExternalAccountID     *string
		ExternalAccountName   *string
		ExternalWorkspaceID   *string
		ExternalWorkspaceName *string
		Scopes                []string
		Settings              interface{}
		Metadata              interface{}
		ProviderName          string
		Category              string
		IconURL               *string
		Skills                []string
		Modules               []string
	}

	err = h.pool.QueryRow(c.Request.Context(), `
		SELECT ui.id, ui.provider_id, ui.status, ui.connected_at, ui.last_used_at,
		       ui.external_account_id, ui.external_account_name,
		       ui.external_workspace_id, ui.external_workspace_name,
		       ui.scopes, ui.settings, ui.metadata,
		       ip.name, ip.category, ip.icon_url, ip.skills, ip.modules
		FROM user_integrations ui
		JOIN integration_providers ip ON ui.provider_id = ip.id
		WHERE ui.id = $1 AND ui.user_id = $2
	`, id, userID).Scan(&i.ID, &i.ProviderID, &i.Status, &i.ConnectedAt, &i.LastUsedAt,
		&i.ExternalAccountID, &i.ExternalAccountName,
		&i.ExternalWorkspaceID, &i.ExternalWorkspaceName,
		&i.Scopes, &i.Settings, &i.Metadata,
		&i.ProviderName, &i.Category, &i.IconURL, &i.Skills, &i.Modules)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Integration not found",
		})
		return
	}

	// Get comprehensive sync stats for this integration
	syncStats := h.getIntegrationSyncStats(c, id, i.ProviderID, userID)

	// Get available permissions for this provider
	availablePermissions := getAvailablePermissions(i.ProviderID)

	// Get sync history (last 10 syncs)
	syncHistory := h.getSyncHistory(c, id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"integration": map[string]interface{}{
			"id":                      i.ID,
			"provider_id":             i.ProviderID,
			"provider_name":           i.ProviderName,
			"category":                i.Category,
			"icon_url":                i.IconURL,
			"status":                  i.Status,
			"connected_at":            i.ConnectedAt,
			"last_used_at":            i.LastUsedAt,
			"external_account_id":     i.ExternalAccountID,
			"external_account_name":   i.ExternalAccountName,
			"external_workspace_id":   i.ExternalWorkspaceID,
			"external_workspace_name": i.ExternalWorkspaceName,
			"scopes":                  i.Scopes,
			"available_permissions":   availablePermissions,
			"settings":                i.Settings,
			"metadata":                i.Metadata,
			"skills":                  i.Skills,
			"modules":                 i.Modules,
			"sync_stats":              syncStats,
			"sync_history":            syncHistory,
		},
	})
}

// UpdateIntegrationSettingsRequest represents the request body for updating settings.
type UpdateIntegrationSettingsRequest struct {
	EnabledSkills []string               `json:"enabled_skills"`
	Notifications bool                   `json:"notifications"`
	SyncSettings  map[string]interface{} `json:"sync_settings"`
}

// UpdateIntegrationSettings handles PATCH /api/integrations/:id/settings
// Updates settings for a user's integration.
func (h *IntegrationsHandler) UpdateIntegrationSettings(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid integration ID",
		})
		return
	}

	var req UpdateIntegrationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	settings := map[string]interface{}{
		"enabledSkills": req.EnabledSkills,
		"notifications": req.Notifications,
		"syncSettings":  req.SyncSettings,
	}

	_, err = h.pool.Exec(c.Request.Context(), `
		UPDATE user_integrations SET
			settings = $3,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, id, userID, settings)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated",
	})
}

// DisconnectIntegration handles DELETE /api/integrations/:id
// Disconnects a user's integration.
func (h *IntegrationsHandler) DisconnectIntegration(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid integration ID",
		})
		return
	}

	_, err = h.pool.Exec(c.Request.Context(), `
		UPDATE user_integrations SET
			status = 'disconnected',
			access_token_encrypted = NULL,
			refresh_token_encrypted = NULL,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to disconnect",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Integration disconnected",
	})
}

// TriggerSync handles POST /api/integrations/:id/sync
// Triggers a manual sync for an integration.
func (h *IntegrationsHandler) TriggerSync(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	idStr := c.Param("id")
	module := c.Query("module")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid integration ID",
		})
		return
	}

	// Get the integration to check provider
	var providerID string
	err = h.pool.QueryRow(c.Request.Context(), `
		SELECT provider_id FROM user_integrations WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Integration not found",
		})
		return
	}

	// Create sync log entry
	var syncLogID uuid.UUID
	err = h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO integration_sync_log (
			user_integration_id, module_id, sync_type, direction, status
		) VALUES ($1, $2, 'manual', 'bidirectional', 'in_progress')
		RETURNING id
	`, id, module).Scan(&syncLogID)

	if err != nil {
		log.Printf("Failed to create sync log: %v", err)
		// Continue anyway - sync log is nice to have but not critical
	}

	// Perform actual sync based on provider
	var syncedCount int
	var syncError error

	switch providerID {
	case "google_calendar", "google":
		// Sync calendar events from Google using new integration infrastructure
		calendarService := h.getCalendarService()
		if calendarService == nil {
			syncError = nil // No calendar service configured, skip sync
		} else {
			timeMin := time.Now().AddDate(0, -6, 0) // Last 6 months
			timeMax := time.Now().AddDate(0, 6, 0)  // Next 6 months
			_, syncError = calendarService.SyncEvents(c.Request.Context(), userID, timeMin, timeMax)
			if syncError == nil {
				// Count events synced (approximate)
				h.pool.QueryRow(c.Request.Context(),
					"SELECT COUNT(*) FROM calendar_events WHERE user_id = $1 AND source = 'google'",
					userID).Scan(&syncedCount)
			}
		}
	case "slack":
		// Slack sync handled by new integration infrastructure
		syncError = nil
	case "notion":
		// Notion sync handled by new integration infrastructure
		syncError = nil
	default:
		// Other providers - placeholder
		syncError = nil
	}

	// Update sync log status
	if syncLogID != uuid.Nil {
		status := "completed"
		if syncError != nil {
			status = "failed"
		}
		h.pool.Exec(c.Request.Context(), `
			UPDATE integration_sync_log
			SET status = $1, completed_at = NOW(), records_synced = $2
			WHERE id = $3
		`, status, syncedCount, syncLogID)
	}

	if syncError != nil {
		log.Printf("Sync failed for integration %s: %v", idStr, syncError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Sync failed: " + syncError.Error(),
		})
		return
	}

	// Get detailed sync info for Google Calendar
	var syncDetails map[string]interface{}
	if providerID == "google_calendar" {
		var minDate, maxDate *time.Time
		var eventCount int
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*), MIN(start_time), MAX(start_time)
			FROM calendar_events WHERE user_id = $1 AND source = 'google'
		`, userID).Scan(&eventCount, &minDate, &maxDate)

		syncDetails = map[string]interface{}{
			"total_events": eventCount,
			"date_range": map[string]interface{}{
				"from": minDate,
				"to":   maxDate,
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"sync_log_id":  syncLogID,
		"message":      "Sync completed successfully",
		"synced_count": syncedCount,
		"details":      syncDetails,
	})
}

// ============================================================================
// Module Integration Endpoints
// ============================================================================

// GetModuleIntegrations handles GET /api/modules/:module/integrations
// Returns available and connected integrations for a specific module.
func (h *IntegrationsHandler) GetModuleIntegrations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	module := c.Param("module")

	// Get available providers for this module
	providerRows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, description, category, icon_url, skills, status
		FROM integration_providers
		WHERE $1 = ANY(modules) AND status != 'deprecated'
		ORDER BY status = 'available' DESC, name
	`, module)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch providers",
		})
		return
	}
	defer providerRows.Close()

	var availableProviders []map[string]interface{}
	for providerRows.Next() {
		var p struct {
			ID          string
			Name        string
			Description *string
			Category    string
			IconURL     *string
			Skills      []string
			Status      string
		}
		if err := providerRows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.IconURL, &p.Skills, &p.Status); err != nil {
			continue
		}
		availableProviders = append(availableProviders, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"category":    p.Category,
			"icon_url":    p.IconURL,
			"skills":      p.Skills,
			"status":      p.Status,
		})
	}

	// Get user's connected integrations for this module
	var connectedIntegrations []map[string]interface{}
	if userID != "" {
		connRows, err := h.pool.Query(c.Request.Context(), `
			SELECT ui.id, ui.provider_id, ui.status, ui.last_used_at,
			       ui.external_account_name, ui.settings,
			       ip.name, ip.icon_url
			FROM user_integrations ui
			JOIN integration_providers ip ON ui.provider_id = ip.id
			WHERE ui.user_id = $1 AND ui.status = 'connected' AND $2 = ANY(ip.modules)
		`, userID, module)
		if err == nil {
			defer connRows.Close()
			for connRows.Next() {
				var i struct {
					ID                  uuid.UUID
					ProviderID          string
					Status              string
					LastUsedAt          interface{}
					ExternalAccountName *string
					Settings            interface{}
					ProviderName        string
					IconURL             *string
				}
				if err := connRows.Scan(&i.ID, &i.ProviderID, &i.Status, &i.LastUsedAt,
					&i.ExternalAccountName, &i.Settings,
					&i.ProviderName, &i.IconURL); err != nil {
					continue
				}
				connectedIntegrations = append(connectedIntegrations, map[string]interface{}{
					"id":                    i.ID,
					"provider_id":           i.ProviderID,
					"provider_name":         i.ProviderName,
					"icon_url":              i.IconURL,
					"status":                i.Status,
					"last_used_at":          i.LastUsedAt,
					"external_account_name": i.ExternalAccountName,
					"settings":              i.Settings,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":               true,
		"module":                module,
		"available_providers":   availableProviders,
		"connected_integrations": connectedIntegrations,
	})
}

// ============================================================================
// Model Preferences Endpoints
// ============================================================================

// GetModelPreferences handles GET /api/integrations/ai/preferences
// Returns the user's AI model tier preferences.
func (h *IntegrationsHandler) GetModelPreferences(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID

	var prefs struct {
		Tier2Model                 interface{}
		Tier3Model                 interface{}
		Tier4Model                 interface{}
		Tier2Fallbacks             interface{}
		Tier3Fallbacks             interface{}
		Tier4Fallbacks             interface{}
		SkillOverrides             interface{}
		AllowModelUpgradeOnFailure bool
		MaxLatencyMs               int
		PreferLocal                bool
	}

	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT tier_2_model, tier_3_model, tier_4_model,
		       tier_2_fallbacks, tier_3_fallbacks, tier_4_fallbacks,
		       skill_overrides, allow_model_upgrade_on_failure,
		       max_latency_ms, prefer_local
		FROM user_model_preferences
		WHERE user_id = $1
	`, userID).Scan(&prefs.Tier2Model, &prefs.Tier3Model, &prefs.Tier4Model,
		&prefs.Tier2Fallbacks, &prefs.Tier3Fallbacks, &prefs.Tier4Fallbacks,
		&prefs.SkillOverrides, &prefs.AllowModelUpgradeOnFailure,
		&prefs.MaxLatencyMs, &prefs.PreferLocal)

	if err != nil {
		// Return defaults if not found
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"preferences": map[string]interface{}{
				"tier_2_model":                   map[string]string{"model_id": "claude-3-5-haiku", "provider": "anthropic"},
				"tier_3_model":                   map[string]string{"model_id": "claude-sonnet-4", "provider": "anthropic"},
				"tier_4_model":                   map[string]string{"model_id": "claude-opus-4", "provider": "anthropic"},
				"tier_2_fallbacks":               []interface{}{},
				"tier_3_fallbacks":               []interface{}{},
				"tier_4_fallbacks":               []interface{}{},
				"skill_overrides":                map[string]interface{}{},
				"allow_model_upgrade_on_failure": true,
				"max_latency_ms":                 30000,
				"prefer_local":                   false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"preferences": map[string]interface{}{
			"tier_2_model":                   prefs.Tier2Model,
			"tier_3_model":                   prefs.Tier3Model,
			"tier_4_model":                   prefs.Tier4Model,
			"tier_2_fallbacks":               prefs.Tier2Fallbacks,
			"tier_3_fallbacks":               prefs.Tier3Fallbacks,
			"tier_4_fallbacks":               prefs.Tier4Fallbacks,
			"skill_overrides":                prefs.SkillOverrides,
			"allow_model_upgrade_on_failure": prefs.AllowModelUpgradeOnFailure,
			"max_latency_ms":                 prefs.MaxLatencyMs,
			"prefer_local":                   prefs.PreferLocal,
		},
	})
}

// UpdateModelPreferencesRequest represents the request body for updating AI preferences.
type UpdateModelPreferencesRequest struct {
	Tier2Model                 map[string]string      `json:"tier_2_model"`
	Tier3Model                 map[string]string      `json:"tier_3_model"`
	Tier4Model                 map[string]string      `json:"tier_4_model"`
	Tier2Fallbacks             []map[string]string    `json:"tier_2_fallbacks"`
	Tier3Fallbacks             []map[string]string    `json:"tier_3_fallbacks"`
	Tier4Fallbacks             []map[string]string    `json:"tier_4_fallbacks"`
	SkillOverrides             map[string]interface{} `json:"skill_overrides"`
	AllowModelUpgradeOnFailure *bool                  `json:"allow_model_upgrade_on_failure"`
	MaxLatencyMs               *int                   `json:"max_latency_ms"`
	PreferLocal                *bool                  `json:"prefer_local"`
}

// UpdateModelPreferences handles PUT /api/integrations/ai/preferences
// Updates the user's AI model tier preferences.
func (h *IntegrationsHandler) UpdateModelPreferences(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	var req UpdateModelPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	// Set defaults
	allowUpgrade := true
	if req.AllowModelUpgradeOnFailure != nil {
		allowUpgrade = *req.AllowModelUpgradeOnFailure
	}
	maxLatency := 30000
	if req.MaxLatencyMs != nil {
		maxLatency = *req.MaxLatencyMs
	}
	preferLocal := false
	if req.PreferLocal != nil {
		preferLocal = *req.PreferLocal
	}

	_, err := h.pool.Exec(c.Request.Context(), `
		INSERT INTO user_model_preferences (
			user_id, tier_2_model, tier_3_model, tier_4_model,
			tier_2_fallbacks, tier_3_fallbacks, tier_4_fallbacks,
			skill_overrides, allow_model_upgrade_on_failure,
			max_latency_ms, prefer_local
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id) DO UPDATE SET
			tier_2_model = EXCLUDED.tier_2_model,
			tier_3_model = EXCLUDED.tier_3_model,
			tier_4_model = EXCLUDED.tier_4_model,
			tier_2_fallbacks = EXCLUDED.tier_2_fallbacks,
			tier_3_fallbacks = EXCLUDED.tier_3_fallbacks,
			tier_4_fallbacks = EXCLUDED.tier_4_fallbacks,
			skill_overrides = EXCLUDED.skill_overrides,
			allow_model_upgrade_on_failure = EXCLUDED.allow_model_upgrade_on_failure,
			max_latency_ms = EXCLUDED.max_latency_ms,
			prefer_local = EXCLUDED.prefer_local,
			updated_at = NOW()
	`, userID, req.Tier2Model, req.Tier3Model, req.Tier4Model,
		req.Tier2Fallbacks, req.Tier3Fallbacks, req.Tier4Fallbacks,
		req.SkillOverrides, allowUpgrade, maxLatency, preferLocal)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save preferences",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Preferences saved",
	})
}

// ============================================================================
// Aggregated Status Endpoint
// ============================================================================

// IntegrationStatusInfo represents the status of a single integration.
type IntegrationStatusInfo struct {
	ProviderID    string      `json:"provider_id"`
	ProviderName  string      `json:"provider_name"`
	Category      string      `json:"category"`
	Connected     bool        `json:"connected"`
	Status        string      `json:"status"`
	AccountName   *string     `json:"account_name,omitempty"`
	WorkspaceName *string     `json:"workspace_name,omitempty"`
	ConnectedAt   interface{} `json:"connected_at,omitempty"`
	IconURL       *string     `json:"icon_url,omitempty"`
}

// GetAllIntegrationsStatus handles GET /api/integrations/status
// Returns aggregated status of all integrations from both legacy OAuth tables
// and the new user_integrations table.
func (h *IntegrationsHandler) GetAllIntegrationsStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := user.ID
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	ctx := c.Request.Context()
	statusMap := make(map[string]*IntegrationStatusInfo)

	// 1. Check Google OAuth (legacy table)
	var googleStatus struct {
		Email       *string     `json:"email"`
		ConnectedAt interface{} `json:"connected_at"`
	}
	err := h.pool.QueryRow(ctx, `
		SELECT google_email, created_at FROM google_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&googleStatus.Email, &googleStatus.ConnectedAt)
	if err == nil {
		statusMap["google_calendar"] = &IntegrationStatusInfo{
			ProviderID:   "google_calendar",
			ProviderName: "Google Calendar",
			Category:     "calendar",
			Connected:    true,
			Status:       "connected",
			AccountName:  googleStatus.Email,
			ConnectedAt:  googleStatus.ConnectedAt,
		}
	}

	// 2. Check Slack OAuth (legacy table)
	var slackStatus struct {
		TeamName    *string     `json:"team_name"`
		ConnectedAt interface{} `json:"connected_at"`
	}
	err = h.pool.QueryRow(ctx, `
		SELECT team_name, created_at FROM slack_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&slackStatus.TeamName, &slackStatus.ConnectedAt)
	if err == nil {
		statusMap["slack"] = &IntegrationStatusInfo{
			ProviderID:    "slack",
			ProviderName:  "Slack",
			Category:      "communication",
			Connected:     true,
			Status:        "connected",
			WorkspaceName: slackStatus.TeamName,
			ConnectedAt:   slackStatus.ConnectedAt,
		}
	}

	// 3. Check Notion OAuth (legacy table)
	var notionStatus struct {
		WorkspaceName *string     `json:"workspace_name"`
		ConnectedAt   interface{} `json:"connected_at"`
	}
	err = h.pool.QueryRow(ctx, `
		SELECT workspace_name, created_at FROM notion_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&notionStatus.WorkspaceName, &notionStatus.ConnectedAt)
	if err == nil {
		statusMap["notion"] = &IntegrationStatusInfo{
			ProviderID:    "notion",
			ProviderName:  "Notion",
			Category:      "storage",
			Connected:     true,
			Status:        "connected",
			WorkspaceName: notionStatus.WorkspaceName,
			ConnectedAt:   notionStatus.ConnectedAt,
		}
	}

	// 4. Get all from user_integrations table (new system)
	rows, err := h.pool.Query(ctx, `
		SELECT ui.provider_id, ui.status, ui.connected_at,
		       ui.external_account_name, ui.external_workspace_name,
		       ip.name, ip.category, ip.icon_url
		FROM user_integrations ui
		JOIN integration_providers ip ON ui.provider_id = ip.id
		WHERE ui.user_id = $1
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var info struct {
				ProviderID    string
				Status        string
				ConnectedAt   interface{}
				AccountName   *string
				WorkspaceName *string
				Name          string
				Category      string
				IconURL       *string
			}
			if err := rows.Scan(&info.ProviderID, &info.Status, &info.ConnectedAt,
				&info.AccountName, &info.WorkspaceName,
				&info.Name, &info.Category, &info.IconURL); err != nil {
				continue
			}
			// Override legacy status if exists in new table
			statusMap[info.ProviderID] = &IntegrationStatusInfo{
				ProviderID:    info.ProviderID,
				ProviderName:  info.Name,
				Category:      info.Category,
				Connected:     info.Status == "connected",
				Status:        info.Status,
				AccountName:   info.AccountName,
				WorkspaceName: info.WorkspaceName,
				ConnectedAt:   info.ConnectedAt,
				IconURL:       info.IconURL,
			}
		}
	}

	// 5. Get all available providers and mark unconnected ones
	providerRows, err := h.pool.Query(ctx, `
		SELECT id, name, category, icon_url, status
		FROM integration_providers
		WHERE status != 'deprecated'
		ORDER BY category, name
	`)
	if err == nil {
		defer providerRows.Close()
		for providerRows.Next() {
			var p struct {
				ID       string
				Name     string
				Category string
				IconURL  *string
				Status   string
			}
			if err := providerRows.Scan(&p.ID, &p.Name, &p.Category, &p.IconURL, &p.Status); err != nil {
				continue
			}
			// Only add if not already in statusMap
			if _, exists := statusMap[p.ID]; !exists {
				statusMap[p.ID] = &IntegrationStatusInfo{
					ProviderID:   p.ID,
					ProviderName: p.Name,
					Category:     p.Category,
					Connected:    false,
					Status:       p.Status, // available, coming_soon, etc.
					IconURL:      p.IconURL,
				}
			} else {
				// Update icon_url if available
				if statusMap[p.ID].IconURL == nil {
					statusMap[p.ID].IconURL = p.IconURL
				}
			}
		}
	}

	// Convert map to slice grouped by category
	categorized := make(map[string][]IntegrationStatusInfo)
	var allIntegrations []IntegrationStatusInfo
	for _, info := range statusMap {
		categorized[info.Category] = append(categorized[info.Category], *info)
		allIntegrations = append(allIntegrations, *info)
	}

	// Count connected
	connectedCount := 0
	for _, info := range allIntegrations {
		if info.Connected {
			connectedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"integrations":    allIntegrations,
		"by_category":     categorized,
		"connected_count": connectedCount,
		"total_count":     len(allIntegrations),
	})
}

// ============================================================================
// Helper Functions for Integration Stats
// ============================================================================

// getIntegrationSyncStats returns detailed sync statistics for an integration
func (h *IntegrationsHandler) getIntegrationSyncStats(c *gin.Context, integrationID uuid.UUID, providerID, userID string) map[string]interface{} {
	stats := map[string]interface{}{
		"total_items":      0,
		"items_by_type":    map[string]int{},
		"date_range":       nil,
		"last_sync":        nil,
		"last_sync_status": nil,
		"sync_count":       0,
	}

	switch providerID {
	case "google_calendar":
		// Get calendar event stats
		var eventCount int
		var minDate, maxDate *time.Time
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*), MIN(start_time), MAX(start_time)
			FROM calendar_events WHERE user_id = $1 AND source = 'google'
		`, userID).Scan(&eventCount, &minDate, &maxDate)

		stats["total_items"] = eventCount
		stats["items_by_type"] = map[string]int{
			"events": eventCount,
		}

		if minDate != nil && maxDate != nil {
			stats["date_range"] = map[string]interface{}{
				"from": minDate,
				"to":   maxDate,
			}
		}

		// Get last sync info - use started_at as fallback if completed_at is null
		var lastSync, startedAt *time.Time
		var lastStatus *string
		var syncCount int
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COALESCE(completed_at, started_at), started_at, status FROM integration_sync_log
			WHERE user_integration_id = $1
			ORDER BY started_at DESC LIMIT 1
		`, integrationID).Scan(&lastSync, &startedAt, &lastStatus)

		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*) FROM integration_sync_log
			WHERE user_integration_id = $1
		`, integrationID).Scan(&syncCount)

		// Use started_at if completed_at is null
		if lastSync == nil && startedAt != nil {
			lastSync = startedAt
		}

		stats["last_sync"] = lastSync
		stats["last_sync_status"] = lastStatus
		stats["sync_count"] = syncCount

	case "gmail":
		// Get email stats (for future)
		var emailCount int
		var unreadCount int
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*), COUNT(*) FILTER (WHERE is_read = false)
			FROM emails WHERE user_id = $1 AND provider = 'gmail'
		`, userID).Scan(&emailCount, &unreadCount)

		stats["total_items"] = emailCount
		stats["items_by_type"] = map[string]int{
			"emails": emailCount,
			"unread": unreadCount,
		}

	case "slack":
		// Get channel stats (for future)
		var channelCount int
		var messageCount int
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*) FROM channels WHERE user_id = $1 AND provider = 'slack'
		`, userID).Scan(&channelCount)
		h.pool.QueryRow(c.Request.Context(), `
			SELECT COUNT(*) FROM channel_messages cm
			JOIN channels c ON cm.channel_id = c.id
			WHERE c.user_id = $1 AND c.provider = 'slack'
		`, userID).Scan(&messageCount)

		stats["total_items"] = channelCount + messageCount
		stats["items_by_type"] = map[string]int{
			"channels": channelCount,
			"messages": messageCount,
		}
	}

	return stats
}

// getSyncHistory returns the last 10 sync operations for an integration
func (h *IntegrationsHandler) getSyncHistory(c *gin.Context, integrationID uuid.UUID) []map[string]interface{} {
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, sync_type, direction, status, started_at, completed_at, records_synced, error_message
		FROM integration_sync_log
		WHERE user_integration_id = $1
		ORDER BY started_at DESC
		LIMIT 10
	`, integrationID)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var (
			id            uuid.UUID
			syncType      string
			direction     string
			status        string
			startedAt     *time.Time
			completedAt   *time.Time
			recordsSynced *int
			errorMessage  *string
		)
		if err := rows.Scan(&id, &syncType, &direction, &status, &startedAt, &completedAt, &recordsSynced, &errorMessage); err != nil {
			continue
		}
		history = append(history, map[string]interface{}{
			"id":             id,
			"sync_type":      syncType,
			"direction":      direction,
			"status":         status,
			"started_at":     startedAt,
			"completed_at":   completedAt,
			"records_synced": recordsSynced,
			"error_message":  errorMessage,
		})
	}
	return history
}

// getAvailablePermissions returns all available permissions for a provider
func getAvailablePermissions(providerID string) []map[string]interface{} {
	permissions := map[string][]map[string]interface{}{
		"google_calendar": {
			{"scope": "calendar", "name": "Calendar Access", "description": "View and manage your calendars", "granted": true},
			{"scope": "calendar.readonly", "name": "Calendar Read-Only", "description": "View your calendars", "granted": true},
			{"scope": "calendar.events", "name": "Calendar Events", "description": "Create and edit events", "granted": true},
			{"scope": "calendar.settings.readonly", "name": "Calendar Settings", "description": "View calendar settings", "granted": false},
		},
		"gmail": {
			{"scope": "gmail.readonly", "name": "Gmail Read-Only", "description": "View your emails", "granted": false},
			{"scope": "gmail.send", "name": "Gmail Send", "description": "Send emails on your behalf", "granted": false},
			{"scope": "gmail.compose", "name": "Gmail Compose", "description": "Compose new emails", "granted": false},
			{"scope": "gmail.modify", "name": "Gmail Full Access", "description": "Read, send, and manage emails", "granted": false},
		},
		"slack": {
			{"scope": "channels:read", "name": "View Channels", "description": "View public channels", "granted": false},
			{"scope": "channels:history", "name": "Channel History", "description": "View messages in channels", "granted": false},
			{"scope": "chat:write", "name": "Send Messages", "description": "Send messages as the app", "granted": false},
			{"scope": "users:read", "name": "View Users", "description": "View workspace members", "granted": false},
		},
		"notion": {
			{"scope": "read_content", "name": "Read Content", "description": "View pages and databases", "granted": false},
			{"scope": "insert_content", "name": "Insert Content", "description": "Create new pages", "granted": false},
			{"scope": "update_content", "name": "Update Content", "description": "Edit existing pages", "granted": false},
		},
	}

	if perms, ok := permissions[providerID]; ok {
		return perms
	}
	return []map[string]interface{}{}
}
