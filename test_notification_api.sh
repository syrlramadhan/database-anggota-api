#!/bin/bash

# Notification System API Test Script
# This script tests the notification system endpoints

# Configuration
BASE_URL="http://localhost:8080"
TOKEN="your_jwt_token_here"

echo "=== Notification System API Test ==="
echo "Base URL: $BASE_URL"
echo "Using token: ${TOKEN:0:20}..."
echo ""

# Function to make HTTP requests with proper headers
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ "$method" = "GET" ]; then
        curl -s -X GET "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" | jq .
    elif [ "$method" = "PUT" ] && [ -z "$data" ]; then
        curl -s -X PUT "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" | jq .
    elif [ "$method" = "POST" ]; then
        curl -s -X POST "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" | jq .
    fi
}

# Test 1: Get all notifications
echo "1. Testing GET /api/notifications"
make_request "GET" "/api/notifications"
echo ""

# Test 2: Get unread notification count
echo "2. Testing GET /api/notifications/unread/count"
make_request "GET" "/api/notifications/unread/count"
echo ""

# Test 3: Create a status change request (replace with actual member IDs)
echo "3. Testing POST /api/status-change/request"
test_data='{
  "target_member_id": "replace-with-actual-member-id",
  "from_status": "bph",
  "to_status": "dpo"
}'
make_request "POST" "/api/status-change/request" "$test_data"
echo ""

# Test 4: Mark notification as read (replace with actual notification ID)
echo "4. Testing PUT /api/notifications/{id}/read"
notification_id="replace-with-actual-notification-id"
make_request "PUT" "/api/notifications/$notification_id/read"
echo ""

# Test 5: Accept status change request (replace with actual request ID)
echo "5. Testing PUT /api/status-change/{id}/accept"
request_id="replace-with-actual-request-id"
make_request "PUT" "/api/status-change/$request_id/accept"
echo ""

# Test 6: Reject status change request (replace with actual request ID)
echo "6. Testing PUT /api/status-change/{id}/reject"
request_id="replace-with-actual-request-id"
make_request "PUT" "/api/status-change/$request_id/reject"
echo ""

echo "=== Test Completed ==="
echo "Note: Replace placeholder IDs with actual values from your database"
echo "Make sure to:"
echo "1. Update the TOKEN variable with a valid JWT token"
echo "2. Replace placeholder member/notification/request IDs with real ones"
echo "3. Run the migration script first to create the database tables"
