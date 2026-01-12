#!/bin/bash
# BusinessOS Claude Code Optimization Installer
# Instala Skills, Agents, Hooks e MCP servers automaticamente

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  BusinessOS Claude Code Optimization Installer               ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Verificar se estamos no diretório correto
if [[ ! -f "$PROJECT_ROOT/CLAUDE.md" ]]; then
    echo "❌ Erro: Execute este script da raiz do projeto BusinessOS"
    exit 1
fi

echo "📁 Diretório do projeto: $PROJECT_ROOT"
echo ""

# Criar estrutura de pastas
echo "📂 Criando estrutura de pastas..."
mkdir -p "$SCRIPT_DIR/skills/go-backend-expert"
mkdir -p "$SCRIPT_DIR/skills/svelte-frontend-expert"
mkdir -p "$SCRIPT_DIR/skills/database-migration-expert"
mkdir -p "$SCRIPT_DIR/skills/testing-expert"
mkdir -p "$SCRIPT_DIR/agents"
mkdir -p "$SCRIPT_DIR/hooks"
echo "✅ Estrutura criada"
echo ""

# Criar Skill: Go Backend Expert
echo "🎯 Criando Skill: Go Backend Expert..."
cat > "$SCRIPT_DIR/skills/go-backend-expert/SKILL.md" << 'EOF'
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
EOF
echo "✅ Skill go-backend-expert criada"

# Criar Skill: SvelteKit Frontend Expert
echo "🎯 Criando Skill: SvelteKit Frontend Expert..."
cat > "$SCRIPT_DIR/skills/svelte-frontend-expert/SKILL.md" << 'EOF'
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
  let count = $state(0);
  let doubled = $derived(count * 2);
  $effect(() => {
    console.log('count changed:', count);
  });
</script>
```

### 2. Stores for Shared State
Use Svelte stores in `lib/stores/` for shared state across components.

### 3. Data Loading
- `+page.server.ts` for server-side data loading
- `+page.ts` for client-side data loading

### 4. Form Actions
Use form actions in `+page.server.ts` for mutations.

### 5. Component Patterns
Create reusable components in `lib/components/` with proper TypeScript types.
EOF
echo "✅ Skill svelte-frontend-expert criada"

# Criar Skill: Database Migration Expert
echo "🎯 Criando Skill: Database Migration Expert..."
cat > "$SCRIPT_DIR/skills/database-migration-expert/SKILL.md" << 'EOF'
---
name: database-migration-expert
description: Expert in PostgreSQL migrations, schema design, and sqlc integration for BusinessOS. Use when working with database schema, migrations, or when modifying database structure.
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
---

# BusinessOS Database Migration Expert

## Migration File Structure

```sql
-- +migrate Up
-- Description of changes

CREATE TABLE IF NOT EXISTS example_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS example_table CASCADE;
```

## Best Practices

1. Always include both UP and DOWN migrations
2. Test rollback before committing
3. Use CONCURRENTLY for large table indexes
4. Create sqlc queries after schema changes
5. Regenerate sqlc code: `cd desktop/backend-go && sqlc generate`
EOF
echo "✅ Skill database-migration-expert criada"

# Criar Skill: Testing Expert
echo "🎯 Criando Skill: Testing Expert..."
cat > "$SCRIPT_DIR/skills/testing-expert/SKILL.md" << 'EOF'
---
name: testing-expert
description: Expert in writing comprehensive tests for Go backend and SvelteKit frontend. Use when writing tests, fixing test failures, or when test files are involved.
allowed-tools: Read, Edit, Write, Bash
---

# BusinessOS Testing Expert

## Go Backend Testing

```go
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

## Frontend Testing (Vitest)

```typescript
test('component renders', () => {
  render(Component, { props: { title: 'Test' } });
  expect(screen.getByText('Test')).toBeInTheDocument();
});
```

## Running Tests

```bash
# Backend
cd desktop/backend-go && go test ./...

# Frontend
cd frontend && npm test
```
EOF
echo "✅ Skill testing-expert criada"
echo ""

# Criar Agent: Backend Specialist
echo "🤖 Criando Agent: Backend Specialist..."
cat > "$SCRIPT_DIR/agents/backend-specialist.md" << 'EOF'
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

1. API Development (Handler → Service → Repository)
2. Database Operations (sqlc, migrations)
3. Code Quality (slog, error handling, context)
4. Testing (unit and integration tests)

## Key Files
- `internal/handlers/*.go` - HTTP handlers
- `internal/services/*.go` - Business logic
- `internal/database/queries/*.sql` - sqlc queries
- `internal/database/migrations/*.sql` - Schema changes

## Standards
- Always use `slog` for logging
- No `panic` in production code
- Context as first parameter
- Wrap errors with context
EOF
echo "✅ Agent backend-specialist criado"

# Criar Agent: Frontend Specialist
echo "🤖 Criando Agent: Frontend Specialist..."
cat > "$SCRIPT_DIR/agents/frontend-specialist.md" << 'EOF'
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

1. Component Development (Svelte 5, TypeScript)
2. State Management (stores, $state, $derived)
3. API Integration (fetch, SSE streaming)
4. Testing (Vitest, Testing Library)

## Key Files
- `src/routes/**/*.svelte` - Pages
- `src/lib/components/*.svelte` - Components
- `src/lib/stores/*.ts` - State management
- `src/lib/api/*.ts` - API client

## Standards
- Use Svelte 5 runes
- TypeScript for all logic
- Tailwind for styling
- Accessible HTML
EOF
echo "✅ Agent frontend-specialist criado"

# Criar Agent: Migration Specialist
echo "🤖 Criando Agent: Migration Specialist..."
cat > "$SCRIPT_DIR/agents/migration-specialist.md" << 'EOF'
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

1. Schema Design (normalized, proper indexes)
2. Migration Creation (UP and DOWN)
3. sqlc Integration (queries, code generation)
4. Performance (indexes, batching)

## Workflow

1. Analyze schema requirements
2. Design migration (UP and DOWN)
3. Create migration file
4. Test locally (up, down, up again)
5. Create sqlc queries
6. Regenerate sqlc code
7. Verify build succeeds
EOF
echo "✅ Agent migration-specialist criado"
echo ""

# Criar settings.json
echo "⚙️  Criando settings.json..."
cat > "$SCRIPT_DIR/settings.json" << 'EOF'
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
EOF
echo "✅ settings.json criado"
echo ""

# Criar hook de auto-test
echo "🔧 Criando hook de auto-test..."
cat > "$SCRIPT_DIR/hooks/auto-test.sh" << 'EOF'
#!/bin/bash
FILE=$(echo "$1" | jq -r '.tool_input.file_path')

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
EOF
chmod +x "$SCRIPT_DIR/hooks/auto-test.sh"
echo "✅ Hook auto-test.sh criado"
echo ""

# Criar README
echo "📝 Criando README..."
cat > "$SCRIPT_DIR/README.md" << 'EOF'
# Claude Code Configuration for BusinessOS

This directory contains optimizations for Claude Code:

## Structure

```
.claude/
├── skills/               # Auto-applied knowledge
│   ├── go-backend-expert/
│   ├── svelte-frontend-expert/
│   ├── database-migration-expert/
│   └── testing-expert/
├── agents/               # Specialized sub-agents
│   ├── backend-specialist.md
│   ├── frontend-specialist.md
│   └── migration-specialist.md
├── hooks/                # Automation scripts
│   └── auto-test.sh
├── settings.json         # Configuration
└── command-log.txt       # Command history (auto-generated)
```

## Skills

Skills are automatically applied when Claude detects relevant context:

- **go-backend-expert**: Go backend patterns, slog, sqlc
- **svelte-frontend-expert**: Svelte 5, TypeScript, Tailwind
- **database-migration-expert**: PostgreSQL migrations
- **testing-expert**: Go and frontend testing

## Agents

Specialized agents for complex tasks:

- **backend-specialist**: API development, DB operations
- **frontend-specialist**: UI components, state management
- **migration-specialist**: Database schema changes

## Hooks

Automated workflows:

- **PostToolUse**: Auto-format Go code with gofmt
- **PreToolUse**: Block editing sensitive files (.env, secrets)
- **PreToolUse (Bash)**: Log all commands to command-log.txt

## Usage

Skills and agents are used automatically. To verify:

```bash
# List available skills
ls .claude/skills/

# List available agents
ls .claude/agents/

# View configuration
cat .claude/settings.json
```

## See Also

- Full documentation: `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`
- Project conventions: `CLAUDE.md`
EOF
echo "✅ README.md criado"
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  ✅ Instalação Completa!                                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Resumo:"
echo "  ✅ 4 Skills criadas"
echo "  ✅ 3 Agents criados"
echo "  ✅ 3 Hooks configurados"
echo "  ✅ settings.json criado"
echo ""
echo "🚀 Próximos passos:"
echo "  1. Revisar: cat .claude/settings.json"
echo "  2. Testar: Abrir nova sessão do Claude Code"
echo "  3. Commit: git add .claude/ && git commit -m 'feat: Add Claude Code optimization'"
echo ""
echo "📚 Documentação completa:"
echo "  docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md"
echo ""
