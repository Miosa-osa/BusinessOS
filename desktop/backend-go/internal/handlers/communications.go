package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// CommunicationsHandler exposes the workspace-managed overlay for provider data.
// It never changes the original Gmail, Slack, or calendar source records.
type CommunicationsHandler struct {
	pool *pgxpool.Pool
}

func NewCommunicationsHandler(pool *pgxpool.Pool) *CommunicationsHandler {
	return &CommunicationsHandler{pool: pool}
}

func RegisterCommunicationsRoutes(api *gin.RouterGroup, h *CommunicationsHandler, auth gin.HandlerFunc) {
	communications := api.Group("/communications")
	communications.Use(auth, middleware.RequireAuth())
	{
		communications.GET("/workspaces/:workspaceId/overview", h.GetWorkspaceOverview)
		communications.PUT("/workspaces/:workspaceId/items", h.UpsertWorkspaceItem)
		communications.DELETE("/workspaces/:workspaceId/items/:itemId", h.RemoveWorkspaceItem)
	}
}

type communicationItem struct {
	ID               string     `json:"id,omitempty"`
	Provider         string     `json:"provider"`
	SourceKind       string     `json:"source_kind"`
	SourceExternalID string     `json:"source_external_id"`
	Title            string     `json:"title"`
	Preview          string     `json:"preview,omitempty"`
	Participant      string     `json:"participant,omitempty"`
	ParticipantEmail string     `json:"participant_email,omitempty"`
	OccurredAt       time.Time  `json:"occurred_at"`
	State            string     `json:"state,omitempty"`
	AssignedTo       string     `json:"assigned_to,omitempty"`
	WaitingOn        string     `json:"waiting_on,omitempty"`
	ProjectID        *string    `json:"project_id,omitempty"`
	ClientID         *string    `json:"client_id,omitempty"`
	ContextID        *string    `json:"context_id,omitempty"`
	SnoozedUntil     *time.Time `json:"snoozed_until,omitempty"`
}

type communicationOverview struct {
	Items           []communicationItem `json:"items"`
	UnroutedItems   []communicationItem `json:"unrouted_items"`
	ConnectionState map[string]bool     `json:"connection_state"`
}

type workspaceItemRequest struct {
	Provider         string     `json:"provider" binding:"required"`
	SourceKind       string     `json:"source_kind" binding:"required"`
	SourceExternalID string     `json:"source_external_id" binding:"required"`
	State            string     `json:"state"`
	AssignedTo       string     `json:"assigned_to"`
	WaitingOn        string     `json:"waiting_on"`
	ProjectID        *string    `json:"project_id"`
	ClientID         *string    `json:"client_id"`
	ContextID        *string    `json:"context_id"`
	SnoozedUntil     *time.Time `json:"snoozed_until"`
}

var validCommunicationStates = map[string]bool{
	"triage": true, "waiting": true, "delegated": true,
	"scheduled": true, "resolved": true, "archived": true,
}

func (h *CommunicationsHandler) GetWorkspaceOverview(c *gin.Context) {
	userID, workspaceID, ok := h.authorizeWorkspace(c)
	if !ok {
		return
	}

	items, err := h.loadItems(c, userID, workspaceID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load workspace communications"})
		return
	}

	unroutedItems, err := h.loadItems(c, userID, workspaceID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load unrouted communications"})
		return
	}

	c.JSON(http.StatusOK, communicationOverview{
		Items:         items,
		UnroutedItems: unroutedItems,
		ConnectionState: map[string]bool{
			"gmail":    h.hasGmail(c, userID),
			"slack":    h.hasSlack(c, userID),
			"calendar": h.hasCalendar(c, userID),
			"whatsapp": false,
		},
	})
}

func (h *CommunicationsHandler) UpsertWorkspaceItem(c *gin.Context) {
	userID, workspaceID, ok := h.authorizeWorkspace(c)
	if !ok {
		return
	}

	var req workspaceItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid communication item"})
		return
	}
	if req.State == "" {
		req.State = "triage"
	}
	if !validCommunicationStates[req.State] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid communication state"})
		return
	}

	var itemID uuid.UUID
	err := h.pool.QueryRow(c.Request.Context(), `
		INSERT INTO communication_workspace_items (
			workspace_id, owner_user_id, provider, source_kind, source_external_id,
			state, assigned_to, waiting_on, project_id, client_id, context_id, snoozed_until
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11, $12)
		ON CONFLICT (owner_user_id, provider, source_kind, source_external_id)
		DO UPDATE SET
			workspace_id = EXCLUDED.workspace_id,
			state = EXCLUDED.state,
			assigned_to = EXCLUDED.assigned_to,
			waiting_on = EXCLUDED.waiting_on,
			project_id = EXCLUDED.project_id,
			client_id = EXCLUDED.client_id,
			context_id = EXCLUDED.context_id,
			snoozed_until = EXCLUDED.snoozed_until,
			updated_at = NOW()
		RETURNING id
	`, workspaceID, userID, req.Provider, req.SourceKind, req.SourceExternalID, req.State,
		req.AssignedTo, req.WaitingOn, req.ProjectID, req.ClientID, req.ContextID, req.SnoozedUntil).Scan(&itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save communication routing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": itemID, "state": req.State})
}

func (h *CommunicationsHandler) RemoveWorkspaceItem(c *gin.Context) {
	userID, workspaceID, ok := h.authorizeWorkspace(c)
	if !ok {
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid communication item ID"})
		return
	}

	result, err := h.pool.Exec(c.Request.Context(), `
		DELETE FROM communication_workspace_items
		WHERE id = $1 AND workspace_id = $2 AND owner_user_id = $3
	`, itemID, workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove workspace routing"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Communication routing not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CommunicationsHandler) authorizeWorkspace(c *gin.Context) (string, uuid.UUID, bool) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return "", uuid.Nil, false
	}

	workspaceID, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return "", uuid.Nil, false
	}

	var isMember bool
	err = h.pool.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1 FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2 AND status = 'active'
		)
	`, workspaceID, user.ID).Scan(&isMember)
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return "", uuid.Nil, false
	}

	return user.ID, workspaceID, true
}

func (h *CommunicationsHandler) loadItems(c *gin.Context, userID string, workspaceID uuid.UUID, routed bool) ([]communicationItem, error) {
	items := make([]communicationItem, 0)
	queries := make([]string, 0, 3)
	if h.tableExists(c, "emails") {
		queries = append(queries, `SELECT cwi.id, 'gmail', 'email', e.external_id, COALESCE(e.subject, '(No subject)'),
			COALESCE(e.snippet, ''), COALESCE(e.from_name, ''), COALESCE(e.from_email, ''),
			COALESCE(e.date, e.received_at, e.created_at), cwi.state, cwi.assigned_to, cwi.waiting_on,
			cwi.project_id, cwi.client_id, cwi.context_id, cwi.snoozed_until
		FROM emails e
		LEFT JOIN communication_workspace_items cwi
			ON cwi.owner_user_id = e.user_id AND cwi.provider = 'gmail'
			AND cwi.source_kind = 'email' AND cwi.source_external_id = e.external_id
		WHERE e.user_id = $1 AND e.is_trash = FALSE AND %s`)
	}
	if h.tableExists(c, "calendar_events") {
		queries = append(queries, `SELECT cwi.id, 'calendar', 'meeting', COALESCE(e.google_event_id, e.id::text),
			COALESCE(e.title, '(Untitled meeting)'), COALESCE(e.location, ''), '', '', e.start_time,
			cwi.state, cwi.assigned_to, cwi.waiting_on, cwi.project_id, cwi.client_id, cwi.context_id, cwi.snoozed_until
		FROM calendar_events e
		LEFT JOIN communication_workspace_items cwi
			ON cwi.owner_user_id = e.user_id AND cwi.provider = 'calendar'
			AND cwi.source_kind = 'meeting' AND cwi.source_external_id = COALESCE(e.google_event_id, e.id::text)
		WHERE e.user_id = $1 AND e.status != 'cancelled' AND %s`)
	}
	if h.tableExists(c, "channel_messages") && h.tableExists(c, "channels") {
		queries = append(queries, `SELECT cwi.id, 'slack', 'channel_message', c.external_id || ':' || m.external_id,
			'#' || c.name, COALESCE(m.content, ''), COALESCE(m.sender_name, ''), '',
			COALESCE(m.sent_at, m.created_at), cwi.state, cwi.assigned_to, cwi.waiting_on,
			cwi.project_id, cwi.client_id, cwi.context_id, cwi.snoozed_until
		FROM channel_messages m
		JOIN channels c ON c.id = m.channel_id
		LEFT JOIN communication_workspace_items cwi
			ON cwi.owner_user_id = m.user_id AND cwi.provider = 'slack'
			AND cwi.source_kind = 'channel_message' AND cwi.source_external_id = c.external_id || ':' || m.external_id
		WHERE m.user_id = $1 AND c.provider = 'slack' AND %s`)
	}

	condition := "cwi.workspace_id = $2"
	if !routed {
		condition = "cwi.id IS NULL"
	}

	for _, query := range queries {
		rows, err := h.pool.Query(c.Request.Context(), fmt.Sprintf(query, condition), userID, workspaceID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var item communicationItem
			var itemID, state, assignedTo, waitingOn *string
			var projectID, clientID, contextID *uuid.UUID
			if err := rows.Scan(&itemID, &item.Provider, &item.SourceKind, &item.SourceExternalID, &item.Title,
				&item.Preview, &item.Participant, &item.ParticipantEmail, &item.OccurredAt, &state, &assignedTo,
				&waitingOn, &projectID, &clientID, &contextID, &item.SnoozedUntil); err != nil {
				rows.Close()
				return nil, err
			}
			if itemID != nil {
				item.ID = *itemID
			}
			if state != nil {
				item.State = *state
			}
			if assignedTo != nil {
				item.AssignedTo = *assignedTo
			}
			if waitingOn != nil {
				item.WaitingOn = *waitingOn
			}
			if projectID != nil {
				value := projectID.String()
				item.ProjectID = &value
			}
			if clientID != nil {
				value := clientID.String()
				item.ClientID = &value
			}
			if contextID != nil {
				value := contextID.String()
				item.ContextID = &value
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	sort.Slice(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
	if len(items) > 150 {
		items = items[:150]
	}
	return items, nil
}

func (h *CommunicationsHandler) hasGmail(c *gin.Context, userID string) bool {
	if !h.tableExists(c, "emails") {
		return false
	}
	var exists bool
	_ = h.pool.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM emails WHERE user_id = $1 AND provider = 'gmail')`, userID).Scan(&exists)
	return exists
}

func (h *CommunicationsHandler) hasSlack(c *gin.Context, userID string) bool {
	if !h.tableExists(c, "channels") {
		return false
	}
	var exists bool
	_ = h.pool.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM channels WHERE user_id = $1 AND provider = 'slack')`, userID).Scan(&exists)
	return exists
}

func (h *CommunicationsHandler) hasCalendar(c *gin.Context, userID string) bool {
	if !h.tableExists(c, "calendar_events") {
		return false
	}
	var exists bool
	_ = h.pool.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM calendar_events WHERE user_id = $1)`, userID).Scan(&exists)
	return exists
}

func (h *CommunicationsHandler) tableExists(c *gin.Context, table string) bool {
	var exists bool
	if err := h.pool.QueryRow(c.Request.Context(), `SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&exists); err != nil {
		return false
	}
	return exists
}
