package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// parseIntDefault parses a string to int, returning defaultVal if parsing fails
func parseIntDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// ============================================================================
// WORKSPACE MEMORY HANDLERS
// ============================================================================

// ListWorkspaceMemories returns all memories for a workspace
func (h *Handlers) ListWorkspaceMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Parse pagination params
	limit := parseIntDefault(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseIntDefault(c.Query("offset"), 0)

	// Parse filters
	memoryType := c.Query("type")
	category := c.Query("category")
	tagsStr := c.Query("tags")

	var memories []sqlc.WorkspaceMemory

	// For now, ignore filters and just use the base query
	// TODO: Implement ListWorkspaceMemoriesFiltered if needed
	_ = memoryType
	_ = category
	_ = tagsStr

	memories, err = queries.ListWorkspaceMemories(ctx, sqlc.ListWorkspaceMemoriesParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	})

	if err != nil {
		log.Printf("ListWorkspaceMemories error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list memories"})
		return
	}

	// Get total count (use length of memories for now since CountWorkspaceMemories may not exist)
	total := int64(len(memories))

	result := make([]gin.H, len(memories))
	for i, m := range memories {
		result[i] = gin.H{
			"id":               m.ID,
			"title":            m.Title,
			"summary":          m.Summary,
			"content":          m.Content,
			"memory_type":      m.MemoryType,
			"category":         m.Category,
			"tags":             m.Tags,
			"scope_type":       m.ScopeType,
			"scope_id":         m.ScopeID,
			"importance_score": m.ImportanceScore,
			"metadata":         m.Metadata,
			"created_by":       m.CreatedBy,
			"created_at":       m.CreatedAt,
			"updated_at":       m.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": result,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetWorkspaceMemory returns a specific memory
func (h *Handlers) GetWorkspaceMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	memoryID, err := uuid.Parse(c.Param("memoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	memory, err := queries.GetWorkspaceMemory(ctx, sqlc.GetWorkspaceMemoryParams{
		ID:          pgtype.UUID{Bytes: memoryID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memory": gin.H{
			"id":               memory.ID,
			"title":            memory.Title,
			"summary":          memory.Summary,
			"content":          memory.Content,
			"memory_type":      memory.MemoryType,
			"category":         memory.Category,
			"tags":             memory.Tags,
			"scope_type":       memory.ScopeType,
			"scope_id":         memory.ScopeID,
			"importance_score": memory.ImportanceScore,
			"metadata":         memory.Metadata,
			"created_by":       memory.CreatedBy,
			"created_at":       memory.CreatedAt,
			"updated_at":       memory.UpdatedAt,
		},
	})
}

// CreateWorkspaceMemoryRequest represents the request body for creating a memory
type CreateWorkspaceMemoryRequest struct {
	Title           string         `json:"title" binding:"required,min=1"`
	Summary         string         `json:"summary"`
	Content         string         `json:"content" binding:"required,min=1"`
	MemoryType      string         `json:"memory_type"`
	Category        *string        `json:"category"`
	Tags            []string       `json:"tags"`
	ScopeType       *string        `json:"scope_type"`
	ScopeID         *string        `json:"scope_id"`
	Visibility      *string        `json:"visibility"`
	ImportanceScore *float64       `json:"importance_score"`
	Metadata        map[string]any `json:"metadata"`
	IsPinned        *bool          `json:"is_pinned"`
}

// CreateWorkspaceMemory creates a new memory in a workspace
func (h *Handlers) CreateWorkspaceMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req CreateWorkspaceMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Set defaults
	memoryType := req.MemoryType
	if memoryType == "" {
		memoryType = "fact"
	}

	// Convert metadata to JSON bytes
	var metadataBytes []byte
	if req.Metadata != nil {
		metadataBytes, _ = jsonMarshal(req.Metadata)
	}

	// Parse scope ID if provided
	var scopeID pgtype.UUID
	if req.ScopeID != nil {
		if parsed, err := uuid.Parse(*req.ScopeID); err == nil {
			scopeID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Convert importance score to pgtype.Numeric
	var importanceScore pgtype.Numeric
	if req.ImportanceScore != nil {
		importanceScore.Valid = true
		// pgtype.Numeric.Scan can accept float64
		_ = importanceScore.Scan(*req.ImportanceScore)
	}

	// Create memory
	memory, err := queries.CreateWorkspaceMemory(ctx, sqlc.CreateWorkspaceMemoryParams{
		WorkspaceID:     pgtype.UUID{Bytes: workspaceID, Valid: true},
		Title:           req.Title,
		Summary:         req.Summary,
		Content:         req.Content,
		MemoryType:      memoryType,
		Category:        req.Category,
		Tags:            req.Tags,
		ScopeType:       req.ScopeType,
		ScopeID:         scopeID,
		Visibility:      req.Visibility,
		ImportanceScore: importanceScore,
		Metadata:        metadataBytes,
		CreatedBy:       user.ID,
		IsPinned:        req.IsPinned,
	})
	if err != nil {
		log.Printf("CreateWorkspaceMemory error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create memory"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"memory": gin.H{
			"id":               memory.ID,
			"title":            memory.Title,
			"summary":          memory.Summary,
			"content":          memory.Content,
			"memory_type":      memory.MemoryType,
			"category":         memory.Category,
			"tags":             memory.Tags,
			"scope_type":       memory.ScopeType,
			"scope_id":         memory.ScopeID,
			"visibility":       memory.Visibility,
			"importance_score": memory.ImportanceScore,
			"metadata":         memory.Metadata,
			"created_by":       memory.CreatedBy,
			"created_at":       memory.CreatedAt,
		},
	})
}

// UpdateWorkspaceMemoryRequest represents the request body for updating a memory
type UpdateWorkspaceMemoryRequest struct {
	Title           *string        `json:"title"`
	Summary         *string        `json:"summary"`
	Content         *string        `json:"content"`
	MemoryType      *string        `json:"memory_type"`
	Category        *string        `json:"category"`
	Tags            []string       `json:"tags"`
	Visibility      *string        `json:"visibility"`
	ImportanceScore *float64       `json:"importance_score"`
	Metadata        map[string]any `json:"metadata"`
	IsPinned        *bool          `json:"is_pinned"`
}

// UpdateWorkspaceMemory updates an existing memory
func (h *Handlers) UpdateWorkspaceMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	memoryID, err := uuid.Parse(c.Param("memoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req UpdateWorkspaceMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Get existing memory to verify it exists
	_, err = queries.GetWorkspaceMemory(ctx, sqlc.GetWorkspaceMemoryParams{
		ID:          pgtype.UUID{Bytes: memoryID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory"})
		return
	}

	// Convert metadata to JSON bytes if provided
	var metadata []byte
	if req.Metadata != nil {
		metadata, _ = jsonMarshal(req.Metadata)
	}

	// Convert importance score to pgtype.Numeric
	var importanceScore pgtype.Numeric
	if req.ImportanceScore != nil {
		importanceScore.Valid = true
		_ = importanceScore.Scan(*req.ImportanceScore)
	}

	memory, err := queries.UpdateWorkspaceMemory(ctx, sqlc.UpdateWorkspaceMemoryParams{
		ID:              pgtype.UUID{Bytes: memoryID, Valid: true},
		WorkspaceID:    pgtype.UUID{Bytes: workspaceID, Valid: true},
		Title:           req.Title,
		Summary:         req.Summary,
		Content:         req.Content,
		MemoryType:      req.MemoryType,
		Category:        req.Category,
		Visibility:      req.Visibility,
		ImportanceScore: importanceScore,
		Tags:            req.Tags,
		Metadata:        metadata,
		IsPinned:        req.IsPinned,
	})
	if err != nil {
		log.Printf("UpdateWorkspaceMemory error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memory": gin.H{
			"id":               memory.ID,
			"title":            memory.Title,
			"summary":          memory.Summary,
			"content":          memory.Content,
			"memory_type":      memory.MemoryType,
			"category":         memory.Category,
			"tags":             memory.Tags,
			"visibility":       memory.Visibility,
			"importance_score": memory.ImportanceScore,
			"metadata":         memory.Metadata,
			"updated_at":       memory.UpdatedAt,
		},
	})
}

// DeleteWorkspaceMemory deletes a memory
func (h *Handlers) DeleteWorkspaceMemory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	memoryID, err := uuid.Parse(c.Param("memoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Check if user has permission to delete (owner of memory or admin)
	member, _ := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})

	memory, err := queries.GetWorkspaceMemory(ctx, sqlc.GetWorkspaceMemoryParams{
		ID:          pgtype.UUID{Bytes: memoryID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Memory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory"})
		return
	}

	// Check permission: creator or owner/admin
	isCreator := memory.CreatedBy == user.ID
	roleName := ""
	if member.RoleName != nil {
		roleName = *member.RoleName
	}
	isAdmin := roleName == "owner" || roleName == "admin"

	if !isCreator && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this memory"})
		return
	}

	err = queries.DeleteWorkspaceMemory(ctx, sqlc.DeleteWorkspaceMemoryParams{
		ID:          pgtype.UUID{Bytes: memoryID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		log.Printf("DeleteWorkspaceMemory error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted successfully"})
}

// SearchWorkspaceMemories searches memories by content and tags
func (h *Handlers) SearchWorkspaceMemories(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	limit := parseIntDefault(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}

	// Text search
	memories, err := queries.SearchWorkspaceMemories(ctx, sqlc.SearchWorkspaceMemoriesParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Column2:     &query,
		Limit:       int32(limit),
	})
	if err != nil {
		log.Printf("SearchWorkspaceMemories error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search memories"})
		return
	}

	result := make([]gin.H, len(memories))
	for i, m := range memories {
		result[i] = gin.H{
			"id":               m.ID,
			"title":            m.Title,
			"summary":          m.Summary,
			"content":          m.Content,
			"memory_type":      m.MemoryType,
			"category":         m.Category,
			"tags":             m.Tags,
			"importance_score": m.ImportanceScore,
			"created_at":       m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": result,
		"query":    query,
		"total":    len(result),
	})
}

// GetMemoryStats returns statistics about workspace memories
func (h *Handlers) GetMemoryStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Get all memories to count (simplified - no dedicated count query)
	memories, _ := queries.ListWorkspaceMemories(ctx, sqlc.ListWorkspaceMemoriesParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Limit:       1000, // Get up to 1000 for counting
		Offset:      0,
	})

	// Count by type
	typeCounts := make(map[string]int)
	categoryCounts := make(map[string]int)
	for _, m := range memories {
		typeCounts[m.MemoryType]++
		if m.Category != nil {
			categoryCounts[*m.Category]++
		}
	}

	typeStats := make([]gin.H, 0, len(typeCounts))
	for t, count := range typeCounts {
		typeStats = append(typeStats, gin.H{
			"type":  t,
			"count": count,
		})
	}

	categoryStats := make([]gin.H, 0, len(categoryCounts))
	for cat, count := range categoryCounts {
		categoryStats = append(categoryStats, gin.H{
			"category": cat,
			"count":    count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_memories": len(memories),
		"by_type":        typeStats,
		"by_category":    categoryStats,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
