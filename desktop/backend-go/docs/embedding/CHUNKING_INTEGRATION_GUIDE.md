# Smart Chunking Integration Guide

## Overview

This guide shows how to integrate the Smart Chunking Service with the existing RAG infrastructure in BusinessOS.

## Integration Points

### 1. Document Indexing Pipeline

The Smart Chunking Service integrates with the existing embedding pipeline to provide better document chunking before embedding generation.

#### Before (Simple Splitting)

```go
// Old approach: simple block-based splitting
blocks := []services.Block{
    {ID: "1", Type: "paragraph", Content: "..."},
    {ID: "2", Type: "paragraph", Content: "..."},
}
embeddingSvc.IndexDocument(ctx, contextID, blocks)
```

#### After (Smart Chunking)

```go
// New approach: intelligent chunking with overlap and boundary respect
chunkingSvc := services.NewSmartChunkingService()
embeddingSvc := services.NewEmbeddingService(pool, ollamaURL)

// Step 1: Chunk the document
opts := services.ChunkOptions{
    ChunkSize:    512,
    OverlapRatio: 0.2,
    MinChunkSize: 128,
}

chunks, err := chunkingSvc.ChunkDocument(
    ctx,
    documentContent,
    services.DocTypeMarkdown,
    documentID,
    opts,
)

// Step 2: Convert chunks to blocks for embedding
blocks := make([]services.Block, len(chunks))
for i, chunk := range chunks {
    blocks[i] = services.Block{
        ID:      chunk.ID,
        Type:    "smart_chunk",
        Content: chunk.Content,
    }
}

// Step 3: Generate embeddings and index
err = embeddingSvc.IndexDocument(ctx, contextID, blocks)
```

### 2. Handler Integration

Update handlers to use smart chunking for document processing.

#### Example: Document Upload Handler

```go
package handlers

import (
    "github.com/labstack/echo/v4"
    "internal/services"
)

type DocumentHandler struct {
    chunkingSvc  *services.SmartChunkingService
    embeddingSvc *services.EmbeddingService
    // ... other dependencies
}

func NewDocumentHandler(
    chunkingSvc *services.SmartChunkingService,
    embeddingSvc *services.EmbeddingService,
) *DocumentHandler {
    return &DocumentHandler{
        chunkingSvc:  chunkingSvc,
        embeddingSvc: embeddingSvc,
    }
}

func (h *DocumentHandler) UploadDocument(c echo.Context) error {
    ctx := c.Request().Context()
    userID := c.Get("user_id").(string)

    // Parse request
    var req struct {
        Content     string `json:"content"`
        Type        string `json:"type"`
        Name        string `json:"name"`
        ProjectID   string `json:"project_id"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(400, "Invalid request")
    }

    // Determine document type
    docType := h.detectDocumentType(req.Type, req.Content)

    // Smart chunking
    opts := services.ChunkOptions{
        ChunkSize:    512,
        OverlapRatio: 0.2,
        MinChunkSize: 128,
    }

    chunks, err := h.chunkingSvc.ChunkDocument(
        ctx,
        req.Content,
        docType,
        req.Name,
        opts,
    )
    if err != nil {
        return echo.NewHTTPError(500, "Chunking failed: "+err.Error())
    }

    // Create context entry
    contextID := uuid.New()
    // ... save context to database

    // Convert to blocks and index
    blocks := h.chunksToBlocks(chunks)
    err = h.embeddingSvc.IndexDocument(ctx, contextID, blocks)
    if err != nil {
        return echo.NewHTTPError(500, "Indexing failed: "+err.Error())
    }

    return c.JSON(200, map[string]interface{}{
        "context_id":  contextID,
        "chunks":      len(chunks),
        "total_tokens": h.getTotalTokens(chunks),
    })
}

func (h *DocumentHandler) detectDocumentType(fileType, content string) services.DocumentType {
    switch fileType {
    case "markdown", "md":
        return services.DocTypeMarkdown
    case "code", "go", "py", "js", "ts":
        return services.DocTypeCode
    case "json":
        return services.DocTypeJSON
    case "xml":
        return services.DocTypeXML
    default:
        return services.DocTypePlainText
    }
}

func (h *DocumentHandler) chunksToBlocks(chunks []services.Chunk) []services.Block {
    blocks := make([]services.Block, len(chunks))
    for i, chunk := range chunks {
        blocks[i] = services.Block{
            ID:      chunk.ID,
            Type:    "smart_chunk",
            Content: chunk.Content,
        }
    }
    return blocks
}

func (h *DocumentHandler) getTotalTokens(chunks []services.Chunk) int {
    total := 0
    for _, chunk := range chunks {
        total += chunk.TokenCount
    }
    return total
}
```

### 3. Integration with Agentic RAG

Enhance the Agentic RAG service to leverage smart chunking metadata.

```go
package services

// Enhanced AgenticRAGService with chunking awareness
type AgenticRAGServiceV2 struct {
    *AgenticRAGService // Embed existing service
    chunkingSvc *SmartChunkingService
}

func NewAgenticRAGServiceV2(
    pool *pgxpool.Pool,
    hybridSearch *HybridSearchService,
    reranker *ReRankerService,
    embeddingSvc *EmbeddingService,
    learningSvc *LearningService,
    chunkingSvc *SmartChunkingService,
) *AgenticRAGServiceV2 {
    return &AgenticRAGServiceV2{
        AgenticRAGService: NewAgenticRAGService(
            pool,
            hybridSearch,
            reranker,
            embeddingSvc,
            learningSvc,
        ),
        chunkingSvc: chunkingSvc,
    }
}

// RetrieveWithContext uses chunk metadata for better retrieval
func (a *AgenticRAGServiceV2) RetrieveWithContext(
    ctx context.Context,
    req AgenticRAGRequest,
) (*AgenticRAGResponse, error) {
    // Use base retrieval
    response, err := a.Retrieve(ctx, req)
    if err != nil {
        return nil, err
    }

    // Enhance with chunk context
    for i, result := range response.Results {
        // Load chunk metadata if available
        if chunkMetadata, ok := result.Metadata["chunk_metadata"]; ok {
            // Use metadata to improve relevance
            // E.g., boost chunks from same section
            if sectionIdx, ok := chunkMetadata.(map[string]interface{})["section_idx"]; ok {
                response.Results[i].FinalScore *= 1.1 // 10% boost for same section
            }
        }
    }

    return response, nil
}
```

### 4. Database Schema Extension

Optionally extend the schema to store chunk metadata.

```sql
-- Add chunk metadata column to context_embeddings
ALTER TABLE context_embeddings
ADD COLUMN chunk_metadata JSONB;

-- Index for chunk metadata queries
CREATE INDEX idx_context_embeddings_chunk_metadata
ON context_embeddings USING gin(chunk_metadata);

-- Example query using chunk metadata
SELECT
    ce.*,
    ce.chunk_metadata->>'heading' as heading,
    ce.chunk_metadata->>'section_idx' as section_idx
FROM context_embeddings ce
WHERE ce.chunk_metadata->>'type' = 'markdown_section'
  AND ce.chunk_metadata->>'level' = '2';
```

### 5. Service Initialization

Update the main server initialization to include smart chunking.

```go
// In cmd/server/main.go or service initialization

func initializeServices(pool *pgxpool.Pool, config *Config) *Services {
    // Create smart chunking service
    chunkingSvc := services.NewSmartChunkingService()

    // Create embedding service
    embeddingSvc := services.NewEmbeddingService(pool, config.OllamaURL)

    // Create other RAG services
    hybridSearch := services.NewHybridSearchService(pool, embeddingSvc)
    reranker := services.NewReRankerService(pool, embeddingSvc)
    ragCache := services.NewRAGCacheService(pool)
    queryExpansion := services.NewQueryExpansionService(pool, embeddingSvc)

    // Set up service dependencies
    embeddingSvc.SetCache(ragCache)

    // Create agentic RAG with chunking support
    agenticRAG := services.NewAgenticRAGServiceV2(
        pool,
        hybridSearch,
        reranker,
        embeddingSvc,
        learningSvc,
        chunkingSvc,
    )
    agenticRAG.SetCache(ragCache)
    agenticRAG.SetQueryExpansion(queryExpansion)

    return &Services{
        ChunkingSvc:  chunkingSvc,
        EmbeddingSvc: embeddingSvc,
        AgenticRAG:   agenticRAG,
        // ... other services
    }
}
```

### 6. API Endpoints

Add endpoints for chunking preview and statistics.

```go
// Preview how a document would be chunked
func (h *DocumentHandler) PreviewChunking(c echo.Context) error {
    ctx := c.Request().Context()

    var req struct {
        Content   string                    `json:"content"`
        Type      services.DocumentType      `json:"type"`
        Options   *services.ChunkOptions     `json:"options"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(400, err)
    }

    // Use defaults if not provided
    opts := services.DefaultChunkOptions()
    if req.Options != nil {
        opts = *req.Options
    }

    // Chunk document
    chunks, err := h.chunkingSvc.ChunkDocument(
        ctx,
        req.Content,
        req.Type,
        "preview",
        opts,
    )
    if err != nil {
        return echo.NewHTTPError(500, err)
    }

    // Get statistics
    stats := h.chunkingSvc.GetChunkStatistics(chunks)

    return c.JSON(200, map[string]interface{}{
        "chunks":     chunks,
        "statistics": stats,
    })
}

// Get chunking statistics for a document
func (h *DocumentHandler) GetChunkingStats(c echo.Context) error {
    contextID := c.Param("context_id")

    // Load chunks from database
    chunks, err := h.loadChunksForContext(c.Request().Context(), contextID)
    if err != nil {
        return echo.NewHTTPError(500, err)
    }

    stats := h.chunkingSvc.GetChunkStatistics(chunks)

    return c.JSON(200, stats)
}
```

### 7. Configuration

Add chunking configuration to the application config.

```go
type RAGConfig struct {
    // Chunking settings
    Chunking struct {
        DefaultChunkSize    int     `json:"default_chunk_size" default:"512"`
        DefaultOverlap      float64 `json:"default_overlap" default:"0.2"`
        MinChunkSize        int     `json:"min_chunk_size" default:"128"`
        PreserveCodeBlocks  bool    `json:"preserve_code_blocks" default:"true"`
    } `json:"chunking"`

    // Embedding settings
    Embedding struct {
        Model      string `json:"model" default:"nomic-embed-text"`
        Dimensions int    `json:"dimensions" default:"768"`
        BatchSize  int    `json:"batch_size" default:"100"`
    } `json:"embedding"`

    // Search settings
    Search struct {
        DefaultLimit       int     `json:"default_limit" default:"10"`
        MinSimilarity      float64 `json:"min_similarity" default:"0.5"`
        SemanticWeight     float64 `json:"semantic_weight" default:"0.6"`
        KeywordWeight      float64 `json:"keyword_weight" default:"0.4"`
    } `json:"search"`
}

// Load from config file
func LoadRAGConfig(path string) (*RAGConfig, error) {
    // ... load and parse config
}

// Use in service initialization
func NewSmartChunkingServiceFromConfig(config *RAGConfig) *SmartChunkingService {
    svc := NewSmartChunkingService()
    // Could extend to allow custom configuration
    return svc
}
```

## Usage Examples

### Example 1: Processing a Knowledge Base Article

```go
func ProcessKnowledgeBaseArticle(
    ctx context.Context,
    article string,
    chunkingSvc *services.SmartChunkingService,
    embeddingSvc *services.EmbeddingService,
) error {
    // Chunk the article
    opts := services.ChunkOptions{
        ChunkSize:    512,
        OverlapRatio: 0.25, // Slightly higher overlap for KB articles
        MinChunkSize: 128,
    }

    chunks, err := chunkingSvc.ChunkMarkdown(ctx, article, "kb-article", opts)
    if err != nil {
        return err
    }

    // Validate chunks
    validChunks := []services.Chunk{}
    for _, chunk := range chunks {
        if valid, _ := chunkingSvc.ValidateChunk(chunk, opts); valid {
            validChunks = append(validChunks, chunk)
        }
    }

    // Optimize
    optimized := chunkingSvc.OptimizeChunks(validChunks, opts)

    // Convert and index
    blocks := chunksToBlocks(optimized)
    contextID := uuid.New()
    return embeddingSvc.IndexDocument(ctx, contextID, blocks)
}
```

### Example 2: Processing Code Documentation

```go
func ProcessCodeDocumentation(
    ctx context.Context,
    code string,
    language string,
    chunkingSvc *services.SmartChunkingService,
    embeddingSvc *services.EmbeddingService,
) error {
    // Use smaller chunks for code (more precise retrieval)
    opts := services.ChunkOptions{
        ChunkSize:    256,
        OverlapRatio: 0.15,
        PreserveCode: true,
        MinChunkSize: 64,
    }

    chunks, err := chunkingSvc.ChunkCode(ctx, code, language, "code-doc", opts)
    if err != nil {
        return err
    }

    // Enrich chunks with code-specific metadata
    for i := range chunks {
        if funcName, ok := chunks[i].Metadata["name"]; ok {
            chunks[i].Metadata["searchable_name"] = funcName
        }
    }

    // Index
    blocks := chunksToBlocks(chunks)
    contextID := uuid.New()
    return embeddingSvc.IndexDocument(ctx, contextID, blocks)
}
```

### Example 3: Batch Processing

```go
func BatchProcessDocuments(
    ctx context.Context,
    documents []Document,
    chunkingSvc *services.SmartChunkingService,
    embeddingSvc *services.EmbeddingService,
) error {
    opts := services.DefaultChunkOptions()

    // Process in parallel
    type result struct {
        docID string
        err   error
    }
    results := make(chan result, len(documents))

    for _, doc := range documents {
        go func(d Document) {
            chunks, err := chunkingSvc.ChunkDocument(
                ctx,
                d.Content,
                d.Type,
                d.ID,
                opts,
            )
            if err != nil {
                results <- result{d.ID, err}
                return
            }

            blocks := chunksToBlocks(chunks)
            contextID := uuid.New()
            err = embeddingSvc.IndexDocument(ctx, contextID, blocks)
            results <- result{d.ID, err}
        }(doc)
    }

    // Collect results
    for i := 0; i < len(documents); i++ {
        r := <-results
        if r.err != nil {
            log.Printf("Failed to process document %s: %v", r.docID, r.err)
        }
    }

    return nil
}
```

## Testing Integration

```go
func TestChunkingIntegration(t *testing.T) {
    // Setup
    pool := setupTestDB(t)
    chunkingSvc := services.NewSmartChunkingService()
    embeddingSvc := services.NewEmbeddingService(pool, testOllamaURL)

    ctx := context.Background()
    content := `# Test Document

## Section 1
Content here.

## Section 2
More content.`

    // Chunk
    opts := services.DefaultChunkOptions()
    chunks, err := chunkingSvc.ChunkMarkdown(ctx, content, "test-doc", opts)
    if err != nil {
        t.Fatal(err)
    }

    // Index
    blocks := chunksToBlocks(chunks)
    contextID := uuid.New()
    err = embeddingSvc.IndexDocument(ctx, contextID, blocks)
    if err != nil {
        t.Fatal(err)
    }

    // Search
    results, err := embeddingSvc.SimilaritySearch(ctx, "Section 1", 5, testUserID)
    if err != nil {
        t.Fatal(err)
    }

    if len(results) == 0 {
        t.Error("Expected search results")
    }
}
```

## Migration Strategy

### Phase 1: Deploy Service (No Breaking Changes)

1. Deploy smart chunking service
2. Keep existing chunking as fallback
3. Add feature flag for smart chunking

### Phase 2: Parallel Running

1. Process new documents with smart chunking
2. Keep old documents with existing chunks
3. Monitor performance and quality

### Phase 3: Re-indexing (Optional)

1. Re-chunk and re-index important documents
2. Compare retrieval quality
3. Gradually migrate all documents

### Phase 4: Full Migration

1. Make smart chunking the default
2. Remove old chunking code
3. Update documentation

## Monitoring

Add metrics to track chunking performance:

```go
type ChunkingMetrics struct {
    TotalDocuments      int64
    TotalChunks         int64
    AvgChunksPerDoc     float64
    AvgTokensPerChunk   float64
    ChunkingErrorRate   float64
    ProcessingTimeP95   time.Duration
}

func (h *DocumentHandler) GetChunkingMetrics(c echo.Context) error {
    metrics := h.calculateMetrics()
    return c.JSON(200, metrics)
}
```

## Troubleshooting

### Issue: Chunks too small

**Solution**: Increase `MinChunkSize` or adjust content splitting logic

### Issue: Chunks too large

**Solution**: Decrease `ChunkSize` or check for large unbreakable blocks

### Issue: Poor retrieval quality

**Solution**: Adjust overlap ratio, try different chunking strategies

### Issue: High memory usage

**Solution**: Process documents in batches, stream large files

## Conclusion

The Smart Chunking Service integrates seamlessly with the existing RAG infrastructure and provides significant improvements in document processing quality. Follow this guide to adopt smart chunking in your application while maintaining backward compatibility.
