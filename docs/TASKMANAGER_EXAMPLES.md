# TaskManager Automático - Exemplos Práticos

Este documento mostra o TaskManager em ação com diferentes tipos de tarefas.

---

## 📋 Exemplo 1: Full-Stack Feature

**Usuário pede:**
> "Adicionar sistema de comentários nas memórias"

**TaskManager detecta:**
- Tipo: Full-Stack Feature (Frontend + Backend + Database)
- Complexidade: Moderada
- Ativação: ✅ Automática

---

### 🤖 Resposta do Claude com TaskManager:

```
╔══════════════════════════════════════════════════════════════╗
║ 🤖 TASKMANAGER ATIVO                                         ║
╠══════════════════════════════════════════════════════════════╣
║ Tarefa: Adicionar sistema de comentários nas memórias       ║
║ Tipo: Full-Stack                                             ║
║ Complexidade: Moderada                                       ║
╠══════════════════════════════════════════════════════════════╣
║ 📊 DECOMPOSIÇÃO AUTOMÁTICA                                   ║
║                                                              ║
║ 🔄 PARALELO (3 tracks):                                      ║
║   • Track A: Database schema → Subagent: general-purpose    ║
║   • Track B: Backend API → Subagent: Explore + Plan         ║
║   • Track C: Frontend UI → Subagent: Explore                ║
║                                                              ║
║ ➡️ SEQUENCIAL (3 steps):                                     ║
║   • Step 1: Integrar frontend ↔ backend → Aguarda: A,B,C    ║
║   • Step 2: Testes E2E → Aguarda: Step 1                    ║
║   • Step 3: Verificação completa → Aguarda: Step 2          ║
║                                                              ║
║ ✅ VERIFICAÇÃO:                                              ║
║   • Backend build + migration                                ║
║   • Frontend build                                           ║
║   • Testes passam                                            ║
╠══════════════════════════════════════════════════════════════╣
║ Total: 6 subtasks | Paralelas: 3 | Sequenciais: 3           ║
╚══════════════════════════════════════════════════════════════╝
```

### 📝 TodoWrite ativado automaticamente:

```
✅ TodoWrite criado com 6 subtasks:

1. ⏳ Track A: Criar schema de comentários (migrations)
2. ⏳ Track B: Explorar e planejar API de comentários
3. ⏳ Track C: Explorar padrões de UI para comentários
4. ⏱️ Step 1: Integração frontend ↔ backend
5. ⏱️ Step 2: Testes E2E
6. ⏱️ Step 3: Verificação final
```

### 🚀 Execução Automática:

**1. Lançando 3 subagents em paralelo:**

```markdown
[Task tool call] Subagent: general-purpose
Tarefa: Criar migration para tabela comments com foreign key para memories

[Task tool call] Subagent: Explore (medium thoroughness)
Tarefa: Mapear arquivos memory_handler, memory_service, memory_repository

[Task tool call] Subagent: Explore (quick)
Tarefa: Buscar componentes de comentários ou padrões similares no frontend
```

**2. Aguardando todos os subagents...**

---

**3. Resultados das Tracks Paralelas:**

#### ✅ Track A: Schema completo
```sql
-- migrations/033_add_comments.sql
CREATE TABLE comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  memory_id UUID NOT NULL REFERENCES memories(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id),
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_comments_memory ON comments(memory_id);
CREATE INDEX idx_comments_user ON comments(user_id);
```

#### ✅ Track B: Arquivos identificados
```
Padrão identificado: Handler → Service → Repository

Arquivos a criar/modificar:
- internal/handler/comment_handler.go (novo)
- internal/service/comment_service.go (novo)
- internal/repository/comment_repository.go (novo)
- internal/handler/memory_handler.go (adicionar endpoint list comments)
```

#### ✅ Track C: UI Pattern encontrado
```
Componente similar: frontend/src/lib/components/ChatMessage.svelte
Pode ser adaptado para CommentCard.svelte

Localização para nova UI:
frontend/src/routes/memories/[id]/+page.svelte (adicionar seção de comments)
```

---

**4. TodoWrite atualizado:**
```
1. ✅ Track A: Criar schema de comentários
2. ✅ Track B: Explorar e planejar API de comentários
3. ✅ Track C: Explorar padrões de UI para comentários
4. ⏳ Step 1: Integração frontend ↔ backend (em andamento)
5. ⏱️ Step 2: Testes E2E
6. ⏱️ Step 3: Verificação final
```

---

**5. Execução Sequencial:**

#### Step 1: Implementando Backend + Frontend

[Código do backend: handlers, services, repositories]
[Código do frontend: CommentCard.svelte, integração na página]

#### Step 2: Testes E2E

```go
// internal/handler/comment_handler_test.go
func TestCreateComment(t *testing.T) { ... }
func TestListComments(t *testing.T) { ... }
func TestDeleteComment(t *testing.T) { ... }
```

```typescript
// frontend/src/lib/components/CommentCard.test.ts
describe('CommentCard', () => {
  it('renders comment', ...)
  it('allows delete if owner', ...)
})
```

---

### ✅ VERIFICAÇÃO FINAL

**Backend Build:**
```bash
$ cd desktop/backend-go && go build
# ✓ Build successful
```

**Frontend Build:**
```bash
$ cd frontend && npm run build
✓ built in 2.1s
```

**Migration:**
```bash
$ psql -h localhost businessos < migrations/033_add_comments.sql
CREATE TABLE
CREATE INDEX
CREATE INDEX
```

**Testes Backend:**
```bash
$ go test ./internal/handler/... ./internal/service/...
ok  	handler	0.145s
ok  	service	0.089s
PASS
```

**Testes Frontend:**
```bash
$ npm test
✓ CommentCard renders
✓ Can add comment
✓ Can delete own comment
✓ Cannot delete others' comments
4/4 passed
```

**Teste E2E Manual:**
```
✓ Navegar para memória
✓ Adicionar comentário → aparece na lista
✓ Deletar comentário → remove da lista
✓ Comentários persistem após reload
```

---

### 🎉 Feature Completa

**TaskManager concluiu 6 subtasks automaticamente:**
- ✅ 3 tracks paralelas executadas
- ✅ 3 steps sequenciais completados
- ✅ Todos builds passaram
- ✅ Todos testes passaram
- ✅ Verificação E2E completa

**Arquivos criados/modificados:**
```
desktop/backend-go/
  migrations/033_add_comments.sql
  internal/handler/comment_handler.go
  internal/handler/comment_handler_test.go
  internal/service/comment_service.go
  internal/repository/comment_repository.go
  internal/handler/memory_handler.go (modificado)

frontend/
  src/lib/components/CommentCard.svelte
  src/lib/components/CommentCard.test.ts
  src/routes/memories/[id]/+page.svelte (modificado)
  src/routes/memories/[id]/+page.server.ts (modificado)
```

---

## 🐛 Exemplo 2: Bug Fix

**Usuário pede:**
> "As memórias não estão sendo filtradas corretamente por workspace"

**TaskManager detecta:**
- Tipo: Bug Fix
- Complexidade: Simples-Moderada
- Ativação: ✅ Automática

---

```
╔══════════════════════════════════════════════════════════════╗
║ 🤖 TASKMANAGER ATIVO                                         ║
╠══════════════════════════════════════════════════════════════╣
║ Tarefa: Fix filtro de memórias por workspace                ║
║ Tipo: Bug Fix                                                ║
║ Complexidade: Simples-Moderada                               ║
╠══════════════════════════════════════════════════════════════╣
║ 📊 DECOMPOSIÇÃO AUTOMÁTICA                                   ║
║                                                              ║
║ 🔄 PARALELO (3 tracks de investigação):                      ║
║   • Track A: Backend query → Subagent: Explore              ║
║   • Track B: Frontend request → Subagent: Explore            ║
║   • Track C: Database logs → Subagent: general-purpose      ║
║                                                              ║
║ ➡️ SEQUENCIAL (3 steps):                                     ║
║   • Step 1: Identificar root cause → Aguarda: A,B,C         ║
║   • Step 2: Implementar fix → Aguarda: Step 1               ║
║   • Step 3: Testar e verificar → Aguarda: Step 2            ║
║                                                              ║
║ ✅ VERIFICAÇÃO:                                              ║
║   • Bug reproduzido e corrigido                              ║
║   • Testes de regressão adicionados                          ║
║   • Verificação E2E                                          ║
╠══════════════════════════════════════════════════════════════╣
║ Total: 6 subtasks | Paralelas: 3 | Sequenciais: 3           ║
╚══════════════════════════════════════════════════════════════╝
```

### 📝 TodoWrite:
```
1. ⏳ Track A: Explorar query de memórias no backend
2. ⏳ Track B: Explorar chamada API no frontend
3. ⏳ Track C: Verificar logs de consultas SQL
4. ⏱️ Step 1: Identificar root cause
5. ⏱️ Step 2: Implementar fix
6. ⏱️ Step 3: Testar e verificar
```

### 🔍 Investigação Paralela:

**Track A: Backend Query**
```go
// Found in: internal/repository/memory_repository.go:145
func (r *MemoryRepository) GetMemoriesByWorkspace(ctx context.Context, workspaceID string) {
    // ❌ BUG: Não está usando workspaceID no WHERE!
    query := `SELECT * FROM memories ORDER BY created_at DESC`
    // Should be: WHERE workspace_id = $1
}
```

**Track B: Frontend**
```typescript
// frontend/src/routes/workspace/[id]/+page.server.ts
// ✓ Frontend está passando workspace_id corretamente
const memories = await api.getMemories({ workspaceId })
```

**Track C: Database Logs**
```sql
-- Query executada (do log):
SELECT * FROM memories ORDER BY created_at DESC;
-- ❌ Confirmado: workspace_id não está no WHERE clause
```

### 🎯 Root Cause Identificado:

```
Bug encontrado: memory_repository.go linha 145
Causa: Query SQL não filtra por workspace_id
Fix necessário: Adicionar WHERE workspace_id = $1
```

### 🔧 Fix Implementado:

```go
// internal/repository/memory_repository.go
func (r *MemoryRepository) GetMemoriesByWorkspace(ctx context.Context, workspaceID string) ([]*models.Memory, error) {
    query := `
        SELECT id, workspace_id, content, created_at, updated_at
        FROM memories
        WHERE workspace_id = $1
        ORDER BY created_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, workspaceID)
    // ... rest of implementation
}
```

### ✅ Verificação:

**Teste de Regressão Adicionado:**
```go
func TestGetMemoriesByWorkspace_FiltersCorrectly(t *testing.T) {
    // Criar memórias em workspaces diferentes
    ws1 := createTestWorkspace(t, "ws1")
    ws2 := createTestWorkspace(t, "ws2")

    createTestMemory(t, ws1.ID, "memory 1")
    createTestMemory(t, ws2.ID, "memory 2")

    // Buscar memórias do ws1
    memories, err := repo.GetMemoriesByWorkspace(ctx, ws1.ID)

    assert.NoError(t, err)
    assert.Len(t, memories, 1)
    assert.Equal(t, "memory 1", memories[0].Content)
}
```

**Teste Manual:**
```
✓ Workspace 1: mostra apenas memórias do workspace 1
✓ Workspace 2: mostra apenas memórias do workspace 2
✓ Sem vazamento entre workspaces
```

### 🎉 Bug Corrigido

**TaskManager completou:**
- ✅ Investigação paralela identificou bug
- ✅ Root cause encontrado (1 linha de código)
- ✅ Fix implementado e testado
- ✅ Teste de regressão adicionado
- ✅ Verificação E2E completa

---

## 🔄 Exemplo 3: Refactor

**Usuário pede:**
> "Refatorar sistema de logging para usar slog em vez de fmt.Printf"

**TaskManager detecta:**
- Tipo: Refactor
- Complexidade: Moderada
- Ativação: ✅ Automática

---

```
╔══════════════════════════════════════════════════════════════╗
║ 🤖 TASKMANAGER ATIVO                                         ║
╠══════════════════════════════════════════════════════════════╣
║ Tarefa: Migrar logging para slog                            ║
║ Tipo: Refactor                                               ║
║ Complexidade: Moderada                                       ║
╠══════════════════════════════════════════════════════════════╣
║ 📊 DECOMPOSIÇÃO AUTOMÁTICA                                   ║
║                                                              ║
║ 🔄 PARALELO (3 tracks de análise):                           ║
║   • Track A: Mapear fmt.Printf → Subagent: Explore          ║
║   • Track B: Analisar padrão slog → Subagent: Explore       ║
║   • Track C: Planejar migração → Subagent: Plan             ║
║                                                              ║
║ ➡️ SEQUENCIAL (5 steps):                                     ║
║   • Step 1: Setup slog global → Aguarda: A,B,C              ║
║   • Step 2: Migrar handlers → Aguarda: Step 1               ║
║   • Step 3: Migrar services → Aguarda: Step 2               ║
║   • Step 4: Migrar repositories → Aguarda: Step 3           ║
║   • Step 5: Remover fmt imports → Aguarda: Step 4           ║
║                                                              ║
║ ✅ VERIFICAÇÃO:                                              ║
║   • Todos testes passam                                      ║
║   • Sem fmt.Printf restante                                  ║
║   • Logs estruturados funcionando                            ║
╠══════════════════════════════════════════════════════════════╣
║ Total: 8 subtasks | Paralelas: 3 | Sequenciais: 5           ║
╚══════════════════════════════════════════════════════════════╝
```

### 📝 TodoWrite:
```
1. ⏳ Track A: Mapear todos os fmt.Printf
2. ⏳ Track B: Analisar padrão slog existente
3. ⏳ Track C: Planejar estratégia de migração
4. ⏱️ Step 1: Setup slog global logger
5. ⏱️ Step 2: Migrar handlers
6. ⏱️ Step 3: Migrar services
7. ⏱️ Step 4: Migrar repositories
8. ⏱️ Step 5: Cleanup e verificação
```

### 🔍 Análise Paralela:

**Track A: fmt.Printf mapeado**
```
Encontrados 47 usos de fmt.Printf/Println em:
- internal/handler/*.go: 18 ocorrências
- internal/service/*.go: 15 ocorrências
- internal/repository/*.go: 12 ocorrências
- cmd/main.go: 2 ocorrências
```

**Track B: Padrão slog existente**
```go
// Encontrado em: cmd/main.go (já tem setup parcial)
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
slog.SetDefault(logger)

// Padrão identificado:
slog.Info("message", "key", value)
slog.Error("message", "error", err)
```

**Track C: Plano de migração**
```
Estratégia:
1. Garantir slog.SetDefault() em main.go
2. Passar logger via context ou struct fields
3. Migrar camada por camada (handler → service → repo)
4. Pattern: fmt.Printf("msg %v", val) → slog.Info("msg", "detail", val)
5. Errors: fmt.Printf("error: %v", err) → slog.Error("msg", "error", err)
```

---

### 🔧 Execução Sequencial:

[Implementação detalhada de cada step...]

### ✅ Verificação Final:

```bash
# Verificar se ainda tem fmt.Printf
$ grep -r "fmt.Printf" internal/
# (sem resultados)

# Verificar logs estruturados funcionando
$ go run cmd/main.go
{"time":"2026-01-06T10:30:00","level":"INFO","msg":"Server starting","port":8080}
{"time":"2026-01-06T10:30:01","level":"INFO","msg":"Database connected"}

# Testes
$ go test ./...
ok  	all tests passed
```

### 🎉 Refactor Completo

**TaskManager executou:**
- ✅ Análise paralela (47 ocorrências mapeadas)
- ✅ Plano de migração criado
- ✅ 5 steps sequenciais executados
- ✅ Todos testes passam
- ✅ 0 fmt.Printf restantes
- ✅ Logs estruturados ativos

---

## 📊 Resumo dos Exemplos

### Padrão Consistente:

1. **TaskManager detecta tipo de tarefa**
2. **Decompõe automaticamente**
3. **Cria TodoWrite**
4. **Lança subagents paralelos**
5. **Executa steps sequenciais**
6. **Verifica com outputs reais**

### Benefícios:

- ✅ **Visibilidade**: Usuário vê todo o processo
- ✅ **Eficiência**: Paralelização automática
- ✅ **Qualidade**: Verificação obrigatória
- ✅ **Rastreamento**: TodoWrite sempre ativo
- ✅ **Consistência**: Mesmo padrão para qualquer tarefa

---

**Estes exemplos mostram o TaskManager trabalhando automaticamente em diferentes cenários.**
