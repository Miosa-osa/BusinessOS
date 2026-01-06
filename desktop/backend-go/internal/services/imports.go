package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =============================================================================
// IMPORT SERVICE
// =============================================================================

type ImportService struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewImportService(pool *pgxpool.Pool) *ImportService {
	return &ImportService{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// =============================================================================
// CHATGPT EXPORT FORMAT
// =============================================================================

// ChatGPT exports conversations.json with this structure
type ChatGPTExport struct {
	Conversations []ChatGPTConversation `json:"conversations"`
}

type ChatGPTConversation struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	CreateTime  float64                `json:"create_time"` // Unix timestamp (can be float)
	UpdateTime  float64                `json:"update_time"`
	Mapping     map[string]ChatGPTNode `json:"mapping"`
	CurrentNode string                 `json:"current_node,omitempty"`
}

type ChatGPTNode struct {
	ID       string          `json:"id"`
	Message  *ChatGPTMessage `json:"message,omitempty"`
	Parent   *string         `json:"parent"`
	Children []string        `json:"children"`
}

type ChatGPTMessage struct {
	ID         string         `json:"id"`
	Author     ChatGPTAuthor  `json:"author"`
	CreateTime *float64       `json:"create_time"`
	Content    ChatGPTContent `json:"content"`
	Status     string         `json:"status,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type ChatGPTAuthor struct {
	Role     string         `json:"role"` // "user", "assistant", "system", "tool"
	Name     *string        `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ChatGPTContent struct {
	ContentType string `json:"content_type"` // "text", "code", etc.
	Parts       []any  `json:"parts"`        // Usually strings, but can be other types
}

// =============================================================================
// CLAUDE EXPORT FORMAT
// =============================================================================

// Claude exports conversations.json with this structure
type ClaudeExport struct {
	Conversations []ClaudeConversation `json:"conversations"`
}

type ClaudeConversation struct {
	UUID         string          `json:"uuid"`
	Name         string          `json:"name"`
	CreatedAt    string          `json:"created_at"` // ISO8601
	UpdatedAt    string          `json:"updated_at"`
	ChatMessages []ClaudeMessage `json:"chat_messages"`
	Model        string          `json:"model,omitempty"`
	Project      *ClaudeProject  `json:"project,omitempty"`
}

type ClaudeMessage struct {
	UUID      string       `json:"uuid"`
	Text      string       `json:"text"`
	Sender    string       `json:"sender"` // "human", "assistant"
	CreatedAt string       `json:"created_at,omitempty"`
	Files     []ClaudeFile `json:"files,omitempty"`
}

type ClaudeFile struct {
	FileName         string `json:"file_name"`
	FileType         string `json:"file_type"`
	FileSize         int64  `json:"file_size,omitempty"`
	ExtractedContent string `json:"extracted_content,omitempty"`
}

type ClaudeProject struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// =============================================================================
// NORMALIZED MESSAGE FORMAT
// =============================================================================

// NormalizedMessage is the standardized format we store in the database
type NormalizedMessage struct {
	Role      string         `json:"role"` // "user", "assistant", "system"
	Content   string         `json:"content"`
	Timestamp *time.Time     `json:"timestamp,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// =============================================================================
// IMPORT JOB INPUT/OUTPUT
// =============================================================================

type CreateImportJobInput struct {
	UserID           string
	SourceType       sqlc.ImportSourceType
	OriginalFilename string
	FileSizeBytes    int64
	ContentType      string
	TargetModule     string
	ImportOptions    map[string]any
}

type ImportProgress struct {
	TotalRecords     int
	ProcessedRecords int
	ImportedRecords  int
	SkippedRecords   int
	FailedRecords    int
	ProgressPercent  int
}

type ImportResult struct {
	JobID           uuid.UUID
	Status          sqlc.ImportStatus
	TotalRecords    int
	ImportedRecords int
	SkippedRecords  int
	FailedRecords   int
	Errors          []ImportError
}

type ImportError struct {
	RecordIndex int    `json:"record_index"`
	ExternalID  string `json:"external_id,omitempty"`
	Error       string `json:"error"`
}

// =============================================================================
// POINTER HELPERS
// =============================================================================

func ptr[T any](v T) *T {
	return &v
}

func ptrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// =============================================================================
// IMPORT JOB METHODS
// =============================================================================

// CreateImportJob creates a new import job and returns its ID
func (s *ImportService) CreateImportJob(ctx context.Context, input CreateImportJobInput) (*sqlc.ImportJob, error) {
	optionsJSON, err := json.Marshal(input.ImportOptions)
	if err != nil {
		optionsJSON = []byte("{}")
	}

	var fileSizePtr *int64
	if input.FileSizeBytes > 0 {
		fileSizePtr = &input.FileSizeBytes
	}

	job, err := s.queries.CreateImportJob(ctx, sqlc.CreateImportJobParams{
		UserID:           input.UserID,
		SourceType:       input.SourceType,
		OriginalFilename: ptrIfNotEmpty(input.OriginalFilename),
		FileSizeBytes:    fileSizePtr,
		ContentType:      ptrIfNotEmpty(input.ContentType),
		TargetModule:     input.TargetModule,
		TargetEntity:     nil,
		FieldMapping:     []byte("{}"),
		TransformRules:   []byte("{}"),
		ImportOptions:    optionsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	return &job, nil
}

// GetImportJob retrieves an import job by ID
func (s *ImportService) GetImportJob(ctx context.Context, userID string, jobID uuid.UUID) (*sqlc.ImportJob, error) {
	job, err := s.queries.GetImportJob(ctx, sqlc.GetImportJobParams{
		ID:     pgtype.UUID{Bytes: jobID, Valid: true},
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetUserImportJobs retrieves import jobs for a user with pagination
func (s *ImportService) GetUserImportJobs(ctx context.Context, userID string, limit, offset int32) ([]sqlc.ImportJob, error) {
	return s.queries.GetImportJobsByUser(ctx, sqlc.GetImportJobsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

// UpdateImportProgress updates the progress of an import job
func (s *ImportService) UpdateImportProgress(ctx context.Context, userID string, jobID uuid.UUID, progress ImportProgress) error {
	return s.queries.UpdateImportJobProgress(ctx, sqlc.UpdateImportJobProgressParams{
		ID:               pgtype.UUID{Bytes: jobID, Valid: true},
		UserID:           userID,
		ProgressPercent:  ptr(int32(progress.ProgressPercent)),
		ProcessedRecords: ptr(int32(progress.ProcessedRecords)),
		ImportedRecords:  ptr(int32(progress.ImportedRecords)),
		SkippedRecords:   ptr(int32(progress.SkippedRecords)),
		FailedRecords:    ptr(int32(progress.FailedRecords)),
	})
}

// FailImportJob marks an import job as failed
func (s *ImportService) FailImportJob(ctx context.Context, userID string, jobID uuid.UUID, errMsg string, details map[string]any) error {
	detailsJSON, _ := json.Marshal(details)
	return s.queries.FailImportJob(ctx, sqlc.FailImportJobParams{
		ID:           pgtype.UUID{Bytes: jobID, Valid: true},
		UserID:       userID,
		ErrorMessage: ptrIfNotEmpty(errMsg),
		ErrorDetails: detailsJSON,
	})
}

// CompleteImportJob marks an import job as completed
func (s *ImportService) CompleteImportJob(ctx context.Context, userID string, jobID uuid.UUID, summary map[string]any) error {
	summaryJSON, _ := json.Marshal(summary)
	return s.queries.CompleteImportJob(ctx, sqlc.CompleteImportJobParams{
		ID:            pgtype.UUID{Bytes: jobID, Valid: true},
		UserID:        userID,
		ResultSummary: summaryJSON,
	})
}

// =============================================================================
// CHATGPT IMPORT
// =============================================================================

// ImportChatGPTConversations imports conversations from a ChatGPT export file
func (s *ImportService) ImportChatGPTConversations(ctx context.Context, userID string, reader io.Reader, filename string) (*ImportResult, error) {
	// Parse the export file
	var export ChatGPTExport
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to parse ChatGPT export: %w", err)
	}

	// Create import job
	job, err := s.CreateImportJob(ctx, CreateImportJobInput{
		UserID:           userID,
		SourceType:       sqlc.ImportSourceTypeChatgptExport,
		OriginalFilename: filename,
		TargetModule:     "conversations",
		ImportOptions:    map[string]any{"format": "chatgpt", "version": "1.0"},
	})
	if err != nil {
		return nil, err
	}

	jobID, _ := uuid.FromBytes(job.ID.Bytes[:])
	totalRecords := len(export.Conversations)

	// Update total records
	s.queries.UpdateImportJobTotalRecords(ctx, sqlc.UpdateImportJobTotalRecordsParams{
		ID:           job.ID,
		UserID:       userID,
		TotalRecords: ptr(int32(totalRecords)),
	})

	// Update status to processing
	s.queries.UpdateImportJobStatus(ctx, sqlc.UpdateImportJobStatusParams{
		ID:              job.ID,
		UserID:          userID,
		Status:          sqlc.NullImportStatus{ImportStatus: sqlc.ImportStatusProcessing, Valid: true},
		ProgressPercent: ptr(int32(0)),
	})

	result := &ImportResult{
		JobID:  jobID,
		Status: sqlc.ImportStatusProcessing,
	}
	result.TotalRecords = totalRecords
	var errors []ImportError

	// Process each conversation
	for i, conv := range export.Conversations {
		// Check for duplicate
		exists, _ := s.queries.CheckExternalRecordExists(ctx, sqlc.CheckExternalRecordExistsParams{
			UserID:     userID,
			SourceType: sqlc.ImportSourceTypeChatgptExport,
			ExternalID: conv.ID,
		})
		if exists {
			result.SkippedRecords++
			continue
		}

		// Convert ChatGPT messages to normalized format
		messages, model := s.parseChatGPTMessages(conv)
		messagesJSON, _ := json.Marshal(messages)

		// Build search content
		searchContent := s.buildSearchContent(conv.Title, messages)

		// Calculate data hash for deduplication
		dataHash := s.hashData(messagesJSON)

		// Create imported conversation
		imported, err := s.queries.CreateImportedConversation(ctx, sqlc.CreateImportedConversationParams{
			UserID:                 userID,
			ImportJobID:            job.ID,
			SourceType:             sqlc.ImportSourceTypeChatgptExport,
			ExternalConversationID: ptrIfNotEmpty(conv.ID),
			Title:                  ptrIfNotEmpty(conv.Title),
			Model:                  ptrIfNotEmpty(model),
			Messages:               messagesJSON,
			MessageCount:           ptr(int32(len(messages))),
			OriginalCreatedAt:      s.unixToTimestamp(conv.CreateTime),
			OriginalUpdatedAt:      s.unixToTimestamp(conv.UpdateTime),
			Metadata:               []byte("{}"),
			SearchContent:          ptrIfNotEmpty(searchContent),
			Tags:                   []string{},
		})
		if err != nil {
			errors = append(errors, ImportError{
				RecordIndex: i,
				ExternalID:  conv.ID,
				Error:       err.Error(),
			})
			result.FailedRecords++
			continue
		}

		// Track the imported record for deduplication
		s.queries.CreateImportedRecord(ctx, sqlc.CreateImportedRecordParams{
			UserID:           userID,
			ImportJobID:      job.ID,
			SourceType:       sqlc.ImportSourceTypeChatgptExport,
			ExternalID:       conv.ID,
			TargetModule:     "conversations",
			TargetRecordID:   imported.ID,
			ExternalDataHash: ptrIfNotEmpty(dataHash),
		})

		result.ImportedRecords++

		// Update progress every 10 records
		if (i+1)%10 == 0 || i == totalRecords-1 {
			progress := int((float64(i+1) / float64(totalRecords)) * 100)
			s.UpdateImportProgress(ctx, userID, jobID, ImportProgress{
				TotalRecords:     totalRecords,
				ProcessedRecords: i + 1,
				ImportedRecords:  result.ImportedRecords,
				SkippedRecords:   result.SkippedRecords,
				FailedRecords:    result.FailedRecords,
				ProgressPercent:  progress,
			})
		}
	}

	// Complete the job
	result.Errors = errors
	if result.FailedRecords > 0 && result.ImportedRecords == 0 {
		result.Status = sqlc.ImportStatusFailed
		s.FailImportJob(ctx, userID, jobID, "All records failed to import", map[string]any{
			"errors": errors,
		})
	} else {
		result.Status = sqlc.ImportStatusCompleted
		s.CompleteImportJob(ctx, userID, jobID, map[string]any{
			"imported":    result.ImportedRecords,
			"skipped":     result.SkippedRecords,
			"failed":      result.FailedRecords,
			"error_count": len(errors),
		})
	}

	return result, nil
}

// parseChatGPTMessages extracts messages from ChatGPT's node mapping
func (s *ImportService) parseChatGPTMessages(conv ChatGPTConversation) ([]NormalizedMessage, string) {
	var messages []NormalizedMessage
	var model string

	// Find the root node and traverse the conversation
	// ChatGPT stores messages in a tree structure via parent/children references
	visited := make(map[string]bool)

	// Build ordered message list by traversing from root
	var traverseNode func(nodeID string)
	traverseNode = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true

		node, ok := conv.Mapping[nodeID]
		if !ok || node.Message == nil {
			// Continue to children even if this node has no message
			for _, childID := range node.Children {
				traverseNode(childID)
			}
			return
		}

		msg := node.Message

		// Skip system messages and tool messages typically
		if msg.Author.Role == "user" || msg.Author.Role == "assistant" {
			content := s.extractChatGPTContent(msg.Content)
			if content != "" {
				normalized := NormalizedMessage{
					Role:    msg.Author.Role,
					Content: content,
				}

				if msg.CreateTime != nil {
					t := time.Unix(int64(*msg.CreateTime), 0)
					normalized.Timestamp = &t
				}

				// Extract model from metadata if available
				if msg.Metadata != nil {
					if m, ok := msg.Metadata["model_slug"].(string); ok && model == "" {
						model = m
					}
				}

				messages = append(messages, normalized)
			}
		}

		// Process children in order
		for _, childID := range node.Children {
			traverseNode(childID)
		}
	}

	// Find root node (node with no parent)
	for nodeID, node := range conv.Mapping {
		if node.Parent == nil {
			traverseNode(nodeID)
			break
		}
	}

	return messages, model
}

// extractChatGPTContent extracts text content from ChatGPT's content structure
func (s *ImportService) extractChatGPTContent(content ChatGPTContent) string {
	var parts []string
	for _, part := range content.Parts {
		switch v := part.(type) {
		case string:
			if v != "" {
				parts = append(parts, v)
			}
		case map[string]any:
			// Handle structured content (e.g., images, code blocks)
			if text, ok := v["text"].(string); ok {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}

// =============================================================================
// CLAUDE IMPORT
// =============================================================================

// ImportClaudeConversations imports conversations from a Claude export file
func (s *ImportService) ImportClaudeConversations(ctx context.Context, userID string, reader io.Reader, filename string) (*ImportResult, error) {
	// Parse the export file
	var export ClaudeExport
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to parse Claude export: %w", err)
	}

	// Create import job
	job, err := s.CreateImportJob(ctx, CreateImportJobInput{
		UserID:           userID,
		SourceType:       sqlc.ImportSourceTypeClaudeExport,
		OriginalFilename: filename,
		TargetModule:     "conversations",
		ImportOptions:    map[string]any{"format": "claude", "version": "1.0"},
	})
	if err != nil {
		return nil, err
	}

	jobID, _ := uuid.FromBytes(job.ID.Bytes[:])
	totalRecords := len(export.Conversations)

	// Update total records
	s.queries.UpdateImportJobTotalRecords(ctx, sqlc.UpdateImportJobTotalRecordsParams{
		ID:           job.ID,
		UserID:       userID,
		TotalRecords: ptr(int32(totalRecords)),
	})

	// Update status to processing
	s.queries.UpdateImportJobStatus(ctx, sqlc.UpdateImportJobStatusParams{
		ID:              job.ID,
		UserID:          userID,
		Status:          sqlc.NullImportStatus{ImportStatus: sqlc.ImportStatusProcessing, Valid: true},
		ProgressPercent: ptr(int32(0)),
	})

	result := &ImportResult{
		JobID:  jobID,
		Status: sqlc.ImportStatusProcessing,
	}
	result.TotalRecords = totalRecords
	var errors []ImportError

	// Process each conversation
	for i, conv := range export.Conversations {
		// Check for duplicate
		exists, _ := s.queries.CheckExternalRecordExists(ctx, sqlc.CheckExternalRecordExistsParams{
			UserID:     userID,
			SourceType: sqlc.ImportSourceTypeClaudeExport,
			ExternalID: conv.UUID,
		})
		if exists {
			result.SkippedRecords++
			continue
		}

		// Convert Claude messages to normalized format
		messages := s.parseClaudeMessages(conv)
		messagesJSON, _ := json.Marshal(messages)

		// Build search content
		searchContent := s.buildSearchContent(conv.Name, messages)

		// Calculate data hash for deduplication
		dataHash := s.hashData(messagesJSON)

		// Parse timestamps
		createdAt := s.parseISO8601(conv.CreatedAt)
		updatedAt := s.parseISO8601(conv.UpdatedAt)

		// Build metadata
		metadata := map[string]any{}
		if conv.Project != nil {
			metadata["project_uuid"] = conv.Project.UUID
			metadata["project_name"] = conv.Project.Name
		}
		metadataJSON, _ := json.Marshal(metadata)

		// Create imported conversation
		imported, err := s.queries.CreateImportedConversation(ctx, sqlc.CreateImportedConversationParams{
			UserID:                 userID,
			ImportJobID:            job.ID,
			SourceType:             sqlc.ImportSourceTypeClaudeExport,
			ExternalConversationID: ptrIfNotEmpty(conv.UUID),
			Title:                  ptrIfNotEmpty(conv.Name),
			Model:                  ptrIfNotEmpty(conv.Model),
			Messages:               messagesJSON,
			MessageCount:           ptr(int32(len(messages))),
			OriginalCreatedAt:      createdAt,
			OriginalUpdatedAt:      updatedAt,
			Metadata:               metadataJSON,
			SearchContent:          ptrIfNotEmpty(searchContent),
			Tags:                   []string{},
		})
		if err != nil {
			errors = append(errors, ImportError{
				RecordIndex: i,
				ExternalID:  conv.UUID,
				Error:       err.Error(),
			})
			result.FailedRecords++
			continue
		}

		// Track the imported record for deduplication
		s.queries.CreateImportedRecord(ctx, sqlc.CreateImportedRecordParams{
			UserID:           userID,
			ImportJobID:      job.ID,
			SourceType:       sqlc.ImportSourceTypeClaudeExport,
			ExternalID:       conv.UUID,
			TargetModule:     "conversations",
			TargetRecordID:   imported.ID,
			ExternalDataHash: ptrIfNotEmpty(dataHash),
		})

		result.ImportedRecords++

		// Update progress every 10 records
		if (i+1)%10 == 0 || i == totalRecords-1 {
			progress := int((float64(i+1) / float64(totalRecords)) * 100)
			s.UpdateImportProgress(ctx, userID, jobID, ImportProgress{
				TotalRecords:     totalRecords,
				ProcessedRecords: i + 1,
				ImportedRecords:  result.ImportedRecords,
				SkippedRecords:   result.SkippedRecords,
				FailedRecords:    result.FailedRecords,
				ProgressPercent:  progress,
			})
		}
	}

	// Complete the job
	result.Errors = errors
	if result.FailedRecords > 0 && result.ImportedRecords == 0 {
		result.Status = sqlc.ImportStatusFailed
		s.FailImportJob(ctx, userID, jobID, "All records failed to import", map[string]any{
			"errors": errors,
		})
	} else {
		result.Status = sqlc.ImportStatusCompleted
		s.CompleteImportJob(ctx, userID, jobID, map[string]any{
			"imported":    result.ImportedRecords,
			"skipped":     result.SkippedRecords,
			"failed":      result.FailedRecords,
			"error_count": len(errors),
		})
	}

	return result, nil
}

// parseClaudeMessages converts Claude messages to normalized format
func (s *ImportService) parseClaudeMessages(conv ClaudeConversation) []NormalizedMessage {
	var messages []NormalizedMessage

	for _, msg := range conv.ChatMessages {
		// Map Claude sender to standard role
		role := "user"
		if msg.Sender == "assistant" {
			role = "assistant"
		}

		normalized := NormalizedMessage{
			Role:    role,
			Content: msg.Text,
		}

		// Parse timestamp if available
		if msg.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, msg.CreatedAt); err == nil {
				normalized.Timestamp = &t
			}
		}

		// Include file information in metadata
		if len(msg.Files) > 0 {
			files := make([]map[string]any, len(msg.Files))
			for i, f := range msg.Files {
				files[i] = map[string]any{
					"file_name": f.FileName,
					"file_type": f.FileType,
					"file_size": f.FileSize,
				}
			}
			normalized.Metadata = map[string]any{"files": files}
		}

		messages = append(messages, normalized)
	}

	return messages
}

// =============================================================================
// IMPORTED CONVERSATIONS QUERY METHODS
// =============================================================================

// GetImportedConversations retrieves imported conversations for a user
func (s *ImportService) GetImportedConversations(ctx context.Context, userID string, limit, offset int32) ([]sqlc.ImportedConversation, error) {
	return s.queries.GetImportedConversationsByUser(ctx, sqlc.GetImportedConversationsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

// GetImportedConversationsBySource retrieves imported conversations filtered by source
func (s *ImportService) GetImportedConversationsBySource(ctx context.Context, userID string, sourceType sqlc.ImportSourceType, limit, offset int32) ([]sqlc.ImportedConversation, error) {
	return s.queries.GetImportedConversationsBySource(ctx, sqlc.GetImportedConversationsBySourceParams{
		UserID:     userID,
		SourceType: sourceType,
		Limit:      limit,
		Offset:     offset,
	})
}

// GetImportedConversation retrieves a single imported conversation
func (s *ImportService) GetImportedConversation(ctx context.Context, userID string, conversationID uuid.UUID) (*sqlc.ImportedConversation, error) {
	conv, err := s.queries.GetImportedConversation(ctx, sqlc.GetImportedConversationParams{
		ID:     pgtype.UUID{Bytes: conversationID, Valid: true},
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// SearchImportedConversations searches imported conversations by content
func (s *ImportService) SearchImportedConversations(ctx context.Context, userID string, query string, limit int32) ([]sqlc.ImportedConversation, error) {
	return s.queries.SearchImportedConversations(ctx, sqlc.SearchImportedConversationsParams{
		UserID:         userID,
		PlaintoTsquery: query,
		Limit:          limit,
	})
}

// CountImportedConversations counts imported conversations for a user
func (s *ImportService) CountImportedConversations(ctx context.Context, userID string) (int64, error) {
	return s.queries.CountImportedConversations(ctx, userID)
}

// LinkConversationToContext links an imported conversation to a BusinessOS context
func (s *ImportService) LinkConversationToContext(ctx context.Context, userID string, conversationID, contextID uuid.UUID) error {
	return s.queries.LinkConversationToContext(ctx, sqlc.LinkConversationToContextParams{
		ID:              pgtype.UUID{Bytes: conversationID, Valid: true},
		UserID:          userID,
		LinkedContextID: pgtype.UUID{Bytes: contextID, Valid: true},
	})
}

// LinkConversationToProject links an imported conversation to a BusinessOS project
func (s *ImportService) LinkConversationToProject(ctx context.Context, userID string, conversationID, projectID uuid.UUID) error {
	return s.queries.LinkConversationToProject(ctx, sqlc.LinkConversationToProjectParams{
		ID:              pgtype.UUID{Bytes: conversationID, Valid: true},
		UserID:          userID,
		LinkedProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
	})
}

// UpdateConversationTags updates the tags on an imported conversation
func (s *ImportService) UpdateConversationTags(ctx context.Context, userID string, conversationID uuid.UUID, tags []string) error {
	return s.queries.UpdateConversationTags(ctx, sqlc.UpdateConversationTagsParams{
		ID:     pgtype.UUID{Bytes: conversationID, Valid: true},
		UserID: userID,
		Tags:   tags,
	})
}

// DeleteImportedConversation deletes an imported conversation
func (s *ImportService) DeleteImportedConversation(ctx context.Context, userID string, conversationID uuid.UUID) error {
	return s.queries.DeleteImportedConversation(ctx, sqlc.DeleteImportedConversationParams{
		ID:     pgtype.UUID{Bytes: conversationID, Valid: true},
		UserID: userID,
	})
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// buildSearchContent creates searchable text from conversation title and messages
func (s *ImportService) buildSearchContent(title string, messages []NormalizedMessage) string {
	var parts []string
	if title != "" {
		parts = append(parts, title)
	}
	for _, msg := range messages {
		if msg.Content != "" {
			parts = append(parts, msg.Content)
		}
	}
	return strings.Join(parts, " ")
}

// hashData creates a SHA256 hash of data for change detection
func (s *ImportService) hashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// unixToTimestamp converts Unix timestamp to pgtype.Timestamptz
func (s *ImportService) unixToTimestamp(unix float64) pgtype.Timestamptz {
	if unix <= 0 {
		return pgtype.Timestamptz{}
	}
	sec := int64(unix)
	nsec := int64((unix - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec)
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// parseISO8601 parses an ISO8601 timestamp to pgtype.Timestamptz
func (s *ImportService) parseISO8601(ts string) pgtype.Timestamptz {
	if ts == "" {
		return pgtype.Timestamptz{}
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}
