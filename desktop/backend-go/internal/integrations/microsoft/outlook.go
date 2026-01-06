package microsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// ============================================================================
// OUTLOOK MAIL TYPES
// ============================================================================

// OutlookMessage represents an Outlook email message.
type OutlookMessage struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	MessageID        string          `json:"message_id"`
	ConversationID   string          `json:"conversation_id,omitempty"`
	Subject          string          `json:"subject"`
	BodyPreview      string          `json:"body_preview,omitempty"`
	BodyContent      string          `json:"body_content,omitempty"`
	BodyContentType  string          `json:"body_content_type,omitempty"`
	Importance       string          `json:"importance,omitempty"`
	FromEmail        string          `json:"from_email"`
	FromName         string          `json:"from_name"`
	ToRecipients     []EmailRecipient `json:"to_recipients,omitempty"`
	CcRecipients     []EmailRecipient `json:"cc_recipients,omitempty"`
	BccRecipients    []EmailRecipient `json:"bcc_recipients,omitempty"`
	IsRead           bool            `json:"is_read"`
	IsDraft          bool            `json:"is_draft"`
	HasAttachments   bool            `json:"has_attachments"`
	FolderID         string          `json:"folder_id,omitempty"`
	FolderName       string          `json:"folder_name,omitempty"`
	Categories       []string        `json:"categories,omitempty"`
	FlagStatus       string          `json:"flag_status,omitempty"`
	ReceivedDateTime time.Time       `json:"received_datetime,omitempty"`
	SentDateTime     time.Time       `json:"sent_datetime,omitempty"`
	SyncedAt         time.Time       `json:"synced_at"`
}

// EmailRecipient represents an email recipient.
type EmailRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// ============================================================================
// OUTLOOK CALENDAR TYPES
// ============================================================================

// OutlookEvent represents an Outlook calendar event.
type OutlookEvent struct {
	ID                     string           `json:"id"`
	UserID                 string           `json:"user_id"`
	EventID                string           `json:"event_id"`
	CalendarID             string           `json:"calendar_id,omitempty"`
	Subject                string           `json:"subject"`
	BodyPreview            string           `json:"body_preview,omitempty"`
	BodyContent            string           `json:"body_content,omitempty"`
	LocationDisplayName    string           `json:"location_display_name,omitempty"`
	StartDateTime          time.Time        `json:"start_datetime"`
	StartTimeZone          string           `json:"start_timezone,omitempty"`
	EndDateTime            time.Time        `json:"end_datetime"`
	EndTimeZone            string           `json:"end_timezone,omitempty"`
	IsAllDay               bool             `json:"is_all_day"`
	Attendees              []EventAttendee  `json:"attendees,omitempty"`
	OrganizerEmail         string           `json:"organizer_email,omitempty"`
	OrganizerName          string           `json:"organizer_name,omitempty"`
	IsOnlineMeeting        bool             `json:"is_online_meeting"`
	OnlineMeetingProvider  string           `json:"online_meeting_provider,omitempty"`
	OnlineMeetingURL       string           `json:"online_meeting_url,omitempty"`
	ResponseStatus         string           `json:"response_status,omitempty"`
	Importance             string           `json:"importance,omitempty"`
	ShowAs                 string           `json:"show_as,omitempty"`
	IsCancelled            bool             `json:"is_cancelled"`
	IsReminderOn           bool             `json:"is_reminder_on"`
	ReminderMinutesBefore  int              `json:"reminder_minutes_before"`
	Categories             []string         `json:"categories,omitempty"`
	CreatedDateTime        time.Time        `json:"created_datetime,omitempty"`
	LastModifiedDateTime   time.Time        `json:"last_modified_datetime,omitempty"`
	SyncedAt               time.Time        `json:"synced_at"`
}

// EventAttendee represents a calendar event attendee.
type EventAttendee struct {
	Email          string `json:"email"`
	Name           string `json:"name,omitempty"`
	Type           string `json:"type,omitempty"` // required, optional, resource
	ResponseStatus string `json:"response_status,omitempty"`
}

// ============================================================================
// OUTLOOK SERVICE
// ============================================================================

// OutlookService handles Microsoft Outlook operations (mail and calendar).
type OutlookService struct {
	provider *Provider
}

// NewOutlookService creates a new Outlook service.
func NewOutlookService(provider *Provider) *OutlookService {
	return &OutlookService{provider: provider}
}

// ============================================================================
// MAIL OPERATIONS
// ============================================================================

// SyncMessages syncs messages from Outlook.
func (s *OutlookService) SyncMessages(ctx context.Context, userID string, maxResults int) (*SyncMessagesResult, error) {
	log.Printf("Outlook mail sync starting for user %s: max %d messages", userID, maxResults)

	client, err := s.provider.GetHTTPClient(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTP client: %w", err)
	}

	result := &SyncMessagesResult{}
	skip := 0

	for {
		// Build URL
		apiURL := fmt.Sprintf("%s/me/messages?$top=%d&$skip=%d&$orderby=receivedDateTime desc&$select=id,conversationId,subject,bodyPreview,importance,from,toRecipients,ccRecipients,isRead,isDraft,hasAttachments,parentFolderId,categories,flag,receivedDateTime,sentDateTime",
			GraphAPIBase, maxResults, skip)

		resp, err := client.Get(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API error: %s", resp.Status)
		}

		var msgResp struct {
			Value    []graphMessage `json:"value"`
			NextLink string         `json:"@odata.nextLink"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		result.TotalMessages += len(msgResp.Value)

		for _, msg := range msgResp.Value {
			if err := s.saveMessage(ctx, userID, &msg); err != nil {
				log.Printf("Failed to save message %s: %v", msg.ID, err)
				result.FailedMessages++
			} else {
				result.SyncedMessages++
			}
		}

		// Check if we should continue
		if msgResp.NextLink == "" || result.TotalMessages >= maxResults {
			break
		}
		skip += len(msgResp.Value)
	}

	log.Printf("Outlook mail sync complete for user %s: synced %d/%d messages",
		userID, result.SyncedMessages, result.TotalMessages)

	return result, nil
}

// SyncMessagesResult represents the result of a message sync.
type SyncMessagesResult struct {
	TotalMessages  int `json:"total_messages"`
	SyncedMessages int `json:"synced_messages"`
	FailedMessages int `json:"failed_messages"`
}

// Graph API message structure
type graphMessage struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Subject        string `json:"subject"`
	BodyPreview    string `json:"bodyPreview"`
	Importance     string `json:"importance"`
	From           *struct {
		EmailAddress struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"emailAddress"`
	} `json:"from"`
	ToRecipients []struct {
		EmailAddress struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"emailAddress"`
	} `json:"toRecipients"`
	CcRecipients []struct {
		EmailAddress struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"emailAddress"`
	} `json:"ccRecipients"`
	IsRead         bool     `json:"isRead"`
	IsDraft        bool     `json:"isDraft"`
	HasAttachments bool     `json:"hasAttachments"`
	ParentFolderID string   `json:"parentFolderId"`
	Categories     []string `json:"categories"`
	Flag           struct {
		FlagStatus string `json:"flagStatus"`
	} `json:"flag"`
	ReceivedDateTime string `json:"receivedDateTime"`
	SentDateTime     string `json:"sentDateTime"`
}

func (s *OutlookService) saveMessage(ctx context.Context, userID string, msg *graphMessage) error {
	// Parse recipients
	toRecipients := make([]EmailRecipient, 0)
	for _, r := range msg.ToRecipients {
		toRecipients = append(toRecipients, EmailRecipient{
			Email: r.EmailAddress.Address,
			Name:  r.EmailAddress.Name,
		})
	}

	ccRecipients := make([]EmailRecipient, 0)
	for _, r := range msg.CcRecipients {
		ccRecipients = append(ccRecipients, EmailRecipient{
			Email: r.EmailAddress.Address,
			Name:  r.EmailAddress.Name,
		})
	}

	// Parse dates
	var receivedDateTime, sentDateTime *time.Time
	if msg.ReceivedDateTime != "" {
		t, _ := time.Parse(time.RFC3339, msg.ReceivedDateTime)
		receivedDateTime = &t
	}
	if msg.SentDateTime != "" {
		t, _ := time.Parse(time.RFC3339, msg.SentDateTime)
		sentDateTime = &t
	}

	// Extract from info
	var fromEmail, fromName string
	if msg.From != nil {
		fromEmail = msg.From.EmailAddress.Address
		fromName = msg.From.EmailAddress.Name
	}

	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO microsoft_mail_messages (
			user_id, message_id, conversation_id, subject, body_preview, importance,
			from_email, from_name, to_recipients, cc_recipients,
			is_read, is_draft, has_attachments, folder_id, categories, flag_status,
			received_datetime, sent_datetime, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
		ON CONFLICT (user_id, message_id) DO UPDATE SET
			conversation_id = EXCLUDED.conversation_id,
			subject = EXCLUDED.subject,
			body_preview = EXCLUDED.body_preview,
			importance = EXCLUDED.importance,
			from_email = EXCLUDED.from_email,
			from_name = EXCLUDED.from_name,
			to_recipients = EXCLUDED.to_recipients,
			cc_recipients = EXCLUDED.cc_recipients,
			is_read = EXCLUDED.is_read,
			is_draft = EXCLUDED.is_draft,
			has_attachments = EXCLUDED.has_attachments,
			folder_id = EXCLUDED.folder_id,
			categories = EXCLUDED.categories,
			flag_status = EXCLUDED.flag_status,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, msg.ID, msg.ConversationID, msg.Subject, msg.BodyPreview, msg.Importance,
		fromEmail, fromName, toRecipients, ccRecipients,
		msg.IsRead, msg.IsDraft, msg.HasAttachments, msg.ParentFolderID, msg.Categories, msg.Flag.FlagStatus,
		receivedDateTime, sentDateTime)

	return err
}

// GetMessages retrieves messages for a user.
func (s *OutlookService) GetMessages(ctx context.Context, userID string, folderID string, limit, offset int) ([]*OutlookMessage, error) {
	query := `
		SELECT id, user_id, message_id, conversation_id, subject, body_preview, importance,
			from_email, from_name, is_read, is_draft, has_attachments, folder_id, categories, flag_status,
			received_datetime, sent_datetime, synced_at
		FROM microsoft_mail_messages
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if folderID != "" {
		query += " AND folder_id = $2 ORDER BY received_datetime DESC NULLS LAST LIMIT $3 OFFSET $4"
		args = append(args, folderID, limit, offset)
	} else {
		query += " ORDER BY received_datetime DESC NULLS LAST LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	}

	rows, err := s.provider.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*OutlookMessage
	for rows.Next() {
		var m OutlookMessage
		var conversationID, bodyPreview, importance, folderID, flagStatus *string
		var categories []string
		var receivedDateTime, sentDateTime *time.Time

		err := rows.Scan(
			&m.ID, &m.UserID, &m.MessageID, &conversationID, &m.Subject, &bodyPreview, &importance,
			&m.FromEmail, &m.FromName, &m.IsRead, &m.IsDraft, &m.HasAttachments, &folderID, &categories, &flagStatus,
			&receivedDateTime, &sentDateTime, &m.SyncedAt,
		)
		if err != nil {
			return nil, err
		}

		if conversationID != nil {
			m.ConversationID = *conversationID
		}
		if bodyPreview != nil {
			m.BodyPreview = *bodyPreview
		}
		if importance != nil {
			m.Importance = *importance
		}
		if folderID != nil {
			m.FolderID = *folderID
		}
		if flagStatus != nil {
			m.FlagStatus = *flagStatus
		}
		m.Categories = categories
		if receivedDateTime != nil {
			m.ReceivedDateTime = *receivedDateTime
		}
		if sentDateTime != nil {
			m.SentDateTime = *sentDateTime
		}

		messages = append(messages, &m)
	}

	return messages, nil
}

// SendMessage sends an email via Outlook.
func (s *OutlookService) SendMessage(ctx context.Context, userID string, to []string, subject, body string, isHTML bool) error {
	client, err := s.provider.GetHTTPClient(ctx, userID)
	if err != nil {
		return err
	}

	// Build recipients
	toRecipients := make([]map[string]interface{}, 0)
	for _, email := range to {
		toRecipients = append(toRecipients, map[string]interface{}{
			"emailAddress": map[string]string{
				"address": email,
			},
		})
	}

	contentType := "text"
	if isHTML {
		contentType = "html"
	}

	message := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": contentType,
				"content":     body,
			},
			"toRecipients": toRecipients,
		},
		"saveToSentItems": true,
	}

	jsonBody, _ := json.Marshal(message)

	resp, err := client.Post(GraphAPIBase+"/me/sendMail", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	return nil
}

// ============================================================================
// CALENDAR OPERATIONS
// ============================================================================

// SyncEvents syncs calendar events from Outlook.
func (s *OutlookService) SyncEvents(ctx context.Context, userID string, timeMin, timeMax time.Time) (*SyncEventsResult, error) {
	log.Printf("Outlook calendar sync starting for user %s", userID)

	client, err := s.provider.GetHTTPClient(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTP client: %w", err)
	}

	result := &SyncEventsResult{}

	// Build URL with time filter
	apiURL := fmt.Sprintf("%s/me/calendarView?startDateTime=%s&endDateTime=%s&$top=100&$select=id,subject,bodyPreview,start,end,location,attendees,organizer,isOnlineMeeting,onlineMeeting,responseStatus,importance,showAs,isCancelled,isReminderOn,reminderMinutesBeforeStart,categories,createdDateTime,lastModifiedDateTime,isAllDay",
		GraphAPIBase, url.QueryEscape(timeMin.Format(time.RFC3339)), url.QueryEscape(timeMax.Format(time.RFC3339)))

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var eventResp struct {
		Value []graphEvent `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result.TotalEvents = len(eventResp.Value)

	for _, event := range eventResp.Value {
		if err := s.saveEvent(ctx, userID, &event); err != nil {
			log.Printf("Failed to save event %s: %v", event.ID, err)
			result.FailedEvents++
		} else {
			result.SyncedEvents++
		}
	}

	log.Printf("Outlook calendar sync complete for user %s: synced %d/%d events",
		userID, result.SyncedEvents, result.TotalEvents)

	return result, nil
}

// SyncEventsResult represents the result of an events sync.
type SyncEventsResult struct {
	TotalEvents  int `json:"total_events"`
	SyncedEvents int `json:"synced_events"`
	FailedEvents int `json:"failed_events"`
}

// Graph API event structure
type graphEvent struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	BodyPreview string `json:"bodyPreview"`
	Start       struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"end"`
	IsAllDay bool `json:"isAllDay"`
	Location *struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
	Attendees []struct {
		EmailAddress struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"emailAddress"`
		Type   string `json:"type"`
		Status struct {
			Response string `json:"response"`
		} `json:"status"`
	} `json:"attendees"`
	Organizer *struct {
		EmailAddress struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"emailAddress"`
	} `json:"organizer"`
	IsOnlineMeeting bool `json:"isOnlineMeeting"`
	OnlineMeeting   *struct {
		JoinUrl string `json:"joinUrl"`
	} `json:"onlineMeeting"`
	ResponseStatus *struct {
		Response string `json:"response"`
	} `json:"responseStatus"`
	Importance                   string   `json:"importance"`
	ShowAs                       string   `json:"showAs"`
	IsCancelled                  bool     `json:"isCancelled"`
	IsReminderOn                 bool     `json:"isReminderOn"`
	ReminderMinutesBeforeStart   int      `json:"reminderMinutesBeforeStart"`
	Categories                   []string `json:"categories"`
	CreatedDateTime              string   `json:"createdDateTime"`
	LastModifiedDateTime         string   `json:"lastModifiedDateTime"`
}

func (s *OutlookService) saveEvent(ctx context.Context, userID string, event *graphEvent) error {
	// Parse start/end times
	var startDateTime, endDateTime *time.Time
	if event.Start.DateTime != "" {
		t, _ := time.Parse("2006-01-02T15:04:05.0000000", event.Start.DateTime)
		startDateTime = &t
	}
	if event.End.DateTime != "" {
		t, _ := time.Parse("2006-01-02T15:04:05.0000000", event.End.DateTime)
		endDateTime = &t
	}

	// Parse attendees
	attendees := make([]EventAttendee, 0)
	for _, a := range event.Attendees {
		attendees = append(attendees, EventAttendee{
			Email:          a.EmailAddress.Address,
			Name:           a.EmailAddress.Name,
			Type:           a.Type,
			ResponseStatus: a.Status.Response,
		})
	}

	// Extract organizer
	var organizerEmail, organizerName string
	if event.Organizer != nil {
		organizerEmail = event.Organizer.EmailAddress.Address
		organizerName = event.Organizer.EmailAddress.Name
	}

	// Extract location
	var locationDisplayName string
	if event.Location != nil {
		locationDisplayName = event.Location.DisplayName
	}

	// Extract online meeting URL
	var onlineMeetingURL string
	if event.OnlineMeeting != nil {
		onlineMeetingURL = event.OnlineMeeting.JoinUrl
	}

	// Extract response status
	var responseStatus string
	if event.ResponseStatus != nil {
		responseStatus = event.ResponseStatus.Response
	}

	// Parse timestamps
	var createdDateTime, lastModifiedDateTime *time.Time
	if event.CreatedDateTime != "" {
		t, _ := time.Parse(time.RFC3339, event.CreatedDateTime)
		createdDateTime = &t
	}
	if event.LastModifiedDateTime != "" {
		t, _ := time.Parse(time.RFC3339, event.LastModifiedDateTime)
		lastModifiedDateTime = &t
	}

	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO microsoft_calendar_events (
			user_id, event_id, subject, body_preview,
			location_display_name, start_datetime, start_timezone, end_datetime, end_timezone, is_all_day,
			attendees, organizer_email, organizer_name,
			is_online_meeting, online_meeting_url, response_status,
			importance, show_as, is_cancelled, is_reminder_on, reminder_minutes_before_start,
			categories, created_datetime, last_modified_datetime, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, NOW())
		ON CONFLICT (user_id, event_id) DO UPDATE SET
			subject = EXCLUDED.subject,
			body_preview = EXCLUDED.body_preview,
			location_display_name = EXCLUDED.location_display_name,
			start_datetime = EXCLUDED.start_datetime,
			start_timezone = EXCLUDED.start_timezone,
			end_datetime = EXCLUDED.end_datetime,
			end_timezone = EXCLUDED.end_timezone,
			is_all_day = EXCLUDED.is_all_day,
			attendees = EXCLUDED.attendees,
			organizer_email = EXCLUDED.organizer_email,
			organizer_name = EXCLUDED.organizer_name,
			is_online_meeting = EXCLUDED.is_online_meeting,
			online_meeting_url = EXCLUDED.online_meeting_url,
			response_status = EXCLUDED.response_status,
			importance = EXCLUDED.importance,
			show_as = EXCLUDED.show_as,
			is_cancelled = EXCLUDED.is_cancelled,
			is_reminder_on = EXCLUDED.is_reminder_on,
			reminder_minutes_before_start = EXCLUDED.reminder_minutes_before_start,
			categories = EXCLUDED.categories,
			last_modified_datetime = EXCLUDED.last_modified_datetime,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, event.ID, event.Subject, event.BodyPreview,
		locationDisplayName, startDateTime, event.Start.TimeZone, endDateTime, event.End.TimeZone, event.IsAllDay,
		attendees, organizerEmail, organizerName,
		event.IsOnlineMeeting, onlineMeetingURL, responseStatus,
		event.Importance, event.ShowAs, event.IsCancelled, event.IsReminderOn, event.ReminderMinutesBeforeStart,
		event.Categories, createdDateTime, lastModifiedDateTime)

	return err
}

// GetEvents retrieves calendar events for a user.
func (s *OutlookService) GetEvents(ctx context.Context, userID string, start, end time.Time) ([]*OutlookEvent, error) {
	rows, err := s.provider.Pool().Query(ctx, `
		SELECT id, user_id, event_id, calendar_id, subject, body_preview,
			location_display_name, start_datetime, start_timezone, end_datetime, end_timezone, is_all_day,
			organizer_email, organizer_name, is_online_meeting, online_meeting_url, response_status,
			importance, show_as, is_cancelled, is_reminder_on, reminder_minutes_before_start,
			categories, created_datetime, last_modified_datetime, synced_at
		FROM microsoft_calendar_events
		WHERE user_id = $1 AND start_datetime >= $2 AND start_datetime < $3
		ORDER BY start_datetime
	`, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*OutlookEvent
	for rows.Next() {
		var e OutlookEvent
		var calendarID, bodyPreview, locationDisplayName, startTZ, endTZ *string
		var organizerEmail, organizerName, onlineMeetingURL, responseStatus, importance, showAs *string
		var createdDateTime, lastModifiedDateTime *time.Time
		var categories []string

		err := rows.Scan(
			&e.ID, &e.UserID, &e.EventID, &calendarID, &e.Subject, &bodyPreview,
			&locationDisplayName, &e.StartDateTime, &startTZ, &e.EndDateTime, &endTZ, &e.IsAllDay,
			&organizerEmail, &organizerName, &e.IsOnlineMeeting, &onlineMeetingURL, &responseStatus,
			&importance, &showAs, &e.IsCancelled, &e.IsReminderOn, &e.ReminderMinutesBefore,
			&categories, &createdDateTime, &lastModifiedDateTime, &e.SyncedAt,
		)
		if err != nil {
			return nil, err
		}

		if calendarID != nil {
			e.CalendarID = *calendarID
		}
		if bodyPreview != nil {
			e.BodyPreview = *bodyPreview
		}
		if locationDisplayName != nil {
			e.LocationDisplayName = *locationDisplayName
		}
		if startTZ != nil {
			e.StartTimeZone = *startTZ
		}
		if endTZ != nil {
			e.EndTimeZone = *endTZ
		}
		if organizerEmail != nil {
			e.OrganizerEmail = *organizerEmail
		}
		if organizerName != nil {
			e.OrganizerName = *organizerName
		}
		if onlineMeetingURL != nil {
			e.OnlineMeetingURL = *onlineMeetingURL
		}
		if responseStatus != nil {
			e.ResponseStatus = *responseStatus
		}
		if importance != nil {
			e.Importance = *importance
		}
		if showAs != nil {
			e.ShowAs = *showAs
		}
		e.Categories = categories
		if createdDateTime != nil {
			e.CreatedDateTime = *createdDateTime
		}
		if lastModifiedDateTime != nil {
			e.LastModifiedDateTime = *lastModifiedDateTime
		}

		events = append(events, &e)
	}

	return events, nil
}

// CreateEvent creates a new calendar event.
func (s *OutlookService) CreateEvent(ctx context.Context, userID string, event *OutlookEvent) (*OutlookEvent, error) {
	client, err := s.provider.GetHTTPClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	eventData := map[string]interface{}{
		"subject": event.Subject,
		"start": map[string]string{
			"dateTime": event.StartDateTime.Format("2006-01-02T15:04:05"),
			"timeZone": event.StartTimeZone,
		},
		"end": map[string]string{
			"dateTime": event.EndDateTime.Format("2006-01-02T15:04:05"),
			"timeZone": event.EndTimeZone,
		},
		"isAllDay": event.IsAllDay,
	}

	if event.LocationDisplayName != "" {
		eventData["location"] = map[string]string{
			"displayName": event.LocationDisplayName,
		}
	}

	if event.BodyContent != "" {
		eventData["body"] = map[string]string{
			"contentType": "text",
			"content":     event.BodyContent,
		}
	}

	jsonBody, _ := json.Marshal(eventData)

	resp, err := client.Post(GraphAPIBase+"/me/events", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create event: %s", resp.Status)
	}

	var created graphEvent
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Save to database
	if err := s.saveEvent(ctx, userID, &created); err != nil {
		log.Printf("Failed to save created event to database: %v", err)
	}

	event.EventID = created.ID
	return event, nil
}

// IsMailConnected checks if Outlook Mail is connected for a user.
func (s *OutlookService) IsMailConnected(ctx context.Context, userID string) bool {
	var scopes []string
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT scopes FROM microsoft_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&scopes)
	if err != nil {
		return false
	}

	for _, scope := range scopes {
		if containsMailScope(scope) {
			return true
		}
	}
	return false
}

// IsCalendarConnected checks if Outlook Calendar is connected for a user.
func (s *OutlookService) IsCalendarConnected(ctx context.Context, userID string) bool {
	var scopes []string
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT scopes FROM microsoft_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&scopes)
	if err != nil {
		return false
	}

	for _, scope := range scopes {
		if containsCalendarScope(scope) {
			return true
		}
	}
	return false
}

func containsMailScope(scope string) bool {
	mailScopes := []string{"Mail.Read", "Mail.ReadWrite", "Mail.Send"}
	for _, s := range mailScopes {
		if scope == s {
			return true
		}
	}
	return false
}

func containsCalendarScope(scope string) bool {
	calendarScopes := []string{"Calendars.Read", "Calendars.ReadWrite"}
	for _, s := range calendarScopes {
		if scope == s {
			return true
		}
	}
	return false
}
