package google

import "time"

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

// SyncEmailsResult represents the result of an email sync.
type SyncEmailsResult struct {
	TotalEmails  int `json:"total_emails"`
	SyncedEmails int `json:"synced_emails"`
	FailedEmails int `json:"failed_emails"`
}

// GmailService handles Gmail operations.
type GmailService struct {
	provider *Provider
}

// NewGmailService creates a new Gmail service.
func NewGmailService(provider *Provider) *GmailService {
	return &GmailService{provider: provider}
}
