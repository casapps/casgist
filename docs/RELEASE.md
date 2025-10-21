# CasGists v1.0.0 Release

## 🎉 Welcome to CasGists v1.0.0!

CasGists is now ready for production use as a comprehensive, self-hosted alternative to GitHub Gists. This release represents a complete, feature-rich platform designed specifically for non-technical self-hosters and small to medium businesses.

## ✨ Key Features

### Core Functionality
- **Full GitHub Gists compatibility** - Complete feature parity with GitHub Gists
- **Multi-user support** - Organizations, teams, and individual accounts  
- **Advanced permissions** - Granular access control and visibility settings
- **Git operations** - Full Git history and version control using go-git
- **Multi-format export/import** - Support for GitHub, GitLab, and OpenGist migration

### Self-Hosting Optimized  
- **Single binary deployment** - Zero dependencies (CGO_ENABLED=0)
- **Multiple database support** - SQLite (default), PostgreSQL, MySQL
- **Port range 64000-64999** - Avoids conflicts with common services
- **Path variable substitution** - `${DATA_DIR}` for flexible deployments
- **8-step setup wizard** - Guided configuration for non-technical users

### Security & Compliance
- **JWT authentication** - With refresh tokens and session management
- **Two-factor authentication** - TOTP support for enhanced security
- **GDPR compliance** - Data export, deletion, and consent management
- **SOC2 & HIPAA ready** - Comprehensive audit logging and compliance features
- **CSRF protection** - Built-in security headers and protection
- **Rate limiting** - Configurable rate limits and DDoS protection

### Advanced Features
- **Progressive Web App (PWA)** - Full offline support and mobile-first design
- **Webhook system** - HMAC-signed webhooks with circuit breakers
- **Search system** - SQLite FTS with Redis/Elasticsearch upgrade path  
- **Email system** - SMTP integration with template support
- **Backup/restore** - Automated backups with encryption support
- **Admin panel** - Comprehensive system administration interface

## 🚀 Quick Start

### Requirements
- Linux, macOS, or Windows
- 1GB RAM (2GB recommended)
- 1GB disk space (plus storage for gists)

### Installation

1. **Download the binary:**
```bash
# Replace with actual download URL when published
wget https://github.com/casapps/casgists/releases/download/v1.0.0/casgists-linux-amd64
chmod +x casgists-linux-amd64
sudo mv casgists-linux-amd64 /usr/local/bin/casgists
```

2. **Create data directory:**
```bash
sudo mkdir -p /var/lib/casgists
sudo chown $USER:$USER /var/lib/casgists
```

3. **Run the setup wizard:**
```bash
casgists --setup
```

4. **Start CasGists:**
```bash
# Development
casgists --config development

# Production  
sudo casgists --config production
```

5. **Access the web interface:**
Open http://localhost:64200 in your browser and follow the setup wizard.

### Docker Deployment

```yaml
# docker-compose.yml
version: '3.8'
services:
  casgists:
    image: casgists/casgists:v1.0.0
    ports:
      - "64200:64200"
    volumes:
      - ./data:/var/lib/casgists
      - ./configs:/etc/casgists
    environment:
      - DATA_DIR=/var/lib/casgists
      - CONFIG_DIR=/etc/casgists
      - JWT_SECRET=your-secret-key-here
    restart: unless-stopped
```

## 📋 System Requirements

### Minimum Requirements
- **CPU:** 1 core
- **RAM:** 1GB
- **Storage:** 1GB + gist storage
- **OS:** Linux, macOS, Windows

### Recommended Requirements
- **CPU:** 2+ cores  
- **RAM:** 2GB+
- **Storage:** 10GB+ SSD
- **OS:** Linux (Ubuntu 20.04+ / CentOS 8+)

### Supported Databases
- **SQLite** (default, no setup required)
- **PostgreSQL** 12+
- **MySQL** 8.0+

## 🔧 Configuration

### Environment Variables
```bash
# Required for production
export JWT_SECRET="your-super-secure-jwt-secret-here"
export CSRF_SECRET="your-super-secure-csrf-secret-here"  
export WEBHOOK_SECRET="your-webhook-secret-here"
export BACKUP_ENCRYPTION_KEY="your-backup-encryption-key"

# Database (if using external DB)
export DB_TYPE="postgres"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="casgists"
export DB_USER="casgists"  
export DB_PASSWORD="your-db-password"

# Email (if using SMTP)
export SMTP_HOST="smtp.example.com"
export SMTP_PORT="587"
export SMTP_USER="noreply@example.com"
export SMTP_PASSWORD="your-email-password"
export SMTP_FROM="noreply@example.com"

# Optional
export PUBLIC_URL="https://gists.example.com"
export REDIS_HOST="localhost"
export REDIS_PASSWORD="your-redis-password"
```

### Configuration Files
- `configs/casgists.yaml` - Main configuration
- `configs/development.yaml` - Development overrides  
- `configs/production.yaml` - Production settings

## 📚 Documentation

### Available Documentation
- [Installation Guide](docs/INSTALLATION.md) - Detailed installation instructions
- [Configuration Guide](docs/CONFIGURATION.md) - Complete configuration reference
- [API Documentation](docs/API.md) - REST API reference
- [Development Guide](docs/DEVELOPMENT.md) - Developer setup and contribution guide
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment best practices

### Additional Resources
- **Admin Guide** - Access via `/admin` after setup
- **API Docs** - Interactive docs at `/api/v1/docs`
- **Health Checks** - Status endpoint at `/health`

## 🔒 Security Considerations

### Production Security Checklist
- [ ] Set strong, unique secrets for JWT, CSRF, and webhooks
- [ ] Use HTTPS with valid SSL certificates
- [ ] Enable two-factor authentication for admin accounts
- [ ] Configure rate limiting appropriately
- [ ] Set up regular backups with encryption
- [ ] Review and configure CORS settings
- [ ] Enable audit logging for compliance
- [ ] Set up monitoring and alerting

### Default Security Features
- CSRF protection enabled by default
- Secure session handling with HTTP-only cookies
- Rate limiting on authentication endpoints  
- Comprehensive audit logging
- GDPR compliance tools built-in

## 🚨 Known Limitations

### v1.0.0 Limitations
- **CLI tools** - Command-line interface not included in v1.0.0
- **Advanced search** - Full-text search requires Redis/Elasticsearch for optimal performance
- **Horizontal scaling** - Single-instance deployment (clustering planned for v1.1)
- **Plugin system** - Extension system planned for future release

### Browser Compatibility
- **Modern browsers** - Chrome 90+, Firefox 88+, Safari 14+
- **PWA features** - Requires HTTPS for full offline functionality
- **JavaScript required** - Core functionality needs JavaScript enabled

## 🎯 Roadmap

### Upcoming in v1.1
- [ ] Horizontal scaling and clustering support
- [ ] Advanced search with Elasticsearch integration  
- [ ] CLI tools and automation scripts
- [ ] Plugin system for extensions
- [ ] Advanced analytics dashboard
- [ ] LDAP/SAML authentication

### Long-term Goals
- Container orchestration templates (Kubernetes, Docker Swarm)
- Advanced webhook transformations and filters
- Integration with CI/CD platforms
- Mobile app for iOS and Android
- Real-time collaboration features

## 🤝 Support & Community

### Getting Help
- **Documentation** - Check the docs/ directory first
- **GitHub Issues** - Report bugs and request features
- **Discussions** - Community support and questions

### Contributing
- **Bug Reports** - Use GitHub Issues with bug template
- **Feature Requests** - Use GitHub Issues with feature template  
- **Code Contributions** - Follow the development guide
- **Documentation** - Help improve documentation

## 📝 License

CasGists is released under the MIT License. See [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

CasGists was built with inspiration from:
- **GitHub Gists** - The original and still amazing service
- **OpenGist** - Open source alternative that paved the way
- **GitLab Snippets** - Feature ideas and user experience inspiration

Special thanks to the open source community and all contributors who made this project possible.

---

## 🚀 Ready to Get Started?

1. **Download CasGists v1.0.0** from the releases page
2. **Follow the Quick Start guide** above  
3. **Join the community** for support and updates
4. **Star the project** if you find it useful!

**Happy self-hosting! 🎉**

*CasGists - Your gists, your server, your control.*