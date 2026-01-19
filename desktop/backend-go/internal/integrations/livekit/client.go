package livekit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	lkproto "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// Client provides methods to interact with LiveKit
type Client struct {
	config *Config
	logger *slog.Logger
}

// NewClient creates a new LiveKit client
func NewClient(config *Config, logger *slog.Logger) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Client{
		config: config,
		logger: logger,
	}, nil
}

// createRoom creates a LiveKit room (idempotent - won't fail if room exists)
func (c *Client) createRoom(ctx context.Context, roomName string) error {
	roomService := lksdk.NewRoomServiceClient(c.config.URL, c.config.APIKey, c.config.APISecret)

	_, err := roomService.CreateRoom(ctx, &lkproto.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    30, // 30 seconds - close quickly when empty
		MaxParticipants: 10,
	})

	if err != nil {
		c.logger.Debug("room create (may already exist)", "room", roomName, "error", err)
		// Don't return error - room may already exist, which is fine
	}

	return nil
}

// createAgentDispatch dispatches an agent to a room
func (c *Client) createAgentDispatch(roomName, agentName string) error {
	// Convert wss:// to https:// for REST API
	apiURL := strings.Replace(c.config.URL, "wss://", "https://", 1)
	apiURL = strings.Replace(apiURL, "ws://", "http://", 1)
	apiURL = apiURL + "/twirp/livekit.AgentDispatchService/CreateDispatch"

	payload := map[string]interface{}{
		"room":       roomName,
		"agent_name": agentName,
	}
	body, _ := json.Marshal(payload)

	// Create JWT for API auth
	at := auth.NewAccessToken(c.config.APIKey, c.config.APISecret)
	at.AddGrant(&auth.VideoGrant{
		RoomAdmin: true,
		Room:      roomName,
	}).SetValidFor(time.Minute)

	token, err := at.ToJWT()
	if err != nil {
		return fmt.Errorf("failed to create dispatch token: %w", err)
	}

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch agent: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.logger.Info("agent dispatch response", "status", resp.StatusCode, "result", result)

	return nil
}

// GenerateRoomToken generates a JWT token for a user to join a LiveKit room
func (c *Client) GenerateRoomToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("token request cannot be nil")
	}

	// Validate required fields
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if req.UserName == "" {
		return nil, fmt.Errorf("user name is required")
	}

	// Generate room name if not provided
	roomName := req.RoomName
	if roomName == "" {
		// Format: ws_{workspace_id}_{user_id}_{timestamp}
		// Use first 8 chars of IDs for shorter room names
		userIDShort := req.UserID
		if len(userIDShort) > 8 {
			userIDShort = userIDShort[:8]
		}
		workspaceIDShort := req.WorkspaceID
		if len(workspaceIDShort) > 8 {
			workspaceIDShort = workspaceIDShort[:8]
		}
		roomName = fmt.Sprintf("ws_%s_%s_%d",
			workspaceIDShort,
			userIDShort,
			time.Now().Unix(),
		)
	}

	// Create room first (idempotent - won't fail if room exists)
	if err := c.createRoom(ctx, roomName); err != nil {
		c.logger.Error("failed to create room", "room", roomName, "error", err)
		// Continue anyway - the room may already exist
	}

	// Dispatch agent to the room ASYNCHRONOUSLY to avoid blocking token generation
	// The user can start connecting while the agent spins up in the background
	agentName := req.AgentRole
	if agentName == "" {
		agentName = "groq-agent" // Default to groq
	}
	go func() {
		// Use background context with timeout (not the request context which will be cancelled)
		dispatchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		c.logger.Info("dispatching agent asynchronously", "room", roomName, "agent", agentName)
		if err := c.createAgentDispatch(roomName, agentName); err != nil {
			c.logger.Error("failed to dispatch agent", "room", roomName, "agent", agentName, "error", err)
		} else {
			c.logger.Info("agent dispatched successfully", "room", roomName, "agent", agentName)
		}
		_ = dispatchCtx // Use the context variable
	}()

	// Identity for the participant
	identity := req.UserID

	// Create access token using LiveKit SDK
	at := auth.NewAccessToken(c.config.APIKey, c.config.APISecret)

	// Set participant identity
	at.SetIdentity(identity)

	// Set participant name
	at.SetName(req.UserName)

	// Set metadata if provided
	if req.Metadata != "" {
		at.SetMetadata(req.Metadata)
	}

	// Grant room permissions
	canPublish := true
	canSubscribe := true
	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         roomName,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	at.AddGrant(grant)

	// Set token validity
	expiresAt := time.Now().Add(c.config.TokenTTL)
	at.SetValidFor(c.config.TokenTTL)

	// Generate JWT
	token, err := at.ToJWT()
	if err != nil {
		c.logger.Error("failed to generate LiveKit token",
			"error", err,
			"user_id", req.UserID,
			"room_name", roomName,
		)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate session ID for tracking
	sessionID := uuid.New().String()

	c.logger.Info("generated LiveKit token",
		"user_id", req.UserID,
		"user_name", req.UserName,
		"room_name", roomName,
		"session_id", sessionID,
		"expires_at", expiresAt,
	)

	return &TokenResponse{
		Token:     token,
		URL:       c.config.URL,
		RoomName:  roomName,
		Identity:  identity,
		ExpiresAt: expiresAt,
		SessionID: sessionID,
	}, nil
}

// GetUserContext retrieves user context for Python agents
// This is called by the Python voice agent to get conversation history and RAG context
func (c *Client) GetUserContext(ctx context.Context, sessionID, userID string) (*ContextResponse, error) {
	// This will be implemented by the handler layer which has database access
	// The client just provides the interface - the handler will populate it
	return nil, fmt.Errorf("GetUserContext must be implemented by handler layer with database access")
}

// ValidateToken validates a LiveKit token (for future use)
func (c *Client) ValidateToken(ctx context.Context, token string) (bool, error) {
	// TODO: Implement token validation using LiveKit SDK
	// The LiveKit SDK doesn't provide a simple token verifier API
	// For now, we'll trust tokens generated by our own client
	// In production, LiveKit server validates tokens on connection
	return true, nil
}

// Config returns the client configuration (read-only)
func (c *Client) Config() *Config {
	return c.config
}
