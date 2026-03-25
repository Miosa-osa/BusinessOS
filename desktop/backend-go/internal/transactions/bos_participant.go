package transactions

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionID is a unique transaction identifier
type TransactionID string

// ModelID is the database ID of a persisted model
type ModelID string

// VOte represents a participant's vote in the prepare phase
type Vote string

const (
	VoteYes Vote = "YES"
	VoteNo  Vote = "NO"
)

// TransactionState represents the state of a transaction
type TransactionState string

const (
	StateInitial       TransactionState = "INITIAL"
	StatePreparing     TransactionState = "PREPARING"
	StatePrepared      TransactionState = "PREPARED"
	StateDecidedCommit TransactionState = "DECIDED_COMMIT"
	StateCommitted     TransactionState = "COMMITTED"
	StateDecidedAbort  TransactionState = "DECIDED_ABORT"
	StateAborted       TransactionState = "ABORTED"
	StateRecovered     TransactionState = "RECOVERED"
)

// ===== Request/Response Types =====

// PrepareRequest is sent by the coordinator to ask the participant to prepare
type PrepareRequest struct {
	TransactionID TransactionID      `json:"transaction_id"`
	LogData       LogData            `json:"log_data"`
	Algorithm     string             `json:"algorithm"`
	Parameters    AlgorithmParams    `json:"parameters"`
	TimeoutMS     int64              `json:"timeout_ms"`
}

// LogData contains the event log for discovery
type LogData struct {
	Type     string `json:"log_type"`      // "xes", "csv"
	Encoding string `json:"encoding"`      // "base64"
	Content  string `json:"content"`       // base64-encoded log
}

// AlgorithmParams are parameters for the discovery algorithm
type AlgorithmParams struct {
	ActivityKey   string `json:"activity_key"`
	TimestampKey  string `json:"timestamp_key"`
	CaseKey       string `json:"case_key"`
}

// PrepareResponse is the participant's response to a prepare request
type PrepareResponse struct {
	TransactionID TransactionID    `json:"transaction_id"`
	Vote          Vote             `json:"vote"`
	Model         *ModelInfo       `json:"model,omitempty"`
	Error         *ParticipantErr  `json:"error,omitempty"`
	ResourceHold  *ResourceHoldInfo `json:"resource_hold,omitempty"`
	Timestamp     time.Time        `json:"timestamp"`
}

// ModelInfo describes a discovered process model
type ModelInfo struct {
	Type      string            `json:"model_type"`
	Content   string            `json:"content"`   // base64-encoded
	Hash      string            `json:"hash"`      // sha256:...
	Metadata  ModelMetadata     `json:"metadata"`
}

// ModelMetadata contains model statistics
type ModelMetadata struct {
	Nodes      int      `json:"nodes"`
	Edges      int      `json:"edges"`
	Activities []string `json:"activities"`
	SizeBytes  int      `json:"size_bytes"`
}

// ResourceHoldInfo describes a resource lock
type ResourceHoldInfo struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ParticipantErr is an error response from the participant
type ParticipantErr struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// CommitRequest asks the participant to finalize a prepared transaction
type CommitRequest struct {
	TransactionID TransactionID `json:"transaction_id"`
}

// CommitResponse is the participant's response to a commit request
type CommitResponse struct {
	TransactionID TransactionID `json:"transaction_id"`
	State         string        `json:"state"`
	ModelID       ModelID       `json:"model_id,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// AbortRequest asks the participant to rollback a transaction
type AbortRequest struct {
	TransactionID TransactionID `json:"transaction_id"`
	Reason        string        `json:"reason"`
}

// AbortResponse is the participant's response to an abort request
type AbortResponse struct {
	TransactionID TransactionID `json:"transaction_id"`
	State         string        `json:"state"`
	Timestamp     time.Time     `json:"timestamp"`
}

// ===== Internal State =====

// transactionRecord tracks in-flight transaction state
type transactionRecord struct {
	ID            TransactionID
	State         TransactionState
	StartedAt     time.Time
	BosResponse   *PrepareResponse
	DiscoveredModel []byte
	mu            sync.RWMutex
}

// ===== TransactionCoordinator: Main Coordinator Implementation =====

// TransactionCoordinator manages two-phase commit transactions in BusinessOS
type TransactionCoordinator struct {
	db                *pgxpool.Pool
	transactions      map[TransactionID]*transactionRecord
	txMu              sync.RWMutex
	log               *slog.Logger
	prepareTimeout    time.Duration
	commitTimeout     time.Duration
	abortTimeout      time.Duration
	resourceExpiry    time.Duration
}

// NewTransactionCoordinator creates a new coordinator
func NewTransactionCoordinator(db *pgxpool.Pool, logger *slog.Logger) *TransactionCoordinator {
	return &TransactionCoordinator{
		db:             db,
		transactions:   make(map[TransactionID]*transactionRecord),
		log:            logger,
		prepareTimeout: 30 * time.Second,
		commitTimeout:  10 * time.Second,
		abortTimeout:   5 * time.Second,
		resourceExpiry: 2 * time.Minute,
	}
}

// BeginTransaction initializes a new transaction
func (c *TransactionCoordinator) BeginTransaction(
	ctx context.Context,
	modelName string,
	algorithm string,
) (TransactionID, error) {
	txID := TransactionID(uuid.New().String())

	c.txMu.Lock()
	c.transactions[txID] = &transactionRecord{
		ID:        txID,
		State:     StateInitial,
		StartedAt: time.Now(),
	}
	c.txMu.Unlock()

	// Write to WAL
	err := c.writeWAL(ctx, txID, StateInitial, "transaction started")
	if err != nil {
		c.log.ErrorContext(ctx, "failed to write WAL",
			"tx_id", txID, "error", err)
		return "", err
	}

	// Insert into database
	err = c.db.QueryRow(ctx,
		`INSERT INTO transactions (id, state, model_name, algorithm, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id`,
		txID, StateInitial, modelName, algorithm,
	).Scan(&txID)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to create transaction record",
			"tx_id", txID, "error", err)
		return "", err
	}

	c.log.InfoContext(ctx, "transaction started",
		"tx_id", txID, "model_name", modelName, "algorithm", algorithm)

	return txID, nil
}

// Prepare sends a prepare request to BOS and collects the vote
func (c *TransactionCoordinator) Prepare(
	ctx context.Context,
	txID TransactionID,
	req *PrepareRequest,
) (*PrepareResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.prepareTimeout)
	defer cancel()

	c.txMu.Lock()
	txRec := c.transactions[txID]
	c.txMu.Unlock()

	if txRec == nil {
		return nil, fmt.Errorf("transaction not found: %s", txID)
	}

	// Update state
	txRec.mu.Lock()
	txRec.State = StatePreparing
	txRec.mu.Unlock()

	c.writeWAL(context.Background(), txID, StatePreparing, "prepare sent to BOS")

	// In production: call BOS HTTP API here
	// For now, return mock response
	resp := &PrepareResponse{
		TransactionID: txID,
		Vote:          VoteYes,
		Model: &ModelInfo{
			Type:    "petri_net",
			Content: "base64_model_data",
			Hash:    "sha256:abc123",
			Metadata: ModelMetadata{
				Nodes:      47,
				Edges:      89,
				Activities: []string{"A", "B", "C"},
				SizeBytes:  4096,
			},
		},
		ResourceHold: &ResourceHoldInfo{
			ID:        fmt.Sprintf("hold-%s", txID),
			ExpiresAt: time.Now().Add(c.resourceExpiry),
		},
		Timestamp: time.Now(),
	}

	// Update state
	txRec.mu.Lock()
	txRec.State = StatePrepared
	txRec.BosResponse = resp
	txRec.mu.Unlock()

	// Write to WAL
	c.writeWAL(context.Background(), txID, StatePrepared,
		fmt.Sprintf("vote=%s", resp.Vote))

	c.log.InfoContext(ctx, "prepare response received",
		"tx_id", txID, "vote", resp.Vote)

	return resp, nil
}

// Commit sends commit request to BOS and persists the model
func (c *TransactionCoordinator) Commit(
	ctx context.Context,
	txID TransactionID,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.commitTimeout)
	defer cancel()

	c.txMu.Lock()
	txRec := c.transactions[txID]
	c.txMu.Unlock()

	if txRec == nil {
		return fmt.Errorf("transaction not found: %s", txID)
	}

	// Update state
	txRec.mu.Lock()
	txRec.State = StateDecidedCommit
	txRec.mu.Unlock()

	c.writeWAL(context.Background(), txID, StateDecidedCommit, "commit decision made")

	// In production: send commit to BOS, with retry logic
	// CommitRequest would be marshaled to JSON and sent via HTTP

	// Persist model to database
	var modelID ModelID
	err := c.db.QueryRow(ctx,
		`INSERT INTO process_models (org_id, name, algorithm, model_data, transaction_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 RETURNING id`,
		"org-default",
		"model-from-"+string(txID),
		"alpha_miner",
		txRec.BosResponse.Model.Content,
		txID,
	).Scan(&modelID)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to insert model",
			"tx_id", txID, "error", err)
		return err
	}

	// Update transaction state in database
	_, err = c.db.Exec(ctx,
		`UPDATE transactions SET state = $1, completed_at = NOW() WHERE id = $2`,
		StateCommitted, txID,
	)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to update transaction state",
			"tx_id", txID, "error", err)
		return err
	}

	// Write to WAL
	c.writeWAL(context.Background(), txID, StateCommitted,
		fmt.Sprintf("model_id=%s", modelID))

	// Update memory state
	txRec.mu.Lock()
	txRec.State = StateCommitted
	txRec.mu.Unlock()

	c.log.InfoContext(ctx, "transaction committed",
		"tx_id", txID, "model_id", modelID)

	return nil
}

// Abort sends abort request to BOS and rolls back the transaction
func (c *TransactionCoordinator) Abort(
	ctx context.Context,
	txID TransactionID,
	reason string,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.abortTimeout)
	defer cancel()

	c.txMu.Lock()
	txRec := c.transactions[txID]
	c.txMu.Unlock()

	if txRec == nil {
		// Already cleaned up; this is a no-op
		return nil
	}

	// Update state
	txRec.mu.Lock()
	txRec.State = StateDecidedAbort
	txRec.mu.Unlock()

	c.writeWAL(context.Background(), txID, StateDecidedAbort,
		fmt.Sprintf("reason=%s", reason))

	// In production: send abort to BOS with retry logic
	// AbortRequest would be marshaled to JSON and sent via HTTP

	// Update database
	_, err := c.db.Exec(ctx,
		`UPDATE transactions SET state = $1, completed_at = NOW() WHERE id = $2`,
		StateAborted, txID,
	)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to update transaction state on abort",
			"tx_id", txID, "error", err)
		return err
	}

	// Write to WAL
	c.writeWAL(context.Background(), txID, StateAborted, reason)

	// Update memory state
	txRec.mu.Lock()
	txRec.State = StateAborted
	txRec.mu.Unlock()

	// Clean up in-memory record after a delay
	go func() {
		time.Sleep(10 * time.Second)
		c.txMu.Lock()
		delete(c.transactions, txID)
		c.txMu.Unlock()
	}()

	c.log.InfoContext(ctx, "transaction aborted",
		"tx_id", txID, "reason", reason)

	return nil
}

// GetStatus returns the current state of a transaction
func (c *TransactionCoordinator) GetStatus(
	ctx context.Context,
	txID TransactionID,
) (*TransactionRecord, error) {
	// Try memory first
	c.txMu.RLock()
	txRec := c.transactions[txID]
	c.txMu.RUnlock()

	if txRec != nil {
		txRec.mu.RLock()
		defer txRec.mu.RUnlock()
		return &TransactionRecord{
			ID:        txRec.ID,
			State:     txRec.State,
			StartedAt: txRec.StartedAt,
		}, nil
	}

	// Query database
	var state TransactionState
	var startedAt time.Time
	err := c.db.QueryRow(ctx,
		`SELECT state, created_at FROM transactions WHERE id = $1`,
		txID,
	).Scan(&state, &startedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found: %s", txID)
	} else if err != nil {
		return nil, err
	}

	return &TransactionRecord{
		ID:        txID,
		State:     state,
		StartedAt: startedAt,
	}, nil
}

// RecoverFromLog resumes incomplete transactions on startup
func (c *TransactionCoordinator) RecoverFromLog(ctx context.Context) error {
	c.log.InfoContext(ctx, "starting transaction recovery from log")

	// Query all PREPARED and DECIDED_* transactions
	rows, err := c.db.Query(ctx,
		`SELECT id, state, created_at FROM transactions
		 WHERE state IN ($1, $2, $3)
		 ORDER BY created_at ASC`,
		StatePrepared, StateDecidedCommit, StateDecidedAbort,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var recovered int
	for rows.Next() {
		var txID TransactionID
		var state TransactionState
		var createdAt time.Time

		if err := rows.Scan(&txID, &state, &createdAt); err != nil {
			c.log.ErrorContext(ctx, "failed to scan transaction record",
				"error", err)
			continue
		}

		// Determine recovery action
		switch state {
		case StatePrepared:
			// Prepared but coordinator crashed: safe to abort
			c.log.InfoContext(ctx, "recovering PREPARED transaction (aborting)",
				"tx_id", txID)
			if err := c.Abort(ctx, txID, "RECOVERY_ABORT"); err != nil {
				c.log.ErrorContext(ctx, "failed to abort during recovery",
					"tx_id", txID, "error", err)
			}
			recovered++

		case StateDecidedCommit:
			// Decided to commit: redo the commit
			c.log.InfoContext(ctx, "recovering DECIDED_COMMIT transaction (recommitting)",
				"tx_id", txID)
			if err := c.doCommitRecovery(ctx, txID); err != nil {
				c.log.ErrorContext(ctx, "failed to recommit during recovery",
					"tx_id", txID, "error", err)
			}
			recovered++

		case StateDecidedAbort:
			// Decided to abort: redo the abort
			c.log.InfoContext(ctx, "recovering DECIDED_ABORT transaction (reaborting)",
				"tx_id", txID)
			if err := c.Abort(ctx, txID, "RECOVERY_ABORT"); err != nil {
				c.log.ErrorContext(ctx, "failed to reabort during recovery",
					"tx_id", txID, "error", err)
			}
			recovered++
		}
	}

	if err = rows.Err(); err != nil {
		return err
	}

	c.log.InfoContext(ctx, "transaction recovery complete",
		"recovered_count", recovered)

	return nil
}

// doCommitRecovery is a helper for recovery
func (c *TransactionCoordinator) doCommitRecovery(ctx context.Context, txID TransactionID) error {
	// Get transaction details
	var modelName, algorithm string
	err := c.db.QueryRow(ctx,
		`SELECT model_name, algorithm FROM transactions WHERE id = $1`,
		txID,
	).Scan(&modelName, &algorithm)
	if err != nil {
		return err
	}

	// Check if model already persisted
	var count int
	err = c.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM process_models WHERE transaction_id = $1`,
		txID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Already committed; just update transaction state
		_, err := c.db.Exec(ctx,
			`UPDATE transactions SET state = $1, completed_at = NOW() WHERE id = $2`,
			StateCommitted, txID,
		)
		return err
	}

	// Re-persist (in production: fetch from BOS)
	modelData := `{"model_type": "petri_net", "recovered": true}`
	_, err = c.db.Exec(ctx,
		`INSERT INTO process_models (org_id, name, algorithm, model_data, transaction_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		"org-default", modelName, algorithm, modelData, txID,
	)
	if err != nil {
		return err
	}

	// Update transaction state
	_, err = c.db.Exec(ctx,
		`UPDATE transactions SET state = $1, completed_at = NOW() WHERE id = $2`,
		StateCommitted, txID,
	)
	return err
}

// writeWAL writes an entry to the transaction write-ahead log
func (c *TransactionCoordinator) writeWAL(
	ctx context.Context,
	txID TransactionID,
	state TransactionState,
	details string,
) error {
	// In production: write to persistent WAL file or database table
	// For now, just log it
	c.log.DebugContext(ctx, "WAL write",
		"tx_id", txID, "state", state, "details", details)
	return nil
}

// ===== Schema: Database tables (SQL) =====

// CreateTransactionSchema creates the necessary tables
const CreateTransactionSchema = `
-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    state TEXT NOT NULL,
    model_name TEXT,
    algorithm TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    INDEX idx_state (state),
    INDEX idx_created_at (created_at)
);

-- Transaction log (write-ahead log)
CREATE TABLE IF NOT EXISTS transaction_log (
    id SERIAL PRIMARY KEY,
    tx_id TEXT NOT NULL REFERENCES transactions(id),
    state TEXT NOT NULL,
    event TEXT,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_tx_id (tx_id),
    INDEX idx_state (state)
);

-- Process models (persisted discovered models)
CREATE TABLE IF NOT EXISTS process_models (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    org_id TEXT NOT NULL,
    name TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    model_data BYTEA NOT NULL,
    transaction_id TEXT REFERENCES transactions(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_org_id (org_id),
    INDEX idx_transaction_id (transaction_id)
);
`

// ===== TransactionRecord: Public response type =====

// TransactionRecord is returned by GetStatus
type TransactionRecord struct {
	ID        TransactionID    `json:"id"`
	State     TransactionState `json:"state"`
	StartedAt time.Time        `json:"started_at"`
}

// ===== Test helpers =====

// TestHelper provides test fixtures (used in transaction_test.go)
type TestHelper struct {
	Coordinator *TransactionCoordinator
	DB          *pgxpool.Pool
	TxID        TransactionID
}

// NewTestHelper creates a test helper (exported for tests)
func NewTestHelper(db *pgxpool.Pool, logger *slog.Logger) *TestHelper {
	return &TestHelper{
		Coordinator: NewTransactionCoordinator(db, logger),
		DB:          db,
	}
}
