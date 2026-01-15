package terminal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/logging"
)

// Manager handles terminal session lifecycle
type Manager struct {
	sessions       map[string]*Session
	mu             sync.RWMutex
	maxSessions    int
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
	containerMgr   *container.ContainerManager
	useContainers  bool
	securityConfig *SessionSecurityConfig
	pubsub         *TerminalPubSub // Optional pub/sub for horizontal scaling
}

// NewManager creates a new terminal session manager
// If containerMgr is provided, sessions will use Docker containers instead of local PTY
func NewManager(containerMgr *container.ContainerManager) *Manager {
	securityConfig := DefaultSessionSecurityConfig()

	m := &Manager{
		sessions:       make(map[string]*Session),
		maxSessions:    100,
		stopCleanup:    make(chan struct{}),
		containerMgr:   containerMgr,
		useContainers:  containerMgr != nil,
		securityConfig: securityConfig,
	}

	if m.useContainers {
		log.Printf("[Terminal] Manager initialized with Docker container support")
	} else {
		log.Printf("[Terminal] Manager initialized with local PTY support")
	}

	log.Printf("[Terminal] Security: max_duration=%v, idle_timeout=%v, ip_binding=%v",
		securityConfig.MaxSessionDuration, securityConfig.IdleTimeout, securityConfig.EnableIPBinding)

	// Start cleanup goroutine
	m.startCleanup()

	return m
}

// SetPubSub configures optional pub/sub for horizontal scaling
func (m *Manager) SetPubSub(pubsub *TerminalPubSub) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pubsub = pubsub
	if pubsub != nil {
		log.Printf("[Terminal] Pub/sub enabled for horizontal scaling (instance: %s)", pubsub.InstanceID())
	}
}

// GetPubSub returns the pub/sub manager (may be nil)
func (m *Manager) GetPubSub() *TerminalPubSub {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pubsub
}

// GetSecurityConfig returns the current security configuration
func (m *Manager) GetSecurityConfig() *SessionSecurityConfig {
	return m.securityConfig
}

// UpdateSecurityConfig updates the security configuration
func (m *Manager) UpdateSecurityConfig(config *SessionSecurityConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.securityConfig = config
}

// shouldUseContainer determines if a session should use Docker based on requested mode
func (m *Manager) shouldUseContainer(requestedMode TerminalMode) bool {
	if requestedMode == TerminalModeLocal {
		logging.Info("[Terminal] 🚫 Mode=LOCAL requested - FORCING PTY (not Docker)")
		return false // User explicitly wants local PTY
	}
	if requestedMode == TerminalModeDocker {
		result := m.useContainers && m.containerMgr != nil
		logging.Info("[Terminal] 🐳 Mode=DOCKER requested - useContainers=%v, containerMgr=%v, result=%v", m.useContainers, m.containerMgr != nil, result)
		return result // Docker only if available
	}
	// Default: use global setting
	logging.Info("[Terminal] ⚙️  Mode=DEFAULT (empty) - using global setting useContainers=%v", m.useContainers)
	return m.useContainers
}

// CreateSession creates a new terminal session
// clientIP is optional but recommended for session hijacking protection
// requestedMode allows per-session mode override (docker/local)
func (m *Manager) CreateSession(userID string, cols, rows int, shell, workingDir string, requestedMode TerminalMode, clientIP ...string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check user session limit (max 5 per user)
	userSessions := m.getUserSessionCount(userID)
	if userSessions >= 5 {
		return nil, fmt.Errorf("session limit reached (max 5 per user)")
	}

	// Check global session limit
	if len(m.sessions) >= m.maxSessions {
		return nil, fmt.Errorf("maximum global session limit reached")
	}

	// Determine working directory based on requested mode
	if workingDir == "" {
		if m.shouldUseContainer(requestedMode) {
			workingDir = "/workspace"
		} else {
			workingDir = getDefaultWorkingDir()
		}
	}

	// Calculate session expiration
	now := time.Now()
	var expiresAt time.Time
	if m.securityConfig.MaxSessionDuration > 0 {
		expiresAt = now.Add(m.securityConfig.MaxSessionDuration)
	}

	// Extract client IP for binding
	var ip, subnet string
	if len(clientIP) > 0 && clientIP[0] != "" {
		ip = clientIP[0]
		subnet = extractSubnet(ip)
	}

	// Create session with security fields
	// Determine if we should use container BEFORE creating session
	// so we can set the correct PS1 prompt
	willUseContainer := m.shouldUseContainer(requestedMode)

	session := &Session{
		ID:            uuid.New().String(),
		UserID:        userID,
		CreatedAt:     now,
		LastActivity:  now,
		Cols:          cols,
		Rows:          rows,
		Shell:         shell,
		WorkingDir:    workingDir,
		Status:        StatusActive,
		RequestedMode: requestedMode,
		Environment:   m.buildEnvironment(userID, willUseContainer),
		ClientIP:      ip,
		ClientSubnet:  subnet,
		ExpiresAt:     expiresAt,
	}

	// Start container or PTY based on requested mode
	if willUseContainer {
		if err := m.startContainer(session); err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}
	} else {
		if err := startPTY(session); err != nil {
			return nil, fmt.Errorf("failed to start PTY: %w", err)
		}
	}

	// Store session
	m.sessions[session.ID] = session

	// Publish session created event for horizontal scaling
	if m.pubsub != nil {
		ctx := context.Background()
		if err := m.pubsub.PublishSessionEvent(ctx, "session_created", session.ID, session.UserID); err != nil {
			log.Printf("[Terminal] Warning: Failed to publish session_created event: %v", err)
		}
	}

	return session, nil
}

// GetSession retrieves a session by ID (without ownership validation)
// WARNING: Use GetSessionSecure for user-facing APIs to prevent unauthorized access
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// GetSessionSecure retrieves a session with full security validation
// Checks: ownership, expiration, and optionally IP binding
func (m *Manager) GetSessionSecure(sessionID, userID, clientIP string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Validate ownership
	if session.UserID != userID {
		log.Printf("[Security] Session %s ownership mismatch: expected %s, got %s",
			sessionID[:8], session.UserID[:8], userID[:8])
		return nil, fmt.Errorf("session access denied")
	}

	// Check expiration
	if session.IsExpired() {
		log.Printf("[Security] Session %s has expired", sessionID[:8])
		return nil, fmt.Errorf("session expired")
	}

	// Check idle timeout
	if session.IsIdle(m.securityConfig.IdleTimeout) {
		log.Printf("[Security] Session %s is idle (timeout: %v)", sessionID[:8], m.securityConfig.IdleTimeout)
		return nil, fmt.Errorf("session timed out due to inactivity")
	}

	// Validate IP binding
	if clientIP != "" {
		valid, reason := session.ValidateIP(clientIP, m.securityConfig)
		if !valid {
			log.Printf("[Security] Session %s IP validation failed: %s (expected %s, got %s)",
				sessionID[:8], reason, session.ClientIP, clientIP)
			return nil, fmt.Errorf("session security violation: %s", reason)
		}
	}

	return session, nil
}

// ValidateSessionAccess performs security validation without retrieving the full session
// Returns (valid bool, reason string)
func (m *Manager) ValidateSessionAccess(sessionID, userID, clientIP string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return false, "session not found"
	}

	if session.UserID != userID {
		return false, "unauthorized access"
	}

	if session.IsExpired() {
		return false, "session expired"
	}

	if session.IsIdle(m.securityConfig.IdleTimeout) {
		return false, "session timed out"
	}

	if clientIP != "" {
		valid, reason := session.ValidateIP(clientIP, m.securityConfig)
		if !valid {
			return false, reason
		}
	}

	return true, ""
}

// GetUserSessions retrieves all sessions for a user
func (m *Manager) GetUserSessions(userID string) []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []SessionInfo
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session.ToInfo())
		}
	}

	return sessions
}

// UpdateActivity updates the last activity time for a session
func (m *Manager) UpdateActivity(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// CloseSession closes a terminal session
func (m *Manager) CloseSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Publish session closed event before cleanup for horizontal scaling
	if m.pubsub != nil {
		ctx := context.Background()
		if err := m.pubsub.PublishSessionEvent(ctx, "session_closed", session.ID, session.UserID); err != nil {
			log.Printf("[Terminal] Warning: Failed to publish session_closed event: %v", err)
		}
	}

	// Close container or PTY based on session type
	if session.IsContainerized() {
		m.closeContainer(session)
	} else {
		closePTY(session)
	}

	// Update status
	session.Status = StatusClosed

	// Remove from active sessions
	delete(m.sessions, sessionID)

	return nil
}

// ResizeSession resizes the terminal
func (m *Manager) ResizeSession(sessionID string, cols, rows int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Cols = cols
	session.Rows = rows

	// Resize container exec or PTY based on session type
	var err error
	if session.IsContainerized() {
		if session.ExecID == "" {
			return fmt.Errorf("exec ID not set for containerized session")
		}
		err = m.containerMgr.ResizeExec(session.ExecID, uint(rows), uint(cols))
	} else {
		err = resizePTY(session, cols, rows)
	}

	// Publish resize event after successful resize for horizontal scaling
	if err == nil && m.pubsub != nil {
		ctx := context.Background()
		if pubErr := m.pubsub.PublishResize(ctx, sessionID, cols, rows); pubErr != nil {
			log.Printf("[Terminal] Warning: Failed to publish resize event: %v", pubErr)
		}
	}

	return err
}

// Shutdown closes all sessions and stops cleanup
func (m *Manager) Shutdown() {
	close(m.stopCleanup)
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all sessions
	for sessionID, session := range m.sessions {
		if session.IsContainerized() {
			m.closeContainer(session)
		} else {
			closePTY(session)
		}
		delete(m.sessions, sessionID)
	}
}

// Helper functions

func (m *Manager) getUserSessionCount(userID string) int {
	count := 0
	for _, session := range m.sessions {
		if session.UserID == userID && session.Status == StatusActive {
			count++
		}
	}
	return count
}

func (m *Manager) buildEnvironment(userID string, useContainer bool) map[string]string {
	env := make(map[string]string)
	env["TERM"] = "xterm-256color"
	env["LANG"] = "en_US.UTF-8"
	env["COLORTERM"] = "truecolor"

	// Enhanced PS1 to show terminal mode
	if useContainer {
		// Running in Docker container - show Docker indicator
		// Shows: [🐳 docker] ~/workspace $
		env["PS1"] = "\\[\\033[1;32m\\][🐳 docker]\\[\\033[0m\\] \\[\\033[1;36m\\]\\w\\[\\033[0m\\] \\$ "
	} else {
		// Running in local PTY (Glimpse mode) - show real Mac prompt
		// Shows: username@hostname currentdir %
		// DON'T set PS1 at all - let zsh/bash use their defaults
		// This will show the actual user@hostname prompt from the system
		// For zsh (default on macOS), this will be: rhl@Robertos-Mac-Studio ~ %
		// We intentionally don't set PS1 here to use system defaults
	}

	env["BUSINESSOS_USER_ID"] = userID // Pass user ID to container for internal API calls
	return env
}

func (m *Manager) startCleanup() {
	m.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.cleanupIdleSessions()
			case <-m.stopCleanup:
				return
			}
		}
	}()
}

func (m *Manager) cleanupIdleSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiredCount, idleCount int

	for sessionID, session := range m.sessions {
		shouldClose := false
		reason := ""

		// Check for hard expiration
		if session.IsExpired() {
			shouldClose = true
			reason = "expired"
			expiredCount++
		} else if session.IsIdle(m.securityConfig.IdleTimeout) {
			// Check for idle timeout
			shouldClose = true
			reason = "idle"
			idleCount++
		}

		if shouldClose {
			log.Printf("[Terminal] Closing session %s (reason: %s)", sessionID[:8], reason)
			if session.IsContainerized() {
				m.closeContainer(session)
			} else {
				closePTY(session)
			}
			delete(m.sessions, sessionID)
		}
	}

	if expiredCount > 0 || idleCount > 0 {
		log.Printf("[Terminal] Cleanup completed: %d expired, %d idle sessions closed",
			expiredCount, idleCount)
	}
}

// startContainer creates and starts a Docker container for the session
func (m *Manager) startContainer(session *Session) error {
	log.Printf("[Terminal] Starting container for session %s (user: %s)", session.ID, session.UserID)

	// Create volume for user workspace
	volumeName, err := m.containerMgr.CreateVolume(session.UserID)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}
	log.Printf("[Terminal] Volume created/verified: %s", volumeName)

	// Create container with the default image (include session ID for unique naming)
	containerID, err := m.containerMgr.CreateContainer(session.UserID, session.ID, m.containerMgr.GetDefaultImage())
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	log.Printf("[Terminal] Container created: %s", containerID[:12])

	// Start container
	if err := m.containerMgr.StartContainer(containerID); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	log.Printf("[Terminal] Container started: %s", containerID[:12])

	// Determine shell command - load BusinessOS init script for osa command
	// The init script is bind-mounted at /etc/businessos/init.sh
	initScript := "/etc/businessos/init.sh"

	// Force bash in containers for reliable --rcfile support
	// zsh in containers has complex rc loading; bash is simpler and universal
	shell := "/bin/bash"
	shellCmd := []string{shell, "--rcfile", initScript}

	// Create and start exec session with shell and environment variables
	execID, hijacked, err := m.containerMgr.StartExecWithEnv(containerID, shellCmd, true, session.Environment)
	if err != nil {
		// If exec fails, try to stop the container
		m.containerMgr.StopContainer(containerID, 10)
		return fmt.Errorf("failed to start exec: %w", err)
	}
	log.Printf("[Terminal] Exec started: %s with env: %v", execID, session.Environment)

	// Store container information in session
	session.ContainerID = containerID
	session.VolumeID = volumeName
	session.ExecID = execID
	session.ExecConn = &hijacked

	// Set initial terminal size
	if err := m.containerMgr.ResizeExec(execID, uint(session.Rows), uint(session.Cols)); err != nil {
		log.Printf("[Terminal] Warning: Failed to set initial terminal size: %v", err)
	}

	log.Printf("[Terminal] Container session ready: container=%s exec=%s", containerID[:12], execID[:12])
	return nil
}

// closeContainer closes the container session and removes the container
func (m *Manager) closeContainer(session *Session) {
	log.Printf("[Terminal] Closing container session %s", session.ID)

	// Close exec connection first
	if session.ExecConn != nil {
		session.ExecConn.Close()
		log.Printf("[Terminal] Exec connection closed for session %s", session.ID)
	}

	// Stop and remove container for immediate cleanup
	// The ContainerMonitor acts as a safety net for any missed containers
	if session.ContainerID != "" {
		containerID := session.ContainerID[:12]

		// Stop container first (with graceful timeout)
		if err := m.containerMgr.StopContainer(session.ContainerID, 5); err != nil {
			log.Printf("[Terminal] Warning: Failed to stop container %s: %v", containerID, err)
		} else {
			log.Printf("[Terminal] Container stopped: %s", containerID)
		}

		// Remove container immediately (force=true handles any remaining state)
		if err := m.containerMgr.RemoveContainer(session.ContainerID, true); err != nil {
			// Not critical - ContainerMonitor will clean up orphaned containers
			log.Printf("[Terminal] Warning: Failed to remove container %s: %v (will be cleaned by monitor)", containerID, err)
		} else {
			log.Printf("[Terminal] Container removed: %s", containerID)
		}
	}

	log.Printf("[Terminal] Container session closed: %s", session.ID)
}
