# Installation Guide

This guide covers all installation methods for CasGists, from simple local development to production deployment.

## System Requirements

### Minimum Requirements
- **Operating System**: Linux, macOS, or Windows
- **Architecture**: AMD64 or ARM64
- **Memory**: 256MB RAM
- **Storage**: 1GB available disk space
- **Network**: Port access (default: 64080)

### Recommended for Production
- **Operating System**: Linux (Ubuntu 20.04+ or CentOS 8+)
- **Memory**: 1GB+ RAM
- **Storage**: 10GB+ SSD storage
- **Network**: Reverse proxy (Nginx, Caddy, Traefik)
- **Database**: PostgreSQL or MySQL (for high-load environments)

## Installation Methods

### 1. System Service Installation (Recommended for Production)

This method installs CasGists as a proper system service with automatic startup, security hardening, and proper file permissions.

#### Linux (systemd)

```bash
# Download the binary
wget https://github.com/casapps/casgists/releases/latest/download/casgists-linux-amd64
chmod +x casgists-linux-amd64

# Install system service (requires root)
sudo ./casgists-linux-amd64 install

# The installer will:
# - Create 'casgists' system user
# - Set up directory structure in /opt/casgists and /var/lib/casgists
# - Install systemd service file
# - Set up log rotation
# - Configure port binding capabilities

# Start and enable the service
sudo systemctl start casgists
sudo systemctl enable casgists

# Check service status
sudo systemctl status casgists
```

#### Linux (SysV Init)

For older systems without systemd:

```bash
sudo ./casgists-linux-amd64 install
# Automatically detects SysV and creates appropriate init script

# Start service
sudo service casgists start

# Enable on boot
sudo update-rc.d casgists defaults
```

#### macOS (launchd)

```bash
# Download macOS binary
wget https://github.com/casapps/casgists/releases/latest/download/casgists-darwin-amd64
chmod +x casgists-darwin-amd64

# Install system service (requires admin)
sudo ./casgists-darwin-amd64 install

# Start service
sudo launchctl start com.casapps.casgists

# Enable on boot (done automatically by installer)
```

#### Custom Installation Options

```bash
# Custom port and user
sudo ./casgists install --port 8080 --user myuser --data-dir /custom/data

# Install without creating system service
sudo ./casgists install --no-service

# Custom installation path
sudo ./casgists install --install-path /usr/local/casgists
```

### 2. Local Development Setup

For development, testing, or single-user installations:

```bash
# Download binary
wget https://github.com/casapps/casgists/releases/latest/download/casgists-linux-amd64
chmod +x casgists-linux-amd64

# Run interactive setup wizard
./casgists-linux-amd64 setup

# The setup wizard will:
# - Choose between system and local installation
# - Configure port and data directory
# - Generate configuration file
# - Set up initial admin user (optional)

# Start server
./casgists-linux-amd64 serve

# Or specify custom config
./casgists-linux-amd64 serve --config /path/to/config.yaml
```

### 3. Docker Installation

#### Using Docker Compose (Recommended)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  casgists:
    image: casapps/casgists:latest
    container_name: casgists
    restart: unless-stopped
    ports:
      - "64080:64080"
    volumes:
      - casgists_data:/data
      - casgists_config:/config
    environment:
      - CASGISTS_SERVER_HOST=0.0.0.0
      - CASGISTS_SERVER_PORT=64080
      - CASGISTS_DATA_DIR=/data
      - CASGISTS_DB_PATH=/data/casgists.db
      # Add your custom environment variables here
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:64080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Add PostgreSQL for production
  postgres:
    image: postgres:15
    container_name: casgists-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=casgists
      - POSTGRES_USER=casgists
      - POSTGRES_PASSWORD=changeme
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  # Optional: Add Redis for enhanced search
  redis:
    image: redis:7-alpine
    container_name: casgists-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  casgists_data:
  casgists_config:
  postgres_data:
  redis_data:
```

Start with Docker Compose:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f casgists

# Stop services
docker-compose down
```

#### Using Docker Run

```bash
# Basic setup with SQLite
docker run -d \
  --name casgists \
  --restart unless-stopped \
  -p 64080:64080 \
  -v casgists_data:/data \
  -e CASGISTS_SERVER_HOST=0.0.0.0 \
  -e CASGISTS_SERVER_PORT=64080 \
  casapps/casgists:latest

# With PostgreSQL
docker run -d \
  --name casgists \
  --restart unless-stopped \
  -p 64080:64080 \
  -v casgists_data:/data \
  -e CASGISTS_DB_TYPE=postgresql \
  -e CASGISTS_DB_HOST=postgres \
  -e CASGISTS_DB_NAME=casgists \
  -e CASGISTS_DB_USER=casgists \
  -e CASGISTS_DB_PASSWORD=changeme \
  --link postgres:postgres \
  casapps/casgists:latest
```

### 4. Building from Source

For developers or custom builds:

```bash
# Prerequisites: Go 1.21+, Node.js 18+ (for web assets)
git clone https://github.com/casapps/casgists.git
cd casgists

# Install Go dependencies
go mod download

# Build web assets (optional, embedded versions included)
make build-assets

# Build binary
make build

# The binary will be in ./build/casgists
./build/casgists --version
```

#### Custom Build Options

```bash
# Build for specific platform
make build GOOS=linux GOARCH=arm64

# Build with custom version
make build VERSION=v1.0.0-custom

# Build without CGO (fully static binary)
make build CGO_ENABLED=0

# Build debug version
make build-debug
```

## Post-Installation Configuration

### 1. Initial Setup Wizard

After installation, visit your CasGists instance in a web browser:

```
http://localhost:64080
```

The setup wizard will guide you through:

1. **Database Configuration** - Choose SQLite, PostgreSQL, or MySQL
2. **Administrator Account** - Create the first admin user
3. **Basic Settings** - Site name, URL, registration settings
4. **Email Configuration** - SMTP settings for notifications (optional)
5. **Security Settings** - 2FA, session timeout, rate limiting

### 2. Configuration File

CasGists uses YAML configuration files. The default locations are:

- **System installation**: `/etc/casgists/config.yaml`
- **Local installation**: `./config.yaml`
- **Docker**: `/config/config.yaml`

Example minimal configuration:

```yaml
server:
  host: 0.0.0.0
  port: 64080
  base_url: https://gists.yourdomain.com

database:
  type: sqlite
  path: ${DATA_DIR}/casgists.db

security:
  secret_key: your-secure-random-key-here

logging:
  level: info
  file: ${DATA_DIR}/logs/casgists.log
```

### 3. Environment Variables

All configuration options can be set via environment variables:

```bash
# Server configuration
export CASGISTS_SERVER_HOST=0.0.0.0
export CASGISTS_SERVER_PORT=64080
export CASGISTS_SERVER_BASE_URL=https://gists.example.com

# Database
export CASGISTS_DB_TYPE=postgresql
export CASGISTS_DB_HOST=localhost
export CASGISTS_DB_PORT=5432
export CASGISTS_DB_NAME=casgists
export CASGISTS_DB_USER=casgists
export CASGISTS_DB_PASSWORD=secure_password

# Security
export CASGISTS_SECRET_KEY=your-secret-key
export CASGISTS_JWT_EXPIRY=24h

# Features
export CASGISTS_FEATURES_REGISTRATION=true
export CASGISTS_FEATURES_ANONYMOUS_GISTS=false
export CASGISTS_FEATURES_ORGANIZATIONS=true
```

## Verification

### Check Installation

```bash
# Verify binary
casgists --version

# Check configuration
casgists config-check

# Verify system installation
casgists verify-install
```

### Health Check

```bash
# Check if service is running
curl http://localhost:64080/health

# Expected response:
{
  "status": "ok",
  "version": "v1.0.0",
  "database": "connected",
  "uptime": "5m30s"
}
```

### Service Status

```bash
# Systemd
sudo systemctl status casgists
journalctl -u casgists -f

# Docker
docker logs casgists -f

# Manual process
ps aux | grep casgists
```

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Check what's using the port
   sudo netstat -tulpn | grep :64080
   
   # Use different port
   sudo ./casgists install --port 64081
   ```

2. **Permission denied**
   ```bash
   # Ensure proper permissions
   sudo chown -R casgists:casgists /var/lib/casgists
   sudo chmod 755 /var/lib/casgists
   ```

3. **Database connection failed**
   ```bash
   # Check database connectivity
   casgists config-check
   
   # Verify database exists and user has permissions
   ```

4. **Service fails to start**
   ```bash
   # Check logs
   journalctl -u casgists --no-pager
   
   # Check configuration
   casgists config-check
   
   # Run in foreground for debugging
   sudo -u casgists /opt/casgists/bin/casgists serve
   ```

### Getting Help

- **Configuration Issues**: Check the [Configuration Guide](configuration.md)
- **Performance**: See the [Performance Tuning Guide](performance.md)
- **Security**: Review the [Security Guide](security.md)
- **Bugs**: Report on [GitHub Issues](https://github.com/casapps/casgists/issues)
- **Questions**: Use [GitHub Discussions](https://github.com/casapps/casgists/discussions)

## Next Steps

After successful installation:

1. **Security**: Set up reverse proxy with TLS - see [Deployment Guide](deployment.md)
2. **Backup**: Configure automated backups - see [Backup Guide](backup.md)
3. **Monitoring**: Set up monitoring and alerting - see [Monitoring Guide](monitoring.md)
4. **Integration**: Configure webhooks and API access - see [API Documentation](../api/README.md)
5. **Migration**: Import existing gists - see [Migration Guide](migration.md)