# CasGists API Reference

## Overview

The CasGists API is a RESTful API that provides programmatic access to all CasGists features. All API endpoints return JSON responses and use standard HTTP response codes.

## Base URL

```
https://gists.example.com/api/v1
```

## Authentication

CasGists uses JWT (JSON Web Tokens) for API authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Obtaining a Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "your-username",
  "password": "your-password"
}
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "your-username",
    "email": "user@example.com"
  }
}
```

## Rate Limiting

API requests are rate-limited to prevent abuse:

- **Authenticated requests**: 5000 per hour
- **Unauthenticated requests**: 60 per hour
- **Search requests**: 30 per minute

Rate limit headers:
```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1640995200
```

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Gist not found",
    "details": {
      "gist_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

Common error codes:
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `429` - Too Many Requests
- `500` - Internal Server Error

## Pagination

List endpoints support pagination using query parameters:

```
GET /api/v1/gists?page=2&per_page=20
```

Pagination response:
```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Authentication Endpoints

### Register

Create a new user account.

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "SecurePassword123!",
  "password_confirm": "SecurePassword123!"
}
```

Response: `201 Created`
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "newuser",
    "email": "newuser@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Login

Authenticate and receive access tokens.

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "password",
  "totp_code": "123456"  // Optional, if 2FA enabled
}
```

Response: `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

### Refresh Token

Get a new access token using a refresh token.

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Logout

Invalidate the current session.

```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

Response: `204 No Content`

## Gist Endpoints

### List Gists

Get a list of gists.

```http
GET /api/v1/gists?visibility=public&sort=created&page=1&per_page=20
```

Query parameters:
- `visibility` - Filter by visibility: `public`, `private`, `unlisted`
- `username` - Filter by username
- `sort` - Sort by: `created`, `updated`, `stars`
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 20, max: 100)

Response: `200 OK`
```json
{
  "gists": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Example Gist",
      "description": "This is an example",
      "visibility": "public",
      "view_count": 42,
      "star_count": 5,
      "fork_count": 2,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "user": {
        "id": "user-id",
        "username": "john",
        "avatar_url": "https://example.com/avatar.jpg"
      },
      "files": [
        {
          "id": "file-id",
          "filename": "example.py",
          "language": "python",
          "size": 256,
          "line_count": 10
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Create Gist

Create a new gist.

```http
POST /api/v1/gists
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "My New Gist",
  "description": "Example gist created via API",
  "visibility": "public",
  "files": [
    {
      "filename": "hello.py",
      "content": "print('Hello, World!')",
      "language": "python"
    }
  ]
}
```

Response: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "My New Gist",
  "description": "Example gist created via API",
  "visibility": "public",
  "created_at": "2024-01-15T10:30:00Z",
  "files": [
    {
      "id": "file-id",
      "filename": "hello.py",
      "content": "print('Hello, World!')",
      "language": "python",
      "size": 21,
      "line_count": 1
    }
  ]
}
```

### Get Gist

Get a specific gist by ID.

```http
GET /api/v1/gists/{gist_id}
```

Response: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Example Gist",
  "description": "Detailed gist information",
  "visibility": "public",
  "view_count": 42,
  "star_count": 5,
  "fork_count": 2,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "user": {
    "id": "user-id",
    "username": "john",
    "email": "john@example.com",
    "avatar_url": "https://example.com/avatar.jpg"
  },
  "files": [
    {
      "id": "file-id",
      "filename": "example.py",
      "content": "# Full file content here",
      "language": "python",
      "size": 1024,
      "line_count": 42
    }
  ]
}
```

### Update Gist

Update an existing gist.

```http
PUT /api/v1/gists/{gist_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "visibility": "private",
  "files": [
    {
      "filename": "updated.py",
      "content": "# Updated content",
      "language": "python"
    }
  ]
}
```

### Delete Gist

Delete a gist (soft delete).

```http
DELETE /api/v1/gists/{gist_id}
Authorization: Bearer <token>
```

Response: `204 No Content`

### Star Gist

Star a gist.

```http
POST /api/v1/gists/{gist_id}/star
Authorization: Bearer <token>
```

Response: `201 Created`

### Unstar Gist

Remove star from a gist.

```http
DELETE /api/v1/gists/{gist_id}/star
Authorization: Bearer <token>
```

Response: `204 No Content`

### Fork Gist

Create a fork of a gist.

```http
POST /api/v1/gists/{gist_id}/fork
Authorization: Bearer <token>
```

Response: `201 Created`
```json
{
  "id": "new-gist-id",
  "title": "Forked: Original Title",
  "forked_from_id": "original-gist-id",
  // ... rest of gist data
}
```

### Get Gist Stars

Get users who starred a gist.

```http
GET /api/v1/gists/{gist_id}/stars?page=1&per_page=20
```

### Get Gist Forks

Get all forks of a gist.

```http
GET /api/v1/gists/{gist_id}/forks?page=1&per_page=20
```

## File Operations

### Get Raw File

Get raw content of a gist file.

```http
GET /api/v1/gists/{gist_id}/files/{filename}/raw
```

Response: Raw file content with appropriate Content-Type header.

### Download Gist

Download gist as archive.

```http
GET /api/v1/gists/{gist_id}/download?format=zip
```

Query parameters:
- `format` - Archive format: `zip`, `tar`, `tar.gz` (default: `zip`)

## User Endpoints

### Get Current User

Get the authenticated user's profile.

```http
GET /api/v1/user
Authorization: Bearer <token>
```

Response: `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john",
  "email": "john@example.com",
  "display_name": "John Doe",
  "bio": "Software Developer",
  "website": "https://johndoe.com",
  "location": "San Francisco, CA",
  "avatar_url": "https://example.com/avatar.jpg",
  "created_at": "2024-01-01T00:00:00Z",
  "gist_count": 42,
  "star_count": 100,
  "follower_count": 50,
  "following_count": 30
}
```

### Update Current User

Update the authenticated user's profile.

```http
PUT /api/v1/user
Authorization: Bearer <token>
Content-Type: application/json

{
  "display_name": "John Doe",
  "bio": "Updated bio",
  "website": "https://newsite.com",
  "location": "New York, NY"
}
```

### Get User

Get a user's public profile.

```http
GET /api/v1/users/{username}
```

### Get User Gists

Get a user's gists.

```http
GET /api/v1/users/{username}/gists?page=1&per_page=20
```

### Follow User

Follow a user.

```http
POST /api/v1/users/{username}/follow
Authorization: Bearer <token>
```

### Unfollow User

Unfollow a user.

```http
DELETE /api/v1/users/{username}/follow
Authorization: Bearer <token>
```

## Search

### Search Gists

Search for gists.

```http
GET /api/v1/search?q=python+flask&type=gists&page=1&per_page=20
```

Query parameters:
- `q` - Search query (required)
- `type` - Search type: `gists`, `users`, `files`
- `language` - Filter by programming language
- `user` - Filter by username
- `sort` - Sort results: `relevance`, `stars`, `created`, `updated`
- `order` - Order: `asc`, `desc`

Response: `200 OK`
```json
{
  "results": [
    {
      "type": "gist",
      "gist": {
        "id": "gist-id",
        "title": "Flask Example",
        "description": "Example Flask application",
        "matches": [
          {
            "filename": "app.py",
            "line_number": 5,
            "content": "from flask import Flask"
          }
        ]
      }
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 42
  }
}
```

## Comments

### Get Comments

Get comments on a gist.

```http
GET /api/v1/gists/{gist_id}/comments?page=1&per_page=20
```

### Create Comment

Add a comment to a gist.

```http
POST /api/v1/gists/{gist_id}/comments
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Great gist! Thanks for sharing."
}
```

Response: `201 Created`
```json
{
  "id": "comment-id",
  "content": "Great gist! Thanks for sharing.",
  "user": {
    "id": "user-id",
    "username": "commenter",
    "avatar_url": "https://example.com/avatar.jpg"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Update Comment

Update a comment.

```http
PUT /api/v1/gists/{gist_id}/comments/{comment_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Updated comment content"
}
```

### Delete Comment

Delete a comment.

```http
DELETE /api/v1/gists/{gist_id}/comments/{comment_id}
Authorization: Bearer <token>
```

## Organizations

### List Organizations

Get user's organizations.

```http
GET /api/v1/orgs
Authorization: Bearer <token>
```

### Create Organization

Create a new organization.

```http
POST /api/v1/orgs
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "my-org",
  "display_name": "My Organization",
  "description": "Organization description",
  "website": "https://myorg.com",
  "email": "contact@myorg.com"
}
```

### Get Organization

Get organization details.

```http
GET /api/v1/orgs/{org_name}
```

### Update Organization

Update organization details.

```http
PUT /api/v1/orgs/{org_name}
Authorization: Bearer <token>
Content-Type: application/json

{
  "display_name": "Updated Name",
  "description": "Updated description"
}
```

### Organization Members

#### List Members

```http
GET /api/v1/orgs/{org_name}/members
```

#### Add Member

```http
POST /api/v1/orgs/{org_name}/members
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newmember",
  "role": "member"  // member, admin, owner
}
```

#### Update Member Role

```http
PUT /api/v1/orgs/{org_name}/members/{username}
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "admin"
}
```

#### Remove Member

```http
DELETE /api/v1/orgs/{org_name}/members/{username}
Authorization: Bearer <token>
```

## Teams

### List Teams

Get organization teams.

```http
GET /api/v1/orgs/{org_name}/teams
```

### Create Team

Create a new team.

```http
POST /api/v1/orgs/{org_name}/teams
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "developers",
  "description": "Development team",
  "privacy": "closed"  // closed, secret
}
```

### Team Members

#### Add Team Member

```http
POST /api/v1/teams/{team_id}/members
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "developer1"
}
```

#### Remove Team Member

```http
DELETE /api/v1/teams/{team_id}/members/{username}
Authorization: Bearer <token>
```

## Webhooks

### List Webhooks

Get webhooks for authenticated user.

```http
GET /api/v1/webhooks
Authorization: Bearer <token>
```

### Create Webhook

Create a new webhook.

```http
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://example.com/webhook",
  "events": ["gist.created", "gist.updated", "gist.deleted"],
  "active": true,
  "secret": "webhook-secret"
}
```

Response: `201 Created`
```json
{
  "id": "webhook-id",
  "url": "https://example.com/webhook",
  "events": ["gist.created", "gist.updated", "gist.deleted"],
  "active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Update Webhook

Update webhook configuration.

```http
PUT /api/v1/webhooks/{webhook_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "events": ["gist.created"],
  "active": false
}
```

### Delete Webhook

Delete a webhook.

```http
DELETE /api/v1/webhooks/{webhook_id}
Authorization: Bearer <token>
```

### Test Webhook

Send a test payload to webhook.

```http
POST /api/v1/webhooks/{webhook_id}/test
Authorization: Bearer <token>
```

### Webhook Events

Available webhook events:
- `gist.created` - A gist was created
- `gist.updated` - A gist was updated
- `gist.deleted` - A gist was deleted
- `gist.starred` - A gist was starred
- `gist.unstarred` - A gist was unstarred
- `gist.forked` - A gist was forked
- `comment.created` - A comment was created
- `comment.updated` - A comment was updated
- `comment.deleted` - A comment was deleted
- `user.followed` - A user was followed
- `user.unfollowed` - A user was unfollowed

Webhook payload example:
```json
{
  "event": "gist.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "payload": {
    "gist": {
      "id": "gist-id",
      "title": "New Gist",
      // ... gist data
    },
    "user": {
      "id": "user-id",
      "username": "creator"
    }
  }
}
```

## API Tokens

### List Tokens

Get all API tokens for authenticated user.

```http
GET /api/v1/tokens
Authorization: Bearer <token>
```

### Create Token

Create a new API token.

```http
POST /api/v1/tokens
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My API Client",
  "scopes": ["gist:read", "gist:write"],
  "expires_at": "2025-01-01T00:00:00Z"
}
```

Response: `201 Created`
```json
{
  "id": "token-id",
  "name": "My API Client",
  "token": "cst_1234567890abcdef",  // Only shown once!
  "scopes": ["gist:read", "gist:write"],
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2025-01-01T00:00:00Z",
  "last_used_at": null
}
```

### Revoke Token

Revoke an API token.

```http
DELETE /api/v1/tokens/{token_id}
Authorization: Bearer <token>
```

### Token Scopes

Available scopes:
- `gist:read` - Read access to gists
- `gist:write` - Write access to gists
- `user:read` - Read access to user profile
- `user:write` - Write access to user profile
- `user:email` - Access to email addresses
- `org:read` - Read access to organizations
- `org:write` - Write access to organizations
- `webhook:read` - Read access to webhooks
- `webhook:write` - Write access to webhooks

## Migration

### Import from GitHub

Import gists from GitHub.

```http
POST /api/v1/import/github
Authorization: Bearer <token>
Content-Type: application/json

{
  "github_token": "ghp_xxxxxxxxxxxx",
  "import_starred": true,
  "visibility_mapping": {
    "public": "public",
    "secret": "unlisted"
  }
}
```

### Import from GitLab

Import snippets from GitLab.

```http
POST /api/v1/import/gitlab
Authorization: Bearer <token>
Content-Type: application/json

{
  "gitlab_url": "https://gitlab.com",
  "gitlab_token": "glpat-xxxxxxxxxxxx",
  "project_id": "12345"
}
```

### Import Status

Check import job status.

```http
GET /api/v1/import/status/{job_id}
Authorization: Bearer <token>
```

Response: `200 OK`
```json
{
  "job_id": "import-job-id",
  "status": "processing",  // pending, processing, completed, failed
  "progress": 45,
  "total": 100,
  "errors": [],
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": null
}
```

## Admin Endpoints

### System Status

Get system status (admin only).

```http
GET /api/v1/admin/system
Authorization: Bearer <admin-token>
```

Response: `200 OK`
```json
{
  "version": "1.0.0",
  "uptime": "72h 15m 30s",
  "database": {
    "type": "postgresql",
    "version": "14.5",
    "size": "256MB",
    "connections": 15
  },
  "storage": {
    "used": "1.2GB",
    "available": "48.8GB",
    "gist_count": 1234,
    "file_count": 5678
  },
  "users": {
    "total": 100,
    "active": 85,
    "new_today": 5
  }
}
```

### User Management

#### List All Users

```http
GET /api/v1/admin/users?page=1&per_page=50
Authorization: Bearer <admin-token>
```

#### Suspend User

```http
POST /api/v1/admin/users/{user_id}/suspend
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "reason": "Terms of Service violation",
  "duration": "7d"  // Optional, permanent if not specified
}
```

#### Unsuspend User

```http
POST /api/v1/admin/users/{user_id}/unsuspend
Authorization: Bearer <admin-token>
```

### Audit Logs

Get audit logs.

```http
GET /api/v1/admin/audit?from=2024-01-01&to=2024-01-31&page=1
Authorization: Bearer <admin-token>
```

### Backup

Trigger system backup.

```http
POST /api/v1/admin/backup
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "include_git": true,
  "include_uploads": true,
  "compress": true
}
```

## GraphQL API

CasGists also provides a GraphQL API endpoint:

```
POST /api/graphql
```

Example query:
```graphql
query {
  viewer {
    username
    email
    gists(first: 10, visibility: PUBLIC) {
      edges {
        node {
          id
          title
          description
          starCount
          files {
            filename
            language
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
```

## SDK Examples

### JavaScript/TypeScript

```typescript
import { CasGistsClient } from '@casgists/sdk';

const client = new CasGistsClient({
  baseUrl: 'https://gists.example.com',
  token: 'your-api-token'
});

// Create a gist
const gist = await client.gists.create({
  title: 'My Gist',
  description: 'Created via SDK',
  visibility: 'public',
  files: [
    {
      filename: 'example.js',
      content: 'console.log("Hello, World!");'
    }
  ]
});

// Search gists
const results = await client.search.gists({
  query: 'javascript',
  sort: 'stars',
  perPage: 10
});
```

### Python

```python
from casgists import Client

client = Client(
    base_url='https://gists.example.com',
    token='your-api-token'
)

# Create a gist
gist = client.gists.create(
    title='My Gist',
    description='Created via SDK',
    visibility='public',
    files=[
        {
            'filename': 'example.py',
            'content': 'print("Hello, World!")'
        }
    ]
)

# List your gists
my_gists = client.gists.list(username='me', per_page=20)
```

### Go

```go
package main

import (
    "github.com/casapps/casgists-go"
)

func main() {
    client := casgists.NewClient("https://gists.example.com", "your-api-token")
    
    // Create a gist
    gist, err := client.Gists.Create(&casgists.CreateGistRequest{
        Title:       "My Gist",
        Description: "Created via SDK",
        Visibility:  "public",
        Files: []casgists.File{
            {
                Filename: "example.go",
                Content:  "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}",
            },
        },
    })
}
```

### cURL Examples

```bash
# Create a gist
curl -X POST https://gists.example.com/api/v1/gists \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Gist",
    "visibility": "public",
    "files": [{
      "filename": "test.sh",
      "content": "echo \"Hello, World!\""
    }]
  }'

# Search gists
curl "https://gists.example.com/api/v1/search?q=bash&type=gists" \
  -H "Authorization: Bearer your-token"

# Star a gist
curl -X POST https://gists.example.com/api/v1/gists/gist-id/star \
  -H "Authorization: Bearer your-token"
```

## Best Practices

1. **Rate Limiting**: Implement exponential backoff when hitting rate limits
2. **Pagination**: Always use pagination for list endpoints
3. **Caching**: Cache responses when appropriate
4. **Error Handling**: Check status codes and handle errors gracefully
5. **Security**: Never expose API tokens in client-side code
6. **Webhooks**: Verify webhook signatures for security

## API Versioning

The API uses URL versioning. The current version is `v1`. When breaking changes are introduced, a new version will be released while maintaining the previous version for backward compatibility.

## Support

- API Status: https://status.casgists.com
- API Documentation: https://docs.casgists.com/api
- Support: api-support@casgists.com