#!/bin/bash

# Quick RAG Benchmark Runner
# Runs a specific benchmark or set of benchmarks quickly

if [ $# -eq 0 ]; then
    echo "Usage: ./quick_benchmark.sh <benchmark_name>"
    echo ""
    echo "Available benchmarks:"
    echo "  embedding     - Test text embedding generation"
    echo "  vector        - Test vector search"
    echo "  hybrid        - Test hybrid search"
    echo "  rerank        - Test re-ranking"
    echo "  chunk         - Test document chunking"
    echo "  cache         - Test caching"
    echo "  pipeline      - Test full RAG pipeline"
    echo "  all           - Run all benchmarks (use with caution)"
    echo ""
    echo "Examples:"
    echo "  ./quick_benchmark.sh embedding"
    echo "  ./quick_benchmark.sh hybrid"
    exit 1
fi

BENCHMARK=$1
BENCH_TIME=${BENCH_TIME:-3s}

cd internal/services

case $BENCHMARK in
    embedding)
        echo "Running Text Embedding Benchmarks..."
        go test -bench=BenchmarkTextEmbedding -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    vector)
        echo "Running Vector Search Benchmarks..."
        go test -bench=BenchmarkVectorSearch -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    hybrid)
        echo "Running Hybrid Search Benchmarks..."
        go test -bench=BenchmarkHybridSearch -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    rerank)
        echo "Running Re-Ranking Benchmarks..."
        go test -bench=BenchmarkReRanking -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    chunk)
        echo "Running Chunking Benchmarks..."
        go test -bench=BenchmarkChunking -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    cache)
        echo "Running Cache Benchmarks..."
        go test -bench=BenchmarkCache -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    pipeline)
        echo "Running Full Pipeline Benchmark..."
        go test -bench=BenchmarkFullRAGPipeline -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    all)
        echo "Running ALL Benchmarks (this will take a while)..."
        go test -bench=. -benchmem -benchtime=$BENCH_TIME -run=^$
        ;;
    *)
        echo "Unknown benchmark: $BENCHMARK"
        echo "Run './quick_benchmark.sh' without arguments to see available options"
        exit 1
        ;;
esac

cd ../..
