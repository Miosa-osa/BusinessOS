package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/integrations/google"
)

// EmailAnalyzerService extracts metadata from user emails for onboarding analysis
type EmailAnalyzerService struct {
	pool         *pgxpool.Pool
	gmailService *google.GmailService
}

// EmailAnalysisMetadata represents extracted metadata from emails
type EmailAnalysisMetadata struct {
	TotalEmails     int                    `json:"total_emails"`
	SenderDomains   map[string]int         `json:"sender_domains"`
	SubjectKeywords []string               `json:"subject_keywords"`
	BodyKeywords    []string               `json:"body_keywords"`
	DetectedTools   map[string]int         `json:"detected_tools"`
	TopicFrequency  map[string]int         `json:"topic_frequency"`
	EmailDates      []time.Time            `json:"email_dates"`
	RawMetadata     map[string]interface{} `json:"raw_metadata,omitempty"`
}

// Known tools/platforms to detect in emails
var knownTools = []string{
	// Design & Creative
	"Figma", "Sketch", "Adobe XD", "Canva", "Photoshop", "Illustrator",
	// Development
	"GitHub", "GitLab", "VS Code", "Visual Studio", "IntelliJ", "Docker", "Kubernetes",
	// Project Management
	"Notion", "Asana", "Trello", "Jira", "Monday.com", "ClickUp", "Linear",
	// Communication
	"Slack", "Discord", "Teams", "Zoom", "Meet", "Loom",
	// Productivity
	"Airtable", "Coda", "Roam", "Obsidian", "Evernote", "Todoist",
	// Marketing
	"HubSpot", "Mailchimp", "SendGrid", "Intercom", "Segment",
	// Analytics
	"Google Analytics", "Mixpanel", "Amplitude", "Heap",
	// No-code
	"Webflow", "Bubble", "Zapier", "Make", "n8n", "Retool",
	// CRM
	"Salesforce", "HubSpot CRM", "Pipedrive", "Copper",
	// Other
	"Stripe", "PayPal", "Shopify", "WordPress", "Framer",
}

// Common topic patterns to detect
var topicPatterns = map[string]*regexp.Regexp{
	"design":        regexp.MustCompile(`(?i)\b(design|ui|ux|prototype|mockup|wireframe)\b`),
	"development":   regexp.MustCompile(`(?i)\b(code|develop|build|api|backend|frontend|deploy)\b`),
	"marketing":     regexp.MustCompile(`(?i)\b(market|campaign|seo|content|social|ads)\b`),
	"sales":         regexp.MustCompile(`(?i)\b(sale|deal|prospect|lead|client|revenue)\b`),
	"product":       regexp.MustCompile(`(?i)\b(product|feature|roadmap|launch|release)\b`),
	"analytics":     regexp.MustCompile(`(?i)\b(analytic|metric|data|report|insight)\b`),
	"automation":    regexp.MustCompile(`(?i)\b(automat|workflow|integration|api)\b`),
	"collaboration": regexp.MustCompile(`(?i)\b(collaborat|team|meeting|sync|standup)\b`),
}

func NewEmailAnalyzerService(pool *pgxpool.Pool, gmailService *google.GmailService) *EmailAnalyzerService {
	return &EmailAnalyzerService{
		pool:         pool,
		gmailService: gmailService,
	}
}

// AnalyzeRecentEmails fetches and analyzes recent emails for a user
func (s *EmailAnalyzerService) AnalyzeRecentEmails(ctx context.Context, userID string, maxEmails int) (*EmailAnalysisMetadata, error) {
	slog.Info("EmailAnalyzerService starting analysis",
		"user_id", userID,
		"max_emails", maxEmails,
	)

	// Fetch recent emails from Gmail
	slog.Info("Syncing emails from Gmail")
	syncResult, err := s.gmailService.SyncEmails(ctx, userID, int64(maxEmails))
	if err != nil {
		slog.Error("Failed to sync emails", "error", err)
		return nil, fmt.Errorf("failed to sync emails: %w", err)
	}

	slog.Info("Emails synced",
		"total", syncResult.TotalEmails,
		"synced", syncResult.SyncedEmails,
		"failed", syncResult.FailedEmails,
	)

	// Get synced emails from database
	emails, err := s.gmailService.GetEmails(ctx, userID, google.FolderInbox, maxEmails, 0)
	if err != nil {
		slog.Error("Failed to retrieve emails", "error", err)
		return nil, fmt.Errorf("failed to retrieve emails: %w", err)
	}

	if len(emails) == 0 {
		slog.Warn("No emails found for user", "user_id", userID)
		return &EmailAnalysisMetadata{
			TotalEmails:     0,
			SenderDomains:   make(map[string]int),
			SubjectKeywords: []string{},
			BodyKeywords:    []string{},
			DetectedTools:   make(map[string]int),
			TopicFrequency:  make(map[string]int),
			EmailDates:      []time.Time{},
		}, nil
	}

	// Extract metadata from emails
	metadata := s.extractMetadata(emails)
	metadata.TotalEmails = len(emails)

	slog.Info("Email analysis complete",
		"emails_analyzed", metadata.TotalEmails,
		"unique_domains", len(metadata.SenderDomains),
		"tools_detected", len(metadata.DetectedTools),
		"topics_detected", len(metadata.TopicFrequency),
	)

	return metadata, nil
}

// extractMetadata processes emails and extracts structured metadata
func (s *EmailAnalyzerService) extractMetadata(emails []*google.Email) *EmailAnalysisMetadata {
	metadata := &EmailAnalysisMetadata{
		SenderDomains:   make(map[string]int),
		SubjectKeywords: []string{},
		BodyKeywords:    []string{},
		DetectedTools:   make(map[string]int),
		TopicFrequency:  make(map[string]int),
		EmailDates:      []time.Time{},
	}

	subjectWords := make(map[string]int)
	bodyWords := make(map[string]int)

	for _, email := range emails {
		// Extract sender domain
		domain := extractDomainFromEmail(email.FromEmail)
		if domain != "" {
			metadata.SenderDomains[domain]++
		}

		// Extract date
		if !email.Date.IsZero() {
			metadata.EmailDates = append(metadata.EmailDates, email.Date)
		}

		// Process subject
		subjectText := strings.ToLower(email.Subject)
		for _, word := range extractKeywordsFromText(subjectText) {
			subjectWords[word]++
		}

		// Process body (snippet + full text if available)
		bodyText := strings.ToLower(email.Snippet)
		if email.BodyText != "" {
			bodyText += " " + strings.ToLower(email.BodyText)
		}

		for _, word := range extractKeywordsFromText(bodyText) {
			bodyWords[word]++
		}

		// Detect tools
		fullText := email.Subject + " " + email.Snippet + " " + email.BodyText
		for _, tool := range knownTools {
			if strings.Contains(strings.ToLower(fullText), strings.ToLower(tool)) {
				metadata.DetectedTools[tool]++
			}
		}

		// Detect topics
		for topic, pattern := range topicPatterns {
			if pattern.MatchString(fullText) {
				metadata.TopicFrequency[topic]++
			}
		}
	}

	// Convert word maps to top keywords (top 20)
	metadata.SubjectKeywords = topNWords(subjectWords, 20)
	metadata.BodyKeywords = topNWords(bodyWords, 20)

	return metadata
}

// extractDomainFromEmail extracts domain from email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// extractKeywordsFromText extracts meaningful words from text (filters stop words)
func extractKeywordsFromText(text string) []string {
	// Simple word extraction (split on non-alphanumeric)
	words := regexp.MustCompile(`[^\w]+`).Split(text, -1)

	// Filter stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
		"to": true, "of": true, "in": true, "for": true, "on": true, "with": true,
		"as": true, "at": true, "by": true, "from": true, "it": true, "that": true,
		"this": true, "you": true, "your": true, "we": true, "our": true, "have": true,
		"has": true, "had": true, "can": true, "will": true, "would": true,
	}

	var keywords []string
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// topNWords returns the top N words by frequency
func topNWords(wordMap map[string]int, n int) []string {
	type wordCount struct {
		word  string
		count int
	}

	var counts []wordCount
	for word, count := range wordMap {
		counts = append(counts, wordCount{word, count})
	}

	// Sort by count (descending)
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	// Take top N
	if len(counts) > n {
		counts = counts[:n]
	}

	var result []string
	for _, wc := range counts {
		result = append(result, wc.word)
	}

	return result
}
