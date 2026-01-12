# ArticleToUse - AI Research Paper Management System

Sistema local para gerenciar papers de pesquisa em IA relevantes para o BusinessOS.

## 📋 Visão Geral

Este sistema:
1. Busca papers diários do arXiv via MCP server
2. Calcula relevância para features do BusinessOS
3. Armazena localmente com metadata estruturada
4. Gera ideias de aplicação automáticas
5. Rastreia implementação e review

## 🏗️ Estrutura de Pastas

```
articletouse/
├── papers/                     # Papers armazenados
│   ├── arxiv/
│   │   └── {arxiv_id}/
│   │       ├── paper.pdf
│   │       ├── metadata.json
│   │       ├── extracted_text.md
│   │       └── notes.md
│   └── other/
│
├── database/                   # Dados locais
│   ├── papers.db              # SQLite (futuro)
│   └── embeddings/
│
├── index/                      # Índices de busca
│   ├── paper_index.json       # Índice master
│   ├── by_relevance.json      # Por score
│   ├── by_topic.json          # Por tópico
│   └── reviewed_status.json   # Status de review
│
├── applications/               # Papers por feature
│   ├── memory_hierarchy/
│   ├── rag_enhancement/
│   ├── agent_orchestration/
│   └── ...
│
├── collections/                # Coleções curadas
│   ├── must_read.json
│   ├── implementation_ready.json
│   └── long_term.json
│
└── config/                     # Configuração
    ├── schema.json            # Schema de metadata
    ├── businessos_features.json # Features do BusinessOS
    └── sources.json           # Configuração de fontes
```

## ⚙️ Setup

### 1. Instalar arXiv MCP Server

```bash
# Opção 1: Via Smithery (recomendado)
npx -y @smithery/cli install arxiv-mcp-server --client claude

# Opção 2: Manual com uv
uv tool install arxiv-mcp-server
```

### 2. Configurar MCP no Claude Code

```bash
claude mcp add --transport stdio arxiv -- uv tool run arxiv-mcp-server \
  --storage-path "C:\Users\Pichau\Desktop\BusinessOS-main-dev\articletouse\papers"
```

### 3. Verificar Configuração

```bash
claude mcp list
# Deve mostrar: arxiv (stdio)
```

## 🎯 Uso Básico

### Buscar Papers Relevantes

No Claude Code, use o MCP:

```
/mcp

Tools disponíveis do arxiv MCP:
- search_papers: Buscar papers por query
- download_paper: Baixar PDF por arXiv ID
- list_papers: Listar papers baixados
- read_paper: Ler conteúdo de paper
```

### Exemplo de Busca Diária

```
Buscar papers do arXiv publicados hoje sobre:
- Vector databases
- RAG (retrieval-augmented generation)
- Agent orchestration
- Memory systems for LLMs

Para cada paper relevante (score > 60):
1. Baixar PDF
2. Criar metadata.json
3. Calcular relevance score
4. Gerar application ideas
5. Atualizar índices
```

## 📊 Relevance Scoring

Sistema de scoring multidimensional (0-100):

```
Score Final = (
  Technology Match × 0.25 +        # 0-25 pontos
  Feature Alignment × 0.30 +       # 0-30 pontos
  Implementation Feasibility × 0.20 + # 0-20 pontos
  Innovation Potential × 0.15 +    # 0-15 pontos
  Recency × 0.10                   # 0-10 pontos
)
```

### Technology Match (0-25)

**Tier 1 (5 pontos cada):**
- Vector databases / pgvector
- RAG / Retrieval-Augmented Generation
- Embeddings / Semantic search
- Agent orchestration / Multi-agent
- LLM optimization

**Tier 2 (3 pontos cada):**
- Hybrid search
- Query expansion
- Re-ranking
- Redis caching
- Go concurrency

**Tier 3 (1 ponto cada):**
- API design
- Database indexing
- Authentication
- TypeScript patterns

### Feature Alignment (0-30)

Match com features do BusinessOS:

- **Memory Hierarchy (10 pts):** CUS-25 - Workspace vs private memories
- **RAG Enhancement (10 pts):** CUS-41 - Agentic RAG, hybrid search, re-ranking
- **Agent Orchestration (10 pts):** COT reasoning, multi-agent, tool calling

### Thresholds

- **70-100:** Alta relevância - implementar ASAP
- **50-69:** Média relevância - watch list
- **40-49:** Baixa relevância - referência futura
- **< 40:** Filtrado

## 🔧 Workflow Completo

### 1. Busca Diária (Automática)

```bash
# Script futuro: daily_check.sh
# 1. Consultar arXiv via MCP
# 2. Filtrar por categorias: cs.AI, cs.CL, cs.LG
# 3. Calcular relevance score
# 4. Baixar papers com score > 60
# 5. Gerar metadata
# 6. Atualizar índices
# 7. Enviar notificação
```

### 2. Review Manual

Para cada paper de alta relevância:

1. Ler abstract e conclusões
2. Avaliar feasibility
3. Anotar key insights
4. Gerar application ideas
5. Priorizar implementação
6. Marcar como reviewed

### 3. Implementação

1. Escolher paper prioritário
2. Ler detalhadamente
3. Prototipar solução
4. Integrar com BusinessOS
5. Testar e validar
6. Atualizar metadata (implemented: true)

## 📝 Formato de Metadata

Exemplo: `papers/arxiv/2501.12345/metadata.json`

```json
{
  "id": "uuid",
  "arxiv_id": "2501.12345",
  "title": "Hierarchical Memory for LLMs",
  "abstract": "...",
  "authors": ["Jane Smith", "John Doe"],
  "published_date": "2025-01-15T00:00:00Z",

  "relevance": {
    "score": 87,
    "score_breakdown": {
      "technology_match": 20,
      "feature_alignment": 28,
      "implementation_feasibility": 18,
      "innovation_potential": 13,
      "recency_relevance": 8
    },
    "businessos_features": ["memory_hierarchy", "rag"],
    "confidence": 0.92
  },

  "applications": [
    {
      "feature_area": "memory_hierarchy",
      "description": "Implement attention-based memory prioritization",
      "target_files": ["desktop/backend-go/internal/services/memory_hierarchy_service.go"],
      "implementation_difficulty": "medium",
      "priority": 8,
      "estimated_effort": "3-5 days",
      "impact": "High - 20-30% improvement"
    }
  ],

  "review_status": {
    "reviewed": true,
    "review_date": "2026-01-06T15:30:00Z",
    "rating": 5,
    "summary": "Excellent paper with clear implementation path",
    "key_insights": [
      "Three-tier memory approach matches our design",
      "Query expansion technique could improve our RAG",
      "Attention mechanism for memory prioritization"
    ]
  },

  "files": {
    "pdf_path": "papers/arxiv/2501.12345/paper.pdf",
    "extracted_text_path": "papers/arxiv/2501.12345/extracted_text.md",
    "notes_path": "papers/arxiv/2501.12345/notes.md"
  },

  "tags": ["high-priority", "implementation-ready", "memory", "rag"],
  "timestamps": {
    "added": "2026-01-06T10:00:00Z",
    "last_accessed": "2026-01-06T15:30:00Z",
    "last_modified": "2026-01-06T15:30:00Z"
  }
}
```

## 🔍 Busca e Query

### Query por Relevância

```json
// index/by_relevance.json
{
  "high": [
    {"id": "uuid", "title": "...", "score": 87}
  ],
  "medium": [...],
  "low": [...]
}
```

### Query por Feature

```json
// applications/memory_hierarchy/index.json
{
  "feature_area": "memory_hierarchy",
  "businessos_files": [
    "desktop/backend-go/internal/services/memory_hierarchy_service.go"
  ],
  "related_papers": [
    {
      "id": "uuid",
      "title": "...",
      "key_takeaways": [...],
      "implementation_ideas": [...]
    }
  ]
}
```

## 🚀 Próximos Passos

### Fase 1: Storage (✅ COMPLETO)
- [x] Criar estrutura de pastas
- [x] Definir schema de metadata
- [x] Configurar MCP server
- [x] Criar índices iniciais

### Fase 2: Scorer (Em Progresso)
- [ ] Implementar algoritmo de relevance scoring
- [ ] Criar script de extração de keywords
- [ ] Integrar com embeddings do BusinessOS
- [ ] Gerar application ideas automaticamente

### Fase 3: Automação
- [ ] Script de busca diária
- [ ] Atualização automática de índices
- [ ] Sistema de notificações
- [ ] Dashboard de status

### Fase 4: Integração
- [ ] Importar papers como workspace memories
- [ ] Semantic search via RAG
- [ ] Link papers → código BusinessOS
- [ ] Tracking de implementação

## 📚 Recursos

- **MCP Server:** [blazickjp/arxiv-mcp-server](https://github.com/blazickjp/arxiv-mcp-server)
- **arXiv API:** [arXiv.org API](https://arxiv.org/help/api)
- **BusinessOS Features:** `config/businessos_features.json`
- **Metadata Schema:** `config/schema.json`

## 💡 Tips

1. **Busca eficiente:** Use categorias específicas do arXiv (cs.AI, cs.CL, cs.LG)
2. **Filtre early:** Score < 40 é automaticamente descartado
3. **Priorize:** Papers com code available são mais fáceis de implementar
4. **Link sempre:** Conecte papers a arquivos específicos do BusinessOS
5. **Review rápido:** Leia abstract + conclusões primeiro

---

**Sistema criado:** 2026-01-06
**Última atualização:** 2026-01-06
