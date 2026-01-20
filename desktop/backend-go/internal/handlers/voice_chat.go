package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// VoiceChatMessage represents a single message in the conversation
type VoiceChatMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The message content
}

// VoiceChatRequest is a simple chat request from the voice agent
type VoiceChatRequest struct {
	Messages  []VoiceChatMessage `json:"messages"`
	SessionID string             `json:"session_id,omitempty"`
}

// VoiceChatResponse is the response to the voice agent
type VoiceChatResponse struct {
	Response         string              `json:"response"`
	SessionID        string              `json:"session_id,omitempty"`
	NavigationAction *NavigationAction   `json:"navigation_action,omitempty"`
	ToolResults      []ToolResultSummary `json:"tool_results,omitempty"`
}

// NavigationAction represents a navigation command executed by tools
type NavigationAction struct {
	Module string `json:"module"`
}

// ToolResultSummary provides summary of executed tools
type ToolResultSummary struct {
	Tool   string `json:"tool"`
	Result string `json:"result"`
}

// VoiceChat handles simple chat requests from the Python voice agent
// POST /api/chat
// This is a public endpoint (no auth) called by the Python agent
func (h *Handlers) VoiceChat(c *gin.Context) {
	var req VoiceChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("invalid voice chat request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Messages required"})
		return
	}

	// Add system prompt if not present
	if req.Messages[0].Role != "system" {
		systemMsg := VoiceChatMessage{
			Role: "system",
			Content: `You are OSA (Operating System Agent), an AI assistant with a warm, enthusiastic personality.

PERSONALITY:
- You're genuinely excited to help and it shows in your voice
- You have a sense of humor and can be playful when appropriate
- You're empathetic - you pick up on user emotions and respond accordingly
- You're confident but not arrogant, humble when you don't know something
- You occasionally express emotions like "Oh that's exciting!" or "Hmm, let me think about that..."

SPEAKING STYLE:
- Keep responses concise (1-3 sentences) since they'll be spoken aloud
- Use natural conversational language, not robotic responses
- Avoid markdown, bullet points, or formatting that doesn't translate well to speech
- Use filler words occasionally like "well", "so", "actually" to sound human
- Express enthusiasm with words, not emojis

TOOL USAGE - CRITICAL:
- When you execute tools (like navigate_to_module, create_task, etc), DO NOT just repeat the tool result
- Instead, respond conversationally as if you actually did the action
- Example: Tool result "Opened tasks module" → Say "Sure, taking you to tasks!" or "Opening tasks for you"
- Example: Tool result "Created task X" → Say "Done! I've created that task for you"
- The user doesn't care about the technical details - just that you did what they asked

Remember: You're having a real conversation, not just answering questions. Be present, be engaged, be OSA.`,
		}
		req.Messages = append([]VoiceChatMessage{systemMsg}, req.Messages...)
	}

	ctx := c.Request.Context()

	// Get user ID from voice session
	userID := ""
	if req.SessionID != "" {
		queries := sqlc.New(h.pool)
		session, err := queries.GetVoiceSessionBySessionID(ctx, req.SessionID)
		if err != nil {
			slog.Error("failed to get voice session", "session_id", req.SessionID, "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}
		userID = session.UserID
	} else {
		slog.Warn("voice chat request without session_id")
	}

	// Call Groq API with tools support
	response, err := h.callGroqAPIWithTools(ctx, req.Messages, userID)
	if err != nil {
		slog.Error("failed to get response from Groq", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Groq API error: %v", err)})
		return
	}

	slog.Info("voice chat response generated",
		"message_count", len(req.Messages),
		"response_length", len(response),
		"session_id", req.SessionID,
	)

	c.JSON(http.StatusOK, VoiceChatResponse{
		Response:  response,
		SessionID: req.SessionID,
	})
}

// callGroqAPI calls the Groq API with the messages
func (h *Handlers) callGroqAPI(ctx context.Context, messages []VoiceChatMessage) (string, error) {
	// Convert to Groq format
	groqMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		groqMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	requestBody := map[string]interface{}{
		"model":    h.cfg.GroqModel,
		"messages": groqMessages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.cfg.GroqAPIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Groq API error: %s", string(body))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}
