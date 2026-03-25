package transactions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ParticipantState represents the state of a participant transaction
type ParticipantState string

const (
	StateInitial      ParticipantState = "INITIAL"
	StatePreparing    ParticipantState = "PREPARING"
	StateReady        ParticipantState = "READY"
	StateCommitting   ParticipantState = "COMMITTING"
	StateCommitted    ParticipantState = "COMMITTED"
	StateRollingBack  ParticipantState = "ROLLING_BACK"
	StateAborted      ParticipantState = "ABORTED"
)

// PrepareRequest contains data for prepare phase
type PrepareRequest struct {
	TransactionID string          `json:"transaction_id"`
	Operation     string          `json:"operation"`
	Data          json.RawMessage `json:"data"`
	Deadline      time.Time       `json:"deadline"`
	Checksum      string          `json:"checksum"`
}

// PrepareResponse contains participant vote
type PrepareResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"` // READY or ABORT
	UndoLog       string    `json:"undo_log"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorReason   string    `json:"error_reason,omitempty"`
}

// CommitRequest signals final commit decision
type CommitRequest struct {
	TransactionID string `json:"transaction_id"`
}

// CommitAck acknowledges successful commit
type CommitAck struct {
	TransactionID string    `json:"transaction_id"`
	CommittedAt   time.Time `json:"committed_at"`
}

// AbortRequest signals abort decision
type AbortRequest struct {
	TransactionID string `json:"transaction_id"`
}

// AbortAck acknowledges successful abort/rollback
type AbortAck struct {
	TransactionID string    `json:"transaction_id"`
	AbortedAt     time.Time `json:"aborted_at"`
}

// UndoOperation represents a single undo action
type UndoOperation struct {
	Type          string          `json:"type"` // DELETE, UNLOCK, etc
	Table         string          `json:"table,omitempty"`
	ID            string          `json:"id,omitempty"`
	Resource      string          `json:"resource,omitempty"`
	PreviousState json.RawMessage `json:"previous_state,omitempty"`
}

// UndoLog contains operations to reverse a transaction
type UndoLog struct {
	TransactionID string           `json:"transaction_id"`
	Operations    []UndoOperation  `json:"operations"`
}

// ParticipantTransaction represents in-memory state
type ParticipantTransaction struct {
	ID              string
	State           ParticipantState
	CreatedAt       time.Time
	Deadline        time.Time
	UndoLog         *UndoLog
	LockedResources map[string]string // resource -> lock_id
	mu              sync.RWMutex
}

// ParticipantLogEntry represents persistent log entry
type ParticipantLogEntry struct {
	Version       int32     `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
	TransactionID string    `json:"transaction_id"`
	CoordinatorID string    `json:"coordinator_id"`
	ParticipantID string    `json:"participant_id"`
	State         string    `json:"state"`
	Operation     string    `json:"operation"`
	DataHash      string    `json:"data_hash"`
	UndoLogID     string    `json:"undo_log_id,omitempty"`
	Error         string    `json:"error,omitempty"`
}

// Participant implements the participant role in 2PC
type Participant struct {
	participantID    string
	logPath          string
	prepareTimeout   time.Duration
	lockTimeout      time.Duration
	transactions     map[string]*ParticipantTransaction
	resourceLocks    map[string]string // resource -> txn_id
	mu               sync.RWMutex
	logger           *slog.Logger
}

// NewParticipant creates a new participant instance
func NewParticipant(
	participantID string,
	logPath string,
	prepareTimeout time.Duration,
	logger *slog.Logger,
) *Participant {
	// Ensure log directory exists
	_ = os.MkdirAll(logPath, 0755)

	if logger == nil {
		logger = slog.Default()
	}

	return &Participant{
		participantID:   participantID,
		logPath:         logPath,
		prepareTimeout:  prepareTimeout,
		lockTimeout:     prepareTimeout + 5*time.Second,
		transactions:    make(map[string]*ParticipantTransaction),
		resourceLocks:   make(map[string]string),
		logger:          logger,
	}
}

// HandlePrepareRequest processes prepare request and returns vote
func (p *Participant) HandlePrepareRequest(
	ctx context.Context,
	req *PrepareRequest,
) (*PrepareResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	txnID := req.TransactionID
	p.logger.Info("Handling prepare request",
		slog.String("txn_id", txnID),
		slog.String("operation", req.Operation),
	)

	// Check if transaction already exists
	if _, exists := p.transactions[txnID]; exists {
		return nil, fmt.Errorf("transaction already exists: %s", txnID)
	}

	// Verify checksum
	dataStr := string(req.Data)
	expectedChecksum := calculateChecksum(dataStr)
	if expectedChecksum != req.Checksum {
		return &PrepareResponse{
			TransactionID: txnID,
			Status:        "ABORT",
			Timestamp:     time.Now().UTC(),
			ErrorReason:   "Checksum mismatch",
		}, nil
	}

	// Create transaction
	txn := &ParticipantTransaction{
		ID:              txnID,
		State:           StatePreparing,
		CreatedAt:       time.Now().UTC(),
		Deadline:        req.Deadline,
		UndoLog:         &UndoLog{TransactionID: txnID, Operations: []UndoOperation{}},
		LockedResources: make(map[string]string),
	}

	// Log PREPARING state
	if err := p.logEntry(txnID, "PREPARING", req.Operation, dataStr, ""); err != nil {
		return &PrepareResponse{
			TransactionID: txnID,
			Status:        "ABORT",
			Timestamp:     time.Now().UTC(),
			ErrorReason:   fmt.Sprintf("Failed to log: %v", err),
		}, nil
	}

	// Validate operation (schema, constraints, auth)
	if err := p.validateOperation(ctx, req); err != nil {
		p.logger.Error("Validation failed",
			slog.String("txn_id", txnID),
			slog.String("error", err.Error()),
		)
		return &PrepareResponse{
			TransactionID: txnID,
			Status:        "ABORT",
			Timestamp:     time.Now().UTC(),
			ErrorReason:   err.Error(),
		}, nil
	}

	// Attempt resource locking
	if err := p.acquireLocks(txn); err != nil {
		p.logger.Error("Lock acquisition failed",
			slog.String("txn_id", txnID),
			slog.String("error", err.Error()),
		)
		return &PrepareResponse{
			TransactionID: txnID,
			Status:        "ABORT",
			Timestamp:     time.Now().UTC(),
			ErrorReason:   fmt.Sprintf("Failed to acquire locks: %v", err),
		}, nil
	}

	// Store transaction
	p.transactions[txnID] = txn
	txn.State = StateReady

	p.logger.Info("Prepare phase succeeded",
		slog.String("txn_id", txnID),
	)

	return &PrepareResponse{
		TransactionID: txnID,
		Status:        "READY",
		UndoLog:       txnID, // Use txn_id as undo_log reference
		Timestamp:     time.Now().UTC(),
	}, nil
}

// HandleCommitRequest persists transaction
func (p *Participant) HandleCommitRequest(
	ctx context.Context,
	req *CommitRequest,
) (*CommitAck, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	txnID := req.TransactionID
	p.logger.Info("Handling commit request",
		slog.String("txn_id", txnID),
	)

	txn, exists := p.transactions[txnID]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txnID)
	}

	if txn.State != StateReady {
		return nil, fmt.Errorf("invalid state for commit: %s", txn.State)
	}

	// Update state to COMMITTING
	txn.State = StateCommitting
	if err := p.logEntry(txnID, "COMMITTING", "", "", ""); err != nil {
		return nil, fmt.Errorf("failed to log: %v", err)
	}

	// Persist to database (simulate with logging)
	p.logger.Info("Persisting transaction",
		slog.String("txn_id", txnID),
	)

	// Update state to COMMITTED
	txn.State = StateCommitted
	if err := p.logEntry(txnID, "COMMITTED", "", "", ""); err != nil {
		return nil, fmt.Errorf("failed to log commit: %v", err)
	}

	// Release locks
	p.releaseLocks(txn)

	p.logger.Info("Commit phase succeeded",
		slog.String("txn_id", txnID),
	)

	return &CommitAck{
		TransactionID: txnID,
		CommittedAt:   time.Now().UTC(),
	}, nil
}

// HandleAbortRequest rolls back transaction
func (p *Participant) HandleAbortRequest(
	ctx context.Context,
	req *AbortRequest,
) (*AbortAck, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	txnID := req.TransactionID
	p.logger.Info("Handling abort request",
		slog.String("txn_id", txnID),
	)

	txn, exists := p.transactions[txnID]
	if !exists {
		// Transaction already cleaned up
		return &AbortAck{
			TransactionID: txnID,
			AbortedAt:     time.Now().UTC(),
		}, nil
	}

	// Update state to ROLLING_BACK
	txn.State = StateRollingBack
	if err := p.logEntry(txnID, "ROLLING_BACK", "", "", ""); err != nil {
		p.logger.Error("Failed to log rollback",
			slog.String("txn_id", txnID),
			slog.String("error", err.Error()),
		)
	}

	// Execute undo operations
	if err := p.executeUndo(txn); err != nil {
		p.logger.Error("Undo execution failed",
			slog.String("txn_id", txnID),
			slog.String("error", err.Error()),
		)
	}

	// Release locks
	p.releaseLocks(txn)

	// Update state to ABORTED
	txn.State = StateAborted
	if err := p.logEntry(txnID, "ABORTED", "", "", ""); err != nil {
		p.logger.Error("Failed to log abort",
			slog.String("txn_id", txnID),
			slog.String("error", err.Error()),
		)
	}

	p.logger.Info("Abort phase succeeded",
		slog.String("txn_id", txnID),
	)

	return &AbortAck{
		TransactionID: txnID,
		AbortedAt:     time.Now().UTC(),
	}, nil
}

// RecoverTransactions recovers from crash
func (p *Participant) RecoverTransactions(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.Info("Starting transaction recovery")

	entries, err := os.ReadDir(p.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read log directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(p.logPath, entry.Name()))
		if err != nil {
			p.logger.Error("Failed to read log file",
				slog.String("file", entry.Name()),
				slog.String("error", err.Error()),
			)
			continue
		}

		var logEntry ParticipantLogEntry
		if err := json.Unmarshal(content, &logEntry); err != nil {
			p.logger.Error("Failed to parse log entry",
				slog.String("file", entry.Name()),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Recover based on state
		switch logEntry.State {
		case "PREPARING", "READY":
			// Prepare was logged but no commit received
			// Conservative: abort via undo
			p.logger.Info("Recovering transaction (abort on recovery)",
				slog.String("txn_id", logEntry.TransactionID),
			)
			// Would apply undo_log here in production
			p.logEntry(logEntry.TransactionID, "RECOVERED_ABORT", "", "", "")
		case "COMMITTED", "ABORTED":
			// Terminal state: nothing to do
			p.logger.Info("Transaction already in terminal state",
				slog.String("txn_id", logEntry.TransactionID),
				slog.String("state", logEntry.State),
			)
		}
	}

	p.logger.Info("Transaction recovery completed")
	return nil
}

// Private helper methods

// validateOperation checks if operation is valid
func (p *Participant) validateOperation(ctx context.Context, req *PrepareRequest) error {
	// In production, would validate against actual schema/constraints
	// For now, simulate successful validation
	if req.Operation == "" {
		return fmt.Errorf("operation cannot be empty")
	}

	// Check deadline
	if time.Now().UTC().After(req.Deadline) {
		return fmt.Errorf("transaction deadline exceeded")
	}

	return nil
}

// acquireLocks locks resources for transaction
func (p *Participant) acquireLocks(txn *ParticipantTransaction) error {
	// In production, would acquire actual database locks
	// For now, simulate lock acquisition
	resource := fmt.Sprintf("process_model_%s", txn.ID[:8])
	lockID := fmt.Sprintf("lock_%s", txn.ID[:8])

	if existing, locked := p.resourceLocks[resource]; locked {
		return fmt.Errorf("resource locked by %s", existing)
	}

	p.resourceLocks[resource] = txn.ID
	txn.LockedResources[resource] = lockID

	// Add undo operation for unlock
	txn.UndoLog.Operations = append(txn.UndoLog.Operations, UndoOperation{
		Type:     "UNLOCK",
		Resource: resource,
	})

	return nil
}

// releaseLocks releases all locks held by transaction
func (p *Participant) releaseLocks(txn *ParticipantTransaction) {
	for resource := range txn.LockedResources {
		delete(p.resourceLocks, resource)
	}
	txn.LockedResources = make(map[string]string)
}

// executeUndo executes undo operations in reverse order
func (p *Participant) executeUndo(txn *ParticipantTransaction) error {
	if txn.UndoLog == nil || len(txn.UndoLog.Operations) == 0 {
		return nil
	}

	// Execute in reverse order
	for i := len(txn.UndoLog.Operations) - 1; i >= 0; i-- {
		op := txn.UndoLog.Operations[i]
		p.logger.Info("Executing undo operation",
			slog.String("txn_id", txn.ID),
			slog.String("type", op.Type),
		)
		// In production, would execute actual undo operations
	}

	return nil
}

// logEntry writes transaction log entry
func (p *Participant) logEntry(
	txnID string,
	state string,
	operation string,
	dataStr string,
	undoLogID string,
) error {
	entry := ParticipantLogEntry{
		Version:       1,
		Timestamp:     time.Now().UTC(),
		TransactionID: txnID,
		ParticipantID: p.participantID,
		State:         state,
		Operation:     operation,
		DataHash:      calculateChecksum(dataStr),
		UndoLogID:     undoLogID,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	logFile := filepath.Join(p.logPath, fmt.Sprintf("txn_%s.log", txnID))
	if err := os.WriteFile(logFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %v", err)
	}

	return nil
}

// calculateChecksum computes SHA256 of data
func calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(hash[:]))
}

// GetTransactionState returns the current state of a transaction
func (p *Participant) GetTransactionState(txnID string) (ParticipantState, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	txn, exists := p.transactions[txnID]
	if !exists {
		return "", fmt.Errorf("transaction not found: %s", txnID)
	}

	return txn.State, nil
}
