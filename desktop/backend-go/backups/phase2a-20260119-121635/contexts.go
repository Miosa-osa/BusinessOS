package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ListContexts returns all contexts for the current user
func (h *Handlers) ListContexts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional filters
	contextType := c.Query("type")
	isArchived := c.Query("is_archived") == "true"
	isTemplate := c.Query("is_template") == "true"
	search := c.Query("search")

	var ctxType sqlc.Contexttype
	if contextType != "" {
		ctxType = stringToContextType(contextType)
	}

	contexts, err := queries.ListContexts(c.Request.Context(), sqlc.ListContextsParams{
		UserID:      user.ID,
		IsArchived:  &isArchived,
		ContextType: sqlc.NullContexttype{Contexttype: ctxType, Valid: contextType != ""},
		IsTemplate:  &isTemplate,
		Search:      stringPtr(search),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list contexts"})
		return
	}

	c.JSON(http.StatusOK, TransformContexts(contexts))
}

// CreateContext creates a new context
func (h *Handlers) CreateContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name                 string          `json:"name" binding:"required"`
		Type                 *string         `json:"type"`
		Content              *string         `json:"content"`
		StructuredData       json.RawMessage `json:"structured_data"`
		SystemPromptTemplate *string         `json:"system_prompt_template"`
		Blocks               json.RawMessage `json:"blocks"`
		CoverImage           *string         `json:"cover_image"`
		Icon                 *string         `json:"icon"`
		ParentID             *string         `json:"parent_id"`
		IsTemplate           *bool           `json:"is_template"`
		PropertySchema       json.RawMessage `json:"property_schema"`
		Properties           json.RawMessage `json:"properties"`
		ClientID             *string         `json:"client_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional IDs
	var parentID, clientID pgtype.UUID
	if req.ParentID != nil {
		if parsed, err := uuid.Parse(*req.ParentID); err == nil {
			parentID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.ClientID != nil {
		if parsed, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Parse context type
	var ctxType sqlc.NullContexttype
	if req.Type != nil {
		ctxType = sqlc.NullContexttype{
			Contexttype: stringToContextType(*req.Type),
			Valid:       true,
		}
	}

	// Handle JSON fields
	structuredData := []byte("{}")
	if req.StructuredData != nil {
		structuredData = req.StructuredData
	}
	blocks := []byte("[]")
	if req.Blocks != nil {
		blocks = req.Blocks
	}
	propertySchema := []byte("{}")
	if req.PropertySchema != nil {
		propertySchema = req.PropertySchema
	}
	properties := []byte("{}")
	if req.Properties != nil {
		properties = req.Properties
	}

	// Default type to 'document' if not specified (lowercase for DB compatibility)
	if !ctxType.Valid {
		ctxType = sqlc.NullContexttype{
			Contexttype: sqlc.ContexttypeDocument, // lowercase 'document' for DB enum
			Valid:       true,
		}
	}

	// Default is_template to false if not specified
	isTemplate := false
	if req.IsTemplate != nil {
		isTemplate = *req.IsTemplate
	}

	context, err := queries.CreateContext(c.Request.Context(), sqlc.CreateContextParams{
		UserID:               user.ID,
		Name:                 req.Name,
		Type:                 ctxType,
		Content:              req.Content,
		StructuredData:       structuredData,
		SystemPromptTemplate: req.SystemPromptTemplate,
		Blocks:               blocks,
		CoverImage:           req.CoverImage,
		Icon:                 req.Icon,
		ParentID:             parentID,
		IsTemplate:           &isTemplate,
		PropertySchema:       propertySchema,
		Properties:           properties,
		ClientID:             clientID,
	})
	if err != nil {
		log.Printf("CreateContext error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create context"})
		return
	}

	c.JSON(http.StatusCreated, TransformContext(context))
}

// GetContext returns a single context
func (h *Handlers) GetContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)
	context, err := queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	// Check if children are requested
	if c.Query("include_children") == "true" {
		children, err := queries.GetContextChildren(c.Request.Context(), sqlc.GetContextChildrenParams{
			ParentID: pgtype.UUID{Bytes: id, Valid: true},
			UserID:   user.ID,
		})
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"context":  TransformContext(context),
				"children": TransformContexts(children),
			})
			return
		}
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// UpdateContext updates an existing context
func (h *Handlers) UpdateContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	var req struct {
		Name                 *string         `json:"name"`
		Type                 *string         `json:"type"`
		Content              *string         `json:"content"`
		StructuredData       json.RawMessage `json:"structured_data"`
		SystemPromptTemplate *string         `json:"system_prompt_template"`
		CoverImage           *string         `json:"cover_image"`
		Icon                 *string         `json:"icon"`
		ParentID             *string         `json:"parent_id"`
		IsTemplate           *bool           `json:"is_template"`
		PropertySchema       json.RawMessage `json:"property_schema"`
		Properties           json.RawMessage `json:"properties"`
		ClientID             *string         `json:"client_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing context first
	existing, err := queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	// Build update params with existing values as defaults
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	var ctxType sqlc.NullContexttype
	if req.Type != nil {
		ctxType = sqlc.NullContexttype{
			Contexttype: stringToContextType(*req.Type),
			Valid:       true,
		}
	} else {
		ctxType = existing.Type
	}

	content := existing.Content
	if req.Content != nil {
		content = req.Content
	}

	structuredData := existing.StructuredData
	if req.StructuredData != nil {
		structuredData = req.StructuredData
	}

	systemPromptTemplate := existing.SystemPromptTemplate
	if req.SystemPromptTemplate != nil {
		systemPromptTemplate = req.SystemPromptTemplate
	}

	coverImage := existing.CoverImage
	if req.CoverImage != nil {
		coverImage = req.CoverImage
	}

	icon := existing.Icon
	if req.Icon != nil {
		icon = req.Icon
	}

	parentID := existing.ParentID
	if req.ParentID != nil {
		if parsed, err := uuid.Parse(*req.ParentID); err == nil {
			parentID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	isTemplate := existing.IsTemplate
	if req.IsTemplate != nil {
		isTemplate = req.IsTemplate
	}

	propertySchema := existing.PropertySchema
	if req.PropertySchema != nil {
		propertySchema = req.PropertySchema
	}

	properties := existing.Properties
	if req.Properties != nil {
		properties = req.Properties
	}

	clientID := existing.ClientID
	if req.ClientID != nil {
		if parsed, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	context, err := queries.UpdateContext(c.Request.Context(), sqlc.UpdateContextParams{
		ID:                   pgtype.UUID{Bytes: id, Valid: true},
		Name:                 name,
		Type:                 ctxType,
		Content:              content,
		StructuredData:       structuredData,
		SystemPromptTemplate: systemPromptTemplate,
		CoverImage:           coverImage,
		Icon:                 icon,
		ParentID:             parentID,
		IsTemplate:           isTemplate,
		PropertySchema:       propertySchema,
		Properties:           properties,
		ClientID:             clientID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update context"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// UpdateContextBlocks updates only the blocks field of a context
func (h *Handlers) UpdateContextBlocks(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	var req struct {
		Blocks    json.RawMessage `json:"blocks" binding:"required"`
		WordCount *int32          `json:"word_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	context, err := queries.UpdateContextBlocks(c.Request.Context(), sqlc.UpdateContextBlocksParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Blocks:    req.Blocks,
		WordCount: req.WordCount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blocks"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// ShareContext makes a context publicly accessible
func (h *Handlers) ShareContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	// Generate unique share ID
	shareID := generateShareID()

	context, err := queries.ShareContext(c.Request.Context(), sqlc.ShareContextParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		ShareID: &shareID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share context"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// UnshareContext makes a context private
func (h *Handlers) UnshareContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	context, err := queries.UnshareContext(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare context"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// GetPublicContext returns a publicly shared context by share ID
func (h *Handlers) GetPublicContext(c *gin.Context) {
	shareID := c.Param("share_id")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share ID required"})
		return
	}

	queries := sqlc.New(h.pool)
	context, err := queries.GetPublicContext(c.Request.Context(), &shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found or not public"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// DuplicateContext creates a copy of an existing context
func (h *Handlers) DuplicateContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get original context
	original, err := queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	// Create duplicate
	newName := original.Name + " (Copy)"
	duplicate, err := queries.CreateContext(c.Request.Context(), sqlc.CreateContextParams{
		UserID:               user.ID,
		Name:                 newName,
		Type:                 original.Type,
		Content:              original.Content,
		StructuredData:       original.StructuredData,
		SystemPromptTemplate: original.SystemPromptTemplate,
		Blocks:               original.Blocks,
		CoverImage:           original.CoverImage,
		Icon:                 original.Icon,
		ParentID:             original.ParentID,
		IsTemplate:           original.IsTemplate,
		PropertySchema:       original.PropertySchema,
		Properties:           original.Properties,
		ClientID:             original.ClientID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate context"})
		return
	}

	c.JSON(http.StatusCreated, TransformContext(duplicate))
}

// ArchiveContext archives a context
func (h *Handlers) ArchiveContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	context, err := queries.ArchiveContext(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive context"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// UnarchiveContext unarchives a context
func (h *Handlers) UnarchiveContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	context, err := queries.UnarchiveContext(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive context"})
		return
	}

	c.JSON(http.StatusOK, TransformContext(context))
}

// DeleteContext deletes a context
func (h *Handlers) DeleteContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteContext(c.Request.Context(), sqlc.DeleteContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Context deleted"})
}

// AggregateContext provides aggregated data for a context
func (h *Handlers) AggregateContext(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get the context
	context, err := queries.GetContext(c.Request.Context(), sqlc.GetContextParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context not found"})
		return
	}

	// Get children
	children, _ := queries.GetContextChildren(c.Request.Context(), sqlc.GetContextChildrenParams{
		ParentID: pgtype.UUID{Bytes: id, Valid: true},
		UserID:   user.ID,
	})

	// Get related artifacts
	artifacts, _ := queries.ListArtifacts(c.Request.Context(), sqlc.ListArtifactsParams{
		UserID:    user.ID,
		ContextID: pgtype.UUID{Bytes: id, Valid: true},
	})

	c.JSON(http.StatusOK, gin.H{
		"context":   TransformContext(context),
		"children":  TransformContexts(children),
		"artifacts": TransformArtifacts(artifacts),
	})
}

// stringToContextType converts a string to sqlc.Contexttype
func stringToContextType(t string) sqlc.Contexttype {
	typeMap := map[string]sqlc.Contexttype{
		"person":   sqlc.ContexttypePERSON,
		"business": sqlc.ContexttypeBUSINESS,
		"project":  sqlc.ContexttypePROJECT,
		"document": sqlc.ContexttypeDocument, // Use lowercase version for DB compatibility
		"custom":   sqlc.ContexttypeCUSTOM,
	}
	if enum, ok := typeMap[strings.ToLower(t)]; ok {
		return enum
	}
	return sqlc.ContexttypeCUSTOM
}

// generateShareID generates a random share ID using internal/utils
func generateShareID() string {
	return utils.MustGenerateRandomHex(8) // 8 bytes = 16 hex chars
}

// stringPtr returns a pointer to a string, or nil if the string is empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
