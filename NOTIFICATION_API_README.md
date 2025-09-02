# Notification System API Documentation

This document describes the notification system endpoints that enable status change requests and notifications for the member management system.

## Overview

The notification system allows BPH (Board) members to request status changes for other BPH members through a notification-based approval system. When a BPH member attempts to change another BPH member's status, instead of making the change immediately, a notification is sent to the target member for approval.

## Database Setup

Before using the notification endpoints, run the migration script to create the required tables:

```sql
-- Run the migration_notification_system.sql file
mysql -u your_username -p your_database < migration_notification_system.sql
```

This will create:
- `notifications` table: Stores all notifications
- `status_change_requests` table: Stores status change request details

## Authentication

All notification endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer {your_jwt_token}
```

## API Endpoints

### 1. Get All Notifications

**GET** `/api/notifications`

Retrieves all notifications for the authenticated user.

**Headers:**
- `Authorization: Bearer {token}`

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": [
    {
      "id_notification": "550e8400-e29b-41d4-a716-446655440000",
      "type": "status_change_request",
      "title": "Status Change Request",
      "message": "Ahmad Fadli requests to change your status from bph to dpo",
      "pending": true,
      "accepted": null,
      "read_at": null,
      "created_at": "2025-08-30T10:30:00Z",
      "from_member": {
        "id_member": "550e8400-e29b-41d4-a716-446655440001",
        "nama": "Ahmad Fadli",
        "nra": "12.23.002",
        "status_keanggotaan": "bph"
      },
      "metadata": {
        "request_id": "550e8400-e29b-41d4-a716-446655440002",
        "from_status": "bph",
        "to_status": "dpo"
      }
    }
  ],
  "message": "Notifications retrieved successfully"
}
```

### 2. Mark Notification as Read

**PUT** `/api/notifications/{id}/read`

Marks a specific notification as read.

**Headers:**
- `Authorization: Bearer {token}`

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "message": "Notification marked as read"
  },
  "message": "Notification marked as read"
}
```

### 3. Get Unread Notification Count

**GET** `/api/notifications/unread/count`

Gets the count of unread notifications for the authenticated user.

**Headers:**
- `Authorization: Bearer {token}`

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "unread_count": 3
  },
  "message": "Unread notification count retrieved successfully"
}
```

### 4. Create Status Change Request

**POST** `/api/status-change/request`

Creates a new status change request (notification) for another member.

**Headers:**
- `Authorization: Bearer {token}`
- `Content-Type: application/json`

**Body:**
```json
{
  "target_member_id": "550e8400-e29b-41d4-a716-446655440003",
  "from_status": "bph",
  "to_status": "dpo"
}
```

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "request_id": "550e8400-e29b-41d4-a716-446655440004",
    "notification_id": "550e8400-e29b-41d4-a716-446655440005",
    "message": "Status change request sent successfully"
  },
  "message": "Status change request created successfully"
}
```

### 5. Accept Status Change Request

**PUT** `/api/status-change/{id}/accept`

Accepts a status change request (only the target member can accept).

**Headers:**
- `Authorization: Bearer {token}`

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "message": "Status change accepted",
    "new_status": "dpo"
  },
  "message": "Status change request accepted successfully"
}
```

### 6. Reject Status Change Request

**PUT** `/api/status-change/{id}/reject`

Rejects a status change request (only the target member can reject).

**Headers:**
- `Authorization: Bearer {token}`

**Response:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "message": "Status change rejected"
  },
  "message": "Status change request rejected successfully"
}
```

### 7. Enhanced Member Update (with Notification)

**PUT** `/api/member/{id}/notify`

Updates a member with automatic notification system for BPH status changes.

**Headers:**
- `Authorization: Bearer {token}`
- `Content-Type: multipart/form-data`

**Form Data:**
- `status_keanggotaan`: New status (required if changing status)
- Other member fields as needed

**Response (when notification is sent):**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "member": {
      "id_member": "550e8400-e29b-41d4-a716-446655440000",
      "nra": "12.23.001",
      "nama": "John Doe",
      "status_keanggotaan": "bph",
      "angkatan": "015"
    },
    "notification_sent": true,
    "notification_id": "550e8400-e29b-41d4-a716-446655440006"
  },
  "message": "status change request sent for approval"
}
```

**Response (when direct update):**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "member": {
      "id_member": "550e8400-e29b-41d4-a716-446655440000",
      "nra": "12.23.001",
      "nama": "John Doe Updated",
      "status_keanggotaan": "dpo"
    },
    "notification_sent": false,
    "notification_id": ""
  },
  "message": "updated successfully"
}
```

## Business Logic

### Status Change Rules

1. **BPH → BPH Changes**: When a BPH member tries to change another BPH member's status, a notification is sent for approval instead of making the change directly.

2. **Other Changes**: All other status changes are applied immediately without notification.

3. **Self Updates**: Members can update their own information directly without notifications.

### Notification Flow

1. **Request Creation**: BPH member A requests to change BPH member B's status
2. **Notification Sent**: System creates a notification for member B
3. **Approval/Rejection**: Member B can accept or reject the request
4. **Status Update**: If accepted, member B's status is updated; if rejected, no change occurs

## Error Handling

### Common Error Responses

**401 Unauthorized:**
```json
{
  "code": 401,
  "status": "Unauthorized",
  "message": "authorization header is required"
}
```

**403 Forbidden:**
```json
{
  "code": 403,
  "status": "Forbidden", 
  "message": "only BPH members can request status changes"
}
```

**400 Bad Request:**
```json
{
  "code": 400,
  "status": "Bad Request",
  "message": "target_member_id, from_status, and to_status are required"
}
```

**404 Not Found:**
```json
{
  "code": 404,
  "status": "Not Found",
  "message": "notification not found or not owned by member"
}
```

## Usage Examples

### Frontend Integration

```javascript
// Get unread notification count
const getUnreadCount = async () => {
  const response = await fetch('/api/notifications/unread/count', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  const data = await response.json();
  return data.data.unread_count;
};

// Accept a status change request
const acceptStatusChange = async (requestId) => {
  const response = await fetch(`/api/status-change/${requestId}/accept`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return await response.json();
};

// Send status change request
const sendStatusChangeRequest = async (targetMemberId, fromStatus, toStatus) => {
  const response = await fetch('/api/status-change/request', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      target_member_id: targetMemberId,
      from_status: fromStatus,
      to_status: toStatus
    })
  });
  return await response.json();
};
```

### Testing with curl

```bash
# Get all notifications
curl -X GET "http://localhost:8080/api/notifications" \
  -H "Authorization: Bearer your_jwt_token"

# Create status change request
curl -X POST "http://localhost:8080/api/status-change/request" \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "target_member_id": "target-member-uuid",
    "from_status": "bph", 
    "to_status": "dpo"
  }'

# Accept status change request
curl -X PUT "http://localhost:8080/api/status-change/request-uuid/accept" \
  -H "Authorization: Bearer your_jwt_token"
```

## Security Considerations

1. **Authorization**: All endpoints validate JWT tokens
2. **Permission Checks**: Only BPH members can create status change requests
3. **Ownership Validation**: Users can only accept/reject their own notifications
4. **Input Validation**: All inputs are validated before processing
5. **SQL Injection Prevention**: Uses prepared statements

## Performance Notes

The system includes database indexes for optimal query performance:
- `idx_notifications_target_pending`: For fetching user's pending notifications
- `idx_notifications_type_created`: For chronological notification queries  
- `idx_status_requests_target_status`: For status change request lookups

## Future Enhancements

1. **Real-time Notifications**: WebSocket support for instant notification delivery
2. **Email Notifications**: Send email alerts for important status changes
3. **Notification Templates**: Customizable notification message templates
4. **Batch Operations**: Accept/reject multiple notifications at once
5. **Notification History**: Archive and search historical notifications
