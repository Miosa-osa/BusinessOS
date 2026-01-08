# Smart Chunking Service

## Overview

The Smart Chunking Service implements intelligent document chunking for RAG (Retrieval Augmented Generation) optimization. It provides context-aware splitting of documents while preserving semantic boundaries, maintaining overlap for context continuity, and respecting document structure.

## Features

### Core Capabilities

- **Multiple Document Types**: Supports Markdown, code (Go, Python, JavaScript/TypeScript), plain text, JSON, and XML
- **Intelligent Boundary Detection**: Respects natural boundaries like paragraphs, code blocks, functions, and sections
- **Configurable Chunking**: Customizable chunk size, overlap ratio, and minimum chunk size
- **Overlap Strategy**: Maintains context continuity between chunks through configurable overlap
- **Metadata Preservation**: Tracks source, position, document structure, and custom metadata
- **Token-Aware**: Uses token estimation for accurate chunk sizing
- **Optimization**: Post-processing to merge small chunks and normalize whitespace

### Document-Specific Features

#### Markdown
- Splits by headers (H1-H6)
- Preserves section hierarchy
- Maintains header context in metadata

#### Code
- Language-specific parsing (Go, Python, JavaScript/TypeScript)
- Splits by functions, classes, and methods
- Preserves code block boundaries
- Automatic language detection

#### Plain Text
- Splits by paragraphs
- Falls back to sentence splitting for large paragraphs
- Maintains natural reading flow

#### Structured (JSON/XML)
- Preserves structure
- Splits by logical elements
- Maintains formatting

## Installation

```go
import "path/to/services"

svc := services.NewSmartChunkingService()
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "path/to/services"
)

func main() {
    // Create service
    svc := services.NewSmartChunkingService()
    ctx := context.Background()

    // Configure options
    opts := services.ChunkOptions{
        ChunkSize:    512,  // Target 512 tokens per chunk
        OverlapRatio: 0.2,  // 20% overlap between chunks
        PreserveCode: true,
        MinChunkSize: 128,  // Minimum 128 tokens per chunk
    }

    // Chunk a markdown document
    content := `# My Document

## Introduction
This is an introduction.

## Details
More detailed content here.`

    chunks, err := svc.ChunkMarkdown(ctx, content, "doc-001", opts)
    if err != nil {
        log.Fatal(err)
    }

    // Process chunks
    for i, chunk := range chunks {
        fmt.Printf("Chunk %d: %d tokens\n", i+1, chunk.TokenCount)
        fmt.Printf("  Content: %s\n", chunk.Content[:min(50, len(chunk.Content))])
        fmt.Printf("  Metadata: %v\n", chunk.Metadata)
    }
}
```

### Chunking Strategies

#### Strategy 1: Precise Retrieval (Small chunks, high overlap)

```go
preciseOpts := services.ChunkOptions{
    ChunkSize:    256,
    OverlapRatio: 0.3, // 30% overlap
    MinChunkSize: 64,
}
chunks, _ := svc.ChunkDocument(ctx, content, services.DocTypeMarkdown, "doc", preciseOpts)
```

**Best for**: Factual lookup, precise information retrieval

#### Strategy 2: Context-Rich (Large chunks, low overlap)

```go
contextOpts := services.ChunkOptions{
    ChunkSize:    1024,
    OverlapRatio: 0.1, // 10% overlap
    MinChunkSize: 256,
}
chunks, _ := svc.ChunkDocument(ctx, content, services.DocTypeMarkdown, "doc", contextOpts)
```

**Best for**: Understanding complex topics, maintaining narrative flow

#### Strategy 3: Balanced (Default)

```go
balancedOpts := services.DefaultChunkOptions()
chunks, _ := svc.ChunkDocument(ctx, content, services.DocTypeMarkdown, "doc", balancedOpts)
```

**Best for**: General purpose RAG applications

### Document Types

```go
// Markdown
chunks, err := svc.ChunkMarkdown(ctx, markdownContent, parentDocID, opts)

// Code (with language)
chunks, err := svc.ChunkCode(ctx, codeContent, "go", parentDocID, opts)

// Code (auto-detect language)
chunks, err := svc.ChunkCode(ctx, codeContent, "", parentDocID, opts)

// Plain text
chunks, err := svc.ChunkPlainText(ctx, textContent, parentDocID, opts)

// Structured (JSON/XML)
chunks, err := svc.ChunkStructured(ctx, jsonContent, "json", parentDocID, opts)

// Auto-routing based on type
chunks, err := svc.ChunkDocument(ctx, content, services.DocTypeMarkdown, parentDocID, opts)
```

### Statistics and Validation

```go
// Get chunk statistics
stats := svc.GetChunkStatistics(chunks)
fmt.Printf("Total chunks: %v\n", stats["count"])
fmt.Printf("Average tokens: %v\n", stats["avg_tokens"])
fmt.Printf("Min/Max tokens: %v / %v\n", stats["min_tokens"], stats["max_tokens"])

// Validate individual chunk
valid, msg := svc.ValidateChunk(chunk, opts)
if !valid {
    fmt.Printf("Chunk validation failed: %s\n", msg)
}

// Optimize chunks
optimizedChunks := svc.OptimizeChunks(chunks, opts)
```

## API Reference

### Types

#### ChunkOptions

```go
type ChunkOptions struct {
    ChunkSize    int     // Target chunk size in tokens (default: 512)
    OverlapRatio float64 // Overlap ratio between chunks (default: 0.2)
    PreserveCode bool    // Preserve code block boundaries (default: true)
    MinChunkSize int     // Minimum chunk size in tokens (default: 128)
}
```

#### Chunk

```go
type Chunk struct {
    ID          string                 // Unique chunk identifier
    Content     string                 // Chunk content
    TokenCount  int                    // Estimated token count
    Position    int                    // Position in document (0-indexed)
    Metadata    map[string]interface{} // Chunk metadata
    ParentDocID string                 // Parent document identifier
}
```

#### DocumentType

```go
type DocumentType string

const (
    DocTypeMarkdown  DocumentType = "markdown"
    DocTypeCode      DocumentType = "code"
    DocTypePlainText DocumentType = "plaintext"
    DocTypeJSON      DocumentType = "json"
    DocTypeXML       DocumentType = "xml"
)
```

### Methods

#### ChunkDocument

```go
func (s *SmartChunkingService) ChunkDocument(
    ctx context.Context,
    content string,
    docType DocumentType,
    parentDocID string,
    options ChunkOptions,
) ([]Chunk, error)
```

Routes to appropriate chunking strategy based on document type.

#### ChunkMarkdown

```go
func (s *SmartChunkingService) ChunkMarkdown(
    ctx context.Context,
    content string,
    parentDocID string,
    options ChunkOptions,
) ([]Chunk, error)
```

Chunks markdown content respecting headers and structure.

#### ChunkCode

```go
func (s *SmartChunkingService) ChunkCode(
    ctx context.Context,
    content string,
    language string,
    parentDocID string,
    options ChunkOptions,
) ([]Chunk, error)
```

Chunks code content respecting language syntax. Language can be empty for auto-detection.

#### ChunkPlainText

```go
func (s *SmartChunkingService) ChunkPlainText(
    ctx context.Context,
    content string,
    parentDocID string,
    options ChunkOptions,
) ([]Chunk, error)
```

Chunks plain text respecting paragraphs and sentences.

#### GetChunkStatistics

```go
func (s *SmartChunkingService) GetChunkStatistics(chunks []Chunk) map[string]interface{}
```

Returns statistics about a set of chunks.

#### ValidateChunk

```go
func (s *SmartChunkingService) ValidateChunk(chunk Chunk, options ChunkOptions) (bool, string)
```

Validates a chunk meets quality criteria.

#### OptimizeChunks

```go
func (s *SmartChunkingService) OptimizeChunks(chunks []Chunk, options ChunkOptions) []Chunk
```

Post-processes chunks to improve quality.

## Integration with RAG Pipeline

### Step 1: Chunk Documents

```go
svc := services.NewSmartChunkingService()
chunks, err := svc.ChunkDocument(ctx, document, docType, docID, opts)
```

### Step 2: Generate Embeddings

```go
embeddingSvc := services.NewEmbeddingService(pool, ollamaURL)
for _, chunk := range chunks {
    embedding, err := embeddingSvc.GenerateEmbedding(ctx, chunk.Content)
    // Store embedding with chunk metadata
}
```

### Step 3: Store in Vector Database

```go
// Store chunks with embeddings in database
for _, chunk := range chunks {
    // Insert into context_embeddings table with metadata
}
```

### Step 4: Retrieve During Query

```go
// Query retrieves most relevant chunks
results, err := embeddingSvc.SimilaritySearch(ctx, query, limit, userID)
```

## Metadata Fields

Each chunk includes metadata that varies by document type:

### Common Metadata
- `position`: Chunk position in document (0-indexed)
- `token_count`: Estimated token count
- `char_count`: Character count
- `has_overlap`: Whether chunk includes overlap from previous
- `overlap_tokens`: Number of overlap tokens (if applicable)

### Markdown-Specific
- `type`: "markdown_section"
- `heading`: Section heading text
- `level`: Heading level (1-6)
- `section_idx`: Section index

### Code-Specific
- `type`: "code_block"
- `language`: Programming language
- `block_type`: "function", "class", "method", etc.
- `name`: Function/class name
- `block_idx`: Block index

### Plain Text-Specific
- `type`: "plain_text"
- `paragraph_count`: Number of paragraphs in chunk
- `sentence_count`: Number of sentences in chunk

## Performance Considerations

### Token Estimation

The service uses a simple estimation of ~4 characters per token. For production use with specific models, consider:

1. Using model-specific tokenizers (e.g., tiktoken for OpenAI models)
2. Caching token counts for repeated content
3. Adjusting `charsPerToken` based on your content language

### Optimization Tips

1. **Batch Processing**: Process multiple documents in parallel
2. **Caching**: Cache chunk results for frequently accessed documents
3. **Lazy Loading**: Only chunk documents when needed
4. **Incremental Updates**: Re-chunk only modified sections

### Memory Usage

For very large documents (>10MB), consider:

1. Streaming approaches for reading content
2. Processing in sections
3. Limiting maximum document size

## Error Handling

```go
chunks, err := svc.ChunkDocument(ctx, content, docType, parentDocID, opts)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "empty content"):
        // Handle empty content
    case strings.Contains(err.Error(), "chunk_size"):
        // Handle invalid options
    default:
        // Handle other errors
    }
}
```

## Testing

```bash
# Run all tests
go test ./internal/services/ -run TestSmartChunking

# Run specific test
go test ./internal/services/ -run TestChunkMarkdown

# Run with coverage
go test ./internal/services/ -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test ./internal/services/ -bench=BenchmarkChunk -benchmem
```

## Best Practices

### 1. Choose Appropriate Chunk Size

- **Small (256 tokens)**: Precise retrieval, more chunks to search
- **Medium (512 tokens)**: Balanced approach, good for most use cases
- **Large (1024 tokens)**: Context-rich, fewer chunks, slower retrieval

### 2. Set Overlap Based on Use Case

- **High (30%)**: When context continuity is critical
- **Medium (20%)**: Default, good balance
- **Low (10%)**: When storage is a concern

### 3. Validate Chunks Before Indexing

```go
for _, chunk := range chunks {
    if valid, msg := svc.ValidateChunk(chunk, opts); !valid {
        log.Printf("Invalid chunk: %s", msg)
        continue
    }
    // Index valid chunk
}
```

### 4. Use Appropriate Document Type

Correct document type detection improves chunking quality:

```go
// Detect type based on file extension or content
docType := detectDocumentType(filename, content)
chunks, err := svc.ChunkDocument(ctx, content, docType, docID, opts)
```

### 5. Monitor Statistics

```go
stats := svc.GetChunkStatistics(chunks)
if avgTokens := stats["avg_tokens"].(int); avgTokens < opts.MinChunkSize {
    log.Printf("Warning: Average chunk size (%d) is small", avgTokens)
}
```

## Limitations

1. **Language Detection**: Simple heuristic-based detection may not work for all languages
2. **Token Estimation**: Approximation may differ from actual model tokenization
3. **Complex Structures**: Some document structures may not split optimally
4. **Non-Text Content**: Images, tables, and diagrams require special handling

## Future Enhancements

1. **Advanced Tokenization**: Integration with model-specific tokenizers
2. **Semantic Splitting**: Use embeddings to determine optimal split points
3. **Table Support**: Specialized handling for tables and structured data
4. **Image Context**: Extract and preserve image captions and context
5. **Cross-Reference**: Maintain links between related chunks
6. **Adaptive Chunking**: Dynamically adjust chunk size based on content complexity

## Contributing

When contributing improvements:

1. Add tests for new functionality
2. Update documentation
3. Consider backward compatibility
4. Add examples for new features
5. Benchmark performance changes

## License

Part of BusinessOS - Internal use only.
