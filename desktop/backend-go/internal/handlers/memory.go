package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// MemoryHandler handles memory-related HTTP endpoints
type MemoryHandler struct {
	pool             *pgxpool.Pool
	embeddingService *services.EmbeddingService
}

// NewMemoryHandler creates a new MemoryHandler
func NewMemoryHandler(pool *pgxpool.Pool, embeddingService *services.EmbeddingService) *MemoryHandler {
	return &MemoryHandler{
		pool:             pool,
		embeddingService: embeddingService,
	}
}

// ================================================
// RESPONSE TYPES
// ================================================

// MemoryResponse represents a memory in API responses
type MemoryResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	Content         string                 `json:"content"`
	MemoryType      string                 `json:"memory_type"`
	Category        *string                `json:"category"`
	SourceType      string                 `json:"source_type"`
	SourceID        *string                `json:"source_id"`
	SourceContext   *string                `json:"source_context"`
	ProjectID       *string                `json:"project_id"`
	NodeID          *string                `json:"node_id"`
	ImportanceScore float64                `json:"importance_score"`
	AccessCount     int                    `json:"access_count"`
	LastAccessedAt  *string                `json:"last_accessed_at"`
	IsActive        bool                   `json:"is_active"`
	IsPinned        bool                   `json:"is_pinned"`
	ExpiresAt       *string                `json:"expires_at"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// UserFactResponse represents a user fact in API responses
type UserFactResponse struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	FactKey         string  `json:"fact_key"`
	FactValue       string  `json:"fact_value"`
	FactType        string  `json:"fact_type"`
	SourceMemoryID  *string `json:"source_memory_id"`
	ConfidenceScore float64 `json:"confidence_score"`
	IsActive        bool    `json:"is_active"`
	LastConfirmedAt *string `json:"last_confirmed_at"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// MemoryStatsResponse contains memory statistics
type MemoryStatsResponse struct {
	TotalMemories     int            `json:"total_memories"`
	ActiveMemories    int            `json:"active_memories"`
	PinnedMemories    int            `json:"pinned_memories"`
	ByType            map[string]int `json:"by_type"`
	ByCategory        map[string]int `json:"by_category"`
	TotalFacts        int            `json:"total_facts"`
	RecentAccessCount int            `json:"recent_access_count"`
}

// ================================================
// REQUEST TYPES
// ================================================

// CreateMemoryRequest represents a request to create a memory
type CreateMemoryRequest struct {
	Title         string                 `json:"title" binding:"required"`
	Summary       string                 `json:"summary" binding:"required"`
	Content       string                 `json:"content" binding:"required"`
	MemoryType    string                 `json:"memory_type" binding:"required"` // fact, preference, decision, pattern, insight, interaction, learning
	Category      *string                `json:"category"`
	SourceType    string                 `json:"source_type" binding:"required"` // conversation, voice_note, document, task, project, manual, inferred
	SourceID      *string                `json:"source_id"`
	SourceContext *string                `json:"source_context"`
	ProjectID     *string                `json:"project_id"`
	NodeID        *string                `json:"node_id"`
	Tags          []string               `json:"tags"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateMemoryRequest represents a request to update a memory
type UpdateMemoryRequest struct {
	Title           *string                `json:"title"`
	Summary         *string                `json:"summary"`
	Content         *string                `json:"content"`
	MemoryType      *string                `json:"memory_type"`
	Category        *string                `json:"category"`
	ImportanceScore *float64               `json:"importance_score"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	IsActive        *bool                  `json:"is_active"`
}

// MemorySearchRequest represents a semantic search request
type MemorySearchRequest struct {
	Query      string   `json:"query" binding:"required"`
	MemoryType *string  `json:"memory_type"`
	ProjectID  *string  `json:"project_id"`
	NodeID     *string  `json:"node_id"`
	Limit      int      `json:"limit"`
	Tags       []string `json:"tags"`
}

// RelevantMemoriesRequest represents a request for relevant memories
type RelevantMemoriesRequest struct {
	Context        string   `json:"context" binding:"required"`
	ConversationID *string  `json:"conversation_id"`
	ProjectID      *string  `json:"project_id"`
	NodeID         *string  `json:"node_id"`
	Limit          int      `json:"limit"`
	MemoryTypes    []string `json:"memory_types"`
}

// UpdateFactRequest represents a request to update a user fact
type UpdateFactRequest struct {
	FactValue       string   `json:"fact_value" binding:"required"`
	FactType        *string  `json:"fact_type"`
	ConfidenceScore *float64 `json:"confidence_score"`
}

// ================================================
// MEMORY CRUD HANDLERS
// ================================================

// ListMemories returns all memories for the current user
// GET /api/memories
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse query params
	memoryType := c.Query("type")
	category := c.Query("category")
	pinnedOnly := c.Query("pinned") == "true"
	activeOnly := c.Query("active") != "false" // Default to active only
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Build query
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at
		FROM memories
		WHERE user_id = $1
	`
	args := []interface{}{user.ID}
	argIdx := 2

	if activeOnly {
		query += ` AND is_active = true`
	}
	if pinnedOnly {
		query += ` AND is_pinned = true`
	}
	if memoryType != "" {
		query += ` AND memory_type = $` + strconv.Itoa(argIdx)
		args = append(args, memoryType)
		argIdx++
	}
	if category != "" {
		query += ` AND category = $` + strconv.Itoa(argIdx)
		args = append(args, category)
		argIdx++
	}

	query += ` ORDER BY is_pinned DESC, importance_score DESC, created_at DESC`
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list memories"})
		return
	}
	defer rows.Close()

	memories := []MemoryResponse{}
	for rows.Next() {
		memory, err := scanMemoryRow(rows)
		if err != nil {
			continue
		}
		memories = append(memories, memory)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"count":    len(memories),
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateMemory creates a new memory
// POST /api/memories
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate memory type
	validTypes := map[string]bool{
		"fact": true, "preference": true, "decision": true,
		"pattern": true, "insight": true, "interaction": true, "learning": true,
	}
	if !validTypes[req.MemoryType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory_type"})
		return
	}

	// Validate source type
	validSources := map[string]bool{
		"conversation": true, "voice_note": true, "document": true,
		"task": true, "project": true, "manual": true, "inferred": true,
	}
	if !validSources[req.SourceType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source_type"})
		return
	}

	// Generate embedding if service is available
	var embeddingJSON []byte
	var embeddingModel *string
	if h.embeddingService != nil {
		textToEmbed := req.Title + " " + req.Summary + " " + req.Content
		embedding, err := h.embeddingService.GenerateEmbedding(c.Request.Context(), textToEmbed)
		if err == nil && len(embedding) > 0 {
			embeddingJSON, _ = json.Marshal(embedding)
			model := "text-embedding-ada-002"
			embeddingModel = &model
		}
	}

	// Serialize tags and metadata
	tagsJSON, _ := json.Marshal(req.Tags)
	if req.Tags == nil {
		tagsJSON = []byte("[]")
	}
	metadataJSON, _ := json.Marshal(req.Metadata)
	if req.Metadata == nil {
		metadataJSON = []byte("{}")
	}

	// Parse optional UUIDs
	var sourceID, projectID, nodeID *uuid.UUID
	if req.SourceID != nil {
		if parsed, err := uuid.Parse(*req.SourceID); err == nil {
			sourceID = &parsed
		}
	}
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = &parsed
		}
	}
	if req.NodeID != nil {
		if parsed, err := uuid.Parse(*req.NodeID); err == nil {
			nodeID = &parsed
		}
	}

	query := `
		INSERT INTO memories (
			user_id, title, summary, content, memory_type, category,
			source_type, source_id, source_context, project_id, node_id,
			tags, metadata, embedding, embedding_model
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, user_id, title, summary, content, memory_type, category,
		          source_type, source_id, source_context, project_id, node_id,
		          importance_score, access_count, last_accessed_at,
		          is_active, is_pinned, expires_at, tags, metadata,
		          created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), query,
		user.ID, req.Title, req.Summary, req.Content, req.MemoryType, req.Category,
		req.SourceType, sourceID, req.SourceContext, projectID, nodeID,
		tagsJSON, metadataJSON, embeddingJSON, embeddingModel,
	)

	memory, err := scanMemoryRowSingle(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create memory: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, memory)
}

// GetMemory returns a specific memory
// GET /api/memories/:id
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at
		FROM memories
		WHERE id = $1 AND user_id = $2
	`

	row := h.pool.QueryRow(c.Request.Context(), query, id, user.ID)
	memory, err := scanMemoryRowSingle(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory"})
		}
		return
	}

	// Update access count and last accessed
	go h.recordMemoryAccess(id, user.ID, "user_view", nil, nil)

	c.JSON(http.StatusOK, memory)
}

// UpdateMemory updates a memory
// PUT /api/memories/:id
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership and get existing
	var existingTitle, existingSummary, existingContent, existingType string
	var existingCategory *string
	var existingScore float64
	var existingActive bool
	var existingTags, existingMetadata []byte

	checkQuery := `
		SELECT title, summary, content, memory_type, category, importance_score, is_active, tags, metadata
		FROM memories WHERE id = $1 AND user_id = $2
	`
	err = h.pool.QueryRow(c.Request.Context(), checkQuery, id, user.ID).Scan(
		&existingTitle, &existingSummary, &existingContent, &existingType,
		&existingCategory, &existingScore, &existingActive, &existingTags, &existingMetadata,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify memory"})
		}
		return
	}

	// Apply updates
	title := existingTitle
	if req.Title != nil {
		title = *req.Title
	}
	summary := existingSummary
	if req.Summary != nil {
		summary = *req.Summary
	}
	content := existingContent
	if req.Content != nil {
		content = *req.Content
	}
	memoryType := existingType
	if req.MemoryType != nil {
		memoryType = *req.MemoryType
	}
	category := existingCategory
	if req.Category != nil {
		category = req.Category
	}
	importanceScore := existingScore
	if req.ImportanceScore != nil {
		importanceScore = *req.ImportanceScore
	}
	isActive := existingActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tagsJSON := existingTags
	if req.Tags != nil {
		tagsJSON, _ = json.Marshal(req.Tags)
	}
	metadataJSON := existingMetadata
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	// Regenerate embedding if content changed
	var embeddingJSON []byte
	if req.Title != nil || req.Summary != nil || req.Content != nil {
		if h.embeddingService != nil {
			textToEmbed := title + " " + summary + " " + content
			embedding, err := h.embeddingService.GenerateEmbedding(c.Request.Context(), textToEmbed)
			if err == nil && len(embedding) > 0 {
				embeddingJSON, _ = json.Marshal(embedding)
			}
		}
	}

	updateQuery := `
		UPDATE memories
		SET title = $1, summary = $2, content = $3, memory_type = $4, category = $5,
		    importance_score = $6, is_active = $7, tags = $8, metadata = $9,
		    embedding = COALESCE($10, embedding), updated_at = NOW()
		WHERE id = $11 AND user_id = $12
		RETURNING id, user_id, title, summary, content, memory_type, category,
		          source_type, source_id, source_context, project_id, node_id,
		          importance_score, access_count, last_accessed_at,
		          is_active, is_pinned, expires_at, tags, metadata,
		          created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), updateQuery,
		title, summary, content, memoryType, category,
		importanceScore, isActive, tagsJSON, metadataJSON,
		embeddingJSON, id, user.ID,
	)

	memory, err := scanMemoryRowSingle(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memory"})
		return
	}

	c.JSON(http.StatusOK, memory)
}

// DeleteMemory deletes a memory
// DELETE /api/memories/:id
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	query := `DELETE FROM memories WHERE id = $1 AND user_id = $2`
	result, err := h.pool.Exec(c.Request.Context(), query, id, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete memory"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted"})
}

// PinMemory pins or unpins a memory
// POST /api/memories/:id/pin
func (h *MemoryHandler) PinMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req struct {
		Pinned bool `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to toggle
		req.Pinned = true
	}

	query := `
		UPDATE memories
		SET is_pinned = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, title, summary, content, memory_type, category,
		          source_type, source_id, source_context, project_id, node_id,
		          importance_score, access_count, last_accessed_at,
		          is_active, is_pinned, expires_at, tags, metadata,
		          created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), query, req.Pinned, id, user.ID)
	memory, err := scanMemoryRowSingle(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memory"})
		}
		return
	}

	c.JSON(http.StatusOK, memory)
}

// ================================================
// SEMANTIC SEARCH HANDLERS
// ================================================

// SearchMemories performs semantic search on memories
// POST /api/memories/search
func (h *MemoryHandler) SearchMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req MemorySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Generate query embedding
	if h.embeddingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Embedding service not available"})
		return
	}

	queryEmbedding, err := h.embeddingService.GenerateEmbedding(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embedding"})
		return
	}

	embeddingJSON, _ := json.Marshal(queryEmbedding)

	// Build semantic search query
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at,
		       1 - (embedding <=> $1::vector) as similarity
		FROM memories
		WHERE user_id = $2 AND is_active = true AND embedding IS NOT NULL
	`
	args := []interface{}{string(embeddingJSON), user.ID}
	argIdx := 3

	if req.MemoryType != nil {
		query += ` AND memory_type = $` + strconv.Itoa(argIdx)
		args = append(args, *req.MemoryType)
		argIdx++
	}
	if req.ProjectID != nil {
		query += ` AND project_id = $` + strconv.Itoa(argIdx)
		args = append(args, *req.ProjectID)
		argIdx++
	}
	if req.NodeID != nil {
		query += ` AND node_id = $` + strconv.Itoa(argIdx)
		args = append(args, *req.NodeID)
		argIdx++
	}

	query += ` ORDER BY embedding <=> $1::vector LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, req.Limit)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	defer rows.Close()

	type SearchResult struct {
		MemoryResponse
		Similarity float64 `json:"similarity"`
	}

	results := []SearchResult{}
	for rows.Next() {
		var similarity float64
		memory, err := scanMemoryRowWithExtra(rows, &similarity)
		if err != nil {
			continue
		}
		results = append(results, SearchResult{
			MemoryResponse: memory,
			Similarity:     similarity,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
		"query":   req.Query,
	})
}

// GetRelevantMemories gets memories relevant to a context
// POST /api/memories/relevant
func (h *MemoryHandler) GetRelevantMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req RelevantMemoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}

	// Generate embedding for context
	if h.embeddingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Embedding service not available"})
		return
	}

	contextEmbedding, err := h.embeddingService.GenerateEmbedding(c.Request.Context(), req.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embedding"})
		return
	}

	embeddingJSON, _ := json.Marshal(contextEmbedding)

	// Build query with hybrid scoring (semantic + recency + importance)
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at,
		       (1 - (embedding <=> $1::vector)) * 0.6 +
		       importance_score * 0.2 +
		       CASE WHEN is_pinned THEN 0.2 ELSE 0 END as relevance_score
		FROM memories
		WHERE user_id = $2 AND is_active = true AND embedding IS NOT NULL
	`
	args := []interface{}{string(embeddingJSON), user.ID}
	argIdx := 3

	if req.ProjectID != nil {
		query += ` AND (project_id = $` + strconv.Itoa(argIdx) + ` OR project_id IS NULL)`
		args = append(args, *req.ProjectID)
		argIdx++
	}
	if req.NodeID != nil {
		query += ` AND (node_id = $` + strconv.Itoa(argIdx) + ` OR node_id IS NULL)`
		args = append(args, *req.NodeID)
		argIdx++
	}
	if len(req.MemoryTypes) > 0 {
		query += ` AND memory_type = ANY($` + strconv.Itoa(argIdx) + `)`
		args = append(args, req.MemoryTypes)
		argIdx++
	}

	query += ` ORDER BY relevance_score DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, req.Limit)

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get relevant memories"})
		return
	}
	defer rows.Close()

	type RelevantResult struct {
		MemoryResponse
		RelevanceScore float64 `json:"relevance_score"`
	}

	results := []RelevantResult{}
	for rows.Next() {
		var relevance float64
		memory, err := scanMemoryRowWithExtra(rows, &relevance)
		if err != nil {
			continue
		}
		results = append(results, RelevantResult{
			MemoryResponse: memory,
			RelevanceScore: relevance,
		})

		// Log access for learning
		if memID, err := uuid.Parse(memory.ID); err == nil {
			go h.recordMemoryAccess(memID, user.ID, "auto_inject", req.ConversationID, &relevance)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": results,
		"count":    len(results),
	})
}

// ================================================
// PROJECT/NODE SCOPED HANDLERS
// ================================================

// GetProjectMemories returns memories for a specific project
// GET /api/memories/project/:projectId
func (h *MemoryHandler) GetProjectMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND project_id = $2 AND is_active = true
		ORDER BY is_pinned DESC, importance_score DESC, created_at DESC
		LIMIT $3
	`

	rows, err := h.pool.Query(c.Request.Context(), query, user.ID, projectID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project memories"})
		return
	}
	defer rows.Close()

	memories := []MemoryResponse{}
	for rows.Next() {
		memory, err := scanMemoryRow(rows)
		if err != nil {
			continue
		}
		memories = append(memories, memory)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories":   memories,
		"count":      len(memories),
		"project_id": projectID.String(),
	})
}

// GetNodeMemories returns memories for a specific node
// GET /api/memories/node/:nodeId
func (h *MemoryHandler) GetNodeMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	nodeID, err := uuid.Parse(c.Param("nodeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, source_context, project_id, node_id,
		       importance_score, access_count, last_accessed_at,
		       is_active, is_pinned, expires_at, tags, metadata,
		       created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND node_id = $2 AND is_active = true
		ORDER BY is_pinned DESC, importance_score DESC, created_at DESC
		LIMIT $3
	`

	rows, err := h.pool.Query(c.Request.Context(), query, user.ID, nodeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get node memories"})
		return
	}
	defer rows.Close()

	memories := []MemoryResponse{}
	for rows.Next() {
		memory, err := scanMemoryRow(rows)
		if err != nil {
			continue
		}
		memories = append(memories, memory)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": memories,
		"count":    len(memories),
		"node_id":  nodeID.String(),
	})
}

// ================================================
// USER FACTS HANDLERS
// ================================================

// ListUserFacts returns all user facts
// GET /api/user-facts
func (h *MemoryHandler) ListUserFacts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	factType := c.Query("type")
	activeOnly := c.Query("active") != "false"

	query := `
		SELECT id, user_id, fact_key, fact_value, fact_type,
		       source_memory_id, confidence_score, is_active,
		       last_confirmed_at, created_at, updated_at
		FROM user_facts
		WHERE user_id = $1
	`
	args := []interface{}{user.ID}
	argIdx := 2

	if activeOnly {
		query += ` AND is_active = true`
	}
	if factType != "" {
		query += ` AND fact_type = $` + strconv.Itoa(argIdx)
		args = append(args, factType)
	}

	query += ` ORDER BY fact_type, fact_key`

	rows, err := h.pool.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list user facts"})
		return
	}
	defer rows.Close()

	facts := []UserFactResponse{}
	for rows.Next() {
		fact, err := scanUserFactRow(rows)
		if err != nil {
			continue
		}
		facts = append(facts, fact)
	}

	c.JSON(http.StatusOK, gin.H{
		"facts": facts,
		"count": len(facts),
	})
}

// UpdateUserFact updates or creates a user fact
// PUT /api/user-facts/:key
func (h *MemoryHandler) UpdateUserFact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	factKey := c.Param("key")
	if factKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fact key is required"})
		return
	}

	var req UpdateFactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	factType := "fact"
	if req.FactType != nil {
		factType = *req.FactType
	}

	confidence := 1.0
	if req.ConfidenceScore != nil {
		confidence = *req.ConfidenceScore
	}

	// Upsert fact
	query := `
		INSERT INTO user_facts (user_id, fact_key, fact_value, fact_type, confidence_score, last_confirmed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, fact_key)
		DO UPDATE SET
			fact_value = EXCLUDED.fact_value,
			fact_type = EXCLUDED.fact_type,
			confidence_score = EXCLUDED.confidence_score,
			last_confirmed_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, fact_key, fact_value, fact_type,
		          source_memory_id, confidence_score, is_active,
		          last_confirmed_at, created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), query, user.ID, factKey, req.FactValue, factType, confidence)
	fact, err := scanUserFactRowSingle(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fact"})
		return
	}

	c.JSON(http.StatusOK, fact)
}

// ConfirmUserFact marks a user fact as confirmed/active
// POST /api/user-facts/:key/confirm
func (h *MemoryHandler) ConfirmUserFact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	factKey := c.Param("key")
	if factKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fact key is required"})
		return
	}

	query := `
		UPDATE user_facts
		SET is_active = true,
		    last_confirmed_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1 AND fact_key = $2
		RETURNING id, user_id, fact_key, fact_value, fact_type,
		          source_memory_id, confidence_score, is_active,
		          last_confirmed_at, created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), query, user.ID, factKey)
	fact, err := scanUserFactRowSingle(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm fact"})
		return
	}

	c.JSON(http.StatusOK, fact)
}

// RejectUserFact marks a user fact as inactive (not injected)
// POST /api/user-facts/:key/reject
func (h *MemoryHandler) RejectUserFact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	factKey := c.Param("key")
	if factKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fact key is required"})
		return
	}

	query := `
		UPDATE user_facts
		SET is_active = false,
		    updated_at = NOW()
		WHERE user_id = $1 AND fact_key = $2
		RETURNING id, user_id, fact_key, fact_value, fact_type,
		          source_memory_id, confidence_score, is_active,
		          last_confirmed_at, created_at, updated_at
	`

	row := h.pool.QueryRow(c.Request.Context(), query, user.ID, factKey)
	fact, err := scanUserFactRowSingle(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject fact"})
		return
	}

	c.JSON(http.StatusOK, fact)
}

// DeleteUserFact deletes a user fact
// DELETE /api/user-facts/:key
func (h *MemoryHandler) DeleteUserFact(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	factKey := c.Param("key")
	if factKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fact key is required"})
		return
	}

	query := `DELETE FROM user_facts WHERE user_id = $1 AND fact_key = $2`
	result, err := h.pool.Exec(c.Request.Context(), query, user.ID, factKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete fact"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fact not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fact deleted"})
}

// ================================================
// STATISTICS HANDLER
// ================================================

// GetMemoryStats returns memory statistics for the user
// GET /api/memories/stats
func (h *MemoryHandler) GetMemoryStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	stats := MemoryStatsResponse{
		ByType:     make(map[string]int),
		ByCategory: make(map[string]int),
	}

	// Total and active memories
	var total, active, pinned int
	err := h.pool.QueryRow(c.Request.Context(), `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_active = true),
			COUNT(*) FILTER (WHERE is_pinned = true)
		FROM memories WHERE user_id = $1
	`, user.ID).Scan(&total, &active, &pinned)
	if err == nil {
		stats.TotalMemories = total
		stats.ActiveMemories = active
		stats.PinnedMemories = pinned
	}

	// By type
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT memory_type, COUNT(*)
		FROM memories WHERE user_id = $1 AND is_active = true
		GROUP BY memory_type
	`, user.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			var count int
			if rows.Scan(&t, &count) == nil {
				stats.ByType[t] = count
			}
		}
	}

	// By category
	rows, err = h.pool.Query(c.Request.Context(), `
		SELECT COALESCE(category, 'uncategorized'), COUNT(*)
		FROM memories WHERE user_id = $1 AND is_active = true
		GROUP BY category
	`, user.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cat string
			var count int
			if rows.Scan(&cat, &count) == nil {
				stats.ByCategory[cat] = count
			}
		}
	}

	// Total facts
	h.pool.QueryRow(c.Request.Context(), `
		SELECT COUNT(*) FROM user_facts WHERE user_id = $1 AND is_active = true
	`, user.ID).Scan(&stats.TotalFacts)

	// Recent access count (last 7 days)
	h.pool.QueryRow(c.Request.Context(), `
		SELECT COUNT(*) FROM memory_access_log
		WHERE user_id = $1 AND created_at > NOW() - INTERVAL '7 days'
	`, user.ID).Scan(&stats.RecentAccessCount)

	c.JSON(http.StatusOK, stats)
}

// ================================================
// HELPER FUNCTIONS
// ================================================

func (h *MemoryHandler) recordMemoryAccess(memoryID uuid.UUID, userID, accessType string, conversationID *string, relevanceScore *float64) {
	query := `
		INSERT INTO memory_access_log (memory_id, user_id, access_type, conversation_id, relevance_score)
		VALUES ($1, $2, $3, $4, $5)
	`
	var convID *uuid.UUID
	if conversationID != nil {
		if parsed, err := uuid.Parse(*conversationID); err == nil {
			convID = &parsed
		}
	}
	// Use background context for async logging
	h.pool.Exec(context.Background(), query, memoryID, userID, accessType, convID, relevanceScore)
}

func scanMemoryRow(rows pgx.Rows) (MemoryResponse, error) {
	var m MemoryResponse
	var id, sourceID, projectID, nodeID uuid.UUID
	var sourceIDValid, projectIDValid, nodeIDValid bool
	var lastAccessed, expires *time.Time
	var tags, metadata []byte

	err := rows.Scan(
		&id, &m.UserID, &m.Title, &m.Summary, &m.Content, &m.MemoryType, &m.Category,
		&m.SourceType, &sourceID, &m.SourceContext, &projectID, &nodeID,
		&m.ImportanceScore, &m.AccessCount, &lastAccessed,
		&m.IsActive, &m.IsPinned, &expires, &tags, &metadata,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return m, err
	}

	m.ID = id.String()
	if sourceIDValid {
		s := sourceID.String()
		m.SourceID = &s
	}
	if projectIDValid {
		s := projectID.String()
		m.ProjectID = &s
	}
	if nodeIDValid {
		s := nodeID.String()
		m.NodeID = &s
	}
	if lastAccessed != nil {
		s := lastAccessed.Format(time.RFC3339)
		m.LastAccessedAt = &s
	}
	if expires != nil {
		s := expires.Format(time.RFC3339)
		m.ExpiresAt = &s
	}

	json.Unmarshal(tags, &m.Tags)
	if m.Tags == nil {
		m.Tags = []string{}
	}
	json.Unmarshal(metadata, &m.Metadata)
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}

	return m, nil
}

func scanMemoryRowSingle(row pgx.Row) (MemoryResponse, error) {
	var m MemoryResponse
	var id uuid.UUID
	var sourceID, projectID, nodeID *uuid.UUID
	var lastAccessed, expires *time.Time
	var tags, metadata []byte
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&id, &m.UserID, &m.Title, &m.Summary, &m.Content, &m.MemoryType, &m.Category,
		&m.SourceType, &sourceID, &m.SourceContext, &projectID, &nodeID,
		&m.ImportanceScore, &m.AccessCount, &lastAccessed,
		&m.IsActive, &m.IsPinned, &expires, &tags, &metadata,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return m, err
	}

	m.ID = id.String()
	m.CreatedAt = createdAt.Format(time.RFC3339)
	m.UpdatedAt = updatedAt.Format(time.RFC3339)

	if sourceID != nil {
		s := sourceID.String()
		m.SourceID = &s
	}
	if projectID != nil {
		s := projectID.String()
		m.ProjectID = &s
	}
	if nodeID != nil {
		s := nodeID.String()
		m.NodeID = &s
	}
	if lastAccessed != nil {
		s := lastAccessed.Format(time.RFC3339)
		m.LastAccessedAt = &s
	}
	if expires != nil {
		s := expires.Format(time.RFC3339)
		m.ExpiresAt = &s
	}

	json.Unmarshal(tags, &m.Tags)
	if m.Tags == nil {
		m.Tags = []string{}
	}
	json.Unmarshal(metadata, &m.Metadata)
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}

	return m, nil
}

func scanMemoryRowWithExtra(rows pgx.Rows, extra *float64) (MemoryResponse, error) {
	var m MemoryResponse
	var id uuid.UUID
	var sourceID, projectID, nodeID *uuid.UUID
	var lastAccessed, expires *time.Time
	var tags, metadata []byte
	var createdAt, updatedAt time.Time

	err := rows.Scan(
		&id, &m.UserID, &m.Title, &m.Summary, &m.Content, &m.MemoryType, &m.Category,
		&m.SourceType, &sourceID, &m.SourceContext, &projectID, &nodeID,
		&m.ImportanceScore, &m.AccessCount, &lastAccessed,
		&m.IsActive, &m.IsPinned, &expires, &tags, &metadata,
		&createdAt, &updatedAt, extra,
	)
	if err != nil {
		return m, err
	}

	m.ID = id.String()
	m.CreatedAt = createdAt.Format(time.RFC3339)
	m.UpdatedAt = updatedAt.Format(time.RFC3339)

	if sourceID != nil {
		s := sourceID.String()
		m.SourceID = &s
	}
	if projectID != nil {
		s := projectID.String()
		m.ProjectID = &s
	}
	if nodeID != nil {
		s := nodeID.String()
		m.NodeID = &s
	}
	if lastAccessed != nil {
		s := lastAccessed.Format(time.RFC3339)
		m.LastAccessedAt = &s
	}
	if expires != nil {
		s := expires.Format(time.RFC3339)
		m.ExpiresAt = &s
	}

	json.Unmarshal(tags, &m.Tags)
	if m.Tags == nil {
		m.Tags = []string{}
	}
	json.Unmarshal(metadata, &m.Metadata)
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}

	return m, nil
}

func scanUserFactRow(rows pgx.Rows) (UserFactResponse, error) {
	var f UserFactResponse
	var id uuid.UUID
	var sourceMemoryID *uuid.UUID
	var lastConfirmed *time.Time
	var createdAt, updatedAt time.Time

	err := rows.Scan(
		&id, &f.UserID, &f.FactKey, &f.FactValue, &f.FactType,
		&sourceMemoryID, &f.ConfidenceScore, &f.IsActive,
		&lastConfirmed, &createdAt, &updatedAt,
	)
	if err != nil {
		return f, err
	}

	f.ID = id.String()
	f.CreatedAt = createdAt.Format(time.RFC3339)
	f.UpdatedAt = updatedAt.Format(time.RFC3339)

	if sourceMemoryID != nil {
		s := sourceMemoryID.String()
		f.SourceMemoryID = &s
	}
	if lastConfirmed != nil {
		s := lastConfirmed.Format(time.RFC3339)
		f.LastConfirmedAt = &s
	}

	return f, nil
}

func scanUserFactRowSingle(row pgx.Row) (UserFactResponse, error) {
	var f UserFactResponse
	var id uuid.UUID
	var sourceMemoryID *uuid.UUID
	var lastConfirmed *time.Time
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&id, &f.UserID, &f.FactKey, &f.FactValue, &f.FactType,
		&sourceMemoryID, &f.ConfidenceScore, &f.IsActive,
		&lastConfirmed, &createdAt, &updatedAt,
	)
	if err != nil {
		return f, err
	}

	f.ID = id.String()
	f.CreatedAt = createdAt.Format(time.RFC3339)
	f.UpdatedAt = updatedAt.Format(time.RFC3339)

	if sourceMemoryID != nil {
		s := sourceMemoryID.String()
		f.SourceMemoryID = &s
	}
	if lastConfirmed != nil {
		s := lastConfirmed.Format(time.RFC3339)
		f.LastConfirmedAt = &s
	}

	return f, nil
}
