# CasGists Feature Reference

## Core Features

### Authentication & Security

#### JWT Authentication
- **Access Tokens**: Short-lived tokens (15 minutes default)
- **Refresh Tokens**: Long-lived tokens (7 days default)
- **Token Rotation**: Automatic refresh token rotation on use
- **Revocation**: Individual token revocation support

#### Two-Factor Authentication (2FA)
- **TOTP Support**: Time-based one-time passwords
- **QR Code Generation**: Easy setup with authenticator apps
- **Backup Codes**: 10 single-use recovery codes
- **Enforcement**: Can be required for admin accounts

#### Session Management
- **Concurrent Sessions**: Configurable limit per user
- **Session Tracking**: IP address, user agent, last activity
- **Session Termination**: Users can view and terminate sessions
- **Idle Timeout**: Automatic session expiration

### Gist Management

#### Gist Types
- **Public**: Visible to everyone, searchable
- **Private**: Only visible to owner and collaborators
- **Unlisted**: Accessible via direct link only

#### File Operations
- **Multiple Files**: Up to 100 files per gist
- **Syntax Highlighting**: 200+ languages supported
- **File Size Limit**: 10MB per file (configurable)
- **Binary Files**: Image preview support

#### Social Features
- **Stars**: Bookmark favorite gists
- **Forks**: Create personal copies
- **Comments**: Threaded discussions
- **Watch**: Get notifications on updates

### Search System

#### Search Capabilities
- **Full-Text Search**: Content and metadata
- **Advanced Filters**:
  - `user:username` - Filter by author
  - `language:python` - Filter by language
  - `stars:>10` - Filter by star count
  - `created:2024-01-01` - Filter by date
  - `tag:snippet` - Filter by tags

#### Search Performance
- **Redis Integration**: Sub-millisecond response times
- **SQLite FTS Fallback**: When Redis unavailable
- **Search Suggestions**: Autocomplete as you type
- **Popular Searches**: Track and display trending searches

### Organization Management

#### Organization Features
- **Team Management**: Create teams with different permissions
- **Role-Based Access**:
  - **Owner**: Full control
  - **Admin**: Manage members and gists
  - **Member**: Create and manage own gists
  - **Viewer**: Read-only access

#### Organization Gists
- **Shared Ownership**: Organization owns the gists
- **Team Collaboration**: Multiple members can edit
- **Access Control**: Fine-grained permissions

### Webhook System

#### Event Types
- **Gist Events**:
  - `gist.created`
  - `gist.updated`
  - `gist.deleted`
  - `gist.starred`
  - `gist.forked`
- **User Events**:
  - `user.created`
  - `user.updated`
  - `user.followed`

#### Webhook Features
- **HMAC Signatures**: Verify webhook authenticity
- **Retry Logic**: Exponential backoff with jitter
- **Circuit Breaker**: Prevent cascading failures
- **Delivery History**: Track success/failure
- **Filtering**: Complex rules for event filtering

#### Webhook Filtering Rules
```json
{
  "rules": [
    {
      "field": "event_type",
      "operator": "equals",
      "value": "gist.created"
    },
    {
      "field": "gist.language",
      "operator": "in",
      "value": ["python", "go", "javascript"]
    }
  ],
  "logic": "AND"
}
```

### Email System

#### Email Types
- **Welcome Email**: New user registration
- **Email Verification**: Confirm email address
- **Password Reset**: Secure password recovery
- **Notifications**: Gist activity, mentions
- **Weekly Digest**: Summary of activity

#### Email Configuration
- **SMTP Support**: Any SMTP server
- **Template Engine**: HTML and plain text
- **Queue System**: Retry failed deliveries
- **Unsubscribe**: Per-category preferences

### Performance Features

#### Caching Strategy
- **Multi-Level Cache**:
  - L1: In-memory cache (hot data)
  - L2: Redis cache (distributed)
  - L3: Database (persistent)
- **Cache Invalidation**: Smart invalidation on updates
- **Cache Warming**: Preload popular content

#### Database Optimization
- **Connection Pooling**: Reuse connections
- **Query Optimization**: Selective loading
- **Batch Operations**: Reduce round trips
- **Index Management**: Automatic index creation

## Advanced Features

### Path Variables System

#### Variable Substitution
- **Syntax**: `{VARIABLE_NAME}`
- **Environment Variables**: All env vars available
- **Platform Defaults**: OS-specific paths
- **Nested Variables**: `{CASGISTS_DATA_DIR}/repos`

#### Built-in Variables
```yaml
# Linux/macOS (privileged)
CASGISTS_DATA_DIR: /var/lib/casgists
CASGISTS_LOG_DIR: /var/log/casgists
CASGISTS_CONFIG_DIR: /etc/casgists

# Windows (privileged)
CASGISTS_DATA_DIR: C:\ProgramData\casgists
CASGISTS_LOG_DIR: C:\ProgramData\casgists\logs
CASGISTS_CONFIG_DIR: C:\ProgramData\casgists\config
```

### Privilege Escalation

#### Smart Detection
- **Unix Systems**: Detects sudo requirement
- **Windows**: Detects UAC requirement
- **Automatic Prompting**: When needed for operations
- **Graceful Fallback**: User-mode directories

#### Escalation Operations
- **System-wide Install**: `/usr/local/bin`
- **Service Installation**: Systemd/Windows Service
- **Port Binding**: Ports below 1024
- **System Directories**: Write to `/etc`, `/var`

### Setup Wizard

#### Configuration Steps
1. **Welcome**: Introduction and overview
2. **Database**: Type selection and connection
3. **Admin Account**: Create first admin user
4. **Server Settings**: URL, port, domain
5. **Features**: Enable/disable features
6. **Security**: Set secrets and policies
7. **Notifications**: Email configuration
8. **Review**: Confirm all settings

### Migration Tools

#### OpenGist Migration
- **Complete Data Transfer**:
  - Users and profiles
  - Gists and files
  - Stars and likes
  - SSH keys (tracked)
- **Password Handling**: Optional reset required
- **ID Mapping**: Preserve relationships
- **Progress Tracking**: Real-time status

#### GitHub Import
- **API Integration**: Uses GitHub API v3
- **Import Options**:
  - Public gists only
  - Include private gists
  - Include starred gists
  - Preserve timestamps
- **Rate Limiting**: Respects GitHub limits
- **Batch Processing**: Efficient bulk import

### Git Integration

#### Native Git Operations
- **go-git Backend**: No external dependencies
- **Repository Management**:
  - Create bare repositories
  - Handle commits and branches
  - Track history and diffs
- **Clone Support**: Standard Git protocols
- **Web Hooks**: Post-receive hooks

#### Git URLs
```bash
# HTTPS Clone
git clone https://gists.example.com/username/gist-id.git

# SSH Clone (future)
git clone git@gists.example.com:username/gist-id.git
```

### Custom Domains

#### Domain Features
- **Per-Organization Domains**: `org.gists.com`
- **SSL Management**:
  - Let's Encrypt integration
  - Custom certificates
  - Auto-renewal
- **Domain Verification**: DNS TXT records
- **Subdomain Routing**: Automatic routing

### GDPR Compliance

#### Data Protection
- **Right to Access**: Export all user data
- **Right to Erasure**: Complete data deletion
- **Data Portability**: Standard formats (JSON/CSV)
- **Consent Tracking**: Audit trail of consent

#### Export Formats
```json
{
  "user": {
    "id": "...",
    "username": "...",
    "email": "...",
    "created_at": "..."
  },
  "gists": [...],
  "comments": [...],
  "stars": [...],
  "audit_logs": [...]
}
```

### Transfer System

#### Transfer Workflow
1. **Request Creation**: Owner initiates transfer
2. **Notification**: Target user notified
3. **Review Period**: 7 days to accept/reject
4. **Approval**: Target user accepts
5. **Execution**: Ownership transferred
6. **History**: Permanent record created

#### Transfer Options
- **User to User**: Personal transfers
- **User to Organization**: Team ownership
- **Organization to User**: Personal ownership
- **Organization to Organization**: Team transfers

### Audit Logging

#### Audit Events
- **Authentication**: Login, logout, 2FA
- **Resource Access**: View, create, update, delete
- **Security Events**: Failed attempts, rate limits
- **Administrative**: User management, settings
- **Compliance**: GDPR requests, transfers

#### Audit Features
- **Comprehensive Tracking**: All HTTP requests
- **Sensitive Data Redaction**: Passwords, tokens
- **Search and Filter**: Find specific events
- **Retention Policies**: Configurable duration
- **Export Capabilities**: CSV, JSON formats

## CLI Administration

### User Management
```bash
# Create user
./casgists user create --username john --email john@example.com

# Grant admin
./casgists user grant-admin john

# Reset password
./casgists user reset-password john

# Disable account
./casgists user disable john
```

### Database Operations
```bash
# Run migrations
./casgists db migrate

# Create backup
./casgists db backup --output backup.sql

# Restore backup
./casgists db restore --input backup.sql

# Optimize tables
./casgists db optimize
```

### System Maintenance
```bash
# Clean old sessions
./casgists cleanup sessions --days 30

# Remove orphaned files
./casgists cleanup files

# Purge old logs
./casgists cleanup logs --days 90

# Rebuild search index
./casgists search reindex
```

### Email Management
```bash
# Test email configuration
./casgists email test recipient@example.com

# Process email queue
./casgists email process-queue

# View queue status
./casgists email queue-status

# Retry failed emails
./casgists email retry-failed
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh tokens
- `POST /api/v1/auth/2fa/setup` - Setup 2FA
- `POST /api/v1/auth/2fa/verify` - Verify 2FA code

### Gists
- `GET /api/v1/gists` - List gists
- `POST /api/v1/gists` - Create gist
- `GET /api/v1/gists/:id` - Get gist
- `PUT /api/v1/gists/:id` - Update gist
- `DELETE /api/v1/gists/:id` - Delete gist
- `POST /api/v1/gists/:id/star` - Star gist
- `POST /api/v1/gists/:id/fork` - Fork gist

### Organizations
- `GET /api/v1/orgs` - List organizations
- `POST /api/v1/orgs` - Create organization
- `GET /api/v1/orgs/:name` - Get organization
- `PUT /api/v1/orgs/:name` - Update organization
- `POST /api/v1/orgs/:name/members` - Add member

### Webhooks
- `GET /api/v1/webhooks` - List webhooks
- `POST /api/v1/webhooks` - Create webhook
- `PUT /api/v1/webhooks/:id` - Update webhook
- `DELETE /api/v1/webhooks/:id` - Delete webhook
- `POST /api/v1/webhooks/:id/ping` - Test webhook

### GDPR
- `POST /api/v1/gdpr/export` - Request data export
- `GET /api/v1/gdpr/export/:id` - Download export
- `POST /api/v1/gdpr/delete` - Request deletion
- `GET /api/v1/gdpr/requests` - List requests

### Transfers
- `POST /api/v1/transfers` - Create transfer
- `GET /api/v1/transfers/:id` - Get transfer
- `POST /api/v1/transfers/:id/accept` - Accept transfer
- `POST /api/v1/transfers/:id/reject` - Reject transfer
- `DELETE /api/v1/transfers/:id` - Cancel transfer

## Configuration Reference

### Server Configuration
```yaml
server:
  host: 0.0.0.0
  port: 3000
  url: https://gists.example.com
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 30s
```

### Security Configuration
```yaml
security:
  secret_key: minimum-32-character-secret
  jwt_secret: jwt-signing-secret
  jwt_access_expiry: 15m
  jwt_refresh_expiry: 168h
  session_secret: session-encryption-secret
  session_lifetime: 24h
  bcrypt_cost: 12
  password_min_length: 8
  password_require_uppercase: true
  password_require_numbers: true
  password_require_special: true
```

### Feature Flags
```yaml
features:
  registration: true
  email_verification: true
  two_factor_auth: true
  organizations: true
  webhooks: true
  public_gists: true
  anonymous_gists: false
  import_export: true
  api_tokens: true
```

### Rate Limiting
```yaml
rate_limit:
  enabled: true
  requests_per_minute: 60
  burst_size: 10
  auth_multiplier: 2
  skip_private_ips: true
```

### Metrics Configuration
```yaml
metrics:
  enabled: true
  path: /metrics
  include_method_label: true
  include_path_label: true
  buckets: [0.1, 0.5, 1, 2.5, 5, 10]
```