#!/bin/bash
# Test script for username API endpoints

BASE_URL="http://localhost:8080/api"

echo "==================================="
echo "Username System API Tests"
echo "==================================="
echo ""

# Test 1: Check availability - valid username
echo "Test 1: Check availability for 'johndoe'"
curl -s "${BASE_URL}/users/check-username/johndoe" | jq .
echo ""

# Test 2: Check availability - too short
echo "Test 2: Check availability for 'ab' (too short)"
curl -s "${BASE_URL}/users/check-username/ab" | jq .
echo ""

# Test 3: Check availability - invalid characters
echo "Test 3: Check availability for 'john-doe' (invalid chars)"
curl -s "${BASE_URL}/users/check-username/john-doe" | jq .
echo ""

# Test 4: Check availability - reserved name
echo "Test 4: Check availability for 'admin' (reserved)"
curl -s "${BASE_URL}/users/check-username/admin" | jq .
echo ""

# Test 5: Set username (requires authentication)
echo "Test 5: Set username to 'johndoe' (requires session token)"
echo "Usage: export SESSION_TOKEN=your_token_here"
if [ -n "$SESSION_TOKEN" ]; then
  curl -s -X PATCH "${BASE_URL}/users/me/username" \
    -H "Cookie: better-auth.session_token=${SESSION_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"username": "johndoe"}' | jq .
else
  echo "Skipped - no SESSION_TOKEN set"
fi
echo ""

echo "==================================="
echo "Tests complete!"
echo "==================================="
