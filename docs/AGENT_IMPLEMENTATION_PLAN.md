# Business OS - Agent Architecture Implementation Plan

**Data:** 2025-12-26  
**Status:** Em Progresso  
**Versão:** 1.0

---

## Visão Geral

Este documento detalha o plano de implementação da nova arquitetura de agentes do Business OS, transformando o sistema atual de 4 agentes básicos em um sistema multi-agente robusto com 6 agentes especializados e routing inteligente.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     BUSINESS OS AGENT ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                              USER MESSAGE                                    │
│                                   │                                         │
│                                   ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      ORCHESTRATOR AGENT                              │   │
│  │                                                                      │   │
│  │  • Primary interface for all user interactions                       │   │
│  │  • Handles general questions, advice, discussions                    │   │
│  │  • Routes to specialists when deep domain work needed                │   │
│  └─────────────────────────────────┬───────────────────────────────────┘   │
│                                    │                                        │
│          ┌─────────────┬───────────┼───────────┬─────────────┐             │
│          │             │           │           │             │             │
│          ▼             ▼           ▼           ▼             ▼             │
│  ┌─────────────┐┌─────────────┐┌─────────────┐┌─────────────┐┌──────────┐ │
│  │  DOCUMENT   ││   PROJECT   ││    TASK     ││   CLIENT    ││ ANALYST  │ │
│  │   AGENT     ││   AGENT     ││   AGENT     ││   AGENT     ││  AGENT   │ │
│  └─────────────┘└─────────────┘└─────────────┘└─────────────┘└──────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Estado Atual vs. Estado Desejado

### Estado Atual
```
internal/
├── agents/
│   └── agents.go              # 4 agentes básicos (Orchestrator, Document, Analysis, Planning)
├── prompts/
│   └── prompts.go             # Prompts monolíticos
├── tools/
│   ├── artifacts.go           # Criação de artifacts
│   └── context_tools.go       # get_entity_context
└── services/
    └── tiered_context.go      # Sistema de contexto em 3 níveis (já implementado)
```

### Estado Desejado
```
internal/
├── agents/
│   ├── types.go               # Interfaces e tipos compartilhados
│   ├── base.go                # BaseAgent com funcionalidade comum
│   ├── registry.go            # AgentRegistry - factory pattern
│   ├── events.go              # StreamEvent, EventType
│   │
│   ├── orchestrator/
│   │   ├── agent.go           # OrchestratorAgent
│   │   ├── router.go          # IntentRouter - classifica e roteia
│   │   └── prompt.go          # Prompt do Orchestrator
│   │
│   ├── document/
│   │   ├── agent.go           # DocumentAgent
│   │   └── prompt.go          # Prompt do Document Agent
│   │
│   ├── project/
│   │   ├── agent.go           # ProjectAgent (inclui Task)
│   │   └── prompt.go          # Prompt do Project/Task Agent
│   │
│   ├── client/
│   │   ├── agent.go           # ClientAgent
│   │   └── prompt.go          # Prompt do Client Agent
│   │
│   └── analyst/
│       ├── agent.go           # AnalystAgent
│       └── prompt.go          # Prompt do Analyst Agent
│
├── prompts/
│   ├── core/
│   │   ├── identity.go        # Identidade OSA compartilhada
│   │   ├── formatting.go      # Regras de formatação
│   │   ├── artifacts.go       # Sistema de artifacts
│   │   ├── context.go         # Integração de contexto
│   │   ├── tools.go           # Instruções de uso de tools
│   │   └── errors.go          # Error handling
│   │
│   └── composer.go            # Monta prompts finais
│
├── tools/
│   ├── registry.go            # ToolRegistry
│   ├── interface.go           # Tool interface
│   ├── artifacts.go           # create_artifact, update_artifact (existente)
│   ├── knowledge.go           # search_knowledge, get_entity_context
│   ├── projects.go            # create_project, update_project
│   ├── tasks.go               # create_task, update_task, bulk_create
│   ├── clients.go             # create_client, update_client, log_interaction
│   ├── team.go                # get_team_capacity, assign_team_member
│   ├── analytics.go           # query_metrics, get_trends
│   └── activity.go            # log_activity, get_recent_activity
│
└── streaming/
    ├── parser.go              # Stream chunk parsing
    ├── artifact_detector.go   # Detecção de artifacts em stream
    └── events.go              # SSE event types
```

---

## Fases de Implementação

### Fase 1: Estrutura Base (Foundation)
**Estimativa:** 2-3 horas  
**Dependências:** Nenhuma

#### Arquivos a Criar:

1. **`internal/agents/types.go`**
   - Interface `Agent` expandida
   - `AgentType` enum
   - `AgentInput` struct
   - `UserSelections` struct
   - `ContextRequirements` struct

2. **`internal/agents/base.go`**
   - `BaseAgent` struct refatorado
   - Suporte a tools e context requirements
   - Método `Run()` com streaming

3. **`internal/agents/registry.go`**
   - `AgentRegistry` struct
   - Factory methods para todos os agentes
   - `GetAgentForFocusMode()` atualizado

4. **`internal/agents/events.go`**
   - `StreamEvent` struct
   - `EventType` enum (token, artifact_start, artifact_complete, tool_call, etc.)

#### Checklist:
- [ ] types.go criado
- [ ] base.go refatorado
- [ ] registry.go criado
- [ ] events.go criado
- [ ] Testes básicos passando

---

### Fase 2: Sistema de Prompts Modular
**Estimativa:** 2-3 horas  
**Dependências:** Fase 1

#### Arquivos a Criar:

1. **`internal/prompts/core/identity.go`**
   - `CoreIdentity` const
   - Identidade OSA compartilhada

2. **`internal/prompts/core/formatting.go`**
   - `OutputFormattingStandards` const
   - Regras de formatação de output

3. **`internal/prompts/core/artifacts.go`**
   - `ArtifactSystem` const
   - Instruções de criação de artifacts

4. **`internal/prompts/core/context.go`**
   - `ContextIntegration` const
   - Como usar o sistema de contexto em 3 níveis

5. **`internal/prompts/core/tools.go`**
   - `ToolUsagePatterns` const
   - Instruções de uso de tools

6. **`internal/prompts/core/errors.go`**
   - `ErrorHandling` const
   - Tratamento de edge cases

7. **`internal/prompts/composer.go`**
   - `ComposePrompt()` function
   - Monta prompt final com contexto dinâmico

#### Checklist:
- [ ] core/identity.go criado
- [ ] core/formatting.go criado
- [ ] core/artifacts.go criado
- [ ] core/context.go criado
- [ ] core/tools.go criado
- [ ] core/errors.go criado
- [ ] composer.go criado

---

### Fase 3: Orchestrator Agent
**Estimativa:** 3-4 horas  
**Dependências:** Fases 1, 2

#### Arquivos a Criar:

1. **`internal/agents/orchestrator/agent.go`**
   - `OrchestratorAgent` struct
   - Implementação de `Agent` interface
   - Lógica de delegação

2. **`internal/agents/orchestrator/router.go`**
   - `IntentRouter` struct
   - `ClassifyIntent()` method
   - Keywords e patterns para cada agente

3. **`internal/agents/orchestrator/prompt.go`**
   - `OrchestratorPrompt` const
   - Prompt específico do Orchestrator

#### Checklist:
- [ ] agent.go criado
- [ ] router.go criado com intent classification
- [ ] prompt.go criado
- [ ] Delegação funcionando

---

### Fase 4: Document Agent
**Estimativa:** 2 horas  
**Dependências:** Fases 1, 2

#### Arquivos a Criar:

1. **`internal/agents/document/agent.go`**
   - `DocumentAgent` struct
   - Tools: create_artifact, search_knowledge, get_entity_context

2. **`internal/agents/document/prompt.go`**
   - `DocumentAgentPrompt` const (prompt completo fornecido)

#### Checklist:
- [ ] agent.go criado
- [ ] prompt.go criado com prompt completo
- [ ] Criação de artifacts funcionando

---

### Fase 5: Project Agent (Task Agent)
**Estimativa:** 2-3 horas  
**Dependências:** Fases 1, 2

#### Arquivos a Criar:

1. **`internal/agents/project/agent.go`**
   - `ProjectAgent` struct
   - Tools: create_project, update_project, create_task, assign_team_member, etc.

2. **`internal/agents/project/prompt.go`**
   - `ProjectAgentPrompt` const (Planning Agent prompt fornecido)

#### Checklist:
- [ ] agent.go criado
- [ ] prompt.go criado
- [ ] Tools de projeto funcionando

---

### Fase 6: Client Agent
**Estimativa:** 2 horas  
**Dependências:** Fases 1, 2

#### Arquivos a Criar:

1. **`internal/agents/client/agent.go`**
   - `ClientAgent` struct
   - Tools: create_client, update_client, log_interaction, update_pipeline

2. **`internal/agents/client/prompt.go`**
   - `ClientAgentPrompt` const

#### Checklist:
- [ ] agent.go criado
- [ ] prompt.go criado
- [ ] Tools de cliente funcionando

---

### Fase 7: Analyst Agent
**Estimativa:** 2 horas  
**Dependências:** Fases 1, 2

#### Arquivos a Criar:

1. **`internal/agents/analyst/agent.go`**
   - `AnalystAgent` struct
   - Tools: query_metrics, get_trends, create_artifact, search_knowledge

2. **`internal/agents/analyst/prompt.go`**
   - `AnalystAgentPrompt` const (prompt completo fornecido)

#### Checklist:
- [ ] agent.go criado
- [ ] prompt.go criado com prompt completo
- [ ] Análise de dados funcionando

---

### Fase 8: Tool Registry Expandido
**Estimativa:** 3-4 horas  
**Dependências:** Fases 1-7

#### Arquivos a Criar/Atualizar:

1. **`internal/tools/interface.go`**
   - `Tool` interface padronizada
   - `ToolResult` struct

2. **`internal/tools/registry.go`**
   - `ToolRegistry` struct
   - Registro de tools por agente

3. **`internal/tools/projects.go`**
   - `CreateProjectTool`
   - `UpdateProjectTool`
   - `GetProjectStatusTool`

4. **`internal/tools/tasks.go`**
   - `CreateTaskTool`
   - `UpdateTaskTool`
   - `BulkCreateTasksTool`
   - `MoveTaskTool`
   - `AssignTaskTool`

5. **`internal/tools/clients.go`**
   - `CreateClientTool`
   - `UpdateClientTool`
   - `LogInteractionTool`
   - `UpdatePipelineTool`

6. **`internal/tools/team.go`**
   - `GetTeamCapacityTool`
   - `AssignTeamMemberTool`

7. **`internal/tools/analytics.go`**
   - `QueryMetricsTool`
   - `GetTrendsTool`

8. **`internal/tools/activity.go`**
   - `LogActivityTool`
   - `GetRecentActivityTool`

#### Checklist:
- [ ] interface.go criado
- [ ] registry.go criado
- [ ] Todas as tools implementadas
- [ ] Tools testadas

---

### Fase 9: Streaming/Artifact Detector
**Estimativa:** 2-3 horas  
**Dependências:** Fase 1

#### Arquivos a Criar:

1. **`internal/streaming/events.go`**
   - `EventType` enum
   - `StreamEvent` struct

2. **`internal/streaming/artifact_detector.go`**
   - `ArtifactDetector` struct
   - `ProcessChunk()` method
   - Detecção de artifacts em stream

3. **`internal/streaming/parser.go`**
   - Parsing de chunks
   - Handling de edge cases

#### Checklist:
- [ ] events.go criado
- [ ] artifact_detector.go criado
- [ ] parser.go criado
- [ ] Detecção em tempo real funcionando

---

### Fase 10: Integração com Chat Handler
**Estimativa:** 2-3 horas  
**Dependências:** Todas as fases anteriores

#### Arquivos a Atualizar:

1. **`internal/handlers/chat.go`**
   - Usar novo `AgentRegistry`
   - Integrar `ArtifactDetector`
   - SSE events tipados
   - Agent traces para logging

#### Checklist:
- [ ] chat.go atualizado
- [ ] SSE events funcionando
- [ ] Artifacts salvos automaticamente
- [ ] Logging de agent traces

---

## Matriz de Tools por Agente

| Tool | Orchestrator | Document | Project | Task | Client | Analyst |
|------|:------------:|:--------:|:-------:|:----:|:------:|:-------:|
| create_artifact | - | ✓ | ✓ | - | ✓ | ✓ |
| search_knowledge | ✓ | ✓ | ✓ | - | ✓ | ✓ |
| get_entity_context | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| create_project | - | - | ✓ | - | - | - |
| update_project | - | - | ✓ | - | - | - |
| create_task | ✓ | - | ✓ | ✓ | - | - |
| update_task | - | - | ✓ | ✓ | - | - |
| bulk_create_tasks | - | - | ✓ | ✓ | - | - |
| move_task | - | - | - | ✓ | - | - |
| assign_task | - | - | ✓ | ✓ | - | - |
| create_client | - | - | - | - | ✓ | - |
| update_client | - | - | - | - | ✓ | - |
| log_interaction | - | - | - | - | ✓ | - |
| update_pipeline | - | - | - | - | ✓ | - |
| get_team_capacity | - | - | ✓ | ✓ | - | ✓ |
| assign_team_member | - | - | ✓ | - | - | - |
| query_metrics | - | - | - | - | - | ✓ |
| get_trends | - | - | - | - | - | ✓ |
| log_activity | ✓ | - | - | - | - | - |

---

## Routing Logic (IntentRouter)

```
USER REQUEST
     │
     ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Is this a request for a FORMAL DOCUMENT?                            │
│ (proposal, SOP, report, framework, plan, playbook)                  │
└─────────────────────────────────────────────────────────────────────┘
     │
     ├─── YES ──► DOCUMENT AGENT
     │
     ▼ NO
┌─────────────────────────────────────────────────────────────────────┐
│ Is this about PROJECT management?                                   │
│ (create project, project planning, milestones, team allocation)     │
└─────────────────────────────────────────────────────────────────────┘
     │
     ├─── YES ──► PROJECT AGENT
     │
     ▼ NO
┌─────────────────────────────────────────────────────────────────────┐
│ Is this about TASK management?                                      │
│ (bulk tasks, prioritization, scheduling, dependencies)              │
└─────────────────────────────────────────────────────────────────────┘
     │
     ├─── YES ──► TASK AGENT (Project Agent)
     │
     ▼ NO
┌─────────────────────────────────────────────────────────────────────┐
│ Is this about CLIENT management?                                    │
│ (client profiles, interactions, pipeline, communications)           │
└─────────────────────────────────────────────────────────────────────┘
     │
     ├─── YES ──► CLIENT AGENT
     │
     ▼ NO
┌─────────────────────────────────────────────────────────────────────┐
│ Is this about DATA ANALYSIS or METRICS?                             │
│ (performance analysis, trends, reports, dashboards)                 │
└─────────────────────────────────────────────────────────────────────┘
     │
     ├─── YES ──► ANALYST AGENT
     │
     ▼ NO
┌─────────────────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR HANDLES DIRECTLY                    │
│  • General questions and conversation                               │
│  • Quick status checks                                              │
│  • Simple single operations                                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Cronograma Estimado

| Fase | Descrição | Estimativa | Status |
|------|-----------|------------|--------|
| 1 | Estrutura Base | 2-3h | 🔲 Pendente |
| 2 | Sistema de Prompts | 2-3h | 🔲 Pendente |
| 3 | Orchestrator Agent | 3-4h | 🔲 Pendente |
| 4 | Document Agent | 2h | 🔲 Pendente |
| 5 | Project Agent | 2-3h | 🔲 Pendente |
| 6 | Client Agent | 2h | 🔲 Pendente |
| 7 | Analyst Agent | 2h | 🔲 Pendente |
| 8 | Tool Registry | 3-4h | 🔲 Pendente |
| 9 | Streaming/Artifact | 2-3h | 🔲 Pendente |
| 10 | Integração | 2-3h | 🔲 Pendente |
| **Total** | | **22-30h** | |

---

## Notas de Implementação

### Backward Compatibility
- Manter `agents.go` original funcionando durante transição
- Adicionar feature flag para novo sistema
- Migração gradual

### Testing Strategy
- Unit tests para cada agente
- Integration tests para routing
- E2E tests para fluxo completo

### Rollback Plan
- Feature flag permite reverter instantaneamente
- Logs detalhados para debugging
- Métricas de comparação entre sistemas

---

## Status Atual (2025-12-26)

### ✅ Concluído

1. **Documento de Plano** - `docs/AGENT_IMPLEMENTATION_PLAN.md`
2. **Sistema de Prompts Modular**
   - `internal/prompts/core/identity.go` - Identidade OSA
   - `internal/prompts/core/formatting.go` - Padrões de formatação
   - `internal/prompts/core/artifacts.go` - Sistema de artifacts
   - `internal/prompts/core/context.go` - Integração de contexto
   - `internal/prompts/core/tools.go` - Padrões de uso de tools
   - `internal/prompts/core/errors.go` - Tratamento de erros
   - `internal/prompts/agent_prompts.go` - Prompts dos 5 agentes
   - `internal/prompts/composer.go` - Compositor de prompts

3. **Streaming/Artifact Detection**
   - `internal/streaming/events.go` - Tipos de eventos SSE
   - `internal/streaming/artifact_detector.go` - Detector de artifacts em stream

### 🔲 Próximos Passos

**Fase 1: Migrar Interface de Agentes (Requer Refatoração Cuidadosa)**

O código atual em `internal/agents/agents.go` define:
- `Agent` interface com `Run(ctx, messages) (<-chan string, <-chan error)`
- `BaseAgent` struct básico

Para implementar a nova arquitetura, precisamos:

1. **Opção A: Migração Gradual**
   - Criar nova interface `AgentV2` com `Run(ctx, AgentInput) (<-chan StreamEvent, <-chan error)`
   - Manter compatibilidade com código existente
   - Migrar handlers gradualmente

2. **Opção B: Refatoração Completa**
   - Atualizar `agents.go` para nova interface
   - Atualizar todos os handlers que usam agentes
   - Mais arriscado mas mais limpo

**Fase 2: Implementar Agentes Especializados**

Após resolver a interface, criar:
- `internal/agents/orchestrator/` - Com IntentRouter
- `internal/agents/document/`
- `internal/agents/project/`
- `internal/agents/client/`
- `internal/agents/analyst/`

**Fase 3: Integrar com chat.go**

- Usar `ArtifactDetector` para detectar artifacts em stream
- Emitir eventos SSE tipados
- Salvar artifacts automaticamente

---

## Arquivos Criados Nesta Sessão

```
internal/
├── prompts/
│   ├── core/
│   │   ├── identity.go
│   │   ├── formatting.go
│   │   ├── artifacts.go
│   │   ├── context.go
│   │   ├── tools.go
│   │   └── errors.go
│   ├── agent_prompts.go
│   └── composer.go
└── streaming/
    ├── events.go
    └── artifact_detector.go
```

---

*Documento atualizado: 2025-12-26*
