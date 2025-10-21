# CasGists Deployment Guide

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Production Deployment](#production-deployment)
- [Security Considerations](#security-considerations)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements
- **OS**: Linux, macOS, or Windows
- **Go**: 1.21+ (for building from source)
- **Database**: SQLite (default), PostgreSQL 12+, or MySQL 8+
- **Redis**: 6+ (optional, for caching)
- **Git**: 2.30+ (for repository operations)

### Hardware Requirements
- **Minimum**: 1 CPU, 1GB RAM, 10GB storage
- **Recommended**: 2+ CPUs, 4GB RAM, 50GB+ storage
- **Enterprise**: 4+ CPUs, 8GB+ RAM, 100GB+ SSD storage

## Installation Methods

### Method 1: Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start services
docker-compose up -d

# Access at http://localhost:3000
```

### Method 2: Pre-built Binary

```bash
# Download latest release
curl -L https://github.com/casapps/casgists/releases/latest/download/casgists-linux-amd64.tar.gz | tar xz

# Make executable
chmod +x casgists

# Run setup wizard
./casgists setup

# Start server
./casgists serve
```

### Method 3: Build from Source

```bash
# Clone repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Build
make build

# Run setup wizard
./build/casgists setup

# Start server
./build/casgists serve
```

### Method 4: Systemd Service

```bash
# Copy binary
sudo cp casgists /usr/local/bin/

# Create service user
sudo useradd -r -s /bin/false casgists

# Create directories
sudo mkdir -p /etc/casgists /var/lib/casgists /var/log/casgists
sudo chown casgists:casgists /var/lib/casgists /var/log/casgists

# Create service file
sudo tee /etc/systemd/system/casgists.service > /dev/null <<EOF
[Unit]
Description=CasGists - Self-hosted Gist Service
After=network.target

[Service]
Type=simple
User=casgists
Group=casgists
ExecStart=/usr/local/bin/casgists serve
Restart=always
RestartSec=5
StandardOutput=append:/var/log/casgists/casgists.log
StandardError=append:/var/log/casgists/error.log
Environment="CASGISTS_CONFIG_PATH=/etc/casgists/config.yaml"

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable casgists
sudo systemctl start casgists
```

## Configuration

### Environment Variables

```bash
# Server Configuration
CASGISTS_SERVER_HOST=0.0.0.0
CASGISTS_SERVER_PORT=3000
CASGISTS_SERVER_URL=https://gists.example.com

# Database Configuration
CASGISTS_DATABASE_TYPE=postgres
CASGISTS_DATABASE_DSN=postgres://user:pass@localhost/casgists?sslmode=disable

# Security
CASGISTS_SECURITY_SECRET_KEY=your-secret-key-min-32-chars
CASGISTS_SECURITY_JWT_SECRET=your-jwt-secret
CASGISTS_SECURITY_SESSION_SECRET=your-session-secret

# Features
CASGISTS_FEATURES_REGISTRATION=true
CASGISTS_FEATURES_ORGANIZATIONS=true
CASGISTS_FEATURES_WEBHOOKS=true

# Email
CASGISTS_EMAIL_ENABLED=true
CASGISTS_EMAIL_FROM=noreply@example.com
CASGISTS_EMAIL_SMTP_HOST=smtp.example.com
CASGISTS_EMAIL_SMTP_PORT=587
CASGISTS_EMAIL_SMTP_USER=user
CASGISTS_EMAIL_SMTP_PASS=pass

# Redis Cache
CASGISTS_CACHE_TYPE=redis
CASGISTS_CACHE_REDIS_URL=redis://localhost:6379

# Path Variables
CASGISTS_DATA_DIR=/var/lib/casgists
CASGISTS_LOG_DIR=/var/log/casgists
CASGISTS_BACKUP_DIR={CASGISTS_DATA_DIR}/backups
```

### Configuration File (config.yaml)

```yaml
server:
  host: 0.0.0.0
  port: 3000
  url: https://gists.example.com
  
database:
  type: postgres
  dsn: postgres://user:pass@localhost/casgists?sslmode=disable
  max_open_conns: 25
  max_idle_conns: 5
  
security:
  secret_key: your-secret-key-min-32-chars
  jwt_secret: your-jwt-secret
  session_secret: your-session-secret
  bcrypt_cost: 12
  
features:
  registration: true
  organizations: true
  webhooks: true
  email_verification: true
  two_factor_auth: true
  
rate_limit:
  enabled: true
  requests_per_minute: 60
  
paths:
  data_dir: /var/lib/casgists
  log_dir: /var/log/casgists
  backup_dir: "{CASGISTS_DATA_DIR}/backups"
  repo_dir: "{CASGISTS_DATA_DIR}/repos"
```

## Database Setup

### PostgreSQL

```sql
-- Create database and user
CREATE USER casgists WITH ENCRYPTED PASSWORD 'secure_password';
CREATE DATABASE casgists OWNER casgists;
GRANT ALL PRIVILEGES ON DATABASE casgists TO casgists;

-- Enable extensions
\c casgists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For search
```

### MySQL

```sql
-- Create database and user
CREATE DATABASE casgists CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'casgists'@'%' IDENTIFIED BY 'secure_password';
GRANT ALL PRIVILEGES ON casgists.* TO 'casgists'@'%';
FLUSH PRIVILEGES;
```

### SQLite (Default)

No setup required. Database file will be created automatically at:
- Linux/macOS: `~/.local/share/casgists/casgists.db`
- Windows: `%APPDATA%\casgists\casgists.db`

## Production Deployment

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name gists.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name gists.example.com;
    
    ssl_certificate /etc/ssl/certs/gists.example.com.crt;
    ssl_certificate_key /etc/ssl/private/gists.example.com.key;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # Proxy settings
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Git operations
    location ~ /(.+\.git)/ {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Git specific
        client_max_body_size 50m;
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
```

### Docker Production Setup

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  casgists:
    image: casgists:latest
    container_name: casgists
    restart: always
    ports:
      - "127.0.0.1:3000:3000"
    environment:
      - CASGISTS_DATABASE_TYPE=postgres
      - CASGISTS_DATABASE_DSN=postgres://casgists:${DB_PASSWORD}@postgres:5432/casgists?sslmode=disable
      - CASGISTS_CACHE_TYPE=redis
      - CASGISTS_CACHE_REDIS_URL=redis://redis:6379
    volumes:
      - ./data:/data
      - ./config:/etc/casgists
    depends_on:
      - postgres
      - redis
    
  postgres:
    image: postgres:15-alpine
    container_name: casgists-db
    restart: always
    environment:
      - POSTGRES_USER=casgists
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=casgists
    volumes:
      - postgres_data:/var/lib/postgresql/data
    
  redis:
    image: redis:7-alpine
    container_name: casgists-cache
    restart: always
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment

```yaml
# casgists-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: casgists
spec:
  replicas: 3
  selector:
    matchLabels:
      app: casgists
  template:
    metadata:
      labels:
        app: casgists
    spec:
      containers:
      - name: casgists
        image: casgists:latest
        ports:
        - containerPort: 3000
        env:
        - name: CASGISTS_DATABASE_TYPE
          value: postgres
        - name: CASGISTS_DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: casgists-secrets
              key: database-dsn
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: casgists-data-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: casgists
spec:
  selector:
    app: casgists
  ports:
  - port: 80
    targetPort: 3000
  type: LoadBalancer
```

## Security Considerations

### SSL/TLS Configuration
- Always use HTTPS in production
- Configure strong cipher suites
- Enable HSTS headers
- Use Let's Encrypt for free certificates

### Database Security
- Use strong passwords
- Enable SSL connections
- Regular backups
- Restrict network access

### Application Security
- Change all default secrets
- Enable rate limiting
- Configure CORS properly
- Regular security updates
- Enable 2FA for admin accounts

### Firewall Rules
```bash
# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow SSH (restricted)
sudo ufw allow from 10.0.0.0/8 to any port 22

# Database (internal only)
# Don't expose database ports publicly
```

## Monitoring & Maintenance

### Health Checks
```bash
# Check service health
curl http://localhost:3000/api/v1/health

# Check metrics
curl http://localhost:3000/metrics
```

### Prometheus Configuration
```yaml
scrape_configs:
  - job_name: 'casgists'
    static_configs:
      - targets: ['localhost:3000']
    metrics_path: '/metrics'
```

### Backup Strategy
```bash
# Automated daily backups
0 2 * * * /usr/local/bin/casgists backup --output /backups/casgists-$(date +\%Y\%m\%d).tar.gz

# Backup retention (30 days)
0 3 * * * find /backups -name "casgists-*.tar.gz" -mtime +30 -delete
```

### Log Rotation
```bash
# /etc/logrotate.d/casgists
/var/log/casgists/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 644 casgists casgists
    postrotate
        systemctl reload casgists > /dev/null 2>&1 || true
    endscript
}
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check logs
journalctl -u casgists -f

# Verify configuration
./casgists validate-config

# Check permissions
ls -la /var/lib/casgists
```

#### Database Connection Failed
```bash
# Test connection
psql -h localhost -U casgists -d casgists

# Check firewall
sudo ufw status

# Verify credentials
echo $CASGISTS_DATABASE_DSN
```

#### Performance Issues
```bash
# Check resource usage
htop

# Database connections
SELECT count(*) FROM pg_stat_activity;

# Clear cache
redis-cli FLUSHALL
```

### Debug Mode
```bash
# Enable debug logging
export CASGISTS_LOG_LEVEL=debug
./casgists serve
```

### Support Resources
- GitHub Issues: https://github.com/casapps/casgists/issues
- Documentation: https://docs.casgists.com
- Community Forum: https://forum.casgists.com

## Migration from Other Platforms

### From OpenGist
```bash
./casgists migrate opengist \
  --source-db "postgres://user:pass@old-host/opengist" \
  --map-users
```

### From GitHub
```bash
./casgists import github \
  --token YOUR_GITHUB_TOKEN \
  --include-private \
  --preserve-dates
```

## Maintenance Commands

```bash
# Database migrations
./casgists migrate up

# Clean old sessions
./casgists cleanup sessions --older-than 30d

# Optimize database
./casgists db optimize

# Export metrics
./casgists metrics export --format prometheus
```