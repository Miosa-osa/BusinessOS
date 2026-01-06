package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// Database represents a Notion database.
type Database struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	NotionID    string                 `json:"notion_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Cover       string                 `json:"cover,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	URL         string                 `json:"url,omitempty"`
	SyncedAt    *time.Time             `json:"synced_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Page represents a Notion page or database entry.
type Page struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	NotionID   string                 `json:"notion_id"`
	DatabaseID string                 `json:"database_id,omitempty"`
	Title      string                 `json:"title"`
	Icon       string                 `json:"icon,omitempty"`
	Cover      string                 `json:"cover,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Archived   bool                   `json:"archived"`
	SyncedAt   *time.Time             `json:"synced_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// DatabaseService handles Notion database operations.
type DatabaseService struct {
	provider *Provider
}

// NewDatabaseService creates a new database service.
func NewDatabaseService(provider *Provider) *DatabaseService {
	return &DatabaseService{provider: provider}
}

// SyncDatabasesResult represents the result of a database sync.
type SyncDatabasesResult struct {
	TotalDatabases  int `json:"total_databases"`
	SyncedDatabases int `json:"synced_databases"`
	FailedDatabases int `json:"failed_databases"`
}

// SyncDatabases syncs databases from Notion.
func (s *DatabaseService) SyncDatabases(ctx context.Context, userID string) (*SyncDatabasesResult, error) {
	log.Printf("Notion database sync starting for user %s", userID)

	result := &SyncDatabasesResult{}

	// Search for all databases
	searchBody := `{"filter":{"property":"object","value":"database"}}`
	body, err := s.provider.APIRequest(ctx, userID, "POST", "/search", strings.NewReader(searchBody))
	if err != nil {
		return nil, fmt.Errorf("failed to search databases: %w", err)
	}

	var searchResp struct {
		Results []NotionDatabase `json:"results"`
		HasMore bool             `json:"has_more"`
	}

	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	result.TotalDatabases = len(searchResp.Results)

	for _, db := range searchResp.Results {
		if err := s.saveDatabase(ctx, userID, db); err != nil {
			log.Printf("Failed to save database %s: %v", db.ID, err)
			result.FailedDatabases++
		} else {
			result.SyncedDatabases++
		}
	}

	log.Printf("Notion database sync complete for user %s: synced %d/%d databases",
		userID, result.SyncedDatabases, result.TotalDatabases)

	return result, nil
}

// NotionDatabase represents a database from the Notion API.
type NotionDatabase struct {
	Object      string `json:"object"`
	ID          string `json:"id"`
	Title       []Text `json:"title"`
	Description []Text `json:"description"`
	Icon        *Icon  `json:"icon"`
	Cover       *Cover `json:"cover"`
	Properties  map[string]interface{} `json:"properties"`
	URL         string `json:"url"`
	Archived    bool   `json:"archived"`
}

// Text represents a Notion rich text element.
type Text struct {
	Type      string `json:"type"`
	PlainText string `json:"plain_text"`
}

// Icon represents a Notion icon.
type Icon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji,omitempty"`
	File  *struct {
		URL string `json:"url"`
	} `json:"file,omitempty"`
	External *struct {
		URL string `json:"url"`
	} `json:"external,omitempty"`
}

// Cover represents a Notion cover image.
type Cover struct {
	Type string `json:"type"`
	File *struct {
		URL string `json:"url"`
	} `json:"file,omitempty"`
	External *struct {
		URL string `json:"url"`
	} `json:"external,omitempty"`
}

// saveDatabase saves a Notion database to our database.
func (s *DatabaseService) saveDatabase(ctx context.Context, userID string, db NotionDatabase) error {
	title := ""
	if len(db.Title) > 0 {
		title = db.Title[0].PlainText
	}

	description := ""
	if len(db.Description) > 0 {
		description = db.Description[0].PlainText
	}

	icon := ""
	if db.Icon != nil {
		if db.Icon.Type == "emoji" {
			icon = db.Icon.Emoji
		} else if db.Icon.File != nil {
			icon = db.Icon.File.URL
		} else if db.Icon.External != nil {
			icon = db.Icon.External.URL
		}
	}

	cover := ""
	if db.Cover != nil {
		if db.Cover.File != nil {
			cover = db.Cover.File.URL
		} else if db.Cover.External != nil {
			cover = db.Cover.External.URL
		}
	}

	propertiesJSON, _ := json.Marshal(db.Properties)

	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO notion_databases (
			user_id, notion_id, title, description, icon, cover,
			properties, url, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (user_id, notion_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon,
			cover = EXCLUDED.cover,
			properties = EXCLUDED.properties,
			url = EXCLUDED.url,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, db.ID, title, description, icon, cover, propertiesJSON, db.URL)

	return err
}

// GetDatabases retrieves databases for a user.
func (s *DatabaseService) GetDatabases(ctx context.Context, userID string) ([]*Database, error) {
	rows, err := s.provider.Pool().Query(ctx, `
		SELECT id, user_id, notion_id, title, description, icon, cover,
			properties, url, synced_at, created_at, updated_at
		FROM notion_databases
		WHERE user_id = $1
		ORDER BY title
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []*Database
	for rows.Next() {
		var db Database
		var propertiesJSON []byte

		err := rows.Scan(
			&db.ID, &db.UserID, &db.NotionID, &db.Title, &db.Description,
			&db.Icon, &db.Cover, &propertiesJSON, &db.URL, &db.SyncedAt,
			&db.CreatedAt, &db.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(propertiesJSON) > 0 {
			json.Unmarshal(propertiesJSON, &db.Properties)
		}

		databases = append(databases, &db)
	}

	return databases, nil
}

// SyncPagesResult represents the result of a page sync.
type SyncPagesResult struct {
	TotalPages  int `json:"total_pages"`
	SyncedPages int `json:"synced_pages"`
	FailedPages int `json:"failed_pages"`
}

// SyncPages syncs pages from a Notion database.
func (s *DatabaseService) SyncPages(ctx context.Context, userID, databaseID string) (*SyncPagesResult, error) {
	log.Printf("Notion page sync starting for user %s, database %s", userID, databaseID)

	// Get the Notion database ID
	var notionDatabaseID string
	err := s.provider.Pool().QueryRow(ctx, `
		SELECT notion_id FROM notion_databases WHERE id = $1 AND user_id = $2
	`, databaseID, userID).Scan(&notionDatabaseID)
	if err != nil {
		return nil, fmt.Errorf("database not found: %w", err)
	}

	result := &SyncPagesResult{}

	// Query database pages
	body, err := s.provider.APIRequest(ctx, userID, "POST",
		fmt.Sprintf("/databases/%s/query", notionDatabaseID),
		strings.NewReader("{}"))
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	var queryResp struct {
		Results []NotionPage `json:"results"`
		HasMore bool         `json:"has_more"`
	}

	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	result.TotalPages = len(queryResp.Results)

	for _, page := range queryResp.Results {
		if err := s.savePage(ctx, userID, databaseID, page); err != nil {
			log.Printf("Failed to save page %s: %v", page.ID, err)
			result.FailedPages++
		} else {
			result.SyncedPages++
		}
	}

	log.Printf("Notion page sync complete: synced %d/%d pages",
		result.SyncedPages, result.TotalPages)

	return result, nil
}

// NotionPage represents a page from the Notion API.
type NotionPage struct {
	Object     string                 `json:"object"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Icon       *Icon                  `json:"icon"`
	Cover      *Cover                 `json:"cover"`
	URL        string                 `json:"url"`
	Archived   bool                   `json:"archived"`
}

// savePage saves a Notion page to our database.
func (s *DatabaseService) savePage(ctx context.Context, userID, databaseID string, page NotionPage) error {
	// Extract title from properties
	title := extractTitle(page.Properties)

	icon := ""
	if page.Icon != nil {
		if page.Icon.Type == "emoji" {
			icon = page.Icon.Emoji
		} else if page.Icon.File != nil {
			icon = page.Icon.File.URL
		} else if page.Icon.External != nil {
			icon = page.Icon.External.URL
		}
	}

	cover := ""
	if page.Cover != nil {
		if page.Cover.File != nil {
			cover = page.Cover.File.URL
		} else if page.Cover.External != nil {
			cover = page.Cover.External.URL
		}
	}

	propertiesJSON, _ := json.Marshal(page.Properties)

	_, err := s.provider.Pool().Exec(ctx, `
		INSERT INTO notion_pages (
			user_id, notion_id, database_id, title, icon, cover,
			properties, url, archived, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (user_id, notion_id) DO UPDATE SET
			database_id = EXCLUDED.database_id,
			title = EXCLUDED.title,
			icon = EXCLUDED.icon,
			cover = EXCLUDED.cover,
			properties = EXCLUDED.properties,
			url = EXCLUDED.url,
			archived = EXCLUDED.archived,
			synced_at = NOW(),
			updated_at = NOW()
	`, userID, page.ID, databaseID, title, icon, cover, propertiesJSON, page.URL, page.Archived)

	return err
}

// extractTitle extracts the title from page properties.
func extractTitle(properties map[string]interface{}) string {
	// Look for a "title" or "Name" property
	for _, key := range []string{"title", "Title", "name", "Name"} {
		if prop, ok := properties[key]; ok {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if titleArr, ok := propMap["title"].([]interface{}); ok && len(titleArr) > 0 {
					if textObj, ok := titleArr[0].(map[string]interface{}); ok {
						if plainText, ok := textObj["plain_text"].(string); ok {
							return plainText
						}
					}
				}
			}
		}
	}
	return "Untitled"
}

// GetPages retrieves pages for a database.
func (s *DatabaseService) GetPages(ctx context.Context, userID, databaseID string, limit, offset int) ([]*Page, error) {
	rows, err := s.provider.Pool().Query(ctx, `
		SELECT id, user_id, notion_id, database_id, title, icon, cover,
			properties, url, archived, synced_at, created_at, updated_at
		FROM notion_pages
		WHERE user_id = $1 AND database_id = $2 AND archived = false
		ORDER BY title
		LIMIT $3 OFFSET $4
	`, userID, databaseID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*Page
	for rows.Next() {
		var p Page
		var propertiesJSON []byte

		err := rows.Scan(
			&p.ID, &p.UserID, &p.NotionID, &p.DatabaseID, &p.Title,
			&p.Icon, &p.Cover, &propertiesJSON, &p.URL, &p.Archived,
			&p.SyncedAt, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(propertiesJSON) > 0 {
			json.Unmarshal(propertiesJSON, &p.Properties)
		}

		pages = append(pages, &p)
	}

	return pages, nil
}

// ============================================
// MCP-Compatible Methods (Notion API Direct)
// These methods make direct API calls for MCP tools.
// ============================================

// GetToken returns the user's Notion token (delegates to provider).
func (s *DatabaseService) GetToken(ctx context.Context, userID string) (interface{}, error) {
	return s.provider.GetToken(ctx, userID)
}

// NotionDatabaseAPI represents a database from the Notion API (for MCP).
type NotionDatabaseAPI struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	Title          []Text                 `json:"title"`
	URL            string                 `json:"url"`
	Properties     map[string]interface{} `json:"properties"`
	CreatedTime    string                 `json:"created_time"`
	LastEditedTime string                 `json:"last_edited_time"`
}

// NotionPageAPI represents a page from the Notion API (for MCP).
type NotionPageAPI struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	URL            string                 `json:"url"`
	Properties     map[string]interface{} `json:"properties"`
	Parent         interface{}            `json:"parent"`
	Archived       bool                   `json:"archived"`
	CreatedTime    string                 `json:"created_time"`
	LastEditedTime string                 `json:"last_edited_time"`
}

// NotionQueryResponse represents a query response from Notion API.
type NotionQueryResponse struct {
	Results    []NotionPageAPI `json:"results"`
	HasMore    bool            `json:"has_more"`
	NextCursor *string         `json:"next_cursor"`
}

// NotionSearchResponse represents a search response from Notion API.
type NotionSearchResponse struct {
	Results    []interface{} `json:"results"`
	HasMore    bool          `json:"has_more"`
	NextCursor *string       `json:"next_cursor"`
}

// ListDatabases returns databases from Notion API.
func (s *DatabaseService) ListDatabases(ctx context.Context, userID string, pageSize int, cursor string) ([]NotionDatabaseAPI, string, bool, error) {
	requestBody := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "object",
			"value":    "database",
		},
		"page_size": pageSize,
	}
	if cursor != "" {
		requestBody["start_cursor"] = cursor
	}

	bodyJSON, _ := json.Marshal(requestBody)
	body, err := s.provider.APIRequest(ctx, userID, "POST", "/search", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, "", false, err
	}

	var resp struct {
		Results    []NotionDatabaseAPI `json:"results"`
		HasMore    bool                `json:"has_more"`
		NextCursor *string             `json:"next_cursor"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", false, err
	}

	nextCursor := ""
	if resp.NextCursor != nil {
		nextCursor = *resp.NextCursor
	}

	return resp.Results, nextCursor, resp.HasMore, nil
}

// GetDatabase returns a specific database from Notion API.
func (s *DatabaseService) GetDatabase(ctx context.Context, userID, databaseID string) (*NotionDatabaseAPI, error) {
	body, err := s.provider.APIRequest(ctx, userID, "GET", fmt.Sprintf("/databases/%s", databaseID), nil)
	if err != nil {
		return nil, err
	}

	var db NotionDatabaseAPI
	if err := json.Unmarshal(body, &db); err != nil {
		return nil, err
	}

	return &db, nil
}

// QueryDatabase queries a Notion database.
func (s *DatabaseService) QueryDatabase(ctx context.Context, userID, databaseID string, pageSize int, cursor string, filter, sorts interface{}) (*NotionQueryResponse, error) {
	requestBody := map[string]interface{}{
		"page_size": pageSize,
	}
	if cursor != "" {
		requestBody["start_cursor"] = cursor
	}
	if filter != nil {
		requestBody["filter"] = filter
	}
	if sorts != nil {
		requestBody["sorts"] = sorts
	}

	bodyJSON, _ := json.Marshal(requestBody)
	body, err := s.provider.APIRequest(ctx, userID, "POST", fmt.Sprintf("/databases/%s/query", databaseID), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	var resp NotionQueryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPage returns a specific page from Notion API.
func (s *DatabaseService) GetPage(ctx context.Context, userID, pageID string) (*NotionPageAPI, error) {
	body, err := s.provider.APIRequest(ctx, userID, "GET", fmt.Sprintf("/pages/%s", pageID), nil)
	if err != nil {
		return nil, err
	}

	var page NotionPageAPI
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// CreatePage creates a new page in a Notion database.
func (s *DatabaseService) CreatePage(ctx context.Context, userID, databaseID string, properties map[string]interface{}) (*NotionPageAPI, error) {
	requestBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": databaseID,
		},
		"properties": properties,
	}

	bodyJSON, _ := json.Marshal(requestBody)
	body, err := s.provider.APIRequest(ctx, userID, "POST", "/pages", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	var page NotionPageAPI
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// UpdatePage updates an existing Notion page.
func (s *DatabaseService) UpdatePage(ctx context.Context, userID, pageID string, properties map[string]interface{}) (*NotionPageAPI, error) {
	requestBody := map[string]interface{}{
		"properties": properties,
	}

	bodyJSON, _ := json.Marshal(requestBody)
	body, err := s.provider.APIRequest(ctx, userID, "PATCH", fmt.Sprintf("/pages/%s", pageID), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	var page NotionPageAPI
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// Search searches Notion for pages and databases.
func (s *DatabaseService) Search(ctx context.Context, userID, query, filterType string, pageSize int, cursor string) (*NotionSearchResponse, error) {
	requestBody := map[string]interface{}{
		"page_size": pageSize,
	}
	if query != "" {
		requestBody["query"] = query
	}
	if filterType != "" {
		requestBody["filter"] = map[string]interface{}{
			"property": "object",
			"value":    filterType,
		}
	}
	if cursor != "" {
		requestBody["start_cursor"] = cursor
	}

	bodyJSON, _ := json.Marshal(requestBody)
	body, err := s.provider.APIRequest(ctx, userID, "POST", "/search", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	var resp NotionSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetDatabaseTitle extracts the title from a NotionDatabaseAPI.
func GetDatabaseTitle(db *NotionDatabaseAPI) string {
	if len(db.Title) > 0 {
		return db.Title[0].PlainText
	}
	return "Untitled"
}

// GetPageTitle extracts the title from a NotionPageAPI.
func GetPageTitle(page *NotionPageAPI) string {
	// Look for a "title" or "Name" property
	for _, key := range []string{"title", "Title", "name", "Name"} {
		if prop, ok := page.Properties[key]; ok {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if titleArr, ok := propMap["title"].([]interface{}); ok && len(titleArr) > 0 {
					if textObj, ok := titleArr[0].(map[string]interface{}); ok {
						if plainText, ok := textObj["plain_text"].(string); ok {
							return plainText
						}
					}
				}
			}
		}
	}
	return "Untitled"
}
