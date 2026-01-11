# Daily Paper Check - Workflow Manual

Como fazer check diário de papers relevantes do arXiv.

## 🚀 Quick Start

### Opção 1: Via Claude Code (Recomendado)

1. Abrir Claude Code no projeto BusinessOS
2. Usar o comando:

```
/mcp

Buscar papers do arXiv publicados ontem nas categorias:
- cs.AI (Artificial Intelligence)
- cs.CL (Computation and Language)
- cs.LG (Machine Learning)

Para cada paper:
1. Ler título e abstract
2. Calcular relevance score usando scorer.py
3. Se score > 60:
   - Baixar PDF
   - Criar pasta em papers/arxiv/{arxiv_id}/
   - Gerar metadata.json
   - Salvar em index/paper_index.json
   - Adicionar a applications/{feature}/ se relevante
```

### Opção 2: Manual via arXiv

1. Visitar https://arxiv.org/list/cs.AI/recent
2. Filtrar por "New submissions" do dia anterior
3. Para cada paper interessante:
   - Copiar arXiv ID (ex: 2501.12345)
   - Usar scorer.py para calcular relevância
   - Baixar manualmente se relevante

### Opção 3: Python Script (Futuro)

```python
#!/usr/bin/env python3
# daily_check.py

import arxiv
from scorer import PaperScorer

def check_daily_papers():
    """Busca e analisa papers do dia"""

    # Buscar papers do último dia em categorias relevantes
    client = arxiv.Client()
    search = arxiv.Search(
        query='cat:cs.AI OR cat:cs.CL OR cat:cs.LG',
        max_results=100,
        sort_by=arxiv.SortCriterion.SubmittedDate,
        sort_order=arxiv.SortOrder.Descending
    )

    scorer = PaperScorer()
    relevant_papers = []

    for paper in client.results(search):
        score = scorer.score_paper(
            paper.title,
            paper.summary,
            paper.published.isoformat()
        )

        if score.total >= 60:
            relevant_papers.append({
                'arxiv_id': paper.get_short_id(),
                'title': paper.title,
                'score': score.total,
                'pdf_url': paper.pdf_url
            })

    # Salvar papers relevantes
    print(f"Found {len(relevant_papers)} relevant papers")
    for p in relevant_papers:
        print(f"  [{p['score']:.1f}] {p['arxiv_id']}: {p['title']}")

    return relevant_papers

if __name__ == "__main__":
    check_daily_papers()
```

## 📋 Workflow Detalhado

### 1. Busca (Manhã - 5 min)

Buscar novos papers publicados ontem:

```bash
# Via MCP no Claude Code
search_papers(
  query="(cat:cs.AI OR cat:cs.CL OR cat:cs.LG) AND submittedDate:[yesterday TO today]",
  max_results=50
)
```

Categorias priorizadas:
- `cs.AI` - Inteligência Artificial
- `cs.CL` - Processamento de Linguagem Natural
- `cs.LG` - Machine Learning
- `cs.DB` - Databases (ocasionalmente)

### 2. Filtro Inicial (10 min)

Para cada paper, ler:
1. Título
2. Abstract (primeiro parágrafo)
3. Conclusão (se disponível)

Filtrar se contém palavras-chave relevantes:
- ✅ "RAG", "retrieval", "embedding"
- ✅ "agent", "orchestration", "multi-agent"
- ✅ "memory", "context", "workspace"
- ✅ "vector database", "pgvector", "semantic search"
- ❌ "computer vision", "image", "audio"
- ❌ "robotics", "hardware"

### 3. Scoring (5 min por paper)

```bash
cd articletouse
python scorer.py --title "..." --abstract "..." --date "2026-01-06"
```

Ou usar Claude Code:

```
Calcular relevance score para este paper:

Title: [título]
Abstract: [abstract]
Published: 2026-01-06

Usar as regras em config/businessos_features.json
```

### 4. Download e Organização (10 min)

Para papers com score > 60:

```bash
# Criar estrutura
mkdir -p papers/arxiv/2501.12345

# Baixar PDF via MCP
download_paper(arxiv_id="2501.12345")

# Ou manual
wget https://arxiv.org/pdf/2501.12345.pdf -O papers/arxiv/2501.12345/paper.pdf
```

### 5. Metadata e Indexação (5 min)

Criar `papers/arxiv/2501.12345/metadata.json`:

```json
{
  "id": "generated-uuid",
  "arxiv_id": "2501.12345",
  "title": "Paper Title",
  "abstract": "...",
  "relevance": {
    "score": 87,
    "score_breakdown": {...},
    "businessos_features": ["memory_hierarchy", "rag"]
  },
  "applications": [...],
  "timestamps": {
    "added": "2026-01-06T10:00:00Z"
  }
}
```

Atualizar `index/paper_index.json`:

```json
{
  "papers": [
    {
      "id": "uuid",
      "arxiv_id": "2501.12345",
      "title": "...",
      "relevance_score": 87,
      "added_date": "2026-01-06"
    }
  ]
}
```

### 6. Review Rápido (15-30 min)

Para papers de alta prioridade (score > 70):

1. Ler introdução completa
2. Ler seção de métodos (overview)
3. Ler resultados principais
4. Anotar key insights em `notes.md`
5. Gerar application ideas
6. Priorizar implementação

### 7. Notificação (Opcional)

Criar resumo diário:

```
Daily Paper Report - 2026-01-06
================================

Papers analisados: 47
Papers relevantes: 5
Alta prioridade (>70): 2

High Priority:
1. [87] 2501.12345 - Adaptive Query Expansion for RAG
   → Feature: rag_enhancement
   → Implementation: 2-4 days
   → Impact: High

2. [82] 2501.67890 - Memory Hierarchies in Multi-Agent Systems
   → Feature: memory_hierarchy, agent_orchestration
   → Implementation: 5-7 days
   → Impact: Very High

Medium Priority:
3. [65] 2501.11111 - Efficient Vector Indexing
4. [63] 2501.22222 - LLM Context Management
5. [61] 2501.33333 - Hybrid Search Optimization
```

## ⏱️ Tempo Total

- **Busca**: 5 min
- **Filtro inicial**: 10 min
- **Scoring (5 papers)**: 25 min
- **Download/organização**: 10 min
- **Review rápido (2 papers)**: 30 min

**Total: ~1h20min por dia**

## 🎯 Goals

- **Mínimo**: 1-2 papers relevantes por semana
- **Ideal**: 3-5 papers relevantes por semana
- **Review profundo**: 1 paper por semana
- **Implementação**: 1 paper por mês

## 📊 Métricas

Rastrear mensalmente:
- Total de papers analisados
- Papers salvos (score > 60)
- Papers implementados
- Impacto medido (performance improvements)

## 💡 Tips

1. **Seja seletivo**: Melhor 2 papers bem analisados que 10 superficiais
2. **Priorize implementação**: Papers com code disponível são mais valiosos
3. **Link ao código**: Sempre conecte papers a arquivos específicos do BusinessOS
4. **Revise semanalmente**: Olhe os papers da semana e priorize 1-2 para deep dive
5. **Compartilhe insights**: Anote descobertas para a equipe

## 🔗 Links Úteis

- arXiv AI: https://arxiv.org/list/cs.AI/recent
- arXiv CL: https://arxiv.org/list/cs.CL/recent
- arXiv LG: https://arxiv.org/list/cs.LG/recent
- Papers with Code: https://paperswithcode.com/
- Hugging Face Daily Papers: https://huggingface.co/papers

---

**Próximo check:** Verificar diariamente às 9h (hora que arXiv atualiza)
