package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectContextService handles loading context for projects and nodes
type ProjectContextService struct {
	pool           *pgxpool.Pool
	contextService *ContextService
	logger         *slog.Logger
}

// NewProjectContextService creates a new project context service
func NewProjectContextService(pool *pgxpool.Pool, contextService *ContextService) *ProjectContextService {
	return &ProjectContextService{
		pool:           pool,
		contextService: contextService,
		logger:         slog.Default().With("service", "project_context"),
	}
}

// ============================================================================
// Types
// ============================================================================

// Project represents a project entity
type Project struct {
	ID          uuid.UUID  `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority,omitempty"`
	ClientID    *uuid.UUID `json:"client_id,omitempty"`
	ClientName  string     `json:"client_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Node represents a business node entity
type Node struct {
	ID          uuid.UUID  `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Path        string     `json:"path,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Memory represents a memory entity
type Memory struct {
	ID              uuid.UUID  `json:"id"`
	UserID          string     `json:"user_id"`
	Title           string     `json:"title"`
	Summary         string     `json:"summary"`
	Content         string     `json:"content"`
	MemoryType      string     `json:"memory_type"`
	Category        string     `json:"category,omitempty"`
	SourceType      string     `json:"source_type"`
	SourceID        *uuid.UUID `json:"source_id,omitempty"`
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	NodeID          *uuid.UUID `json:"node_id,omitempty"`
	ImportanceScore float64    `json:"importance_score"`
	AccessCount     int        `json:"access_count"`
	IsPinned        bool       `json:"is_pinned"`
	Tags            []string   `json:"tags,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// VoiceNote represents a voice note entity
type VoiceNote struct {
	ID         uuid.UUID  `json:"id"`
	UserID     string     `json:"user_id"`
	Title      string     `json:"title,omitempty"`
	Transcript string     `json:"transcript,omitempty"`
	Duration   int        `json:"duration"`
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`
	NodeID     *uuid.UUID `json:"node_id,omitempty"`
	KeyTopics  []string   `json:"key_topics,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Artifact represents an artifact entity
type Artifact struct {
	ID           uuid.UUID  `json:"id"`
	UserID       string     `json:"user_id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	ArtifactType string     `json:"artifact_type"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Document represents an uploaded document
type Document struct {
	ID           uuid.UUID `json:"id"`
	UserID       string    `json:"user_id"`
	Filename     string    `json:"filename"`
	DisplayName  string    `json:"display_name,omitempty"`
	Description  string    `json:"description,omitempty"`
	FileType     string    `json:"file_type"`
	DocumentType string    `json:"document_type,omitempty"`
	ExtractedText string   `json:"extracted_text,omitempty"`
	WordCount    int       `json:"word_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// Context represents a KB context entity
type Context struct {
	ID           uuid.UUID `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Content      string    `json:"content,omitempty"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProjectContext contains all context for a project
type ProjectContext struct {
	Project              *Project               `json:"project"`
	Profile              *ContextProfile        `json:"profile,omitempty"`
	Memories             []Memory               `json:"memories"`
	Documents            []Document             `json:"documents"`
	Artifacts            []Artifact             `json:"artifacts"`
	VoiceNotes           []VoiceNote            `json:"voice_notes"`
	Conversations        []ConversationSummary  `json:"conversations"`
	KnowledgeBaseContexts []Context             `json:"kb_contexts"`
	UserFacts            []UserFact             `json:"user_facts"`
	TotalTokenEstimate   int                    `json:"total_token_estimate"`
}

// NodeContext contains all context for a node
type NodeContext struct {
	Node          *Node           `json:"node"`
	Ancestors     []*Node         `json:"ancestors,omitempty"`
	Profile       *ContextProfile `json:"profile,omitempty"`
	Memories      []Memory        `json:"memories"`
	Projects      []Project       `json:"projects"`
	ParentContext *ProjectContext `json:"parent_context,omitempty"`
}

// InjectedContext represents context ready to be injected into a conversation
type InjectedContext struct {
	SystemPromptAddition string        `json:"system_prompt_addition"`
	LoadedItems          []ContextItem `json:"loaded_items"`
	TotalTokens          int           `json:"total_tokens"`
	TokenBreakdown       map[string]int `json:"token_breakdown"`
}

// ============================================================================
// Project Context Methods
// ============================================================================

// LoadProjectContext loads all relevant context when a project is selected
func (s *ProjectContextService) LoadProjectContext(ctx context.Context, userID string, projectID uuid.UUID) (*ProjectContext, error) {
	pc := &ProjectContext{
		Memories:              make([]Memory, 0),
		Documents:             make([]Document, 0),
		Artifacts:             make([]Artifact, 0),
		VoiceNotes:            make([]VoiceNote, 0),
		Conversations:         make([]ConversationSummary, 0),
		KnowledgeBaseContexts: make([]Context, 0),
		UserFacts:             make([]UserFact, 0),
	}

	// 1. Get project details
	project, err := s.getProject(ctx, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	pc.Project = project

	// 2. Get project's context profile (if exists)
	if s.contextService != nil {
		profile, _ := s.contextService.GetContextProfile(ctx, userID, "project", projectID)
		pc.Profile = profile
	}

	// 3. Load memories associated with project
	memories, err := s.getProjectMemories(ctx, userID, projectID, 10)
	if err == nil {
		pc.Memories = memories
	}

	// 4. Load documents linked to project
	documents, err := s.getProjectDocuments(ctx, userID, projectID, 5)
	if err == nil {
		pc.Documents = documents
	}

	// 5. Load artifacts for project
	artifacts, err := s.getProjectArtifacts(ctx, userID, projectID, 5)
	if err == nil {
		pc.Artifacts = artifacts
	}

	// 6. Get recent voice notes for project
	voiceNotes, err := s.getProjectVoiceNotes(ctx, userID, projectID, 3)
	if err == nil {
		pc.VoiceNotes = voiceNotes
	}

	// 7. Get recent conversations in project
	conversations, err := s.getProjectConversations(ctx, userID, projectID, 3)
	if err == nil {
		pc.Conversations = conversations
	}

	// 8. Get KB contexts linked to project
	kbContexts, err := s.getProjectKBContexts(ctx, userID, projectID, 5)
	if err == nil {
		pc.KnowledgeBaseContexts = kbContexts
	}

	// 9. Get user facts
	userFacts, err := s.getUserFacts(ctx, userID)
	if err == nil {
		pc.UserFacts = userFacts
	}

	// Estimate total tokens
	pc.TotalTokenEstimate = s.estimateProjectTokens(pc)

	return pc, nil
}

// getProject retrieves project details
func (s *ProjectContextService) getProject(ctx context.Context, userID string, projectID uuid.UUID) (*Project, error) {
	var p Project
	var clientID *uuid.UUID
	var clientName *string

	err := s.pool.QueryRow(ctx, `
		SELECT p.id, p.user_id, p.name, p.description, p.status, p.priority, p.client_id,
		       c.name as client_name, p.created_at, p.updated_at
		FROM projects p
		LEFT JOIN clients c ON c.id = p.client_id
		WHERE p.id = $1 AND p.user_id = $2
	`, projectID, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.Priority, &clientID,
		&clientName, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, err
	}

	p.ClientID = clientID
	if clientName != nil {
		p.ClientName = *clientName
	}

	return &p, nil
}

// getProjectMemories retrieves memories for a project
func (s *ProjectContextService) getProjectMemories(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]Memory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, title, summary, content, memory_type, category, source_type,
		       source_id, project_id, node_id, importance_score, access_count, is_pinned,
		       tags, created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND project_id = $2 AND is_active = true
		ORDER BY is_pinned DESC, importance_score DESC, created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.UserID, &m.Title, &m.Summary, &m.Content, &m.MemoryType, &m.Category,
			&m.SourceType, &m.SourceID, &m.ProjectID, &m.NodeID, &m.ImportanceScore,
			&m.AccessCount, &m.IsPinned, &m.Tags, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			continue
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// getProjectDocuments retrieves documents for a project
func (s *ProjectContextService) getProjectDocuments(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, filename, display_name, description, file_type, document_type,
		       LEFT(extracted_text, 1000), word_count, created_at
		FROM uploaded_documents
		WHERE user_id = $1 AND project_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var d Document
		err := rows.Scan(
			&d.ID, &d.UserID, &d.Filename, &d.DisplayName, &d.Description, &d.FileType,
			&d.DocumentType, &d.ExtractedText, &d.WordCount, &d.CreatedAt,
		)
		if err != nil {
			continue
		}
		documents = append(documents, d)
	}

	return documents, nil
}

// getProjectArtifacts retrieves artifacts for a project
func (s *ProjectContextService) getProjectArtifacts(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]Artifact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, title, LEFT(content, 1000), type, project_id, created_at
		FROM artifacts
		WHERE user_id = $1 AND project_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []Artifact
	for rows.Next() {
		var a Artifact
		err := rows.Scan(&a.ID, &a.UserID, &a.Title, &a.Content, &a.ArtifactType, &a.ProjectID, &a.CreatedAt)
		if err != nil {
			continue
		}
		artifacts = append(artifacts, a)
	}

	return artifacts, nil
}

// getProjectVoiceNotes retrieves voice notes for a project
func (s *ProjectContextService) getProjectVoiceNotes(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]VoiceNote, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, title, transcript, duration, project_id, node_id, key_topics, created_at
		FROM voice_notes
		WHERE user_id = $1 AND project_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voiceNotes []VoiceNote
	for rows.Next() {
		var v VoiceNote
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Transcript, &v.Duration, &v.ProjectID, &v.NodeID, &v.KeyTopics, &v.CreatedAt)
		if err != nil {
			continue
		}
		voiceNotes = append(voiceNotes, v)
	}

	return voiceNotes, nil
}

// getProjectConversations retrieves recent conversation summaries for a project
func (s *ProjectContextService) getProjectConversations(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]ConversationSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cs.id, cs.conversation_id, cs.summary, cs.key_points, cs.decisions_made,
		       cs.topics, cs.message_count, cs.created_at
		FROM conversation_summaries cs
		JOIN conversations c ON c.id = cs.conversation_id
		WHERE cs.user_id = $1 AND c.project_id = $2
		ORDER BY cs.created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ConversationSummary
	for rows.Next() {
		var s ConversationSummary
		var convID uuid.UUID
		err := rows.Scan(&s.ID, &convID, &s.Summary, &s.KeyPoints, &s.DecisionsMade, &s.Topics, &s.MessageCount, &s.CreatedAt)
		if err != nil {
			continue
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// getProjectKBContexts retrieves KB contexts linked to a project
func (s *ProjectContextService) getProjectKBContexts(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]Context, error) {
	// Get contexts that are linked to this project via context_profile_items
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.user_id, c.name, c.type, LEFT(c.content, 500), c.system_prompt, c.created_at
		FROM contexts c
		JOIN context_profile_items cpi ON cpi.item_id = c.id AND cpi.item_type = 'kb_context'
		JOIN context_profiles cp ON cp.id = cpi.context_profile_id
		WHERE c.user_id = $1 AND cp.entity_type = 'project' AND cp.entity_id = $2 AND c.is_archived = false
		ORDER BY cpi.sort_order, c.created_at DESC
		LIMIT $3
	`, userID, projectID, limit)
	if err != nil {
		// Fallback: get contexts directly if profile items don't exist yet
		return s.getDirectProjectContexts(ctx, userID, projectID, limit)
	}
	defer rows.Close()

	var contexts []Context
	for rows.Next() {
		var c Context
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.Content, &c.SystemPrompt, &c.CreatedAt)
		if err != nil {
			continue
		}
		contexts = append(contexts, c)
	}

	if len(contexts) == 0 {
		return s.getDirectProjectContexts(ctx, userID, projectID, limit)
	}

	return contexts, nil
}

// getDirectProjectContexts retrieves contexts that might be related to the project name
func (s *ProjectContextService) getDirectProjectContexts(ctx context.Context, userID string, projectID uuid.UUID, limit int) ([]Context, error) {
	// Get project name first
	var projectName string
	s.pool.QueryRow(ctx, `SELECT name FROM projects WHERE id = $1`, projectID).Scan(&projectName)

	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, type, LEFT(content, 500), system_prompt, created_at
		FROM contexts
		WHERE user_id = $1 AND is_archived = false
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contexts []Context
	for rows.Next() {
		var c Context
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.Content, &c.SystemPrompt, &c.CreatedAt)
		if err != nil {
			continue
		}
		contexts = append(contexts, c)
	}

	return contexts, nil
}

// getUserFacts retrieves user facts
func (s *ProjectContextService) getUserFacts(ctx context.Context, userID string) ([]UserFact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, fact_key, fact_value, fact_type, confidence_score, is_active, created_at
		FROM user_facts
		WHERE user_id = $1 AND is_active = true
		ORDER BY confidence_score DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []UserFact
	for rows.Next() {
		var f UserFact
		err := rows.Scan(&f.ID, &f.UserID, &f.FactKey, &f.FactValue, &f.FactType, &f.ConfidenceScore, &f.IsActive, &f.CreatedAt)
		if err != nil {
			continue
		}
		facts = append(facts, f)
	}

	return facts, nil
}

// estimateProjectTokens estimates total tokens for project context
func (s *ProjectContextService) estimateProjectTokens(pc *ProjectContext) int {
	total := 0

	// Project info (~100 tokens)
	if pc.Project != nil {
		total += 100
	}

	// Memories (average ~200 tokens each)
	total += len(pc.Memories) * 200

	// Documents (average ~250 tokens each for summary)
	total += len(pc.Documents) * 250

	// Artifacts (average ~250 tokens each for summary)
	total += len(pc.Artifacts) * 250

	// Voice notes (average ~150 tokens each)
	total += len(pc.VoiceNotes) * 150

	// Conversations (average ~100 tokens each)
	total += len(pc.Conversations) * 100

	// KB Contexts (average ~150 tokens each)
	total += len(pc.KnowledgeBaseContexts) * 150

	// User facts (average ~20 tokens each)
	total += len(pc.UserFacts) * 20

	return total
}

// ============================================================================
// Node Context Methods
// ============================================================================

// LoadNodeContext loads context when a specific node is selected
func (s *ProjectContextService) LoadNodeContext(ctx context.Context, userID string, nodeID uuid.UUID) (*NodeContext, error) {
	nc := &NodeContext{
		Ancestors: make([]*Node, 0),
		Memories:  make([]Memory, 0),
		Projects:  make([]Project, 0),
	}

	// 1. Get node details
	node, err := s.getNode(ctx, userID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	nc.Node = node

	// 2. Get node ancestors (path to root)
	ancestors, err := s.getNodeAncestors(ctx, userID, nodeID)
	if err == nil {
		nc.Ancestors = ancestors
	}

	// 3. Get node's context profile
	if s.contextService != nil {
		profile, _ := s.contextService.GetContextProfile(ctx, userID, "node", nodeID)
		nc.Profile = profile
	}

	// 4. Load memories for this specific node
	memories, err := s.getNodeMemories(ctx, userID, nodeID, 10)
	if err == nil {
		nc.Memories = memories
	}

	// 5. Get projects under this node
	projects, err := s.getNodeProjects(ctx, userID, nodeID, 10)
	if err == nil {
		nc.Projects = projects
	}

	return nc, nil
}

// getNode retrieves node details
func (s *ProjectContextService) getNode(ctx context.Context, userID string, nodeID uuid.UUID) (*Node, error) {
	var n Node

	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, type, description, parent_id, created_at, updated_at
		FROM nodes
		WHERE id = $1 AND user_id = $2
	`, nodeID, userID).Scan(
		&n.ID, &n.UserID, &n.Name, &n.Type, &n.Description, &n.ParentID, &n.CreatedAt, &n.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("node not found")
	}
	if err != nil {
		return nil, err
	}

	return &n, nil
}

// getNodeAncestors retrieves ancestor nodes (path to root)
func (s *ProjectContextService) getNodeAncestors(ctx context.Context, userID string, nodeID uuid.UUID) ([]*Node, error) {
	// Use recursive CTE to get ancestors
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE ancestors AS (
			SELECT id, user_id, name, type, description, parent_id, created_at, updated_at, 0 as depth
			FROM nodes
			WHERE id = $1 AND user_id = $2

			UNION ALL

			SELECT n.id, n.user_id, n.name, n.type, n.description, n.parent_id, n.created_at, n.updated_at, a.depth + 1
			FROM nodes n
			JOIN ancestors a ON n.id = a.parent_id
			WHERE n.user_id = $2
		)
		SELECT id, user_id, name, type, description, parent_id, created_at, updated_at
		FROM ancestors
		WHERE depth > 0
		ORDER BY depth DESC
	`, nodeID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ancestors []*Node
	for rows.Next() {
		var n Node
		err := rows.Scan(&n.ID, &n.UserID, &n.Name, &n.Type, &n.Description, &n.ParentID, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			continue
		}
		ancestors = append(ancestors, &n)
	}

	return ancestors, nil
}

// getNodeMemories retrieves memories for a node
func (s *ProjectContextService) getNodeMemories(ctx context.Context, userID string, nodeID uuid.UUID, limit int) ([]Memory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, title, summary, content, memory_type, category, source_type,
		       source_id, project_id, node_id, importance_score, access_count, is_pinned,
		       tags, created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND node_id = $2 AND is_active = true
		ORDER BY is_pinned DESC, importance_score DESC, created_at DESC
		LIMIT $3
	`, userID, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.UserID, &m.Title, &m.Summary, &m.Content, &m.MemoryType, &m.Category,
			&m.SourceType, &m.SourceID, &m.ProjectID, &m.NodeID, &m.ImportanceScore,
			&m.AccessCount, &m.IsPinned, &m.Tags, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			continue
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// getNodeProjects retrieves projects under a node
func (s *ProjectContextService) getNodeProjects(ctx context.Context, userID string, nodeID uuid.UUID, limit int) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, description, status, priority, client_id, created_at, updated_at
		FROM projects
		WHERE user_id = $1 AND node_id = $2 AND is_archived = false
		ORDER BY updated_at DESC
		LIMIT $3
	`, userID, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.Priority, &p.ClientID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// ============================================================================
// Context Injection Methods
// ============================================================================

// InjectContextIntoConversation prepares context for injection into a conversation
func (s *ProjectContextService) InjectContextIntoConversation(
	ctx context.Context,
	userID string,
	projectID, nodeID *uuid.UUID,
	agentType, focusMode string,
	maxTokens int,
) (*InjectedContext, error) {
	ic := &InjectedContext{
		LoadedItems:    make([]ContextItem, 0),
		TokenBreakdown: make(map[string]int),
	}

	items := make([]BudgetItem, 0, 32)
	items = append(items, BudgetItem{
		Key:      "header",
		Type:     "meta",
		Content:  "\n## Loaded Context\n\n",
		Priority: 100,
		Pinned:   true,
	})

	// Load project context if project is selected
	if projectID != nil {
		pc, err := s.LoadProjectContext(ctx, userID, *projectID)
		if err == nil {
			// Add project info (pinned)
			var psb strings.Builder
			psb.WriteString(fmt.Sprintf("### Active Project: %s\n", pc.Project.Name))
			if pc.Project.Description != "" {
				psb.WriteString(fmt.Sprintf("%s\n", pc.Project.Description))
			}
			psb.WriteString("\n")
			items = append(items, BudgetItem{Key: "project", Type: "project", Content: psb.String(), Priority: 90, Pinned: true})

			// Add memories
			if len(pc.Memories) > 0 {
				for _, m := range pc.Memories {
					items = append(items, BudgetItem{
						Key:      "memory:" + m.ID.String(),
						Type:     "memory",
						Content:  fmt.Sprintf("### Relevant Memories\n- [%s] %s: %s\n\n", m.MemoryType, m.Title, m.Summary),
						Priority: 60,
						Pinned:   false,
					})
					ic.LoadedItems = append(ic.LoadedItems, ContextItem{
						ID:    m.ID,
						Type:  "memory",
						Title: m.Title,
					})
				}
			}

			// Add user facts
			if len(pc.UserFacts) > 0 {
				var fsb strings.Builder
				fsb.WriteString("### User Facts\n")
				for _, f := range pc.UserFacts {
					fsb.WriteString(fmt.Sprintf("- %s: %s\n", f.FactKey, f.FactValue))
				}
				fsb.WriteString("\n")
				items = append(items, BudgetItem{Key: "user_facts", Type: "user_facts", Content: fsb.String(), Priority: 80, Pinned: true})
			}

			// Add recent conversation context
			if len(pc.Conversations) > 0 {
				for _, c := range pc.Conversations {
					items = append(items, BudgetItem{
						Key:      "conversation:" + c.ID.String(),
						Type:     "conversation",
						Content:  fmt.Sprintf("### Recent Discussion Context\n- %s\n\n", c.Summary),
						Priority: 50,
						Pinned:   false,
					})
				}
			}
		}
	}

	// Load node context if node is selected
	if nodeID != nil {
		nc, err := s.LoadNodeContext(ctx, userID, *nodeID)
		if err == nil {
			var nsb strings.Builder
			nsb.WriteString(fmt.Sprintf("### Business Context: %s (%s)\n", nc.Node.Name, nc.Node.Type))
			if nc.Node.Description != "" {
				nsb.WriteString(fmt.Sprintf("%s\n", nc.Node.Description))
			}
			nsb.WriteString("\n")
			items = append(items, BudgetItem{Key: "node", Type: "node", Content: nsb.String(), Priority: 70, Pinned: true})

			// Add node memories
			if len(nc.Memories) > 0 {
				for _, m := range nc.Memories {
					items = append(items, BudgetItem{
						Key:      "node_memory:" + m.ID.String(),
						Type:     "node_memory",
						Content:  fmt.Sprintf("### Node-Specific Memories\n- [%s] %s\n\n", m.MemoryType, m.Summary),
						Priority: 40,
						Pinned:   false,
					})
				}
			}
		}
	}

	res := ApplyTokenBudget(items, maxTokens)

	// Rebuild prompt addition from kept items
	var sb strings.Builder
	for _, it := range res.Kept {
		if it.Content == "" {
			continue
		}
		sb.WriteString(it.Content)
	}
	ic.SystemPromptAddition = sb.String()
	ic.TotalTokens = res.UsedTokens

	// Token breakdown from kept items
	ic.TokenBreakdown = make(map[string]int)
	for _, it := range res.Kept {
		if it.Type == "meta" {
			continue
		}
		it = ensureTokenCount(it)
		ic.TokenBreakdown[it.Type] += it.TokenCount
	}

	return ic, nil
}
