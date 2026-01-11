#!/bin/bash

# Custom Agents Backend Testing Suite
# 20 comprehensive tests for Custom Agents endpoints

BASE_URL="http://localhost:8001"
TOTAL_TESTS=20
PASSED=0
FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test result tracker
declare -a TEST_RESULTS

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║       CUSTOM AGENTS BACKEND TEST SUITE - 20 TESTS                ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Testing backend at: $BASE_URL"
echo "Start time: $(date)"
echo ""

# Helper function to run test
run_test() {
    local test_num=$1
    local test_name=$2
    local method=$3
    local endpoint=$4
    local expected_code=$5
    local data=$6

    echo -n "Test $test_num: $test_name ... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        ((PASSED++))
        TEST_RESULTS[$test_num]="✓ PASS | $test_name | HTTP $http_code"
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_code, got $http_code)"
        ((FAILED++))
        TEST_RESULTS[$test_num]="✗ FAIL | $test_name | Expected $expected_code, got $http_code"
    fi

    # Save response for later inspection if needed
    echo "$body" > "/tmp/test_${test_num}_response.json" 2>/dev/null
}

echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 1: CRUD OPERATIONS (5 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 1: List all custom agents
run_test 1 "GET /api/ai/custom-agents (List agents)" "GET" "/api/ai/custom-agents" "200"

# Test 2: Create custom agent
CREATE_DATA='{
  "name": "test-agent-'$(date +%s)'",
  "display_name": "Test Agent",
  "description": "Test agent created by automated test",
  "category": "custom",
  "system_prompt": "You are a helpful test agent.",
  "is_active": true
}'
run_test 2 "POST /api/ai/custom-agents (Create agent)" "POST" "/api/ai/custom-agents" "201" "$CREATE_DATA"

# Extract created agent ID from response (if successful)
if [ -f "/tmp/test_2_response.json" ]; then
    CREATED_AGENT_ID=$(grep -o '"id":"[^"]*"' /tmp/test_2_response.json | head -1 | cut -d'"' -f4)
fi

# Test 3: Get single agent (using created ID if available, otherwise use placeholder)
if [ -n "$CREATED_AGENT_ID" ]; then
    run_test 3 "GET /api/ai/custom-agents/:id (Get agent)" "GET" "/api/ai/custom-agents/$CREATED_AGENT_ID" "200"
else
    echo "Test 3: GET /api/ai/custom-agents/:id ... ${YELLOW}SKIP${NC} (No agent created)"
    TEST_RESULTS[3]="⊘ SKIP | GET /api/ai/custom-agents/:id | No agent created in test 2"
fi

# Test 4: Update agent
if [ -n "$CREATED_AGENT_ID" ]; then
    UPDATE_DATA='{
      "display_name": "Updated Test Agent",
      "description": "Updated by automated test"
    }'
    run_test 4 "PUT /api/ai/custom-agents/:id (Update agent)" "PUT" "/api/ai/custom-agents/$CREATED_AGENT_ID" "200" "$UPDATE_DATA"
else
    echo "Test 4: PUT /api/ai/custom-agents/:id ... ${YELLOW}SKIP${NC} (No agent to update)"
    TEST_RESULTS[4]="⊘ SKIP | PUT /api/ai/custom-agents/:id | No agent to update"
fi

# Test 5: Delete agent (save for last to cleanup)
# Will run at the end

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 2: FILTERING & QUERY (2 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 6: List agents with include_inactive
run_test 6 "GET /api/ai/custom-agents?include_inactive=true" "GET" "/api/ai/custom-agents?include_inactive=true" "200"

# Test 7: Get agents by category
run_test 7 "GET /api/ai/custom-agents/category/:category" "GET" "/api/ai/custom-agents/category/custom" "200"

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 3: PRESETS (3 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 8: List presets
run_test 8 "GET /api/ai/agents/presets (List presets)" "GET" "/api/ai/agents/presets" "200"

# Extract first preset ID
if [ -f "/tmp/test_8_response.json" ]; then
    PRESET_ID=$(grep -o '"id":"[^"]*"' /tmp/test_8_response.json | head -1 | cut -d'"' -f4)
fi

# Test 9: Get preset detail
if [ -n "$PRESET_ID" ]; then
    run_test 9 "GET /api/ai/agents/presets/:id (Get preset)" "GET" "/api/ai/agents/presets/$PRESET_ID" "200"
else
    echo "Test 9: GET /api/ai/agents/presets/:id ... ${YELLOW}SKIP${NC} (No presets found)"
    TEST_RESULTS[9]="⊘ SKIP | GET /api/ai/agents/presets/:id | No presets found"
fi

# Test 10: Create agent from preset
if [ -n "$PRESET_ID" ]; then
    PRESET_DATA='{"name": "agent-from-preset-'$(date +%s)'"}'
    run_test 10 "POST /api/ai/custom-agents/from-preset/:id" "POST" "/api/ai/custom-agents/from-preset/$PRESET_ID" "201" "$PRESET_DATA"

    # Extract created agent ID for cleanup
    if [ -f "/tmp/test_10_response.json" ]; then
        PRESET_AGENT_ID=$(grep -o '"id":"[^"]*"' /tmp/test_10_response.json | head -1 | cut -d'"' -f4)
    fi
else
    echo "Test 10: POST /api/ai/custom-agents/from-preset/:id ... ${YELLOW}SKIP${NC} (No preset available)"
    TEST_RESULTS[10]="⊘ SKIP | POST /api/ai/custom-agents/from-preset/:id | No preset available"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 4: TESTING & SANDBOX (2 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 11: Test agent (expects SSE stream, so 200 is success)
if [ -n "$CREATED_AGENT_ID" ]; then
    TEST_MESSAGE='{"message": "Hello, this is a test"}'
    run_test 11 "POST /api/ai/custom-agents/:id/test" "POST" "/api/ai/custom-agents/$CREATED_AGENT_ID/test" "200" "$TEST_MESSAGE"
else
    echo "Test 11: POST /api/ai/custom-agents/:id/test ... ${YELLOW}SKIP${NC} (No agent available)"
    TEST_RESULTS[11]="⊘ SKIP | POST /api/ai/custom-agents/:id/test | No agent available"
fi

# Test 12: Sandbox test
SANDBOX_DATA='{
  "system_prompt": "You are a test assistant.",
  "message": "Hello from sandbox test",
  "model": "gpt-4",
  "temperature": 0.7
}'
run_test 12 "POST /api/ai/custom-agents/sandbox" "POST" "/api/ai/custom-agents/sandbox" "200" "$SANDBOX_DATA"

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 5: GENERAL AGENTS (2 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 13: List all agents (built-in + custom)
run_test 13 "GET /api/ai/agents (All agents)" "GET" "/api/ai/agents" "200"

# Test 14: Get specific agent
if [ -n "$CREATED_AGENT_ID" ]; then
    run_test 14 "GET /api/ai/agents/:id" "GET" "/api/ai/agents/$CREATED_AGENT_ID" "200"
else
    echo "Test 14: GET /api/ai/agents/:id ... ${YELLOW}SKIP${NC} (No agent available)"
    TEST_RESULTS[14]="⊘ SKIP | GET /api/ai/agents/:id | No agent available"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CATEGORY 6: ERROR HANDLING & EDGE CASES (6 tests)"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 15: Get nonexistent agent (404)
run_test 15 "GET nonexistent agent (404 test)" "GET" "/api/ai/custom-agents/nonexistent-id-999" "404"

# Test 16: Create agent with invalid data (400)
INVALID_DATA='{"invalid": "data", "missing": "required_fields"}'
run_test 16 "POST invalid agent data (400 test)" "POST" "/api/ai/custom-agents" "400" "$INVALID_DATA"

# Test 17: Update nonexistent agent (404)
run_test 17 "PUT nonexistent agent (404 test)" "PUT" "/api/ai/custom-agents/nonexistent-id-999" "404" '{"display_name": "test"}'

# Test 18: Delete nonexistent agent (404)
run_test 18 "DELETE nonexistent agent (404 test)" "DELETE" "/api/ai/custom-agents/nonexistent-id-999" "404"

# Test 19: Test agent without message (400)
if [ -n "$CREATED_AGENT_ID" ]; then
    run_test 19 "POST test without message (400 test)" "POST" "/api/ai/custom-agents/$CREATED_AGENT_ID/test" "400" '{}'
else
    echo "Test 19: POST test without message ... ${YELLOW}SKIP${NC} (No agent available)"
    TEST_RESULTS[19]="⊘ SKIP | POST test without message | No agent available"
fi

# Test 20: Get agents by invalid category (should return empty or 200)
run_test 20 "GET invalid category (edge case)" "GET" "/api/ai/custom-agents/category/invalid-category-xyz" "200"

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  CLEANUP"
echo "════════════════════════════════════════════════════════════════════"
echo ""

# Test 5 (delayed): Delete created agent
if [ -n "$CREATED_AGENT_ID" ]; then
    run_test 5 "DELETE /api/ai/custom-agents/:id (Delete agent)" "DELETE" "/api/ai/custom-agents/$CREATED_AGENT_ID" "200"
else
    echo "Test 5: DELETE /api/ai/custom-agents/:id ... ${YELLOW}SKIP${NC} (No agent to delete)"
    TEST_RESULTS[5]="⊘ SKIP | DELETE /api/ai/custom-agents/:id | No agent to delete"
fi

# Delete preset-created agent if exists
if [ -n "$PRESET_AGENT_ID" ]; then
    echo -n "Cleanup: Deleting preset-created agent ... "
    curl -s -X DELETE "$BASE_URL/api/ai/custom-agents/$PRESET_AGENT_ID" > /dev/null
    echo -e "${GREEN}Done${NC}"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  FINAL REPORT"
echo "════════════════════════════════════════════════════════════════════"
echo ""
echo "End time: $(date)"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

# Calculate percentage
if [ $TOTAL_TESTS -gt 0 ]; then
    PERCENTAGE=$((PASSED * 100 / TOTAL_TESTS))
    echo "Success Rate: $PERCENTAGE%"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════"
echo "  TEST RESULTS SUMMARY"
echo "════════════════════════════════════════════════════════════════════"
echo ""

for i in "${!TEST_RESULTS[@]}"; do
    echo "  $i. ${TEST_RESULTS[$i]}"
done

echo ""
echo "════════════════════════════════════════════════════════════════════"

# Exit with failure if any tests failed
if [ $FAILED -gt 0 ]; then
    exit 1
else
    exit 0
fi
