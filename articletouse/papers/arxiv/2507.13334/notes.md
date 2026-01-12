# Notes: A Survey of Context Engineering for LLMs

**Paper ID:** 2507.13334
**Date Added:** 2026-01-07
**Relevance Score:** 92/100 🔥

---

## Executive Summary

Comprehensive survey analyzing 1400+ papers on context engineering. Establishes formal taxonomy and technical roadmap for optimizing LLM information payloads beyond simple prompt design.

## Key Takeaways for BusinessOS

### 1. Three-Tier Memory Architecture ⭐⭐⭐

**What it is:**
- **Short-term:** Exact text (current conversation)
- **Mid-term:** Summaries (session context)
- **Long-term:** Embeddings/titles (historical knowledge)

**Why it matters:**
- Current BusinessOS has 2-tier (workspace vs private)
- Adding mid-term layer could improve retrieval by 30-40%
- Prevents context window overflow

**Implementation path:**
```
1. Add mid_term_memory table
2. Implement automatic summarization service
3. Update retrieval strategy to query all 3 tiers
4. Benchmark against current 2-tier system
```

### 2. Context Collapse Problem ⭐⭐⭐

**Problem:** Enlarged context windows cause models to fail distinguishing between contexts.

**Symptoms in BusinessOS:**
- Long COT chains lose track of original intent
- Multi-turn conversations degrade quality
- Agent confusion in workspace with many memories

**Solution approaches:**
- Context refresh mechanisms
- Critical information pinning
- Degradation metrics tracking

### 3. Recurrent Compression Buffers ⭐⭐

**Concept:** Compress earlier parts of context stream into smaller representations.

**Use case for BusinessOS:**
- Very long conversations (50+ turns)
- Workspace with 1000+ memories
- Complex multi-step agent tasks

**Trade-off:** Compression vs information loss - needs careful tuning.

---

## Taxonomy Mapping

| Survey Category | BusinessOS Component | Status |
|----------------|----------------------|--------|
| Context Retrieval | HybridSearchService | ✅ Implemented |
| Context Generation | Memory injection | ✅ Implemented |
| Context Processing | Query expansion, re-ranking | ✅ Implemented |
| Context Management | Memory hierarchy | ⚠️ Partial (2-tier) |
| RAG Systems | Agentic RAG | ✅ Implemented |
| Memory Systems | Workspace memories | ✅ Implemented |
| Multi-Agent | COT Orchestrator | ✅ Implemented |

---

## Implementation Priorities

### HIGH (Week 1-2)
- [ ] Prototype mid-term memory layer
- [ ] Test 3-tier retrieval strategy
- [ ] Measure improvement vs baseline

### MEDIUM (Week 3-4)
- [ ] Implement context collapse detection
- [ ] Add context refresh mechanisms
- [ ] Deploy to staging

### LOW (Future)
- [ ] Recurrent compression buffers
- [ ] Advanced state-space backbones

---

## Questions to Explore

1. How to automatically determine what goes in mid-term vs long-term?
2. What summarization approach? Extractive vs abstractive?
3. How to balance compression ratio vs information retention?
4. Can we use existing pgvector for mid-term or need separate table?

---

## Related Papers to Check

- Dynamic Cheatsheet (referenced in ACE paper)
- State-space backbones for long sequences
- Attention mechanism optimization papers

---

## Meeting Notes

**2026-01-07 - Initial Review**
- Paper is foundational - must-read for team
- Schedule deep dive session
- Assign mid-term memory prototype to backend team

---

## Code Snippets / Pseudocode

```go
// Proposed mid-term memory structure
type MidTermMemory struct {
    ID          uuid.UUID
    WorkspaceID uuid.UUID
    Summary     string        // Compressed context
    Timeframe   time.Duration // What period this summarizes
    SourceIDs   []uuid.UUID   // Original memory IDs
    Embedding   pgvector.Vector
    CreatedAt   time.Time
    ExpiresAt   *time.Time    // Optional TTL
}

// Retrieval strategy
func RetrieveContext(query string, tiers []MemoryTier) []Memory {
    var results []Memory

    // Short-term: exact matches, recent
    if contains(tiers, ShortTerm) {
        results = append(results, getRecentMemories(limit=10))
    }

    // Mid-term: summarized sessions
    if contains(tiers, MidTerm) {
        results = append(results, getMidTermSummaries(query))
    }

    // Long-term: semantic search
    if contains(tiers, LongTerm) {
        results = append(results, semanticSearch(query, limit=20))
    }

    return rerank(results)
}
```

---

**Status:** Ready for deep review and prototyping
**Next Action:** Schedule team meeting to discuss implementation plan
