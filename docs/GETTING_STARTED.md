# Getting Started with CasGists

Welcome to CasGists! This guide will help you get up and running quickly.

## Quick Start

### 1. Download and Install

```bash
# Clone the repository
git clone https://github.com/yourusername/casgists.git
cd casgists

# Build the application
make build

# Or use Docker
docker-compose up -d
```

### 2. Initial Configuration

Create a `config.yaml` file (copy from `config.example.yaml`):

```bash
cp config.example.yaml config.yaml
```

Edit the configuration to match your environment:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  base_url: "https://gists.yourdomain.com"

database:
  type: "sqlite"  # or "postgres", "mysql"
  dsn: "./data/casgists.db"

auth:
  jwt_secret: "your-secret-key-here"  # Generate with: openssl rand -base64 32
```

### 3. Run Database Migrations

```bash
./casgists migrate up
```

### 4. Create Admin User

```bash
./casgists user create --admin \
  --username admin \
  --email admin@example.com \
  --password yourpassword
```

### 5. Start the Server

```bash
# Development mode
./casgists server --dev

# Production mode
./casgists server
```

Visit http://localhost:8080 to access your CasGists instance!

## First Steps

### 1. Log In
- Navigate to http://localhost:8080/login
- Use the admin credentials you created
- Enable 2FA for enhanced security (recommended)

### 2. Create Your First Gist
- Click "New Gist" in the navigation
- Add your code or text
- Choose visibility (public, private, or unlisted)
- Add tags for organization
- Click "Create Gist"

### 3. Explore Features
- **Search**: Use the search bar with advanced syntax (e.g., `user:admin language:go`)
- **Tags**: Click on tags to filter gists
- **Fork**: Create your own copy of any public gist
- **Star**: Bookmark gists you find useful
- **Follow**: Keep track of other users' activities

## Key Features

### Authentication & Security
- JWT-based authentication with refresh tokens
- Two-factor authentication (2FA) with TOTP
- Session management with concurrent login limits
- Secure password policies

### Gist Management
- Multi-file gists with syntax highlighting
- Version history (via forks)
- Public, private, and unlisted visibility options
- Tag-based organization
- Full-text search

### Collaboration
- Follow other users
- Star favorite gists
- Fork gists for modifications
- Comment on gists (coming soon)

### Developer Features
- RESTful API with authentication
- Webhook support for real-time events
- CLI administration tools
- Multiple database support

## Common Tasks

### Backup Your Data
```bash
./casgists backup create --output ./backups/
```

### Update Email Settings
```bash
# Test email configuration
./casgists email test recipient@example.com

# View email queue
./casgists email queue
```

### Monitor System Health
```bash
# Check system status
./casgists system health

# View configuration
./casgists system config
```

### Manage Users
```bash
# List all users
./casgists user list

# Deactivate a user
./casgists user deactivate username

# Reset user password
./casgists user password username
```

## API Access

### Get API Token
1. Log in to your account
2. Go to Settings → API Tokens
3. Create a new token with desired permissions

### Example API Usage
```bash
# List your gists
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://gists.yourdomain.com/api/v1/gists

# Create a gist
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Gist",
    "description": "Example gist",
    "files": [{
      "name": "hello.go",
      "content": "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"
    }],
    "visibility": "public"
  }' \
  https://gists.yourdomain.com/api/v1/gists
```

## Troubleshooting

### Can't Connect to Database
- Check your database configuration in `config.yaml`
- Ensure the database server is running
- Verify connection string format

### Authentication Issues
- Clear browser cookies and cache
- Check JWT secret hasn't changed
- Verify system time is synchronized

### Email Not Sending
- Check SMTP configuration
- Test with: `./casgists email test your@email.com`
- Check email queue: `./casgists email queue`

### Performance Issues
- Enable Redis caching
- Check database indexes: `./casgists system cleanup`
- Monitor with: `curl http://localhost:8080/metrics`

## Next Steps

1. **Secure Your Instance**
   - Enable HTTPS with SSL/TLS
   - Configure firewall rules
   - Set up regular backups

2. **Customize**
   - Modify templates in `internal/web/templates/`
   - Add custom CSS/JS in `internal/web/static/`
   - Configure webhooks for integrations

3. **Scale**
   - Enable Redis for caching
   - Use PostgreSQL/MySQL for production
   - Deploy behind a load balancer

4. **Monitor**
   - Set up Prometheus for metrics
   - Configure alerts
   - Monitor logs

## Getting Help

- **Documentation**: See `/docs` directory
- **API Reference**: Visit `/api/v1/docs` when server is running
- **Issues**: Report bugs on GitHub
- **Community**: Join our Discord/Slack (coming soon)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

Happy Gisting! 🚀