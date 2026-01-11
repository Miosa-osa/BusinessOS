# BusinessOS - Claude Code Workflow

## 🎯 Filosofia: Decomposição Automática em Subtasks

**SEMPRE divida trabalho complexo em subtasks paralelas usando subagents nativos do Claude Code.**

---

## 🤖 TaskManager Automático

### Sistema de Auto-Decomposição

O **TaskManager** é ativado AUTOMATICAMENTE para TODA tarefa não-trivial.

### Como Funciona:

```
1. Usuário pede algo
2. TaskManager analisa e classifica a tarefa
3. TaskManager decomponha AUTOMATICAMENTE em subtasks
4. TaskManager lança subagents em paralelo
5. TaskManager rastreia com TodoWrite
6. TaskManager verifica e reporta resultados
```

### Análise Automática de Tarefas:

| Tipo Detectado | Decomposição Automática | Subagents Usados |
|----------------|-------------------------|------------------|
| **Full-Stack Feature** | Frontend + Backend + DB em paralelo | Explore (x2) + Plan |
| **Bug Fix** | Investigação paralela → Fix → Verify | Explore (x3) + general-purpose |
| **Refactor** | Análise paralela → Plan → Executar | Explore + Plan + general-purpose |
| **Frontend Only** | Explorar padrões → Implementar → Testar | Explore + general-purpose |
| **Backend Only** | Explorar código → Planejar → Implementar | Explore + Plan |
| **Database** | Explorar schema → Migration → Validar | Explore + general-purpose |

### Ativação Automática:

TaskManager ativa quando detecta:
- ✅ Palavras-chave: "adicionar", "criar", "implementar", "fix", "refactor"
- ✅ Múltiplos arquivos/módulos envolvidos
- ✅ Frontend E Backend juntos
- ✅ Mudanças em banco de dados
- ✅ Tarefa com mais de 2 passos

TaskManager NÃO ativa para:
- ❌ Perguntas simples ("o que é X?")
- ❌ Leitura de código ("mostre o arquivo X")
- ❌ Comandos diretos de 1 passo ("compile o projeto")

### 🚀 TaskManager Avançado: Microtasks, Milestones e Feedback Loop

Para **tarefas complexas** (3+ dias de trabalho), o TaskManager usa modo avançado:

**Hierarquia Multinível:**
```
MILESTONE (Marco principal - ex: "Q1 Implementation")
    ↓
TASKS (Tarefas principais - ex: "Auth System")
    ↓
SUBTASKS (Tracks paralelas/sequenciais)
    ↓
MICROTASKS (Unidades atômicas mínimas)
    ↓
FEEDBACK LOOP (Métricas + Aprendizado contínuo)
```

**Quando ativa:**
- ✅ Features que levam 3+ dias
- ✅ Sistemas completos (auth, notifications, etc)
- ✅ Múltiplas tarefas inter-relacionadas
- ✅ Quando usuário menciona "milestone" ou "fase"

**Benefícios:**
- 📊 Dashboard de progresso em tempo real
- 🎯 Estimativas que melhoram continuamente
- 🔁 Feedback automático após cada microtask
- 🤖 Distribuição inteligente de agentes
- 📈 Aprendizado: Sistema fica mais eficiente com uso

**Documentação completa:** `docs/ADVANCED_TASKMANAGER.md`

---

## 📋 Quando Usar Decomposição

| Situação | Ação | Ferramenta |
|----------|------|------------|
| 3+ passos independentes | Múltiplos subagents em paralelo | `Task` tool |
| Frontend + Backend juntos | 2 tracks paralelas | `Task` tool |
| Pesquisa de código | Agente Explore | `Task` (Explore) |
| Planejamento complexo | Agente Plan | `Task` (Plan) |
| Implementação | Agente general-purpose | `Task` (general-purpose) |
| Rastreamento visual | Criar todos | `TodoWrite` |
| Buscar decisões passadas | Memória episódica | `Skill` (episodic-memory) |

---

## 🚀 Workflow Padrão para TODA Tarefa

### 1️⃣ INICIALIZAÇÃO (primeira resposta)

```markdown
1. Ler TASKS.md para entender contexto atual
2. Buscar memória episódica se relevante
3. Mostrar status do projeto
4. DECOMPOR em subtasks antes de começar
5. Usar TodoWrite para rastrear subtasks
```

### 2️⃣ PLANEJAMENTO: Decomposição em Subtasks

Para QUALQUER tarefa não-trivial:

```markdown
## 📊 Plano de Execução

### 🔄 Paralelo (executam juntos)
- [ ] TRACK A: [tarefa] → Subagent: Explore/Plan/general-purpose
- [ ] TRACK B: [tarefa] → Subagent: Explore/Plan/general-purpose

### ➡️ Sequencial (depois das paralelas)
- [ ] STEP 1: [tarefa] → Aguarda: A, B completarem
- [ ] STEP 2: [tarefa] → Aguarda: Step 1

### ✅ Verificação
- [ ] Build/compile funciona
- [ ] Testes passam
- [ ] Sem regressões
```

**OBRIGATÓRIO: Usar TodoWrite imediatamente após decomposição.**

### 3️⃣ EXECUÇÃO: Lançar Subagents

**Exemplo de execução paralela:**

```markdown
Vou executar em paralelo:
1. 🔍 Explore: Mapear arquivos de autenticação
2. 📋 Plan: Planejar implementação da feature
3. 🔨 general-purpose: Buscar padrões similares no código
```

**Lançar múltiplos agentes em UMA ÚNICA mensagem:**
- Use múltiplos `Task` tool calls na mesma resposta
- Claude Code executa todos em paralelo
- Aguarde TODOS terminarem antes de prosseguir

### 3️⃣-B FORMATO VISUAL DO TASKMANAGER

Quando TaskManager ativa, SEMPRE mostrar este cabeçalho:

```
╔══════════════════════════════════════════════════════════════╗
║ 🤖 TASKMANAGER ATIVO                                         ║
╠══════════════════════════════════════════════════════════════╣
║ Tarefa: [descrição breve]                                    ║
║ Tipo: [Full-Stack/Frontend/Backend/Bug/Refactor/DB]         ║
║ Complexidade: [Simples/Moderada/Complexa/Crítica]           ║
╠══════════════════════════════════════════════════════════════╣
║ 📊 DECOMPOSIÇÃO AUTOMÁTICA                                   ║
║                                                              ║
║ 🔄 PARALELO (X tracks):                                      ║
║   • Track A: [descrição] → Subagent: [tipo]                 ║
║   • Track B: [descrição] → Subagent: [tipo]                 ║
║   • Track C: [descrição] → Subagent: [tipo]                 ║
║                                                              ║
║ ➡️ SEQUENCIAL (Y steps):                                     ║
║   • Step 1: [descrição] → Aguarda: [dependências]           ║
║   • Step 2: [descrição] → Aguarda: [dependências]           ║
║                                                              ║
║ ✅ VERIFICAÇÃO:                                              ║
║   • Build/compile                                            ║
║   • Testes                                                   ║
║   • Sem regressões                                           ║
╠══════════════════════════════════════════════════════════════╣
║ Total: X subtasks | Paralelas: Y | Sequenciais: Z           ║
╚══════════════════════════════════════════════════════════════╝
```

**Após decomposição:**
1. Usar TodoWrite com TODAS as subtasks
2. Lançar subagents paralelos imediatamente
3. Atualizar TodoWrite conforme progresso
4. Mostrar resultados de cada track
5. Executar steps sequenciais
6. Verificação final com outputs

### 4️⃣ VERIFICAÇÃO OBRIGATÓRIA

Antes de dizer "pronto/concluído/done":

```markdown
## ✅ VERIFICAÇÃO

□ Código compila sem erros
□ Testes passam (mostrar output)
□ Sem regressões
□ Edge cases tratados

**PROVA:**
[mostrar output de comandos: build, test, etc]
```

---

## 🛠️ Subagents Nativos do Claude Code

### Explore
- **Quando**: Encontrar arquivos, mapear código, buscar padrões
- **Throughness**: "quick", "medium", "very thorough"
- **Exemplo**: "Find all authentication-related files"

### Plan
- **Quando**: Planejar implementação antes de escrever código
- **Output**: Plano passo-a-passo, arquivos críticos, trade-offs
- **Exemplo**: "Plan implementation of notifications system"

### general-purpose
- **Quando**: Tarefas complexas, multi-step, pesquisa + implementação
- **Exemplo**: "Research and implement caching layer"

---

## 📚 Padrões BusinessOS

### Stack Atual
```
Frontend:  SvelteKit + TypeScript + Tailwind + Svelte 5
Backend:   Go 1.24.1 + Gin + PostgreSQL + Redis + pgvector
Deploy:    Docker + GCP Cloud Run
```

### Convenções de Código

**Svelte/SvelteKit:**
- Stores para estado compartilhado
- Form actions para mutations
- `+page.server.ts` para data loading
- `+page.ts` para client-side data

**Go Backend:**
- **SEMPRE use `slog`** para logging (NÃO `fmt.Printf`)
- Context propagation em todas funções
- Handler → Service → Repository (camadas)
- Error handling adequado (no `panic`)
- Graceful shutdown

**Estrutura de Pastas:**
```
desktop/backend-go/
  ├── scripts/
  │   ├── debug/      # Scripts de debug/diagnóstico
  │   ├── migrations/ # Scripts de migração DB
  │   └── tests/      # Scripts de teste
  ├── internal/
  │   ├── handler/
  │   ├── service/
  │   └── repository/
  └── ...

frontend/
  ├── src/
  │   ├── routes/
  │   ├── lib/
  │   └── components/
  └── ...
```

### Arquitetura

**Backend Flow:**
```
HTTP Request → Handler → Service → Repository → Database
                  ↓         ↓          ↓
               Validation  Logic   Data Access
```

**Features Principais:**
- Memory hierarchy system (workspace/project/agent)
- Role-based agent behavior
- RAG with pgvector embeddings
- SSE streaming
- COT (Chain of Thought) orchestration

---

## 🎯 Exemplos Práticos

### Exemplo 1: Nova Feature (Frontend + Backend)

**Request do usuário:**
> "Adicionar notificações em tempo real"

**Sua resposta:**

```markdown
## 📊 Plano de Execução

Vou decompor esta feature em subtasks paralelas:

### 🔄 Paralelo
- [ ] TRACK A: Backend - Criar endpoints de notificações
  - Subagent: Plan → planejar API
  - Files: internal/handler/notifications.go, service, repository

- [ ] TRACK B: Frontend - UI de notificações
  - Subagent: Explore → encontrar padrões de UI existentes
  - Files: src/routes/notifications/+page.svelte

- [ ] TRACK C: Database - Schema de notificações
  - Subagent: general-purpose → criar migration
  - Files: migrations/XXX_notifications.sql

### ➡️ Sequencial
- [ ] Integração Frontend ↔ Backend (SSE streaming)
- [ ] Testes E2E
- [ ] Verificação completa

Vou usar TodoWrite para rastrear e começar com Track A, B, C em paralelo.
```

### Exemplo 2: Bug Fix

**Request:**
> "Login está mostrando tela branca"

**Sua resposta:**

```markdown
## 🔍 Investigação + Fix

### 🔄 Paralelo (investigação)
- [ ] TRACK A: Explore frontend login files
  - Buscar: src/routes/login, auth components

- [ ] TRACK B: Explore backend auth handlers
  - Buscar: internal/handler/auth.go

- [ ] TRACK C: Verificar logs recentes
  - Buscar erros de console, logs do servidor

Após investigação paralela → identificar causa → implementar fix → verificar.
```

### Exemplo 3: Refactor

**Request:**
> "Limpar código de autenticação"

**Sua resposta:**

```markdown
## 🔄 Refactor Sistemático

### 🔄 Paralelo (análise)
- [ ] Explore: Mapear TODOS arquivos de auth
- [ ] general-purpose: Identificar código duplicado
- [ ] Plan: Planejar nova estrutura

### ➡️ Sequencial (execução)
- [ ] Criar abstrações comuns
- [ ] Migrar código gradualmente
- [ ] Testes após cada migração
- [ ] Remover código antigo

### ✅ Verificação
- [ ] Todos testes de auth passam
- [ ] Sem regressões
- [ ] Cobertura mantida/melhorada
```

---

## 🚨 Regras Não Negociáveis

### ❌ NUNCA faça isso:
1. Começar a implementar SEM decomposição prévia
2. Dizer "pronto" sem mostrar output de build/tests
3. Modificar código sem ler o arquivo primeiro
4. Usar `fmt.Printf` no backend Go (use `slog`)
5. Criar código sem tratar erros adequadamente
6. Fazer múltiplas tarefas independentes sequencialmente (use paralelo!)

### ✅ SEMPRE faça isso:
1. Decompor tarefa complexa antes de começar
2. Usar TodoWrite para rastrear subtasks
3. Lançar subagents em paralelo quando possível
4. Mostrar verificação com output real
5. Ler arquivos antes de modificar
6. Seguir padrões do projeto (Handler→Service→Repository)

---

## 📝 Template de Resposta

Use este template para TODA tarefa não-trivial:

```markdown
## 📊 Plano de Execução

[Decomposição clara da tarefa]

### 🔄 Subtasks Paralelas
- [ ] Track A: ...
- [ ] Track B: ...

### ➡️ Subtasks Sequenciais
- [ ] Step 1: ...
- [ ] Step 2: ...

### ✅ Verificação
- [ ] Build
- [ ] Tests
- [ ] No regressions

---

Iniciando execução com TodoWrite...

[Lançar subagents conforme necessário]

[Implementação]

[Verificação com output real]
```

---

## 🎓 Memória Episódica

Use o plugin `episodic-memory` para:
- Buscar decisões arquiteturais passadas
- Encontrar soluções para problemas similares
- Recuperar padrões de código já aprovados

**Quando usar:**
- Início de sessão (se relevante)
- Antes de decisões arquiteturais
- Quando encontrar problema já resolvido

---

**Última atualização:** Q1 2026 - v2.1.0
**Projeto:** BusinessOS (SvelteKit + Go)
**Branch principal:** main
**Branch de desenvolvimento:** pedro-dev
