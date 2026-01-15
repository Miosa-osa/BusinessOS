package windowcapture

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// StreamMessage represents a message in the capture stream
type StreamMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StartCapturePayload is sent to start capturing a window
type StartCapturePayload struct {
	BundleID string  `json:"bundle_id"`
	Quality  float32 `json:"quality,omitempty"` // 0.0 to 1.0, default 0.7
	FPS      int     `json:"fps,omitempty"`     // Frames per second, default 30
}

// FramePayload contains a captured frame
type FramePayload struct {
	WindowID int    `json:"window_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Data     string `json:"data"` // Base64 encoded JPEG
}

// WindowListPayload contains the list of windows
type WindowListPayload struct {
	Windows []WindowInfo `json:"windows"`
}

// PermissionPayload indicates permission status
type PermissionPayload struct {
	Granted bool `json:"granted"`
}

// StreamSession manages a window capture streaming session
type StreamSession struct {
	conn         *websocket.Conn
	bundleID     string
	windowID     int
	quality      float32
	fps          int
	running      bool
	mu           sync.Mutex
	stopChan     chan struct{}
	logger       *slog.Logger
	windowBounds WindowBounds // Current window position for input coordinate translation
}

// NewStreamSession creates a new stream session
func NewStreamSession(conn *websocket.Conn, logger *slog.Logger) *StreamSession {
	return &StreamSession{
		conn:     conn,
		quality:  0.7,
		fps:      30,
		stopChan: make(chan struct{}),
		logger:   logger,
	}
}

// Run handles the WebSocket connection
func (s *StreamSession) Run() {
	defer s.conn.Close()

	// Check screen capture permission first
	if !HasScreenCapturePermission() {
		s.sendMessage("permission", PermissionPayload{Granted: false})
		RequestScreenCapturePermission()
		// Wait a bit for user to grant permission
		time.Sleep(2 * time.Second)
		if !HasScreenCapturePermission() {
			s.sendError("Screen capture permission denied. Please grant permission in System Preferences > Security & Privacy > Privacy > Screen Recording")
			return
		}
	}
	s.sendMessage("permission", PermissionPayload{Granted: true})

	// Handle incoming messages
	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket error", "error", err)
			}
			break
		}

		var msg StreamMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			s.sendError("Invalid message format")
			continue
		}

		switch msg.Type {
		case "start":
			s.handleStart(msg.Payload)
		case "stop":
			s.handleStop()
		case "list_windows":
			s.handleListWindows(msg.Payload)
		case "select_window":
			s.handleSelectWindow(msg.Payload)
		case "input":
			s.handleInput(msg.Payload)
		default:
			s.sendError("Unknown message type: " + msg.Type)
		}
	}

	// Stop capture on disconnect
	s.handleStop()
}

func (s *StreamSession) handleStart(payload interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.sendError("Capture already running")
		return
	}

	// Parse payload
	payloadBytes, _ := json.Marshal(payload)
	var startPayload StartCapturePayload
	if err := json.Unmarshal(payloadBytes, &startPayload); err != nil {
		s.sendError("Invalid start payload")
		return
	}

	if startPayload.BundleID == "" {
		s.sendError("bundle_id is required")
		return
	}

	s.bundleID = startPayload.BundleID
	if startPayload.Quality > 0 && startPayload.Quality <= 1.0 {
		s.quality = startPayload.Quality
	}
	if startPayload.FPS > 0 && startPayload.FPS <= 60 {
		s.fps = startPayload.FPS
	}

	// Find windows for the bundle ID
	windows, err := GetWindowsForBundleID(s.bundleID)
	if err != nil {
		s.sendError("No windows found for app: " + s.bundleID)
		return
	}

	// Send window list to client
	s.sendMessage("windows", WindowListPayload{Windows: windows})

	// If only one window, auto-select it
	if len(windows) == 1 {
		s.windowID = windows[0].WindowID
		s.startCapture()
	}
}

func (s *StreamSession) handleSelectWindow(payload interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payloadBytes, _ := json.Marshal(payload)
	var selectPayload struct {
		WindowID int `json:"window_id"`
	}
	if err := json.Unmarshal(payloadBytes, &selectPayload); err != nil {
		s.sendError("Invalid select_window payload")
		return
	}

	if selectPayload.WindowID <= 0 {
		s.sendError("Invalid window_id")
		return
	}

	s.windowID = selectPayload.WindowID
	s.startCapture()
}

func (s *StreamSession) startCapture() {
	if s.running {
		return
	}

	// Initialize window bounds for input coordinate translation
	s.updateWindowBounds()

	s.running = true
	s.stopChan = make(chan struct{})

	s.sendMessage("started", map[string]interface{}{
		"window_id": s.windowID,
		"fps":       s.fps,
		"quality":   s.quality,
	})

	go s.captureLoop()
}

func (s *StreamSession) captureLoop() {
	interval := time.Second / time.Duration(s.fps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("Starting window capture",
		"bundle_id", s.bundleID,
		"window_id", s.windowID,
		"fps", s.fps,
		"quality", s.quality,
	)

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("Stopping window capture", "window_id", s.windowID)
			return
		case <-ticker.C:
			s.captureFrame()
		}
	}
}

func (s *StreamSession) captureFrame() {
	data, err := CaptureWindowAsJPEG(s.windowID, s.quality)
	if err != nil {
		// Window might have been closed
		s.mu.Lock()
		if s.running {
			s.sendError("Window closed or capture failed")
			s.running = false
			close(s.stopChan)
		}
		s.mu.Unlock()
		return
	}

	// Send frame to client
	frame := FramePayload{
		WindowID: s.windowID,
		Data:     base64.StdEncoding.EncodeToString(data),
	}

	s.sendMessage("frame", frame)
}

func (s *StreamSession) handleListWindows(payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)
	var listPayload struct {
		BundleID string `json:"bundle_id"`
	}
	if err := json.Unmarshal(payloadBytes, &listPayload); err != nil {
		s.sendError("Invalid list_windows payload")
		return
	}

	if listPayload.BundleID == "" {
		s.sendError("bundle_id is required")
		return
	}

	windows, err := GetWindowsForBundleID(listPayload.BundleID)
	if err != nil {
		s.sendMessage("windows", WindowListPayload{Windows: []WindowInfo{}})
		return
	}

	s.sendMessage("windows", WindowListPayload{Windows: windows})
}

func (s *StreamSession) handleStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)
	s.sendMessage("stopped", nil)
}

func (s *StreamSession) sendMessage(msgType string, payload interface{}) {
	msg := StreamMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, _ := json.Marshal(msg)

	if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Error("Failed to send WebSocket message", "error", err)
	}
}

func (s *StreamSession) sendError(errMsg string) {
	msg := StreamMessage{
		Type:  "error",
		Error: errMsg,
	}
	data, _ := json.Marshal(msg)

	if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		s.logger.Error("Failed to send error message", "error", err)
	}
}

func (s *StreamSession) handleInput(payload interface{}) {
	if !s.running || s.windowID <= 0 {
		return
	}

	// Parse input event
	payloadBytes, _ := json.Marshal(payload)
	var input InputEvent
	if err := json.Unmarshal(payloadBytes, &input); err != nil {
		s.logger.Error("Invalid input payload", "error", err)
		return
	}

	// Use stored window bounds for coordinate translation
	bounds := s.windowBounds

	switch input.Type {
	case "mousemove":
		// NOTE: Intentionally NOT forwarding mouse moves to the system cursor.
		// This would hijack the user's mouse and move it to the captured window's position.
		// For a VNC-style viewer, we only forward clicks, scrolls, and keyboard events.
		// The visual cursor hover effect is handled purely in the frontend canvas.
		return

	case "mousedown":
		// Bring window to front on first click
		BringWindowToFront(s.windowID)
		InjectMouseClick(bounds, input.X, input.Y, input.Button, true)

	case "mouseup":
		InjectMouseClick(bounds, input.X, input.Y, input.Button, false)

	case "click":
		BringWindowToFront(s.windowID)
		InjectMouseClick(bounds, input.X, input.Y, input.Button, true)
		InjectMouseClick(bounds, input.X, input.Y, input.Button, false)

	case "dblclick":
		BringWindowToFront(s.windowID)
		InjectDoubleClick(bounds, input.X, input.Y)

	case "scroll":
		InjectScroll(bounds, input.X, input.Y, input.DeltaX, input.DeltaY)

	case "keydown":
		InjectKeyEvent(input.KeyCode, true, input.Modifiers)

	case "keyup":
		InjectKeyEvent(input.KeyCode, false, input.Modifiers)

	case "char":
		if len(input.Char) > 0 {
			for _, char := range input.Char {
				InjectCharacter(char)
			}
		}
	}
}

// updateWindowBounds fetches current window position
func (s *StreamSession) updateWindowBounds() {
	windows, err := GetWindowsForBundleID(s.bundleID)
	if err != nil {
		return
	}

	for _, win := range windows {
		if win.WindowID == s.windowID {
			s.windowBounds = WindowBounds{
				X:      win.X,
				Y:      win.Y,
				Width:  win.Width,
				Height: win.Height,
			}
			break
		}
	}
}
