package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

var (
	dispatchMutex   sync.Mutex
	dispatchedRooms = make(map[string]bool)
)

// LiveKitTokenRequest represents a request for a LiveKit room token
type LiveKitTokenRequest struct {
	RoomName string `json:"room_name"` // Optional - will generate if empty
	Identity string `json:"identity"`  // Optional - will use user ID
}

// LiveKitTokenResponse contains the token and room info
type LiveKitTokenResponse struct {
	Token    string `json:"token"`
	RoomName string `json:"room_name"`
	Identity string `json:"identity"`
	URL      string `json:"url"`
}

// HandleLiveKitToken generates a LiveKit access token for voice sessions
// POST /api/livekit/token
func (h *Handlers) HandleLiveKitToken(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		slog.Error("[LiveKit] Unauthorized access attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req LiveKitTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body - use defaults
		req = LiveKitTokenRequest{}
	}

	// Get LiveKit credentials from environment
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	if apiKey == "" || apiSecret == "" {
		slog.Error("[LiveKit] Missing API credentials",
			"has_key", apiKey != "",
			"has_secret", apiSecret != "")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LiveKit not configured"})
		return
	}

	// Generate room name if not provided
	roomName := req.RoomName
	if roomName == "" {
		// Use consistent room per user (prevents duplicates)
		roomName = "osa-voice-" + user.ID
	}

	// Use user ID as identity if not provided
	identity := req.Identity
	if identity == "" {
		// Use FULL user ID, not truncated
		identity = "user-" + user.ID
	}

	slog.Info("[LiveKit] Generating token",
		"user_id", user.ID,
		"room_name", roomName,
		"identity", identity)

	// Create access token
	at := auth.NewAccessToken(apiKey, apiSecret)

	// Grant permissions for the room
	grant := &auth.VideoGrant{
		RoomJoin:       true,
		Room:           roomName,
		CanPublish:     boolPtr(true),
		CanSubscribe:   boolPtr(true),
		CanPublishData: boolPtr(true),
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetName(user.Name). // Display name
		SetValidFor(time.Hour * 24)
		// NOTE: NOT using SetMetadata for agent dispatch - we do manual dispatch below

	// Generate JWT token
	token, err := at.ToJWT()
	if err != nil {
		slog.Error("[LiveKit] Failed to generate token",
			"user_id", user.ID,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	slog.Info("[LiveKit] Token generated successfully",
		"user_id", user.ID,
		"room_name", roomName)

	// Dispatch agent explicitly via LiveKit API
	// CRITICAL FIX: Prevent duplicate dispatches by holding mutex until after API call
	go func() {
		slog.Info("[LiveKit] 🔵 DISPATCH GOROUTINE STARTED", "room", roomName)

		// Use mutex to prevent race condition when multiple requests come in simultaneously
		dispatchMutex.Lock()
		defer dispatchMutex.Unlock() // FIXED: Hold mutex for entire dispatch operation

		slog.Info("[LiveKit] 🔒 Mutex acquired", "room", roomName)

		// Check if already dispatched
		if dispatchedRooms[roomName] {
			slog.Info("[LiveKit] ⚠️ Agent already dispatched for room, skipping", "room", roomName)
			return
		}

		// Mark as dispatching BEFORE making API call (prevents duplicate requests)
		dispatchedRooms[roomName] = true
		slog.Info("[LiveKit] ✅ Room marked as dispatched", "room", roomName, "map_size", len(dispatchedRooms))

		ctx := context.Background()

		// NOTE: NOT creating room explicitly - LiveKit will auto-create when user joins
		// This prevents triggering auto-dispatch rules that might be configured in LiveKit Cloud

		// Dispatch the voice agent to the room
		slog.Info("[LiveKit] 🚀 About to call CreateDispatch", "room", roomName, "agent_name", "osa-voice-agent")
		agentClient := lksdk.NewAgentDispatchServiceClient(livekitURL, apiKey, apiSecret)
		dispatch, err := agentClient.CreateDispatch(ctx, &livekit.CreateAgentDispatchRequest{
			AgentName: "osa-voice-agent",
			Room:      roomName,
		})
		if err != nil {
			slog.Error("[LiveKit] ❌ Failed to dispatch agent",
				"room", roomName,
				"agent", "osa-voice-agent",
				"error", err)
			// FIXED: Don't delete marker on error - let it persist to prevent retries
			// User will need to reconnect if dispatch fails
			return
		}
		slog.Info("[LiveKit] ✅✅ Agent dispatched successfully",
			"room", roomName,
			"dispatch_id", dispatch.Id)

		// Note: NOT cleaning up dispatchedRooms marker
		// This prevents duplicate agents for the same room
		// Entry will be cleared on server restart or when implementing proper cleanup
	}()

	c.JSON(http.StatusOK, LiveKitTokenResponse{
		Token:    token,
		RoomName: roomName,
		Identity: identity,
		URL:      livekitURL,
	})
}

// HandleLiveKitRoomInfo returns info about available voice rooms
// GET /api/livekit/rooms
func (h *Handlers) HandleLiveKitRooms(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// For now, just return config status
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	livekitURL := os.Getenv("LIVEKIT_URL")

	c.JSON(http.StatusOK, gin.H{
		"configured": apiKey != "",
		"url":        livekitURL,
	})
}
