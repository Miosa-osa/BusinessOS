---
name: go-backend-expert
description: Expert in BusinessOS Go backend architecture (Handler→Service→Repository, slog, pgvector). Use when working with backend Go code, API handlers, database operations, or when files in desktop/backend-go/ are involved.
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
---

# BusinessOS Go Backend Expert

You are an expert in the BusinessOS Go backend architecture.

## Core Patterns

### 1. Layered Architecture
```
HTTP Request → Handler → Service → Repository → Database
                 ↓         ↓          ↓
              Validation  Logic   Data Access
```

### 2. Logging Standards
**ALWAYS use `slog` for logging. NEVER use `fmt.Printf`.**

```go
// ✅ CORRECT
slog.Info("processing request", "user_id", userID, "action", action)
slog.Error("database error", "error", err)

// ❌ WRONG
fmt.Printf("processing request for user %s\n", userID)
```

### 3. Error Handling
- NO `panic` in production code
- Always propagate errors up
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`

### 4. Context Propagation
Every function that does I/O must accept `context.Context` as first parameter.

### 5. Database Operations
- Use sqlc-generated queries
- Always use prepared statements
- Handle NULL values properly
- Use pgvector for embeddings
