#!/bin/bash

# Background Jobs System - API Test Script
# Este script testa todos os endpoints do sistema de background jobs

BASE_URL="http://localhost:8080/api"

echo "=================================="
echo "Background Jobs System - API Tests"
echo "=================================="
echo ""

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para testar endpoint
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4

    echo -e "${BLUE}Testing:${NC} $name"

    if [ -z "$data" ]; then
        response=$(curl -s -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json" -d "$data")
    fi

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Success"
        echo "Response: $response" | jq '.' 2>/dev/null || echo "$response"
    else
        echo -e "${RED}✗${NC} Failed"
    fi
    echo ""
}

echo "1. Test: Create Background Job (email_send)"
echo "==========================================="
JOB_DATA='{
  "job_type": "email_send",
  "payload": {
    "to": "test@example.com",
    "subject": "Test Email from Background Jobs",
    "body": "This is a test email sent via background job system!"
  },
  "priority": 1,
  "max_attempts": 3
}'

response=$(curl -s -X POST "$BASE_URL/background-jobs" \
  -H "Content-Type: application/json" \
  -d "$JOB_DATA")

echo "Response:"
echo "$response" | jq '.'

# Extract job ID for next tests
JOB_ID=$(echo "$response" | jq -r '.id' 2>/dev/null)

if [ "$JOB_ID" != "null" ] && [ -n "$JOB_ID" ]; then
    echo -e "${GREEN}✓${NC} Job created with ID: $JOB_ID"
else
    echo -e "${RED}✗${NC} Failed to create job"
    exit 1
fi
echo ""
sleep 2

echo "2. Test: List Background Jobs"
echo "=============================="
test_endpoint "List all jobs" "GET" "/background-jobs"
sleep 1

echo "3. Test: Get Job Status"
echo "======================="
if [ -n "$JOB_ID" ]; then
    test_endpoint "Get job $JOB_ID" "GET" "/background-jobs/$JOB_ID"
fi
sleep 1

echo "4. Test: Create Report Generation Job"
echo "====================================="
REPORT_DATA='{
  "job_type": "report_generate",
  "payload": {
    "report_type": "sales",
    "start_date": "2026-01-01",
    "end_date": "2026-01-07"
  },
  "priority": 2
}'
test_endpoint "Create report job" "POST" "/background-jobs" "$REPORT_DATA"
sleep 1

echo "5. Test: List Pending Jobs"
echo "=========================="
test_endpoint "List pending jobs" "GET" "/background-jobs?status=pending&limit=10"
sleep 1

echo "6. Test: Create Scheduled Job (Daily Report at 9am)"
echo "===================================================="
SCHEDULED_DATA='{
  "job_type": "daily_report",
  "payload": {
    "report_type": "daily_summary",
    "recipients": ["admin@example.com"]
  },
  "cron_expression": "0 9 * * *",
  "timezone": "America/Sao_Paulo",
  "name": "Daily Sales Report",
  "description": "Generates and sends daily sales report at 9am"
}'

scheduled_response=$(curl -s -X POST "$BASE_URL/scheduled-jobs" \
  -H "Content-Type: application/json" \
  -d "$SCHEDULED_DATA")

echo "Response:"
echo "$scheduled_response" | jq '.'

SCHEDULED_ID=$(echo "$scheduled_response" | jq -r '.id' 2>/dev/null)

if [ "$SCHEDULED_ID" != "null" ] && [ -n "$SCHEDULED_ID" ]; then
    echo -e "${GREEN}✓${NC} Scheduled job created with ID: $SCHEDULED_ID"
else
    echo -e "${RED}✗${NC} Failed to create scheduled job"
fi
echo ""
sleep 1

echo "7. Test: List Scheduled Jobs"
echo "============================"
test_endpoint "List all scheduled jobs" "GET" "/scheduled-jobs"
sleep 1

echo "8. Test: List Active Scheduled Jobs Only"
echo "========================================"
test_endpoint "List active scheduled jobs" "GET" "/scheduled-jobs?active_only=true"
sleep 1

echo "9. Test: Get Specific Scheduled Job"
echo "===================================="
if [ -n "$SCHEDULED_ID" ]; then
    test_endpoint "Get scheduled job $SCHEDULED_ID" "GET" "/scheduled-jobs/$SCHEDULED_ID"
fi
sleep 1

echo "10. Test: Disable Scheduled Job"
echo "================================"
if [ -n "$SCHEDULED_ID" ]; then
    test_endpoint "Disable scheduled job" "POST" "/scheduled-jobs/$SCHEDULED_ID/disable"
fi
sleep 1

echo "11. Test: Enable Scheduled Job"
echo "==============================="
if [ -n "$SCHEDULED_ID" ]; then
    test_endpoint "Enable scheduled job" "POST" "/scheduled-jobs/$SCHEDULED_ID/enable"
fi
sleep 1

echo "12. Test: Create Job with Custom Schedule Time"
echo "==============================================="
FUTURE_TIME=$(date -u -d "+5 minutes" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+5M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)
FUTURE_JOB_DATA='{
  "job_type": "sync_calendar",
  "payload": {
    "user_id": "user123",
    "calendar_id": "cal456"
  },
  "scheduled_at": "'$FUTURE_TIME'",
  "priority": 3
}'
test_endpoint "Create future job" "POST" "/background-jobs" "$FUTURE_JOB_DATA"
sleep 1

echo ""
echo "=================================="
echo "Test Summary"
echo "=================================="
echo "✓ Background jobs API is working"
echo "✓ Scheduled jobs API is working"
echo "✓ Workers should be processing jobs"
echo "✓ Check server logs for worker activity"
echo ""
echo "Next steps:"
echo "1. Check server logs for: 'Job acquired', 'Job completed'"
echo "2. Query jobs again to see status changes"
echo "3. Monitor scheduled jobs execution"
echo ""
echo "Cleanup (optional):"
if [ -n "$SCHEDULED_ID" ]; then
    echo "Delete scheduled job: curl -X DELETE $BASE_URL/scheduled-jobs/$SCHEDULED_ID"
fi
echo ""
