# P2: Advanced RAG & Search

> **Priority:** P2 - Nice to Have
> **Backend Status:** Complete (14 endpoints)
> **Frontend Status:** Partial (4 basic endpoints)
> **Estimated Effort:** 1 sprint

---

## Overview

BusinessOS has sophisticated RAG (Retrieval-Augmented Generation) capabilities including hybrid search, re-ranking, and agentic retrieval. Currently only basic embedding search is used.

---

## Backend API Endpoints

### Basic (Currently Used)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/embeddings/search` | Basic semantic search |
| POST | `/api/embeddings/context` | Build AI context |
| GET | `/api/embeddings/stats` | Embedding statistics |
| GET | `/api/embeddings/health` | Health check |

### Advanced (Not Used)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/rag/search/hybrid` | Hybrid semantic + keyword search |
| POST | `/api/rag/search/hybrid/explain` | Explain hybrid results |
| GET | `/api/rag/search/explain` | Explain any search results |
| POST | `/api/rag/search/rerank` | Re-rank search results |
| POST | `/api/rag/search/rerank/explain` | Explain re-ranking |
| POST | `/api/rag/retrieve` | Agentic adaptive retrieval |
| POST | `/api/search/multimodal` | Multi-modal search (text + image) |

---

## Frontend Implementation Tasks

### Phase 1: Enhanced Search UI
- [ ] Advanced search modal with options
- [ ] Toggle: Semantic vs Hybrid vs Keyword
- [ ] Search results with relevance scores
- [ ] "Explain results" button

### Phase 2: Search Explanation
- [ ] Show why results matched
- [ ] Highlight matching terms
- [ ] Display relevance breakdown

### Phase 3: Multi-modal Search
- [ ] Image upload for search
- [ ] Combined text + image queries
- [ ] Visual search results

---

## Linear Issues to Create

1. **[RAG-001]** Create advanced search modal
2. **[RAG-002]** Implement hybrid search toggle
3. **[RAG-003]** Add search explanation UI
4. **[RAG-004]** Build multi-modal search
5. **[RAG-005]** API client updates

---

## Notes

- Hybrid search significantly improves results
- Re-ranking is expensive but valuable for precision
- Multi-modal opens up image-based workflows
