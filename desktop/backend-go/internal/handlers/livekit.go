package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
)

// Removed: Agent dispatch logic (handled by Python agent in dev mode)

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

	// DISABLED: Agent dispatch (Python agent handles this via dev mode)
	// The Python agent running in dev mode will automatically connect when room is created
	slog.Info("[LiveKit] 🔵 Skipping explicit agent dispatch - Python agent will auto-connect", "room", roomName)

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
