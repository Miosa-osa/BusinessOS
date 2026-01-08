# RAG System Performance Benchmarks

This directory contains comprehensive performance benchmarks for the RAG (Retrieval-Augmented Generation) system.

## Overview

The benchmark suite tests all critical components of the RAG system:

1. **Embedding Generation** - Text and image embeddings
2. **Vector Search** - Similarity search performance
3. **Hybrid Search** - Combined semantic + keyword search
4. **Re-Ranking** - Multi-signal result re-ranking
5. **Document Chunking** - Smart chunking strategies
6. **Caching** - Redis cache performance
7. **End-to-End Pipeline** - Complete RAG workflow

## Quick Start

### Run All Benchmarks

**Linux/Mac:**
```bash
cd desktop/backend-go
./run_rag_benchmarks.sh
```

**Windows:**
```powershell
cd desktop\backend-go
.\run_rag_benchmarks.ps1
```

This will:
- Run all benchmarks
- Generate a detailed performance report in `docs/rag_performance_report.md`
- Include memory allocation analysis
- Profile CPU and memory usage

### Run Specific Benchmarks

**Quick benchmarks (3 seconds each):**
```bash
./quick_benchmark.sh embedding  # Text embedding only
./quick_benchmark.sh vector     # Vector search only
./quick_benchmark.sh hybrid     # Hybrid search only
./quick_benchmark.sh cache      # Cache performance only
```

**Manual benchmark execution:**
```bash
cd internal/services

# Run specific benchmark
go test -bench=BenchmarkTextEmbedding -benchmem -benchtime=5s -run=^$

# Run benchmark with CPU profiling
go test -bench=BenchmarkHybridSearch -benchmem -cpuprofile=cpu.prof

# Run benchmark with memory profiling
go test -bench=BenchmarkVectorSearch -benchmem -memprofile=mem.prof
```

## Benchmark Categories

### 1. Embedding Benchmarks

```bash
# Text embeddings
go test -bench=BenchmarkTextEmbedding -benchmem

# Image embeddings
go test -bench=BenchmarkImageEmbedding -benchmem

# Parallel embedding generation
go test -bench=BenchmarkTextEmbeddingParallel -benchmem
```

**What it tests:**
- Embedding generation speed for different text sizes
- Parallel embedding performance
- Memory allocations
- Image embedding latency

### 2. Vector Search Benchmarks

```bash
# Search by dataset size
go test -bench=BenchmarkVectorSearch -benchmem

# Search by result limit
go test -bench=BenchmarkVectorSearchLimits -benchmem
```

**What it tests:**
- Search performance vs dataset size (10, 100, 1000 blocks)
- Performance vs result limits (5, 10, 25, 50, 100)
- PostgreSQL pgvector performance

### 3. Hybrid Search Benchmarks

```bash
# Overall hybrid search
go test -bench=BenchmarkHybridSearch -benchmem

# Different weight combinations
go test -bench=BenchmarkHybridSearchWeights -benchmem
```

**What it tests:**
- RRF (Reciprocal Rank Fusion) performance
- Semantic vs keyword search trade-offs
- Different weight combinations (semantic-heavy, balanced, keyword-heavy)

### 4. Re-Ranking Benchmarks

```bash
# Basic re-ranking
go test -bench=BenchmarkReRanking -benchmem

# Re-ranking different result counts
go test -bench=BenchmarkReRankingResultCounts -benchmem
```

**What it tests:**
- Multi-signal re-ranking overhead
- Performance vs number of results
- Metadata lookup efficiency

### 5. Chunking Benchmarks

```bash
# Chunking by document size
go test -bench=BenchmarkSmartChunking -benchmem

# Different chunking strategies
go test -bench=BenchmarkChunkingStrategies -benchmem

# Chunk size impact
go test -bench=BenchmarkChunkingSizes -benchmem
```

**What it tests:**
- Chunking performance for different document sizes
- Header-based vs fixed-size chunking
- Impact of chunk overlap
- Memory efficiency

### 6. Cache Benchmarks

```bash
# Cache hit performance
go test -bench=BenchmarkCacheHit -benchmem

# Cache miss performance
go test -bench=BenchmarkCacheMiss -benchmem

# Cache write performance
go test -bench=BenchmarkCacheSet -benchmem

# Hybrid search with caching
go test -bench=BenchmarkHybridSearchWithCache -benchmem
```

**What it tests:**
- Redis cache latency
- Cache hit speedup
- Cache miss overhead
- Overall cache effectiveness

### 7. End-to-End Pipeline

```bash
go test -bench=BenchmarkFullRAGPipeline -benchmem
```

**What it tests:**
- Complete RAG workflow performance
- Query → Search → Re-rank pipeline
- Real-world performance characteristics

## Understanding Results

### Benchmark Output Format

```
BenchmarkTextEmbedding/Small-50chars-8    1000    1234567 ns/op    5678 B/op    123 allocs/op
│                                        │       │         │        │            │
│                                        │       │         │        │            └─ Allocations per operation
│                                        │       │         │        └─ Bytes allocated per operation
│                                        │       │         └─ Nanoseconds per operation
│                                        │       └─ Iterations run
│                                        └─ CPU cores used (GOMAXPROCS)
```

### Key Metrics

- **ns/op (nanoseconds per operation)**: Lower is better
  - <100ms = Excellent
  - 100-300ms = Good
  - 300-500ms = Acceptable
  - >500ms = Needs optimization

- **B/op (bytes per operation)**: Memory used per operation
  - Lower is better
  - Watch for memory leaks (increasing over time)

- **allocs/op (allocations per operation)**: Number of heap allocations
  - Lower is better
  - High allocation counts can cause GC pressure

### Performance Targets

| Operation | Target Latency | Memory Budget |
|-----------|---------------|---------------|
| Text Embedding (500 chars) | <100ms | <10KB |
| Vector Search (1000 blocks) | <50ms | <50KB |
| Hybrid Search | <150ms | <100KB |
| Re-Ranking (25 results) | <50ms | <25KB |
| Document Chunking (100KB) | <100ms | <500KB |
| Cache Hit | <5ms | <1KB |
| Full RAG Pipeline | <300ms | <200KB |

## Profiling

### CPU Profiling

```bash
cd internal/services
go test -bench=BenchmarkHybridSearch -cpuprofile=cpu.prof
go tool pprof -http=:8080 cpu.prof
```

This opens an interactive web interface showing:
- Function call graphs
- Hot spots (where CPU time is spent)
- Flame graphs

### Memory Profiling

```bash
go test -bench=BenchmarkVectorSearch -memprofile=mem.prof
go tool pprof -http=:8080 mem.prof
```

This shows:
- Memory allocation hot spots
- Allocation call stacks
- Memory usage over time

### Trace Analysis

```bash
go test -bench=BenchmarkFullRAGPipeline -trace=trace.out
go tool trace trace.out
```

This shows:
- Goroutine scheduling
- GC pauses
- System calls
- Blocking operations

## Continuous Benchmarking

### Integration with CI/CD

Add to your CI pipeline:

```yaml
# .github/workflows/benchmarks.yml
name: Performance Benchmarks

on:
  pull_request:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run benchmarks
        run: |
          cd desktop/backend-go/internal/services
          go test -bench=. -benchmem -benchtime=5s > bench.txt

      - name: Compare with baseline
        run: |
          # Compare with previous results
          benchstat baseline.txt bench.txt
```

### Benchmark Comparison

```bash
# Baseline benchmarks
go test -bench=. -benchmem > baseline.txt

# After changes
go test -bench=. -benchmem > current.txt

# Compare
benchstat baseline.txt current.txt
```

Output:
```
name                    old time/op    new time/op    delta
TextEmbedding-8           125ms ± 2%     98ms ± 1%  -21.60%  (p=0.000 n=10+10)
HybridSearch-8            189ms ± 3%    165ms ± 2%  -12.70%  (p=0.000 n=10+10)

name                    old alloc/op   new alloc/op   delta
TextEmbedding-8          5.2kB ± 0%     4.1kB ± 0%  -21.15%  (p=0.000 n=10+10)
```

## Optimization Tips

### 1. Embedding Generation
- Use local Ollama instance (avoid network latency)
- Implement aggressive caching (24h TTL)
- Batch requests when possible
- Pre-generate embeddings for static content

### 2. Vector Search
- Ensure pgvector indexes are created
- Use appropriate result limits
- Consider read replicas for high load
- Regularly VACUUM ANALYZE

### 3. Hybrid Search
- Tune semantic/keyword weights for your use case
- Cache frequently searched queries
- Limit re-ranking to top N results

### 4. Caching
- Co-locate Redis with application
- Monitor cache hit rates (target >80%)
- Implement cache warming for common queries
- Use Redis pipeline for batch operations

### 5. General
- Use connection pooling
- Profile regularly to find bottlenecks
- Monitor p95/p99 latencies, not just averages
- Test with production-like data volumes

## Troubleshooting

### Benchmarks are slow

1. **Check external dependencies:**
   - Is Ollama running locally?
   - Is PostgreSQL optimized?
   - Is Redis accessible?

2. **Reduce benchmark time:**
   ```bash
   BENCH_TIME=1s ./quick_benchmark.sh embedding
   ```

3. **Run fewer iterations:**
   ```bash
   go test -bench=BenchmarkTextEmbedding -benchtime=10x
   ```

### Memory usage is high

1. **Check for leaks:**
   ```bash
   go test -bench=. -memprofile=mem.prof
   go tool pprof mem.prof
   ```

2. **Look for large allocations:**
   ```
   (pprof) top
   (pprof) list functionName
   ```

### Inconsistent results

1. **Run more iterations:**
   ```bash
   go test -bench=. -benchtime=30s
   ```

2. **Use benchstat for statistical comparison:**
   ```bash
   go test -bench=. -count=10 > results.txt
   benchstat results.txt
   ```

3. **Disable CPU frequency scaling:**
   ```bash
   # Linux
   sudo cpupower frequency-set --governor performance
   ```

## Contributing

When adding new benchmarks:

1. **Follow naming convention:**
   - `BenchmarkOperationName`
   - Use sub-benchmarks for variations: `b.Run("variant", ...)`

2. **Include memory profiling:**
   ```go
   b.ReportAllocs()
   ```

3. **Reset timer before measured code:**
   ```go
   b.ResetTimer()
   ```

4. **Document what you're testing:**
   ```go
   // BenchmarkNewFeature tests the performance of the new feature
   // with different input sizes
   func BenchmarkNewFeature(b *testing.B) { ... }
   ```

5. **Update documentation:**
   - Add to this README
   - Update performance report template

## Resources

- [Go Benchmark Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Performance Best Practices](https://go.dev/doc/effective_go#optimization)

---

**Questions or Issues?**

If you encounter problems with benchmarks or have suggestions for improvement, please open an issue or submit a PR.
