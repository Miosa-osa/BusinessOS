package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// ContextService handles context profiles, loading rules, and tree operations
type ContextService struct {
	pool             *pgxpool.Pool
	embeddingService *EmbeddingService
	logger           *slog.Logger
}

// NewContextService creates a new context service
func NewContextService(pool *pgxpool.Pool, embeddingService *EmbeddingService) *ContextService {
	return &ContextService{
		pool:             pool,
		embeddingService: embeddingService,
		logger:           slog.Default().With("service", "context"),
	}
}

// ============================================================================
// Types
// ============================================================================

// ContextProfile represents a context profile entity
type ContextProfile struct {
	ID                 uuid.UUID         `json:"id"`
	UserID             string            `json:"user_id"`
	EntityType         string            `json:"entity_type"`
	EntityID           uuid.UUID         `json:"entity_id"`
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	ContextTree        map[string]any    `json:"context_tree"`
	Summary            string            `json:"summary,omitempty"`
	KeyFacts           []string          `json:"key_facts,omitempty"`
	DocumentTypes      []string          `json:"document_types,omitempty"`
	TotalDocuments     int               `json:"total_documents"`
	TotalFileSizeBytes int64             `json:"total_file_size_bytes"`
	TotalContexts      int               `json:"total_contexts"`
	TotalMemories      int               `json:"total_memories"`
	TotalArtifacts     int               `json:"total_artifacts"`
	TotalTasks         int               `json:"total_tasks"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// ContextLoadingRule defines what context to load automatically
type ContextLoadingRule struct {
	ID                       uuid.UUID `json:"id"`
	UserID                   string    `json:"user_id"`
	Name                     string    `json:"name"`
	Description              string    `json:"description,omitempty"`
	TriggerType              string    `json:"trigger_type"`
	TriggerValue             string    `json:"trigger_value,omitempty"`
	LoadMemories             bool      `json:"load_memories"`
	MemoryTypes              []string  `json:"memory_types,omitempty"`
	MemoryLimit              int       `json:"memory_limit"`
	LoadContexts             bool      `json:"load_contexts"`
	ContextCategories        []string  `json:"context_categories,omitempty"`
	ContextLimit             int       `json:"context_limit"`
	LoadArtifacts            bool      `json:"load_artifacts"`
	ArtifactTypes            []string  `json:"artifact_types,omitempty"`
	ArtifactLimit            int       `json:"artifact_limit"`
	LoadRecentConversations  bool      `json:"load_recent_conversations"`
	ConversationLimit        int       `json:"conversation_limit"`
	Priority                 int       `json:"priority"`
	IsActive                 bool      `json:"is_active"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// AgentContextSession tracks context for an agent session
type AgentContextSession struct {
	ID                      uuid.UUID   `json:"id"`
	UserID                  string      `json:"user_id"`
	ConversationID          uuid.UUID   `json:"conversation_id"`
	AgentType               string      `json:"agent_type"`
	AgentID                 *uuid.UUID  `json:"agent_id,omitempty"`
	MaxContextTokens        int         `json:"max_context_tokens"`
	UsedContextTokens       int         `json:"used_context_tokens"`
	AvailableTokens         int         `json:"available_tokens"`
	LoadedMemories          []uuid.UUID `json:"loaded_memories"`
	LoadedContexts          []uuid.UUID `json:"loaded_contexts"`
	LoadedArtifacts         []uuid.UUID `json:"loaded_artifacts"`
	LoadedDocuments         []uuid.UUID `json:"loaded_documents"`
	BaseSystemPrompt        string      `json:"base_system_prompt,omitempty"`
	InjectedContext         string      `json:"injected_context,omitempty"`
	TotalSystemPromptTokens int         `json:"total_system_prompt_tokens"`
	ProjectID               *uuid.UUID  `json:"project_id,omitempty"`
	NodeID                  *uuid.UUID  `json:"node_id,omitempty"`
	FocusMode               string      `json:"focus_mode,omitempty"`
	StartedAt               time.Time   `json:"started_at"`
	LastActivityAt          time.Time   `json:"last_activity_at"`
	EndedAt                 *time.Time  `json:"ended_at,omitempty"`
}

// TreeSearchParams contains parameters for tree search
type TreeSearchParams struct {
	Query        string   `json:"query"`
	SearchType   string   `json:"search_type"`    // 'title', 'content', 'semantic', 'browse'
	EntityTypes  []string `json:"entity_types"`   // 'memories', 'contexts', 'artifacts', 'documents'
	ProjectScope *string  `json:"project_scope"`
	NodeScope    *string  `json:"node_scope"`
	MaxResults   int      `json:"max_results"`
}

// TreeSearchResult represents a search result from the context tree
type TreeSearchResult struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	Summary        string    `json:"summary,omitempty"`
	RelevanceScore float64   `json:"relevance_score"`
	TreePath       []string  `json:"tree_path"`
	TokenEstimate  int       `json:"token_estimate"`
}

// ContextItem represents a loaded context item
type ContextItem struct {
	ID         uuid.UUID      `json:"id"`
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	TokenCount int            `json:"token_count"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ContextTree represents the hierarchical context structure
type ContextTree struct {
	RootNode    *ContextTreeNode `json:"root_node"`
	TotalItems  int              `json:"total_items"`
	LastUpdated time.Time        `json:"last_updated"`
}

// ContextTreeNode represents a node in the context tree
type ContextTreeNode struct {
	ID          uuid.UUID          `json:"id"`
	Type        string             `json:"type"` // 'root', 'project', 'node', 'category', 'item'
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Icon        string             `json:"icon,omitempty"`
	ItemCount   int                `json:"item_count"`
	TokenCount  int                `json:"token_count"`
	Children    []*ContextTreeNode `json:"children,omitempty"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// TreeStatistics contains statistics about the context tree
type TreeStatistics struct {
	TotalProjects   int            `json:"total_projects"`
	TotalNodes      int            `json:"total_nodes"`
	TotalMemories   int            `json:"total_memories"`
	TotalContexts   int            `json:"total_contexts"`
	TotalArtifacts  int            `json:"total_artifacts"`
	TotalDocuments  int            `json:"total_documents"`
	TotalVoiceNotes int            `json:"total_voice_notes"`
	ByType          map[string]int `json:"by_type"`
	TotalTokens     int            `json:"total_tokens"`
}

// AgentContext contains the built context for an agent
type AgentContext struct {
	SystemPromptAddition string                 `json:"system_prompt_addition"`
	LoadedMemories       []ContextItem          `json:"loaded_memories"`
	LoadedContexts       []ContextItem          `json:"loaded_contexts"`
	LoadedArtifacts      []ContextItem          `json:"loaded_artifacts"`
	LoadedDocuments      []ContextItem          `json:"loaded_documents"`
	RecentConversations  []ConversationSummary  `json:"recent_conversations"`
	UserFacts            []UserFact             `json:"user_facts"`
	TotalTokens          int                    `json:"total_tokens"`
	TokenBreakdown       map[string]int         `json:"token_breakdown"`
}

// ConversationSummary represents a summarized conversation
type ConversationSummary struct {
	ID            uuid.UUID `json:"id"`
	Summary       string    `json:"summary"`
	KeyPoints     []string  `json:"key_points,omitempty"`
	DecisionsMade []string  `json:"decisions_made,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
	MessageCount  int       `json:"message_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserFact represents a user fact
type UserFact struct {
	ID              uuid.UUID `json:"id"`
	UserID          string    `json:"user_id"`
	FactKey         string    `json:"fact_key"`
	FactValue       string    `json:"fact_value"`
	FactType        string    `json:"fact_type"`
	ConfidenceScore float64   `json:"confidence_score"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// ContextBuildInput contains parameters for building agent context
type ContextBuildInput struct {
	UserID         string
	ProjectID      *uuid.UUID
	NodeID         *uuid.UUID
	ConversationID *uuid.UUID
	AgentType      string
	FocusMode      string
	CurrentQuery   string
	MaxTokens      int
}

// ============================================================================
// Context Profile Methods
// ============================================================================

// CreateContextProfile creates a new context profile
func (s *ContextService) CreateContextProfile(ctx context.Context, userID, entityType string, entityID uuid.UUID, name, description string) (*ContextProfile, error) {
	profile := &ContextProfile{
		ID:          uuid.New(),
		UserID:      userID,
		EntityType:  entityType,
		EntityID:    entityID,
		Name:        name,
		Description: description,
		ContextTree: make(map[string]any),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	contextTree, _ := json.Marshal(profile.ContextTree)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO context_profiles (id, user_id, entity_type, entity_id, name, description, context_tree, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, profile.ID, userID, entityType, entityID, name, description, contextTree, profile.CreatedAt, profile.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create context profile: %w", err)
	}

	return profile, nil
}

// GetContextProfile retrieves a context profile by entity
func (s *ContextService) GetContextProfile(ctx context.Context, userID, entityType string, entityID uuid.UUID) (*ContextProfile, error) {
	var profile ContextProfile
	var contextTreeJSON []byte
	var keyFacts, documentTypes []string

	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, entity_type, entity_id, name, description, context_tree, summary,
		       key_facts, document_types, total_documents, total_file_size_bytes,
		       total_contexts, total_memories, total_artifacts, total_tasks, created_at, updated_at
		FROM context_profiles
		WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3
	`, userID, entityType, entityID).Scan(
		&profile.ID, &profile.UserID, &profile.EntityType, &profile.EntityID,
		&profile.Name, &profile.Description, &contextTreeJSON, &profile.Summary,
		&keyFacts, &documentTypes, &profile.TotalDocuments, &profile.TotalFileSizeBytes,
		&profile.TotalContexts, &profile.TotalMemories, &profile.TotalArtifacts, &profile.TotalTasks,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get context profile: %w", err)
	}

	if contextTreeJSON != nil {
		json.Unmarshal(contextTreeJSON, &profile.ContextTree)
	}
	profile.KeyFacts = keyFacts
	profile.DocumentTypes = documentTypes

	return &profile, nil
}

// UpdateContextProfile updates a context profile
func (s *ContextService) UpdateContextProfile(ctx context.Context, profile *ContextProfile) error {
	contextTree, _ := json.Marshal(profile.ContextTree)

	_, err := s.pool.Exec(ctx, `
		UPDATE context_profiles SET
			name = $2, description = $3, context_tree = $4, summary = $5,
			key_facts = $6, document_types = $7, total_documents = $8, total_file_size_bytes = $9,
			total_contexts = $10, total_memories = $11, total_artifacts = $12, total_tasks = $13,
			updated_at = NOW()
		WHERE id = $1
	`, profile.ID, profile.Name, profile.Description, contextTree, profile.Summary,
		profile.KeyFacts, profile.DocumentTypes, profile.TotalDocuments, profile.TotalFileSizeBytes,
		profile.TotalContexts, profile.TotalMemories, profile.TotalArtifacts, profile.TotalTasks)

	if err != nil {
		return fmt.Errorf("update context profile: %w", err)
	}

	return nil
}

// ============================================================================
// Context Tree Methods
// ============================================================================

// GetContextTree retrieves the context tree for a user
func (s *ContextService) GetContextTree(ctx context.Context, userID string, projectID, nodeID *uuid.UUID) (*ContextTree, error) {
	tree := &ContextTree{
		RootNode: &ContextTreeNode{
			ID:       uuid.Nil,
			Type:     "root",
			Name:     "Context Tree",
			Children: make([]*ContextTreeNode, 0),
		},
		LastUpdated: time.Now(),
	}

	// Get projects
	projectNodes, err := s.getProjectNodes(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	tree.RootNode.Children = append(tree.RootNode.Children, projectNodes...)

	// Count total items
	for _, child := range tree.RootNode.Children {
		tree.TotalItems += countTreeItems(child)
	}
	tree.RootNode.ItemCount = tree.TotalItems

	return tree, nil
}

// getProjectNodes retrieves project nodes for the tree
func (s *ContextService) getProjectNodes(ctx context.Context, userID string, projectID *uuid.UUID) ([]*ContextTreeNode, error) {
	query := `
		SELECT id, name, description, status
		FROM projects
		WHERE user_id = $1 AND is_archived = false
	`
	args := []any{userID}

	if projectID != nil {
		query += " AND id = $2"
		args = append(args, *projectID)
	}

	query += " ORDER BY updated_at DESC LIMIT 50"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var nodes []*ContextTreeNode
	for rows.Next() {
		var id uuid.UUID
		var name, description, status string

		if err := rows.Scan(&id, &name, &description, &status); err != nil {
			continue
		}

		node := &ContextTreeNode{
			ID:          id,
			Type:        "project",
			Name:        name,
			Description: description,
			Icon:        "folder",
			Children:    make([]*ContextTreeNode, 0),
			Metadata: map[string]any{
				"status": status,
			},
		}

		// Get project children (memories, documents, etc.)
		children, itemCount := s.getProjectChildren(ctx, userID, id)
		node.Children = children
		node.ItemCount = itemCount

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// getProjectChildren retrieves child items for a project
func (s *ContextService) getProjectChildren(ctx context.Context, userID string, projectID uuid.UUID) ([]*ContextTreeNode, int) {
	children := make([]*ContextTreeNode, 0)
	totalItems := 0

	// Memories category
	var memoryCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM memories WHERE user_id = $1 AND project_id = $2 AND is_active = true
	`, userID, projectID).Scan(&memoryCount)

	if memoryCount > 0 {
		children = append(children, &ContextTreeNode{
			ID:        uuid.New(),
			Type:      "category",
			Name:      "Memories",
			Icon:      "brain",
			ItemCount: memoryCount,
		})
		totalItems += memoryCount
	}

	// Documents category
	var docCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM uploaded_documents WHERE user_id = $1 AND project_id = $2
	`, userID, projectID).Scan(&docCount)

	if docCount > 0 {
		children = append(children, &ContextTreeNode{
			ID:        uuid.New(),
			Type:      "category",
			Name:      "Documents",
			Icon:      "file-text",
			ItemCount: docCount,
		})
		totalItems += docCount
	}

	// Artifacts category
	var artifactCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM artifacts WHERE user_id = $1 AND project_id = $2
	`, userID, projectID).Scan(&artifactCount)

	if artifactCount > 0 {
		children = append(children, &ContextTreeNode{
			ID:        uuid.New(),
			Type:      "category",
			Name:      "Artifacts",
			Icon:      "file-code",
			ItemCount: artifactCount,
		})
		totalItems += artifactCount
	}

	// Conversations category
	var convCount int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM conversations WHERE user_id = $1 AND project_id = $2
	`, userID, projectID).Scan(&convCount)

	if convCount > 0 {
		children = append(children, &ContextTreeNode{
			ID:        uuid.New(),
			Type:      "category",
			Name:      "Conversations",
			Icon:      "message-square",
			ItemCount: convCount,
		})
		totalItems += convCount
	}

	return children, totalItems
}

// GetTreeStatistics returns statistics about the context tree
func (s *ContextService) GetTreeStatistics(ctx context.Context, userID string) (*TreeStatistics, error) {
	stats := &TreeStatistics{
		ByType: make(map[string]int),
	}

	// Count projects
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE user_id = $1 AND is_archived = false`, userID).Scan(&stats.TotalProjects)

	// Count nodes
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM nodes WHERE user_id = $1`, userID).Scan(&stats.TotalNodes)

	// Count memories
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM memories WHERE user_id = $1 AND is_active = true`, userID).Scan(&stats.TotalMemories)

	// Count contexts
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM contexts WHERE user_id = $1 AND is_archived = false`, userID).Scan(&stats.TotalContexts)

	// Count artifacts
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM artifacts WHERE user_id = $1`, userID).Scan(&stats.TotalArtifacts)

	// Count documents
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM uploaded_documents WHERE user_id = $1`, userID).Scan(&stats.TotalDocuments)

	// Count voice notes
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM voice_notes WHERE user_id = $1`, userID).Scan(&stats.TotalVoiceNotes)

	// Rough token estimate across common tree sources.
	// NOTE: This is an approximation (chars/4) to avoid loading all rows.
	var totalTokens int
	{
		var t int
		// Contexts
		s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(LENGTH(COALESCE(content, '')) / 4), 0)
			FROM contexts
			WHERE user_id = $1 AND is_archived = false
		`, userID).Scan(&t)
		totalTokens += t
	}
	{
		var t int
		// Artifacts
		s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(LENGTH(COALESCE(content, '')) / 4), 0)
			FROM artifacts
			WHERE user_id = $1
		`, userID).Scan(&t)
		totalTokens += t
	}
	{
		var t int
		// Documents (extracted_text)
		s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(LENGTH(COALESCE(extracted_text, '')) / 4), 0)
			FROM uploaded_documents
			WHERE user_id = $1
		`, userID).Scan(&t)
		totalTokens += t
	}
	{
		var t int
		// Memories
		s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(LENGTH(COALESCE(content, '')) / 4), 0)
			FROM memories
			WHERE user_id = $1 AND is_active = true
		`, userID).Scan(&t)
		totalTokens += t
	}
	{
		var t int
		// Voice notes
		s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(LENGTH(COALESCE(transcript, '')) / 4), 0)
			FROM voice_notes
			WHERE user_id = $1
		`, userID).Scan(&t)
		totalTokens += t
	}
	stats.TotalTokens = totalTokens

	stats.ByType["projects"] = stats.TotalProjects
	stats.ByType["nodes"] = stats.TotalNodes
	stats.ByType["memories"] = stats.TotalMemories
	stats.ByType["contexts"] = stats.TotalContexts
	stats.ByType["artifacts"] = stats.TotalArtifacts
	stats.ByType["documents"] = stats.TotalDocuments
	stats.ByType["voice_notes"] = stats.TotalVoiceNotes

	return stats, nil
}

// ============================================================================
// Tree Search Methods
// ============================================================================

// SearchTree searches the context tree based on parameters
func (s *ContextService) SearchTree(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error) {
	if params.MaxResults <= 0 {
		params.MaxResults = 10
	}

	switch params.SearchType {
	case "semantic":
		return s.semanticSearch(ctx, userID, params)
	case "title":
		return s.titleSearch(ctx, userID, params)
	case "content":
		return s.contentSearch(ctx, userID, params)
	default:
		return s.titleSearch(ctx, userID, params)
	}
}

// semanticSearch performs semantic search using embeddings
func (s *ContextService) semanticSearch(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error) {
	if s.embeddingService == nil {
		return s.titleSearch(ctx, userID, params)
	}

	// Generate query embedding
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, params.Query)
	if err != nil {
		s.logger.Warn("failed to generate embedding, falling back to title search", "error", err)
		return s.titleSearch(ctx, userID, params)
	}

	vec := pgvector.NewVector(queryEmbedding)
	var results []TreeSearchResult

	// Search memories
	if containsType(params.EntityTypes, "memories") || len(params.EntityTypes) == 0 {
		memResults, _ := s.searchMemoriesSemantic(ctx, userID, vec, params)
		results = append(results, memResults...)
	}

	// Search documents
	if containsType(params.EntityTypes, "documents") || len(params.EntityTypes) == 0 {
		docResults, _ := s.searchDocumentsSemantic(ctx, userID, vec, params)
		results = append(results, docResults...)
	}

	// Search voice notes
	if containsType(params.EntityTypes, "voice_notes") || len(params.EntityTypes) == 0 {
		voiceResults, _ := s.searchVoiceNotesSemantic(ctx, userID, vec, params)
		results = append(results, voiceResults...)
	}

	// Search conversation summaries (past chat history)
	if containsType(params.EntityTypes, "conversations") || containsType(params.EntityTypes, "conversation_summaries") || len(params.EntityTypes) == 0 {
		convResults, _ := s.searchConversationSummariesSemantic(ctx, userID, vec, params)
		results = append(results, convResults...)
	}

	// Sort by relevance and limit
	sortByRelevance(results)
	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

// searchConversationSummariesSemantic searches conversation summaries using embedding.
func (s *ContextService) searchConversationSummariesSemantic(ctx context.Context, userID string, vec pgvector.Vector, params TreeSearchParams) ([]TreeSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cs.conversation_id,
		       COALESCE(cs.title, ''),
		       LEFT(cs.summary, 200) as snippet,
		       COALESCE(cs.summarized_at, cs.created_at) as ts,
		       1 - (cs.embedding <=> $1) as similarity
		FROM conversation_summaries cs
		WHERE cs.user_id = $2 AND cs.embedding IS NOT NULL
		ORDER BY cs.embedding <=> $1
		LIMIT $3
	`, vec, userID, params.MaxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TreeSearchResult
	for rows.Next() {
		var r TreeSearchResult
		var title string
		var snippet string
		var ts time.Time
		if err := rows.Scan(&r.ID, &title, &snippet, &ts, &r.RelevanceScore); err != nil {
			continue
		}
		r.Type = "conversation"
		if strings.TrimSpace(title) != "" {
			r.Title = title
		} else {
			r.Title = "Conversation (" + ts.UTC().Format("2006-01-02") + ")"
		}
		r.Summary = snippet
		r.TreePath = []string{"Conversations"}
		results = append(results, r)
	}

	return results, nil
}

// searchVoiceNotesSemantic searches voice notes using embedding
func (s *ContextService) searchVoiceNotesSemantic(ctx context.Context, userID string, vec pgvector.Vector, params TreeSearchParams) ([]TreeSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, LEFT(transcript, 200) as snippet, created_at, 1 - (embedding <=> $1) as similarity
		FROM voice_notes
		WHERE user_id = $2 AND is_context_source = true AND embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $3
	`, vec, userID, params.MaxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TreeSearchResult
	for rows.Next() {
		var r TreeSearchResult
		var snippet string
		var createdAt time.Time
		if err := rows.Scan(&r.ID, &snippet, &createdAt, &r.RelevanceScore); err != nil {
			continue
		}
		r.Type = "voice_note"
		r.Title = "Voice note (" + createdAt.UTC().Format("2006-01-02") + ")"
		r.Summary = snippet
		r.TreePath = []string{"Voice Notes"}
		results = append(results, r)
	}

	return results, nil
}

// searchMemoriesSemantic searches memories using embedding
func (s *ContextService) searchMemoriesSemantic(ctx context.Context, userID string, vec pgvector.Vector, params TreeSearchParams) ([]TreeSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, memory_type, summary, 1 - (embedding <=> $1) as similarity
		FROM memories
		WHERE user_id = $2 AND is_active = true AND embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $3
	`, vec, userID, params.MaxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TreeSearchResult
	for rows.Next() {
		var r TreeSearchResult
		var memType string
		if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary, &r.RelevanceScore); err != nil {
			continue
		}
		r.Type = "memory"
		r.TreePath = []string{"Memories", memType}
		results = append(results, r)
	}

	return results, nil
}

// searchDocumentsSemantic searches documents using embedding
func (s *ContextService) searchDocumentsSemantic(ctx context.Context, userID string, vec pgvector.Vector, params TreeSearchParams) ([]TreeSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, display_name, document_type, description, 1 - (embedding <=> $1) as similarity
		FROM uploaded_documents
		WHERE user_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $3
	`, vec, userID, params.MaxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TreeSearchResult
	for rows.Next() {
		var r TreeSearchResult
		var docType string
		if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary, &r.RelevanceScore); err != nil {
			continue
		}
		r.Type = "document"
		r.TreePath = []string{"Documents", docType}
		results = append(results, r)
	}

	return results, nil
}

// titleSearch searches by title
func (s *ContextService) titleSearch(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error) {
	var results []TreeSearchResult
	searchPattern := "%" + params.Query + "%"

	// Search memories by title
	if containsType(params.EntityTypes, "memories") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, title, memory_type, summary
			FROM memories
			WHERE user_id = $1 AND is_active = true AND (title ILIKE $2 OR summary ILIKE $2)
			ORDER BY importance_score DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var memType string
				if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary); err != nil {
					continue
				}
				r.Type = "memory"
				r.TreePath = []string{"Memories", memType}
				r.RelevanceScore = 0.8
				results = append(results, r)
			}
			rows.Close()
		}
	}

	// Search documents by title
	if containsType(params.EntityTypes, "documents") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, COALESCE(display_name, filename), document_type, description
			FROM uploaded_documents
			WHERE user_id = $1 AND (display_name ILIKE $2 OR filename ILIKE $2 OR description ILIKE $2)
			ORDER BY created_at DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var docType *string
				if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary); err != nil {
					continue
				}
				r.Type = "document"
				if docType != nil {
					r.TreePath = []string{"Documents", *docType}
				} else {
					r.TreePath = []string{"Documents"}
				}
				r.RelevanceScore = 0.7
				results = append(results, r)
			}
			rows.Close()
		}
	}

	// Search voice notes by transcript (best-effort "title")
	if containsType(params.EntityTypes, "voice_notes") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, created_at, LEFT(transcript, 200)
			FROM voice_notes
			WHERE user_id = $1 AND is_context_source = true AND transcript ILIKE $2
			ORDER BY created_at DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var createdAt time.Time
				if err := rows.Scan(&r.ID, &createdAt, &r.Summary); err != nil {
					continue
				}
				r.Type = "voice_note"
				r.Title = "Voice note (" + createdAt.UTC().Format("2006-01-02") + ")"
				r.TreePath = []string{"Voice Notes"}
				r.RelevanceScore = 0.65
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

// contentSearch searches by content
func (s *ContextService) contentSearch(ctx context.Context, userID string, params TreeSearchParams) ([]TreeSearchResult, error) {
	var results []TreeSearchResult
	searchPattern := "%" + params.Query + "%"

	// Search memories by content
	if containsType(params.EntityTypes, "memories") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, title, memory_type, LEFT(content, 200)
			FROM memories
			WHERE user_id = $1 AND is_active = true AND content ILIKE $2
			ORDER BY importance_score DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var memType string
				if err := rows.Scan(&r.ID, &r.Title, &memType, &r.Summary); err != nil {
					continue
				}
				r.Type = "memory"
				r.TreePath = []string{"Memories", memType}
				r.RelevanceScore = 0.7
				results = append(results, r)
			}
			rows.Close()
		}
	}

	// Search documents by extracted text
	if containsType(params.EntityTypes, "documents") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, COALESCE(display_name, filename), document_type, LEFT(extracted_text, 200)
			FROM uploaded_documents
			WHERE user_id = $1 AND extracted_text ILIKE $2
			ORDER BY created_at DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var docType *string
				if err := rows.Scan(&r.ID, &r.Title, &docType, &r.Summary); err != nil {
					continue
				}
				r.Type = "document"
				if docType != nil {
					r.TreePath = []string{"Documents", *docType}
				} else {
					r.TreePath = []string{"Documents"}
				}
				r.RelevanceScore = 0.6
				results = append(results, r)
			}
			rows.Close()
		}
	}

	// Search voice notes by transcript
	if containsType(params.EntityTypes, "voice_notes") || len(params.EntityTypes) == 0 {
		rows, _ := s.pool.Query(ctx, `
			SELECT id, created_at, LEFT(transcript, 200)
			FROM voice_notes
			WHERE user_id = $1 AND is_context_source = true AND transcript ILIKE $2
			ORDER BY created_at DESC
			LIMIT $3
		`, userID, searchPattern, params.MaxResults)
		if rows != nil {
			for rows.Next() {
				var r TreeSearchResult
				var createdAt time.Time
				if err := rows.Scan(&r.ID, &createdAt, &r.Summary); err != nil {
					continue
				}
				r.Type = "voice_note"
				r.Title = "Voice note (" + createdAt.UTC().Format("2006-01-02") + ")"
				r.TreePath = []string{"Voice Notes"}
				r.RelevanceScore = 0.6
				results = append(results, r)
			}
			rows.Close()
		}
	}

	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return results, nil
}

// ============================================================================
// Context Loading Methods
// ============================================================================

// LoadContextItem loads a specific context item by ID and type
func (s *ContextService) LoadContextItem(ctx context.Context, userID string, itemID uuid.UUID, itemType string) (*ContextItem, error) {
	item := &ContextItem{
		ID:   itemID,
		Type: itemType,
	}

	switch itemType {
	case "memory":
		err := s.pool.QueryRow(ctx, `
			SELECT title, content FROM memories WHERE id = $1 AND user_id = $2
		`, itemID, userID).Scan(&item.Title, &item.Content)
		if err != nil {
			return nil, fmt.Errorf("load memory: %w", err)
		}

	case "document":
		err := s.pool.QueryRow(ctx, `
			SELECT COALESCE(display_name, filename), extracted_text FROM uploaded_documents WHERE id = $1 AND user_id = $2
		`, itemID, userID).Scan(&item.Title, &item.Content)
		if err != nil {
			return nil, fmt.Errorf("load document: %w", err)
		}

	case "artifact":
		err := s.pool.QueryRow(ctx, `
			SELECT title, content FROM artifacts WHERE id = $1 AND user_id = $2
		`, itemID, userID).Scan(&item.Title, &item.Content)
		if err != nil {
			return nil, fmt.Errorf("load artifact: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown item type: %s", itemType)
	}

	// Estimate tokens (rough estimate: 4 chars per token)
	item.TokenCount = len(item.Content) / 4

	return item, nil
}

// ============================================================================
// Context Loading Rules Methods
// ============================================================================

// GetLoadingRules retrieves context loading rules for a user
func (s *ContextService) GetLoadingRules(ctx context.Context, userID, triggerType, triggerValue string) ([]ContextLoadingRule, error) {
	query := `
		SELECT id, user_id, name, description, trigger_type, trigger_value,
		       load_memories, memory_types, memory_limit,
		       load_contexts, context_categories, context_limit,
		       load_artifacts, artifact_types, artifact_limit,
		       load_recent_conversations, conversation_limit,
		       priority, is_active, created_at, updated_at
		FROM context_loading_rules
		WHERE user_id = $1 AND is_active = true
	`
	args := []any{userID}

	if triggerType != "" {
		query += " AND (trigger_type = $2 OR trigger_type = 'always')"
		args = append(args, triggerType)
	}

	query += " ORDER BY priority DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query loading rules: %w", err)
	}
	defer rows.Close()

	var rules []ContextLoadingRule
	for rows.Next() {
		var r ContextLoadingRule
		err := rows.Scan(
			&r.ID, &r.UserID, &r.Name, &r.Description, &r.TriggerType, &r.TriggerValue,
			&r.LoadMemories, &r.MemoryTypes, &r.MemoryLimit,
			&r.LoadContexts, &r.ContextCategories, &r.ContextLimit,
			&r.LoadArtifacts, &r.ArtifactTypes, &r.ArtifactLimit,
			&r.LoadRecentConversations, &r.ConversationLimit,
			&r.Priority, &r.IsActive, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			continue
		}
		rules = append(rules, r)
	}

	return rules, nil
}

// CreateLoadingRule creates a new context loading rule
func (s *ContextService) CreateLoadingRule(ctx context.Context, rule *ContextLoadingRule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO context_loading_rules (
			id, user_id, name, description, trigger_type, trigger_value,
			load_memories, memory_types, memory_limit,
			load_contexts, context_categories, context_limit,
			load_artifacts, artifact_types, artifact_limit,
			load_recent_conversations, conversation_limit,
			priority, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, rule.ID, rule.UserID, rule.Name, rule.Description, rule.TriggerType, rule.TriggerValue,
		rule.LoadMemories, rule.MemoryTypes, rule.MemoryLimit,
		rule.LoadContexts, rule.ContextCategories, rule.ContextLimit,
		rule.LoadArtifacts, rule.ArtifactTypes, rule.ArtifactLimit,
		rule.LoadRecentConversations, rule.ConversationLimit,
		rule.Priority, rule.IsActive, rule.CreatedAt, rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create loading rule: %w", err)
	}

	return nil
}

// ============================================================================
// Agent Context Session Methods
// ============================================================================

// CreateContextSession creates a new agent context session
func (s *ContextService) CreateContextSession(ctx context.Context, userID string, conversationID uuid.UUID, agentType string, maxTokens int) (*AgentContextSession, error) {
	session := &AgentContextSession{
		ID:               uuid.New(),
		UserID:           userID,
		ConversationID:   conversationID,
		AgentType:        agentType,
		MaxContextTokens: maxTokens,
		AvailableTokens:  maxTokens,
		LoadedMemories:   []uuid.UUID{},
		LoadedContexts:   []uuid.UUID{},
		LoadedArtifacts:  []uuid.UUID{},
		LoadedDocuments:  []uuid.UUID{},
		StartedAt:        time.Now(),
		LastActivityAt:   time.Now(),
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO agent_context_sessions (
			id, user_id, conversation_id, agent_type, max_context_tokens,
			used_context_tokens, available_tokens, loaded_memories, loaded_contexts,
			loaded_artifacts, loaded_documents, started_at, last_activity_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, session.ID, userID, conversationID, agentType, maxTokens,
		0, maxTokens, session.LoadedMemories, session.LoadedContexts,
		session.LoadedArtifacts, session.LoadedDocuments, session.StartedAt, session.LastActivityAt)

	if err != nil {
		return nil, fmt.Errorf("create context session: %w", err)
	}

	return session, nil
}

// GetContextSession retrieves an agent context session
func (s *ContextService) GetContextSession(ctx context.Context, sessionID uuid.UUID) (*AgentContextSession, error) {
	var session AgentContextSession

	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, conversation_id, agent_type, agent_id, max_context_tokens,
		       used_context_tokens, available_tokens, loaded_memories, loaded_contexts,
		       loaded_artifacts, loaded_documents, base_system_prompt, injected_context,
		       total_system_prompt_tokens, project_id, node_id, focus_mode,
		       started_at, last_activity_at, ended_at
		FROM agent_context_sessions WHERE id = $1
	`, sessionID).Scan(
		&session.ID, &session.UserID, &session.ConversationID, &session.AgentType,
		&session.AgentID, &session.MaxContextTokens, &session.UsedContextTokens,
		&session.AvailableTokens, &session.LoadedMemories, &session.LoadedContexts,
		&session.LoadedArtifacts, &session.LoadedDocuments, &session.BaseSystemPrompt,
		&session.InjectedContext, &session.TotalSystemPromptTokens, &session.ProjectID,
		&session.NodeID, &session.FocusMode, &session.StartedAt, &session.LastActivityAt,
		&session.EndedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get context session: %w", err)
	}

	return &session, nil
}

// UpdateSessionTokenUsage updates the token usage for a session
func (s *ContextService) UpdateSessionTokenUsage(ctx context.Context, sessionID uuid.UUID, usedTokens int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE agent_context_sessions
		SET used_context_tokens = $2, available_tokens = max_context_tokens - $2, last_activity_at = NOW()
		WHERE id = $1
	`, sessionID, usedTokens)

	if err != nil {
		return fmt.Errorf("update session token usage: %w", err)
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func countTreeItems(node *ContextTreeNode) int {
	count := node.ItemCount
	for _, child := range node.Children {
		count += countTreeItems(child)
	}
	return count
}

func containsType(types []string, t string) bool {
	for _, typ := range types {
		if typ == t {
			return true
		}
	}
	return false
}

func sortByRelevance(results []TreeSearchResult) {
	// Simple bubble sort for small arrays
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].RelevanceScore > results[i].RelevanceScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
