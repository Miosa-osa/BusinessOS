package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// DocumentHandler handles document processing operations
type DocumentHandler struct {
	processor *services.DocumentProcessor
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(processor *services.DocumentProcessor) *DocumentHandler {
	return &DocumentHandler{
		processor: processor,
	}
}

// getUserIDFromContext gets user ID from auth context (set by middleware)
func getUserIDFromContext(c *gin.Context) string {
	// First try auth middleware's GetCurrentUser
	if user := middleware.GetCurrentUser(c); user != nil {
		return user.ID
	}
	// Fallback to header for backward compatibility
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	return "default-user"
}

// UploadDocument handles document upload and processing
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	// Parse multipart form (max 50MB)
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file: " + err.Error()})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return
	}

	// Get optional parameters
	projectIDStr := c.Request.FormValue("project_id")
	nodeIDStr := c.Request.FormValue("node_id")
	displayName := c.Request.FormValue("display_name")
	description := c.Request.FormValue("description")
	documentType := c.Request.FormValue("document_type")
	category := c.Request.FormValue("category")

	// Parse UUIDs if provided
	var projectID, nodeID *uuid.UUID
	if projectIDStr != "" {
		if id, err := uuid.Parse(projectIDStr); err == nil {
			projectID = &id
		}
	}
	if nodeIDStr != "" {
		if id, err := uuid.Parse(nodeIDStr); err == nil {
			nodeID = &id
		}
	}

	// Create upload request
	input := services.ProcessDocumentInput{
		UserID:           userID,
		Filename:         header.Filename,
		OriginalFilename: header.Filename,
		DisplayName:      displayName,
		Description:      description,
		MimeType:         header.Header.Get("Content-Type"),
		Content:          content,
		ProjectID:        projectID,
		NodeID:           nodeID,
		DocumentType:     documentType,
		Category:         category,
	}

	// Process the document
	doc, err := h.processor.ProcessDocument(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process document: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// GetDocument retrieves a document by ID
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, err := h.processor.GetDocument(ctx, userID, docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// ListDocuments lists documents for a user
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get optional project/node filters
	projectIDStr := c.Query("project_id")
	nodeIDStr := c.Query("node_id")

	var projectID, nodeID *uuid.UUID
	if projectIDStr != "" {
		if id, err := uuid.Parse(projectIDStr); err == nil {
			projectID = &id
		}
	}
	if nodeIDStr != "" {
		if id, err := uuid.Parse(nodeIDStr); err == nil {
			nodeID = &id
		}
	}

	// Search with empty query to list all
	docs, err := h.processor.SearchDocuments(ctx, userID, "", limit, projectID, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, docs)
}

// SearchDocuments searches documents semantically
func (h *DocumentHandler) SearchDocuments(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	var req struct {
		Query     string `json:"query"`
		Limit     int    `json:"limit"`
		ProjectID string `json:"project_id,omitempty"`
		NodeID    string `json:"node_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Parse UUIDs if provided
	var projectID, nodeID *uuid.UUID
	if req.ProjectID != "" {
		if id, err := uuid.Parse(req.ProjectID); err == nil {
			projectID = &id
		}
	}
	if req.NodeID != "" {
		if id, err := uuid.Parse(req.NodeID); err == nil {
			nodeID = &id
		}
	}

	results, err := h.processor.SearchDocuments(ctx, userID, req.Query, req.Limit, projectID, nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetRelevantChunks retrieves relevant chunks for a query
func (h *DocumentHandler) GetRelevantChunks(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}

	chunks, err := h.processor.GetRelevantChunks(ctx, userID, req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chunks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, chunks)
}

// DeleteDocument deletes a document
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	if err := h.processor.DeleteDocument(ctx, userID, docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReprocessDocument reprocesses a document (useful after embedding model changes)
func (h *DocumentHandler) ReprocessDocument(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	if err := h.processor.ReprocessDocument(ctx, userID, docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reprocess document: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "reprocessing"})
}

// GetDocumentContent retrieves the raw content of a document
func (h *DocumentHandler) GetDocumentContent(c *gin.Context) {
	ctx := c.Request.Context()
	userID := getUserIDFromContext(c)

	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	content, err := h.processor.GetDocumentContent(ctx, userID, docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get content: " + err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

// RegisterDocumentRoutes registers document routes on a Gin router group
func RegisterDocumentRoutes(r *gin.RouterGroup, handler *DocumentHandler) {
	docs := r.Group("/documents")
	{
		docs.POST("", handler.UploadDocument)
		docs.GET("", handler.ListDocuments)
		docs.POST("/search", handler.SearchDocuments)
		docs.POST("/chunks", handler.GetRelevantChunks)
		docs.GET("/:id", handler.GetDocument)
		docs.DELETE("/:id", handler.DeleteDocument)
		docs.POST("/:id/reprocess", handler.ReprocessDocument)
		docs.GET("/:id/content", handler.GetDocumentContent)
	}
}
