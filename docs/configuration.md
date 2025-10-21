# Configuration Reference

CasGists can be configured using YAML configuration files, environment variables, or command-line flags. This document provides a complete reference for all configuration options.

## Configuration Hierarchy

Configuration is loaded in this order (later sources override earlier ones):

1. Default values
2. Configuration file (`config.yaml`)
3. Environment variables (`CASGISTS_*`)
4. Command-line flags

## Configuration File Locations

### Default Locations

- **System installation**: `/etc/casgists/config.yaml`
- **Local installation**: `./config.yaml` (in working directory)
- **Docker**: `/config/config.yaml`

### Custom Location

```bash
# Specify custom config file
casgists serve --config /path/to/custom-config.yaml

# Environment variable
export CASGISTS_CONFIG_FILE=/path/to/config.yaml
```

## Complete Configuration Reference

### Server Configuration

```yaml
server:
  # Host to bind to (default: 127.0.0.1)
  host: 0.0.0.0
  
  # Port to listen on (default: random 64000-64999)
  port: 64080
  
  # Base URL for the application (used in emails, webhooks)
  base_url: https://gists.yourdomain.com
  
  # Read timeout for requests
  read_timeout: 30s
  
  # Write timeout for responses  
  write_timeout: 30s
  
  # Maximum request body size (default: 10MB)
  max_request_size: 10485760
  
  # Enable HTTPS (requires cert_file and key_file)
  https: false
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem
  
  # Auto-generate self-signed certificate for HTTPS
  auto_cert: false
```

### Database Configuration

#### SQLite (Default)
```yaml
database:
  type: sqlite
  path: ${DATA_DIR}/casgists.db
  
  # SQLite-specific options
  sqlite:
    # Enable WAL mode for better concurrency
    wal_mode: true
    
    # Enable foreign key constraints
    foreign_keys: true
    
    # Connection timeout
    timeout: 5s
```

#### PostgreSQL
```yaml
database:
  type: postgresql
  host: localhost
  port: 5432
  name: casgists
  user: casgists
  password: secure_password
  
  # SSL configuration
  ssl_mode: require # disable, allow, prefer, require, verify-ca, verify-full
  ssl_cert: /path/to/client-cert.pem
  ssl_key: /path/to/client-key.pem
  ssl_ca: /path/to/ca-cert.pem
  
  # Connection pool settings
  max_open_connections: 25
  max_idle_connections: 5
  connection_max_lifetime: 5m
```

#### MySQL
```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  name: casgists
  user: casgists
  password: secure_password
  
  # MySQL-specific options
  mysql:
    charset: utf8mb4
    collation: utf8mb4_unicode_ci
    parse_time: true
    timeout: 30s
  
  # Connection pool settings
  max_open_connections: 25
  max_idle_connections: 5
  connection_max_lifetime: 5m
```

### Paths Configuration

```yaml
paths:
  # Main data directory (supports ${DATA_DIR} variable)
  data_dir: /var/lib/casgists
  
  # Git repositories storage
  repo_dir: ${DATA_DIR}/repos
  
  # Cache directory
  cache_dir: ${DATA_DIR}/cache
  
  # Upload directory for files
  upload_dir: ${DATA_DIR}/uploads
  
  # Backup storage
  backup_dir: ${DATA_DIR}/backups
  
  # GDPR export files
  gdpr_exports: ${DATA_DIR}/gdpr_exports
  
  # Log files
  log_dir: ${DATA_DIR}/logs
```

### Security Configuration

```yaml
security:
  # Secret key for JWT tokens and CSRF protection (REQUIRED)
  secret_key: your-super-secret-key-at-least-32-characters-long
  
  # JWT token expiration
  jwt_expiry: 24h
  
  # Session configuration
  session:
    lifetime: 24h
    cookie_name: casgists_session
    secure: true # Set to true when using HTTPS
    same_site: lax # none, lax, strict
  
  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_size: 10
  
  # CORS settings
  cors:
    enabled: false
    allowed_origins: ["https://example.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["Authorization", "Content-Type"]
    credentials: true
  
  # Content Security Policy
  csp:
    enabled: true
    default_src: "'self'"
    script_src: "'self' 'unsafe-inline'"
    style_src: "'self' 'unsafe-inline'"
    img_src: "'self' data: https:"
```

### Authentication Configuration

```yaml
auth:
  # Local authentication
  local:
    enabled: true
    registration: true
    password_min_length: 8
    require_email_verification: true
  
  # OAuth2 providers
  oauth2:
    github:
      enabled: false
      client_id: your_github_client_id
      client_secret: your_github_client_secret
      scopes: ["user:email"]
    
    gitlab:
      enabled: false
      client_id: your_gitlab_client_id
      client_secret: your_gitlab_client_secret
      url: https://gitlab.com
    
    google:
      enabled: false
      client_id: your_google_client_id
      client_secret: your_google_client_secret
    
    microsoft:
      enabled: false
      client_id: your_microsoft_client_id
      client_secret: your_microsoft_client_secret
      tenant: common # or specific tenant ID
  
  # LDAP/Active Directory
  ldap:
    enabled: false
    host: ldap.example.com
    port: 389
    use_ssl: false
    use_tls: true
    bind_dn: cn=admin,dc=example,dc=com
    bind_password: admin_password
    user_search:
      base_dn: ou=users,dc=example,dc=com
      filter: (uid=%s)
      username_attr: uid
      email_attr: mail
      name_attr: cn
    group_search:
      enabled: true
      base_dn: ou=groups,dc=example,dc=com
      filter: (member=%s)
      name_attr: cn
  
  # SAML 2.0
  saml:
    enabled: false
    entity_id: https://gists.yourdomain.com/saml/metadata
    sso_url: https://idp.example.com/saml/sso
    cert_file: /path/to/idp-cert.pem
    private_key_file: /path/to/sp-private-key.pem
    certificate_file: /path/to/sp-certificate.pem
```

### Two-Factor Authentication

```yaml
auth:
  two_factor:
    # TOTP (Time-based One-Time Password)
    totp:
      enabled: true
      issuer: CasGists
      period: 30
      digits: 6
    
    # WebAuthn/FIDO2
    webauthn:
      enabled: true
      display_name: CasGists
      rp_id: yourdomain.com
      rp_origins: ["https://gists.yourdomain.com"]
      require_resident_key: false
      user_verification: preferred # required, preferred, discouraged
```

### Email Configuration

```yaml
email:
  enabled: true
  
  # SMTP settings
  smtp_host: smtp.gmail.com
  smtp_port: 587
  smtp_username: your-email@gmail.com
  smtp_password: your-app-password
  smtp_encryption: tls # none, tls, ssl
  
  # Email addresses
  from_email: noreply@yourdomain.com
  from_name: CasGists
  reply_to: support@yourdomain.com
  
  # Email templates
  templates:
    # Custom template directory (optional)
    path: /path/to/custom/templates
  
  # Email sending options
  max_retries: 3
  retry_delay: 5s
  timeout: 30s
```

### Search Configuration

```yaml
search:
  # Search backend: sqlite_fts, redis
  backend: sqlite_fts
  
  # SQLite FTS configuration
  sqlite_fts:
    # Enable full-text search
    enabled: true
    
    # Rebuild index on startup
    rebuild_index: false
  
  # Redis configuration
  redis:
    enabled: false
    host: localhost
    port: 6379
    password: ""
    database: 0
    
    # Connection pool
    pool_size: 10
    min_idle_connections: 5
    
    # Index settings
    index_name: casgists_search
    rebuild_index: false
```

### Cache Configuration

```yaml
cache:
  # Cache backend: memory, redis
  backend: memory
  
  # Memory cache
  memory:
    # Maximum number of items
    max_items: 10000
    
    # Default TTL
    default_ttl: 1h
  
  # Redis cache
  redis:
    enabled: false
    host: localhost
    port: 6379
    password: ""
    database: 1
    
    # Default TTL
    default_ttl: 1h
```

### Storage Configuration

```yaml
storage:
  # Storage backend: local, s3
  backend: local
  
  # Local storage
  local:
    path: ${DATA_DIR}/uploads
  
  # S3-compatible storage
  s3:
    enabled: false
    endpoint: https://s3.amazonaws.com
    region: us-east-1
    bucket: casgists-uploads
    access_key: your_access_key
    secret_key: your_secret_key
    
    # Custom endpoint for S3-compatible services
    custom_endpoint: https://minio.example.com
    
    # Force path style (for MinIO and others)
    path_style: false
    
    # SSL verification
    ssl_verify: true
```

### Logging Configuration

```yaml
logging:
  # Log level: debug, info, warn, error
  level: info
  
  # Log format: text, json
  format: text
  
  # Log output: stdout, file
  output: file
  
  # Log file path
  file: ${DATA_DIR}/logs/casgists.log
  
  # Log rotation
  rotation:
    enabled: true
    max_size: 100 # MB
    max_files: 10
    max_age: 30 # days
    compress: true
  
  # Request logging
  http:
    enabled: true
    skip_paths: ["/health", "/metrics"]
    
  # Database query logging
  database:
    enabled: false
    slow_threshold: 200ms
```

### Features Configuration

```yaml
features:
  # User registration
  registration: true
  
  # Anonymous gists (public without account)
  anonymous_gists: true
  
  # Organization support
  organizations: true
  
  # Team collaboration
  teams: true
  
  # Social features (following, stars)
  social: true
  
  # Webhook system
  webhooks: true
  
  # API access
  api: true
  
  # Git operations via API
  git_api: true
  
  # Email notifications
  email_notifications: true
  
  # Real-time notifications
  realtime_notifications: true
```

### Limits Configuration

```yaml
limits:
  # Maximum gist size (bytes)
  max_gist_size: 10485760 # 10MB
  
  # Maximum file size per file (bytes)
  max_file_size: 1048576 # 1MB
  
  # Maximum files per gist
  max_files_per_gist: 100
  
  # Maximum filename length
  max_filename_length: 255
  
  # Maximum gists per user (0 = unlimited)
  max_gists_per_user: 1000
  
  # Maximum organizations per user
  max_orgs_per_user: 10
  
  # Rate limiting
  rate_limits:
    # API requests per minute
    api_requests: 60
    
    # Gist creations per hour
    gist_creation: 10
    
    # Login attempts per hour
    login_attempts: 5
```

### Webhook Configuration

```yaml
webhooks:
  # Enable webhook system
  enabled: true
  
  # Webhook timeout
  timeout: 30s
  
  # Maximum retries
  max_retries: 3
  
  # Retry delays
  retry_delays: [5s, 30s, 300s]
  
  # User-Agent header
  user_agent: CasGists-Webhooks/1.0
  
  # SSL verification
  ssl_verify: true
  
  # Maximum payload size
  max_payload_size: 1048576 # 1MB
```

### Backup Configuration

```yaml
backup:
  # Enable automatic backups
  enabled: false
  
  # Backup schedule (cron format)
  schedule: "0 2 * * *" # Daily at 2 AM
  
  # Backup retention
  retention:
    daily: 7
    weekly: 4
    monthly: 12
  
  # Backup location
  path: ${DATA_DIR}/backups
  
  # Compression
  compression: gzip
  
  # Include uploads in backup
  include_uploads: true
  
  # S3 backup
  s3:
    enabled: false
    bucket: casgists-backups
    prefix: backups/
```

### Compliance Configuration

```yaml
compliance:
  # GDPR compliance
  gdpr:
    enabled: true
    contact_email: privacy@yourdomain.com
    
    # Data retention periods (days)
    data_retention:
      user_data: 2555 # 7 years
      audit_logs: 2555
      deleted_data: 30
  
  # Audit logging
  audit:
    enabled: true
    
    # Events to log
    events:
      - user_login
      - user_logout
      - gist_create
      - gist_update
      - gist_delete
      - admin_action
  
  # Cookie consent
  cookies:
    enabled: false
    policy_url: https://yourdomain.com/cookies
```

## Environment Variables

All configuration options can be set using environment variables with the `CASGISTS_` prefix:

```bash
# Server
export CASGISTS_SERVER_HOST=0.0.0.0
export CASGISTS_SERVER_PORT=64080
export CASGISTS_SERVER_BASE_URL=https://gists.example.com

# Database
export CASGISTS_DATABASE_TYPE=postgresql
export CASGISTS_DATABASE_HOST=localhost
export CASGISTS_DATABASE_PORT=5432
export CASGISTS_DATABASE_NAME=casgists
export CASGISTS_DATABASE_USER=casgists
export CASGISTS_DATABASE_PASSWORD=secure_password

# Security
export CASGISTS_SECURITY_SECRET_KEY=your-secret-key
export CASGISTS_SECURITY_JWT_EXPIRY=24h

# Email
export CASGISTS_EMAIL_SMTP_HOST=smtp.gmail.com
export CASGISTS_EMAIL_SMTP_PORT=587
export CASGISTS_EMAIL_SMTP_USERNAME=user@gmail.com
export CASGISTS_EMAIL_SMTP_PASSWORD=app-password

# Features
export CASGISTS_FEATURES_REGISTRATION=true
export CASGISTS_FEATURES_ANONYMOUS_GISTS=false
export CASGISTS_FEATURES_ORGANIZATIONS=true
```

## Command-Line Flags

```bash
# Basic flags
casgists serve --config /path/to/config.yaml
casgists serve --port 8080
casgists serve --host 0.0.0.0

# Database flags
casgists serve --db-type postgresql --db-host localhost --db-name casgists

# Debug flags
casgists serve --debug --log-level debug
casgists serve --dev-mode
```

## Configuration Validation

```bash
# Validate configuration file
casgists config-check

# Validate and show resolved configuration
casgists config-check --show-config

# Test database connection
casgists config-check --test-db

# Test email configuration
casgists config-check --test-email
```

## Example Configurations

### Minimal Configuration

```yaml
server:
  port: 64080

database:
  type: sqlite
  path: ./casgists.db

security:
  secret_key: generate-a-secure-random-key-here
```

### Production Configuration

```yaml
server:
  host: 0.0.0.0
  port: 64080
  base_url: https://gists.company.com

database:
  type: postgresql
  host: postgres.company.com
  port: 5432
  name: casgists
  user: casgists
  password: ${DB_PASSWORD}
  ssl_mode: require

security:
  secret_key: ${SECRET_KEY}
  session:
    secure: true
    same_site: strict
  rate_limit:
    enabled: true
    requests_per_minute: 100

auth:
  local:
    registration: false
  ldap:
    enabled: true
    host: ldap.company.com
    bind_dn: cn=service,ou=services,dc=company,dc=com
    bind_password: ${LDAP_PASSWORD}

email:
  enabled: true
  smtp_host: smtp.company.com
  smtp_port: 587
  smtp_username: casgists@company.com
  smtp_password: ${SMTP_PASSWORD}
  from_email: gists@company.com
  from_name: Company Gists

search:
  backend: redis
  redis:
    enabled: true
    host: redis.company.com

logging:
  level: info
  format: json
  file: ${DATA_DIR}/logs/casgists.log

features:
  registration: false
  anonymous_gists: false
  organizations: true

compliance:
  gdpr:
    enabled: true
    contact_email: privacy@company.com
  audit:
    enabled: true
```

### Development Configuration

```yaml
server:
  host: 127.0.0.1
  port: 3000

database:
  type: sqlite
  path: ./dev.db

security:
  secret_key: dev-secret-key-not-for-production

logging:
  level: debug
  output: stdout

features:
  registration: true
  anonymous_gists: true
  organizations: true

auth:
  local:
    require_email_verification: false

email:
  enabled: false
```

## Configuration Best Practices

1. **Use Environment Variables for Secrets**: Never store passwords or keys in config files
2. **Enable HTTPS in Production**: Always use TLS for production deployments
3. **Configure Rate Limiting**: Protect against abuse and DDoS attacks
4. **Enable Audit Logging**: Required for compliance and security monitoring
5. **Set Appropriate Limits**: Prevent resource exhaustion
6. **Use External Databases**: PostgreSQL/MySQL for production workloads
7. **Configure Backups**: Automated, encrypted backups to external storage
8. **Monitor Configuration**: Use `config-check` command regularly
9. **Version Control**: Keep configuration templates in version control
10. **Documentation**: Document any custom configuration for your team