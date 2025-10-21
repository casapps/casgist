# Changelog

All notable changes to CasGists will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-27

### 🎉 Initial Release

CasGists v1.0.0 is the first production-ready release of the self-hosted GitHub Gist alternative designed for non-technical self-hosters and small/medium businesses.

### ✨ Features

#### Core Functionality
- **Gist Management**: Create, edit, delete, and organize code snippets with multiple files
- **Syntax Highlighting**: Support for 200+ programming languages
- **Markdown Support**: Full markdown rendering with GFM compatibility
- **Version Control**: Git-based versioning for all gists
- **Search**: Full-text search with SQLite FTS and Redis/Valkey support
- **Organizations**: Team collaboration with role-based permissions

#### Security
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **2FA Support**: TOTP-based two-factor authentication
- **WebAuthn/Passkeys**: Modern passwordless authentication
- **CSRF Protection**: Built-in CSRF protection for all forms
- **Rate Limiting**: Configurable rate limiting for API endpoints
- **Privilege Escalation**: Safe system-level operations with sudo/UAC support

#### Migration & Import
- **OpenGist Migration**: Complete migration from existing OpenGist instances
- **GitHub Import**: Import gists from GitHub with comment preservation
- **GitLab Import**: Import snippets from GitLab
- **URL Preservation**: Maintain existing URLs during migration

#### Administration
- **Setup Wizard**: 8-step guided setup for first-time installation
- **Admin Panel**: Comprehensive administrative interface
- **Backup/Restore**: Full system backup with scheduled automation
- **Service Installation**: Automatic systemd/launchd/Windows service setup
- **Email System**: SMTP email with customizable templates

#### Developer Features
- **REST API**: Complete RESTful API with OpenAPI documentation
- **Webhooks**: Event-driven webhooks with HMAC signatures
- **Swagger UI**: Interactive API documentation
- **ReDoc**: Alternative API documentation interface
- **API Explorer**: Built-in API testing interface

#### Deployment
- **Single Binary**: Zero-dependency static binary deployment
- **Multi-Platform**: Support for Linux, macOS, Windows, FreeBSD, OpenBSD
- **Multi-Architecture**: AMD64, ARM64, ARMv6, ARMv7 support
- **Docker Ready**: Container-friendly design
- **Automatic HTTPS**: Let's Encrypt integration

#### User Experience
- **Mobile-First Design**: Responsive UI optimized for all devices
- **PWA Support**: Progressive Web App with offline capabilities
- **Dark Mode**: Built-in dark theme support
- **Keyboard Shortcuts**: Vim-style navigation
- **Quick Actions**: Command palette for rapid navigation

### 🔧 Technical Specifications
- **Go 1.21+**: Modern Go with generics support
- **Echo Framework**: High-performance web framework
- **GORM**: Multi-database ORM (SQLite, PostgreSQL, MySQL)
- **go-git**: Pure Go Git implementation (no Git dependency)
- **Port Range**: 64000-64999 to avoid conflicts
- **CGO_ENABLED=0**: Maximum portability

### 📋 Requirements
- **Memory**: 512MB RAM minimum (1GB recommended)
- **Storage**: 100MB for application + data storage
- **Database**: SQLite (default), PostgreSQL, or MySQL
- **OS**: Linux, macOS, Windows, FreeBSD, OpenBSD

### 🚀 Getting Started
1. Download the appropriate binary for your platform
2. Run `./casgists --setup` to start the setup wizard
3. Follow the 8-step configuration process
4. Access CasGists at the configured URL

### 📝 Notes
- This is a production-ready v1.0.0 release
- No alpha, beta, or development versions - 100% stable
- Designed for non-technical users and enterprise deployment
- Security-first design with enterprise features invisible to regular users

[1.0.0]: https://github.com/casapps/casgists/releases/tag/v1.0.0