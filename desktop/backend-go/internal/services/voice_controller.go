package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	voicev1 "github.com/rhl/businessos-backend/proto/voice/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VoiceController orchestrates the complete voice pipeline:
// Audio In → STT (Whisper) → LLM (Agent V2) → TTS (ElevenLabs) → Audio Out
type VoiceController struct {
	voicev1.UnimplementedVoiceServiceServer
	pool           *pgxpool.Pool
	sttService     *WhisperService
	ttsService     *ElevenLabsService
	contextService *TieredContextService

	// Session management
	sessions   map[string]*VoiceSession
	sessionsMu sync.RWMutex
}

// VoiceSession represents an active voice conversation
type VoiceSession struct {
	SessionID   string
	UserID      string
	WorkspaceID string
	AgentRole   string
	State       voicev1.SessionState
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Audio buffering for STT
	audioBuffer []byte
	bufferMu    sync.Mutex

	// Conversation history (uses existing Message type from conversation_intelligence.go)
	messages   []Message
	messagesMu sync.Mutex

	// Cancel function for cleanup
	cancel context.CancelFunc
}

// NewVoiceController creates a new voice controller
func NewVoiceController(
	pool *pgxpool.Pool,
	sttService *WhisperService,
	ttsService *ElevenLabsService,
	contextService *TieredContextService,
) *VoiceController {
	return &VoiceController{
		pool:           pool,
		sttService:     sttService,
		ttsService:     ttsService,
		contextService: contextService,
		sessions:       make(map[string]*VoiceSession),
	}
}

// ProcessVoice handles bidirectional audio streaming
func (vc *VoiceController) ProcessVoice(stream voicev1.VoiceService_ProcessVoiceServer) error {
	ctx := stream.Context()
	var session *VoiceSession

	// First frame should establish the session
	firstFrame, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive first frame: %v", err)
	}

	// Get or create session
	session, err = vc.getOrCreateSession(ctx, firstFrame.SessionId, firstFrame.UserId)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	slog.Info("[VoiceController] Session started",
		"session_id", session.SessionID,
		"user_id", session.UserID)

	// Update session state
	session.State = voicev1.SessionState_LISTENING

	// Send initial state update
	if err := stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_LISTENING,
	}); err != nil {
		return err
	}

	// Process audio frames in a loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("[VoiceController] Session ended",
				"session_id", session.SessionID,
				"reason", ctx.Err())
			vc.cleanupSession(session.SessionID)
			return nil

		default:
			frame, err := stream.Recv()
			if err == io.EOF {
				// Client closed stream
				slog.Info("[VoiceController] Client closed stream", "session_id", session.SessionID)
				vc.cleanupSession(session.SessionID)
				return nil
			}
			if err != nil {
				return status.Errorf(codes.Internal, "receive error: %v", err)
			}

			// Process the audio frame
			if err := vc.processAudioFrame(ctx, session, frame, stream); err != nil {
				slog.Error("[VoiceController] Error processing frame",
					"session_id", session.SessionID,
					"error", err)
				// Send error to client but continue
				stream.Send(&voicev1.AudioResponse{
					Type:  voicev1.ResponseType_ERROR,
					Error: err.Error(),
				})
			}
		}
	}
}

// processAudioFrame handles a single audio frame
func (vc *VoiceController) processAudioFrame(
	ctx context.Context,
	session *VoiceSession,
	frame *voicev1.AudioFrame,
	stream voicev1.VoiceService_ProcessVoiceServer,
) error {
	// Only process user audio (not agent audio echoed back)
	if frame.Direction != "user" {
		return nil
	}

	// Buffer audio data
	session.bufferMu.Lock()
	session.audioBuffer = append(session.audioBuffer, frame.AudioData...)
	session.bufferMu.Unlock()

	// If this is a final frame, process the complete utterance
	if frame.IsFinal {
		return vc.processCompleteUtterance(ctx, session, stream)
	}

	return nil
}

// processCompleteUtterance handles a complete user speech segment
func (vc *VoiceController) processCompleteUtterance(
	ctx context.Context,
	session *VoiceSession,
	stream voicev1.VoiceService_ProcessVoiceServer,
) error {
	// Get buffered audio
	session.bufferMu.Lock()
	audioData := make([]byte, len(session.audioBuffer))
	copy(audioData, session.audioBuffer)
	session.audioBuffer = session.audioBuffer[:0] // Clear buffer
	session.bufferMu.Unlock()

	slog.Info("[VoiceController] Processing complete utterance",
		"session_id", session.SessionID,
		"audio_bytes", len(audioData))

	// 1. STT: Convert audio to text
	session.State = voicev1.SessionState_THINKING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_THINKING,
	})

	// Convert audio bytes to io.Reader for existing Whisper service
	audioReader := bytes.NewReader(audioData)
	transcriptionResult, err := vc.sttService.Transcribe(ctx, audioReader, "wav")
	if err != nil {
		return fmt.Errorf("STT failed: %w", err)
	}

	transcript := transcriptionResult.Text

	slog.Info("[VoiceController] User transcript",
		"session_id", session.SessionID,
		"text", transcript)

	// Send user transcript to client
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_TRANSCRIPT_USER,
		Text: transcript,
	})

	// Add to session history
	session.messagesMu.Lock()
	session.messages = append(session.messages, Message{
		Role:      "user",
		Content:   transcript,
		Timestamp: time.Now(),
	})
	session.messagesMu.Unlock()

	// 2. LLM: Get agent response using Agent V2 system
	agentResponse, err := vc.getAgentResponse(ctx, session, transcript)
	if err != nil {
		return fmt.Errorf("LLM failed: %w", err)
	}

	slog.Info("[VoiceController] Agent response",
		"session_id", session.SessionID,
		"text", agentResponse)

	// Send agent transcript to client
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_TRANSCRIPT_AGENT,
		Text: agentResponse,
	})

	// Add to session history
	session.messagesMu.Lock()
	session.messages = append(session.messages, Message{
		Role:      "agent",
		Content:   agentResponse,
		Timestamp: time.Now(),
	})
	session.messagesMu.Unlock()

	// 3. TTS: Convert text to audio
	session.State = voicev1.SessionState_SPEAKING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_SPEAKING,
	})

	audioBytes, err := vc.ttsService.TextToSpeech(ctx, agentResponse)
	if err != nil {
		return fmt.Errorf("TTS failed: %w", err)
	}

	// Stream audio back in chunks (for low latency)
	chunkSize := 4096
	for i := 0; i < len(audioBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(audioBytes) {
			end = len(audioBytes)
		}

		if err := stream.Send(&voicev1.AudioResponse{
			Type:      voicev1.ResponseType_AUDIO,
			AudioData: audioBytes[i:end],
			Sequence:  uint64(i / chunkSize),
		}); err != nil {
			return fmt.Errorf("failed to send audio chunk: %w", err)
		}
	}

	// Send DONE signal
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_DONE,
	})

	// Back to listening
	session.State = voicev1.SessionState_LISTENING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_LISTENING,
	})

	return nil
}

// getAgentResponse gets LLM response using Agent V2 system
func (vc *VoiceController) getAgentResponse(
	ctx context.Context,
	session *VoiceSession,
	userMessage string,
) (string, error) {
	// TODO: Integrate with Agent V2 system
	// For now, return a placeholder
	// This will be replaced with actual Agent V2 orchestration

	slog.Warn("[VoiceController] Using placeholder agent response (Agent V2 integration pending)")

	// Placeholder response
	return fmt.Sprintf("I heard you say: %s. This is a placeholder response. Agent V2 integration coming next.", userMessage), nil
}

// GetSessionContext retrieves user context for voice session
func (vc *VoiceController) GetSessionContext(
	ctx context.Context,
	req *voicev1.SessionRequest,
) (*voicev1.SessionContext, error) {
	session, err := vc.getOrCreateSession(ctx, req.SessionId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get session: %v", err)
	}

	// TODO: Fetch actual user data from database
	// TODO: Fetch workspace context
	// TODO: Fetch conversation history
	// TODO: Fetch RAG context

	return &voicev1.SessionContext{
		SessionId:   session.SessionID,
		UserId:      session.UserID,
		UserName:    "User", // TODO: fetch from DB
		WorkspaceId: session.WorkspaceID,
		AgentRole:   session.AgentRole,
	}, nil
}

// UpdateSessionState updates session state
func (vc *VoiceController) UpdateSessionState(
	ctx context.Context,
	req *voicev1.SessionStateUpdate,
) (*voicev1.SessionStateResponse, error) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	session, exists := vc.sessions[req.SessionId]
	if !exists {
		return &voicev1.SessionStateResponse{
			Success: false,
			Message: "session not found",
		}, nil
	}

	session.State = req.State
	session.UpdatedAt = time.Now()

	slog.Info("[VoiceController] Session state updated",
		"session_id", req.SessionId,
		"new_state", req.State.String())

	return &voicev1.SessionStateResponse{
		Success: true,
		Message: "state updated",
	}, nil
}

// getOrCreateSession gets an existing session or creates a new one
func (vc *VoiceController) getOrCreateSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (*VoiceSession, error) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	if session, exists := vc.sessions[sessionID]; exists {
		return session, nil
	}

	// Create new session
	sessionCtx, cancel := context.WithCancel(ctx)
	session := &VoiceSession{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: "", // TODO: fetch from user
		AgentRole:   "assistant",
		State:       voicev1.SessionState_IDLE,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		audioBuffer: make([]byte, 0, 1024*1024), // 1MB initial capacity
		messages:    make([]Message, 0, 100),
		cancel:      cancel,
	}

	vc.sessions[sessionID] = session

	// Cleanup after 1 hour of inactivity
	go vc.sessionTimeout(sessionCtx, sessionID, 1*time.Hour)

	return session, nil
}

// sessionTimeout cleans up session after timeout
func (vc *VoiceController) sessionTimeout(ctx context.Context, sessionID string, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		slog.Info("[VoiceController] Session timeout", "session_id", sessionID)
		vc.cleanupSession(sessionID)
	}
}

// cleanupSession removes session and frees resources
func (vc *VoiceController) cleanupSession(sessionID string) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	if session, exists := vc.sessions[sessionID]; exists {
		if session.cancel != nil {
			session.cancel()
		}
		delete(vc.sessions, sessionID)
		slog.Info("[VoiceController] Session cleaned up", "session_id", sessionID)
	}
}
