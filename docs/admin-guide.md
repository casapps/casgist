# CasGists Administrator Guide

This comprehensive guide covers administration, configuration, and management of CasGists instances.

## Table of Contents

1. [Initial Setup](#initial-setup)
2. [Configuration Management](#configuration-management)
3. [User Management](#user-management)
4. [System Monitoring](#system-monitoring)
5. [Backup and Recovery](#backup-and-recovery)
6. [Security Management](#security-management)
7. [Performance Tuning](#performance-tuning)
8. [Troubleshooting](#troubleshooting)
9. [Maintenance Tasks](#maintenance-tasks)
10. [Advanced Administration](#advanced-administration)

## Initial Setup

### First Run Setup Wizard

When CasGists starts for the first time, it launches an 8-step setup wizard:

1. **Welcome & Requirements Check**
   - System requirements validation
   - Directory permissions check
   - Database connectivity test

2. **Database Configuration**
   ```yaml
   database:
     type: postgresql  # sqlite, postgresql, mysql
     host: localhost
     port: 5432
     name: casgists
     user: casgists
     password: secure-password
   ```

3. **Admin Account Creation**
   - Username (3-20 characters)
   - Email address
   - Strong password (min 8 chars)
   - This account gets full admin privileges

4. **Basic Settings**
   - Site name and URL
   - Default language
   - Time zone
   - Date/time formats

5. **Security Configuration**
   ```yaml
   security:
     jwt_secret: # Auto-generated if not provided
     enable_registration: true
     require_email_verification: false
     password_min_length: 8
     max_login_attempts: 5
     lockout_duration: 15m
   ```

6. **Email Settings** (Optional)
   ```yaml
   email:
     enabled: true
     smtp_host: smtp.gmail.com
     smtp_port: 587
     smtp_username: your-email@gmail.com
     smtp_password: app-password
     smtp_encryption: starttls
   ```

7. **Feature Toggles**
   - Public registration
   - Organizations
   - Social features (stars, follows)
   - API access
   - Webhooks

8. **Final Review**
   - Configuration summary
   - Test email sending
   - Apply configuration

### Post-Installation Tasks

1. **SSL/TLS Setup**
   ```bash
   # Using Let's Encrypt
   sudo certbot --nginx -d gists.yourdomain.com
   ```

2. **Backup Configuration**
   ```bash
   # Set up automated backups
   sudo crontab -e
   0 2 * * * /usr/local/bin/casgists-backup.sh
   ```

3. **Monitoring Setup**
   - Configure health checks
   - Set up log aggregation
   - Enable metrics collection

## Configuration Management

### Configuration Files

CasGists uses a hierarchical configuration system:

1. **Default values** (built-in)
2. **Configuration file** (`/etc/casgists/config.yaml`)
3. **Environment variables** (highest priority)

### Key Configuration Sections

#### Server Configuration
```yaml
server:
  port: 64080
  host: 0.0.0.0
  url: https://gists.example.com
  enable_https: false
  tls_cert: /path/to/cert.pem
  tls_key: /path/to/key.pem
  
  # Request handling
  read_timeout: 30s
  write_timeout: 30s
  max_request_size: 52428800  # 50MB
  
  # CORS settings
  cors_enabled: true
  cors_origins:
    - https://example.com
    - https://app.example.com
```

#### Storage Configuration
```yaml
storage:
  # Main data directory
  data_dir: /var/lib/casgists
  
  # Git repositories
  git_root: ${DATA_DIR}/repos
  
  # File uploads
  upload_dir: ${DATA_DIR}/uploads
  max_file_size: 10485760  # 10MB
  
  # Allowed file extensions (empty = all)
  allowed_extensions: []
  
  # S3-compatible storage (optional)
  s3:
    enabled: false
    endpoint: s3.amazonaws.com
    bucket: casgists-files
    region: us-east-1
    access_key: ${S3_ACCESS_KEY}
    secret_key: ${S3_SECRET_KEY}
```

#### Feature Flags
```yaml
features:
  # Core features
  registration: true
  email_verification: false
  password_reset: true
  
  # Social features  
  organizations: true
  teams: true
  stars: true
  follows: true
  comments: true
  
  # Advanced features
  webhooks: true
  api: true
  graphql: false
  migration: true
  export: true
  
  # Limits
  max_gists_per_user: 1000
  max_files_per_gist: 10
  max_organizations_per_user: 10
```

### Dynamic Configuration

Some settings can be changed at runtime via the admin API:

```bash
# Update configuration
curl -X PUT https://gists.example.com/api/v1/admin/config \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "features.registration": false,
    "security.max_login_attempts": 3
  }'
```

### Environment Variables

All configuration values can be overridden with environment variables:

```bash
# Format: CASGISTS_<SECTION>_<KEY>
CASGISTS_SERVER_PORT=8080
CASGISTS_DATABASE_HOST=db.example.com
CASGISTS_FEATURES_REGISTRATION=false
CASGISTS_SECURITY_JWT_SECRET=your-secret-key
```

## User Management

### Admin Dashboard

Access the admin dashboard at `/admin` (requires admin privileges).

Dashboard sections:
- **Overview**: System stats, recent activity
- **Users**: User management interface
- **Gists**: Content moderation
- **Organizations**: Org management
- **System**: Configuration, logs, backups

### User Administration

#### Viewing Users
```bash
# CLI command
casgists admin users list --page 1 --per-page 50

# API
GET /api/v1/admin/users?page=1&per_page=50
```

#### User Actions

1. **Suspend User**
   ```bash
   casgists admin users suspend <username> --reason "TOS violation" --duration 7d
   ```

2. **Delete User**
   ```bash
   casgists admin users delete <username> --permanent
   ```

3. **Reset Password**
   ```bash
   casgists admin users reset-password <username>
   ```

4. **Grant Admin**
   ```bash
   casgists admin users promote <username>
   ```

5. **Modify Quotas**
   ```bash
   casgists admin users quota <username> --gists 500 --storage 1GB
   ```

### Batch Operations

```bash
# Export user list
casgists admin users export --format csv --output users.csv

# Bulk suspend inactive users
casgists admin users suspend-inactive --days 365 --dry-run

# Send announcement
casgists admin users notify --subject "Maintenance Notice" --template maintenance.html
```

### Organization Management

```bash
# List all organizations
casgists admin orgs list

# Transfer ownership
casgists admin orgs transfer <org-name> --to <new-owner>

# Set organization limits
casgists admin orgs limits <org-name> --members 50 --gists 1000
```

## System Monitoring

### Health Checks

1. **Basic Health Check**
   ```bash
   curl http://localhost:64080/health
   ```
   
   Response:
   ```json
   {
     "status": "healthy",
     "version": "1.0.0",
     "uptime": "72h 15m 30s",
     "database": "ok",
     "storage": "ok",
     "search": "ok"
   }
   ```

2. **Detailed Health Check**
   ```bash
   casgists health --detailed
   ```

### Metrics and Monitoring

#### Prometheus Metrics
```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
  include_go_metrics: true
  include_process_metrics: true
```

Key metrics:
- `casgists_http_requests_total` - Total HTTP requests
- `casgists_http_request_duration_seconds` - Request latency
- `casgists_active_users` - Currently active users
- `casgists_gist_operations_total` - Gist CRUD operations
- `casgists_storage_bytes` - Storage usage
- `casgists_database_connections` - DB connection pool

#### Grafana Dashboard

Import dashboard ID: `15172` or use the provided JSON:

```json
{
  "dashboard": {
    "title": "CasGists Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [{
          "expr": "rate(casgists_http_requests_total[5m])"
        }]
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(casgists_http_requests_total{status=~\"5..\"}[5m])"
        }]
      }
    ]
  }
}
```

### Log Management

#### Log Configuration
```yaml
logging:
  level: info  # debug, info, warn, error
  format: json  # text, json
  outputs:
    - type: file
      path: /var/log/casgists/app.log
      max_size: 100  # MB
      max_backups: 7
      compress: true
    - type: syslog
      network: udp
      address: localhost:514
      tag: casgists
```

#### Log Analysis

Common log queries:

```bash
# Failed login attempts
grep "login_failed" /var/log/casgists/app.log | jq '.user'

# Slow queries
grep "slow_query" /var/log/casgists/app.log | jq 'select(.duration > 1000)'

# Error summary
grep "error" /var/log/casgists/app.log | jq -r '.error' | sort | uniq -c

# User activity
grep "gist_created" /var/log/casgists/app.log | jq -r '.user' | sort | uniq -c
```

### Performance Monitoring

```bash
# Real-time performance stats
casgists admin perf --watch

# Generate performance report
casgists admin report performance --period 7d --output perf-report.html

# Database query analysis
casgists admin db analyze --slow-queries --limit 20
```

## Backup and Recovery

### Backup Strategy

#### Full System Backup

```bash
#!/bin/bash
# /usr/local/bin/casgists-backup.sh

BACKUP_DIR="/var/backups/casgists"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/casgists_full_$DATE.tar.gz"

# Create backup
casgists backup \
  --output "$BACKUP_FILE" \
  --include-config \
  --include-database \
  --include-git \
  --include-uploads \
  --compress \
  --encrypt \
  --progress

# Verify backup
casgists backup verify --input "$BACKUP_FILE"

# Upload to S3
aws s3 cp "$BACKUP_FILE" s3://backups/casgists/

# Clean old backups
find "$BACKUP_DIR" -mtime +7 -name "*.tar.gz" -delete
```

#### Incremental Backups

```bash
# Daily incremental
casgists backup \
  --incremental \
  --base /var/backups/casgists/weekly_full.tar.gz \
  --output /var/backups/casgists/daily_$(date +%Y%m%d).tar.gz
```

### Recovery Procedures

#### Full Recovery

1. **Stop the service**
   ```bash
   sudo systemctl stop casgists
   ```

2. **Restore from backup**
   ```bash
   casgists restore \
     --input /var/backups/casgists/casgists_full_20240115.tar.gz \
     --verify \
     --progress
   ```

3. **Verify data integrity**
   ```bash
   casgists admin verify --database --files --git-repos
   ```

4. **Start service**
   ```bash
   sudo systemctl start casgists
   ```

#### Selective Recovery

```bash
# Restore specific user's data
casgists restore \
  --input backup.tar.gz \
  --filter user:john \
  --target /tmp/john-restore

# Restore specific gists
casgists restore \
  --input backup.tar.gz \
  --filter gist:id1,id2,id3 \
  --merge
```

### Disaster Recovery Plan

1. **RPO (Recovery Point Objective)**: 1 hour
   - Hourly incremental backups
   - Real-time replication for critical data

2. **RTO (Recovery Time Objective)**: 2 hours
   - Automated recovery scripts
   - Pre-tested recovery procedures
   - Hot standby server

3. **Recovery Steps**:
   ```bash
   # 1. Assess damage
   casgists admin diagnose --full
   
   # 2. Switch to DR site (if available)
   casgists admin failover --to dr-site
   
   # 3. Restore from backup
   ./disaster-recovery.sh --latest-backup
   
   # 4. Verify functionality
   casgists admin verify --all
   
   # 5. Notify users
   casgists admin notify --template dr-complete.html
   ```

## Security Management

### Security Configuration

```yaml
security:
  # Authentication
  jwt_secret: ${JWT_SECRET}  # Min 32 chars
  jwt_algorithm: HS256
  jwt_expiry: 24h
  refresh_expiry: 168h
  
  # Password policy
  password_min_length: 12
  password_require_uppercase: true
  password_require_lowercase: true
  password_require_numbers: true
  password_require_special: true
  password_history: 5
  password_expiry_days: 90
  
  # Account security
  max_login_attempts: 5
  lockout_duration: 30m
  require_2fa_admin: true
  session_timeout: 2h
  concurrent_sessions: 3
  
  # API security
  api_rate_limit: 5000/hour
  api_key_expiry: 365d
  
  # Content security
  max_file_size: 10485760
  allowed_file_types: []
  scan_uploads: true
  
  # Network security
  trusted_proxies:
    - 10.0.0.0/8
    - 172.16.0.0/12
  ip_whitelist: []
  ip_blacklist: []
```

### Access Control

#### Role-Based Access Control (RBAC)

```yaml
roles:
  admin:
    permissions:
      - "*"
  
  moderator:
    permissions:
      - "gist:read"
      - "gist:delete"
      - "user:read"
      - "user:suspend"
  
  user:
    permissions:
      - "gist:read"
      - "gist:write:own"
      - "user:read:own"
      - "user:write:own"
```

#### API Token Management

```bash
# List all API tokens
casgists admin tokens list --include-expired

# Revoke compromised tokens
casgists admin tokens revoke --pattern "suspicious-*"

# Audit token usage
casgists admin tokens audit --user <username> --period 30d
```

### Security Auditing

#### Audit Log Configuration

```yaml
audit:
  enabled: true
  log_file: /var/log/casgists/audit.log
  
  # What to audit
  events:
    - login_attempt
    - login_success
    - login_failed
    - logout
    - password_change
    - permission_change
    - gist_access
    - api_access
    - admin_action
    - config_change
    
  # Retention
  retention_days: 365
  archive_to_s3: true
```

#### Security Reports

```bash
# Generate security audit report
casgists admin report security --period 30d --output security-audit.pdf

# Failed login analysis
casgists admin report failed-logins --group-by ip,user --limit 100

# Suspicious activity detection
casgists admin detect suspicious --sensitivity high
```

### Compliance

#### GDPR Compliance

```bash
# Export user data
casgists admin gdpr export --user <email> --format json

# Delete user data (right to be forgotten)
casgists admin gdpr delete --user <email> --confirm

# Consent audit
casgists admin gdpr consent-audit --output consent-report.csv
```

#### Data Retention

```yaml
retention:
  # Soft-deleted items
  gists_deleted: 30d
  users_deleted: 90d
  
  # Logs
  access_logs: 90d
  audit_logs: 365d
  error_logs: 30d
  
  # Backups
  daily_backups: 7d
  weekly_backups: 30d
  monthly_backups: 365d
```

## Performance Tuning

### Database Optimization

#### PostgreSQL Tuning

```sql
-- Connection pooling
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '1GB';
ALTER SYSTEM SET effective_cache_size = '3GB';

-- Performance
ALTER SYSTEM SET work_mem = '16MB';
ALTER SYSTEM SET maintenance_work_mem = '256MB';
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;

-- Checkpoints
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET min_wal_size = '1GB';
ALTER SYSTEM SET max_wal_size = '4GB';

-- Apply changes
SELECT pg_reload_conf();
```

#### Query Optimization

```bash
# Analyze slow queries
casgists admin db slow-queries --duration 1s --limit 20

# Update statistics
casgists admin db analyze --tables all

# Suggest missing indexes
casgists admin db suggest-indexes --workload 7d
```

### Caching Strategy

```yaml
cache:
  # Provider: memory, redis, memcached
  provider: redis
  
  # Redis configuration
  redis:
    url: redis://localhost:6379/0
    max_connections: 100
    timeout: 5s
  
  # Cache settings
  default_ttl: 3600
  
  # Specific TTLs
  ttls:
    user_profile: 300
    gist_content: 3600
    search_results: 600
    api_responses: 300
  
  # Cache warming
  warm_on_start: true
  warm_popular: true
  warm_threshold: 100  # views
```

### Search Optimization

```yaml
search:
  # Provider: sqlite, postgresql, elasticsearch, meilisearch
  provider: elasticsearch
  
  elasticsearch:
    url: http://localhost:9200
    index_prefix: casgists
    shards: 3
    replicas: 1
  
  # Indexing
  batch_size: 100
  index_on_create: true
  index_on_update: true
  
  # Performance
  max_results: 1000
  timeout: 10s
  cache_results: true
```

### CDN Integration

```nginx
# Nginx caching for static assets
location ~* \.(js|css|png|jpg|jpeg|gif|ico|woff|woff2)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    add_header Vary "Accept-Encoding";
}

# Cloudflare page rules
# /static/* - Cache Everything, TTL 1 month
# /api/* - Bypass Cache
# /* - Standard Caching
```

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check for port conflicts
sudo lsof -i :64080

# Verify permissions
ls -la /var/lib/casgists
ls -la /etc/casgists/config.yaml

# Test database connection
casgists admin db test

# Validate configuration
casgists config validate

# Check system resources
free -h
df -h
```

#### Database Issues

```bash
# Connection pool exhausted
casgists admin db connections --list

# Lock contention
casgists admin db locks --active

# Corrupted data
casgists admin db check --repair

# Migration failures
casgists admin db migrate --status
casgists admin db migrate --rollback
```

#### Performance Issues

```bash
# CPU profiling
casgists admin profile cpu --duration 30s --output cpu.pprof

# Memory profiling
casgists admin profile memory --output mem.pprof

# Goroutine analysis
casgists admin profile goroutines

# Trace requests
casgists admin trace --duration 60s --filter "duration>1s"
```

### Debug Mode

```bash
# Enable debug logging
CASGISTS_LOG_LEVEL=debug casgists serve

# Enable SQL query logging
CASGISTS_DATABASE_LOG_QUERIES=true casgists serve

# Enable profiling endpoint
CASGISTS_DEBUG_PROFILE=true casgists serve
```

### Recovery Tools

```bash
# Database repair
casgists admin db repair --table gists --check-constraints

# Reindex search
casgists admin search reindex --force --parallel 4

# Clear caches
casgists admin cache clear --all

# Reset user password
casgists admin users reset-password <username> --send-email
```

## Maintenance Tasks

### Daily Tasks

```bash
#!/bin/bash
# Daily maintenance script

echo "Starting daily maintenance..."

# Backup
/usr/local/bin/casgists-backup.sh

# Clean old sessions
casgists admin sessions clean --older-than 7d

# Update search index
casgists admin search update --incremental

# Generate daily report
casgists admin report daily --email admin@example.com

echo "Daily maintenance completed"
```

### Weekly Tasks

```bash
#!/bin/bash
# Weekly maintenance script

# Database maintenance
casgists admin db vacuum --analyze
casgists admin db reindex --concurrent

# Security audit
casgists admin security audit --full

# Performance analysis
casgists admin performance analyze --period 7d

# Update GeoIP database
casgists admin geoip update
```

### Monthly Tasks

```bash
#!/bin/bash
# Monthly maintenance script

# Full backup
casgists backup --full --verify

# Clean soft-deleted data
casgists admin cleanup --older-than 30d

# Certificate renewal check
casgists admin certs check --warn-days 30

# License audit
casgists admin licenses audit

# Capacity planning
casgists admin capacity report --projection 6m
```

## Advanced Administration

### Multi-Node Setup

```yaml
cluster:
  enabled: true
  node_id: node1
  
  # Cluster communication
  gossip_port: 7946
  gossip_key: ${CLUSTER_KEY}
  
  # Service discovery
  discovery: consul
  consul_address: consul.service.consul:8500
  
  # Load balancing
  load_balancer: haproxy
  health_check_interval: 10s
```

### High Availability

```yaml
ha:
  # Database
  database:
    type: postgresql
    primary: db-primary.example.com
    replicas:
      - db-replica1.example.com
      - db-replica2.example.com
    failover_timeout: 30s
  
  # Storage
  storage:
    type: s3
    replicate: true
    consistency: eventual
  
  # Cache
  cache:
    type: redis
    mode: sentinel
    master: mymaster
    sentinels:
      - sentinel1:26379
      - sentinel2:26379
      - sentinel3:26379
```

### Custom Integrations

#### Webhook Integration

```yaml
webhooks:
  # Slack notifications
  slack:
    url: https://hooks.slack.com/services/xxx/yyy/zzz
    events:
      - gist.created
      - user.registered
      - system.alert
  
  # Custom webhook
  custom:
    url: https://api.example.com/casgists-webhook
    secret: ${WEBHOOK_SECRET}
    retry_attempts: 3
    retry_delay: 5s
```

#### LDAP/AD Integration

```yaml
ldap:
  enabled: true
  url: ldap://ldap.example.com:389
  bind_dn: cn=admin,dc=example,dc=com
  bind_password: ${LDAP_PASSWORD}
  
  # User search
  user_base: ou=users,dc=example,dc=com
  user_filter: (uid={{username}})
  
  # Group mapping
  group_base: ou=groups,dc=example,dc=com
  admin_group: cn=casgists-admins,ou=groups,dc=example,dc=com
  
  # Attribute mapping
  attributes:
    username: uid
    email: mail
    display_name: displayName
```

### API Automation

```python
#!/usr/bin/env python3
# Admin automation script

import requests
import json

class CasGistsAdmin:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {'Authorization': f'Bearer {token}'}
    
    def daily_report(self):
        """Generate daily admin report"""
        stats = self.get_stats()
        users = self.get_active_users(period='24h')
        gists = self.get_popular_gists(period='24h')
        
        return {
            'date': datetime.now().isoformat(),
            'stats': stats,
            'active_users': users,
            'popular_gists': gists
        }
    
    def cleanup_inactive_users(self, days=365):
        """Clean up inactive users"""
        users = self.get_inactive_users(days=days)
        for user in users:
            print(f"Suspending user: {user['username']}")
            self.suspend_user(user['id'], reason='Inactivity')
```

### Monitoring Integration

```yaml
monitoring:
  # Datadog
  datadog:
    enabled: true
    api_key: ${DATADOG_API_KEY}
    tags:
      - env:production
      - service:casgists
  
  # New Relic
  newrelic:
    enabled: true
    license_key: ${NEWRELIC_LICENSE}
    app_name: CasGists Production
  
  # Custom metrics
  statsd:
    enabled: true
    host: localhost
    port: 8125
    prefix: casgists
```

## Best Practices

### Security Best Practices

1. **Regular Updates**
   - Check for updates weekly
   - Test updates in staging first
   - Keep dependencies updated

2. **Access Control**
   - Use principle of least privilege
   - Regular permission audits
   - Strong password policy

3. **Monitoring**
   - Set up alerts for critical events
   - Monitor resource usage
   - Track error rates

### Operational Best Practices

1. **Documentation**
   - Document all customizations
   - Maintain runbooks
   - Keep contact lists updated

2. **Change Management**
   - Test all changes in staging
   - Have rollback plans
   - Communicate maintenance windows

3. **Capacity Planning**
   - Monitor growth trends
   - Plan for peak usage
   - Scale before hitting limits

### Backup Best Practices

1. **3-2-1 Rule**
   - 3 copies of data
   - 2 different storage types
   - 1 offsite backup

2. **Regular Testing**
   - Test restores monthly
   - Document recovery time
   - Verify backup integrity

3. **Automation**
   - Automate backup processes
   - Automate verification
   - Automate cleanup

## Support Resources

- **Admin Forum**: https://forum.casgists.com/c/admin
- **Slack Channel**: casgists-admin.slack.com
- **Documentation**: https://docs.casgists.com/admin
- **Training**: https://training.casgists.com
- **Enterprise Support**: enterprise@casgists.com

## Conclusion

Effective administration of CasGists requires a combination of proactive monitoring, regular maintenance, and strategic planning. This guide provides the foundation for managing a successful CasGists deployment. Remember to stay updated with new releases and security advisories.