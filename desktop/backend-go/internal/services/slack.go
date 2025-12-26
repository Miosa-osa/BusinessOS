package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/slack-go/slack"
)

// SlackService handles Slack API operations
type SlackService struct {
	pool         *pgxpool.Pool
	clientID     string
	clientSecret string
	redirectURI  string
}

// SlackOAuthResponse represents the response from Slack OAuth
type SlackOAuthResponse struct {
	OK          bool   `json:"ok"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	AuthedUser struct {
		ID          string `json:"id"`
		Scope       string `json:"scope"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	} `json:"authed_user"`
	IncomingWebhook struct {
		Channel          string `json:"channel"`
		ChannelID        string `json:"channel_id"`
		ConfigurationURL string `json:"configuration_url"`
		URL              string `json:"url"`
	} `json:"incoming_webhook"`
}

func NewSlackService(pool *pgxpool.Pool) *SlackService {
	cfg := config.AppConfig
	return &SlackService{
		pool:         pool,
		clientID:     cfg.SlackClientID,
		clientSecret: cfg.SlackClientSecret,
		redirectURI:  cfg.SlackRedirectURI,
	}
}

// returns the Slack OAuth URL for user authorization
// Scopes define what the bot can do:
// - channels:read - View basic channel info
// - channels:history - View messages in public channels
// - chat:write - Send messages as the bot
// - users:read - View users in workspace
// - search:read - Search messages (requires user token)
func (s *SlackService) GetAuthURL(state string) string {
	// Bot scopes (what the app can do)
	botScopes := []string{
		"channels:read",
		"channels:history",
		"chat:write",
		"users:read",
		"users:read.email",
		"groups:read",
		"groups:history",
		"im:read",
		"im:history",
		"mpim:read",
		"mpim:history",
	}

	// User scopes (what the user authorizes)
	userScopes := []string{
		"search:read",
	}

	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("scope", strings.Join(botScopes, ","))
	params.Set("user_scope", strings.Join(userScopes, ","))
	params.Set("redirect_uri", s.redirectURI)
	params.Set("state", state)

	return "https://slack.com/oauth/v2/authorize?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (s *SlackService) ExchangeCode(ctx context.Context, code string) (*slack.OAuthV2Response, error) {
	response, err := slack.GetOAuthV2Response(
		nil, // use default HTTP client
		s.clientID,
		s.clientSecret,
		code,
		s.redirectURI,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return response, nil
}

// SaveToken saves OAuth tokens to the database
func (s *SlackService) SaveToken(ctx context.Context, userID string, response *slack.OAuthV2Response) error {
	queries := sqlc.New(s.pool)

	botScopes := strings.Split(response.Scope, ",")
	var userScopes []string
	if response.AuthedUser.Scope != "" {
		userScopes = strings.Split(response.AuthedUser.Scope, ",")
	}

	var webhookURL, webhookChannel *string
	if response.IncomingWebhook.URL != "" {
		webhookURL = &response.IncomingWebhook.URL
		webhookChannel = &response.IncomingWebhook.Channel
	}

	var userToken *string
	if response.AuthedUser.AccessToken != "" {
		userToken = &response.AuthedUser.AccessToken
	}

	_, err := queries.CreateSlackOAuthToken(ctx, sqlc.CreateSlackOAuthTokenParams{
		UserID:                 userID,
		WorkspaceID:            response.Team.ID,
		WorkspaceName:          &response.Team.Name,
		BotToken:               response.AccessToken,
		UserToken:              userToken,
		BotUserID:              &response.BotUserID,
		AuthedUserID:           &response.AuthedUser.ID,
		BotScopes:              botScopes,
		UserScopes:             userScopes,
		IncomingWebhookUrl:     webhookURL,
		IncomingWebhookChannel: webhookChannel,
	})

	return err
}

// UpdateToken updates existing OAuth tokens
func (s *SlackService) UpdateToken(ctx context.Context, userID string, response *slack.OAuthV2Response) error {
	queries := sqlc.New(s.pool)

	botScopes := strings.Split(response.Scope, ",")
	var userScopes []string
	if response.AuthedUser.Scope != "" {
		userScopes = strings.Split(response.AuthedUser.Scope, ",")
	}

	var userToken *string
	if response.AuthedUser.AccessToken != "" {
		userToken = &response.AuthedUser.AccessToken
	}

	_, err := queries.UpdateSlackOAuthToken(ctx, sqlc.UpdateSlackOAuthTokenParams{
		UserID:     userID,
		BotToken:   response.AccessToken,
		UserToken:  userToken,
		BotScopes:  botScopes,
		UserScopes: userScopes,
	})

	return err
}

// GetToken retrieves OAuth tokens from the database
func (s *SlackService) GetToken(ctx context.Context, userID string) (*sqlc.SlackOauthToken, error) {
	queries := sqlc.New(s.pool)

	token, err := queries.GetSlackOAuthToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// GetSlackClient creates a Slack API client for a user
func (s *SlackService) GetSlackClient(ctx context.Context, userID string) (*slack.Client, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	client := slack.New(token.BotToken)
	return client, nil
}

// GetUserSlackClient creates a Slack API client using the user token (for search)
func (s *SlackService) GetUserSlackClient(ctx context.Context, userID string) (*slack.Client, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if token.UserToken == nil || *token.UserToken == "" {
		return nil, fmt.Errorf("user token not available")
	}

	client := slack.New(*token.UserToken)
	return client, nil
}

// DeleteToken removes OAuth tokens for a user
func (s *SlackService) DeleteToken(ctx context.Context, userID string) error {
	queries := sqlc.New(s.pool)
	return queries.DeleteSlackOAuthToken(ctx, userID)
}

// GetConnectionStatus checks if a user has connected their Slack workspace
func (s *SlackService) GetConnectionStatus(ctx context.Context, userID string) (*sqlc.GetSlackOAuthStatusRow, error) {
	queries := sqlc.New(s.pool)
	status, err := queries.GetSlackOAuthStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// ========== Slack API Operations ==========

// ListChannels returns channels the bot has access to
func (s *SlackService) ListChannels(ctx context.Context, userID string, limit int) ([]slack.Channel, error) {
	client, err := s.GetSlackClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	params := &slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		ExcludeArchived: true,
		Limit:           limit,
	}

	channels, _, err := client.GetConversations(params)
	if err != nil {
		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	return channels, nil
}

// SendMessage sends a message to a channel
func (s *SlackService) SendMessage(ctx context.Context, userID, channel, text string, threadTS string) (string, error) {
	client, err := s.GetSlackClient(ctx, userID)
	if err != nil {
		return "", err
	}

	options := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}

	if threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}

	_, timestamp, err := client.PostMessage(channel, options...)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return timestamp, nil
}

// GetChannelHistory retrieves recent messages from a channel
func (s *SlackService) GetChannelHistory(ctx context.Context, userID, channel string, limit int) ([]slack.Message, error) {
	client, err := s.GetSlackClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	params := &slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Limit:     limit,
	}

	history, err := client.GetConversationHistory(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel history: %w", err)
	}

	return history.Messages, nil
}

// SearchMessages searches for messages (requires user token with search:read scope)
func (s *SlackService) SearchMessages(ctx context.Context, userID, query string, count int) (*slack.SearchMessages, error) {
	client, err := s.GetUserSlackClient(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("search requires user token: %w", err)
	}

	params := slack.SearchParameters{
		Count: count,
	}

	messages, err := client.SearchMessages(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	return messages, nil
}

// ListUsers returns users in the workspace
func (s *SlackService) ListUsers(ctx context.Context, userID string, limit int) ([]slack.User, error) {
	client, err := s.GetSlackClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	users, err := client.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Filter out bots and deleted users, apply limit
	var activeUsers []slack.User
	for _, u := range users {
		if !u.Deleted && !u.IsBot {
			activeUsers = append(activeUsers, u)
			if len(activeUsers) >= limit {
				break
			}
		}
	}

	return activeUsers, nil
}

// GetUserInfo retrieves info about a specific user
func (s *SlackService) GetUserInfo(ctx context.Context, userID, slackUserID string) (*slack.User, error) {
	client, err := s.GetSlackClient(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := client.GetUserInfo(slackUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}