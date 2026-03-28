package errors

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// HintForError
// ---------------------------------------------------------------------------

func TestHintForError_NilReturnsEmpty(t *testing.T) {
	got := HintForError(nil)
	if got != "" {
		t.Errorf("HintForError(nil) = %q, want empty string", got)
	}
}

func TestHintForError_SpecificPatterns(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		phrase string // substring that must appear in the hint
	}{
		{
			name:   "nil pointer dereference",
			err:    errors.New("runtime error: invalid memory address or nil pointer dereference"),
			phrase: "Nil pointer dereference",
		},
		{
			name:   "nil pointer dereference via panic",
			err:    errors.New("panic: nil pointer dereference"),
			phrase: "Nil pointer dereference",
		},
		{
			name:   "undefined method",
			err:    errors.New("undefined method Bar on *Foo"),
			phrase: "Undefined method",
		},
		{
			name:   "undefined method mixed case",
			err:    errors.New("Undefined Method: Foo.Bar"),
			phrase: "Undefined method",
		},
		{
			name:   "import cycle",
			err:    errors.New("import cycle not allowed"),
			phrase: "Import cycle detected",
		},
		{
			name:   "type mismatch",
			err:    errors.New("cannot use x (type int) as type string"),
			phrase: "Type mismatch",
		},
		{
			name:   "type mismatch short message",
			err:    errors.New("cannot use type"),
			phrase: "Type mismatch",
		},
		{
			name:   "slice bounds out of range",
			err:    errors.New("runtime error: index out of range [5] with length 3"),
			phrase: "Slice bounds out of range",
		},
		{
			name:   "slice bounds literal",
			err:    errors.New("slice bounds out of range"),
			phrase: "Slice bounds out of range",
		},
		{
			name:   "send on closed channel",
			err:    errors.New("send on closed channel"),
			phrase: "Send on closed channel",
		},
		{
			name:   "concurrent map access",
			err:    errors.New("concurrent map writes"),
			phrase: "Concurrent map access",
		},
		{
			name:   "concurrent map read write",
			err:    errors.New("fatal error: concurrent map read and map write"),
			phrase: "Concurrent map access",
		},
		{
			name:   "context deadline exceeded",
			err:    errors.New("context deadline exceeded"),
			phrase: "Context canceled or deadline exceeded",
		},
		{
			name:   "context canceled",
			err:    errors.New("context canceled"),
			phrase: "Context canceled or deadline exceeded",
		},
		{
			name:   "connection refused",
			err:    errors.New("dial tcp 127.0.0.1:8090: connect: connection refused"),
			phrase: "Connection refused",
		},
		{
			name:   "connection reset",
			err:    errors.New("read tcp 127.0.0.1:8090->127.0.0.1:54321: read: connection reset by peer"),
			phrase: "Connection reset",
		},
		{
			name:   "unexpected EOF",
			err:    errors.New("unexpected EOF"),
			phrase: "Unexpected EOF",
		},
		{
			name:   "invalid JSON",
			err:    errors.New("invalid JSON: unexpected token"),
			phrase: "Invalid JSON/data format",
		},
		{
			name:   "invalid syntax",
			err:    errors.New("invalid syntax in JSON body"),
			phrase: "Invalid JSON/data format",
		},
		{
			name:   "invalid syntax standalone",
			err:    errors.New("invalid syntax"),
			phrase: "Invalid JSON/data format",
		},
		{
			name:   "permission denied",
			err:    errors.New("open /etc/config: permission denied"),
			phrase: "Permission denied",
		},
		{
			name:   "access denied",
			err:    errors.New("access denied for user"),
			phrase: "Permission denied",
		},
		{
			name:   "no such file",
			err:    errors.New("open /tmp/foo.txt: no such file or directory"),
			phrase: "File not found",
		},
		{
			name:   "not found",
			err:    errors.New("file not found"),
			phrase: "File not found",
		},
		{
			name:   "panic",
			err:    errors.New("panic: runtime error: integer divide by zero"),
			phrase: "Goroutine panicked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HintForError(tt.err)
			if got == "" {
				t.Fatal("HintForError returned empty string, want non-empty hint")
			}
			if !contains(got, tt.phrase) {
				t.Errorf("HintForError = %q\nwant hint containing %q", got, tt.phrase)
			}
		})
	}
}

func TestHintForError_GenericFallback(t *testing.T) {
	err := errors.New("something completely unexpected happened")
	got := HintForError(err)

	if got == "" {
		t.Fatal("HintForError returned empty string for unknown error, want generic fallback")
	}
	if !contains(got, "Error: something completely unexpected happened") {
		t.Errorf("HintForError = %q\nwant generic fallback starting with error message", got)
	}
}

func TestHintForError_CaseInsensitive(t *testing.T) {
	// Verify that casing does not matter for pattern matching
	errUpper := errors.New("NIL POINTER DEREFERENCE")
	got := HintForError(errUpper)
	if !contains(got, "Nil pointer dereference") {
		t.Errorf("case-insensitive match failed: got %q", got)
	}

	errMixed := errors.New("Connection Refused on port 8090")
	got = HintForError(errMixed)
	if !contains(got, "Connection refused") {
		t.Errorf("case-insensitive match failed: got %q", got)
	}
}

// ---------------------------------------------------------------------------
// WvdAViolationHint
// ---------------------------------------------------------------------------

func TestWvdAViolationHint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		phrase  string
	}{
		{
			name:   "deadlock",
			input:  "deadlock detected between goroutine 1 and 2",
			phrase: "Deadlock risk",
		},
		{
			name:   "blocking",
			input:  "blocking operation without timeout in handler",
			phrase: "Deadlock risk",
		},
		{
			name:   "liveness",
			input:  "liveness violation: no termination guarantee",
			phrase: "Liveness violation",
		},
		{
			name:   "infinite loop",
			input:  "infinite loop in sync_with_limit",
			phrase: "Liveness violation",
		},
		{
			name:   "loop without exit",
			input:  "loop without escape condition",
			phrase: "Liveness violation",
		},
		{
			name:   "boundedness",
			input:  "boundedness violation: unbounded queue growth",
			phrase: "Boundedness violation",
		},
		{
			name:   "unbounded",
			input:  "unbounded memory allocation",
			phrase: "Boundedness violation",
		},
		{
			name:   "queue overflow",
			input:  "queue exceeded max size",
			phrase: "Boundedness violation",
		},
		{
			name:   "resource exhaustion",
			input:  "resource exhaustion: too many goroutines",
			phrase: "Resource exhaustion",
		},
		{
			name:   "exhausted connections",
			input:  "connection pool exhausted",
			phrase: "Resource exhaustion",
		},
		{
			name:   "generic soundness violation",
			input:  "some other soundness issue",
			phrase: "Soundness violation",
		},
		{
			name:   "empty input",
			input:  "",
			phrase: "Soundness violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WvdAViolationHint(tt.input)
			if got == "" {
				t.Fatal("WvdAViolationHint returned empty string")
			}
			if !contains(got, tt.phrase) {
				t.Errorf("WvdAViolationHint = %q\nwant hint containing %q", got, tt.phrase)
			}
		})
	}
}

func TestWvdAViolationHint_CaseInsensitive(t *testing.T) {
	got := WvdAViolationHint("DEADLOCK in main goroutine")
	if !contains(got, "Deadlock risk") {
		t.Errorf("case-insensitive match failed: got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ArmstrongViolationHint
// ---------------------------------------------------------------------------

func TestArmstrongViolationHint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		phrase  string
	}{
		{
			name:   "supervision missing",
			input:  "no supervision tree for worker",
			phrase: "Supervision violation",
		},
		{
			name:   "unsupervised",
			input:  "unsupervised goroutine launched",
			phrase: "Supervision violation",
		},
		{
			name:   "swallow error",
			input:  "error swallowed in defer recover",
			phrase: "Let-it-crash violation",
		},
		{
			name:   "silent error",
			input:  "silent failure ignored",
			phrase: "Let-it-crash violation",
		},
		{
			name:   "catch exception",
			input:  "broad catch hides root cause",
			phrase: "Let-it-crash violation",
		},
		{
			name:   "shared state",
			input:  "shared mutable state between goroutines",
			phrase: "Shared mutable state detected",
		},
		{
			name:   "race condition",
			input:  "data race detected",
			phrase: "Shared mutable state detected",
		},
		{
			name:   "mutex missing",
			input:  "missing mutex around global variable",
			phrase: "Shared mutable state detected",
		},
		{
			name:   "budget exceeded",
			input:  "budget exceeded: operation took 30s",
			phrase: "Budget violation",
		},
		{
			name:   "timeout",
			input:  "timeout without fallback",
			phrase: "Budget violation",
		},
		{
			name:   "generic fault tolerance violation",
			input:  "unknown fault tolerance issue",
			phrase: "Fault tolerance violation",
		},
		{
			name:   "empty input",
			input:  "",
			phrase: "Fault tolerance violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArmstrongViolationHint(tt.input)
			if got == "" {
				t.Fatal("ArmstrongViolationHint returned empty string")
			}
			if !contains(got, tt.phrase) {
				t.Errorf("ArmstrongViolationHint = %q\nwant hint containing %q", got, tt.phrase)
			}
		})
	}
}

func TestArmstrongViolationHint_CaseInsensitive(t *testing.T) {
	got := ArmstrongViolationHint("SUPERVISION tree missing for worker")
	if !contains(got, "Supervision violation") {
		t.Errorf("case-insensitive match failed: got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestHintForError_MultiplePatternsFirstWins(t *testing.T) {
	// An error message containing multiple keywords should match the first
	// pattern in the if-chain (nil pointer check comes first).
	err := errors.New("nil pointer dereference and panic and EOF")
	got := HintForError(err)
	if !contains(got, "Nil pointer dereference") {
		t.Errorf("expected nil pointer to win over panic/EOF, got %q", got)
	}
}

func TestHintForError_ErrorMessageEmbeddedInTypeMismatch(t *testing.T) {
	// Type mismatch with enough words should embed the original message
	err := errors.New("cannot use x (type int) as type string in assignment")
	got := HintForError(err)
	if !contains(got, "cannot use") {
		t.Errorf("type mismatch with long message should embed original, got %q", got)
	}
}

func TestWvdAViolationHint_DeadlockBeforeBoundedness(t *testing.T) {
	// "blocking queue deadlock" -- deadlock/blocking checked first
	got := WvdAViolationHint("blocking queue deadlock detected")
	if !contains(got, "Deadlock risk") {
		t.Errorf("expected deadlock pattern to win, got %q", got)
	}
}

func TestArmstrongViolationHint_SupervisionBeforeSharedState(t *testing.T) {
	// "supervision shared state" -- supervision checked first
	got := ArmstrongViolationHint("supervision shared state violation")
	if !contains(got, "Supervision violation") {
		t.Errorf("expected supervision pattern to win, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
