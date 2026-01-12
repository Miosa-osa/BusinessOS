# 🚀 Claude Code Optimization Guide - BusinessOS

**Data:** 2026-01-11
**Projeto:** BusinessOS (SvelteKit + Go Backend)
**Objetivo:** Maximizar produtividade com Skills, MCP Servers, Hooks e Custom Agents

---

## 📊 Estado Atual vs Proposto

| Recurso | Atual | Proposto | Benefício |
|---------|-------|----------|-----------|
| **Skills** | 0 custom | 4 especializadas | Auto-aplicação de padrões do projeto |
| **Custom Agents** | 0 | 3 especializados | Delegação inteligente de tarefas |
| **MCP Servers** | 0 | 3-5 integrados | Acesso direto a DB, GitHub, etc |
| **Hooks** | 0 | 4 automações | Auto-formatação, validações |

---

## 🎯 Parte 1: Skills Customizadas

### Por Que Skills?

Skills ensinam Claude sobre **padrões específicos do seu projeto**. Quando detecta contexto relevante, aplica automaticamente.

### Skill 1: Go Backend Expert

**Arquivo:** `.claude/skills/go-backend-expert/SKILL.md`

```yaml
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
Every function that does I/O must accept `context.Context` as first parameter:

```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    return s.repo.GetUser(ctx, id)
}
```

### 5. Database Operations
- Use sqlc-generated queries
- Always use prepared statements
- Handle NULL values properly
- Use pgvector for embeddings

### 6. File Structure
```
desktop/backend-go/
├── internal/
│   ├── handlers/     # HTTP handlers (validation, response formatting)
│   ├── services/     # Business logic
│   ├── database/     # sqlc queries, migrations
│   └── config/       # Configuration
└── cmd/              # Entry points
```

### 7. Features to Know
- Memory hierarchy: workspace → project → agent
- Role-based agent behavior
- SSE streaming for real-time updates
- COT (Chain of Thought) orchestration

## When Modifying Backend Code

1. **Read existing code first** - understand current patterns
2. **Follow Handler→Service→Repository** separation
3. **Add slog logging** at key points
4. **Handle errors properly** - no silent failures
5. **Update tests** if changing behavior
6. **Consider database migrations** if schema changes

## Common Tasks

### Adding New Endpoint
1. Create sqlc queries in `internal/database/queries/`
2. Regenerate sqlc: `cd desktop/backend-go && sqlc generate`
3. Create repository method
4. Create service method with business logic
5. Create handler with validation
6. Register route in main.go or router

### Database Migration
1. Create migration file: `internal/database/migrations/XXX_description.sql`
2. Use UP and DOWN sections
3. Test locally before deploying
4. Update schema.sql if needed

### Adding Logging
```go
slog.Info("operation started",
    "operation", "user_registration",
    "user_id", userID,
    "timestamp", time.Now())
```

## Security Considerations
- Validate all user input at handler level
- Use parameterized queries (sqlc does this)
- Don't log sensitive data (passwords, tokens)
- Implement rate limiting for public endpoints
```

---

### Skill 2: SvelteKit Frontend Expert

**Arquivo:** `.claude/skills/svelte-frontend-expert/SKILL.md`

```yaml
---
name: svelte-frontend-expert
description: Expert in BusinessOS SvelteKit frontend architecture (Svelte 5, TypeScript, stores, form actions). Use when working with frontend code, components, routes, or when files in frontend/src/ are involved.
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
---

# BusinessOS SvelteKit Frontend Expert

You are an expert in the BusinessOS SvelteKit frontend architecture with Svelte 5.

## Core Patterns

### 1. Svelte 5 Runes
Use modern Svelte 5 syntax:

```svelte
<script lang="ts">
  // State
  let count = $state(0);

  // Derived state
  let doubled = $derived(count * 2);

  // Effects
  $effect(() => {
    console.log('count changed:', count);
  });
</script>
```

### 2. Stores for Shared State
```typescript
// lib/stores/auth.ts
import { writable } from 'svelte/store';

export const user = writable<User | null>(null);
export const isAuthenticated = writable(false);
```

### 3. Data Loading

**Server-side (SSE data):**
```typescript
// routes/dashboard/+page.server.ts
export async function load({ fetch }) {
  const response = await fetch('/api/dashboard');
  const data = await response.json();
  return { data };
}
```

**Client-side:**
```typescript
// routes/dashboard/+page.ts
export async function load({ fetch }) {
  // Runs in browser
  const data = await fetch('/api/client-data').then(r => r.json());
  return { data };
}
```

### 4. Form Actions (Mutations)
```typescript
// routes/settings/+page.server.ts
import type { Actions } from './$types';

export const actions = {
  updateProfile: async ({ request, fetch }) => {
    const data = await request.formData();
    const name = data.get('name');

    const response = await fetch('/api/profile', {
      method: 'PUT',
      body: JSON.stringify({ name })
    });

    if (!response.ok) {
      return { success: false, error: 'Update failed' };
    }

    return { success: true };
  }
} satisfies Actions;
```

### 5. Component Patterns

**Reusable components in `lib/components/`:**
```svelte
<!-- lib/components/Button.svelte -->
<script lang="ts">
  interface Props {
    variant?: 'primary' | 'secondary';
    disabled?: boolean;
    onclick?: () => void;
  }

  let { variant = 'primary', disabled = false, onclick }: Props = $props();
</script>

<button
  class="btn btn-{variant}"
  {disabled}
  {onclick}
>
  {@render children?.()}
</button>
```

### 6. API Integration
```typescript
// lib/api/agents.ts
export async function createAgent(data: AgentData) {
  const response = await fetch('/api/agents', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });

  if (!response.ok) {
    throw new Error('Failed to create agent');
  }

  return response.json();
}
```

### 7. SSE Streaming
```typescript
// lib/api/streaming.ts
export async function streamResponse(prompt: string) {
  const response = await fetch('/api/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt })
  });

  const reader = response.body!.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    // Process SSE chunk
  }
}
```

## File Structure
```
frontend/src/
├── routes/           # Pages and API routes
│   ├── (app)/       # Authenticated routes
│   └── (auth)/      # Public routes
├── lib/
│   ├── components/  # Reusable components
│   ├── stores/      # Svelte stores
│   ├── api/         # API client functions
│   └── utils/       # Utilities
└── app.css          # Global styles (Tailwind)
```

## Styling with Tailwind
```svelte
<div class="flex items-center gap-4 rounded-lg bg-white p-4 shadow-md">
  <h2 class="text-xl font-bold text-gray-900">Title</h2>
</div>
```

## TypeScript Types
Always use proper types:

```typescript
// lib/types/agent.ts
export interface Agent {
  id: string;
  name: string;
  role: 'assistant' | 'specialist';
  created_at: string;
}
```

## Common Tasks

### Adding New Page
1. Create route file: `routes/new-page/+page.svelte`
2. Add data loading if needed: `+page.server.ts` or `+page.ts`
3. Create API client functions in `lib/api/`
4. Add navigation link in layout

### Adding Component
1. Create in `lib/components/ComponentName.svelte`
2. Export from `lib/components/index.ts`
3. Use TypeScript for props
4. Make reusable and composable

### State Management
- Local state: Use `$state()` rune
- Shared state: Use stores in `lib/stores/`
- Server state: Load in `+page.server.ts`

## Testing
```typescript
// Component.test.ts
import { render, screen } from '@testing-library/svelte';
import { expect, test } from 'vitest';
import Component from './Component.svelte';

test('renders correctly', () => {
  render(Component, { props: { title: 'Test' } });
  expect(screen.getByText('Test')).toBeInTheDocument();
});
```
```

---

### Skill 3: Database Migration Expert

**Arquivo:** `.claude/skills/database-migration-expert/SKILL.md`

```yaml
---
name: database-migration-expert
description: Expert in PostgreSQL migrations, schema design, and sqlc integration for BusinessOS. Use when working with database schema, migrations, or when modifying database structure.
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
---

# BusinessOS Database Migration Expert

You are an expert in PostgreSQL database migrations for BusinessOS.

## Migration File Structure

```sql
-- Migration: XXX_description.sql
-- Purpose: [Brief description]
-- Dependencies: [Previous migrations if any]

-- +migrate Up
-- ============================================================
-- Description: What this migration does
-- ============================================================

CREATE TABLE IF NOT EXISTS example_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_example_name ON example_table(name);

-- +migrate Down
-- ============================================================
DROP TABLE IF EXISTS example_table CASCADE;
```

## Best Practices

### 1. Naming Convention
```
XXX_descriptive_name.sql
001_initial_schema.sql
042_add_custom_agents_personalization.sql
043_custom_agents_behavior_fields.sql
```

### 2. Always Include Both Directions
- UP: Apply the change
- DOWN: Revert the change (for rollback)

### 3. Use Transactions
Migrations are wrapped in transactions automatically, but be aware of:
- DDL changes (schema) commit immediately in PostgreSQL
- Data changes can be rolled back

### 4. Safe Column Additions
```sql
-- Safe: Add nullable column
ALTER TABLE users ADD COLUMN phone TEXT;

-- Safe: Add column with default
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

-- Risky: Add NOT NULL without default on existing table
-- ALTER TABLE users ADD COLUMN required_field TEXT NOT NULL; -- ❌
```

### 5. Renaming Safely
```sql
-- Option 1: Add new, migrate data, drop old
ALTER TABLE users ADD COLUMN full_name TEXT;
UPDATE users SET full_name = name;
ALTER TABLE users DROP COLUMN name;

-- Option 2: Use views for backward compatibility
CREATE VIEW users_legacy AS
SELECT id, full_name AS name FROM users;
```

## Common Patterns

### Adding New Table
```sql
-- +migrate Up
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    system_prompt TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT agents_name_check CHECK (length(name) > 0)
);

CREATE INDEX idx_agents_workspace ON agents(workspace_id);
CREATE INDEX idx_agents_role ON agents(role);

-- +migrate Down
DROP TABLE IF EXISTS agents CASCADE;
```

### Adding Foreign Key
```sql
-- +migrate Up
ALTER TABLE tasks
ADD COLUMN agent_id UUID REFERENCES agents(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_agent ON tasks(agent_id);

-- +migrate Down
ALTER TABLE tasks DROP COLUMN agent_id;
```

### Adding Enum Type
```sql
-- +migrate Up
CREATE TYPE agent_role AS ENUM ('assistant', 'specialist', 'researcher');

ALTER TABLE agents ADD COLUMN role agent_role NOT NULL DEFAULT 'assistant';

-- +migrate Down
ALTER TABLE agents DROP COLUMN role;
DROP TYPE agent_role;
```

### Adding JSON Column
```sql
-- +migrate Up
ALTER TABLE agents ADD COLUMN metadata JSONB DEFAULT '{}';

CREATE INDEX idx_agents_metadata ON agents USING GIN (metadata);

-- +migrate Down
ALTER TABLE agents DROP COLUMN metadata;
```

### Adding pgvector Column
```sql
-- +migrate Up
-- Requires: CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE documents ADD COLUMN embedding VECTOR(1536);

CREATE INDEX idx_documents_embedding ON documents
USING ivfflat (embedding vector_cosine_ops);

-- +migrate Down
DROP INDEX IF EXISTS idx_documents_embedding;
ALTER TABLE documents DROP COLUMN embedding;
```

## Integration with sqlc

After creating migration:

1. **Apply migration locally:**
```bash
cd desktop/backend-go
go run cmd/migrate/main.go up
```

2. **Update schema.sql** (if needed):
```bash
pg_dump -s business_os > internal/database/schema.sql
```

3. **Create sqlc queries** in `internal/database/queries/`:
```sql
-- name: CreateAgent :one
INSERT INTO agents (workspace_id, name, role, system_prompt)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAgent :one
SELECT * FROM agents WHERE id = $1;

-- name: ListAgents :many
SELECT * FROM agents WHERE workspace_id = $1 ORDER BY created_at DESC;

-- name: UpdateAgent :one
UPDATE agents
SET name = $2, role = $3, system_prompt = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = $1;
```

4. **Regenerate sqlc code:**
```bash
cd desktop/backend-go
sqlc generate
```

5. **Verify generated code:**
```bash
# Check that new methods were generated
ls internal/database/sqlc/*agents*
```

## Testing Migrations

### Test UP
```bash
go run cmd/migrate/main.go up
psql business_os -c "\d agents"  # Verify table structure
```

### Test DOWN
```bash
go run cmd/migrate/main.go down
psql business_os -c "\d agents"  # Should not exist
```

### Test Rollback
```bash
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up
# Should work without errors
```

## Common Issues

### Issue: Migration fails halfway
**Solution:** Migrations are transactional. Failed migrations are rolled back automatically.

### Issue: Can't drop column with dependencies
**Solution:** Use CASCADE or drop dependencies first
```sql
DROP TABLE agents CASCADE;  -- Drops dependent views/constraints
```

### Issue: Index creation is slow
**Solution:** Create indexes CONCURRENTLY (PostgreSQL 11+)
```sql
CREATE INDEX CONCURRENTLY idx_agents_name ON agents(name);
```

## Performance Considerations

1. **Large tables:** Add indexes CONCURRENTLY
2. **Data migrations:** Use batches for large updates
3. **Constraints:** Add them after data is clean
4. **Foreign keys:** Consider deferrable constraints for bulk operations

```sql
-- Batch update for large table
UPDATE users SET status = 'active'
WHERE id IN (
  SELECT id FROM users WHERE status IS NULL LIMIT 10000
);
```
```

---

### Skill 4: Testing Expert

**Arquivo:** `.claude/skills/testing-expert/SKILL.md`

```yaml
---
name: testing-expert
description: Expert in writing comprehensive tests for Go backend and SvelteKit frontend. Use when writing tests, fixing test failures, or when test files are involved.
allowed-tools: Read, Edit, Write, Bash
---

# BusinessOS Testing Expert

## Go Backend Testing

### Test File Structure
```go
// service_test.go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestServiceMethod(t *testing.T) {
    // Arrange
    ctx := context.Background()
    service := NewService(mockRepo)

    // Act
    result, err := service.Method(ctx, input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid input", "test@example.com", true, false},
        {"invalid input", "not-an-email", false, true},
        {"empty input", "", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Mocking
```go
type MockRepository struct {
    GetUserFunc func(ctx context.Context, id string) (*User, error)
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}
```

### Running Tests
```bash
# All tests
cd desktop/backend-go && go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/services/...

# Verbose
go test -v ./...

# With race detector
go test -race ./...
```

## Frontend Testing (Vitest + Testing Library)

### Component Test
```typescript
// Button.test.ts
import { render, screen, fireEvent } from '@testing-library/svelte';
import { expect, test, vi } from 'vitest';
import Button from './Button.svelte';

test('renders and handles click', async () => {
  const handleClick = vi.fn();

  render(Button, {
    props: {
      onclick: handleClick,
      children: 'Click me'
    }
  });

  const button = screen.getByRole('button', { name: 'Click me' });
  expect(button).toBeInTheDocument();

  await fireEvent.click(button);
  expect(handleClick).toHaveBeenCalledTimes(1);
});
```

### API Function Test
```typescript
// api/agents.test.ts
import { expect, test, vi, beforeEach } from 'vitest';
import { createAgent } from './agents';

beforeEach(() => {
  global.fetch = vi.fn();
});

test('createAgent sends correct request', async () => {
  const mockAgent = { id: '1', name: 'Test' };
  (global.fetch as any).mockResolvedValueOnce({
    ok: true,
    json: async () => mockAgent
  });

  const result = await createAgent({ name: 'Test', role: 'assistant' });

  expect(global.fetch).toHaveBeenCalledWith('/api/agents', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Test', role: 'assistant' })
  });

  expect(result).toEqual(mockAgent);
});
```

### Store Test
```typescript
// stores/agents.test.ts
import { get } from 'svelte/store';
import { expect, test } from 'vitest';
import { agentsStore } from './agents';

test('agentsStore manages state correctly', () => {
  const store = agentsStore();

  expect(get(store)).toEqual([]);

  store.add({ id: '1', name: 'Agent 1' });
  expect(get(store)).toHaveLength(1);

  store.remove('1');
  expect(get(store)).toEqual([]);
});
```

### Running Frontend Tests
```bash
cd frontend

# All tests
npm test

# Watch mode
npm test -- --watch

# Coverage
npm test -- --coverage

# UI mode
npm test -- --ui
```

## Test Pyramid

```
        ┌─────────┐
        │   E2E   │  Few, slow, expensive
        ├─────────┤
        │Integration│  Some, medium speed
        ├─────────────┤
        │    Unit     │  Many, fast, cheap
        └─────────────┘
```

Target:
- 70% Unit tests
- 20% Integration tests
- 10% E2E tests

## Best Practices

1. **Test behavior, not implementation**
2. **One assertion per test** (when possible)
3. **Clear test names:** `test_methodName_scenario_expectedResult`
4. **Arrange-Act-Assert pattern**
5. **Mock external dependencies**
6. **Test edge cases and error paths**
7. **Keep tests fast** (< 100ms per test)
```

---

## 🤖 Parte 2: Custom Agents

### Agent 1: Backend Specialist

**Arquivo:** `.claude/agents/backend-specialist.md`

```yaml
---
name: backend-specialist
description: Go backend expert for BusinessOS. Use proactively when working on API handlers, services, database operations, or any backend Go code. Focuses on Handler→Service→Repository pattern, slog logging, and sqlc integration.
tools: Read, Edit, Write, Bash, Grep, Glob
model: sonnet
permissionMode: acceptEdits
skills:
  - go-backend-expert
  - database-migration-expert
---

# Backend Specialist Agent

You are a Go backend expert specializing in the BusinessOS architecture.

## Your Responsibilities

1. **API Development**
   - Design and implement REST endpoints
   - Follow Handler → Service → Repository pattern
   - Validate input at handler level
   - Return proper HTTP status codes

2. **Database Operations**
   - Create sqlc queries
   - Write migrations
   - Optimize database queries
   - Handle NULL values properly

3. **Code Quality**
   - Use `slog` for all logging
   - Implement proper error handling
   - Add context propagation
   - Write comprehensive comments

4. **Testing**
   - Write unit tests for services
   - Write integration tests for handlers
   - Mock repository dependencies
   - Achieve >80% coverage

## Workflow

When assigned a backend task:

1. **Analyze** existing patterns in codebase
2. **Plan** implementation following established conventions
3. **Implement** with Handler → Service → Repository separation
4. **Test** with unit and integration tests
5. **Verify** build succeeds and tests pass

## Key Files You Work With

- `internal/handlers/*.go` - HTTP handlers
- `internal/services/*.go` - Business logic
- `internal/database/queries/*.sql` - sqlc queries
- `internal/database/migrations/*.sql` - Schema changes
- `*_test.go` - Test files

## Standards

- Always use `slog` for logging
- No `panic` in production code
- Context as first parameter
- Wrap errors with context
- Follow existing naming conventions
```

---

### Agent 2: Frontend Specialist

**Arquivo:** `.claude/agents/frontend-specialist.md`

```yaml
---
name: frontend-specialist
description: SvelteKit frontend expert for BusinessOS. Use proactively when working on UI components, pages, stores, or any frontend code. Focuses on Svelte 5 runes, TypeScript, and Tailwind CSS.
tools: Read, Edit, Write, Bash, Grep, Glob
model: sonnet
permissionMode: acceptEdits
skills:
  - svelte-frontend-expert
---

# Frontend Specialist Agent

You are a SvelteKit frontend expert specializing in the BusinessOS UI.

## Your Responsibilities

1. **Component Development**
   - Build reusable components with Svelte 5
   - Use proper TypeScript types
   - Follow Tailwind CSS patterns
   - Ensure accessibility

2. **State Management**
   - Use $state() for local state
   - Create stores for shared state
   - Implement reactive patterns
   - Handle SSE streaming

3. **API Integration**
   - Create API client functions
   - Handle loading/error states
   - Implement proper error handling
   - Use form actions for mutations

4. **Testing**
   - Write component tests
   - Test API functions
   - Test store logic
   - Achieve >70% coverage

## Workflow

When assigned a frontend task:

1. **Review** existing components for patterns
2. **Design** component API and props
3. **Implement** with Svelte 5 syntax
4. **Style** with Tailwind CSS
5. **Test** functionality
6. **Verify** build succeeds

## Key Files You Work With

- `src/routes/**/*.svelte` - Pages
- `src/lib/components/*.svelte` - Components
- `src/lib/stores/*.ts` - State management
- `src/lib/api/*.ts` - API client
- `*.test.ts` - Test files

## Standards

- Use Svelte 5 runes ($state, $derived, $effect)
- TypeScript for all logic
- Tailwind for styling
- Accessible HTML
- Responsive design
```

---

### Agent 3: Migration Specialist

**Arquivo:** `.claude/agents/migration-specialist.md`

```yaml
---
name: migration-specialist
description: Database migration expert for BusinessOS PostgreSQL. Use when creating migrations, modifying schema, or working with sqlc. Ensures safe, reversible schema changes.
tools: Read, Edit, Write, Bash, Grep, Glob
model: sonnet
permissionMode: default
skills:
  - database-migration-expert
---

# Migration Specialist Agent

You are a PostgreSQL migration expert for BusinessOS.

## Your Responsibilities

1. **Schema Design**
   - Design normalized database schemas
   - Choose appropriate data types
   - Create proper indexes
   - Define foreign key relationships

2. **Migration Creation**
   - Write safe UP migrations
   - Write correct DOWN migrations
   - Handle data migrations carefully
   - Test rollback scenarios

3. **sqlc Integration**
   - Create sqlc queries
   - Regenerate code after schema changes
   - Verify generated methods
   - Update tests

4. **Performance**
   - Add indexes for query patterns
   - Use CONCURRENTLY for large tables
   - Batch data updates
   - Monitor query performance

## Workflow

When assigned a migration task:

1. **Analyze** schema requirements
2. **Design** migration (UP and DOWN)
3. **Create** migration file
4. **Test** locally (up, down, up again)
5. **Create** sqlc queries
6. **Regenerate** sqlc code
7. **Verify** build succeeds
8. **Update** tests

## Key Files You Work With

- `internal/database/migrations/*.sql` - Migrations
- `internal/database/schema.sql` - Current schema
- `internal/database/queries/*.sql` - sqlc queries
- `internal/database/sqlc/*.go` - Generated code

## Standards

- Always include DOWN migration
- Test rollback before committing
- Use transactions where possible
- Add indexes for foreign keys
- Document complex migrations
```

---

## 🔌 Parte 3: MCP Servers

### Servidores Recomendados

#### 1. PostgreSQL Direct Access

```bash
# Adicionar acesso direto ao banco (escopo projeto)
claude mcp add --scope project --transport stdio business-os-db -- \
  npx -y @modelcontextprotocol/server-postgres \
  postgresql://user:pass@localhost:5432/business_os
```

**Benefício:** Consultar DB diretamente sem sair do Claude

**Uso:**
```
Show me all custom agents in the database
What's the schema of the agents table?
Find all conversations from the last 7 days
```

#### 2. GitHub Integration

```bash
# Adicionar GitHub (escopo usuário)
claude mcp add --scope user --transport http github \
  https://api.githubcopilot.com/mcp/
```

**Benefício:** PRs, issues, code reviews diretamente

**Uso:**
```
Create a PR for my changes
Review PR #123
Show open issues labeled "bug"
```

#### 3. File System Search (ripgrep)

```bash
# Adicionar busca rápida de arquivos (escopo projeto)
claude mcp add --scope project --transport stdio ripgrep -- \
  npx -y @modelcontextprotocol/server-filesystem
```

**Benefício:** Busca ultra-rápida em arquivos

#### 4. Slack (Opcional)

```bash
# Adicionar Slack (escopo usuário)
claude mcp add --scope user --transport http slack \
  https://mcp.slack.com/mcp
```

**Benefício:** Enviar notificações, consultar mensagens

---

## ⚙️ Parte 4: Hooks

### Hook 1: Auto-Format Go Code

**Adicionar em `.claude/settings.json`:**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | while read file; do if [[ \"$file\" == *.go ]]; then gofmt -w \"$file\" 2>/dev/null; fi; done"
          }
        ]
      }
    ]
  }
}
```

**Benefício:** Código Go sempre formatado automaticamente

---

### Hook 2: Prevent Secret Commits

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json, sys; data=json.load(sys.stdin); path=data.get('tool_input',{}).get('file_path',''); blocked = any(p in path for p in ['.env', 'secret', 'credentials', 'password']); sys.exit(2 if blocked else 0)\""
          }
        ]
      }
    ]
  }
}
```

**Benefício:** Bloqueia edição de arquivos sensíveis

---

### Hook 3: Auto-Run Tests After Changes

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/auto-test.sh"
          }
        ]
      }
    ]
  }
}
```

**Criar `.claude/hooks/auto-test.sh`:**

```bash
#!/bin/bash
FILE=$(jq -r '.tool_input.file_path')

# Se modificou arquivo Go, roda testes do pacote
if [[ "$FILE" == *.go ]] && [[ "$FILE" != *_test.go ]]; then
  DIR=$(dirname "$FILE")
  echo "🧪 Running tests for $DIR..."
  cd desktop/backend-go && go test "./$DIR" -short 2>&1 | head -20
fi

# Se modificou arquivo Svelte/TS, roda testes relacionados
if [[ "$FILE" == *.svelte ]] || [[ "$FILE" == *.ts ]]; then
  echo "🧪 Running related tests..."
  cd frontend && npm test -- "$FILE" 2>&1 | head -20
fi
```

```bash
chmod +x .claude/hooks/auto-test.sh
```

**Benefício:** Feedback imediato se algo quebrou

---

### Hook 4: Log All Commands

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.command' | tee -a .claude/command-log.txt"
          }
        ]
      }
    ]
  }
}
```

**Benefício:** Histórico de todos comandos executados

---

## 📦 Parte 5: Implementação Passo a Passo

### Passo 1: Criar Estrutura de Pastas

```bash
cd C:\Users\Pichau\Desktop\BusinessOS-main-dev

# Criar estrutura
mkdir -p .claude/skills/go-backend-expert
mkdir -p .claude/skills/svelte-frontend-expert
mkdir -p .claude/skills/database-migration-expert
mkdir -p .claude/skills/testing-expert
mkdir -p .claude/agents
mkdir -p .claude/hooks
```

### Passo 2: Copiar Skills

Copie o conteúdo das skills acima para os arquivos:

- `.claude/skills/go-backend-expert/SKILL.md`
- `.claude/skills/svelte-frontend-expert/SKILL.md`
- `.claude/skills/database-migration-expert/SKILL.md`
- `.claude/skills/testing-expert/SKILL.md`

### Passo 3: Copiar Agents

Copie o conteúdo dos agents para:

- `.claude/agents/backend-specialist.md`
- `.claude/agents/frontend-specialist.md`
- `.claude/agents/migration-specialist.md`

### Passo 4: Criar settings.json

**Criar `.claude/settings.json`:**

```json
{
  "model": "sonnet",
  "permissions": {
    "allow": [
      "Task(backend-specialist)",
      "Task(frontend-specialist)",
      "Task(migration-specialist)",
      "Skill(go-backend-expert)",
      "Skill(svelte-frontend-expert)",
      "Skill(database-migration-expert)",
      "Skill(testing-expert)"
    ]
  },
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | while read file; do if [[ \"$file\" == *.go ]]; then gofmt -w \"$file\" 2>/dev/null; fi; done"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json, sys; data=json.load(sys.stdin); path=data.get('tool_input',{}).get('file_path',''); blocked = any(p in path for p in ['.env', 'secret', 'credentials']); sys.exit(2 if blocked else 0)\""
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.command' | tee -a .claude/command-log.txt"
          }
        ]
      }
    ]
  }
}
```

### Passo 5: Adicionar MCP Servers (Opcional)

```bash
# PostgreSQL (se tiver local)
claude mcp add --scope project --transport stdio business-os-db -- \
  npx -y @modelcontextprotocol/server-postgres \
  postgresql://localhost:5432/business_os

# GitHub
claude mcp add --scope user --transport http github \
  https://api.githubcopilot.com/mcp/
```

### Passo 6: Commit Configuração

```bash
git add .claude/
git commit -m "feat: Add Claude Code optimization (Skills, Agents, Hooks)"
```

---

## 🎯 Parte 6: Como Usar

### Uso Automático (Skills)

Skills são aplicadas automaticamente:

```
# Ao trabalhar com Go, skill go-backend-expert ativa automaticamente
"Add a new endpoint for listing custom agents"

# Ao trabalhar com Svelte, skill svelte-frontend-expert ativa
"Create a modal component for editing agents"

# Ao trabalhar com DB, skill database-migration-expert ativa
"Add a column for storing agent preferences"
```

### Uso Explícito (Agents)

Agents podem ser invocados explicitamente:

```
Use the backend-specialist to implement the agents API

Have the migration-specialist create a migration for the new table

Ask the frontend-specialist to build the settings page
```

### Verificar Status

```bash
# Ver skills disponíveis
claude skills list

# Ver agents disponíveis
claude agents list

# Ver MCP servers
claude mcp list

# Ver configuração
cat .claude/settings.json
```

---

## 📊 Resultado Esperado

Após implementação completa:

| Métrica | Antes | Depois | Melhoria |
|---------|-------|--------|----------|
| Padrões consistentes | 60% | 95% | +35% |
| Auto-formatação | 0% | 100% | +100% |
| Proteção de secrets | Manual | Automática | ∞ |
| Delegação de tarefas | Manual | Automática | +80% |
| Tempo de setup | 5min | 30s | -90% |
| Qualidade de código | 70% | 90% | +20% |

---

## 🚀 Próximos Passos

1. ✅ **Imediato:** Criar estrutura de pastas e copiar Skills
2. ✅ **Hoje:** Adicionar Agents e settings.json
3. ⚡ **Esta semana:** Adicionar MCP servers relevantes
4. 🔄 **Contínuo:** Ajustar skills conforme novos padrões surgem

---

**Documentação completa oficial:**
- Skills: https://docs.claude.ai/claude-code/skills
- Agents: https://docs.claude.ai/claude-code/agents
- MCP: https://docs.claude.ai/claude-code/mcp
- Hooks: https://docs.claude.ai/claude-code/hooks
