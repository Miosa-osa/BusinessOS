// Package services provides business logic for BusinessOS.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =============================================================================
// SYNC SERVICE
// Handles live data synchronization from external providers
// =============================================================================

// SyncService handles syncing data from external providers to BusinessOS.
type SyncService struct {
	pool     *pgxpool.Pool
	logger   *slog.Logger
	eventBus *BuildEventBus
}

// NewSyncService creates a new sync service.
func NewSyncService(pool *pgxpool.Pool, logger *slog.Logger, eventBus *BuildEventBus) *SyncService {
	return &SyncService{
		pool:     pool,
		logger:   logger,
		eventBus: eventBus,
	}
}

// =============================================================================
// CALENDAR EVENTS
// =============================================================================

// SyncedCalendarEvent represents a calendar event synced from an external provider.
type SyncedCalendarEvent struct {
	ID               uuid.UUID              `json:"id"`
	UserID           uuid.UUID              `json:"user_id"`
	WorkspaceID      *uuid.UUID             `json:"workspace_id,omitempty"`
	Provider         string                 `json:"provider"`
	ExternalID       string                 `json:"external_id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description,omitempty"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	AllDay           bool                   `json:"all_day"`
	Location         string                 `json:"location,omitempty"`
	Attendees        []CalendarAttendee     `json:"attendees,omitempty"`
	OrganizerEmail   string                 `json:"organizer_email,omitempty"`
	MeetingLink      string                 `json:"meeting_link,omitempty"`
	RecurringEventID string                 `json:"recurring_event_id,omitempty"`
	RawData          map[string]interface{} `json:"raw_data,omitempty"`
	SyncedAt         time.Time              `json:"synced_at"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CalendarAttendee represents an attendee of a calendar event.
type CalendarAttendee struct {
	Email          string `json:"email"`
	Name           string `json:"name,omitempty"`
	ResponseStatus string `json:"response_status,omitempty"` // accepted, declined, tentative, needsAction
}

// UpsertCalendarEvent creates or updates a synced calendar event.
func (s *SyncService) UpsertCalendarEvent(ctx context.Context, event SyncedCalendarEvent) (*SyncedCalendarEvent, error) {
	attendeesJSON, err := json.Marshal(event.Attendees)
	if err != nil {
		return nil, fmt.Errorf("marshal attendees: %w", err)
	}

	rawDataJSON, err := json.Marshal(event.RawData)
	if err != nil {
		rawDataJSON = []byte("{}")
	}

	var result SyncedCalendarEvent
	err = s.pool.QueryRow(ctx, `
		INSERT INTO synced_calendar_events (
			user_id, workspace_id, provider, external_id,
			title, description, start_time, end_time, all_day,
			location, attendees, organizer_email, meeting_link,
			recurring_event_id, raw_data, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			all_day = EXCLUDED.all_day,
			location = EXCLUDED.location,
			attendees = EXCLUDED.attendees,
			organizer_email = EXCLUDED.organizer_email,
			meeting_link = EXCLUDED.meeting_link,
			recurring_event_id = EXCLUDED.recurring_event_id,
			raw_data = EXCLUDED.raw_data,
			synced_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, provider, external_id, title, start_time, end_time, synced_at, created_at, updated_at
	`,
		event.UserID, event.WorkspaceID, event.Provider, event.ExternalID,
		event.Title, event.Description, event.StartTime, event.EndTime, event.AllDay,
		event.Location, attendeesJSON, event.OrganizerEmail, event.MeetingLink,
		event.RecurringEventID, rawDataJSON,
	).Scan(
		&result.ID, &result.UserID, &result.Provider, &result.ExternalID,
		&result.Title, &result.StartTime, &result.EndTime, &result.SyncedAt,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to upsert calendar event",
			slog.String("provider", event.Provider),
			slog.String("external_id", event.ExternalID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("upsert calendar event: %w", err)
	}

	s.logger.Info("Calendar event synced",
		slog.String("provider", event.Provider),
		slog.String("external_id", event.ExternalID),
		slog.String("title", event.Title),
	)

	// Publish sync event
	s.publishSyncEvent(event.UserID, event.WorkspaceID, "calendar_event", result.ID.String(), "synced")

	return &result, nil
}

// GetCalendarEvents retrieves synced calendar events for a user.
func (s *SyncService) GetCalendarEvents(ctx context.Context, userID uuid.UUID, provider string, startTime, endTime time.Time) ([]SyncedCalendarEvent, error) {
	query := `
		SELECT id, user_id, workspace_id, provider, external_id,
			   title, description, start_time, end_time, all_day,
			   location, attendees, organizer_email, meeting_link,
			   synced_at, created_at, updated_at
		FROM synced_calendar_events
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{userID}

	if provider != "" {
		query += " AND provider = $2"
		args = append(args, provider)
	}

	if !startTime.IsZero() {
		query += fmt.Sprintf(" AND start_time >= $%d", len(args)+1)
		args = append(args, startTime)
	}

	if !endTime.IsZero() {
		query += fmt.Sprintf(" AND end_time <= $%d", len(args)+1)
		args = append(args, endTime)
	}

	query += " ORDER BY start_time ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query calendar events: %w", err)
	}
	defer rows.Close()

	var events []SyncedCalendarEvent
	for rows.Next() {
		var e SyncedCalendarEvent
		var attendeesJSON []byte
		var workspaceID *uuid.UUID

		err := rows.Scan(
			&e.ID, &e.UserID, &workspaceID, &e.Provider, &e.ExternalID,
			&e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.AllDay,
			&e.Location, &attendeesJSON, &e.OrganizerEmail, &e.MeetingLink,
			&e.SyncedAt, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan calendar event: %w", err)
		}

		e.WorkspaceID = workspaceID
		if len(attendeesJSON) > 0 {
			json.Unmarshal(attendeesJSON, &e.Attendees)
		}
		events = append(events, e)
	}

	return events, nil
}

// =============================================================================
// SYNCED TASKS (Linear, ClickUp, etc.)
// =============================================================================

// SyncedTask represents a task synced from an external provider.
type SyncedTask struct {
	ID            uuid.UUID              `json:"id"`
	UserID        uuid.UUID              `json:"user_id"`
	WorkspaceID   *uuid.UUID             `json:"workspace_id,omitempty"`
	Provider      string                 `json:"provider"`
	ExternalID    string                 `json:"external_id"`
	Identifier    string                 `json:"identifier,omitempty"` // e.g., LIN-123
	URL           string                 `json:"url,omitempty"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Status        string                 `json:"status,omitempty"`
	StatusType    string                 `json:"status_type,omitempty"` // backlog, todo, in_progress, done, canceled
	Priority      string                 `json:"priority,omitempty"`
	PriorityOrder int                    `json:"priority_order,omitempty"`
	AssigneeID    string                 `json:"assignee_id,omitempty"`
	AssigneeName  string                 `json:"assignee_name,omitempty"`
	AssigneeEmail string                 `json:"assignee_email,omitempty"`
	ProjectID     string                 `json:"project_id,omitempty"`
	ProjectName   string                 `json:"project_name,omitempty"`
	TeamID        string                 `json:"team_id,omitempty"`
	TeamName      string                 `json:"team_name,omitempty"`
	DueDate       *time.Time             `json:"due_date,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Labels        []TaskLabel            `json:"labels,omitempty"`
	Estimate      int                    `json:"estimate,omitempty"`
	ParentID      string                 `json:"parent_id,omitempty"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
	SyncedAt      time.Time              `json:"synced_at"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// TaskLabel represents a label on a synced task.
type TaskLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// UpsertTask creates or updates a synced task.
func (s *SyncService) UpsertTask(ctx context.Context, task SyncedTask) (*SyncedTask, error) {
	labelsJSON, err := json.Marshal(task.Labels)
	if err != nil {
		labelsJSON = []byte("[]")
	}

	rawDataJSON, err := json.Marshal(task.RawData)
	if err != nil {
		rawDataJSON = []byte("{}")
	}

	var result SyncedTask
	err = s.pool.QueryRow(ctx, `
		INSERT INTO synced_tasks (
			user_id, workspace_id, provider, external_id, identifier, url,
			title, description, status, status_type, priority, priority_order,
			assignee_id, assignee_name, assignee_email,
			project_id, project_name, team_id, team_name,
			due_date, started_at, completed_at, labels, estimate, parent_id, raw_data, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, NOW())
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			identifier = EXCLUDED.identifier,
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			status_type = EXCLUDED.status_type,
			priority = EXCLUDED.priority,
			priority_order = EXCLUDED.priority_order,
			assignee_id = EXCLUDED.assignee_id,
			assignee_name = EXCLUDED.assignee_name,
			assignee_email = EXCLUDED.assignee_email,
			project_id = EXCLUDED.project_id,
			project_name = EXCLUDED.project_name,
			team_id = EXCLUDED.team_id,
			team_name = EXCLUDED.team_name,
			due_date = EXCLUDED.due_date,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			labels = EXCLUDED.labels,
			estimate = EXCLUDED.estimate,
			parent_id = EXCLUDED.parent_id,
			raw_data = EXCLUDED.raw_data,
			synced_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, provider, external_id, identifier, title, status, synced_at, created_at, updated_at
	`,
		task.UserID, task.WorkspaceID, task.Provider, task.ExternalID, task.Identifier, task.URL,
		task.Title, task.Description, task.Status, task.StatusType, task.Priority, task.PriorityOrder,
		task.AssigneeID, task.AssigneeName, task.AssigneeEmail,
		task.ProjectID, task.ProjectName, task.TeamID, task.TeamName,
		task.DueDate, task.StartedAt, task.CompletedAt, labelsJSON, task.Estimate, task.ParentID, rawDataJSON,
	).Scan(
		&result.ID, &result.UserID, &result.Provider, &result.ExternalID,
		&result.Identifier, &result.Title, &result.Status,
		&result.SyncedAt, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to upsert task",
			slog.String("provider", task.Provider),
			slog.String("external_id", task.ExternalID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("upsert task: %w", err)
	}

	s.logger.Info("Task synced",
		slog.String("provider", task.Provider),
		slog.String("identifier", task.Identifier),
		slog.String("title", task.Title),
		slog.String("status", task.Status),
	)

	s.publishSyncEvent(task.UserID, task.WorkspaceID, "task", result.ID.String(), "synced")

	return &result, nil
}

// DeleteTask marks a synced task as deleted.
func (s *SyncService) DeleteTask(ctx context.Context, userID uuid.UUID, provider, externalID string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM synced_tasks
		WHERE user_id = $1 AND provider = $2 AND external_id = $3
	`, userID, provider, externalID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	s.logger.Info("Task deleted",
		slog.String("provider", provider),
		slog.String("external_id", externalID),
	)

	return nil
}

// =============================================================================
// SYNCED CONTACTS (HubSpot, etc.)
// =============================================================================

// SyncedContact represents a contact/lead synced from a CRM.
type SyncedContact struct {
	ID             uuid.UUID              `json:"id"`
	UserID         uuid.UUID              `json:"user_id"`
	WorkspaceID    *uuid.UUID             `json:"workspace_id,omitempty"`
	Provider       string                 `json:"provider"`
	ExternalID     string                 `json:"external_id"`
	Email          string                 `json:"email,omitempty"`
	FirstName      string                 `json:"first_name,omitempty"`
	LastName       string                 `json:"last_name,omitempty"`
	FullName       string                 `json:"full_name,omitempty"`
	Phone          string                 `json:"phone,omitempty"`
	Company        string                 `json:"company,omitempty"`
	CompanyID      string                 `json:"company_id,omitempty"`
	JobTitle       string                 `json:"job_title,omitempty"`
	ContactType    string                 `json:"contact_type,omitempty"` // lead, contact, customer
	LifecycleStage string                 `json:"lifecycle_stage,omitempty"`
	LeadStatus     string                 `json:"lead_status,omitempty"`
	LinkedInURL    string                 `json:"linkedin_url,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	RawData        map[string]interface{} `json:"raw_data,omitempty"`
	SyncedAt       time.Time              `json:"synced_at"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// UpsertContact creates or updates a synced contact.
func (s *SyncService) UpsertContact(ctx context.Context, contact SyncedContact) (*SyncedContact, error) {
	propertiesJSON, err := json.Marshal(contact.Properties)
	if err != nil {
		propertiesJSON = []byte("{}")
	}

	rawDataJSON, err := json.Marshal(contact.RawData)
	if err != nil {
		rawDataJSON = []byte("{}")
	}

	// Compute full name if not provided
	if contact.FullName == "" && (contact.FirstName != "" || contact.LastName != "") {
		contact.FullName = contact.FirstName
		if contact.LastName != "" {
			if contact.FullName != "" {
				contact.FullName += " "
			}
			contact.FullName += contact.LastName
		}
	}

	var result SyncedContact
	err = s.pool.QueryRow(ctx, `
		INSERT INTO synced_contacts (
			user_id, workspace_id, provider, external_id,
			email, first_name, last_name, full_name, phone,
			company, company_id, job_title,
			contact_type, lifecycle_stage, lead_status,
			linkedin_url, properties, raw_data, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			email = EXCLUDED.email,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			full_name = EXCLUDED.full_name,
			phone = EXCLUDED.phone,
			company = EXCLUDED.company,
			company_id = EXCLUDED.company_id,
			job_title = EXCLUDED.job_title,
			contact_type = EXCLUDED.contact_type,
			lifecycle_stage = EXCLUDED.lifecycle_stage,
			lead_status = EXCLUDED.lead_status,
			linkedin_url = EXCLUDED.linkedin_url,
			properties = EXCLUDED.properties,
			raw_data = EXCLUDED.raw_data,
			synced_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, provider, external_id, email, full_name, synced_at, created_at, updated_at
	`,
		contact.UserID, contact.WorkspaceID, contact.Provider, contact.ExternalID,
		contact.Email, contact.FirstName, contact.LastName, contact.FullName, contact.Phone,
		contact.Company, contact.CompanyID, contact.JobTitle,
		contact.ContactType, contact.LifecycleStage, contact.LeadStatus,
		contact.LinkedInURL, propertiesJSON, rawDataJSON,
	).Scan(
		&result.ID, &result.UserID, &result.Provider, &result.ExternalID,
		&result.Email, &result.FullName, &result.SyncedAt, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to upsert contact",
			slog.String("provider", contact.Provider),
			slog.String("external_id", contact.ExternalID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("upsert contact: %w", err)
	}

	s.logger.Info("Contact synced",
		slog.String("provider", contact.Provider),
		slog.String("email", contact.Email),
		slog.String("name", contact.FullName),
	)

	s.publishSyncEvent(contact.UserID, contact.WorkspaceID, "contact", result.ID.String(), "synced")

	return &result, nil
}

// =============================================================================
// SYNCED MESSAGES (Slack, etc.)
// =============================================================================

// SyncedMessage represents a message synced from a messaging platform.
type SyncedMessage struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	WorkspaceID     *uuid.UUID             `json:"workspace_id,omitempty"`
	Provider        string                 `json:"provider"`
	ExternalID      string                 `json:"external_id"`
	ChannelID       string                 `json:"channel_id,omitempty"`
	ChannelName     string                 `json:"channel_name,omitempty"`
	ChannelType     string                 `json:"channel_type,omitempty"` // channel, dm, group
	SenderID        string                 `json:"sender_id,omitempty"`
	SenderName      string                 `json:"sender_name,omitempty"`
	SenderAvatarURL string                 `json:"sender_avatar_url,omitempty"`
	Content         string                 `json:"content,omitempty"`
	ThreadID        string                 `json:"thread_id,omitempty"`
	IsThreadReply   bool                   `json:"is_thread_reply"`
	Attachments     []MessageAttachment    `json:"attachments,omitempty"`
	Reactions       []MessageReaction      `json:"reactions,omitempty"`
	RawData         map[string]interface{} `json:"raw_data,omitempty"`
	SentAt          *time.Time             `json:"sent_at,omitempty"`
	SyncedAt        time.Time              `json:"synced_at"`
}

// MessageAttachment represents an attachment on a message.
type MessageAttachment struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
}

// MessageReaction represents a reaction on a message.
type MessageReaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users,omitempty"`
}

// UpsertMessage creates or updates a synced message.
func (s *SyncService) UpsertMessage(ctx context.Context, message SyncedMessage) (*SyncedMessage, error) {
	attachmentsJSON, err := json.Marshal(message.Attachments)
	if err != nil {
		attachmentsJSON = []byte("[]")
	}

	reactionsJSON, err := json.Marshal(message.Reactions)
	if err != nil {
		reactionsJSON = []byte("[]")
	}

	rawDataJSON, err := json.Marshal(message.RawData)
	if err != nil {
		rawDataJSON = []byte("{}")
	}

	var result SyncedMessage
	err = s.pool.QueryRow(ctx, `
		INSERT INTO synced_messages (
			user_id, workspace_id, provider, external_id,
			channel_id, channel_name, channel_type,
			sender_id, sender_name, sender_avatar_url,
			content, thread_id, is_thread_reply,
			attachments, reactions, raw_data, sent_at, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			channel_name = EXCLUDED.channel_name,
			sender_name = EXCLUDED.sender_name,
			content = EXCLUDED.content,
			attachments = EXCLUDED.attachments,
			reactions = EXCLUDED.reactions,
			raw_data = EXCLUDED.raw_data,
			synced_at = NOW()
		RETURNING id, user_id, provider, external_id, channel_id, content, synced_at
	`,
		message.UserID, message.WorkspaceID, message.Provider, message.ExternalID,
		message.ChannelID, message.ChannelName, message.ChannelType,
		message.SenderID, message.SenderName, message.SenderAvatarURL,
		message.Content, message.ThreadID, message.IsThreadReply,
		attachmentsJSON, reactionsJSON, rawDataJSON, message.SentAt,
	).Scan(
		&result.ID, &result.UserID, &result.Provider, &result.ExternalID,
		&result.ChannelID, &result.Content, &result.SyncedAt,
	)
	if err != nil {
		s.logger.Error("Failed to upsert message",
			slog.String("provider", message.Provider),
			slog.String("external_id", message.ExternalID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("upsert message: %w", err)
	}

	s.logger.Info("Message synced",
		slog.String("provider", message.Provider),
		slog.String("channel", message.ChannelID),
	)

	s.publishSyncEvent(message.UserID, message.WorkspaceID, "message", result.ID.String(), "synced")

	return &result, nil
}

// =============================================================================
// SYNCED MEETINGS (Fathom, etc.)
// =============================================================================

// SyncedMeeting represents a meeting recording synced from a recording service.
type SyncedMeeting struct {
	ID               uuid.UUID              `json:"id"`
	UserID           uuid.UUID              `json:"user_id"`
	WorkspaceID      *uuid.UUID             `json:"workspace_id,omitempty"`
	Provider         string                 `json:"provider"`
	ExternalID       string                 `json:"external_id"`
	Title            string                 `json:"title,omitempty"`
	MeetingType      string                 `json:"meeting_type,omitempty"`
	StartTime        *time.Time             `json:"start_time,omitempty"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	DurationSeconds  int                    `json:"duration_seconds,omitempty"`
	Participants     []MeetingParticipant   `json:"participants,omitempty"`
	ParticipantCount int                    `json:"participant_count,omitempty"`
	OrganizerName    string                 `json:"organizer_name,omitempty"`
	OrganizerEmail   string                 `json:"organizer_email,omitempty"`
	Transcript       string                 `json:"transcript,omitempty"`
	Summary          string                 `json:"summary,omitempty"`
	KeyPoints        []string               `json:"key_points,omitempty"`
	ActionItems      []MeetingActionItem    `json:"action_items,omitempty"`
	RecordingURL     string                 `json:"recording_url,omitempty"`
	CalendarEventID  string                 `json:"calendar_event_id,omitempty"`
	RawData          map[string]interface{} `json:"raw_data,omitempty"`
	SyncedAt         time.Time              `json:"synced_at"`
	CreatedAt        time.Time              `json:"created_at"`
}

// MeetingParticipant represents a participant in a meeting.
type MeetingParticipant struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}

// MeetingActionItem represents an action item from a meeting.
type MeetingActionItem struct {
	Task     string `json:"task"`
	Assignee string `json:"assignee,omitempty"`
	DueDate  string `json:"due_date,omitempty"`
}

// UpsertMeeting creates or updates a synced meeting.
func (s *SyncService) UpsertMeeting(ctx context.Context, meeting SyncedMeeting) (*SyncedMeeting, error) {
	participantsJSON, err := json.Marshal(meeting.Participants)
	if err != nil {
		participantsJSON = []byte("[]")
	}

	keyPointsJSON, err := json.Marshal(meeting.KeyPoints)
	if err != nil {
		keyPointsJSON = []byte("[]")
	}

	actionItemsJSON, err := json.Marshal(meeting.ActionItems)
	if err != nil {
		actionItemsJSON = []byte("[]")
	}

	rawDataJSON, err := json.Marshal(meeting.RawData)
	if err != nil {
		rawDataJSON = []byte("{}")
	}

	var result SyncedMeeting
	err = s.pool.QueryRow(ctx, `
		INSERT INTO synced_meetings (
			user_id, workspace_id, provider, external_id,
			title, meeting_type, start_time, end_time, duration_seconds,
			participants, participant_count, organizer_name, organizer_email,
			transcript, summary, key_points, action_items,
			recording_url, calendar_event_id, raw_data, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW())
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			title = EXCLUDED.title,
			meeting_type = EXCLUDED.meeting_type,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			duration_seconds = EXCLUDED.duration_seconds,
			participants = EXCLUDED.participants,
			participant_count = EXCLUDED.participant_count,
			transcript = EXCLUDED.transcript,
			summary = EXCLUDED.summary,
			key_points = EXCLUDED.key_points,
			action_items = EXCLUDED.action_items,
			recording_url = EXCLUDED.recording_url,
			raw_data = EXCLUDED.raw_data,
			synced_at = NOW()
		RETURNING id, user_id, provider, external_id, title, synced_at, created_at
	`,
		meeting.UserID, meeting.WorkspaceID, meeting.Provider, meeting.ExternalID,
		meeting.Title, meeting.MeetingType, meeting.StartTime, meeting.EndTime, meeting.DurationSeconds,
		participantsJSON, meeting.ParticipantCount, meeting.OrganizerName, meeting.OrganizerEmail,
		meeting.Transcript, meeting.Summary, keyPointsJSON, actionItemsJSON,
		meeting.RecordingURL, meeting.CalendarEventID, rawDataJSON,
	).Scan(
		&result.ID, &result.UserID, &result.Provider, &result.ExternalID,
		&result.Title, &result.SyncedAt, &result.CreatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to upsert meeting",
			slog.String("provider", meeting.Provider),
			slog.String("external_id", meeting.ExternalID),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("upsert meeting: %w", err)
	}

	s.logger.Info("Meeting synced",
		slog.String("provider", meeting.Provider),
		slog.String("title", meeting.Title),
	)

	s.publishSyncEvent(meeting.UserID, meeting.WorkspaceID, "meeting", result.ID.String(), "synced")

	return &result, nil
}

// =============================================================================
// SYNC TOKENS
// =============================================================================

// GetSyncToken retrieves a sync token for incremental sync.
func (s *SyncService) GetSyncToken(ctx context.Context, userID uuid.UUID, provider, resourceType string) (string, error) {
	var token string
	err := s.pool.QueryRow(ctx, `
		SELECT sync_token FROM sync_tokens
		WHERE user_id = $1 AND provider = $2 AND resource_type = $3
	`, userID, provider, resourceType).Scan(&token)

	if err == pgx.ErrNoRows {
		return "", nil // No token yet, will do full sync
	}
	if err != nil {
		return "", fmt.Errorf("get sync token: %w", err)
	}

	return token, nil
}

// SaveSyncToken stores a sync token for incremental sync.
func (s *SyncService) SaveSyncToken(ctx context.Context, userID uuid.UUID, provider, resourceType, syncToken string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sync_tokens (user_id, provider, resource_type, sync_token, last_sync_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, provider, resource_type, COALESCE(resource_id, '')) DO UPDATE SET
			sync_token = EXCLUDED.sync_token,
			last_sync_at = NOW(),
			updated_at = NOW()
	`, userID, provider, resourceType, syncToken)
	if err != nil {
		return fmt.Errorf("save sync token: %w", err)
	}

	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// publishSyncEvent publishes a sync event to the event bus for real-time UI updates.
func (s *SyncService) publishSyncEvent(userID uuid.UUID, workspaceID *uuid.UUID, entityType, entityID, action string) {
	if s.eventBus == nil {
		return
	}

	event := BuildEvent{
		ID:              uuid.New(),
		EventType:       fmt.Sprintf("sync.%s.%s", entityType, action),
		Phase:           "sync",
		ProgressPercent: 100,
		StatusMessage:   fmt.Sprintf("%s %s", entityType, action),
		Data: map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
			"user_id":     userID.String(),
			"action":      action,
		},
		Timestamp: time.Now(),
	}

	if workspaceID != nil {
		event.WorkspaceID = workspaceID
	}

	s.eventBus.Publish(event)
}

// GetUserIDByProviderTeam looks up the user ID from a provider team/workspace ID.
// This is used by webhooks to map provider identifiers to BusinessOS users.
func (s *SyncService) GetUserIDByProviderTeam(ctx context.Context, provider, teamID string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT user_id FROM oauth_tokens
		WHERE provider = $1 AND team_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`, provider, teamID).Scan(&userID)

	if err == pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("no user found for provider %s team %s", provider, teamID)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("lookup user by provider team: %w", err)
	}

	return userID, nil
}

// GetUserIDByProviderPortal looks up the user ID from a provider portal ID (HubSpot).
func (s *SyncService) GetUserIDByProviderPortal(ctx context.Context, provider string, portalID int64) (uuid.UUID, error) {
	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT user_id FROM oauth_tokens
		WHERE provider = $1 AND portal_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`, provider, portalID).Scan(&userID)

	if err == pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("no user found for provider %s portal %d", provider, portalID)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("lookup user by provider portal: %w", err)
	}

	return userID, nil
}
