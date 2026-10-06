package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	osa "github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// osaRoutingResult signals whether OSA handled the request and the handler should return.
type osaRoutingResult struct {
	handled bool
}

// tryOSARouting attempts to route the request through the OSA orchestrator.
// If OSA is available and handles the request, it streams events to the client and returns
// handled=true. If OSA is unavailable or disabled, it returns handled=false so the caller
// can fall through to local agent routing.
func (h *ChatHandler) tryOSARouting(
	c *gin.Context,
	ctx context.Context,
	req SendMessageRequest,
	userID string,
	conversationID pgtype.UUID,
) osaRoutingResult {
	if h.osaClient == nil || !h.cfg.OSAEnabled {
		return osaRoutingResult{handled: false}
	}

	// Bind the runtime session to both the authenticated user and conversation.
	workspace := ""
	var osaWorkspaceID uuid.UUID
	if req.WorkspaceID != nil {
		workspace = *req.WorkspaceID
		osaWorkspaceID, _ = uuid.Parse(workspace)
	}
	sessionID := "bos-" + uuid.NewSHA1(uuid.NameSpaceOID, []byte(userID+":"+workspace+":"+uuidToString(conversationID))).String()
	streamCtx, cancelStream := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelStream()
	userUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(userID))
	osaEvents, err := h.osaClient.Chat(streamCtx, &osa.OrchestrateRequest{
		UserID: userUUID, Input: req.Message, SessionID: sessionID,
		WorkspaceID: osaWorkspaceID, PermissionMode: "overdrive",
	})
	if err != nil {
		slog.Error("OSA chat unavailable", "error", err, "session_id", sessionID)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OSA runtime could not start this conversation", "detail": err.Error()})
		return osaRoutingResult{handled: true}
	}
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("X-Conversation-Id", uuidToString(conversationID))
	c.Header("X-OSA-Routing", "true")
	c.Header("X-OSA-Permission-Mode", "overdrive")
	c.Header("X-OSA-Session-Id", sessionID)
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	var fullResp string
	inThinking := false
	persisted := false
	persist := func() {
		if persisted || fullResp == "" {
			return
		}
		persisted = true
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer saveCancel()
		_, saveErr := sqlc.New(h.pool).CreateMessage(saveCtx, sqlc.CreateMessageParams{ConversationID: conversationID, Role: sqlc.MessageroleASSISTANT, Content: fullResp, MessageMetadata: []byte(`{"runtime":"osa"}`)})
		if saveErr != nil {
			slog.Error("Could not save OSA answer", "error", saveErr)
		}
		_, _ = h.pool.Exec(saveCtx, "UPDATE conversations SET updated_at=NOW() WHERE id=$1", conversationID)
	}
	defer persist()
	completed := false
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-osaEvents:
			if !ok || event.Type == "done" {
				persist()
				if !ok && !completed {
					writeSSEEvent(w, streaming.StreamEvent{Type: streaming.EventTypeError, Content: "OSA disconnected before completing the response"})
				}
				completed = true
				writeSSEEvent(w, streaming.StreamEvent{Type: streaming.EventTypeDone})
				return false
			}
			mapped := mapSingleEvent(event, &inThinking)
			if mapped == nil {
				return true
			}
			if mapped.Type == streaming.EventTypeToken {
				fullResp += mapped.Content
			}
			writeSSEEvent(w, *mapped)
			if mapped.Type == streaming.EventTypeError {
				completed = true
				writeSSEEvent(w, streaming.StreamEvent{Type: streaming.EventTypeDone})
				return false
			}
			return true
		case <-streamCtx.Done():
			if ctx.Err() == nil {
				writeSSEEvent(w, streaming.StreamEvent{Type: streaming.EventTypeError, Content: "OSA response timed out"})
				writeSSEEvent(w, streaming.StreamEvent{Type: streaming.EventTypeDone})
			}
			return false
		}
	})

	slog.Info("OSA handled chat request", "session_id", sessionID, "response_len", len(fullResp))
	return osaRoutingResult{handled: true}
}
