# CasGists

**Production-Ready Self-Hosted GitHub Gist Alternative - v1.0.0**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/casapps/casgists)](https://goreportcard.com/report/github.com/casapps/casgists)
[![Release](https://img.shields.io/github/release/casapps/casgists.svg)](https://github.com/casapps/casgists/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/casapps/casgists)](https://hub.docker.com/r/casapps/casgists)

CasGists is a production-ready, self-hosted GitHub Gist alternative designed for non-technical self-hosters and small/medium businesses. It provides a secure, feature-rich platform for managing code snippets with enterprise-grade security and ease of use.

> **🎉 Version 1.0.0 Release** - Production-ready with complete features, enterprise security, and professional deployment options.

## ✨ Key Features

### 🚀 **Core Functionality**
- **🔐 Security First**: 2FA/TOTP, WebAuthn, audit logs, GDPR compliance
- **⚡ Zero Dependencies**: Single binary with embedded SQLite, no external services required
- **👥 Enterprise Ready**: Organizations, teams, RBAC, LDAP/SAML integration
- **🔍 Intelligent Search**: Full-text search with FTS5/Redis, advanced filtering
- **🎨 Modern UI**: Progressive Web App with dark/light themes, mobile-first design
- **🌐 Full Git Integration**: Real Git repos, branching, history, webhooks with HMAC signing

### 🏢 **Enterprise Features**
- **Organizations & Teams** - Multi-level hierarchy with role-based permissions
- **Compliance Ready** - GDPR, SOC2, HIPAA with data export/deletion
- **Audit Logging** - Complete activity tracking for compliance
- **Single Sign-On** - SAML 2.0, OAuth2, LDAP/Active Directory
- **API & Webhooks** - REST API with scoped tokens, real-time webhooks
- **Migration Tools** - Import from GitHub, GitLab, OpenGist

## 📦 Installation

### System Installation (Recommended for Production)

```bash
# Download latest release
wget https://github.com/casapps/casgists/releases/latest/download/casgists-linux-amd64
chmod +x casgists-linux-amd64

# Install as system service
sudo ./casgists-linux-amd64 install --port 64080

# Start the service
sudo systemctl start casgists
sudo systemctl enable casgists

# Visit http://localhost:64080 to complete setup wizard
```

### Docker Deployment

```bash
# Using Docker Compose (recommended)
curl -o docker-compose.yml https://raw.githubusercontent.com/casapps/casgists/main/docker-compose.yml
docker-compose up -d

# Or run directly with persistent data
docker run -d \
  --name casgists \
  -p 64080:64080 \
  -v casgists_data:/data \
  casapps/casgists:latest
```

### Binary Download

Download the latest binary for your platform from our [releases page](https://github.com/casapps/casgists/releases/latest):

- Linux: `casgists-linux-amd64`, `casgists-linux-arm64`
- macOS: `casgists-darwin-amd64`, `casgists-darwin-arm64`
- Windows: `casgists-windows-amd64.exe`

```bash
# Download and run
./casgists-linux-amd64

# Visit http://localhost:64080 to complete setup wizard
```

## 🚀 Quick Start

After installation, CasGists will guide you through an 8-step setup wizard to configure:

1. **Admin Account** - Create your first admin user
2. **Database** - Choose SQLite, PostgreSQL, or MySQL
3. **Server Settings** - Configure port and URL
4. **Email** - Optional SMTP configuration
5. **Authentication** - Enable 2FA, WebAuthn, SSO
6. **Features** - Enable organizations, registration, etc.
7. **Storage** - Configure data directory and backups
8. **Review** - Confirm and apply configuration

Once setup is complete, you can:
- Create and manage gists
- Organize with organizations and teams
- Configure webhooks and integrations
- Import existing gists from GitHub/GitLab/OpenGist
- Enable enterprise features as needed

## 🔧 Configuration

CasGists can be configured using environment variables or the web-based setup wizard.

### Core Settings

```bash
# Database (required)
CASGISTS_DB_TYPE=sqlite                    # sqlite|postgresql|mysql
CASGISTS_DB_PATH=/data/casgists.db         # For SQLite

# Network (optional)
CASGISTS_LISTEN_PORT=64001                 # Default: random port 64000-64999
CASGISTS_SERVER_URL=https://gists.example.com

# Security (auto-generated if not set)
CASGISTS_SECRET_KEY=your-secret-key

# Storage (optional)
CASGISTS_DATA_DIR=/data                    # Main data directory
```

### Email (Optional)

```bash
CASGISTS_SMTP_HOST=smtp.gmail.com
CASGISTS_SMTP_USERNAME=your-email@gmail.com
CASGISTS_SMTP_PASSWORD=your-app-password
CASGISTS_SMTP_FROM_EMAIL=noreply@example.com
```

### Features (Optional)

```bash
CASGISTS_FEATURES_REGISTRATION=true        # Allow new user registration
CASGISTS_FEATURES_ORGANIZATIONS=true       # Enable organizations
CASGISTS_FEATURES_SOCIAL=true              # Enable social features
```

## 🏗️ Architecture

CasGists is built with:

- **Backend**: Go with Echo web framework
- **Database**: SQLite (default), PostgreSQL, MySQL via GORM
- **Frontend**: Server-rendered HTML with HTMX and modern CSS
- **Authentication**: JWT tokens with 2FA and WebAuthn support
- **Storage**: Local filesystem or S3-compatible storage
- **Search**: Redis/Valkey (preferred) or SQLite FTS (fallback)
- **Version Control**: go-git library (no external Git dependency)

## 🛡️ Security Features

- **Multi-factor Authentication**: TOTP and WebAuthn/Passkeys
- **Audit Logging**: Complete audit trail of all actions
- **Rate Limiting**: Configurable rate limits for all endpoints
- **Content Security Policy**: Comprehensive CSP headers
- **Input Validation**: Rigorous input validation and sanitization
- **Compliance Ready**: GDPR, SOC2, and HIPAA compliance features

## 📖 Documentation

- [Installation Guide](docs/installation.md)
- [Configuration Reference](docs/configuration.md)
- [API Documentation](docs/api.md)
- [User Guide](docs/user-guide.md)
- [Production Deployment](docs/production-deployment.md)

## 📝 License

CasGists is released under the MIT License. See [LICENSE.md](LICENSE.md) for details.

## 📊 Status

- **Build Status**: [![Build Status](https://github.com/casapps/casgists/workflows/CI/badge.svg)](https://github.com/casapps/casgists/actions)
- **Release**: [![GitHub release](https://img.shields.io/github/release/casapps/casgists.svg)](https://github.com/casapps/casgists/releases)
- **License**: [![GitHub license](https://img.shields.io/github/license/casapps/casgists.svg)](https://github.com/casapps/casgists/blob/main/LICENSE.md)

## 💬 Community

- [GitHub Discussions](https://github.com/casapps/casgists/discussions) - Questions and discussions
- [GitHub Issues](https://github.com/casapps/casgists/issues) - Bug reports and feature requests

## 📈 Roadmap

See our [public roadmap](https://github.com/casapps/casgists/projects/1) for upcoming features and improvements.

---

## 🔨 Development

### Development Setup

```bash
# Clone repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Install dependencies
go mod download

# Build binary
make build

# Run the binary
./binaries/casgists

# Run tests
make test

# Build Docker image
make docker

# Build multi-platform release
make release
```

### Project Structure

```
casgists/
├── src/                    # All source code
│   ├── cmd/casgists/      # Main application entry point
│   ├── internal/          # Internal packages
│   └── web/               # Web assets (embedded)
├── scripts/               # Production/install scripts
├── tests/                 # Development/test scripts
├── binaries/              # Built binaries
├── docs/                  # Documentation (ReadTheDocs)
├── Makefile              # Build system
├── Dockerfile            # Docker build
└── release.txt           # Semantic version
```

### Building from Source

```bash
# Build for current platform
make build

# Build Docker image
make docker

# Build multi-platform release
make release

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
make test

# Run specific tests
go test ./src/internal/server/...
```

## 🤝 Contributing

We welcome contributions! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Build the project (`make build`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

Please ensure:
- All tests pass
- Code follows Go conventions
- Documentation is updated
- Commit messages are clear

## 🙏 Acknowledgments

CasGists is inspired by and builds upon the work of several open source projects:

- [OpenGist](https://github.com/thomiceli/opengist) - The original inspiration
- [Gitea](https://gitea.io) - Git management concepts
- [Echo](https://echo.labstack.com) - Web framework
- [GORM](https://gorm.io) - Database ORM

---

Made with ❤️ by the CasApps team
