#!/bin/bash

#############################################################################
# COMPREHENSIVE API TEST SCRIPT
# Tests all critical endpoints in BusinessOS backend
#
# Usage:
#   ./test_all_endpoints.sh [BASE_URL] [AUTH_TOKEN]
#
# Examples:
#   ./test_all_endpoints.sh http://localhost:8001
#   ./test_all_endpoints.sh http://localhost:8001 "Bearer eyJhbGc..."
#
# Requirements:
#   - curl, jq (JSON processor)
#   - Running BusinessOS backend server
#   - Valid authentication token (optional for public endpoints)
#############################################################################

set -e

# ============================================================================
# CONFIGURATION
# ============================================================================

BASE_URL="${1:-http://localhost:8001}"
AUTH_TOKEN="${2:-}"
API_BASE="${BASE_URL}/api"

# Color codes for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Generated test IDs
TEST_WORKSPACE_ID=""
TEST_CONVERSATION_ID=""
TEST_CONTEXT_ID=""
TEST_PROJECT_ID=""
TEST_CLIENT_ID=""

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

# Print section header
section() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Test case: make curl request and validate response
test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="${5:-200}"
    local description="${6:-}"

    ((TESTS_TOTAL++))

    local url="${API_BASE}${endpoint}"
    local curl_args=(-s -w "\n%{http_code}" -X "$method" "$url")

    # Add auth header if token provided
    if [ -n "$AUTH_TOKEN" ]; then
        curl_args+=(-H "Authorization: $AUTH_TOKEN")
    fi

    curl_args+=(-H "Content-Type: application/json")

    # Add data if provided
    if [ -n "$data" ]; then
        curl_args+=(-d "$data")
    fi

    local response=$(curl "${curl_args[@]}")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)

    # Check status code
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $name (HTTP $http_code)"
        if [ -n "$description" ]; then
            echo "   └─ $description"
        fi
        ((TESTS_PASSED++))
        echo "$body"
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $name (Expected $expected_status, got $http_code)"
        if [ -n "$description" ]; then
            echo "   └─ $description"
        fi
        echo "   Response: $body"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Skip test with reason
skip_test() {
    local name="$1"
    local reason="$2"

    ((TESTS_TOTAL++))
    ((TESTS_SKIPPED++))

    echo -e "${YELLOW}⊘ SKIP${NC}: $name"
    if [ -n "$reason" ]; then
        echo "   └─ $reason"
    fi
}

# Extract JSON field from response
extract_field() {
    echo "$1" | jq -r "$2 // empty"
}

# ============================================================================
# PRE-TEST CHECKS
# ============================================================================

section "PRE-TEST VALIDATION"

echo "Configuration:"
echo "  Base URL: $BASE_URL"
echo "  API Base: $API_BASE"
echo "  Auth Token: ${AUTH_TOKEN:0:20}${AUTH_TOKEN:+...}"
echo ""

# Check server health
echo "1. Checking server health..."
HEALTH=$(curl -s "$BASE_URL/health")
if echo "$HEALTH" | jq -e '.status' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Server is healthy${NC}"
    echo "  Response: $HEALTH"
else
    echo -e "${RED}✗ Server is not responding${NC}"
    echo "  Please start the server with: ./bin/server.exe"
    exit 1
fi

echo ""
echo "2. Checking API readiness..."
READY=$(curl -s "$BASE_URL/ready")
echo "  Status: $(echo "$READY" | jq -r '.status')"
echo "  Database: $(echo "$READY" | jq -r '.database')"
echo "  Redis: $(echo "$READY" | jq -r '.redis')"

if [ -z "$AUTH_TOKEN" ]; then
    echo ""
    echo -e "${YELLOW}⚠ Warning: No auth token provided${NC}"
    echo "  Some endpoints require authentication."
    echo "  Pass token as second argument: $0 $BASE_URL 'Bearer YOUR_TOKEN'"
fi

# ============================================================================
# PUBLIC ENDPOINTS (No Auth Required)
# ============================================================================

section "PUBLIC ENDPOINTS (No Auth Required)"

echo "1. Health Check"
test_endpoint "Server Health" "GET" "/health" "" "200" "Basic health endpoint"

echo ""
echo "2. Readiness Check"
test_endpoint "Server Readiness" "GET" "/ready" "" "200" "Includes dependency status"

# ============================================================================
# CHAT ENDPOINTS
# ============================================================================

section "CHAT ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Create Conversation" "No auth token provided"
else
    echo "1. Creating test conversation..."
    CONV_DATA='{"title":"Test Conversation","description":"API test conversation"}'
    CONV_RESPONSE=$(curl -s -X POST "$API_BASE/chat/conversations" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$CONV_DATA")

    TEST_CONVERSATION_ID=$(extract_field "$CONV_RESPONSE" '.id')

    if [ -n "$TEST_CONVERSATION_ID" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Create Conversation"
        echo "   Conversation ID: $TEST_CONVERSATION_ID"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Conversation"
        echo "   Response: $CONV_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Conversations"
    test_endpoint "List Conversations" "GET" "/chat/conversations" "" "200" "Retrieve all conversations"

    echo ""
    if [ -n "$TEST_CONVERSATION_ID" ]; then
        echo "3. Get Conversation"
        test_endpoint "Get Conversation" "GET" "/chat/conversations/$TEST_CONVERSATION_ID" "" "200" "Retrieve specific conversation"
    fi
fi

# ============================================================================
# CONTEXT ENDPOINTS
# ============================================================================

section "CONTEXT ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Create Context" "No auth token provided"
else
    echo "1. Creating test context..."
    CONTEXT_DATA='{"title":"Test Context","content":"Test context content","type":"general"}'
    CONTEXT_RESPONSE=$(curl -s -X POST "$API_BASE/contexts" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$CONTEXT_DATA")

    TEST_CONTEXT_ID=$(extract_field "$CONTEXT_RESPONSE" '.id')

    if [ -n "$TEST_CONTEXT_ID" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Create Context"
        echo "   Context ID: $TEST_CONTEXT_ID"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Context"
        echo "   Response: $CONTEXT_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Contexts"
    test_endpoint "List Contexts" "GET" "/contexts" "" "200" "Retrieve all contexts"
fi

# ============================================================================
# WORKSPACE ENDPOINTS
# ============================================================================

section "WORKSPACE ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Create Workspace" "No auth token provided"
else
    echo "1. Creating test workspace..."
    WORKSPACE_DATA='{"name":"Test Workspace","description":"API test workspace"}'
    WORKSPACE_RESPONSE=$(curl -s -X POST "$API_BASE/workspaces" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$WORKSPACE_DATA")

    TEST_WORKSPACE_ID=$(extract_field "$WORKSPACE_RESPONSE" '.id')

    if [ -n "$TEST_WORKSPACE_ID" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Create Workspace"
        echo "   Workspace ID: $TEST_WORKSPACE_ID"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Workspace"
        echo "   Response: $WORKSPACE_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Workspaces"
    test_endpoint "List Workspaces" "GET" "/workspaces" "" "200" "Retrieve all workspaces"

    echo ""
    if [ -n "$TEST_WORKSPACE_ID" ]; then
        echo "3. Get Workspace"
        test_endpoint "Get Workspace" "GET" "/workspaces/$TEST_WORKSPACE_ID" "" "200" "Retrieve specific workspace"

        echo ""
        echo "4. Get User Role Context"
        test_endpoint "Get Role Context" "GET" "/workspaces/$TEST_WORKSPACE_ID/role-context" "" "200" "Retrieve user's role and permissions in workspace"

        echo ""
        echo "5. List Workspace Members"
        test_endpoint "List Workspace Members" "GET" "/workspaces/$TEST_WORKSPACE_ID/members" "" "200" "Retrieve workspace members"
    fi
fi

# ============================================================================
# WORKSPACE MEMORY ENDPOINTS (CUS-25)
# ============================================================================

section "WORKSPACE MEMORY ENDPOINTS"

if [ -z "$AUTH_TOKEN" ] || [ -z "$TEST_WORKSPACE_ID" ]; then
    skip_test "Create Workspace Memory" "No auth token or workspace ID"
else
    echo "1. Creating workspace memory..."
    MEMORY_DATA='{"title":"Test Memory","content":"Important workspace information","visibility":"workspace"}'
    MEMORY_RESPONSE=$(curl -s -X POST "$API_BASE/workspaces/$TEST_WORKSPACE_ID/memories" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$MEMORY_DATA")

    if echo "$MEMORY_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}: Create Workspace Memory"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Workspace Memory"
        echo "   Response: $MEMORY_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Workspace Memories"
    test_endpoint "List Workspace Memories" "GET" "/workspaces/$TEST_WORKSPACE_ID/memories" "" "200" "Retrieve all workspace memories"

    echo ""
    echo "3. List Private Memories"
    test_endpoint "List Private Memories" "GET" "/workspaces/$TEST_WORKSPACE_ID/memories/private" "" "200" "Retrieve user's private memories in workspace"
fi

# ============================================================================
# RAG SEARCH ENDPOINTS
# ============================================================================

section "RAG SEARCH ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Hybrid Search" "No auth token provided"
else
    echo "1. Hybrid Search (Semantic + Keyword)"
    HYBRID_DATA='{"query":"test search","semantic_weight":0.7,"keyword_weight":0.3}'
    test_endpoint "Hybrid Search" "POST" "/search/hybrid" "$HYBRID_DATA" "200" "Combined semantic and keyword search"

    echo ""
    echo "2. Hybrid Search Explain"
    test_endpoint "Search Explain" "GET" "/search/explain?query=test" "" "200" "Explain search scoring methodology"
fi

# ============================================================================
# MULTIMODAL ENDPOINTS
# ============================================================================

section "MULTIMODAL ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Get Supported Modalities" "No auth token provided"
else
    echo "1. Get Supported Modalities"
    test_endpoint "Supported Modalities" "GET" "/search/modalities" "" "200" "List supported search modalities (text, image)"
fi

# ============================================================================
# PROJECTS ENDPOINTS
# ============================================================================

section "PROJECTS ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Create Project" "No auth token provided"
else
    echo "1. Creating test project..."
    PROJECT_DATA='{"title":"Test Project","description":"API test project","status":"active"}'
    PROJECT_RESPONSE=$(curl -s -X POST "$API_BASE/projects" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$PROJECT_DATA")

    TEST_PROJECT_ID=$(extract_field "$PROJECT_RESPONSE" '.id')

    if [ -n "$TEST_PROJECT_ID" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Create Project"
        echo "   Project ID: $TEST_PROJECT_ID"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Project"
        echo "   Response: $PROJECT_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Projects"
    test_endpoint "List Projects" "GET" "/projects" "" "200" "Retrieve all projects"

    echo ""
    echo "3. Get Project Statistics"
    test_endpoint "Project Stats" "GET" "/projects/stats" "" "200" "Project statistics and metrics"
fi

# ============================================================================
# CLIENTS ENDPOINTS
# ============================================================================

section "CLIENTS ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Create Client" "No auth token provided"
else
    echo "1. Creating test client..."
    CLIENT_DATA='{"name":"Test Client","email":"test@example.com","status":"active"}'
    CLIENT_RESPONSE=$(curl -s -X POST "$API_BASE/clients" \
        -H "Authorization: $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$CLIENT_DATA")

    TEST_CLIENT_ID=$(extract_field "$CLIENT_RESPONSE" '.id')

    if [ -n "$TEST_CLIENT_ID" ]; then
        echo -e "${GREEN}✓ PASS${NC}: Create Client"
        echo "   Client ID: $TEST_CLIENT_ID"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: Create Client"
        echo "   Response: $CLIENT_RESPONSE"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))

    echo ""
    echo "2. List Clients"
    test_endpoint "List Clients" "GET" "/clients" "" "200" "Retrieve all clients"
fi

# ============================================================================
# SEARCH ENDPOINTS
# ============================================================================

section "SEARCH ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Web Search" "No auth token provided"
else
    echo "1. Web Search"
    test_endpoint "Web Search" "GET" "/search/web?q=test" "" "200" "Search the web"

    echo ""
    echo "2. Search History"
    test_endpoint "Search History" "GET" "/search/history" "" "200" "Retrieve search history"
fi

# ============================================================================
# SETTINGS ENDPOINTS
# ============================================================================

section "SETTINGS ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Get Settings" "No auth token provided"
else
    echo "1. Get User Settings"
    test_endpoint "Get Settings" "GET" "/settings" "" "200" "Retrieve user settings"

    echo ""
    echo "2. Get System Settings"
    test_endpoint "System Settings" "GET" "/settings/system" "" "200" "Retrieve system-wide settings"

    echo ""
    echo "3. Get Full State"
    test_endpoint "Full State" "GET" "/settings/full-state" "" "200" "Complete UI state for synchronization"
fi

# ============================================================================
# THINKING/COT ENDPOINTS
# ============================================================================

section "THINKING & REASONING ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "List Thinking Traces" "No auth token provided"
else
    if [ -n "$TEST_CONVERSATION_ID" ]; then
        echo "1. List Thinking Traces"
        test_endpoint "Thinking Traces" "GET" "/thinking/traces/$TEST_CONVERSATION_ID" "" "200" "Retrieve chain-of-thought traces"
    else
        skip_test "List Thinking Traces" "No test conversation created"
    fi

    echo ""
    echo "2. Get Thinking Settings"
    test_endpoint "Thinking Settings" "GET" "/thinking/settings" "" "200" "Retrieve thinking/COT settings"

    echo ""
    echo "3. List Reasoning Templates"
    test_endpoint "Reasoning Templates" "GET" "/reasoning/templates" "" "200" "Retrieve available reasoning templates"
fi

# ============================================================================
# DASHBOARD ENDPOINTS
# ============================================================================

section "DASHBOARD ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "Get Dashboard Summary" "No auth token provided"
else
    echo "1. Get Dashboard Summary"
    test_endpoint "Dashboard Summary" "GET" "/dashboard/summary" "" "200" "Retrieve dashboard overview"

    echo ""
    echo "2. List Focus Items"
    test_endpoint "Focus Items" "GET" "/dashboard/focus" "" "200" "Retrieve focus items"

    echo ""
    echo "3. List Dashboard Tasks"
    test_endpoint "Dashboard Tasks" "GET" "/dashboard/tasks" "" "200" "Retrieve task list"
fi

# ============================================================================
# TEAM ENDPOINTS
# ============================================================================

section "TEAM ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "List Team Members" "No auth token provided"
else
    echo "1. List Team Members"
    test_endpoint "Team Members" "GET" "/team" "" "200" "Retrieve team members"
fi

# ============================================================================
# ARTIFACTS ENDPOINTS
# ============================================================================

section "ARTIFACTS ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "List Artifacts" "No auth token provided"
else
    echo "1. List Artifacts"
    test_endpoint "List Artifacts" "GET" "/artifacts" "" "200" "Retrieve all artifacts"
fi

# ============================================================================
# NODES ENDPOINTS
# ============================================================================

section "NODES ENDPOINTS"

if [ -z "$AUTH_TOKEN" ]; then
    skip_test "List Nodes" "No auth token provided"
else
    echo "1. List Nodes"
    test_endpoint "List Nodes" "GET" "/nodes" "" "200" "Retrieve all nodes"

    echo ""
    echo "2. Get Node Tree"
    test_endpoint "Node Tree" "GET" "/nodes/tree" "" "200" "Retrieve hierarchical node structure"
fi

# ============================================================================
# TEST SUMMARY
# ============================================================================

section "TEST SUMMARY"

echo "Total Tests:   $TESTS_TOTAL"
echo -e "Passed:        ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:        ${RED}$TESTS_FAILED${NC}"
echo -e "Skipped:       ${YELLOW}$TESTS_SKIPPED${NC}"
echo ""

PASS_RATE=$(( (TESTS_PASSED * 100) / TESTS_TOTAL ))
echo "Pass Rate:     $PASS_RATE%"

echo ""
echo "Generated Test IDs:"
[ -n "$TEST_WORKSPACE_ID" ] && echo "  Workspace:     $TEST_WORKSPACE_ID"
[ -n "$TEST_CONVERSATION_ID" ] && echo "  Conversation:  $TEST_CONVERSATION_ID"
[ -n "$TEST_CONTEXT_ID" ] && echo "  Context:       $TEST_CONTEXT_ID"
[ -n "$TEST_PROJECT_ID" ] && echo "  Project:       $TEST_PROJECT_ID"
[ -n "$TEST_CLIENT_ID" ] && echo "  Client:        $TEST_CLIENT_ID"

echo ""
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo ""
    echo "Troubleshooting:"
    echo "  1. Check server logs: tail -f logs/server.log"
    echo "  2. Verify database connection: psql $DATABASE_URL -c 'SELECT 1'"
    echo "  3. Check auth token validity: $0 $BASE_URL 'Bearer YOUR_TOKEN'"
    exit 1
fi
