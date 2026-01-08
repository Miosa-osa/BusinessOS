# RAG Benchmarks - Quick Reference Card

## Quick Start

```bash
# Run all benchmarks (takes ~10-15 minutes)
./run_rag_benchmarks.sh              # Linux/Mac
.\run_rag_benchmarks.ps1             # Windows

# Run specific benchmark (fast, ~30 seconds)
./quick_benchmark.sh embedding       # Test embeddings only
./quick_benchmark.sh hybrid          # Test hybrid search only
```

## Common Commands

| Task | Command |
|------|---------|
| All benchmarks | `./run_rag_benchmarks.sh` |
| Just embeddings | `./quick_benchmark.sh embedding` |
| Just search | `./quick_benchmark.sh vector` |
| Just hybrid | `./quick_benchmark.sh hybrid` |
| Just cache | `./quick_benchmark.sh cache` |
| Full pipeline | `./quick_benchmark.sh pipeline` |

## Manual Benchmark Execution

```bash
cd internal/services

# Basic benchmark
go test -bench=BenchmarkTextEmbedding -benchmem

# Longer benchmark for accuracy
go test -bench=BenchmarkHybridSearch -benchmem -benchtime=10s

# With profiling
go test -bench=BenchmarkVectorSearch -benchmem -cpuprofile=cpu.prof

# Parallel execution test
go test -bench=BenchmarkTextEmbeddingParallel -benchmem
```

## Available Benchmarks

### Embedding Generation
- `BenchmarkTextEmbedding` - Text embedding performance
- `BenchmarkTextEmbeddingParallel` - Parallel embedding generation
- `BenchmarkImageEmbedding` - Image embedding performance

### Vector Search
- `BenchmarkVectorSearch` - Search by dataset size
- `BenchmarkVectorSearchLimits` - Search by result limit

### Hybrid Search
- `BenchmarkHybridSearch` - Overall hybrid search
- `BenchmarkHybridSearchWeights` - Different weight combinations

### Re-Ranking
- `BenchmarkReRanking` - Basic re-ranking
- `BenchmarkReRankingResultCounts` - Re-ranking different counts

### Chunking
- `BenchmarkSmartChunking` - Chunking by document size
- `BenchmarkChunkingSizes` - Different chunk sizes
- `BenchmarkChunkingStrategies` - Different strategies

### Caching
- `BenchmarkCacheHit` - Cache hit performance
- `BenchmarkCacheMiss` - Cache miss performance
- `BenchmarkCacheSet` - Cache write performance
- `BenchmarkHybridSearchWithCache` - Hybrid search with cache

### End-to-End
- `BenchmarkFullRAGPipeline` - Complete RAG workflow

## Performance Targets

| Operation | Target | Good | Needs Work |
|-----------|--------|------|------------|
| Text Embed (500 chars) | <100ms | <200ms | >200ms |
| Vector Search (1k docs) | <50ms | <100ms | >100ms |
| Hybrid Search | <150ms | <300ms | >300ms |
| Re-Ranking (25 results) | <50ms | <100ms | >100ms |
| Full Pipeline | <300ms | <500ms | >500ms |
| Cache Hit | <5ms | <10ms | >10ms |

## Understanding Results

```
BenchmarkTextEmbedding/Small-8    1000    1234567 ns/op    5678 B/op    123 allocs/op
                                  ^^^^    ^^^^^^^          ^^^^          ^^^
                                  iters   nanosec/op       bytes/op      allocs/op
```

- **ns/op**: Lower is better (1ms = 1,000,000 ns)
- **B/op**: Memory used per operation
- **allocs/op**: Heap allocations (lower = less GC pressure)

## Quick Profiling

```bash
# CPU Profile
go test -bench=BenchmarkHybridSearch -cpuprofile=cpu.prof
go tool pprof -http=:8080 cpu.prof

# Memory Profile
go test -bench=BenchmarkVectorSearch -memprofile=mem.prof
go tool pprof -http=:8080 mem.prof
```

## Comparing Benchmarks

```bash
# Before changes
go test -bench=. -benchmem > before.txt

# After changes
go test -bench=. -benchmem > after.txt

# Compare (requires benchstat)
benchstat before.txt after.txt
```

## Environment Variables

```bash
# Change benchmark duration
BENCH_TIME=5s ./quick_benchmark.sh embedding

# Custom output location
OUTPUT_DIR=./my-reports ./run_rag_benchmarks.sh
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Benchmarks skip | Database/Redis not running |
| Slow benchmarks | Reduce BENCH_TIME |
| Inconsistent results | Run with `-benchtime=30s` |
| Out of memory | Reduce dataset sizes in tests |

## Output Locations

- **Full Report**: `docs/rag_performance_report.md`
- **Raw Results**: `internal/services/benchmark_results.txt`
- **CPU Profile**: `internal/services/cpu.prof`
- **Memory Profile**: `internal/services/mem.prof`

## Tips

1. Run benchmarks on idle system for accuracy
2. Close resource-intensive applications
3. Run multiple times and compare
4. Use `-count=10` for statistical accuracy
5. Profile hot paths to find bottlenecks

## Quick Checks

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Check PostgreSQL connection
psql -h localhost -U postgres -d businessos -c "SELECT 1"

# Check Redis connection
redis-cli ping

# Verify pgvector extension
psql -h localhost -U postgres -d businessos -c "SELECT * FROM pg_extension WHERE extname='vector'"
```

## CI/CD Integration

```yaml
# .github/workflows/benchmarks.yml
- name: Run Benchmarks
  run: |
    cd desktop/backend-go/internal/services
    go test -bench=. -benchmem -benchtime=5s
```

## Resources

- Full Documentation: `BENCHMARKS.md`
- Report Template: `docs/rag_performance_report_template.md`
- Benchmark Code: `internal/services/rag_benchmarks_test.go`

---

**Need help?** Check `BENCHMARKS.md` for detailed documentation.
