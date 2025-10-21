# Deployment Guide

This guide covers production deployment scenarios for CasGists, including reverse proxy configuration, SSL/TLS setup, and best practices for different environments.

## Production Deployment Architecture

```
Internet → Load Balancer → Reverse Proxy (Nginx/Caddy) → CasGists → Database
                                ↓
                         Static Assets (CDN)
                                ↓
                         Storage (S3/Local)
```

## Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- Domain name with DNS configured
- SSL certificate (Let's Encrypt recommended)
- Reverse proxy (Nginx, Caddy, or Traefik)
- Database (PostgreSQL or MySQL for production)

## Reverse Proxy Configuration

### Nginx

Install and configure Nginx as a reverse proxy:

```bash
# Install Nginx
sudo apt update
sudo apt install nginx

# Create configuration file
sudo nano /etc/nginx/sites-available/casgists
```

Basic Nginx configuration:

```nginx
server {
    listen 80;
    server_name gists.yourdomain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name gists.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/gists.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/gists.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";

    # Client max body size (for file uploads)
    client_max_body_size 50M;

    # Proxy settings
    location / {
        proxy_pass http://127.0.0.1:64080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Static assets caching (optional)
    location /static/ {
        proxy_pass http://127.0.0.1:64080;
        proxy_cache_valid 200 1d;
        expires 1d;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint
    location /health {
        proxy_pass http://127.0.0.1:64080;
        access_log off;
    }

    # Logging
    access_log /var/log/nginx/casgists.access.log;
    error_log /var/log/nginx/casgists.error.log;
}
```

Enable the site:

```bash
# Enable the site
sudo ln -s /etc/nginx/sites-available/casgists /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Caddy

Caddy provides automatic HTTPS with Let's Encrypt:

```bash
# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

Create Caddyfile:

```bash
sudo nano /etc/caddy/Caddyfile
```

```caddy
gists.yourdomain.com {
    # Automatic HTTPS with Let's Encrypt
    
    # Security headers
    header {
        Strict-Transport-Security "max-age=63072000"
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
    }

    # Reverse proxy to CasGists
    reverse_proxy 127.0.0.1:64080 {
        # Forward real IP
        header_up X-Real-IP {remote}
        header_up X-Forwarded-Proto {scheme}
        header_up X-Forwarded-Host {host}
        
        # Health check
        health_uri /health
        health_interval 10s
    }

    # Logging
    log {
        output file /var/log/caddy/casgists.log
    }

    # Rate limiting (requires caddy-ratelimit plugin)
    # rate_limit {
    #     zone dynamic_rl {
    #         key {remote}
    #         events 100
    #         window 1m
    #     }
    # }
}

# Redirect www to non-www
www.gists.yourdomain.com {
    redir https://gists.yourdomain.com{uri} permanent
}
```

Start Caddy:

```bash
# Reload configuration
sudo systemctl reload caddy

# Check status
sudo systemctl status caddy
```

### Traefik

Traefik with automatic SSL and service discovery:

Create `docker-compose.yml` with Traefik:

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    command:
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.tlschallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "letsencrypt:/letsencrypt"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.routers.traefik.tls.certresolver=letsencrypt"
      - "traefik.http.routers.traefik.service=api@internal"

  casgists:
    image: casapps/casgists:latest
    container_name: casgists
    restart: unless-stopped
    volumes:
      - casgists_data:/data
    environment:
      - CASGISTS_SERVER_HOST=0.0.0.0
      - CASGISTS_SERVER_PORT=64080
      - CASGISTS_SERVER_BASE_URL=https://gists.yourdomain.com
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.casgists.rule=Host(`gists.yourdomain.com`)"
      - "traefik.http.routers.casgists.tls.certresolver=letsencrypt"
      - "traefik.http.services.casgists.loadbalancer.server.port=64080"
      # Security headers
      - "traefik.http.middlewares.casgists-headers.headers.sslredirect=true"
      - "traefik.http.middlewares.casgists-headers.headers.stsseconds=31536000"
      - "traefik.http.middlewares.casgists-headers.headers.framedeny=true"
      - "traefik.http.routers.casgists.middlewares=casgists-headers"

volumes:
  casgists_data:
  letsencrypt:
```

## SSL/TLS Configuration

### Let's Encrypt with Certbot

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d gists.yourdomain.com

# Test automatic renewal
sudo certbot renew --dry-run

# Automatic renewal (already setup by default)
sudo systemctl status certbot.timer
```

### Custom SSL Certificate

```bash
# Copy your certificate files
sudo cp yourdomain.crt /etc/ssl/certs/casgists.crt
sudo cp yourdomain.key /etc/ssl/private/casgists.key

# Set proper permissions
sudo chmod 644 /etc/ssl/certs/casgists.crt
sudo chmod 600 /etc/ssl/private/casgists.key
sudo chown root:root /etc/ssl/certs/casgists.crt /etc/ssl/private/casgists.key
```

Update Nginx configuration to use custom certificates:

```nginx
ssl_certificate /etc/ssl/certs/casgists.crt;
ssl_certificate_key /etc/ssl/private/casgists.key;
```

## Database Setup

### PostgreSQL

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Create user and database
sudo -u postgres psql
```

```sql
-- Create user
CREATE USER casgists WITH ENCRYPTED PASSWORD 'secure_random_password';

-- Create database
CREATE DATABASE casgists OWNER casgists;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE casgists TO casgists;

-- Exit
\q
```

Configure PostgreSQL for production:

```bash
# Edit PostgreSQL configuration
sudo nano /etc/postgresql/15/main/postgresql.conf
```

```conf
# Performance tuning
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 128MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100

# Connection settings
max_connections = 100
```

```bash
# Restart PostgreSQL
sudo systemctl restart postgresql
```

Update CasGists configuration:

```yaml
database:
  type: postgresql
  host: localhost
  port: 5432
  name: casgists
  user: casgists
  password: secure_random_password
  ssl_mode: require
  max_open_connections: 25
  max_idle_connections: 5
  connection_max_lifetime: 5m
```

### MySQL

```bash
# Install MySQL
sudo apt install mysql-server

# Secure installation
sudo mysql_secure_installation

# Create user and database
sudo mysql
```

```sql
-- Create database
CREATE DATABASE casgists CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'casgists'@'localhost' IDENTIFIED BY 'secure_random_password';

-- Grant permissions
GRANT ALL PRIVILEGES ON casgists.* TO 'casgists'@'localhost';
FLUSH PRIVILEGES;

-- Exit
EXIT;
```

## Redis Setup (Optional)

For enhanced search and caching:

```bash
# Install Redis
sudo apt install redis-server

# Configure Redis
sudo nano /etc/redis/redis.conf
```

```conf
# Bind to localhost
bind 127.0.0.1

# Require authentication
requirepass your_redis_password

# Memory management
maxmemory 256mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000
```

```bash
# Restart Redis
sudo systemctl restart redis-server
```

Update CasGists configuration:

```yaml
search:
  backend: redis
  redis:
    enabled: true
    host: localhost
    port: 6379
    password: your_redis_password
    database: 0

cache:
  backend: redis
  redis:
    enabled: true
    host: localhost
    port: 6379
    password: your_redis_password
    database: 1
```

## Monitoring and Logging

### Log Aggregation

Configure centralized logging:

```bash
# Install rsyslog (usually pre-installed)
sudo apt install rsyslog

# Configure CasGists to log to syslog
```

Update CasGists configuration:

```yaml
logging:
  level: info
  format: json
  output: syslog
  syslog:
    network: udp
    address: localhost:514
    tag: casgists
```

### Monitoring with Prometheus

Create monitoring configuration:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'casgists'
    static_configs:
      - targets: ['localhost:64080']
    metrics_path: /metrics
    scrape_interval: 30s
```

Enable metrics in CasGists:

```yaml
monitoring:
  metrics:
    enabled: true
    path: /metrics
    listen_address: 127.0.0.1:9090
```

### Health Checks

Configure health checks for load balancers:

```bash
# Test health endpoint
curl http://localhost:64080/health

# Expected response:
{
  "status": "ok",
  "version": "v1.0.0",
  "database": "connected",
  "uptime": "24h30m15s"
}
```

Add to your load balancer configuration:

```nginx
# Nginx upstream with health check
upstream casgists_backend {
    server 127.0.0.1:64080 max_fails=3 fail_timeout=30s;
    # Add more servers for load balancing
    # server 127.0.0.1:64081 max_fails=3 fail_timeout=30s;
}
```

## Security Hardening

### Firewall Configuration

```bash
# Install UFW
sudo apt install ufw

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH
sudo ufw allow ssh

# Allow HTTP/HTTPS
sudo ufw allow 'Nginx Full'

# Allow specific port for CasGists (if needed)
sudo ufw allow from 127.0.0.1 to any port 64080

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status verbose
```

### System Hardening

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install security updates automatically
sudo apt install unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades

# Disable unused services
sudo systemctl disable apache2 # if installed
sudo systemctl disable mysql # if not using

# Configure fail2ban
sudo apt install fail2ban

# Create fail2ban config for Nginx
sudo nano /etc/fail2ban/jail.local
```

```ini
[nginx-http-auth]
enabled = true

[nginx-limit-req]
enabled = true

[nginx-botsearch]
enabled = true
```

### Application Security

Update CasGists security configuration:

```yaml
security:
  # Strong secret key (64+ characters)
  secret_key: your-very-long-and-secure-secret-key-here
  
  # Secure session settings
  session:
    secure: true
    same_site: strict
    http_only: true
  
  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_minute: 100
    burst_size: 20
  
  # CORS (if needed)
  cors:
    enabled: false
  
  # CSP headers
  csp:
    enabled: true
    default_src: "'self'"
    script_src: "'self'"
    style_src: "'self' 'unsafe-inline'"
    img_src: "'self' data: https:"
    font_src: "'self'"
    connect_src: "'self'"
    frame_ancestors: "'none'"
```

## Backup Strategy

### Automated Database Backups

Create backup script:

```bash
#!/bin/bash
# /usr/local/bin/backup-casgists.sh

BACKUP_DIR="/var/backups/casgists"
DB_NAME="casgists"
DB_USER="casgists"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p ${BACKUP_DIR}

# PostgreSQL backup
export PGPASSWORD="your_db_password"
pg_dump -h localhost -U ${DB_USER} -Fc ${DB_NAME} > ${BACKUP_DIR}/casgists_${DATE}.sql

# Compress and upload to S3 (optional)
# aws s3 cp ${BACKUP_DIR}/casgists_${DATE}.sql s3://your-backup-bucket/

# Cleanup old backups (keep 30 days)
find ${BACKUP_DIR} -name "casgists_*.sql" -mtime +30 -delete

# Set permissions
chmod 600 ${BACKUP_DIR}/casgists_${DATE}.sql
```

Schedule with cron:

```bash
# Edit crontab
sudo crontab -e

# Add daily backup at 2 AM
0 2 * * * /usr/local/bin/backup-casgists.sh
```

### Application Data Backup

```bash
#!/bin/bash
# Backup CasGists data directory

DATA_DIR="/var/lib/casgists"
BACKUP_DIR="/var/backups/casgists"
DATE=$(date +%Y%m%d_%H%M%S)

# Stop service temporarily
sudo systemctl stop casgists

# Create tar backup
tar -czf ${BACKUP_DIR}/casgists_data_${DATE}.tar.gz -C ${DATA_DIR} .

# Start service
sudo systemctl start casgists

# Upload to S3 (optional)
# aws s3 cp ${BACKUP_DIR}/casgists_data_${DATE}.tar.gz s3://your-backup-bucket/
```

## Performance Optimization

### System Tuning

```bash
# Increase file descriptor limits
echo "casgists soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "casgists hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Tune kernel parameters
echo "net.core.somaxconn = 65535" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" | sudo tee -a /etc/sysctl.conf
echo "vm.swappiness = 10" | sudo tee -a /etc/sysctl.conf

# Apply changes
sudo sysctl -p
```

### Application Tuning

```yaml
# CasGists performance configuration
server:
  read_timeout: 30s
  write_timeout: 30s
  max_request_size: 52428800 # 50MB

database:
  max_open_connections: 50
  max_idle_connections: 10
  connection_max_lifetime: 5m

cache:
  backend: redis
  redis:
    pool_size: 20
    min_idle_connections: 5

logging:
  level: warn # Reduce log verbosity in production
```

## Load Balancing

For high availability with multiple instances:

### Nginx Load Balancer

```nginx
upstream casgists_cluster {
    least_conn;
    server casgists1.internal:64080 max_fails=3 fail_timeout=30s;
    server casgists2.internal:64080 max_fails=3 fail_timeout=30s;
    server casgists3.internal:64080 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl http2;
    server_name gists.yourdomain.com;
    
    location / {
        proxy_pass http://casgists_cluster;
        # ... other proxy settings
    }
}
```

### Session Affinity

For sticky sessions (if needed):

```nginx
upstream casgists_cluster {
    ip_hash;
    server casgists1.internal:64080;
    server casgists2.internal:64080;
    server casgists3.internal:64080;
}
```

## Troubleshooting

### Common Issues

1. **502 Bad Gateway**
   ```bash
   # Check if CasGists is running
   sudo systemctl status casgists
   
   # Check port binding
   sudo netstat -tulpn | grep :64080
   
   # Check logs
   sudo journalctl -u casgists -f
   ```

2. **SSL Certificate Issues**
   ```bash
   # Test SSL
   openssl s_client -connect gists.yourdomain.com:443
   
   # Check certificate expiry
   openssl x509 -in /etc/ssl/certs/casgists.crt -text -noout | grep "Not After"
   
   # Renew Let's Encrypt
   sudo certbot renew
   ```

3. **Database Connection Issues**
   ```bash
   # Test database connection
   casgists config-check --test-db
   
   # Check PostgreSQL status
   sudo systemctl status postgresql
   
   # Check PostgreSQL logs
   sudo tail -f /var/log/postgresql/postgresql-15-main.log
   ```

4. **Performance Issues**
   ```bash
   # Check system resources
   htop
   iostat -x 1
   
   # Check application metrics
   curl http://localhost:64080/metrics
   
   # Analyze slow queries (PostgreSQL)
   # Enable log_min_duration_statement in postgresql.conf
   ```

### Monitoring Commands

```bash
# Check service status
sudo systemctl status casgists

# View logs
sudo journalctl -u casgists -f --no-pager

# Check network connections
sudo netstat -tulpn | grep casgists

# Monitor resource usage
sudo htop
sudo iotop

# Check disk space
df -h

# Test application
curl -I http://localhost:64080/health
```

## Deployment Checklist

- [ ] Server provisioned with adequate resources
- [ ] Domain name configured with DNS
- [ ] SSL certificate obtained and configured
- [ ] Reverse proxy configured and tested
- [ ] Database setup and optimized
- [ ] CasGists installed as system service
- [ ] Configuration file created and validated
- [ ] Firewall configured with minimal required ports
- [ ] Monitoring and logging setup
- [ ] Backup strategy implemented and tested
- [ ] Security headers configured
- [ ] Rate limiting enabled
- [ ] Health checks configured
- [ ] Performance tuning applied
- [ ] Documentation updated for your deployment

## Deployment Examples

### Small Team (1-50 users)

- **Server**: 2 vCPU, 4GB RAM, 50GB SSD
- **Database**: SQLite
- **Proxy**: Caddy with automatic HTTPS
- **Monitoring**: Basic systemd logging

### Medium Organization (50-500 users)

- **Server**: 4 vCPU, 8GB RAM, 100GB SSD
- **Database**: PostgreSQL
- **Cache**: Redis
- **Proxy**: Nginx with custom SSL
- **Monitoring**: Prometheus + Grafana
- **Backup**: Daily automated backups to S3

### Large Enterprise (500+ users)

- **Servers**: 3+ instances with load balancing
- **Database**: PostgreSQL cluster with replication
- **Cache**: Redis cluster
- **Proxy**: Nginx or cloud load balancer
- **Storage**: S3-compatible object storage
- **Monitoring**: Full observability stack
- **Backup**: Multi-region automated backups
- **Security**: WAF, DDoS protection, compliance scanning