package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

type NotionOAuthHandler struct {
	notionService *services.NotionService
}

func NewNotionOAuthHandler(notionService *services.NotionService) *NotionOAuthHandler {
	return &NotionOAuthHandler{
		notionService: notionService,
	}
}

func (h *NotionOAuthHandler) InitiateNotionAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	state := generateNotionRandomState()

	c.SetCookie("notion_oauth_state", state, 600, "/", "", false, true)
	c.SetCookie("notion_oauth_user", user.ID, 600, "/", "", false, true)

	authURL := h.notionService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleNotionCallback handles the OAuth callback from Notion
func (h *NotionOAuthHandler) HandleNotionCallback(c *gin.Context) {
	state := c.Query("state")
	storedState, err := c.Cookie("notion_oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	userID, err := c.Cookie("notion_oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	response, err := h.notionService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Printf("Notion OAuth exchange error: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error=exchange_failed")
		return
	}

	// Save tokens to database
	if err := h.notionService.SaveToken(c.Request.Context(), userID, response); err != nil {
		if err := h.notionService.UpdateToken(c.Request.Context(), userID, response); err != nil {
			log.Printf("Notion OAuth save token error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error=save_failed")
			return
		}
	}

	// Clear OAuth cookies
	c.SetCookie("notion_oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("notion_oauth_user", "", -1, "/", "", false, true)

	// Redirect to settings page with success
	c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_connected=true")
}

func (h *NotionOAuthHandler) GetNotionConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.notionService.GetConnectionStatus(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":      true,
		"workspace_id":   status.WorkspaceID,
		"workspace_name": status.WorkspaceName,
		"workspace_icon": status.WorkspaceIcon,
		"owner_name":     status.OwnerUserName,
		"connected_at":   status.CreatedAt,
	})
}


func (h *NotionOAuthHandler) DisconnectNotion(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.notionService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Notion workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notion workspace disconnected"})
}

// GetNotionDatabases returns all databases the user has access to
func (h *NotionOAuthHandler) GetNotionDatabases(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	cursor := c.Query("cursor")
	databases, nextCursor, hasMore, err := h.notionService.ListDatabases(c.Request.Context(), user.ID, 100, cursor)
	if err != nil {
		log.Printf("GetNotionDatabases error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch databases"})
		return
	}

	// Transform to simpler response format
	var result []map[string]interface{}
	for _, db := range databases {
		result = append(result, map[string]interface{}{
			"id":               db.ID,
			"title":            services.GetDatabaseTitle(&db),
			"url":              db.URL,
			"created_time":     db.CreatedTime,
			"last_edited_time": db.LastEditedTime,
			"icon":             db.Icon,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"databases":   result,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

func (h *NotionOAuthHandler) GetNotionPages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	databaseID := c.Query("database_id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	cursor := c.Query("cursor")
	queryResp, err := h.notionService.QueryDatabase(c.Request.Context(), user.ID, databaseID, 100, cursor, nil, nil)
	if err != nil {
		log.Printf("GetNotionPages error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pages"})
		return
	}

	// Transform to simpler response format
	var result []map[string]interface{}
	for _, page := range queryResp.Results {
		result = append(result, map[string]interface{}{
			"id":               page.ID,
			"title":            services.GetPageTitle(&page),
			"url":              page.URL,
			"created_time":     page.CreatedTime,
			"last_edited_time": page.LastEditedTime,
			"icon":             page.Icon,
			"archived":         page.Archived,
		})
	}

	nextCursor := ""
	if queryResp.NextCursor != nil {
		nextCursor = *queryResp.NextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"pages":       result,
		"next_cursor": nextCursor,
		"has_more":    queryResp.HasMore,
	})
}

// SearchNotion searches for pages and databases
func (h *NotionOAuthHandler) SearchNotion(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	query := c.Query("query")
	filter := c.Query("filter") // "database" or "page"
	cursor := c.Query("cursor")

	searchResp, err := h.notionService.Search(c.Request.Context(), user.ID, query, filter, 20, cursor)
	if err != nil {
		log.Printf("SearchNotion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	nextCursor := ""
	if searchResp.NextCursor != nil {
		nextCursor = *searchResp.NextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     searchResp.Results,
		"next_cursor": nextCursor,
		"has_more":    searchResp.HasMore,
	})
}

// Helper function for generating random state
func generateNotionRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
