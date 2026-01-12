# ArticleToUse - Quick Start Guide

## ✅ Instalação Completa!

O sistema ArticleToUse está pronto para uso com:
- ✅ arXiv MCP Server instalado
- ✅ Estrutura de pastas criada
- ✅ Relevance scorer funcionando
- ✅ Configuração do MCP em `.mcp.json`

---

## 🚀 Uso Imediato

### 1. Reiniciar Claude Code

**IMPORTANTE:** Para que o MCP seja carregado, você precisa:

1. Fechar esta sessão do Claude Code
2. Reabrir o Claude Code no projeto BusinessOS
3. O MCP arxiv será carregado automaticamente

### 2. Verificar MCP Ativo

Após reabrir, digite:

```
/mcp
```

Você deve ver:

```
Available MCP Servers:
- arxiv (stdio)
  Tools: search_papers, download_paper, list_papers, read_paper
```

### 3. Primeiro Uso - Buscar Papers

```
Buscar papers do arXiv publicados nos últimos 7 dias sobre:
- Retrieval-Augmented Generation (RAG)
- Vector databases
- Agent orchestration

Categorias: cs.AI, cs.CL, cs.LG
Max: 10 papers

Use search_papers
```

### 4. Analisar Relevância

Para cada paper encontrado, calcular score:

```bash
cd articletouse
python scorer.py
```

Ou pedir ao Claude:

```
Para este paper:
- Title: [título]
- Abstract: [abstract]
- Published: 2026-01-06

Calcular relevance score usando config/businessos_features.json
```

### 5. Baixar Paper Relevante (score > 60)

```
Baixar o paper arXiv:2501.12345 para articletouse/papers/arxiv/2501.12345/
```

### 6. Organizar

```
Criar metadata.json para o paper 2501.12345 com:
- Relevance score calculado
- Features do BusinessOS matched
- Application ideas geradas
- Target files identificados
```

---

## 📋 Exemplo Completo de Workflow

### Cenário: Busca Diária de Papers

**Prompt completo para Claude Code:**

```
# Workflow diário de papers do arXiv

1. Buscar papers do arXiv de ontem nas categorias cs.AI, cs.CL, cs.LG

2. Para os 5 papers mais promissores:
   a. Ler título e abstract
   b. Calcular relevance score usando scorer.py
   c. Se score > 60: baixar PDF e criar estrutura

3. Criar resumo:
   - Total de papers analisados
   - Papers relevantes encontrados
   - Top 3 por score
   - Application ideas principais

Use as tools do MCP arxiv conforme necessário.
```

---

## 🔧 Comandos MCP Disponíveis

### `search_papers`

```
Buscar papers com:
- query: "RAG OR vector database"
- max_results: 20
- sort_by: submittedDate
- sort_order: descending
```

### `download_paper`

```
Baixar paper arXiv:2501.12345
```

### `list_papers`

```
Listar todos os papers que já baixei
```

### `read_paper`

```
Ler conteúdo do paper 2501.12345
```

---

## 📊 Relevance Scorer - Uso Rápido

### Via Python Script

```bash
cd C:\Users\Pichau\Desktop\BusinessOS-main-dev\articletouse
python scorer.py
```

Saída esperada:
```
======================================================================
RELEVANCE SCORE REPORT
======================================================================

Title: Hierarchical Memory Systems for LLMs...

Total Score: 68.7/100

Breakdown:
  Technology Match:      25.0/25
  Feature Alignment:     6.2/30
  Feasibility:           15.0/20
  Innovation:            12.5/15
  Recency:               10.0/10

[MEDIUM] RELEVANCE - Add to watch list

APPLICATION IDEAS (3):
...
```

### Via Claude Code

```
Analisar relevância deste paper para BusinessOS:

Title: "Adaptive Query Expansion for RAG Systems"
Abstract: "We propose a novel approach to query expansion..."
Published: "2026-01-06T00:00:00Z"

Usar:
- config/businessos_features.json para features
- Algoritmo de scoring: Technology (25) + Feature (30) + Feasibility (20) + Innovation (15) + Recency (10)

Retornar:
- Score total
- Score breakdown
- Features matched
- Application ideas com target files
```

---

## 🎯 Thresholds de Relevância

- **70-100**: 🔥 Alta prioridade - Implementar ASAP
- **50-69**: ⚠️ Média - Watch list, revisar semanalmente
- **40-49**: 📋 Baixa - Referência futura
- **< 40**: ❌ Filtrado - Não armazenar

---

## 📁 Estrutura de Um Paper Completo

```
articletouse/papers/arxiv/2501.12345/
├── paper.pdf                 # PDF original
├── metadata.json             # Metadata estruturada
├── extracted_text.md         # Texto extraído (opcional)
└── notes.md                  # Suas anotações
```

**metadata.json exemplo mínimo:**

```json
{
  "id": "uuid-gerado",
  "arxiv_id": "2501.12345",
  "title": "Paper Title",
  "abstract": "Full abstract...",
  "published_date": "2026-01-06T00:00:00Z",
  "relevance": {
    "score": 87,
    "businessos_features": ["memory_hierarchy", "rag_enhancement"]
  },
  "applications": [
    {
      "feature_area": "rag_enhancement",
      "title": "Improve query expansion",
      "target_files": ["desktop/backend-go/internal/services/query_expansion.go"],
      "priority": 8
    }
  ],
  "timestamps": {
    "added": "2026-01-06T10:00:00Z"
  }
}
```

---

## 🔍 Troubleshooting

### MCP não aparece após reiniciar

1. Verificar se `.mcp.json` existe na raiz do projeto:
   ```bash
   cat C:\Users\Pichau\Desktop\BusinessOS-main-dev\.mcp.json
   ```

2. Verificar se `uv` está instalado:
   ```bash
   uv --version
   ```

3. Verificar se `arxiv-mcp-server` está instalado:
   ```bash
   uv tool list
   ```

4. Testar manualmente:
   ```bash
   uv tool run arxiv-mcp-server --help
   ```

### Erro ao baixar paper

- Verificar conexão com internet
- arXiv ID correto? (formato: YYMM.NNNNN)
- Servidor do arXiv pode estar temporariamente indisponível

### Scorer.py não funciona

```bash
cd articletouse
python --version  # Deve ser Python 3.7+
python scorer.py  # Testa com exemplo built-in
```

---

## 📚 Próximos Passos

### Hoje
1. ✅ MCP instalado e configurado
2. ⏳ Reiniciar Claude Code
3. ⏳ Fazer primeira busca
4. ⏳ Baixar 1-2 papers relevantes

### Esta Semana
- [ ] Configurar busca diária
- [ ] Revisar 5-10 papers
- [ ] Escolher 1 paper para implementação
- [ ] Prototipar primeira ideia

### Este Mês
- [ ] Implementar feature de 1 paper
- [ ] Criar automação de busca
- [ ] Integrar com workspace memories
- [ ] Dashboard de visualização

---

## 💡 Tips

1. **Seja seletivo**: 2-3 papers de qualidade > 10 superficiais
2. **Priorize implementação**: Papers com código disponível
3. **Link ao código**: Sempre conecte a arquivos específicos do BusinessOS
4. **Review semanal**: Reserve 1h para revisar papers da semana
5. **Implemente mensalmente**: 1 paper implementado/mês é excelente progresso

---

## 🆘 Suporte

Se algo não funcionar:

1. Verificar logs do Claude Code
2. Verificar `.mcp.json` está correto
3. Reiniciar Claude Code
4. Verificar `uv tool list` mostra arxiv-mcp-server

**Configuração criada em:** 2026-01-06
**Status:** ✅ Pronto para uso após reiniciar Claude Code
