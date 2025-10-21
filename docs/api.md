# API Documentation

CasGists provides a comprehensive REST API for programmatic access to all functionality. This document covers authentication, endpoints, and usage examples.

## Base URL

```
https://your-casgists-instance.com/api/v1
```

## Authentication

### API Tokens

API tokens provide scoped access to the CasGists API. Tokens can be created in the user settings.

```bash
# Include token in Authorization header
curl -H "Authorization: Bearer your_api_token" \
     https://gists.example.com/api/v1/user
```

### Token Scopes

Available scopes for API tokens:

- `read` - Read access to user's gists and profile
- `write` - Create and modify gists
- `delete` - Delete gists
- `admin` - Administrative operations (admin users only)
- `webhook` - Manage webhooks
- `organization` - Organization management

### JWT Authentication

For web applications, you can use JWT tokens obtained through login:

```bash
# Login to get JWT token
curl -X POST https://gists.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Use JWT token
curl -H "Authorization: Bearer jwt_token_here" \
     https://gists.example.com/api/v1/user
```

## Rate Limiting

API requests are rate-limited per user:

- **Authenticated users**: 5000 requests/hour
- **Anonymous users**: 100 requests/hour

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4999
X-RateLimit-Reset: 1640995200
```

## Error Handling

API errors follow RFC 7807 (Problem Details for HTTP APIs):

```json
{
  "type": "https://casgists.com/docs/api/errors#not-found",
  "title": "Gist not found",
  "status": 404,
  "detail": "The gist with ID 'abc123' was not found",
  "instance": "/api/v1/gists/abc123"
}
```

Common HTTP status codes:

- `200` - Success
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Unprocessable Entity
- `429` - Rate Limited
- `500` - Internal Server Error

## Endpoints

### Authentication

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "password",
  "totp_code": "123456"
}
```

Response:
```json
{
  "access_token": "jwt_token_here",
  "refresh_token": "refresh_token_here",
  "expires_in": 86400,
  "user": {
    "id": "user_id",
    "username": "username",
    "email": "user@example.com"
  }
}
```

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer jwt_token_here
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh_token_here"
}
```

### User Management

#### Get Current User
```http
GET /api/v1/user
Authorization: Bearer token
```

Response:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "johndoe",
  "email": "john@example.com",
  "display_name": "John Doe",
  "bio": "Software developer",
  "avatar_url": "https://example.com/avatar.jpg",
  "is_admin": false,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-06-01T12:00:00Z"
}
```

#### Update User Profile
```http
PATCH /api/v1/user
Authorization: Bearer token
Content-Type: application/json

{
  "display_name": "John Smith",
  "bio": "Full-stack developer",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

#### Get User by Username
```http
GET /api/v1/users/{username}
```

#### List Users (Admin only)
```http
GET /api/v1/users?page=1&limit=50&search=john
Authorization: Bearer admin_token
```

### Gist Management

#### List Gists
```http
GET /api/v1/gists?page=1&limit=50&visibility=public&language=go
Authorization: Bearer token
```

Query parameters:
- `page` - Page number (default: 1)
- `limit` - Items per page (max: 100, default: 30)
- `visibility` - Filter by visibility: `public`, `private`, `unlisted`
- `language` - Filter by programming language
- `search` - Search query
- `user` - Filter by username
- `starred` - Show only starred gists (`true`/`false`)

Response:
```json
{
  "gists": [
    {
      "id": "gist_id_here",
      "title": "Example Gist",
      "description": "A sample code snippet",
      "visibility": "public",
      "language": "go",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-02T00:00:00Z",
      "owner": {
        "id": "user_id",
        "username": "johndoe",
        "avatar_url": "https://example.com/avatar.jpg"
      },
      "files": [
        {
          "filename": "main.go",
          "language": "go",
          "size": 1024
        }
      ],
      "stats": {
        "stars": 5,
        "forks": 2,
        "comments": 1
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 30,
    "total": 100,
    "pages": 4
  }
}
```

#### Get Gist
```http
GET /api/v1/gists/{gist_id}
Authorization: Bearer token
```

Response:
```json
{
  "id": "gist_id_here",
  "title": "Example Gist",
  "description": "A sample code snippet",
  "visibility": "public",
  "language": "go",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-02T00:00:00Z",
  "owner": {
    "id": "user_id",
    "username": "johndoe",
    "display_name": "John Doe",
    "avatar_url": "https://example.com/avatar.jpg"
  },
  "files": [
    {
      "filename": "main.go",
      "language": "go",
      "content": "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}",
      "size": 1024,
      "truncated": false
    }
  ],
  "git": {
    "commit_sha": "abc123def456",
    "clone_url": "https://gists.example.com/johndoe/gist_id_here.git"
  },
  "stats": {
    "stars": 5,
    "forks": 2,
    "comments": 1,
    "views": 150
  }
}
```

#### Create Gist
```http
POST /api/v1/gists
Authorization: Bearer token
Content-Type: application/json

{
  "title": "Hello World Example",
  "description": "A simple Hello World program in Go",
  "visibility": "public",
  "files": [
    {
      "filename": "main.go",
      "content": "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}"
    },
    {
      "filename": "README.md",
      "content": "# Hello World\n\nA simple example program."
    }
  ]
}
```

Response: `201 Created` with gist object

#### Update Gist
```http
PATCH /api/v1/gists/{gist_id}
Authorization: Bearer token
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "visibility": "private",
  "files": [
    {
      "filename": "main.go",
      "content": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, Updated World!\")\n}"
    }
  ]
}
```

#### Delete Gist
```http
DELETE /api/v1/gists/{gist_id}
Authorization: Bearer token
```

Response: `204 No Content`

#### Fork Gist
```http
POST /api/v1/gists/{gist_id}/fork
Authorization: Bearer token
```

Response: `201 Created` with new gist object

#### Star/Unstar Gist
```http
PUT /api/v1/gists/{gist_id}/star
Authorization: Bearer token
```

```http
DELETE /api/v1/gists/{gist_id}/star
Authorization: Bearer token
```

#### Get Gist Revisions
```http
GET /api/v1/gists/{gist_id}/revisions
Authorization: Bearer token
```

Response:
```json
{
  "revisions": [
    {
      "sha": "abc123def456",
      "message": "Update main.go",
      "author": {
        "username": "johndoe",
        "email": "john@example.com"
      },
      "committed_at": "2023-01-02T00:00:00Z",
      "stats": {
        "additions": 2,
        "deletions": 1,
        "files_changed": 1
      }
    }
  ]
}
```

#### Get Specific Revision
```http
GET /api/v1/gists/{gist_id}/revisions/{sha}
Authorization: Bearer token
```

### Comments

#### List Comments
```http
GET /api/v1/gists/{gist_id}/comments
Authorization: Bearer token
```

Response:
```json
{
  "comments": [
    {
      "id": "comment_id",
      "body": "Great example!",
      "author": {
        "username": "janedoe",
        "avatar_url": "https://example.com/jane-avatar.jpg"
      },
      "created_at": "2023-01-02T00:00:00Z",
      "updated_at": "2023-01-02T00:00:00Z"
    }
  ]
}
```

#### Create Comment
```http
POST /api/v1/gists/{gist_id}/comments
Authorization: Bearer token
Content-Type: application/json

{
  "body": "This is a great example! Thanks for sharing."
}
```

#### Update Comment
```http
PATCH /api/v1/gists/{gist_id}/comments/{comment_id}
Authorization: Bearer token
Content-Type: application/json

{
  "body": "Updated comment text"
}
```

#### Delete Comment
```http
DELETE /api/v1/gists/{gist_id}/comments/{comment_id}
Authorization: Bearer token
```

### Organizations

#### List Organizations
```http
GET /api/v1/organizations
Authorization: Bearer token
```

#### Get Organization
```http
GET /api/v1/organizations/{org_name}
Authorization: Bearer token
```

#### Create Organization
```http
POST /api/v1/organizations
Authorization: Bearer token
Content-Type: application/json

{
  "name": "my-company",
  "display_name": "My Company",
  "description": "Our company's gists",
  "website": "https://company.com",
  "avatar_url": "https://company.com/logo.png"
}
```

#### Update Organization
```http
PATCH /api/v1/organizations/{org_name}
Authorization: Bearer token
Content-Type: application/json

{
  "display_name": "Updated Company Name",
  "description": "Updated description"
}
```

#### List Organization Members
```http
GET /api/v1/organizations/{org_name}/members
Authorization: Bearer token
```

#### Add Organization Member
```http
PUT /api/v1/organizations/{org_name}/members/{username}
Authorization: Bearer token
Content-Type: application/json

{
  "role": "member"
}
```

#### Remove Organization Member
```http
DELETE /api/v1/organizations/{org_name}/members/{username}
Authorization: Bearer token
```

#### List Organization Gists
```http
GET /api/v1/organizations/{org_name}/gists
Authorization: Bearer token
```

### Teams

#### List Teams
```http
GET /api/v1/organizations/{org_name}/teams
Authorization: Bearer token
```

#### Create Team
```http
POST /api/v1/organizations/{org_name}/teams
Authorization: Bearer token
Content-Type: application/json

{
  "name": "backend-team",
  "display_name": "Backend Team",
  "description": "Backend developers",
  "privacy": "closed"
}
```

#### Get Team
```http
GET /api/v1/organizations/{org_name}/teams/{team_name}
Authorization: Bearer token
```

#### Update Team
```http
PATCH /api/v1/organizations/{org_name}/teams/{team_name}
Authorization: Bearer token
Content-Type: application/json

{
  "display_name": "Updated Team Name",
  "description": "Updated description"
}
```

#### List Team Members
```http
GET /api/v1/organizations/{org_name}/teams/{team_name}/members
Authorization: Bearer token
```

#### Add Team Member
```http
PUT /api/v1/organizations/{org_name}/teams/{team_name}/members/{username}
Authorization: Bearer token
```

#### Remove Team Member
```http
DELETE /api/v1/organizations/{org_name}/teams/{team_name}/members/{username}
Authorization: Bearer token
```

### Webhooks

#### List Webhooks
```http
GET /api/v1/webhooks
Authorization: Bearer token
```

#### Create Webhook
```http
POST /api/v1/webhooks
Authorization: Bearer token
Content-Type: application/json

{
  "url": "https://example.com/webhook",
  "events": ["gist.created", "gist.updated", "gist.deleted"],
  "secret": "webhook_secret",
  "active": true,
  "ssl_verify": true
}
```

#### Get Webhook
```http
GET /api/v1/webhooks/{webhook_id}
Authorization: Bearer token
```

#### Update Webhook
```http
PATCH /api/v1/webhooks/{webhook_id}
Authorization: Bearer token
Content-Type: application/json

{
  "url": "https://new-endpoint.com/webhook",
  "events": ["gist.created", "gist.updated"],
  "active": false
}
```

#### Delete Webhook
```http
DELETE /api/v1/webhooks/{webhook_id}
Authorization: Bearer token
```

#### Test Webhook
```http
POST /api/v1/webhooks/{webhook_id}/test
Authorization: Bearer token
```

### Search

#### Search Gists
```http
GET /api/v1/search/gists?q=hello world&language=go&sort=updated&order=desc
Authorization: Bearer token
```

Query parameters:
- `q` - Search query
- `language` - Filter by language
- `user` - Filter by username
- `sort` - Sort by: `created`, `updated`, `stars`, `forks`
- `order` - Order: `asc`, `desc`
- `page` - Page number
- `limit` - Items per page

#### Search Users
```http
GET /api/v1/search/users?q=john&sort=joined&order=desc
Authorization: Bearer token
```

#### Search Organizations
```http
GET /api/v1/search/organizations?q=company
Authorization: Bearer token
```

### Statistics

#### User Statistics
```http
GET /api/v1/users/{username}/stats
Authorization: Bearer token
```

Response:
```json
{
  "gists": {
    "total": 25,
    "public": 20,
    "private": 5
  },
  "stars": {
    "received": 150,
    "given": 75
  },
  "followers": 45,
  "following": 32,
  "activity": {
    "commits_last_year": 250,
    "most_used_languages": [
      {"language": "Go", "count": 15},
      {"language": "Python", "count": 8},
      {"language": "JavaScript", "count": 2}
    ]
  }
}
```

#### Global Statistics (Admin only)
```http
GET /api/v1/admin/stats
Authorization: Bearer admin_token
```

Response:
```json
{
  "users": {
    "total": 1250,
    "active_last_month": 450,
    "new_this_month": 32
  },
  "gists": {
    "total": 5432,
    "public": 4321,
    "private": 1111
  },
  "organizations": 45,
  "storage": {
    "used_bytes": 1073741824,
    "files": 12450
  }
}
```

## Webhook Events

Webhooks are sent for various events. The payload includes an event type and relevant data.

### Event Types

- `gist.created` - New gist created
- `gist.updated` - Gist modified
- `gist.deleted` - Gist deleted
- `gist.starred` - Gist starred by user
- `gist.unstarred` - Gist unstarred by user
- `gist.forked` - Gist forked by user
- `comment.created` - Comment added to gist
- `comment.updated` - Comment modified
- `comment.deleted` - Comment deleted
- `user.created` - New user registered
- `organization.created` - Organization created
- `organization.member_added` - Member added to organization

### Webhook Payload

```json
{
  "event": "gist.created",
  "timestamp": "2023-01-01T12:00:00Z",
  "delivery_id": "12345678-1234-5678-9012-123456789012",
  "data": {
    "gist": {
      "id": "gist_id_here",
      "title": "Example Gist",
      "description": "A sample code snippet",
      "visibility": "public",
      "language": "go",
      "owner": {
        "username": "johndoe",
        "email": "john@example.com"
      },
      "files": [
        {
          "filename": "main.go",
          "language": "go",
          "content": "package main..."
        }
      ]
    }
  }
}
```

### Webhook Security

Webhooks include HMAC-SHA256 signature in the header:

```
X-CasGists-Signature-256: sha256=hash_value
```

Verify webhook authenticity:

```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected_signature = 'sha256=' + hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected_signature)
```

## SDK Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class CasGistsAPI {
  constructor(baseURL, token) {
    this.client = axios.create({
      baseURL: baseURL + '/api/v1',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createGist(title, description, files, visibility = 'public') {
    const response = await this.client.post('/gists', {
      title,
      description,
      files,
      visibility
    });
    return response.data;
  }

  async getGist(gistId) {
    const response = await this.client.get(`/gists/${gistId}`);
    return response.data;
  }

  async listGists(options = {}) {
    const response = await this.client.get('/gists', { params: options });
    return response.data;
  }

  async starGist(gistId) {
    await this.client.put(`/gists/${gistId}/star`);
  }
}

// Usage
const api = new CasGistsAPI('https://gists.example.com', 'your_token');

const gist = await api.createGist(
  'Hello World',
  'A simple example',
  [{ filename: 'hello.js', content: 'console.log("Hello, World!");' }]
);

console.log('Created gist:', gist.id);
```

### Python

```python
import requests

class CasGistsAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url + '/api/v1'
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })

    def create_gist(self, title, description, files, visibility='public'):
        response = self.session.post(f'{self.base_url}/gists', json={
            'title': title,
            'description': description,
            'files': files,
            'visibility': visibility
        })
        response.raise_for_status()
        return response.json()

    def get_gist(self, gist_id):
        response = self.session.get(f'{self.base_url}/gists/{gist_id}')
        response.raise_for_status()
        return response.json()

    def list_gists(self, **params):
        response = self.session.get(f'{self.base_url}/gists', params=params)
        response.raise_for_status()
        return response.json()

    def star_gist(self, gist_id):
        response = self.session.put(f'{self.base_url}/gists/{gist_id}/star')
        response.raise_for_status()

# Usage
api = CasGistsAPI('https://gists.example.com', 'your_token')

gist = api.create_gist(
    'Hello World',
    'A simple example',
    [{'filename': 'hello.py', 'content': 'print("Hello, World!")'}]
)

print(f'Created gist: {gist["id"]}')
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type CasGistsAPI struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

type CreateGistRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Files       []File `json:"files"`
    Visibility  string `json:"visibility"`
}

type File struct {
    Filename string `json:"filename"`
    Content  string `json:"content"`
}

type Gist struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    // ... other fields
}

func NewCasGistsAPI(baseURL, token string) *CasGistsAPI {
    return &CasGistsAPI{
        BaseURL: baseURL + "/api/v1",
        Token:   token,
        Client:  &http.Client{},
    }
}

func (api *CasGistsAPI) CreateGist(req CreateGistRequest) (*Gist, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", api.BaseURL+"/gists", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Authorization", "Bearer "+api.Token)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := api.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var gist Gist
    err = json.NewDecoder(resp.Body).Decode(&gist)
    if err != nil {
        return nil, err
    }

    return &gist, nil
}

// Usage
func main() {
    api := NewCasGistsAPI("https://gists.example.com", "your_token")
    
    gist, err := api.CreateGist(CreateGistRequest{
        Title:       "Hello World",
        Description: "A simple example",
        Files: []File{
            {Filename: "hello.go", Content: `package main\n\nfunc main() {\n    println("Hello, World!")\n}`},
        },
        Visibility: "public",
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Created gist: %s\n", gist.ID)
}
```

## Rate Limiting Best Practices

1. **Respect Rate Limits**: Check headers and implement backoff
2. **Cache Responses**: Avoid repeated requests for same data
3. **Use Conditional Requests**: Use ETags and If-Modified-Since headers
4. **Batch Operations**: Combine multiple operations when possible
5. **Implement Retry Logic**: Handle temporary failures gracefully

## Error Handling Best Practices

1. **Check Status Codes**: Always verify HTTP response status
2. **Parse Error Messages**: Use structured error responses
3. **Implement Exponential Backoff**: For rate limiting and server errors
4. **Log API Errors**: Keep detailed logs for debugging
5. **Provide User Feedback**: Show meaningful error messages to users

## API Versioning

The API follows semantic versioning:

- **Major version** (`v1`, `v2`): Breaking changes
- **Minor updates**: New features (backwards compatible)
- **Patch updates**: Bug fixes (backwards compatible)

Current version: `v1.0.0`

Always use the versioned endpoint (`/api/v1/`) in your applications. Version deprecation will be announced with 6 months notice.

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:

```
GET /api/v1/openapi.json
```

You can use this with tools like Swagger UI, Postman, or code generators to create client libraries in your preferred language.

## Support

For API questions and issues:

- **Documentation**: This guide and inline examples
- **OpenAPI Spec**: `/api/v1/openapi.json`
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General API questions and usage help
- **Email**: api-support@casgists.com