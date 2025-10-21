# Production Deployment Guide

This guide covers deploying CasGists in a production environment.

## Prerequisites

- Linux server (Ubuntu 20.04+ or CentOS 8+)
- Go 1.21+ (for building from source)
- PostgreSQL 13+ or MySQL 8.0+ (optional, SQLite is default)
- Nginx or Caddy (for reverse proxy)
- Systemd (for service management)
- Minimum 2GB RAM, 10GB disk space

## Quick Start

```bash
# Download and deploy in one command
curl -sSL https://get.casgists.com | sudo bash -s -- --port 64080
```

## Installation Methods

### Method 1: Pre-built Binary (Recommended)

1. Download the latest release:
```bash
# Detect architecture and download
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv7l) ARCH="armv7" ;;
esac

wget "https://github.com/casapps/casgists/releases/latest/download/casgists-linux-${ARCH}"
chmod +x "casgists-linux-${ARCH}"
```

2. Install as system service:
```bash
sudo ./casgists-linux-${ARCH} install --port 64080
```

3. Start the service:
```bash
sudo systemctl start casgists
sudo systemctl enable casgists
```

### Method 2: Docker

1. Using Docker Compose:
```yaml
version: '3.8'
services:
  casgists:
    image: casapps/casgists:latest
    ports:
      - "64080:64080"
    volumes:
      - casgists_data:/data
      - ./config:/config
    environment:
      - CASGISTS_DB_TYPE=sqlite
      - CASGISTS_DB_PATH=/data/casgists.db
      - CASGISTS_SERVER_URL=https://gists.example.com
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:64080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  casgists_data:
```

2. Run with docker-compose:
```bash
docker-compose up -d
```

### Method 3: Kubernetes

See [kubernetes-deployment.yaml](./kubernetes-deployment.yaml) for a complete example.

### Method 4: Building from Source

1. Clone and build:
```bash
git clone https://github.com/casapps/casgists.git
cd casgists
make release
```

2. Install:
```bash
sudo ./build/casgists-linux-amd64 install --port 64080
```

## Configuration

### Environment Variables

```bash
# Core Settings
CASGISTS_LISTEN_PORT=64080                    # Port to listen on
CASGISTS_SERVER_URL=https://gists.example.com # Public URL
CASGISTS_SECRET_KEY=your-secret-key-here      # Generate with: openssl rand -hex 32

# Database (SQLite default)
CASGISTS_DB_TYPE=sqlite                       # sqlite|postgresql|mysql
CASGISTS_DB_PATH=/var/lib/casgists/data.db   # For SQLite

# Database (PostgreSQL)
CASGISTS_DB_TYPE=postgresql
CASGISTS_DB_HOST=localhost
CASGISTS_DB_PORT=5432
CASGISTS_DB_NAME=casgists
CASGISTS_DB_USER=casgists
CASGISTS_DB_PASSWORD=secure-password
CASGISTS_DB_SSL_MODE=require                  # disable|require|verify-full

# Storage
CASGISTS_DATA_DIR=/var/lib/casgists          # Main data directory
CASGISTS_GIT_ROOT=/var/lib/casgists/repos    # Git repositories
CASGISTS_BACKUP_DIR=/var/lib/casgists/backups

# Email (optional)
CASGISTS_SMTP_ENABLED=true
CASGISTS_SMTP_HOST=smtp.gmail.com
CASGISTS_SMTP_PORT=587
CASGISTS_SMTP_USERNAME=your-email@gmail.com
CASGISTS_SMTP_PASSWORD=your-app-password
CASGISTS_SMTP_FROM_EMAIL=noreply@example.com
CASGISTS_SMTP_FROM_NAME=CasGists

# Features
CASGISTS_FEATURES_REGISTRATION=true           # Allow new users
CASGISTS_FEATURES_ORGANIZATIONS=true          # Enable organizations
CASGISTS_FEATURES_SOCIAL=true                 # Enable social features
CASGISTS_FEATURES_API=true                    # Enable API access

# Security
CASGISTS_SECURITY_RATE_LIMIT=60              # Requests per minute
CASGISTS_SECURITY_MAX_FILE_SIZE=10485760      # 10MB
CASGISTS_SECURITY_ALLOWED_ORIGINS=https://gists.example.com
```

### Configuration File

Create `/etc/casgists/config.yaml`:

```yaml
server:
  port: 64080
  host: 0.0.0.0
  url: https://gists.example.com
  enable_https: false
  tls_cert: ""  # Path to cert file if enable_https is true
  tls_key: ""   # Path to key file if enable_https is true
  
database:
  type: postgresql  # sqlite, postgresql, mysql
  host: localhost
  port: 5432
  name: casgists
  user: casgists
  password: ${DB_PASSWORD}  # Use environment variable
  ssl_mode: require
  max_connections: 100
  max_idle_connections: 10
  connection_max_lifetime: 3600
  
storage:
  data_dir: /var/lib/casgists
  git_root: ${DATA_DIR}/repos
  backup_dir: ${DATA_DIR}/backups
  max_file_size: 10485760  # 10MB
  allowed_extensions: []   # Empty = all allowed
  
auth:
  jwt_secret: ${JWT_SECRET}  # Required, generate with openssl
  jwt_expiry: 24h
  refresh_expiry: 168h
  enable_registration: true
  require_email_verification: false
  password_min_length: 8
  
email:
  enabled: true
  smtp_host: smtp.gmail.com
  smtp_port: 587
  smtp_username: ${SMTP_USERNAME}
  smtp_password: ${SMTP_PASSWORD}
  smtp_encryption: starttls  # none, ssl, starttls
  from_address: noreply@example.com
  from_name: CasGists
  
search:
  provider: sqlite  # sqlite, redis, elasticsearch
  redis_url: redis://localhost:6379/0
  elasticsearch_url: http://localhost:9200
  
cache:
  provider: memory  # memory, redis
  redis_url: redis://localhost:6379/1
  ttl: 3600
  max_entries: 10000
  
ratelimit:
  enabled: true
  provider: memory  # memory, redis
  requests_per_minute: 60
  burst: 20
  
features:
  registration: true
  organizations: true
  social: true
  api: true
  webhooks: true
  migration: true
  
security:
  csrf_enabled: true
  cors_enabled: true
  cors_allowed_origins:
    - https://gists.example.com
  cors_allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  cors_allowed_headers:
    - Content-Type
    - Authorization
  content_security_policy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; font-src 'self' https://cdnjs.cloudflare.com; img-src 'self' data: https:;"
```

## Database Setup

### PostgreSQL (Recommended for Production)

1. Install PostgreSQL:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
```

2. Create database and user:
```bash
sudo -u postgres psql << EOF
CREATE DATABASE casgists;
CREATE USER casgists WITH ENCRYPTED PASSWORD 'secure-password';
GRANT ALL PRIVILEGES ON DATABASE casgists TO casgists;
\q
EOF
```

3. Configure PostgreSQL for performance:
```bash
sudo -u postgres psql << EOF
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
SELECT pg_reload_conf();
EOF
```

### MySQL/MariaDB

1. Install MySQL:
```bash
# Ubuntu/Debian
sudo apt install mysql-server

# CentOS/RHEL
sudo yum install mariadb-server
```

2. Create database and user:
```bash
mysql -u root -p << EOF
CREATE DATABASE casgists CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'casgists'@'localhost' IDENTIFIED BY 'secure-password';
GRANT ALL PRIVILEGES ON casgists.* TO 'casgists'@'localhost';
FLUSH PRIVILEGES;
EOF
```

## Reverse Proxy Setup

### Nginx (Recommended)

Create `/etc/nginx/sites-available/casgists`:

```nginx
upstream casgists {
    server localhost:64080 fail_timeout=0;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=casgists_limit:10m rate=10r/s;

# HTTP redirect
server {
    listen 80;
    listen [::]:80;
    server_name gists.example.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name gists.example.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/gists.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/gists.example.com/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    
    # Modern SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # OCSP stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/letsencrypt/live/gists.example.com/chain.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; font-src 'self' https://cdnjs.cloudflare.com; img-src 'self' data: https:;" always;
    
    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/xhtml+xml application/x-font-ttf application/x-font-opentype application/vnd.ms-fontobject image/svg+xml;
    
    # Rate limiting
    location / {
        limit_req zone=casgists_limit burst=20 nodelay;
        
        proxy_pass http://casgists;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Static files with caching
    location /static/ {
        proxy_pass http://casgists;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # Git operations (larger limits)
    location ~ ^/[a-zA-Z0-9_-]+\.(git|pack|idx)$ {
        proxy_pass http://casgists;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Larger limits for git
        client_max_body_size 50m;
        proxy_buffering off;
        proxy_request_buffering off;
    }
    
    # API rate limiting
    location /api/ {
        limit_req zone=casgists_limit burst=10 nodelay;
        
        proxy_pass http://casgists;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable and test:
```bash
sudo ln -s /etc/nginx/sites-available/casgists /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Caddy

Create `/etc/caddy/Caddyfile`:

```caddy
gists.example.com {
    reverse_proxy localhost:64080 {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
        
        # Health checks
        health_uri /health
        health_interval 30s
    }
    
    # Security headers
    header {
        Strict-Transport-Security "max-age=63072000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; font-src 'self' https://cdnjs.cloudflare.com; img-src 'self' data: https:;"
    }
    
    # Compression
    encode gzip
    
    # Git operations
    handle_path /git/* {
        reverse_proxy localhost:64080 {
            flush_interval -1
            transport http {
                read_buffer 8192
            }
        }
    }
    
    # Rate limiting
    rate_limit {
        zone dynamic 10r/s
    }
    
    # Logging
    log {
        output file /var/log/caddy/casgists.log
        format json
    }
}
```

## Security Hardening

### System Security

1. **Create dedicated user**:
```bash
sudo useradd -r -s /bin/false -d /var/lib/casgists -m casgists
```

2. **Set file permissions**:
```bash
sudo chown -R casgists:casgists /var/lib/casgists
sudo chmod 750 /var/lib/casgists
sudo chmod 640 /etc/casgists/config.yaml
```

3. **Configure firewall**:
```bash
# UFW (Ubuntu)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

4. **Enable SELinux** (CentOS/RHEL):
```bash
sudo setsebool -P httpd_can_network_connect 1
sudo semanage port -a -t http_port_t -p tcp 64080
```

### SSL/TLS Setup

1. **Using Let's Encrypt**:
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d gists.example.com --email admin@example.com --agree-tos --no-eff-email

# Auto-renewal
sudo systemctl enable certbot.timer
```

2. **Using custom certificates**:
```bash
# Copy certificates
sudo cp /path/to/cert.pem /etc/ssl/certs/casgists.crt
sudo cp /path/to/key.pem /etc/ssl/private/casgists.key
sudo chmod 644 /etc/ssl/certs/casgists.crt
sudo chmod 600 /etc/ssl/private/casgists.key
```

### Application Security

1. **Generate secure secrets**:
```bash
# JWT secret
openssl rand -hex 32

# Database password
openssl rand -base64 32
```

2. **Configure security settings**:
```yaml
# In config.yaml
security:
  password_min_length: 12
  password_require_uppercase: true
  password_require_lowercase: true
  password_require_numbers: true
  password_require_special: true
  max_login_attempts: 5
  lockout_duration: 15m
  session_timeout: 24h
  enforce_2fa: false
  allowed_email_domains: []
```

3. **Enable audit logging**:
```yaml
audit:
  enabled: true
  log_file: /var/log/casgists/audit.log
  log_login_attempts: true
  log_data_access: true
  log_configuration_changes: true
  retention_days: 90
```

## Monitoring and Logging

### Application Monitoring

1. **Enable metrics endpoint**:
```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
  include_go_metrics: true
```

2. **Prometheus configuration**:
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'casgists'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

3. **Grafana dashboard**:
Import dashboard ID: `15172` for CasGists monitoring

### Logging Configuration

1. **Application logging**:
```yaml
logging:
  level: info  # debug, info, warn, error
  format: json  # text, json
  output: stdout  # stdout, file
  file: /var/log/casgists/app.log
  max_size: 100  # MB
  max_backups: 7
  max_age: 30  # days
  compress: true
```

2. **System logging**:
```bash
# View logs
sudo journalctl -u casgists -f

# Export logs
sudo journalctl -u casgists --since "1 hour ago" -o json > casgists.json

# Log rotation
sudo tee /etc/logrotate.d/casgists << EOF
/var/log/casgists/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 casgists casgists
    postrotate
        systemctl reload casgists
    endscript
}
EOF
```

### Health Monitoring

1. **Health check endpoint**:
```bash
curl -f http://localhost:64080/health || exit 1
```

2. **Monitoring script**:
```bash
#!/bin/bash
# /usr/local/bin/casgists-monitor.sh

HEALTH_URL="http://localhost:64080/health"
WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

if ! curl -sf "$HEALTH_URL" > /dev/null; then
    curl -X POST "$WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d '{"text":"⚠️ CasGists is down!"}'
fi
```

3. **Add to crontab**:
```bash
* * * * * /usr/local/bin/casgists-monitor.sh
```

## Backup and Recovery

### Automated Backup System

1. **Backup script** (`/usr/local/bin/casgists-backup.sh`):
```bash
#!/bin/bash
set -euo pipefail

# Configuration
BACKUP_DIR="/var/backups/casgists"
RETENTION_DAYS=7
S3_BUCKET="s3://your-bucket/casgists-backups"  # Optional

# Create backup
echo "Starting CasGists backup..."
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/casgists_backup_$DATE.tar.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Run CasGists backup
sudo -u casgists casgists backup \
    --output "$BACKUP_FILE" \
    --include-git \
    --include-uploads \
    --compress

# Upload to S3 (optional)
if [ -n "$S3_BUCKET" ]; then
    aws s3 cp "$BACKUP_FILE" "$S3_BUCKET/" || true
fi

# Clean old backups
find "$BACKUP_DIR" -name "casgists_backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $BACKUP_FILE"
```

2. **Cron configuration**:
```bash
# Daily backups at 2 AM
0 2 * * * /usr/local/bin/casgists-backup.sh >> /var/log/casgists/backup.log 2>&1
```

3. **Restore procedure**:
```bash
# Stop service
sudo systemctl stop casgists

# Restore from backup
sudo -u casgists casgists restore \
    --input /var/backups/casgists/casgists_backup_20240115_020000.tar.gz \
    --confirm

# Start service
sudo systemctl start casgists
```

### Database-Specific Backups

1. **PostgreSQL backup**:
```bash
# Backup
pg_dump -h localhost -U casgists -d casgists | gzip > casgists_db_$(date +%Y%m%d).sql.gz

# Restore
gunzip -c casgists_db_20240115.sql.gz | psql -h localhost -U casgists -d casgists
```

2. **MySQL backup**:
```bash
# Backup
mysqldump -u casgists -p casgists | gzip > casgists_db_$(date +%Y%m%d).sql.gz

# Restore
gunzip -c casgists_db_20240115.sql.gz | mysql -u casgists -p casgists
```

## Performance Optimization

### Database Tuning

1. **PostgreSQL optimization**:
```sql
-- Connection pooling
ALTER SYSTEM SET max_connections = 200;

-- Memory settings
ALTER SYSTEM SET shared_buffers = '512MB';
ALTER SYSTEM SET effective_cache_size = '2GB';
ALTER SYSTEM SET work_mem = '8MB';
ALTER SYSTEM SET maintenance_work_mem = '128MB';

-- Performance
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Apply changes
SELECT pg_reload_conf();
```

2. **Indexes**:
```sql
-- Essential indexes for performance
CREATE INDEX idx_gists_user_created ON gists(user_id, created_at DESC);
CREATE INDEX idx_gists_visibility ON gists(visibility) WHERE deleted_at IS NULL;
CREATE INDEX idx_gist_files_gist ON gist_files(gist_id);
CREATE INDEX idx_stars_user_gist ON stars(user_id, gist_id);
CREATE INDEX idx_users_username_lower ON users(LOWER(username));
```

### Redis Caching

1. **Install Redis**:
```bash
sudo apt install redis-server
sudo systemctl enable redis-server
```

2. **Configure CasGists**:
```yaml
cache:
  provider: redis
  redis_url: redis://localhost:6379/1
  ttl: 3600
  
search:
  provider: redis
  redis_url: redis://localhost:6379/0
```

### CDN Integration

1. **CloudFlare configuration**:
- Add CNAME record pointing to your server
- Enable "Proxy" status
- Configure page rules for `/static/*` with "Cache Everything"

2. **Cache headers**:
```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff2?)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    add_header Vary "Accept-Encoding";
}
```

## Troubleshooting

### Common Issues

1. **Service won't start**:
```bash
# Check status
sudo systemctl status casgists

# Check logs
sudo journalctl -u casgists -n 100 --no-pager

# Verify permissions
sudo ls -la /var/lib/casgists

# Test configuration
sudo -u casgists casgists config validate
```

2. **Database connection errors**:
```bash
# Test PostgreSQL connection
psql -h localhost -U casgists -d casgists -c "SELECT 1;"

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log

# Verify pg_hba.conf
sudo grep casgists /etc/postgresql/*/main/pg_hba.conf
```

3. **High memory usage**:
```bash
# Check memory usage
sudo ps aux | grep casgists

# Limit memory in systemd
sudo systemctl edit casgists
# Add:
[Service]
MemoryMax=2G
MemoryHigh=1500M
```

4. **Slow performance**:
```bash
# Enable debug logging
CASGISTS_LOG_LEVEL=debug sudo -u casgists casgists serve

# Check slow queries (PostgreSQL)
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

# Profile the application
curl http://localhost:64080/debug/pprof/profile?seconds=30 > profile.pprof
go tool pprof profile.pprof
```

### Debug Tools

1. **Interactive debugging**:
```bash
# Connect to running instance
sudo -u casgists casgists debug --attach

# Run diagnostics
sudo -u casgists casgists doctor
```

2. **Database analysis**:
```bash
# Analyze database
sudo -u casgists casgists db analyze

# Check migrations
sudo -u casgists casgists db status
```

## Maintenance

### Update Procedure

1. **Backup first**:
```bash
sudo /usr/local/bin/casgists-backup.sh
```

2. **Download and test new version**:
```bash
# Download new version
wget https://github.com/casapps/casgists/releases/latest/download/casgists-linux-amd64
chmod +x casgists-linux-amd64

# Test configuration
./casgists-linux-amd64 config validate

# Dry run migration
./casgists-linux-amd64 db migrate --dry-run
```

3. **Perform update**:
```bash
# Stop service
sudo systemctl stop casgists

# Backup current binary
sudo cp /usr/local/bin/casgists /usr/local/bin/casgists.bak

# Install new version
sudo cp casgists-linux-amd64 /usr/local/bin/casgists
sudo chown root:root /usr/local/bin/casgists
sudo chmod 755 /usr/local/bin/casgists

# Run migrations
sudo -u casgists casgists db migrate

# Start service
sudo systemctl start casgists

# Verify
curl http://localhost:64080/health
```

### Regular Maintenance Tasks

1. **Weekly tasks**:
```bash
# Clean up soft-deleted records (older than 30 days)
sudo -u casgists casgists cleanup --older-than 30d

# Optimize database
sudo -u casgists casgists db optimize

# Update search index
sudo -u casgists casgists search reindex
```

2. **Monthly tasks**:
```bash
# Full system backup
sudo /usr/local/bin/casgists-backup.sh --full

# Security audit
sudo -u casgists casgists security audit

# Performance report
sudo -u casgists casgists report performance --last 30d
```

### Scaling Strategies

1. **Vertical Scaling**:
- Increase CPU cores for better concurrent request handling
- Add RAM for larger cache and database connections
- Use SSD storage for faster I/O

2. **Horizontal Scaling**:
```yaml
# Load balancer configuration
upstream casgists_cluster {
    least_conn;
    server casgists1.internal:64080 max_fails=3 fail_timeout=30s;
    server casgists2.internal:64080 max_fails=3 fail_timeout=30s;
    server casgists3.internal:64080 max_fails=3 fail_timeout=30s;
}
```

3. **Database Clustering**:
- Use PostgreSQL streaming replication
- Configure read replicas for search queries
- Implement connection pooling with PgBouncer

## Support and Resources

- **Documentation**: https://docs.casgists.com
- **Community Forum**: https://github.com/casapps/casgists/discussions
- **Issue Tracker**: https://github.com/casapps/casgists/issues
- **Security**: security@casgists.com
- **Commercial Support**: https://casgists.com/support

## Quick Reference

### Essential Commands

```bash
# Service management
sudo systemctl start|stop|restart|status casgists

# Logs
sudo journalctl -u casgists -f

# Configuration
sudo -u casgists casgists config validate
sudo -u casgists casgists config show

# Database
sudo -u casgists casgists db migrate
sudo -u casgists casgists db status

# Backup/Restore
sudo -u casgists casgists backup --output backup.tar.gz
sudo -u casgists casgists restore --input backup.tar.gz

# Health check
curl http://localhost:64080/health

# Version
casgists version
```

### Configuration Checklist

- [ ] Generated secure JWT secret
- [ ] Configured database connection
- [ ] Set up SSL/TLS certificates
- [ ] Configured email settings
- [ ] Set up backup automation
- [ ] Configured monitoring
- [ ] Enabled firewall
- [ ] Set proper file permissions
- [ ] Configured reverse proxy
- [ ] Tested health endpoint