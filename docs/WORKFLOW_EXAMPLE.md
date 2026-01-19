# Exemplo Prático: Workflow com Subtasks e Subagents

Este documento mostra um exemplo real de como o Claude Code deve responder seguindo o CLAUDE.md.

---

## Cenário: Usuário pede "Adicionar sistema de tags para memórias"

---

## ✅ Resposta Correta (Seguindo CLAUDE.md)

### 📊 Plano de Execução

Vou decompor esta feature em subtasks paralelas e sequenciais:

### 🔄 Paralelo (podem executar simultaneamente)

- [ ] **TRACK A: Database Schema**
  - Subagent: `general-purpose`
  - Tarefa: Criar migration para tabela `memory_tags` e relação many-to-many
  - Files: `desktop/backend-go/migrations/XXX_add_memory_tags.sql`

- [ ] **TRACK B: Backend - Explorar código existente**
  - Subagent: `Explore` (thoroughness: medium)
  - Tarefa: Mapear onde memórias são criadas/buscadas
  - Buscar: `internal/handler/memory*.go`, `internal/service/memory*.go`

- [ ] **TRACK C: Frontend - Padrões de UI**
  - Subagent: `Explore` (thoroughness: quick)
  - Tarefa: Encontrar componentes de tag input existentes
  - Buscar: `frontend/src/lib/components/**/*tag*.svelte`

### ➡️ Sequencial (após paralelo completar)

- [ ] **STEP 1: Backend API**
  - Aguarda: Track A (schema), Track B (padrões)
  - Implementar: Handler → Service → Repository para tags
  - Files: `internal/handler/tags.go`, `internal/service/tags.go`, `internal/repository/tags.go`

- [ ] **STEP 2: Frontend Integration**
  - Aguarda: Step 1 (API pronta), Track C (UI patterns)
  - Implementar: Tag input component + integração com API
  - Files: `frontend/src/lib/components/TagInput.svelte`, `frontend/src/routes/memories/+page.svelte`

- [ ] **STEP 3: Testes**
  - Aguarda: Step 1, Step 2
  - Unit tests backend + Component tests frontend

### ✅ Verificação Final

- [ ] Backend compila: `cd desktop/backend-go && go build`
- [ ] Frontend compila: `cd frontend && npm run build`
- [ ] Testes passam: Backend + Frontend
- [ ] Migration aplica sem erros
- [ ] Feature funciona end-to-end

---

### 🚀 Execução

Agora vou usar **TodoWrite** para rastrear e lançar as **3 tracks paralelas**:

```markdown
TodoWrite:
1. ⏳ TRACK A: Criar schema de tags
2. ⏳ TRACK B: Explorar código de memórias
3. ⏳ TRACK C: Buscar padrões de UI tags
4. ⏱️ STEP 1: Implementar backend API (aguarda A, B)
5. ⏱️ STEP 2: Implementar frontend (aguarda 1, C)
6. ⏱️ STEP 3: Testes (aguarda 2)
7. ⏱️ Verificação final
```

**Lançando subagents em paralelo agora...**

[Aqui o Claude faria múltiplos Task tool calls em UMA mensagem:]

```
Task(Explore, "Mapear código de memórias", ...)
Task(Explore, "Buscar componentes de tags", ...)
Task(general-purpose, "Criar migration de tags", ...)
```

[Aguarda TODOS terminarem]

---

### Resultados das Tracks Paralelas

#### Track A: Schema Created ✅
```sql
-- migrations/032_add_memory_tags.sql
CREATE TABLE tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE memory_tags (
  memory_id UUID REFERENCES memories(id) ON DELETE CASCADE,
  tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (memory_id, tag_id)
);

CREATE INDEX idx_memory_tags_memory ON memory_tags(memory_id);
CREATE INDEX idx_memory_tags_tag ON memory_tags(tag_id);
```

#### Track B: Código Mapeado ✅
Encontrados:
- `internal/handler/memory_handler.go` - CreateMemory, GetMemories
- `internal/service/memory_service.go` - Business logic
- `internal/repository/memory_repository.go` - Data access

Padrão identificado: Handler → Service → Repository

#### Track C: UI Patterns ✅
Não encontrado componente de tags existente.
Encontrado pattern similar em: `frontend/src/lib/components/SearchInput.svelte`
Posso adaptar este padrão.

---

### Agora: Execução Sequencial

**Atualizando TodoWrite:**
```
1. ✅ TRACK A: Criar schema de tags
2. ✅ TRACK B: Explorar código de memórias
3. ✅ TRACK C: Buscar padrões de UI tags
4. ⏳ STEP 1: Implementar backend API (em andamento)
...
```

**STEP 1: Implementando Backend API...**

[Implementação do código seguindo padrões Handler→Service→Repository]

**STEP 2: Implementando Frontend...**

[Implementação da UI de tags]

**STEP 3: Testes...**

[Criação e execução de testes]

---

## ✅ VERIFICAÇÃO FINAL

### Build Backend
```bash
$ cd desktop/backend-go && go build
# Output: sem erros
```

### Build Frontend
```bash
$ cd frontend && npm run build
✓ built in 2.3s
```

### Testes Backend
```bash
$ cd desktop/backend-go && go test ./internal/...
ok  	github.com/rhl/businessos-backend/internal/handler	0.123s
ok  	github.com/rhl/businessos-backend/internal/service	0.089s
ok  	github.com/rhl/businessos-backend/internal/repository	0.156s
```

### Testes Frontend
```bash
$ cd frontend && npm test
✓ TagInput component renders
✓ TagInput allows adding tags
✓ TagInput allows removing tags
✓ Memory page integrates with tags API

Tests passed: 4/4
```

### Migration
```bash
$ psql -h localhost -U postgres -d businessos -f migrations/032_add_memory_tags.sql
CREATE TABLE
CREATE TABLE
CREATE INDEX
CREATE INDEX
```

### Teste Manual E2E
1. ✅ Criar memória com tags
2. ✅ Tags aparecem na lista
3. ✅ Filtrar por tag funciona
4. ✅ Remover tag funciona

---

## 🎉 Feature Completa

**Resultado:**
- ✅ 3 subtasks paralelas executadas
- ✅ 3 steps sequenciais completados
- ✅ Todos builds passam
- ✅ Todos testes passam
- ✅ Feature funciona end-to-end

**Arquivos modificados:**
```
desktop/backend-go/
  migrations/032_add_memory_tags.sql
  internal/handler/tags.go (novo)
  internal/service/tags.go (novo)
  internal/repository/tags.go (novo)
  internal/handler/memory_handler_test.go

frontend/
  src/lib/components/TagInput.svelte (novo)
  src/routes/memories/+page.svelte (modificado)
  src/routes/memories/+page.server.ts (modificado)
```

**Pronto para commit/PR.**

---

## ❌ Resposta INCORRETA (NÃO fazer)

### Exemplo de resposta ruim:

> "Vou adicionar o sistema de tags. Primeiro vou criar a tabela de tags no banco..."
>
> [Começa a escrever código sem decomposição]
>
> [Depois diz] "Pronto, feature implementada!"
>
> [Sem mostrar build, sem testes, sem verificação]

### Por que está errado:
1. ❌ Não decompôs em subtasks
2. ❌ Não usou TodoWrite
3. ❌ Não lançou subagents
4. ❌ Não executou paralelo
5. ❌ Disse "pronto" sem verificação
6. ❌ Não mostrou output de build/tests

---

## 📚 Lições do Exemplo

### ✅ Padrão Correto:
1. **Decompor ANTES de começar**
2. **Identificar paralelo vs sequencial**
3. **Usar TodoWrite para rastrear**
4. **Lançar subagents apropriados**
5. **Executar paralelo quando possível**
6. **SEMPRE verificar com output real**

### 🎯 Benefícios:
- **Visibilidade**: Usuário vê o progresso
- **Eficiência**: Tarefas paralelas economizam tempo
- **Qualidade**: Verificação obrigatória
- **Rastreamento**: TodoWrite mantém organização
- **Padrões**: Segue arquitetura do projeto

---

**Este é o workflow esperado para TODA tarefa não-trivial no BusinessOS.**
