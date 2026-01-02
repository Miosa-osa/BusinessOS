package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// ConversationIntelligenceService provides intelligent analysis of conversations
type ConversationIntelligenceService struct {
	pool   *pgxpool.Pool
	embed  *EmbeddingService
	logger *slog.Logger
}

// ConversationAnalysis represents a detailed analysis of a conversation
type ConversationAnalysis struct {
	ID              string                 `json:"id"`
	ConversationID  string                 `json:"conversation_id"`
	UserID          string                 `json:"user_id"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	KeyPoints       []string               `json:"key_points"`
	Topics          []ConversationTopic    `json:"topics"`
	Sentiment       SentimentAnalysis      `json:"sentiment"`
	Entities        []ConversationEntity   `json:"entities"`
	ActionItems     []ActionItem           `json:"action_items"`
	Questions       []Question             `json:"questions"`
	Decisions       []ConversationDecision `json:"decisions"`
	CodeMentions    []CodeMention          `json:"code_mentions"`
	MessageCount    int                    `json:"message_count"`
	TokenCount      int                    `json:"token_count"`
	Duration        string                 `json:"duration"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ConversationTopic represents a topic discussed in the conversation
type ConversationTopic struct {
	Name       string   `json:"name"`
	Confidence float64  `json:"confidence"`
	Keywords   []string `json:"keywords"`
	FirstMention int    `json:"first_mention"` // Message index
	Frequency  int      `json:"frequency"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Overall     string             `json:"overall"` // positive, negative, neutral, mixed
	Score       float64            `json:"score"`   // -1 to 1
	Progression []SentimentPoint   `json:"progression"`
	Highlights  []SentimentHighlight `json:"highlights"`
}

// SentimentPoint represents sentiment at a point in the conversation
type SentimentPoint struct {
	MessageIndex int     `json:"message_index"`
	Sentiment    string  `json:"sentiment"`
	Score        float64 `json:"score"`
}

// SentimentHighlight represents a significant sentiment moment
type SentimentHighlight struct {
	MessageIndex int    `json:"message_index"`
	Text         string `json:"text"`
	Sentiment    string `json:"sentiment"`
	Reason       string `json:"reason"`
}

// ConversationEntity represents an entity mentioned in the conversation
type ConversationEntity struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"` // person, organization, technology, file, concept
	Mentions   int      `json:"mentions"`
	Context    []string `json:"context"`
	Related    []string `json:"related"`
}

// ActionItem represents a task or action mentioned
type ActionItem struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Priority     string   `json:"priority"` // high, medium, low
	Status       string   `json:"status"`   // pending, completed, cancelled
	AssignedTo   string   `json:"assigned_to,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	MessageIndex int      `json:"message_index"`
	Tags         []string `json:"tags"`
}

// Question represents a question in the conversation
type Question struct {
	Text         string `json:"text"`
	AskedBy      string `json:"asked_by"` // user, assistant
	MessageIndex int    `json:"message_index"`
	Answered     bool   `json:"answered"`
	Answer       string `json:"answer,omitempty"`
}

// ConversationDecision represents a decision made in the conversation
type ConversationDecision struct {
	Description  string   `json:"description"`
	Context      string   `json:"context"`
	Alternatives []string `json:"alternatives,omitempty"`
	Rationale    string   `json:"rationale,omitempty"`
	MessageIndex int      `json:"message_index"`
}

// CodeMention represents code discussed in the conversation
type CodeMention struct {
	FilePath     string `json:"file_path,omitempty"`
	Language     string `json:"language,omitempty"`
	Snippet      string `json:"snippet"`
	Context      string `json:"context"`
	MessageIndex int    `json:"message_index"`
}

// Message represents a conversation message for analysis
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// wordCount is used for keyword frequency analysis
type wordCount struct {
	word  string
	count int
}

// NewConversationIntelligenceService creates a new conversation intelligence service
func NewConversationIntelligenceService(pool *pgxpool.Pool, embeddingService *EmbeddingService) *ConversationIntelligenceService {
	return &ConversationIntelligenceService{
		pool:   pool,
		embed:  embeddingService,
		logger: slog.Default().With("service", "conversation_intelligence"),
	}
}

// AnalyzeConversation performs full analysis on a conversation
func (s *ConversationIntelligenceService) AnalyzeConversation(ctx context.Context, conversationID, userID string, messages []Message) (*ConversationAnalysis, error) {
	analysis := &ConversationAnalysis{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		UserID:         userID,
		KeyPoints:      make([]string, 0),
		Topics:         make([]ConversationTopic, 0),
		Entities:       make([]ConversationEntity, 0),
		ActionItems:    make([]ActionItem, 0),
		Questions:      make([]Question, 0),
		Decisions:      make([]ConversationDecision, 0),
		CodeMentions:   make([]CodeMention, 0),
		Metadata:       make(map[string]interface{}),
		MessageCount:   len(messages),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if len(messages) == 0 {
		return analysis, nil
	}

	// Calculate duration
	if len(messages) > 1 {
		duration := messages[len(messages)-1].Timestamp.Sub(messages[0].Timestamp)
		analysis.Duration = duration.Round(time.Second).String()
	}
	// Track time range for storage
	if analysis.Metadata != nil && len(messages) > 0 {
		analysis.Metadata["time_range_start"] = messages[0].Timestamp
		analysis.Metadata["time_range_end"] = messages[len(messages)-1].Timestamp
	}

	// Extract topics
	analysis.Topics = s.extractTopics(messages)

	// Extract entities
	analysis.Entities = s.extractEntities(messages)

	// Extract questions
	analysis.Questions = s.extractQuestions(messages)

	// Extract action items
	analysis.ActionItems = s.extractActionItems(messages)

	// Extract decisions
	analysis.Decisions = s.extractDecisions(messages)

	// Extract code mentions
	analysis.CodeMentions = s.extractCodeMentions(messages)

	// Analyze sentiment
	analysis.Sentiment = s.analyzeSentiment(messages)

	// Generate title
	analysis.Title = s.generateTitle(messages, analysis.Topics)

	// Generate summary
	analysis.Summary = s.generateSummary(messages, analysis)

	// Extract key points
	analysis.KeyPoints = s.extractKeyPoints(messages, analysis)

	// Calculate token count (approximation)
	for _, msg := range messages {
		analysis.TokenCount += len(msg.Content) / 4
	}

	// Save to database
	if err := s.saveAnalysis(ctx, analysis); err != nil {
		s.logger.Warn("failed to save conversation analysis", "error", err)
	}

	return analysis, nil
}

// BackfillStaleSummaries generates/updates conversation summaries that are missing or stale.
// A summary is considered stale when its updated_at is older than the conversation's latest message.
// If force is true, it will analyze the most recent conversations regardless of staleness.
func (s *ConversationIntelligenceService) BackfillStaleSummaries(ctx context.Context, limit int, maxMessages int, force bool) (int, error) {
	if limit <= 0 {
		limit = 50
	}
	if maxMessages <= 0 {
		maxMessages = 200
	}

	// Select conversations with message activity, prioritize most recently active.
	// We consider a summary stale if cs.updated_at < last_message_at or missing.
	query := `
		WITH last_msg AS (
			SELECT conversation_id, MAX(created_at) AS last_at
			FROM messages
			GROUP BY conversation_id
		)
		SELECT c.id::text, c.user_id
		FROM conversations c
		JOIN last_msg lm ON lm.conversation_id = c.id
		LEFT JOIN conversation_summaries cs ON cs.conversation_id = c.id
		WHERE ($1::boolean = true) OR cs.conversation_id IS NULL OR cs.updated_at < lm.last_at
		ORDER BY lm.last_at DESC
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, force, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	processed := 0
	for rows.Next() {
		var conversationID string
		var userID string
		if err := rows.Scan(&conversationID, &userID); err != nil {
			continue
		}

		messages, err := s.fetchConversationMessages(ctx, conversationID, maxMessages)
		if err != nil || len(messages) == 0 {
			continue
		}

		_, err = s.AnalyzeConversation(ctx, conversationID, userID, messages)
		if err != nil {
			s.logger.Warn("conversation analysis failed", "conversation_id", conversationID, "error", err)
			continue
		}
		processed++
	}

	return processed, nil
}

func (s *ConversationIntelligenceService) fetchConversationMessages(ctx context.Context, conversationID string, maxMessages int) ([]Message, error) {
	// Pull the most recent N messages, then reverse to chronological.
	rows, err := s.pool.Query(ctx, `
		SELECT role::text, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, conversationID, maxMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tmp []Message
	for rows.Next() {
		var role string
		var content string
		var ts time.Time
		if err := rows.Scan(&role, &content, &ts); err != nil {
			continue
		}
		tmp = append(tmp, Message{Role: role, Content: content, Timestamp: ts})
	}

	// Reverse to chronological order.
	for i, j := 0, len(tmp)-1; i < j; i, j = i+1, j-1 {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp, nil
}

// extractTopics extracts topics from messages
func (s *ConversationIntelligenceService) extractTopics(messages []Message) []ConversationTopic {
	// Keyword frequency analysis
	wordFreq := make(map[string]int)
	wordFirstMention := make(map[string]int)

	// Common stop words to filter
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true, "can": true,
		"this": true, "that": true, "these": true, "those": true, "i": true,
		"you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"what": true, "which": true, "who": true, "whom": true, "whose": true,
		"where": true, "when": true, "why": true, "how": true, "all": true,
		"each": true, "every": true, "both": true, "few": true, "more": true,
		"most": true, "other": true, "some": true, "such": true, "no": true,
		"nor": true, "not": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true,
		"and": true, "but": true, "if": true, "or": true, "because": true,
		"as": true, "until": true, "while": true, "of": true, "at": true,
		"by": true, "for": true, "with": true, "about": true, "against": true,
		"between": true, "into": true, "through": true, "during": true,
		"before": true, "after": true, "above": true, "below": true, "to": true,
		"from": true, "up": true, "down": true, "in": true, "out": true,
		"on": true, "off": true, "over": true, "under": true, "again": true,
		"further": true, "then": true, "once": true, "here": true, "there": true,
	}

	for idx, msg := range messages {
		words := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`).FindAllString(strings.ToLower(msg.Content), -1)
		for _, word := range words {
			if stopWords[word] {
				continue
			}
			wordFreq[word]++
			if _, exists := wordFirstMention[word]; !exists {
				wordFirstMention[word] = idx
			}
		}
	}

	// Sort by frequency
	var counts []wordCount
	for word, count := range wordFreq {
		if count >= 2 { // Minimum frequency threshold
			counts = append(counts, wordCount{word, count})
		}
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	// Create topics from top keywords
	topics := make([]ConversationTopic, 0)
	maxTopics := 5
	for i, wc := range counts {
		if i >= maxTopics {
			break
		}

		// Find related keywords
		relatedKeywords := s.findRelatedKeywords(wc.word, counts)

		topics = append(topics, ConversationTopic{
			Name:         wc.word,
			Confidence:   float64(wc.count) / float64(len(messages)),
			Keywords:     relatedKeywords,
			FirstMention: wordFirstMention[wc.word],
			Frequency:    wc.count,
		})
	}

	return topics
}

// findRelatedKeywords finds keywords that often appear with the given word
func (s *ConversationIntelligenceService) findRelatedKeywords(word string, counts []wordCount) []string {
	related := make([]string, 0)
	for _, wc := range counts {
		if wc.word != word && len(related) < 3 {
			related = append(related, wc.word)
		}
	}
	return related
}

// extractEntities extracts named entities from messages
func (s *ConversationIntelligenceService) extractEntities(messages []Message) []ConversationEntity {
	entities := make(map[string]*ConversationEntity)

	// Patterns for different entity types
	patterns := map[string]*regexp.Regexp{
		"file":       regexp.MustCompile(`(?i)[\w-]+\.(go|ts|js|svelte|py|sql|json|yaml|yml|md|txt|css|html|tsx|jsx)`),
		"path":       regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(/[\w/.-]+|[\w]+/[\w/.-]+)`),
		"function":   regexp.MustCompile(`\b[a-z][a-zA-Z0-9]*\([^)]*\)`),
		"technology": regexp.MustCompile(`(?i)\b(react|svelte|vue|angular|node|go|golang|python|typescript|javascript|postgresql|redis|docker|kubernetes|aws|gcp|azure)\b`),
		"url":        regexp.MustCompile(`https?://[^\s]+`),
	}

	for idx, msg := range messages {
		for entityType, pattern := range patterns {
			matches := pattern.FindAllString(msg.Content, -1)
			for _, match := range matches {
				key := strings.ToLower(match)
				if existing, ok := entities[key]; ok {
					existing.Mentions++
					if len(existing.Context) < 3 {
						contextSnippet := s.extractContext(msg.Content, match, 50)
						existing.Context = append(existing.Context, contextSnippet)
					}
				} else {
					entities[key] = &ConversationEntity{
						Name:     match,
						Type:     entityType,
						Mentions: 1,
						Context:  []string{s.extractContext(msg.Content, match, 50)},
						Related:  make([]string, 0),
					}
					_ = idx // Can be used for position tracking
				}
			}
		}
	}

	// Convert to slice and sort by mentions
	result := make([]ConversationEntity, 0, len(entities))
	for _, e := range entities {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Mentions > result[j].Mentions
	})

	// Limit to top entities
	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

// extractContext extracts surrounding context for an entity mention
func (s *ConversationIntelligenceService) extractContext(text, entity string, chars int) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(entity))
	if idx == -1 {
		return ""
	}

	start := idx - chars
	if start < 0 {
		start = 0
	}
	end := idx + len(entity) + chars
	if end > len(text) {
		end = len(text)
	}

	return strings.TrimSpace(text[start:end])
}

// extractQuestions extracts questions from the conversation
func (s *ConversationIntelligenceService) extractQuestions(messages []Message) []Question {
	questions := make([]Question, 0)

	questionPattern := regexp.MustCompile(`[^.!?]*\?`)

	for idx, msg := range messages {
		matches := questionPattern.FindAllString(msg.Content, -1)
		for _, match := range matches {
			match = strings.TrimSpace(match)
			if len(match) < 10 { // Skip very short questions
				continue
			}

			q := Question{
				Text:         match,
				AskedBy:      msg.Role,
				MessageIndex: idx,
				Answered:     false,
			}

			// Check if question was answered in subsequent messages
			if idx < len(messages)-1 {
				for i := idx + 1; i < len(messages) && i <= idx+3; i++ {
					if messages[i].Role != msg.Role && len(messages[i].Content) > 20 {
						q.Answered = true
						q.Answer = s.truncateText(messages[i].Content, 200)
						break
					}
				}
			}

			questions = append(questions, q)
		}
	}

	return questions
}

// extractActionItems extracts action items from messages
func (s *ConversationIntelligenceService) extractActionItems(messages []Message) []ActionItem {
	actionItems := make([]ActionItem, 0)

	// Patterns that indicate action items
	actionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:need to|should|must|have to|going to|will|todo|task:|action:)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:please|could you|can you)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)- \[ \]\s+([^\n]+)`), // Markdown task
		regexp.MustCompile(`(?i)(?:implement|create|add|fix|update|remove|delete|refactor)\s+([^.!?\n]+)`),
	}

	for idx, msg := range messages {
		for _, pattern := range actionPatterns {
			matches := pattern.FindAllStringSubmatch(msg.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					description := strings.TrimSpace(match[1])
					if len(description) < 5 || len(description) > 200 {
						continue
					}

					actionItems = append(actionItems, ActionItem{
						ID:           uuid.New().String(),
						Description:  description,
						Priority:     s.inferPriority(description),
						Status:       "pending",
						MessageIndex: idx,
						Tags:         s.extractTags(description),
					})
				}
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := make([]ActionItem, 0)
	for _, item := range actionItems {
		key := strings.ToLower(item.Description)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}

	return unique
}

// inferPriority infers priority from action item description
func (s *ConversationIntelligenceService) inferPriority(description string) string {
	lower := strings.ToLower(description)

	highPriorityWords := []string{"urgent", "critical", "asap", "immediately", "important", "must", "blocker"}
	for _, word := range highPriorityWords {
		if strings.Contains(lower, word) {
			return "high"
		}
	}

	lowPriorityWords := []string{"maybe", "consider", "could", "nice to have", "eventually", "later"}
	for _, word := range lowPriorityWords {
		if strings.Contains(lower, word) {
			return "low"
		}
	}

	return "medium"
}

// extractTags extracts tags from description
func (s *ConversationIntelligenceService) extractTags(description string) []string {
	tags := make([]string, 0)

	// Common categories
	categories := map[string][]string{
		"bug":         {"fix", "bug", "error", "issue", "broken"},
		"feature":     {"add", "implement", "create", "new", "feature"},
		"refactor":    {"refactor", "clean", "improve", "optimize"},
		"docs":        {"document", "readme", "comment", "docs"},
		"test":        {"test", "spec", "coverage"},
		"security":    {"security", "auth", "permission", "vulnerability"},
		"performance": {"performance", "speed", "optimize", "cache"},
	}

	lower := strings.ToLower(description)
	for tag, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}

// extractDecisions extracts decisions from the conversation
func (s *ConversationIntelligenceService) extractDecisions(messages []Message) []ConversationDecision {
	decisions := make([]ConversationDecision, 0)

	decisionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:decided to|will use|going with|chose|selected|opted for)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:let's|we should|we'll)\s+(?:go with|use|implement)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the (?:best|better) (?:approach|solution|option) is)\s+([^.!?\n]+)`),
	}

	for idx, msg := range messages {
		for _, pattern := range decisionPatterns {
			matches := pattern.FindAllStringSubmatch(msg.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					decisions = append(decisions, ConversationDecision{
						Description:  strings.TrimSpace(match[1]),
						Context:      s.truncateText(msg.Content, 200),
						MessageIndex: idx,
					})
				}
			}
		}
	}

	return decisions
}

// extractCodeMentions extracts code snippets and references
func (s *ConversationIntelligenceService) extractCodeMentions(messages []Message) []CodeMention {
	mentions := make([]CodeMention, 0)

	codeBlockPattern := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")
	filePathPattern := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])([\w/-]+\.(go|ts|js|svelte|py|sql|json|yaml|yml|md|css|html|tsx|jsx))`)

	for idx, msg := range messages {
		// Extract code blocks
		codeMatches := codeBlockPattern.FindAllStringSubmatch(msg.Content, -1)
		for _, match := range codeMatches {
			if len(match) > 2 {
				mentions = append(mentions, CodeMention{
					Language:     match[1],
					Snippet:      s.truncateText(match[2], 500),
					Context:      "code block",
					MessageIndex: idx,
				})
			}
		}

		// Extract file references
		fileMatches := filePathPattern.FindAllStringSubmatch(msg.Content, -1)
		for _, match := range fileMatches {
			if len(match) > 1 {
				mentions = append(mentions, CodeMention{
					FilePath:     match[1],
					Context:      s.extractContext(msg.Content, match[1], 100),
					MessageIndex: idx,
				})
			}
		}
	}

	return mentions
}

// analyzeSentiment performs sentiment analysis on messages
func (s *ConversationIntelligenceService) analyzeSentiment(messages []Message) SentimentAnalysis {
	result := SentimentAnalysis{
		Progression: make([]SentimentPoint, 0),
		Highlights:  make([]SentimentHighlight, 0),
	}

	// Sentiment word lists
	positiveWords := map[string]float64{
		"great": 0.8, "good": 0.6, "excellent": 0.9, "perfect": 1.0,
		"thanks": 0.5, "thank": 0.5, "helpful": 0.7, "awesome": 0.9,
		"love": 0.8, "nice": 0.5, "wonderful": 0.8, "amazing": 0.9,
		"works": 0.6, "working": 0.5, "solved": 0.7, "fixed": 0.7,
		"yes": 0.3, "correct": 0.5, "right": 0.4, "exactly": 0.6,
	}

	negativeWords := map[string]float64{
		"bad": -0.6, "wrong": -0.5, "error": -0.4, "fail": -0.6,
		"failed": -0.6, "broken": -0.7, "issue": -0.3, "problem": -0.4,
		"bug": -0.3, "not working": -0.6, "doesn't work": -0.7,
		"confused": -0.4, "frustrating": -0.7, "annoying": -0.6,
		"hate": -0.8, "terrible": -0.9, "awful": -0.8, "worst": -0.9,
		"unfortunately": -0.3, "sadly": -0.4, "sorry": -0.2,
	}

	totalScore := 0.0
	messageScores := make([]float64, len(messages))

	for idx, msg := range messages {
		if msg.Role == "system" {
			continue
		}

		words := strings.Fields(strings.ToLower(msg.Content))
		score := 0.0
		wordCount := 0

		for _, word := range words {
			word = strings.Trim(word, ".,!?;:\"'")
			if val, ok := positiveWords[word]; ok {
				score += val
				wordCount++
			}
			if val, ok := negativeWords[word]; ok {
				score += val
				wordCount++
			}
		}

		if wordCount > 0 {
			score = score / float64(wordCount)
		}

		messageScores[idx] = score
		totalScore += score

		sentiment := "neutral"
		if score > 0.2 {
			sentiment = "positive"
		} else if score < -0.2 {
			sentiment = "negative"
		}

		result.Progression = append(result.Progression, SentimentPoint{
			MessageIndex: idx,
			Sentiment:    sentiment,
			Score:        score,
		})

		// Add highlights for strong sentiments
		if score > 0.5 || score < -0.5 {
			result.Highlights = append(result.Highlights, SentimentHighlight{
				MessageIndex: idx,
				Text:         s.truncateText(msg.Content, 100),
				Sentiment:    sentiment,
				Reason:       fmt.Sprintf("Strong %s sentiment detected", sentiment),
			})
		}
	}

	// Calculate overall sentiment
	avgScore := 0.0
	if len(messages) > 0 {
		avgScore = totalScore / float64(len(messages))
	}

	result.Score = avgScore
	if avgScore > 0.2 {
		result.Overall = "positive"
	} else if avgScore < -0.2 {
		result.Overall = "negative"
	} else if len(result.Highlights) > 2 {
		result.Overall = "mixed"
	} else {
		result.Overall = "neutral"
	}

	return result
}

// generateTitle generates a title for the conversation
func (s *ConversationIntelligenceService) generateTitle(messages []Message, topics []ConversationTopic) string {
	// Use first user message as basis
	for _, msg := range messages {
		if msg.Role == "user" && len(msg.Content) > 10 {
			// Extract first sentence or line
			content := msg.Content
			if idx := strings.IndexAny(content, ".!?\n"); idx > 0 && idx < 100 {
				content = content[:idx]
			}
			if len(content) > 60 {
				content = content[:60] + "..."
			}
			return strings.TrimSpace(content)
		}
	}

	// Fall back to topics
	if len(topics) > 0 {
		return fmt.Sprintf("Discussion about %s", topics[0].Name)
	}

	return "Untitled Conversation"
}

// generateSummary generates a summary of the conversation
func (s *ConversationIntelligenceService) generateSummary(messages []Message, analysis *ConversationAnalysis) string {
	var sb strings.Builder

	// Opening
	sb.WriteString(fmt.Sprintf("A conversation with %d messages", len(messages)))
	if len(analysis.Topics) > 0 {
		topicNames := make([]string, 0)
		for _, t := range analysis.Topics {
			topicNames = append(topicNames, t.Name)
		}
		sb.WriteString(fmt.Sprintf(" discussing %s", strings.Join(topicNames, ", ")))
	}
	sb.WriteString(". ")

	// Key activities
	if len(analysis.CodeMentions) > 0 {
		sb.WriteString(fmt.Sprintf("Code was discussed in %d instances. ", len(analysis.CodeMentions)))
	}
	if len(analysis.ActionItems) > 0 {
		sb.WriteString(fmt.Sprintf("%d action items were identified. ", len(analysis.ActionItems)))
	}
	if len(analysis.Decisions) > 0 {
		sb.WriteString(fmt.Sprintf("%d decisions were made. ", len(analysis.Decisions)))
	}
	if len(analysis.Questions) > 0 {
		answered := 0
		for _, q := range analysis.Questions {
			if q.Answered {
				answered++
			}
		}
		sb.WriteString(fmt.Sprintf("%d of %d questions were answered. ", answered, len(analysis.Questions)))
	}

	// Sentiment
	sb.WriteString(fmt.Sprintf("Overall sentiment was %s.", analysis.Sentiment.Overall))

	return sb.String()
}

// extractKeyPoints extracts key points from the conversation
func (s *ConversationIntelligenceService) extractKeyPoints(messages []Message, analysis *ConversationAnalysis) []string {
	points := make([]string, 0)

	// Add decisions as key points
	for _, d := range analysis.Decisions {
		if len(d.Description) > 10 {
			points = append(points, "Decision: "+d.Description)
		}
	}

	// Add high-priority action items
	for _, a := range analysis.ActionItems {
		if a.Priority == "high" {
			points = append(points, "Action: "+a.Description)
		}
	}

	// Limit to 5 key points
	if len(points) > 5 {
		points = points[:5]
	}

	return points
}

// truncateText truncates text to maxLen characters
func (s *ConversationIntelligenceService) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// saveAnalysis saves the conversation analysis to the database
func (s *ConversationIntelligenceService) saveAnalysis(ctx context.Context, analysis *ConversationAnalysis) error {
	// conversation_summaries has a mix of TEXT[] (for lightweight context injection)
	// and JSONB columns (for richer structured data). Store both where possible.

	// TEXT[] projections
	topicNames := make([]string, 0, len(analysis.Topics))
	for _, t := range analysis.Topics {
		if t.Name != "" {
			topicNames = append(topicNames, t.Name)
		}
	}

	actionItemTexts := make([]string, 0, len(analysis.ActionItems))
	for _, a := range analysis.ActionItems {
		if strings.TrimSpace(a.Description) == "" {
			continue
		}
		actionItemTexts = append(actionItemTexts, a.Description)
	}

	questionTexts := make([]string, 0, len(analysis.Questions))
	for _, q := range analysis.Questions {
		if strings.TrimSpace(q.Text) == "" {
			continue
		}
		questionTexts = append(questionTexts, q.Text)
	}

	decisionTexts := make([]string, 0, len(analysis.Decisions))
	for _, d := range analysis.Decisions {
		if strings.TrimSpace(d.Description) == "" {
			continue
		}
		decisionTexts = append(decisionTexts, d.Description)
	}

	// Mentioned entities summary for the legacy mentioned_entities column.
	mentioned := map[string][]string{}
	for _, e := range analysis.Entities {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(e.Type))
		switch k {
		case "person", "people":
			mentioned["people"] = append(mentioned["people"], name)
		case "project", "projects":
			mentioned["projects"] = append(mentioned["projects"], name)
		case "client", "clients", "organization", "org":
			mentioned["clients"] = append(mentioned["clients"], name)
		case "task", "tasks":
			mentioned["tasks"] = append(mentioned["tasks"], name)
		default:
			mentioned["other"] = append(mentioned["other"], name)
		}
	}

	// JSONB payloads
	sentimentJSON, _ := json.Marshal(analysis.Sentiment)
	entitiesJSON, _ := json.Marshal(analysis.Entities)
	codeMentionsJSON, _ := json.Marshal(analysis.CodeMentions)
	mentionedJSON, _ := json.Marshal(mentioned)

	// Keep structured details in metadata.
	if analysis.Metadata == nil {
		analysis.Metadata = make(map[string]interface{})
	}
	analysis.Metadata["topics_detail"] = analysis.Topics
	analysis.Metadata["action_items_detail"] = analysis.ActionItems
	analysis.Metadata["questions_detail"] = analysis.Questions
	analysis.Metadata["decisions_detail"] = analysis.Decisions
	metadataJSON, _ := json.Marshal(analysis.Metadata)

	// Best-effort embedding of the summary for semantic search.
	var embedding any = nil
	if s.embed != nil {
		embedText := strings.TrimSpace(strings.Join([]string{
			analysis.Title,
			analysis.Summary,
			"Key points: " + strings.Join(analysis.KeyPoints, "; "),
			"Topics: " + strings.Join(topicNames, ", "),
		}, "\n"))
		if embedText != "" {
			if v, err := s.embed.GenerateEmbedding(ctx, embedText); err == nil {
				embedding = pgvector.NewVector(v)
			}
		}
	}

	// Time range (if available)
	var timeStart, timeEnd *time.Time
	if v, ok := analysis.Metadata["time_range_start"].(time.Time); ok {
		timeStart = &v
	}
	if v, ok := analysis.Metadata["time_range_end"].(time.Time); ok {
		timeEnd = &v
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO conversation_summaries
		 (id, conversation_id, user_id,
		  title, summary,
		  key_points, topics,
		  sentiment, entities, mentioned_entities,
		  action_items, questions,
		  decisions, decisions_made,
		  code_mentions,
		  embedding,
		  message_count, token_count, duration,
		  time_range_start, time_range_end,
		  metadata,
		  summarized_at,
		  created_at, updated_at)
		 VALUES ($1, $2, $3,
		         $4, $5,
		         $6, $7,
		         $8, $9, $10,
		         $11, $12,
		         $13, $14,
		         $15,
		         $16,
		         $17, $18, $19,
		         $20, $21,
		         $22,
		         NOW(),
		         $23, $24)
		 ON CONFLICT (conversation_id) DO UPDATE SET
		    title = EXCLUDED.title,
		    summary = EXCLUDED.summary,
		    key_points = EXCLUDED.key_points,
		    topics = EXCLUDED.topics,
		    sentiment = EXCLUDED.sentiment,
		    entities = EXCLUDED.entities,
		    mentioned_entities = EXCLUDED.mentioned_entities,
		    action_items = EXCLUDED.action_items,
		    questions = EXCLUDED.questions,
		    decisions = EXCLUDED.decisions,
		    decisions_made = EXCLUDED.decisions_made,
		    code_mentions = EXCLUDED.code_mentions,
		    embedding = EXCLUDED.embedding,
		    message_count = EXCLUDED.message_count,
		    token_count = EXCLUDED.token_count,
		    duration = EXCLUDED.duration,
		    time_range_start = EXCLUDED.time_range_start,
		    time_range_end = EXCLUDED.time_range_end,
		    metadata = EXCLUDED.metadata,
		    summarized_at = NOW(),
		    updated_at = EXCLUDED.updated_at`,
		analysis.ID, analysis.ConversationID, analysis.UserID,
		analysis.Title, analysis.Summary,
		analysis.KeyPoints, topicNames,
		sentimentJSON, entitiesJSON, mentionedJSON,
		actionItemTexts, questionTexts,
		decisionTexts, decisionTexts,
		codeMentionsJSON,
		embedding,
		analysis.MessageCount, analysis.TokenCount, analysis.Duration,
		timeStart, timeEnd,
		metadataJSON,
		analysis.CreatedAt, analysis.UpdatedAt)

	return err
}

// GetAnalysis retrieves a conversation analysis
func (s *ConversationIntelligenceService) GetAnalysis(ctx context.Context, conversationID string) (*ConversationAnalysis, error) {
	var analysis ConversationAnalysis
	var keyPoints []string
	var topics []string
	var actionItems []string
	var questions []string
	var decisions []string
	var sentimentJSON, entitiesJSON []byte
	var codeMentionsJSON, metadataJSON []byte

	err := s.pool.QueryRow(ctx,
		`SELECT id, conversation_id, user_id, title, summary,
		        key_points, topics,
		        sentiment, entities,
		        action_items, questions, decisions,
		        code_mentions,
		        message_count, token_count, duration,
		        metadata,
		        created_at, updated_at
		 FROM conversation_summaries
		 WHERE conversation_id = $1`,
		conversationID).Scan(
		&analysis.ID, &analysis.ConversationID, &analysis.UserID, &analysis.Title, &analysis.Summary,
		&keyPoints, &topics,
		&sentimentJSON, &entitiesJSON,
		&actionItems, &questions, &decisions,
		&codeMentionsJSON,
		&analysis.MessageCount, &analysis.TokenCount, &analysis.Duration,
		&metadataJSON,
		&analysis.CreatedAt, &analysis.UpdatedAt)

	if err != nil {
		return nil, err
	}

	analysis.KeyPoints = keyPoints
	json.Unmarshal(sentimentJSON, &analysis.Sentiment)
	json.Unmarshal(entitiesJSON, &analysis.Entities)
	json.Unmarshal(codeMentionsJSON, &analysis.CodeMentions)
	json.Unmarshal(metadataJSON, &analysis.Metadata)

	// Rehydrate structured fields best-effort from metadata; fallback to simple text lists.
	if analysis.Metadata != nil {
		if raw, ok := analysis.Metadata["topics_detail"]; ok {
			if b, err := json.Marshal(raw); err == nil {
				var detailed []ConversationTopic
				if json.Unmarshal(b, &detailed) == nil {
					analysis.Topics = detailed
				}
			}
		}
		if raw, ok := analysis.Metadata["action_items_detail"]; ok {
			if b, err := json.Marshal(raw); err == nil {
				var detailed []ActionItem
				if json.Unmarshal(b, &detailed) == nil {
					analysis.ActionItems = detailed
				}
			}
		}
		if raw, ok := analysis.Metadata["questions_detail"]; ok {
			if b, err := json.Marshal(raw); err == nil {
				var detailed []Question
				if json.Unmarshal(b, &detailed) == nil {
					analysis.Questions = detailed
				}
			}
		}
		if raw, ok := analysis.Metadata["decisions_detail"]; ok {
			if b, err := json.Marshal(raw); err == nil {
				var detailed []ConversationDecision
				if json.Unmarshal(b, &detailed) == nil {
					analysis.Decisions = detailed
				}
			}
		}
	}

	if len(analysis.Topics) == 0 {
		for _, name := range topics {
			analysis.Topics = append(analysis.Topics, ConversationTopic{Name: name, Confidence: 0})
		}
	}
	if len(analysis.ActionItems) == 0 {
		for _, t := range actionItems {
			analysis.ActionItems = append(analysis.ActionItems, ActionItem{Description: t, Priority: "", Status: "pending"})
		}
	}
	if len(analysis.Questions) == 0 {
		for _, t := range questions {
			analysis.Questions = append(analysis.Questions, Question{Text: t, Answered: false})
		}
	}
	if len(analysis.Decisions) == 0 {
		for _, t := range decisions {
			analysis.Decisions = append(analysis.Decisions, ConversationDecision{Description: t})
		}
	}

	return &analysis, nil
}

// SearchConversations searches conversation analyses
func (s *ConversationIntelligenceService) SearchConversations(ctx context.Context, userID, query string, limit int) ([]ConversationAnalysis, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, conversation_id, user_id, title, summary, key_points, topics, sentiment,
		        entities, action_items, message_count, token_count, duration, created_at
		 FROM conversation_summaries
		 WHERE user_id = $1 AND (
		    title ILIKE $2 OR summary ILIKE $2
		 )
		 ORDER BY created_at DESC
		 LIMIT $3`,
		userID, "%"+query+"%", limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analyses := make([]ConversationAnalysis, 0)
	for rows.Next() {
		var a ConversationAnalysis
		var keyPoints []string
		var topics []string
		var actionItems []string
		var sentimentJSON, entitiesJSON []byte

		err := rows.Scan(&a.ID, &a.ConversationID, &a.UserID, &a.Title, &a.Summary,
			&keyPoints, &topics, &sentimentJSON, &entitiesJSON, &actionItems,
			&a.MessageCount, &a.TokenCount, &a.Duration, &a.CreatedAt)
		if err != nil {
			continue
		}

		a.KeyPoints = keyPoints
		for _, name := range topics {
			a.Topics = append(a.Topics, ConversationTopic{Name: name, Confidence: 0})
		}
		json.Unmarshal(sentimentJSON, &a.Sentiment)
		json.Unmarshal(entitiesJSON, &a.Entities)
		for _, t := range actionItems {
			a.ActionItems = append(a.ActionItems, ActionItem{Description: t, Status: "pending"})
		}

		analyses = append(analyses, a)
	}

	return analyses, nil
}
