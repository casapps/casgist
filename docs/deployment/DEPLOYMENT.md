# CasGists Deployment Guide

## Table of Contents
- [Requirements](#requirements)
- [Quick Start](#quick-start)
- [Docker Deployment](#docker-deployment)
- [Manual Deployment](#manual-deployment)
- [Configuration](#configuration)
- [SSL/TLS Setup](#ssltls-setup)
- [Backup & Restore](#backup--restore)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Requirements

### Minimum Hardware Requirements
- CPU: 1 core
- RAM: 512MB
- Storage: 10GB

### Recommended Hardware Requirements
- CPU: 2+ cores
- RAM: 2GB+
- Storage: 50GB+ SSD

### Software Requirements
- Go 1.23+ (for building from source)
- Docker & Docker Compose (for containerized deployment)
- PostgreSQL 14+ (optional, for production database)
- Redis 6+ (optional, for caching)
- Nginx (optional, for reverse proxy)

## Quick Start

### Using Docker (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/casapps/casgists.git
cd casgists
```

2. Copy the example environment file:
```bash
cp deployments/docker/.env.example deployments/docker/.env
```

3. Generate a secure secret key:
```bash
openssl rand -base64 32
```

4. Edit `.env` and set `CASGISTS_SECRET_KEY` with the generated key.

5. Start CasGists:
```bash
cd deployments/docker
docker-compose up -d
```

6. Access CasGists at http://localhost:64001

## Docker Deployment

### Basic Deployment

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f casgists

# Stop services
docker-compose down
```

### Production Deployment with PostgreSQL

```bash
# Start with PostgreSQL
docker-compose --profile postgres up -d

# Configure PostgreSQL connection in .env
CASGISTS_DATABASE_DRIVER=postgres
CASGISTS_DATABASE_DSN=postgres://casgists:password@postgres/casgists?sslmode=disable
```

### Production Deployment with Redis Cache

```bash
# Start with Redis
docker-compose --profile redis up -d

# Configure Redis in .env
CASGISTS_CACHE_DRIVER=redis
REDIS_HOST=redis
REDIS_PORT=6379
```

### Full Production Stack

```bash
# Start all services
docker-compose --profile postgres --profile redis --profile nginx up -d
```

## Manual Deployment

### Building from Source

1. Install Go 1.23+:
```bash
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

2. Build CasGists:
```bash
git clone https://github.com/casapps/casgists.git
cd casgists
make build
```

3. Create system user:
```bash
sudo useradd -r -s /bin/false casgists
```

4. Install binary and files:
```bash
sudo mkdir -p /opt/casgists/{bin,data,logs,storage,git}
sudo cp build/casgists /opt/casgists/bin/
sudo cp -r web /opt/casgists/
sudo cp deployments/docker/config.yaml /opt/casgists/
sudo chown -R casgists:casgists /opt/casgists
```

5. Install systemd service:
```bash
sudo cp deployments/systemd/casgists.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable casgists
```

6. Configure environment:
```bash
sudo mkdir -p /etc/casgists
echo "CASGISTS_SECRET_KEY=$(openssl rand -base64 32)" | sudo tee /etc/casgists/environment
sudo chmod 600 /etc/casgists/environment
```

7. Start service:
```bash
sudo systemctl start casgists
sudo systemctl status casgists
```

## Configuration

### Essential Configuration

```yaml
# /opt/casgists/config.yaml

server:
  host: 0.0.0.0
  port: 64001
  url: https://your-domain.com  # Set your actual domain

security:
  secret_key: ${CASGISTS_SECRET_KEY}  # Must be set!

database:
  # For PostgreSQL (recommended for production)
  driver: postgres
  dsn: postgres://user:pass@localhost/casgists?sslmode=require
  
  # For SQLite (simple deployments)
  driver: sqlite
  dsn: /opt/casgists/data/casgists.db

features:
  registration: true  # Allow new user registration
  public_gists: true  # Allow public gists
```

### Environment Variables

All configuration values can be overridden using environment variables:

```bash
CASGISTS_SERVER_PORT=64001
CASGISTS_SERVER_URL=https://gists.example.com
CASGISTS_DATABASE_DSN=postgres://user:pass@db/casgists
CASGISTS_SECURITY_SECRET_KEY=your-secret-key
```

## SSL/TLS Setup

### Using Let's Encrypt with Nginx

1. Install Certbot:
```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx
```

2. Configure Nginx:
```bash
sudo cp deployments/docker/nginx.conf /etc/nginx/sites-available/casgists
sudo ln -s /etc/nginx/sites-available/casgists /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

3. Obtain certificate:
```bash
sudo certbot --nginx -d your-domain.com
```

### Using Self-Signed Certificate

```bash
# Generate certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Configure in nginx.conf
ssl_certificate /path/to/cert.pem;
ssl_certificate_key /path/to/key.pem;
```

## Backup & Restore

### Automated Backups

Create `/opt/casgists/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/backup/casgists"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
if [ "$DB_TYPE" = "postgres" ]; then
    pg_dump $DATABASE_URL > $BACKUP_DIR/db_$DATE.sql
else
    cp /opt/casgists/data/casgists.db $BACKUP_DIR/db_$DATE.sqlite
fi

# Backup files
tar -czf $BACKUP_DIR/files_$DATE.tar.gz \
    /opt/casgists/storage \
    /opt/casgists/git

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql" -o -name "*.sqlite" -o -name "*.tar.gz" -mtime +7 -delete
```

Add to crontab:
```bash
0 2 * * * /opt/casgists/backup.sh
```

### Manual Backup

```bash
# Backup database (SQLite)
sqlite3 /opt/casgists/data/casgists.db ".backup /backup/casgists.db"

# Backup database (PostgreSQL)
pg_dump -h localhost -U casgists casgists > backup.sql

# Backup files
tar -czf casgists-files.tar.gz /opt/casgists/storage /opt/casgists/git
```

### Restore

```bash
# Restore database (SQLite)
cp /backup/casgists.db /opt/casgists/data/casgists.db

# Restore database (PostgreSQL)
psql -h localhost -U casgists casgists < backup.sql

# Restore files
tar -xzf casgists-files.tar.gz -C /
```

## Monitoring

### Health Checks

```bash
# Check service health
curl http://localhost:64001/healthz

# Detailed health check
curl http://localhost:64001/healthz?detailed=true
```

### Prometheus Metrics

Enable in configuration:
```yaml
monitoring:
  prometheus:
    enabled: true
    path: /metrics
```

### Log Monitoring

```bash
# View logs
journalctl -u casgists -f

# Docker logs
docker-compose logs -f casgists

# Application logs
tail -f /opt/casgists/logs/app.log
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check logs
sudo journalctl -u casgists -n 50

# Check permissions
ls -la /opt/casgists/

# Validate configuration
/opt/casgists/bin/casgists validate
```

#### Database Connection Failed
```bash
# Test PostgreSQL connection
psql -h localhost -U casgists -d casgists

# Check SQLite permissions
ls -la /opt/casgists/data/casgists.db
```

#### Port Already in Use
```bash
# Find process using port
sudo lsof -i :64001

# Change port in config
server:
  port: 64002
```

### Performance Tuning

#### Database Optimization
```sql
-- PostgreSQL
VACUUM ANALYZE;
REINDEX DATABASE casgists;

-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_gists_user_created ON gists(user_id, created_at DESC);
```

#### Resource Limits
```yaml
# Increase connection pool
database:
  max_open_connections: 50
  max_idle_connections: 10

# Increase memory cache
performance:
  max_memory_mb: 2048
```

### Security Checklist

- [ ] Set strong `CASGISTS_SECRET_KEY`
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Set up regular backups
- [ ] Enable rate limiting
- [ ] Configure fail2ban
- [ ] Review user permissions
- [ ] Enable audit logging
- [ ] Keep software updated

### Getting Help

1. Check logs for error messages
2. Review configuration
3. Search [GitHub Issues](https://github.com/casapps/casgists/issues)
4. Join our [Discord community](https://discord.gg/casgists)
5. Contact support at support@casgists.com