# CasGists Documentation

Welcome to **CasGists**, a self-hosted GitHub Gist alternative designed for non-technical self-hosters and small/medium businesses.

## What is CasGists?

CasGists is an open-source, self-hosted alternative to GitHub Gist that provides a complete gist management solution with organization support, modern authentication, and a focus on simplicity and security.

### Key Features

- **🚀 Simple Deployment**: Single static binary with zero dependencies
- **🔒 Security First**: JWT authentication, 2FA (TOTP), WebAuthn/Passkeys
- **📱 Mobile First**: Responsive design optimized for all devices
- **🏢 Organization Support**: Built-in multi-tenant organization management
- **🔍 Full-Text Search**: SQLite FTS or Redis/Valkey powered search
- **🎨 Theme Support**: Multiple themes including dark mode
- **📦 Git Backend**: All gists backed by Git repositories using go-git
- **🔄 Import/Export**: Migrate from GitHub Gist, OpenGist, or GitLab
- **🐳 Docker Ready**: Official Docker images available

## Quick Start

Get started with CasGists in minutes:

```bash
# Using Docker
docker run -d \
  -p 64000:64000 \
  -v casgists-data:/data \
  -e CASGISTS_DATABASE_TYPE=sqlite \
  -e CASGISTS_DATABASE_DSN=/data/casgists.db \
  casgists:latest

# Using binary
./casgists
```

Visit `http://localhost:64000` to access your CasGists instance.

## Core Principles

### Never Die
Always provide maximum functionality given current conditions. CasGists degrades gracefully when optional services are unavailable.

### Security First
Enterprise-grade security that's invisible to users but comprehensive for businesses. Automatic security best practices without configuration.

### Mobile First
Responsive design, intuitive navigation, readable content, follows web standards, and full accessibility support.

## Technology Stack

- **Backend**: Go 1.23+ with Echo web framework
- **Database**: SQLite (default), PostgreSQL, MySQL via GORM
- **Authentication**: JWT with refresh tokens, 2FA, WebAuthn
- **Frontend**: Modern responsive web interface with theme support
- **CLI**: POSIX-compliant shell script generated dynamically
- **Version Control**: go-git library (no external Git dependency)
- **Search**: Redis/Valkey (preferred) or SQLite FTS (fallback)
- **Caching**: In-memory LRU with Redis/Valkey fallback

## Target Audience

### Primary Users
- Non-technical self-hosters who want simple deployment
- Small and medium businesses needing private gist servers

### Secondary Users
- Users who prefer self-hosted solutions over cloud services
- Developers who want to embed gist functionality into their applications

## Getting Help

- **Documentation**: You're reading it! Browse the navigation for detailed guides.
- **GitHub Issues**: [Report bugs or request features](https://github.com/casapps/casgists/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/casapps/casgists/discussions)

## License

CasGists is open-source software licensed under the [MIT License](https://github.com/casapps/casgists/blob/main/LICENSE.md).
