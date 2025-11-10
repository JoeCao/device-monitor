#!/bin/bash

API_URL="http://localhost:3000/api"

echo "=== Device Monitor API Test ==="
echo

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -s "$API_URL/health" | jq .
echo

# Test 2: Start device
echo "2. Starting device..."
SESSION_RESPONSE=$(curl -s -X POST "$API_URL/webhooks/device/start?deviceName=test-device-001" \
  -H "Content-Type: application/json" \
  -d '{"power": "on"}')
echo "$SESSION_RESPONSE" | jq .
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r .sessionId)
echo "Session ID: $SESSION_ID"
echo

# Test 3: Get sessions list
echo "3. Getting sessions list..."
curl -s "$API_URL/sessions?limit=10" | jq .
echo

# Test 4: Get specific session
echo "4. Getting session details..."
curl -s "$API_URL/sessions/$SESSION_ID" | jq .
echo

# Test 5: End device
echo "5. Stopping device..."
sleep 2  # Wait a bit to simulate running time
curl -s -X POST "$API_URL/webhooks/device/end?deviceName=test-device-001" \
  -H "Content-Type: application/json" \
  -d '{"power": "off"}' | jq .
echo

# Test 6: Get session report
echo "6. Getting session report..."
curl -s "$API_URL/sessions/$SESSION_ID/report" | jq .
echo

# Test 7: Get statistics
echo "7. Getting statistics..."
curl -s "$API_URL/sessions/statistics" | jq .
echo

# Test 8: Get IoT data points
echo "8. Getting IoT data points configuration..."
curl -s "$API_URL/iot/data-points" | jq .
echo

# Test 9: Test IoT connection
echo "9. Testing IoT connection..."
curl -s "$API_URL/iot/test-connection" | jq .
echo

echo "=== Test Complete ==="