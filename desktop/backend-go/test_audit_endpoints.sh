#!/bin/bash

# Workspace Audit Endpoints Test Script
# This script tests all 6 audit log endpoints

BASE_URL="${API_BASE_URL:-http://localhost:8080}"
WORKSPACE_ID="${TEST_WORKSPACE_ID:-your-workspace-id}"
USER_ID="${TEST_USER_ID:-your-user-id}"
TOKEN="${AUTH_TOKEN:-your-auth-token}"

echo "======================================"
echo "Workspace Audit Endpoints Test"
echo "======================================"
echo "Base URL: $BASE_URL"
echo "Workspace ID: $WORKSPACE_ID"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: List Audit Logs (with pagination)
echo -e "${YELLOW}Test 1: List Audit Logs${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs?limit=10&offset=0" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 2: List Audit Logs (with filters)
echo -e "${YELLOW}Test 2: List Audit Logs (filtered by action)${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs?action=create&limit=5" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 3: List Audit Logs (with date range)
START_DATE=$(date -u -d '30 days ago' +"%Y-%m-%dT%H:%M:%SZ")
END_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo -e "${YELLOW}Test 3: List Audit Logs (date range)${NC}"
echo "Start: $START_DATE"
echo "End: $END_DATE"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs?start_date=${START_DATE}&end_date=${END_DATE}&limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 4: Get User Activity
echo -e "${YELLOW}Test 4: Get User Activity${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs/user/${USER_ID}?limit=20" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 5: Get Resource History
echo -e "${YELLOW}Test 5: Get Resource History (workspace)${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs/resource/workspace/${WORKSPACE_ID}?limit=10" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 6: Get Action Statistics
echo -e "${YELLOW}Test 6: Get Action Statistics${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs/stats/actions?start_date=${START_DATE}&end_date=${END_DATE}" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 7: Get Most Active Users
echo -e "${YELLOW}Test 7: Get Most Active Users (top 5)${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs/stats/active-users?limit=5&start_date=${START_DATE}&end_date=${END_DATE}" \
  -H "Authorization: Bearer ${TOKEN}" | jq '.'
echo ""

# Test 8: Get Specific Audit Log (needs a log ID from previous test)
LOG_ID=$(curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs?limit=1" \
  -H "Authorization: Bearer ${TOKEN}" | jq -r '.logs[0].id')

if [ "$LOG_ID" != "null" ] && [ -n "$LOG_ID" ]; then
  echo -e "${YELLOW}Test 8: Get Specific Audit Log${NC}"
  echo "Log ID: $LOG_ID"
  curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs/${LOG_ID}" \
    -H "Authorization: Bearer ${TOKEN}" | jq '.'
  echo ""
else
  echo -e "${RED}Test 8: Skipped (no audit logs found)${NC}"
  echo ""
fi

# Test 9: Permission Check (should fail without admin role)
echo -e "${YELLOW}Test 9: Permission Check (expect 403 without admin)${NC}"
curl -s -X GET "${BASE_URL}/api/workspaces/${WORKSPACE_ID}/audit-logs" \
  -H "Authorization: Bearer invalid-token" | jq '.'
echo ""

echo -e "${GREEN}======================================"
echo "All tests completed!"
echo "======================================${NC}"
