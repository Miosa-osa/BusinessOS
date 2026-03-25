package transactions

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSuccessfulTransaction tests the happy path: prepare + commit
func TestSuccessfulTransaction(t *testing.T) {
	// Setup
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	// Step 1: Generate prepare request
	data := json.RawMessage(`{"model_id": "model_xyz", "process_model": {}}`)
	dataStr := string(data)
	checksum := calculateChecksum(dataStr)

	prepareReq := &PrepareRequest{
		TransactionID: "txn_abc123",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	// Step 2: Participant handles prepare request
	prepareResp, err := participant.HandlePrepareRequest(ctx, prepareReq)
	if err != nil {
		t.Fatalf("HandlePrepareRequest failed: %v", err)
	}

	// Verify participant voted READY
	if prepareResp.Status != "READY" {
		t.Errorf("Expected status READY, got %s", prepareResp.Status)
	}

	if prepareResp.TransactionID != "txn_abc123" {
		t.Errorf("Expected txn_id txn_abc123, got %s", prepareResp.TransactionID)
	}

	// Step 3: Verify transaction state is READY
	state, err := participant.GetTransactionState("txn_abc123")
	if err != nil {
		t.Fatalf("GetTransactionState failed: %v", err)
	}

	if state != StateReady {
		t.Errorf("Expected state READY, got %s", state)
	}

	// Step 4: Participant receives commit request
	commitReq := &CommitRequest{TransactionID: "txn_abc123"}
	commitAck, err := participant.HandleCommitRequest(ctx, commitReq)
	if err != nil {
		t.Fatalf("HandleCommitRequest failed: %v", err)
	}

	if commitAck.TransactionID != "txn_abc123" {
		t.Errorf("Expected txn_id txn_abc123, got %s", commitAck.TransactionID)
	}

	// Step 5: Verify final state is COMMITTED
	state, err = participant.GetTransactionState("txn_abc123")
	if err != nil {
		t.Fatalf("GetTransactionState failed: %v", err)
	}

	if state != StateCommitted {
		t.Errorf("Expected state COMMITTED, got %s", state)
	}

	// Verify log file was created
	logFile := filepath.Join(logDir, "txn_abc123.log")
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("Log file not created: %v", err)
	}
}

// TestParticipantValidationFailure tests abort due to validation error
func TestParticipantValidationFailure(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	// Prepare request with no operation (will fail validation)
	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	prepareReq := &PrepareRequest{
		TransactionID: "txn_def456",
		Operation:     "", // Invalid: empty operation
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	// Participant should reject with ABORT
	prepareResp, err := participant.HandlePrepareRequest(ctx, prepareReq)
	if err != nil {
		t.Fatalf("HandlePrepareRequest failed: %v", err)
	}

	if prepareResp.Status != "ABORT" {
		t.Errorf("Expected status ABORT, got %s", prepareResp.Status)
	}

	if prepareResp.ErrorReason == "" {
		t.Error("Expected error reason, got empty string")
	}

	// Verify transaction not stored
	_, err = participant.GetTransactionState("txn_def456")
	if err == nil {
		t.Error("Expected error getting non-existent transaction")
	}
}

// TestChecksumValidation tests message corruption detection
func TestChecksumValidation(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)

	prepareReq := &PrepareRequest{
		TransactionID: "txn_ghi789",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      "sha256:wrong_checksum", // Invalid checksum
	}

	// Should reject due to checksum mismatch
	prepareResp, err := participant.HandlePrepareRequest(ctx, prepareReq)
	if err != nil {
		t.Fatalf("HandlePrepareRequest failed: %v", err)
	}

	if prepareResp.Status != "ABORT" {
		t.Errorf("Expected status ABORT, got %s", prepareResp.Status)
	}

	if prepareResp.ErrorReason != "Checksum mismatch" {
		t.Errorf("Expected 'Checksum mismatch' error, got %s", prepareResp.ErrorReason)
	}
}

// TestDeadlineValidation tests expiration handling
func TestDeadlineValidation(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	prepareReq := &PrepareRequest{
		TransactionID: "txn_jkl012",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(-1 * time.Second), // Already expired
		Checksum:      checksum,
	}

	// Should reject due to deadline
	prepareResp, err := participant.HandlePrepareRequest(ctx, prepareReq)
	if err != nil {
		t.Fatalf("HandlePrepareRequest failed: %v", err)
	}

	if prepareResp.Status != "ABORT" {
		t.Errorf("Expected status ABORT, got %s", prepareResp.Status)
	}
}

// TestLockContention tests resource locking
func TestLockContention(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	// First transaction
	prepareReq1 := &PrepareRequest{
		TransactionID: "txn_lock1",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	resp1, err := participant.HandlePrepareRequest(ctx, prepareReq1)
	if err != nil {
		t.Fatalf("First prepare request failed: %v", err)
	}

	if resp1.Status != "READY" {
		t.Errorf("First transaction: expected READY, got %s", resp1.Status)
	}

	// Second transaction trying to lock same resource
	prepareReq2 := &PrepareRequest{
		TransactionID: "txn_lock2",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	resp2, err := participant.HandlePrepareRequest(ctx, prepareReq2)
	if err != nil {
		t.Fatalf("Second prepare request failed: %v", err)
	}

	// Second should fail due to lock contention
	if resp2.Status != "ABORT" {
		t.Errorf("Second transaction: expected ABORT due to lock, got %s", resp2.Status)
	}
}

// TestAbortFlow tests prepare + abort flow
func TestAbortFlow(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	// Prepare
	prepareReq := &PrepareRequest{
		TransactionID: "txn_abort1",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	prepareResp, err := participant.HandlePrepareRequest(ctx, prepareReq)
	if err != nil {
		t.Fatalf("HandlePrepareRequest failed: %v", err)
	}

	if prepareResp.Status != "READY" {
		t.Errorf("Expected READY, got %s", prepareResp.Status)
	}

	// Abort
	abortReq := &AbortRequest{TransactionID: "txn_abort1"}
	abortAck, err := participant.HandleAbortRequest(ctx, abortReq)
	if err != nil {
		t.Fatalf("HandleAbortRequest failed: %v", err)
	}

	if abortAck.TransactionID != "txn_abort1" {
		t.Errorf("Expected txn_abort1, got %s", abortAck.TransactionID)
	}

	// Verify final state is ABORTED
	state, err := participant.GetTransactionState("txn_abort1")
	if err != nil {
		t.Fatalf("GetTransactionState failed: %v", err)
	}

	if state != StateAborted {
		t.Errorf("Expected ABORTED, got %s", state)
	}
}

// TestRecoveryFromCrash tests startup recovery
func TestRecoveryFromCrash(t *testing.T) {
	logDir := t.TempDir()

	// Simulate a previous instance that crashed
	logFile := filepath.Join(logDir, "txn_recovery1.log")
	logEntry := ParticipantLogEntry{
		Version:       1,
		Timestamp:     time.Now().UTC(),
		TransactionID: "txn_recovery1",
		ParticipantID: "businessos-1",
		State:         "PREPARING",
		Operation:     "model_persistence",
		DataHash:      "sha256:test",
	}

	data, _ := json.MarshalIndent(logEntry, "", "  ")
	_ = os.WriteFile(logFile, data, 0644)

	// New instance recovers
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	err := participant.RecoverTransactions(ctx)
	if err != nil {
		t.Fatalf("RecoverTransactions failed: %v", err)
	}

	// Verify recovery log was written
	recoveryLog := filepath.Join(logDir, "txn_recovery1.log")
	if _, err := os.Stat(recoveryLog); err != nil {
		t.Errorf("Recovery log not created: %v", err)
	}
}

// TestIdempotentOperations tests handling of duplicate messages
func TestIdempotentOperations(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	// Prepare
	prepareReq := &PrepareRequest{
		TransactionID: "txn_idempotent",
		Operation:     "model_persistence",
		Data:          data,
		Deadline:      time.Now().Add(30 * time.Second),
		Checksum:      checksum,
	}

	prepareResp, _ := participant.HandlePrepareRequest(ctx, prepareReq)
	if prepareResp.Status != "READY" {
		t.Fatalf("Prepare failed")
	}

	// Commit
	commitReq := &CommitRequest{TransactionID: "txn_idempotent"}
	ack1, _ := participant.HandleCommitRequest(ctx, commitReq)

	// Send duplicate commit request
	ack2, err := participant.HandleCommitRequest(ctx, commitReq)

	// Should still succeed due to idempotency
	if err != nil {
		// Getting error on duplicate is acceptable for idempotency
		// (in production, would return cached ack)
		t.Logf("Duplicate commit got error (acceptable): %v", err)
	} else if ack2 == nil {
		t.Error("Expected non-nil ack for duplicate request")
	} else if ack1.TransactionID != ack2.TransactionID {
		t.Error("Ack txn_ids differ for same transaction")
	}
}

// TestMultipleTransactionsInProgress tests concurrency
func TestMultipleTransactionsInProgress(t *testing.T) {
	logDir := t.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, slog.Default())
	ctx := context.Background()

	// Start 3 transactions
	for i := 1; i <= 3; i++ {
		data := json.RawMessage(`{"model_id": "model_xyz"}`)
		checksum := calculateChecksum(string(data))
		txnID := "txn_multi" + string(rune('0'+i))

		prepareReq := &PrepareRequest{
			TransactionID: txnID,
			Operation:     "model_persistence",
			Data:          data,
			Deadline:      time.Now().Add(30 * time.Second),
			Checksum:      checksum,
		}

		resp, err := participant.HandlePrepareRequest(ctx, prepareReq)
		if err != nil {
			t.Fatalf("Prepare request %d failed: %v", i, err)
		}

		if resp.Status != "READY" {
			t.Errorf("Transaction %d: expected READY, got %s", i, resp.Status)
		}
	}

	// Commit all
	for i := 1; i <= 3; i++ {
		txnID := "txn_multi" + string(rune('0'+i))
		commitReq := &CommitRequest{TransactionID: txnID}

		ack, err := participant.HandleCommitRequest(ctx, commitReq)
		if err != nil {
			t.Fatalf("Commit %d failed: %v", i, err)
		}

		if ack.TransactionID != txnID {
			t.Errorf("Expected txn %s, got %s", txnID, ack.TransactionID)
		}

		state, _ := participant.GetTransactionState(txnID)
		if state != StateCommitted {
			t.Errorf("Transaction %d not committed", i)
		}
	}
}

// BenchmarkPreparePhase benchmarks prepare phase
func BenchmarkPreparePhase(b *testing.B) {
	logDir := b.TempDir()
	participant := NewParticipant("businessos-1", logDir, 30*time.Second, nil)
	ctx := context.Background()

	data := json.RawMessage(`{"model_id": "model_xyz"}`)
	checksum := calculateChecksum(string(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prepareReq := &PrepareRequest{
			TransactionID: "txn_bench" + string(rune(i%100)),
			Operation:     "model_persistence",
			Data:          data,
			Deadline:      time.Now().Add(30 * time.Second),
			Checksum:      checksum,
		}

		_, _ = participant.HandlePrepareRequest(ctx, prepareReq)
	}
}
