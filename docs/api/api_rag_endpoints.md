# RAG API Endpoints Documentation

**Base URL**: `/api/rag`
**Authentication**: Required (Bearer token)
**Added**: Day 2 Integration (2026-01-05)

---

## Hybrid Search

### POST `/api/rag/search/hybrid`

Performs hybrid search combining semantic (vector) and keyword (full-text) approaches.

**Request Body**:
```json
{
  "query": "machine learning concepts",
  "semantic_weight": 0.7,      // Optional, default: 0.7
  "keyword_weight": 0.3,       // Optional, default: 0.3
  "max_results": 10,           // Optional, default: 10
  "min_similarity": 0.3        // Optional, default: 0.3
}
```

**Response**:
```json
{
  "query": "machine learning concepts",
  "results": [
    {
      "context_id": "uuid",
      "block_id": "block-123",
      "block_type": "paragraph",
      "content": "Machine learning is...",
      "context_name": "AI Documentation",
      "context_type": "document",
      "semantic_score": 0.89,
      "keyword_score": 0.45,
      "hybrid_score": 0.82,
      "search_strategy": "hybrid"
    }
  ],
  "count": 10,
  "options": {
    "semantic_weight": 0.7,
    "keyword_weight": 0.3,
    "max_results": 10
  }
}
```

**Features**:
- Combines vector cosine similarity with PostgreSQL full-text search
- Reciprocal Rank Fusion (RRF) algorithm for result merging
- Configurable semantic/keyword weights
- Returns hybrid scores normalized to 0-1 range

---

### POST `/api/rag/search/hybrid/explain`

Provides detailed explanation of hybrid search results and strategy.

**Request Body**: Same as `/search/hybrid`

**Response**:
```json
{
  "query": "machine learning concepts",
  "total_results": 15,
  "strategy_breakdown": {
    "semantic": 8,
    "keyword": 3,
    "hybrid": 4
  },
  "avg_semantic_score": 0.756,
  "avg_keyword_score": 0.423,
  "avg_hybrid_score": 0.682,
  "options": {
    "semantic_weight": 0.7,
    "keyword_weight": 0.3,
    "rrf_constant": 60,
    "min_similarity": 0.3
  },
  "top_5_results": [...]
}
```

**Use Cases**:
- Debugging search quality
- Understanding which strategy produced results
- Tuning weights for specific query types

---

## Agentic RAG

### POST `/api/rag/retrieve`

Performs intelligent, adaptive retrieval with query understanding and self-critique.

**Request Body**:
```json
{
  "query": "How to implement authentication?",
  "max_results": 10,                  // Optional, default: 10
  "min_quality_score": 0.6,           // Optional, default: 0.5
  "project_id": "uuid",               // Optional
  "task_id": "uuid",                  // Optional
  "use_personalization": true         // Optional, default: false
}
```

**Response**:
```json
{
  "results": [
    {
      // Standard hybrid search result fields
      "context_id": "uuid",
      "block_id": "block-456",
      "content": "Authentication implementation...",

      // Re-ranking scores
      "semantic_score": 0.85,
      "keyword_score": 0.62,
      "hybrid_score": 0.78,
      "recency_score": 0.95,
      "quality_score": 0.82,
      "interaction_score": 0.65,
      "context_score": 1.0,
      "final_score": 0.87,

      // Ranking information
      "original_rank": 3,
      "reranked_position": 1,
      "rank_change": 2,

      // Score breakdown
      "score_breakdown": {
        "semantic": 0.85,
        "recency": 0.95,
        "quality": 0.82,
        "interaction": 0.65,
        "context": 1.0
      }
    }
  ],
  "query_intent": "procedural",
  "strategy_used": "hybrid",
  "strategy_reasoning": "How-to queries need both semantic understanding and keyword precision",
  "quality_score": 0.82,
  "iteration_count": 1,
  "personalized": true,
  "processing_time_ms": 245,
  "metadata": {
    "intent_classification": "procedural",
    "user_preferences": {
      "preferred_tone": "professional",
      "preferred_verbosity": "moderate",
      "expertise_areas": ["go", "authentication"]
    }
  }
}
```

**Query Intent Types**:
- `factual_lookup` - "What is X?"
- `conceptual_search` - Exploring concepts/ideas
- `procedural` - "How to X?"
- `comparison` - "X vs Y"
- `recent` - "Latest/Recent X"
- `exhaustive` - "All/Everything about X"
- `ambiguous` - Unclear queries

**Search Strategies**:
- `semantic_only` - Pure vector search
- `keyword_only` - Pure full-text search
- `hybrid` - Balanced combination
- `multi_pass` - Multiple passes with deduplication

**Features**:
- Automatic query intent classification
- Strategy selection based on intent
- Multi-signal re-ranking (5 signals)
- Self-critique with retry logic (up to 3 iterations)
- Personalization integration
- Detailed execution metadata

---

## Memory Management

### GET `/api/rag/memories`

Lists memories for the authenticated user.

**Query Parameters**:
- `type` (optional) - Filter by memory type
- `limit` (optional) - Max results (default: 50, max: 100)

**Response**:
```json
{
  "memories": [
    {
      "id": "uuid",
      "user_id": "user-123",
      "title": "API Authentication Pattern",
      "summary": "Preferred authentication approach",
      "content": "Use JWT with refresh tokens...",
      "memory_type": "pattern",
      "category": "security",
      "importance_score": 0.85,
      "access_count": 12,
      "is_pinned": false,
      "tags": ["auth", "jwt", "security"],
      "created_at": "2026-01-05T10:30:00Z",
      "updated_at": "2026-01-05T10:30:00Z"
    }
  ],
  "count": 42
}
```

---

### GET `/api/rag/memories/:id`

Retrieves a specific memory.

**Response**: Single memory object (same structure as list item)

---

### POST `/api/rag/memories`

Creates a new memory.

**Request Body**:
```json
{
  "title": "API Authentication Pattern",
  "summary": "Preferred authentication approach",
  "content": "Use JWT with refresh tokens...",
  "memory_type": "pattern",
  "category": "security",
  "source_type": "conversation",
  "source_id": "uuid",             // Optional
  "project_id": "uuid",            // Optional
  "node_id": "uuid",               // Optional
  "importance_score": 0.85,        // Optional, default: 0.5
  "tags": ["auth", "jwt"]          // Optional
}
```

**Response**: Created memory object with ID

**Memory Types**:
- `pattern` - Reusable patterns
- `decision` - Architectural decisions
- `fact` - Confirmed facts
- `preference` - User preferences
- `context` - Contextual information

---

## Integration Examples

### Basic Hybrid Search
```javascript
const response = await fetch('/api/rag/search/hybrid', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    query: 'user authentication best practices',
    semantic_weight: 0.7,
    keyword_weight: 0.3,
    max_results: 10
  })
});

const data = await response.json();
console.log(`Found ${data.count} results`);
data.results.forEach(result => {
  console.log(`${result.context_name}: ${result.content.substring(0, 100)}...`);
  console.log(`Scores - Semantic: ${result.semantic_score}, Keyword: ${result.keyword_score}, Hybrid: ${result.hybrid_score}`);
});
```

### Agentic RAG with Personalization
```javascript
const response = await fetch('/api/rag/retrieve', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    query: 'How do I secure my API endpoints?',
    max_results: 10,
    min_quality_score: 0.6,
    use_personalization: true,
    project_id: currentProjectId
  })
});

const data = await response.json();
console.log(`Intent: ${data.query_intent}`);
console.log(`Strategy: ${data.strategy_used} - ${data.strategy_reasoning}`);
console.log(`Quality: ${data.quality_score.toFixed(2)}`);
console.log(`Iterations: ${data.iteration_count}`);
console.log(`Processing time: ${data.processing_time_ms}ms`);
console.log(`Personalized: ${data.personalized}`);

data.results.forEach((result, i) => {
  console.log(`\n${i+1}. ${result.context_name}`);
  console.log(`   Final score: ${result.final_score.toFixed(2)} (rank change: ${result.rank_change > 0 ? '+' : ''}${result.rank_change})`);
  console.log(`   Breakdown: Semantic=${result.semantic_score.toFixed(2)}, Recency=${result.recency_score.toFixed(2)}, Quality=${result.quality_score.toFixed(2)}`);
});
```

### Create Memory from Conversation
```javascript
const response = await fetch('/api/rag/memories', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    title: 'Preferred Database Indexing Strategy',
    summary: 'Use B-tree for equality, GiST for full-text',
    content: 'Based on project requirements, we decided to use B-tree indexes for equality searches and GiST indexes for full-text search capabilities...',
    memory_type: 'decision',
    category: 'database',
    source_type: 'conversation',
    importance_score: 0.9,
    tags: ['database', 'indexing', 'postgresql']
  })
});

const memory = await response.json();
console.log(`Memory created: ${memory.id}`);
```

---

## Performance Characteristics

| Endpoint | Expected Latency | Notes |
|----------|------------------|-------|
| `/search/hybrid` | 60-250ms | Depends on Ollama response time |
| `/search/hybrid/explain` | 60-250ms | Same as hybrid search |
| `/retrieve` (1 iteration) | 80-320ms | Includes re-ranking |
| `/retrieve` (2 iterations) | 160-640ms | If quality threshold not met |
| `/retrieve` (3 iterations) | 240-960ms | Max iterations |
| `/memories` (list) | 10-30ms | Database query only |
| `/memories/:id` (get) | 5-15ms | Database query only |
| `/memories` (create) | 50-200ms | Includes embedding generation |

---

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "Error message here"
}
```

**Common Status Codes**:
- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `503 Service Unavailable` - Service not initialized
- `500 Internal Server Error` - Server error

---

## Configuration

RAG services are automatically initialized if dependencies are available:
- Hybrid Search requires: PostgreSQL + Embedding Service
- Re-Ranker requires: PostgreSQL + Embedding Service
- Agentic RAG requires: All RAG components + Learning Service
- Memory Service requires: PostgreSQL + Embedding Service

Check service availability at startup in server logs:
```
Hybrid search service initialized (semantic + keyword with RRF)
Re-ranker service initialized (multi-signal relevance scoring)
Agentic RAG service initialized (intelligent adaptive retrieval)
RAG services registered (hybrid search, re-ranker, agentic RAG, memory)
```

---

## Best Practices

1. **Use Agentic RAG for User-Facing Search**:
   - Automatic intent classification
   - Quality-based retries
   - Personalization support
   - Better overall results

2. **Use Hybrid Search for Programmatic Queries**:
   - More control over weights
   - Faster (no re-ranking)
   - Predictable behavior

3. **Tune Weights Based on Query Type**:
   - Exact terms: Higher keyword weight (0.2 semantic, 0.8 keyword)
   - Concepts: Higher semantic weight (0.9 semantic, 0.1 keyword)
   - Mixed: Balanced (0.7 semantic, 0.3 keyword)

4. **Monitor Quality Scores**:
   - Track agentic RAG quality scores
   - If consistently low, adjust min_quality_score threshold
   - Review iteration counts for retry frequency

5. **Leverage Personalization**:
   - Enable for user-facing queries
   - Provides better context awareness
   - Respects user preferences

6. **Create Memories Strategically**:
   - High importance (>0.7) for critical decisions
   - Tag consistently for better retrieval
   - Link to projects/nodes for context

---

**Version**: 1.0.0 (Day 2)
**Last Updated**: 2026-01-05
