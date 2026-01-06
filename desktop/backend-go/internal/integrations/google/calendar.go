package google

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarEvent represents a calendar event.
type CalendarEvent struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	GoogleID     string     `json:"google_id,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Location     string     `json:"location,omitempty"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	AllDay       bool       `json:"all_day"`
	Status       string     `json:"status"`
	MeetingLink  string     `json:"meeting_link,omitempty"`
	MeetingType  string     `json:"meeting_type,omitempty"`
	Attendees    []Attendee `json:"attendees,omitempty"`
	Recurrence   string     `json:"recurrence,omitempty"`
	ColorID      string     `json:"color_id,omitempty"`
	Source       string     `json:"source"` // "local" or "google"
	CalendarID   string     `json:"calendar_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Attendee represents an event attendee.
type Attendee struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"` // "accepted", "declined", "tentative", "needsAction"
	Optional bool   `json:"optional,omitempty"`
}

// CalendarService handles Google Calendar operations.
type CalendarService struct {
	provider *Provider
}

// NewCalendarService creates a new calendar service.
func NewCalendarService(provider *Provider) *CalendarService {
	return &CalendarService{provider: provider}
}

// GetCalendarAPI returns a Google Calendar API service for a user.
func (s *CalendarService) GetCalendarAPI(ctx context.Context, userID string) (*calendar.Service, error) {
	tokenSource, err := s.provider.GetTokenSource(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token source: %w", err)
	}

	srv, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return srv, nil
}

// SyncEvents syncs calendar events from Google Calendar.
func (s *CalendarService) SyncEvents(ctx context.Context, userID string, timeMin, timeMax time.Time) (*SyncEventsResult, error) {
	log.Printf("Calendar sync starting for user %s", userID)

	srv, err := s.GetCalendarAPI(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar API: %w", err)
	}

	// Get events from primary calendar
	events, err := srv.Events.List("primary").
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(250).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	result := &SyncEventsResult{
		TotalEvents: len(events.Items),
	}

	for _, event := range events.Items {
		if err := s.saveEvent(ctx, userID, event); err != nil {
			log.Printf("Failed to save event %s: %v", event.Id, err)
			result.FailedEvents++
		} else {
			result.SyncedEvents++
		}
	}

	log.Printf("Calendar sync complete for user %s: synced %d/%d events",
		userID, result.SyncedEvents, result.TotalEvents)

	return result, nil
}

// SyncEventsResult represents the result of a sync operation.
type SyncEventsResult struct {
	TotalEvents  int `json:"total_events"`
	SyncedEvents int `json:"synced_events"`
	FailedEvents int `json:"failed_events"`
}

// saveEvent saves a Google Calendar event to the database.
func (s *CalendarService) saveEvent(ctx context.Context, userID string, event *calendar.Event) error {
	// Parse start and end times
	var startTime, endTime time.Time
	var allDay bool

	if event.Start.DateTime != "" {
		startTime, _ = time.Parse(time.RFC3339, event.Start.DateTime)
		endTime, _ = time.Parse(time.RFC3339, event.End.DateTime)
	} else {
		// All-day event
		startTime, _ = time.Parse("2006-01-02", event.Start.Date)
		endTime, _ = time.Parse("2006-01-02", event.End.Date)
		allDay = true
	}

	// Extract meeting link
	meetingLink := ""
	meetingType := ""
	if event.ConferenceData != nil && len(event.ConferenceData.EntryPoints) > 0 {
		for _, ep := range event.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				meetingLink = ep.Uri
				if event.ConferenceData.ConferenceSolution != nil {
					meetingType = event.ConferenceData.ConferenceSolution.Name
				}
				break
			}
		}
	}

	// Get status
	status := event.Status
	if status == "" {
		status = "confirmed"
	}

	// Insert or update event
	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO calendar_events (
			user_id, google_event_id, title, description, location,
			start_time, end_time, all_day, status,
			meeting_link, meeting_type, color_id, source, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'google', NOW())
		ON CONFLICT (user_id, google_event_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			location = EXCLUDED.location,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			all_day = EXCLUDED.all_day,
			status = EXCLUDED.status,
			meeting_link = EXCLUDED.meeting_link,
			meeting_type = EXCLUDED.meeting_type,
			color_id = EXCLUDED.color_id,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, event.Id, event.Summary, event.Description, event.Location,
		startTime, endTime, allDay, status,
		meetingLink, meetingType, event.ColorId)

	return err
}

// GetEvents retrieves calendar events for a user.
func (s *CalendarService) GetEvents(ctx context.Context, userID string, start, end time.Time) ([]*CalendarEvent, error) {
	rows, err := s.provider.Pool().Query(ctx, `
		SELECT id, user_id, google_event_id, title, description, location,
			start_time, end_time, all_day, status,
			meeting_link, meeting_type, color_id, source,
			created_at, updated_at
		FROM calendar_events
		WHERE user_id = $1 AND start_time >= $2 AND start_time < $3
		ORDER BY start_time
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		var googleID, description, location, meetingLink, meetingType, colorID pgtype.Text

		err := rows.Scan(
			&e.ID, &e.UserID, &googleID, &e.Title, &description, &location,
			&e.StartTime, &e.EndTime, &e.AllDay, &e.Status,
			&meetingLink, &meetingType, &colorID, &e.Source,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		e.GoogleID = googleID.String
		e.Description = description.String
		e.Location = location.String
		e.MeetingLink = meetingLink.String
		e.MeetingType = meetingType.String
		e.ColorID = colorID.String

		events = append(events, &e)
	}

	return events, nil
}

// CreateEvent creates a new calendar event.
func (s *CalendarService) CreateEvent(ctx context.Context, userID string, event *CalendarEvent) (*CalendarEvent, error) {
	// Insert into database
	var id string
	err := s.provider.Pool().QueryRow(ctx, `
		INSERT INTO calendar_events (
			user_id, title, description, location,
			start_time, end_time, all_day, status,
			meeting_link, meeting_type, source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'local')
		RETURNING id
	`, userID, event.Title, event.Description, event.Location,
		event.StartTime, event.EndTime, event.AllDay, "confirmed",
		event.MeetingLink, event.MeetingType).Scan(&id)

	if err != nil {
		return nil, err
	}

	event.ID = id
	event.UserID = userID
	event.Source = "local"
	event.Status = "confirmed"

	// Optionally push to Google Calendar
	if event.GoogleID == "" {
		go s.pushEventToGoogle(context.Background(), userID, event)
	}

	return event, nil
}

// pushEventToGoogle pushes a local event to Google Calendar.
func (s *CalendarService) pushEventToGoogle(ctx context.Context, userID string, event *CalendarEvent) {
	srv, err := s.GetCalendarAPI(ctx, userID)
	if err != nil {
		log.Printf("Failed to get calendar API for push: %v", err)
		return
	}

	googleEvent := &calendar.Event{
		Summary:     event.Title,
		Description: event.Description,
		Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		},
	}

	if event.AllDay {
		googleEvent.Start = &calendar.EventDateTime{Date: event.StartTime.Format("2006-01-02")}
		googleEvent.End = &calendar.EventDateTime{Date: event.EndTime.Format("2006-01-02")}
	}

	created, err := srv.Events.Insert("primary", googleEvent).Do()
	if err != nil {
		log.Printf("Failed to push event to Google: %v", err)
		return
	}

	// Update local event with Google ID
	s.provider.Pool().Exec(ctx, `
		UPDATE calendar_events SET google_event_id = $1, updated_at = NOW()
		WHERE id = $2
	`, created.Id, event.ID)
}

// DeleteEvent deletes a calendar event.
func (s *CalendarService) DeleteEvent(ctx context.Context, userID, eventID string) error {
	// Get the event to check for Google ID
	var googleID pgtype.Text
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT google_event_id FROM calendar_events WHERE id = $1 AND user_id = $2
	`, eventID, userID).Scan(&googleID)

	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Delete from database
	_, err = s.provider.Pool().Exec(ctx, `
		DELETE FROM calendar_events WHERE id = $1 AND user_id = $2
	`, eventID, userID)
	if err != nil {
		return err
	}

	// Delete from Google if it was synced
	if googleID.Valid && googleID.String != "" {
		go func() {
			srv, err := s.GetCalendarAPI(context.Background(), userID)
			if err != nil {
				log.Printf("Failed to get calendar API for delete: %v", err)
				return
			}
			srv.Events.Delete("primary", googleID.String).Do()
		}()
	}

	return nil
}

// IsConnected checks if the user has Google Calendar connected.
func (s *CalendarService) IsConnected(ctx context.Context, userID string) bool {
	var count int
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM google_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&count)
	return err == nil && count > 0
}

// ============================================
// MCP-Compatible Methods (Google Native Types)
// These methods work with Google's native calendar.Event type
// for advanced features like attendees, recurrence, and Meet links.
// ============================================

// FetchEvents returns events from Google Calendar (native Google types).
// Used by MCP tools that need full Google Calendar features.
func (s *CalendarService) FetchEvents(ctx context.Context, userID string, start, end time.Time) ([]*calendar.Event, error) {
	srv, err := s.GetCalendarAPI(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar API: %w", err)
	}

	events, err := srv.Events.List("primary").
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(100).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events.Items, nil
}

// CreateGoogleEvent creates an event using Google's native event type.
// Supports all Google Calendar features: attendees, Meet links, recurrence.
func (s *CalendarService) CreateGoogleEvent(ctx context.Context, userID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := s.GetCalendarAPI(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar API: %w", err)
	}

	// If event has conference data request, add conferenceDataVersion
	call := srv.Events.Insert("primary", event)
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		call = call.ConferenceDataVersion(1)
	}

	created, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return created, nil
}

// UpdateGoogleEvent updates an event using Google's native event type.
func (s *CalendarService) UpdateGoogleEvent(ctx context.Context, userID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := s.GetCalendarAPI(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar API: %w", err)
	}

	call := srv.Events.Update("primary", eventID, event)
	if event.ConferenceData != nil && event.ConferenceData.CreateRequest != nil {
		call = call.ConferenceDataVersion(1)
	}

	updated, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return updated, nil
}
