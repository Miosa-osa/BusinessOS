package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

const (
	notionAPIBaseURL   = "https://api.notion.com/v1"
	notionOAuthBaseURL = "https://api.notion.com/v1/oauth"
	notionAPIVersion   = "2022-06-28"
)

// NotionService handles Notion API operations
type NotionService struct {
	pool         *pgxpool.Pool
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// NotionOAuthResponse represents the response from Notion OAuth token exchange
type NotionOAuthResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceIcon string `json:"workspace_icon"`
	Owner         struct {
		Type string `json:"type"`
		User struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
			Type      string `json:"type"`
			Person    struct {
				Email string `json:"email"`
			} `json:"person"`
		} `json:"user"`
	} `json:"owner"`
	DuplicatedTemplateID string `json:"duplicated_template_id,omitempty"`
}

// NotionDatabase represents a Notion database
type NotionDatabase struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	Title          []NotionRichText       `json:"title"`
	Description    []NotionRichText       `json:"description"`
	Icon           *NotionIcon            `json:"icon"`
	Cover          *NotionFile            `json:"cover"`
	Properties     map[string]interface{} `json:"properties"`
	URL            string                 `json:"url"`
	CreatedTime    string                 `json:"created_time"`
	LastEditedTime string                 `json:"last_edited_time"`
	Archived       bool                   `json:"archived"`
}

// NotionPage represents a Notion page
type NotionPage struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	CreatedTime    string                 `json:"created_time"`
	LastEditedTime string                 `json:"last_edited_time"`
	CreatedBy      NotionUser             `json:"created_by"`
	LastEditedBy   NotionUser             `json:"last_edited_by"`
	Cover          *NotionFile            `json:"cover"`
	Icon           *NotionIcon            `json:"icon"`
	Parent         NotionParent           `json:"parent"`
	Archived       bool                   `json:"archived"`
	Properties     map[string]interface{} `json:"properties"`
	URL            string                 `json:"url"`
}

// NotionRichText represents rich text in Notion
type NotionRichText struct {
	Type        string `json:"type"`
	PlainText   string `json:"plain_text"`
	Annotations struct {
		Bold          bool   `json:"bold"`
		Italic        bool   `json:"italic"`
		Strikethrough bool   `json:"strikethrough"`
		Underline     bool   `json:"underline"`
		Code          bool   `json:"code"`
		Color         string `json:"color"`
	} `json:"annotations"`
	Href *string `json:"href"`
}

// NotionIcon represents an icon (emoji or file)
type NotionIcon struct {
	Type     string      `json:"type"`
	Emoji    string      `json:"emoji,omitempty"`
	External *NotionFile `json:"external,omitempty"`
}

// NotionFile represents a file or external URL
type NotionFile struct {
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	External struct {
		URL string `json:"url"`
	} `json:"external,omitempty"`
}

// NotionUser represents a Notion user
type NotionUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// NotionParent represents a page's parent
type NotionParent struct {
	Type       string `json:"type"`
	DatabaseID string `json:"database_id,omitempty"`
	PageID     string `json:"page_id,omitempty"`
	Workspace  bool   `json:"workspace,omitempty"`
}

// NotionSearchResponse represents the response from Notion search API
type NotionSearchResponse struct {
	Object     string        `json:"object"`
	Results    []interface{} `json:"results"`
	NextCursor *string       `json:"next_cursor"`
	HasMore    bool          `json:"has_more"`
}

// NotionDatabaseQueryResponse represents a database query response
type NotionDatabaseQueryResponse struct {
	Object     string       `json:"object"`
	Results    []NotionPage `json:"results"`
	NextCursor *string      `json:"next_cursor"`
	HasMore    bool         `json:"has_more"`
}

func NewNotionService(pool *pgxpool.Pool) *NotionService {
	cfg := config.AppConfig
	return &NotionService{
		pool:         pool,
		clientID:     cfg.NotionClientID,
		clientSecret: cfg.NotionClientSecret,
		redirectURI:  cfg.NotionRedirectURI,
		httpClient:   &http.Client{},
	}
}

// GetAuthURL returns the Notion OAuth URL for user authorization
func (s *NotionService) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("response_type", "code")
	params.Set("owner", "user")
	params.Set("redirect_uri", s.redirectURI)
	params.Set("state", state)

	return "https://api.notion.com/v1/oauth/authorize?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for an access token
func (s *NotionService) ExchangeCode(ctx context.Context, code string) (*NotionOAuthResponse, error) {
	tokenURL := notionOAuthBaseURL + "/token"

	// Prepare request body
	body := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": s.redirectURI,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Notion requires Basic auth with client_id:client_secret
	auth := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("notion OAuth error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var oauthResponse NotionOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &oauthResponse, nil
}

// SaveToken saves OAuth token to the database
func (s *NotionService) SaveToken(ctx context.Context, userID string, response *NotionOAuthResponse) error {
	queries := sqlc.New(s.pool)

	var ownerType, ownerUserID, ownerUserName, ownerUserEmail *string
	if response.Owner.Type != "" {
		ownerType = &response.Owner.Type
	}
	if response.Owner.User.ID != "" {
		ownerUserID = &response.Owner.User.ID
	}
	if response.Owner.User.Name != "" {
		ownerUserName = &response.Owner.User.Name
	}
	if response.Owner.User.Person.Email != "" {
		ownerUserEmail = &response.Owner.User.Person.Email
	}

	var workspaceIcon *string
	if response.WorkspaceIcon != "" {
		workspaceIcon = &response.WorkspaceIcon
	}

	_, err := queries.CreateNotionOAuthToken(ctx, sqlc.CreateNotionOAuthTokenParams{
		UserID:         userID,
		WorkspaceID:    response.WorkspaceID,
		WorkspaceName:  &response.WorkspaceName,
		WorkspaceIcon:  workspaceIcon,
		AccessToken:    response.AccessToken,
		BotID:          &response.BotID,
		OwnerType:      ownerType,
		OwnerUserID:    ownerUserID,
		OwnerUserName:  ownerUserName,
		OwnerUserEmail: ownerUserEmail,
	})

	return err
}

// UpdateToken updates existing OAuth token
func (s *NotionService) UpdateToken(ctx context.Context, userID string, response *NotionOAuthResponse) error {
	queries := sqlc.New(s.pool)

	var workspaceIcon *string
	if response.WorkspaceIcon != "" {
		workspaceIcon = &response.WorkspaceIcon
	}

	_, err := queries.UpdateNotionOAuthToken(ctx, sqlc.UpdateNotionOAuthTokenParams{
		UserID:        userID,
		AccessToken:   response.AccessToken,
		WorkspaceName: &response.WorkspaceName,
		WorkspaceIcon: workspaceIcon,
	})

	return err
}

// GetToken retrieves OAuth token from the database
func (s *NotionService) GetToken(ctx context.Context, userID string) (*sqlc.NotionOauthToken, error) {
	queries := sqlc.New(s.pool)

	token, err := queries.GetNotionOAuthToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// DeleteToken removes OAuth token for a user
func (s *NotionService) DeleteToken(ctx context.Context, userID string) error {
	queries := sqlc.New(s.pool)
	return queries.DeleteNotionOAuthToken(ctx, userID)
}

// GetConnectionStatus checks if a user has connected their Notion workspace
func (s *NotionService) GetConnectionStatus(ctx context.Context, userID string) (*sqlc.GetNotionOAuthStatusRow, error) {
	queries := sqlc.New(s.pool)
	status, err := queries.GetNotionOAuthStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// ========== Notion API Operations ==========

// doRequest performs an authenticated request to the Notion API
func (s *NotionService) doRequest(ctx context.Context, method, endpoint string, body interface{}, accessToken string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyJSON)
	}

	req, err := http.NewRequestWithContext(ctx, method, notionAPIBaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("notion API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Search searches for databases and pages in the workspace
func (s *NotionService) Search(ctx context.Context, userID string, query string, filter string, pageSize int, startCursor string) (*NotionSearchResponse, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body := map[string]interface{}{
		"page_size": pageSize,
	}

	if query != "" {
		body["query"] = query
	}

	if filter == "database" || filter == "page" {
		body["filter"] = map[string]string{
			"value":    filter,
			"property": "object",
		}
	}

	if startCursor != "" {
		body["start_cursor"] = startCursor
	}

	respBody, err := s.doRequest(ctx, "POST", "/search", body, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var searchResponse NotionSearchResponse
	if err := json.Unmarshal(respBody, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResponse, nil
}

// ListDatabases lists all databases the integration has access to
func (s *NotionService) ListDatabases(ctx context.Context, userID string, pageSize int, startCursor string) ([]NotionDatabase, string, bool, error) {
	searchResp, err := s.Search(ctx, userID, "", "database", pageSize, startCursor)
	if err != nil {
		return nil, "", false, err
	}

	var databases []NotionDatabase
	for _, result := range searchResp.Results {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			continue
		}
		var db NotionDatabase
		if err := json.Unmarshal(resultJSON, &db); err == nil && db.Object == "database" {
			databases = append(databases, db)
		}
	}

	nextCursor := ""
	if searchResp.NextCursor != nil {
		nextCursor = *searchResp.NextCursor
	}

	return databases, nextCursor, searchResp.HasMore, nil
}

// GetDatabase retrieves a specific database by ID
func (s *NotionService) GetDatabase(ctx context.Context, userID string, databaseID string) (*NotionDatabase, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	respBody, err := s.doRequest(ctx, "GET", "/databases/"+databaseID, nil, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var database NotionDatabase
	if err := json.Unmarshal(respBody, &database); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &database, nil
}

// QueryDatabase queries a database for pages
func (s *NotionService) QueryDatabase(ctx context.Context, userID string, databaseID string, pageSize int, startCursor string, filter interface{}, sorts interface{}) (*NotionDatabaseQueryResponse, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body := map[string]interface{}{
		"page_size": pageSize,
	}

	if startCursor != "" {
		body["start_cursor"] = startCursor
	}

	if filter != nil {
		body["filter"] = filter
	}

	if sorts != nil {
		body["sorts"] = sorts
	}

	respBody, err := s.doRequest(ctx, "POST", "/databases/"+databaseID+"/query", body, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var queryResponse NotionDatabaseQueryResponse
	if err := json.Unmarshal(respBody, &queryResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &queryResponse, nil
}

// GetPage retrieves a specific page by ID
func (s *NotionService) GetPage(ctx context.Context, userID string, pageID string) (*NotionPage, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	respBody, err := s.doRequest(ctx, "GET", "/pages/"+pageID, nil, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &page, nil
}

// CreatePage creates a new page in a database
func (s *NotionService) CreatePage(ctx context.Context, userID string, databaseID string, properties map[string]interface{}) (*NotionPage, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body := map[string]interface{}{
		"parent": map[string]string{
			"database_id": databaseID,
		},
		"properties": properties,
	}

	respBody, err := s.doRequest(ctx, "POST", "/pages", body, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &page, nil
}

// UpdatePage updates an existing page's properties
func (s *NotionService) UpdatePage(ctx context.Context, userID string, pageID string, properties map[string]interface{}) (*NotionPage, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body := map[string]interface{}{
		"properties": properties,
	}

	respBody, err := s.doRequest(ctx, "PATCH", "/pages/"+pageID, body, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &page, nil
}

// GetMe retrieves information about the integration bot user
func (s *NotionService) GetMe(ctx context.Context, userID string) (map[string]interface{}, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	respBody, err := s.doRequest(ctx, "GET", "/users/me", nil, token.AccessToken)
	if err != nil {
		return nil, err
	}

	var user map[string]interface{}
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return user, nil
}

// GetDatabaseTitle extracts the title from a database
func GetDatabaseTitle(db *NotionDatabase) string {
	if len(db.Title) > 0 {
		return db.Title[0].PlainText
	}
	return "Untitled"
}

// GetPageTitle extracts the title from a page's properties
func GetPageTitle(page *NotionPage) string {
	for _, prop := range page.Properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}
		if propMap["type"] == "title" {
			if titleArr, ok := propMap["title"].([]interface{}); ok && len(titleArr) > 0 {
				if titleObj, ok := titleArr[0].(map[string]interface{}); ok {
					if plainText, ok := titleObj["plain_text"].(string); ok {
						return plainText
					}
				}
			}
		}
	}
	return "Untitled"
}
