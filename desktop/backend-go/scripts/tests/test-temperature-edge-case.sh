#!/bin/bash
# Test temperature edge case: 0.0 should be valid

BASE_URL="http://localhost:8080"
TOKEN="your-auth-token-here"  # Replace with valid token

echo "Testing Temperature Edge Cases"
echo "================================"

# Test 1: temperature = 0.0 (should succeed)
echo -e "\n1. Testing temperature = 0.0 (deterministic output)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "test-deterministic-agent",
    "display_name": "Deterministic Test Agent",
    "system_prompt": "You are a test agent with zero randomness.",
    "temperature": 0.0,
    "streaming_enabled": true
  }' | jq

# Test 2: temperature = 0.5 (should succeed)
echo -e "\n2. Testing temperature = 0.5 (balanced)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "test-balanced-agent",
    "display_name": "Balanced Test Agent",
    "system_prompt": "You are a test agent with balanced randomness.",
    "temperature": 0.5,
    "streaming_enabled": true
  }' | jq

# Test 3: temperature = 2.0 (should succeed)
echo -e "\n3. Testing temperature = 2.0 (maximum)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "test-creative-agent",
    "display_name": "Creative Test Agent",
    "system_prompt": "You are a test agent with maximum randomness.",
    "temperature": 2.0,
    "streaming_enabled": true
  }' | jq

# Test 4: temperature = -0.1 (should fail with validation error)
echo -e "\n4. Testing temperature = -0.1 (should fail)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "test-invalid-negative",
    "display_name": "Invalid Negative",
    "system_prompt": "Should not be created.",
    "temperature": -0.1,
    "streaming_enabled": true
  }' | jq

# Test 5: temperature = 2.5 (should fail with validation error)
echo -e "\n5. Testing temperature = 2.5 (should fail)"
curl -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "test-invalid-high",
    "display_name": "Invalid High",
    "system_prompt": "Should not be created.",
    "temperature": 2.5,
    "streaming_enabled": true
  }' | jq

# Test 6: Update agent with temperature = 0.0 (should succeed)
echo -e "\n6. Testing UPDATE with temperature = 0.0"
AGENT_ID="replace-with-actual-id"  # Replace with ID from test 1
curl -X PUT "$BASE_URL/api/agents/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "temperature": 0.0
  }' | jq

echo -e "\n================================"
echo "Tests completed!"
echo ""
echo "Expected results:"
echo "  Tests 1-3: Success (201 Created)"
echo "  Tests 4-5: Fail (400 Bad Request with error message)"
echo "  Test 6: Success (200 OK)"
