#!/bin/bash

# Quick Test Script for Q1 + RAG Features
# Tests the main endpoints we implemented

BASE_URL="http://localhost:8001"
API_BASE="${BASE_URL}/api"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC} Quick Test - Q1 + RAG Features"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Test counter
PASSED=0
FAILED=0

test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"

    echo -n "Testing: $name... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        echo -e "${GREEN}✓ PASS${NC} (Status: $status_code)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Status: $status_code)"
        echo "  Response: $body"
        ((FAILED++))
        return 1
    fi
}

echo ""
echo -e "${BLUE}═══ 1. Server Health${NC}"
test_endpoint "Health Check" "GET" "${BASE_URL}/health"

echo ""
echo -e "${BLUE}═══ 2. Workspace Operations (Public)${NC}"

# Create test workspace
echo -n "Creating test workspace... "
response=$(curl -s -X POST "${API_BASE}/workspaces" \
    -H "Content-Type: application/json" \
    -d '{"name":"Test Workspace","slug":"test-ws-'$(date +%s)'"}')

if echo "$response" | /tmp/jq.exe -e '.id' > /dev/null 2>&1; then
    WORKSPACE_ID=$(echo "$response" | /tmp/jq.exe -r '.id')
    echo -e "${GREEN}✓ Created${NC} (ID: ${WORKSPACE_ID:0:8}...)"
    ((PASSED++))
else
    echo -e "${RED}✗ Failed${NC}"
    echo "  Response: $response"
    ((FAILED++))
    WORKSPACE_ID=""
fi

if [ -n "$WORKSPACE_ID" ]; then
    echo ""
    echo -e "${BLUE}═══ 3. Memory Hierarchy (Q1 Feature)${NC}"

    # Create workspace memory
    test_endpoint "Create Workspace Memory" "POST" \
        "${API_BASE}/workspaces/${WORKSPACE_ID}/memories" \
        '{"title":"Test Decision","content":"Use PostgreSQL","memory_type":"decision","visibility":"workspace"}'

    # List workspace memories
    test_endpoint "List Workspace Memories" "GET" \
        "${API_BASE}/workspaces/${WORKSPACE_ID}/memories"

    # List private memories
    test_endpoint "List Private Memories" "GET" \
        "${API_BASE}/workspaces/${WORKSPACE_ID}/memories/private"

    # List accessible memories
    test_endpoint "List Accessible Memories" "GET" \
        "${API_BASE}/workspaces/${WORKSPACE_ID}/memories/accessible"

    echo ""
    echo -e "${BLUE}═══ 4. Role Context (Q1 Feature)${NC}"

    # Get role context
    test_endpoint "Get Role Context" "GET" \
        "${API_BASE}/workspaces/${WORKSPACE_ID}/role-context"
fi

echo ""
echo -e "${BLUE}═══ 5. Project Access (Q1 Feature)${NC}"

# Create test project
echo -n "Creating test project... "
response=$(curl -s -X POST "${API_BASE}/projects" \
    -H "Content-Type: application/json" \
    -d '{"name":"Test Project","workspace_id":"'${WORKSPACE_ID}'"}')

if echo "$response" | /tmp/jq.exe -e '.id' > /dev/null 2>&1; then
    PROJECT_ID=$(echo "$response" | /tmp/jq.exe -r '.id')
    echo -e "${GREEN}✓ Created${NC} (ID: ${PROJECT_ID:0:8}...)"
    ((PASSED++))

    # Test project members endpoint (may fail without auth)
    test_endpoint "List Project Members" "GET" \
        "${API_BASE}/projects/${PROJECT_ID}/members"
else
    echo -e "${YELLOW}⊘ Skipped${NC} (Project creation requires auth)"
fi

echo ""
echo -e "${BLUE}═══ 6. RAG Features${NC}"

# Test RAG explain
test_endpoint "RAG Search Explain" "GET" \
    "${API_BASE}/rag/search/explain?query=test"

# Test RAG modalities
test_endpoint "RAG Modalities" "GET" \
    "${API_BASE}/rag/search/modalities"

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${NC} TEST SUMMARY"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}Passed:${NC} $PASSED"
echo -e "  ${RED}Failed:${NC} $FAILED"
echo -e "  ${BLUE}Total:${NC}  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
