#!/bin/bash
# Test pm4py-rust integration with BusinessOS BOS Gateway
# This script verifies that the gateway correctly calls pm4py-rust endpoints

set -e

echo "======================================"
echo "Testing pm4py-rust BOS Gateway"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default pm4py-rust URL
PM4PY_URL="${PM4PY_RUST_URL:-http://localhost:8090}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Smoke Test 1: Discover Endpoint"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

DISCOVER_RESPONSE=$(curl -s -X POST http://localhost:8001/api/bos/discover \
  -H "Content-Type: application/json" \
  -d '{"log_path": "/path/to/log.xes", "algorithm": "inductive_miner"}' 2>/dev/null || echo "")

if [ -z "$DISCOVER_RESPONSE" ]; then
  echo -e "${YELLOW}⚠ SKIPPED${NC} - BusinessOS not running on :8001"
  echo "  Start with: make dev"
else
  echo "Response received:"
  echo "$DISCOVER_RESPONSE" | jq . 2>/dev/null || echo "$DISCOVER_RESPONSE"

  # Verify response has expected fields
  if echo "$DISCOVER_RESPONSE" | jq -e '.model_id' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - model_id present"
  else
    echo -e "${RED}✗ FAIL${NC} - model_id missing"
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Smoke Test 2: Conformance Endpoint"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

CONFORMANCE_RESPONSE=$(curl -s -X POST http://localhost:8001/api/bos/conformance \
  -H "Content-Type: application/json" \
  -d '{"log_path": "/path/to/log.xes", "model_id": "petri_net_123"}' 2>/dev/null || echo "")

if [ -z "$CONFORMANCE_RESPONSE" ]; then
  echo -e "${YELLOW}⚠ SKIPPED${NC} - BusinessOS not running on :8001"
else
  echo "Response received:"
  echo "$CONFORMANCE_RESPONSE" | jq . 2>/dev/null || echo "$CONFORMANCE_RESPONSE"

  if echo "$CONFORMANCE_RESPONSE" | jq -e '.fitness' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - fitness metric present"
  else
    echo -e "${RED}✗ FAIL${NC} - fitness metric missing"
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Smoke Test 3: Statistics Endpoint"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

STATISTICS_RESPONSE=$(curl -s -X POST http://localhost:8001/api/bos/statistics \
  -H "Content-Type: application/json" \
  -d '{"log_path": "/path/to/log.xes"}' 2>/dev/null || echo "")

if [ -z "$STATISTICS_RESPONSE" ]; then
  echo -e "${YELLOW}⚠ SKIPPED${NC} - BusinessOS not running on :8001"
else
  echo "Response received:"
  echo "$STATISTICS_RESPONSE" | jq . 2>/dev/null || echo "$STATISTICS_RESPONSE"

  if echo "$STATISTICS_RESPONSE" | jq -e '.num_traces' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC} - num_traces present"
  else
    echo -e "${RED}✗ FAIL${NC} - num_traces missing"
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "To run integration tests:"
echo "  1. Start BusinessOS: make dev"
echo "  2. Start pm4py-rust: cd pm4py-rust && cargo run --example http_server"
echo "  3. Run smoke tests: bash scripts/test-pm4py-integration.sh"
echo ""
