# CasGists API Examples

## Authentication

### Register a New User
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "SecurePassword123!"
  }'
```

Response:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Login
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "SecurePassword123!"
  }'
```

### Refresh Token
```bash
curl -X POST http://localhost:3000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

### Setup 2FA
```bash
# Get QR code and secret
curl -X POST http://localhost:3000/api/v1/auth/2fa/setup \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Response:
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "backup_codes": [
    "12345678",
    "87654321",
    "..."
  ]
}
```

## Gist Management

### Create a Gist
```bash
curl -X POST http://localhost:3000/api/v1/gists \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Code Snippet",
    "description": "A useful Python function",
    "visibility": "public",
    "files": [
      {
        "filename": "utils.py",
        "content": "def hello_world():\n    print(\"Hello, World!\")",
        "language": "python"
      }
    ],
    "tags": ["python", "utility", "example"]
  }'
```

### Update a Gist
```bash
curl -X PUT http://localhost:3000/api/v1/gists/GIST_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Code Snippet",
    "files": [
      {
        "filename": "utils.py",
        "content": "def hello_world():\n    print(\"Hello, CasGists!\")",
        "language": "python"
      },
      {
        "filename": "README.md",
        "content": "# My Utils\n\nUseful utility functions.",
        "language": "markdown"
      }
    ]
  }'
```

### List Gists with Pagination
```bash
# Get public gists
curl "http://localhost:3000/api/v1/gists?page=1&limit=20&visibility=public"

# Get user's gists
curl "http://localhost:3000/api/v1/gists?user=johndoe" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Filter by language
curl "http://localhost:3000/api/v1/gists?language=python&sort=stars"
```

### Star a Gist
```bash
curl -X POST http://localhost:3000/api/v1/gists/GIST_ID/star \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Fork a Gist
```bash
curl -X POST http://localhost:3000/api/v1/gists/GIST_ID/fork \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Search

### Basic Search
```bash
curl "http://localhost:3000/api/v1/search?q=python+function"
```

### Advanced Search
```bash
# Search with filters
curl "http://localhost:3000/api/v1/search?q=async&user=johndoe&language=javascript&stars=>10"

# Search in specific fields
curl "http://localhost:3000/api/v1/search?q=title:API+description:REST"
```

### Search Suggestions
```bash
curl "http://localhost:3000/api/v1/search/suggestions?q=pyth"
```

Response:
```json
{
  "suggestions": [
    "python",
    "python function",
    "python async",
    "python api"
  ]
}
```

## Organizations

### Create Organization
```bash
curl -X POST http://localhost:3000/api/v1/orgs \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "acme-corp",
    "display_name": "ACME Corporation",
    "description": "Building great software",
    "website": "https://acme.example.com"
  }'
```

### Add Member to Organization
```bash
curl -X POST http://localhost:3000/api/v1/orgs/acme-corp/members \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "janedoe",
    "role": "member"
  }'
```

### Create Organization Gist
```bash
curl -X POST http://localhost:3000/api/v1/gists \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization": "acme-corp",
    "title": "Company Standards",
    "visibility": "private",
    "files": [
      {
        "filename": "coding-standards.md",
        "content": "# ACME Coding Standards\n..."
      }
    ]
  }'
```

## Webhooks

### Create Webhook
```bash
curl -X POST http://localhost:3000/api/v1/webhooks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/webhooks/casgists",
    "events": ["gist.created", "gist.updated", "gist.starred"],
    "active": true,
    "secret": "webhook_secret_key",
    "filters": {
      "rules": [
        {
          "field": "gist.language",
          "operator": "in",
          "value": ["python", "javascript"]
        }
      ],
      "logic": "AND"
    }
  }'
```

### Test Webhook
```bash
curl -X POST http://localhost:3000/api/v1/webhooks/WEBHOOK_ID/ping \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Get Webhook Deliveries
```bash
curl "http://localhost:3000/api/v1/webhooks/WEBHOOK_ID/deliveries?status=failed" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## GDPR Compliance

### Request Data Export
```bash
curl -X POST http://localhost:3000/api/v1/gdpr/export \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "format": "json",
    "include": ["profile", "gists", "comments", "stars", "audit_logs"]
  }'
```

Response:
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "estimated_completion": "2024-01-01T12:00:00Z"
}
```

### Check Export Status
```bash
curl "http://localhost:3000/api/v1/gdpr/export/REQUEST_ID" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Request Account Deletion
```bash
curl -X POST http://localhost:3000/api/v1/gdpr/delete \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "No longer using the service",
    "confirm_deletion": true
  }'
```

## Transfer System

### Create Transfer Request
```bash
curl -X POST http://localhost:3000/api/v1/transfers \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "gist_id": "550e8400-e29b-41d4-a716-446655440000",
    "to_user": "janedoe",
    "message": "Transferring ownership as discussed",
    "transfer_reason": "ownership_change",
    "preserve_history": true
  }'
```

### Accept Transfer
```bash
curl -X POST http://localhost:3000/api/v1/transfers/TRANSFER_ID/accept \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Reject Transfer
```bash
curl -X POST http://localhost:3000/api/v1/transfers/TRANSFER_ID/reject \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Not interested in maintaining this gist"
  }'
```

## Batch Operations

### Bulk Create Gists
```bash
curl -X POST http://localhost:3000/api/v1/gists/bulk \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "gists": [
      {
        "title": "Gist 1",
        "files": [{"filename": "file1.txt", "content": "Content 1"}]
      },
      {
        "title": "Gist 2",
        "files": [{"filename": "file2.txt", "content": "Content 2"}]
      }
    ]
  }'
```

### Bulk Operations
```bash
# Star multiple gists
curl -X POST http://localhost:3000/api/v1/gists/bulk/star \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "gist_ids": ["id1", "id2", "id3"]
  }'

# Delete multiple gists
curl -X DELETE http://localhost:3000/api/v1/gists/bulk \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "gist_ids": ["id1", "id2", "id3"]
  }'
```

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input provided",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

Common error codes:
- `UNAUTHORIZED` - Invalid or missing authentication
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Input validation failed
- `RATE_LIMITED` - Too many requests
- `INTERNAL_ERROR` - Server error

## Rate Limiting

Rate limit information is included in response headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 58
X-RateLimit-Reset: 1640995200
```

## Pagination

Paginated responses include metadata:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  },
  "links": {
    "first": "/api/v1/gists?page=1&limit=20",
    "prev": null,
    "next": "/api/v1/gists?page=2&limit=20",
    "last": "/api/v1/gists?page=8&limit=20"
  }
}
```

## WebSocket Events (Future)

```javascript
// Connect to WebSocket
const ws = new WebSocket('wss://gists.example.com/api/v1/ws');

// Authenticate
ws.send(JSON.stringify({
  type: 'auth',
  token: 'YOUR_ACCESS_TOKEN'
}));

// Subscribe to events
ws.send(JSON.stringify({
  type: 'subscribe',
  events: ['gist.updated', 'gist.starred'],
  filters: {
    user_id: 'YOUR_USER_ID'
  }
}));

// Handle events
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data.type, data.payload);
};
```