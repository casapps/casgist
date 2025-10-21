# CasGists Installation Guide

## Quick Start

### 1. Download Binary
Download the appropriate binary for your platform from the [releases page](https://github.com/casapps/casgists/releases).

**Linux AMD64:**
```bash
wget https://github.com/casapps/casgists/releases/download/v1.0.0/casgists-linux-amd64
chmod +x casgists-linux-amd64
sudo mv casgists-linux-amd64 /usr/local/bin/casgists
```

**macOS:**
```bash
wget https://github.com/casapps/casgists/releases/download/v1.0.0/casgists-darwin-amd64
chmod +x casgists-darwin-amd64
sudo mv casgists-darwin-amd64 /usr/local/bin/casgists
```

**Windows:**
Download `casgists-windows-amd64.exe` and place it in your PATH.

### 2. Run Setup Wizard
```bash
casgists --setup
```

The setup wizard will guide you through:
1. System requirements check
2. Database configuration
3. Network settings
4. Email configuration
5. Security setup
6. Admin account creation
7. System service installation
8. Final confirmation

### 3. Start CasGists
```bash
# If installed as service
sudo systemctl start casgists  # Linux
sudo launchctl start com.casapps.casgists  # macOS
sc start casgists  # Windows

# Or run directly
casgists serve
```

## Manual Installation

### Prerequisites
- 512MB RAM (1GB recommended)
- 100MB free disk space + data storage
- Port 64000-64999 available

### Database Setup

**SQLite (Default):**
No setup required - database created automatically.

**PostgreSQL:**
```sql
CREATE DATABASE casgists;
CREATE USER casgists WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE casgists TO casgists;
```

**MySQL:**
```sql
CREATE DATABASE casgists CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'casgists'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON casgists.* TO 'casgists'@'localhost';
FLUSH PRIVILEGES;
```

### Configuration
Create `config.yaml`:
```yaml
server:
  port: 64000
  host: "0.0.0.0"
  
database:
  type: sqlite  # or postgresql, mysql
  dsn: "${DATA_DIR}/casgists.db"
  # PostgreSQL: "host=localhost user=casgists password=your_password dbname=casgists sslmode=disable"
  # MySQL: "casgists:your_password@tcp(localhost:3306)/casgists?charset=utf8mb4&parseTime=True&loc=Local"

paths:
  data_dir: "./data"
  log_dir: "${DATA_DIR}/logs"
  cache_dir: "${DATA_DIR}/cache"
  repo_dir: "${DATA_DIR}/repositories"

security:
  secret_key: "your-secret-key-here"  # Generate with: openssl rand -base64 32
  enable_2fa: true
  enable_webauthn: true
```

### Service Installation

**Linux (systemd):**
```bash
sudo casgists service install
sudo systemctl enable casgists
sudo systemctl start casgists
```

**macOS (launchd):**
```bash
sudo casgists service install
sudo launchctl load -w /Library/LaunchDaemons/com.casapps.casgists.plist
```

**Windows (Service):**
```powershell
# Run as Administrator
casgists.exe service install
sc start casgists
```

### Reverse Proxy Setup

**nginx:**
```nginx
server {
    listen 80;
    server_name gists.example.com;

    location / {
        proxy_pass http://localhost:64000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Caddy:**
```
gists.example.com {
    reverse_proxy localhost:64000
}
```

### Docker Installation
```bash
docker run -d \
  --name casgists \
  -p 64000:64000 \
  -v casgists-data:/data \
  -e CASGISTS_SECRET_KEY="your-secret-key" \
  casapps/casgists:latest
```

### Environment Variables
All configuration can be overridden with environment variables:
- `CASGISTS_PORT` - Server port (default: 64000)
- `CASGISTS_DATA_DIR` - Data directory path
- `CASGISTS_DATABASE_DSN` - Database connection string
- `CASGISTS_SECRET_KEY` - Secret key for JWT tokens
- `CASGISTS_SMTP_HOST` - SMTP server for emails
- `CASGISTS_REDIS_URL` - Redis connection URL

### Post-Installation

1. **Access Admin Panel:** Navigate to `http://your-server:64000/admin`
2. **Configure Backups:** Set up automated backups in Admin > System > Backups
3. **Import Data:** Use Admin > Migration to import from GitHub/GitLab/OpenGist
4. **Configure Webhooks:** Set up integrations in Admin > Webhooks
5. **Enable 2FA:** Secure your admin account in Profile > Security

### Troubleshooting

**Port Already in Use:**
Change port in config.yaml or use environment variable:
```bash
CASGISTS_PORT=64001 casgists serve
```

**Database Connection Failed:**
Check database credentials and ensure database server is running.

**Permission Denied:**
Ensure data directory has correct permissions:
```bash
sudo chown -R casgists:casgists /var/lib/casgists
```

**Service Won't Start:**
Check logs:
```bash
sudo journalctl -u casgists -f  # Linux
sudo tail -f /var/log/casgists.log  # macOS
```

### Updating
1. Download new binary
2. Stop service
3. Replace binary
4. Run migrations: `casgists migrate`
5. Start service

### Uninstallation
```bash
# Stop and remove service
sudo systemctl stop casgists
sudo systemctl disable casgists
sudo rm /etc/systemd/system/casgists.service

# Remove data (optional)
sudo rm -rf /var/lib/casgists

# Remove binary
sudo rm /usr/local/bin/casgists
```

## Support
- Documentation: https://docs.casgists.com
- GitHub Issues: https://github.com/casapps/casgists/issues
- Community Forum: https://forum.casgists.com