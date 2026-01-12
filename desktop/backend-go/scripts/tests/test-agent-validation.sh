#!/bin/bash

# Test script for Custom Agents input validation
# Tests all validation rules added to CreateCustomAgent and UpdateCustomAgent handlers

BASE_URL="http://localhost:8080"
AUTH_HEADER="Authorization: Bearer YOUR_TOKEN_HERE"

echo "=========================================="
echo "Custom Agents Validation Test Suite"
echo "=========================================="
echo ""

# Test 1: Too many suggested prompts (>10)
echo "TEST 1: Too many suggested prompts (should fail)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "name": "test-agent-1",
    "display_name": "Test Agent 1",
    "system_prompt": "You are a helpful assistant",
    "suggested_prompts": [
      "Prompt 1", "Prompt 2", "Prompt 3", "Prompt 4", "Prompt 5",
      "Prompt 6", "Prompt 7", "Prompt 8", "Prompt 9", "Prompt 10", "Prompt 11"
    ]
  }'
echo -e "\n"

# Test 2: Suggested prompt too long (>500 chars)
echo "TEST 2: Suggested prompt exceeds 500 characters (should fail)"
LONG_PROMPT=$(printf 'a%.0s' {1..501})
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"name\": \"test-agent-2\",
    \"display_name\": \"Test Agent 2\",
    \"system_prompt\": \"You are a helpful assistant\",
    \"suggested_prompts\": [\"$LONG_PROMPT\"]
  }"
echo -e "\n"

# Test 3: Empty suggested prompt
echo "TEST 3: Empty suggested prompt (should fail)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "name": "test-agent-3",
    "display_name": "Test Agent 3",
    "system_prompt": "You are a helpful assistant",
    "suggested_prompts": ["Valid prompt", "   ", "Another valid"]
  }'
echo -e "\n"

# Test 4: Welcome message too long (>2000 chars)
echo "TEST 4: Welcome message exceeds 2000 characters (should fail)"
LONG_WELCOME=$(printf 'w%.0s' {1..2001})
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{
    \"name\": \"test-agent-4\",
    \"display_name\": \"Test Agent 4\",
    \"system_prompt\": \"You are a helpful assistant\",
    \"welcome_message\": \"$LONG_WELCOME\"
  }"
echo -e "\n"

# Test 5: Invalid category
echo "TEST 5: Invalid category (should fail)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "name": "test-agent-5",
    "display_name": "Test Agent 5",
    "system_prompt": "You are a helpful assistant",
    "category": "invalid_category"
  }'
echo -e "\n"

# Test 6: Valid agent creation (should succeed)
echo "TEST 6: Valid agent creation (should succeed)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "name": "test-agent-valid",
    "display_name": "Valid Test Agent",
    "system_prompt": "You are a helpful assistant",
    "category": "general",
    "welcome_message": "Hello! How can I help?",
    "suggested_prompts": [
      "Help me with coding",
      "Explain a concept",
      "Review my code"
    ],
    "thinking_enabled": true,
    "streaming_enabled": true
  }'
echo -e "\n"

# Test 7: Update with invalid category (should fail)
echo "TEST 7: Update with invalid category (should fail)"
echo "NOTE: Replace AGENT_ID with actual agent ID from previous test"
curl -X PUT "$BASE_URL/api/agents/AGENT_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "category": "not_a_valid_category"
  }'
echo -e "\n"

# Test 8: Update with too many prompts (should fail)
echo "TEST 8: Update with too many suggested prompts (should fail)"
curl -X PUT "$BASE_URL/api/agents/AGENT_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "suggested_prompts": [
      "P1", "P2", "P3", "P4", "P5", "P6", "P7", "P8", "P9", "P10", "P11"
    ]
  }'
echo -e "\n"

# Test 9: Valid update (should succeed)
echo "TEST 9: Valid update (should succeed)"
curl -X PUT "$BASE_URL/api/agents/AGENT_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "category": "coding",
    "welcome_message": "Updated welcome!",
    "suggested_prompts": ["New prompt 1", "New prompt 2"]
  }'
echo -e "\n"

echo "=========================================="
echo "Test suite complete!"
echo "=========================================="
echo ""
echo "Expected results:"
echo "- Tests 1-5, 7-8: Should return 400 Bad Request with validation error"
echo "- Tests 6, 9: Should return 201/200 with agent data"
echo ""
echo "NOTE: Replace 'YOUR_TOKEN_HERE' with actual auth token"
echo "NOTE: Replace 'AGENT_ID' in tests 7-9 with actual agent ID"
