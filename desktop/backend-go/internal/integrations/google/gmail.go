package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Email represents a synced email.
type Email struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Provider    string         `json:"provider"`
	ExternalID  string         `json:"external_id"`
	ThreadID    string         `json:"thread_id,omitempty"`
	Subject     string         `json:"subject"`
	Snippet     string         `json:"snippet"`
	FromEmail   string         `json:"from_email"`
	FromName    string         `json:"from_name"`
	ToEmails    []EmailAddress `json:"to_emails"`
	CcEmails    []EmailAddress `json:"cc_emails"`
	ReplyTo     string         `json:"reply_to,omitempty"`
	BodyText    string         `json:"body_text,omitempty"`
	BodyHTML    string         `json:"body_html,omitempty"`
	Attachments []Attachment   `json:"attachments,omitempty"`
	IsRead      bool           `json:"is_read"`
	IsStarred   bool           `json:"is_starred"`
	IsImportant bool           `json:"is_important"`
	IsDraft     bool           `json:"is_draft"`
	IsSent      bool           `json:"is_sent"`
	IsArchived  bool           `json:"is_archived"`
	IsTrash     bool           `json:"is_trash"`
	Labels      []string       `json:"labels"`
	Date        time.Time      `json:"date"`
	ReceivedAt  time.Time      `json:"received_at"`
}

// EmailAddress represents an email address with optional name.
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Attachment represents an email attachment.
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// ComposeEmail represents an email to be sent.
type ComposeEmail struct {
	To      []string `json:"to"`
	Cc      []string `json:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

// EmailFolder represents a mail folder.
type EmailFolder string

const (
	FolderInbox   EmailFolder = "inbox"
	FolderSent    EmailFolder = "sent"
	FolderDrafts  EmailFolder = "drafts"
	FolderStarred EmailFolder = "starred"
	FolderArchive EmailFolder = "archive"
	FolderTrash   EmailFolder = "trash"
)

// GmailService handles Gmail operations.
type GmailService struct {
	provider *Provider
}

// NewGmailService creates a new Gmail service.
func NewGmailService(provider *Provider) *GmailService {
	return &GmailService{provider: provider}
}

// GetGmailAPI returns a Gmail API service for a user.
func (s *GmailService) GetGmailAPI(ctx context.Context, userID string) (*gmail.Service, error) {
	tokenSource, err := s.provider.GetTokenSource(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token source: %w", err)
	}

	srv, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return srv, nil
}

// SyncEmails syncs emails from Gmail.
func (s *GmailService) SyncEmails(ctx context.Context, userID string, maxResults int64) (*SyncEmailsResult, error) {
	log.Printf("Gmail sync starting for user %s: max %d emails", userID, maxResults)

	srv, err := s.GetGmailAPI(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Gmail API: %w", err)
	}

	// Get list of messages
	req := srv.Users.Messages.List("me").
		MaxResults(maxResults).
		Q("in:inbox OR in:sent")

	messages, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	result := &SyncEmailsResult{
		TotalEmails: len(messages.Messages),
	}

	for _, msg := range messages.Messages {
		// Get full message details
		fullMsg, err := srv.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("Failed to get message %s: %v", msg.Id, err)
			result.FailedEmails++
			continue
		}

		// Parse and save email
		if err := s.saveEmail(ctx, userID, fullMsg); err != nil {
			log.Printf("Failed to save message %s: %v", msg.Id, err)
			result.FailedEmails++
		} else {
			result.SyncedEmails++
		}
	}

	log.Printf("Gmail sync complete for user %s: synced %d/%d emails",
		userID, result.SyncedEmails, result.TotalEmails)

	return result, nil
}

// SyncEmailsResult represents the result of an email sync.
type SyncEmailsResult struct {
	TotalEmails  int `json:"total_emails"`
	SyncedEmails int `json:"synced_emails"`
	FailedEmails int `json:"failed_emails"`
}

// saveEmail saves an email to the database.
func (s *GmailService) saveEmail(ctx context.Context, userID string, msg *gmail.Message) error {
	// Parse headers
	var subject, from, to, cc, replyTo, date string
	for _, header := range msg.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "subject":
			subject = header.Value
		case "from":
			from = header.Value
		case "to":
			to = header.Value
		case "cc":
			cc = header.Value
		case "reply-to":
			replyTo = header.Value
		case "date":
			date = header.Value
		}
	}

	// Parse from address
	fromName, fromEmail := parseEmailAddress(from)

	// Parse to/cc addresses
	toAddrs := parseEmailAddresses(to)
	ccAddrs := parseEmailAddresses(cc)

	// Get body
	bodyText, bodyHTML := extractBody(msg.Payload)

	// Parse labels
	labels := msg.LabelIds
	isRead := !containsLabel(labels, "UNREAD")
	isStarred := containsLabel(labels, "STARRED")
	isImportant := containsLabel(labels, "IMPORTANT")
	isDraft := containsLabel(labels, "DRAFT")
	isSent := containsLabel(labels, "SENT")
	isArchived := !containsLabel(labels, "INBOX") && !containsLabel(labels, "TRASH") && !containsLabel(labels, "SPAM")
	isTrash := containsLabel(labels, "TRASH")

	// Parse attachments
	attachments := extractAttachments(msg.Payload)

	// Parse date
	var emailDate time.Time
	if date != "" {
		if parsed, err := parseEmailDate(date); err == nil {
			emailDate = parsed
		}
	}
	if emailDate.IsZero() && msg.InternalDate != 0 {
		emailDate = time.UnixMilli(msg.InternalDate)
	}

	// Insert into database
	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO emails (
			user_id, provider, external_id, thread_id,
			subject, snippet, from_email, from_name,
			to_emails, cc_emails, reply_to,
			body_text, body_html, attachments,
			is_read, is_starred, is_important, is_draft, is_sent, is_archived, is_trash,
			labels, date, received_at, synced_at
		) VALUES (
			$1, 'gmail', $2, $3,
			$4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, NOW()
		)
		ON CONFLICT (user_id, provider, external_id) DO UPDATE SET
			subject = EXCLUDED.subject,
			snippet = EXCLUDED.snippet,
			is_read = EXCLUDED.is_read,
			is_starred = EXCLUDED.is_starred,
			is_important = EXCLUDED.is_important,
			is_archived = EXCLUDED.is_archived,
			is_trash = EXCLUDED.is_trash,
			labels = EXCLUDED.labels,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, msg.Id, msg.ThreadId,
		subject, msg.Snippet, fromEmail, fromName,
		toAddrs, ccAddrs, replyTo,
		bodyText, bodyHTML, attachments,
		isRead, isStarred, isImportant, isDraft, isSent, isArchived, isTrash,
		labels, emailDate, emailDate)

	return err
}

// GetEmails retrieves emails for a user.
func (s *GmailService) GetEmails(ctx context.Context, userID string, folder EmailFolder, limit, offset int) ([]*Email, error) {
	query := `
		SELECT id, user_id, provider, external_id, thread_id,
			subject, snippet, from_email, from_name,
			to_emails, cc_emails, reply_to,
			body_text, body_html, attachments,
			is_read, is_starred, is_important, is_draft, is_sent, is_archived, is_trash,
			labels, date, received_at
		FROM emails
		WHERE user_id = $1
	`

	// Add folder filter
	switch folder {
	case FolderInbox:
		query += " AND is_archived = false AND is_trash = false AND is_draft = false"
	case FolderSent:
		query += " AND is_sent = true"
	case FolderDrafts:
		query += " AND is_draft = true"
	case FolderStarred:
		query += " AND is_starred = true"
	case FolderArchive:
		query += " AND is_archived = true"
	case FolderTrash:
		query += " AND is_trash = true"
	}

	query += " ORDER BY date DESC LIMIT $2 OFFSET $3"

	rows, err := s.provider.Pool().Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []*Email
	for rows.Next() {
		var e Email
		var toEmails, ccEmails, attachments, labels []byte
		var bodyText, bodyHTML, replyTo pgtype.Text
		var date, receivedAt pgtype.Timestamptz

		err := rows.Scan(
			&e.ID, &e.UserID, &e.Provider, &e.ExternalID, &e.ThreadID,
			&e.Subject, &e.Snippet, &e.FromEmail, &e.FromName,
			&toEmails, &ccEmails, &replyTo,
			&bodyText, &bodyHTML, &attachments,
			&e.IsRead, &e.IsStarred, &e.IsImportant, &e.IsDraft, &e.IsSent, &e.IsArchived, &e.IsTrash,
			&labels, &date, &receivedAt,
		)
		if err != nil {
			return nil, err
		}

		e.BodyText = bodyText.String
		e.BodyHTML = bodyHTML.String
		e.ReplyTo = replyTo.String
		if date.Valid {
			e.Date = date.Time
		}
		if receivedAt.Valid {
			e.ReceivedAt = receivedAt.Time
		}

		emails = append(emails, &e)
	}

	return emails, nil
}

// GetEmail retrieves a single email by ID.
func (s *GmailService) GetEmail(ctx context.Context, userID, emailID string) (*Email, error) {
	var e Email
	var toEmails, ccEmails, attachments, labels []byte
	var bodyText, bodyHTML, replyTo pgtype.Text
	var date, receivedAt pgtype.Timestamptz

	err := s.provider.Pool().QueryRow(ctx, `
		SELECT id, user_id, provider, external_id, thread_id,
			subject, snippet, from_email, from_name,
			to_emails, cc_emails, reply_to,
			body_text, body_html, attachments,
			is_read, is_starred, is_important, is_draft, is_sent, is_archived, is_trash,
			labels, date, received_at
		FROM emails
		WHERE id = $1 AND user_id = $2
	`, emailID, userID).Scan(
		&e.ID, &e.UserID, &e.Provider, &e.ExternalID, &e.ThreadID,
		&e.Subject, &e.Snippet, &e.FromEmail, &e.FromName,
		&toEmails, &ccEmails, &replyTo,
		&bodyText, &bodyHTML, &attachments,
		&e.IsRead, &e.IsStarred, &e.IsImportant, &e.IsDraft, &e.IsSent, &e.IsArchived, &e.IsTrash,
		&labels, &date, &receivedAt,
	)
	if err != nil {
		return nil, err
	}

	e.BodyText = bodyText.String
	e.BodyHTML = bodyHTML.String
	e.ReplyTo = replyTo.String
	if date.Valid {
		e.Date = date.Time
	}
	if receivedAt.Valid {
		e.ReceivedAt = receivedAt.Time
	}

	return &e, nil
}

// MarkAsRead marks an email as read.
func (s *GmailService) MarkAsRead(ctx context.Context, userID, emailID string) error {
	// Update in database
	_, err := s.provider.Pool().Exec(ctx, `
		UPDATE emails SET is_read = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, emailID, userID)
	if err != nil {
		return err
	}

	// Also update in Gmail
	email, err := s.GetEmail(ctx, userID, emailID)
	if err != nil {
		return nil // Don't fail if we can't get email details
	}

	srv, err := s.GetGmailAPI(ctx, userID)
	if err != nil {
		return nil
	}

	_, err = srv.Users.Messages.Modify("me", email.ExternalID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Do()

	return err
}

// GetEmailByID retrieves a single email by its database ID.
// Alias for GetEmail for handler compatibility.
func (s *GmailService) GetEmailByID(ctx context.Context, userID, emailID string) (*Email, error) {
	return s.GetEmail(ctx, userID, emailID)
}

// ArchiveEmail archives an email by removing it from inbox.
func (s *GmailService) ArchiveEmail(ctx context.Context, userID, emailID string) error {
	// Update in database
	_, err := s.provider.Pool().Exec(ctx, `
		UPDATE emails SET is_archived = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, emailID, userID)
	if err != nil {
		return err
	}

	// Also update in Gmail
	email, err := s.GetEmail(ctx, userID, emailID)
	if err != nil {
		return nil // Don't fail if we can't get email details
	}

	srv, err := s.GetGmailAPI(ctx, userID)
	if err != nil {
		return nil
	}

	_, err = srv.Users.Messages.Modify("me", email.ExternalID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}).Do()

	return err
}

// DeleteEmail moves an email to trash.
func (s *GmailService) DeleteEmail(ctx context.Context, userID, emailID string) error {
	// Get email details first
	email, err := s.GetEmail(ctx, userID, emailID)
	if err != nil {
		return err
	}

	// Move to trash in Gmail
	srv, err := s.GetGmailAPI(ctx, userID)
	if err != nil {
		return err
	}

	_, err = srv.Users.Messages.Trash("me", email.ExternalID).Do()
	if err != nil {
		return err
	}

	// Update in database
	_, err = s.provider.Pool().Exec(ctx, `
		UPDATE emails SET is_trash = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, emailID, userID)

	return err
}

// SendEmail sends an email via Gmail.
func (s *GmailService) SendEmail(ctx context.Context, userID string, email *ComposeEmail) error {
	srv, err := s.GetGmailAPI(ctx, userID)
	if err != nil {
		return err
	}

	// Build the message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	if len(email.Cc) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.Cc, ", ")))
	}
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	if email.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(email.Body)

	// Encode the message
	raw := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))

	// Send the message
	_, err = srv.Users.Messages.Send("me", &gmail.Message{
		Raw: raw,
	}).Do()

	return err
}

// IsConnected checks if Gmail is connected for a user.
func (s *GmailService) IsConnected(ctx context.Context, userID string) bool {
	// Check if user has gmail scopes
	var scopes []string
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT scopes FROM google_oauth_tokens WHERE user_id = $1
	`, userID).Scan(&scopes)
	if err != nil {
		return false
	}

	for _, scope := range scopes {
		if strings.Contains(scope, "gmail") {
			return true
		}
	}
	return false
}

// Helper functions

func parseEmailAddress(addr string) (name, email string) {
	addr = strings.TrimSpace(addr)
	if strings.Contains(addr, "<") {
		parts := strings.Split(addr, "<")
		name = strings.TrimSpace(parts[0])
		name = strings.Trim(name, "\"")
		email = strings.TrimSuffix(parts[1], ">")
	} else {
		email = addr
	}
	return name, email
}

func parseEmailAddresses(addrs string) []EmailAddress {
	var result []EmailAddress
	if addrs == "" {
		return result
	}

	parts := strings.Split(addrs, ",")
	for _, part := range parts {
		name, email := parseEmailAddress(strings.TrimSpace(part))
		if email != "" {
			result = append(result, EmailAddress{Name: name, Email: email})
		}
	}
	return result
}

func extractBody(payload *gmail.MessagePart) (text, html string) {
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
		text = string(decoded)
	} else if payload.MimeType == "text/html" && payload.Body != nil && payload.Body.Data != "" {
		decoded, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
		html = string(decoded)
	}

	for _, part := range payload.Parts {
		partText, partHTML := extractBody(part)
		if partText != "" && text == "" {
			text = partText
		}
		if partHTML != "" && html == "" {
			html = partHTML
		}
	}

	return text, html
}

func extractAttachments(payload *gmail.MessagePart) []Attachment {
	var attachments []Attachment

	if payload.Filename != "" && payload.Body != nil && payload.Body.AttachmentId != "" {
		attachments = append(attachments, Attachment{
			ID:       payload.Body.AttachmentId,
			Filename: payload.Filename,
			MimeType: payload.MimeType,
			Size:     payload.Body.Size,
		})
	}

	for _, part := range payload.Parts {
		attachments = append(attachments, extractAttachments(part)...)
	}

	return attachments
}

func containsLabel(labels []string, label string) bool {
	for _, l := range labels {
		if l == label {
			return true
		}
	}
	return false
}

func parseEmailDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
		"2 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
