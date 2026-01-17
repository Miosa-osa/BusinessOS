package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
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
		// Create unique room for this voice session
		roomName = "osa-voice-" + uuid.New().String()[:8]
	}

	// Use user ID as identity if not provided
	identity := req.Identity
	if identity == "" {
		identity = "user-" + user.ID[:8]
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
		SetValidFor(time.Hour * 24).
		SetMetadata(`{"agent":"osa-voice-agent"}`) // Request agent dispatch via metadata

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
	go func() {
		ctx := context.Background()
		agentClient := lksdk.NewAgentDispatchServiceClient(livekitURL, apiKey, apiSecret)

		// Create the room first (if it doesn't exist)
		roomClient := lksdk.NewRoomServiceClient(livekitURL, apiKey, apiSecret)
		_, err := roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
			Name:            roomName,
			EmptyTimeout:    300, // 5 minutes
			MaxParticipants: 10,
		})
		if err != nil {
			slog.Warn("[LiveKit] Room may already exist", "room", roomName, "error", err)
		}

		// Dispatch the voice agent to the room
		dispatch, err := agentClient.CreateDispatch(ctx, &livekit.CreateAgentDispatchRequest{
			AgentName: "osa-voice-agent",
			Room:      roomName,
		})
		if err != nil {
			slog.Error("[LiveKit] Failed to dispatch agent",
				"room", roomName,
				"agent", "osa-voice-agent",
				"error", err)
			return
		}
		slog.Info("[LiveKit] Agent dispatched successfully",
			"room", roomName,
			"dispatch_id", dispatch.Id)
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
