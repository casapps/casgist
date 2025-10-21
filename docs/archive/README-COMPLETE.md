# CasGists - Self-Hosted GitHub Gist Alternative

[![Build Status](https://img.shields.io/github/workflow/status/casapps/casgists/CI)](https://github.com/casapps/casgists/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/casapps/casgists)](https://hub.docker.com/r/casapps/casgists)

CasGists is a powerful, self-hosted alternative to GitHub Gist with advanced features, modern UI, and enterprise-ready capabilities.

## ✨ Features

### Core Features
- 📝 **Create, edit, and share code snippets** with syntax highlighting
- 🔒 **Privacy controls** - Public, private, and unlisted gists
- 📂 **Multiple files per gist** with language detection
- 🔍 **Full-text search** across all your gists
- 🏷️ **Tagging system** for organization
- ⭐ **Star and watch** gists you love
- 🔄 **Version control** with Git backend
- 💬 **Comments and discussions** on gists

### Advanced Features
- 📱 **Progressive Web App (PWA)** with offline support
- 🌐 **Multi-language support** (10+ languages)
- 🎨 **Multiple themes** including Dracula, Monokai, etc.
- 📊 **Analytics and insights** for your gists
- 🔗 **Embed gists** in websites and blogs
- 📥 **Import from GitHub/GitLab/Bitbucket**
- 🚀 **RESTful API** with OpenAPI documentation
- 🔧 **CLI tool** for command-line operations
- 🪝 **Webhooks** for integrations
- 🔐 **Two-factor authentication** (2FA)

### Enterprise Features
- 👥 **Organizations** with team management
- 🔑 **SSO/SAML** integration (ready for implementation)
- 📋 **Audit logging** for compliance
- 🛡️ **Advanced security** features
- 📈 **Performance monitoring** with Prometheus
- 🐳 **Docker and Kubernetes** ready
- 🔄 **High availability** configuration
- 💾 **Multiple database** support (SQLite, PostgreSQL, MySQL)

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Create a directory for data
mkdir -p ~/casgists/data

# Run CasGists (single static binary with everything embedded)
docker run -d \
  --name casgists \
  -p 64001:64001 \
  -v ~/casgists/data:/app/data \
  -e CASGISTS_SECRET_KEY=$(openssl rand -base64 32) \
  casapps/casgists:latest

# CasGists will automatically detect your server's IP/FQDN and reverse proxy configuration
# Access via your server's IP or domain name on port 64001
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Copy environment file
cp deployments/docker/.env.example deployments/docker/.env

# Edit .env and set CASGISTS_SECRET_KEY
nano deployments/docker/.env

# Start services
cd deployments/docker
docker-compose up -d

# CasGists automatically detects your server configuration and provides the correct URLs
```

## 📦 Installation

### System Requirements
- **Single static binary** - No dependencies needed!
- **Go 1.23+** (only for building from source)
- **Database** - SQLite embedded by default (PostgreSQL/MySQL optional)
- **No external services** required (Redis optional for enhanced caching)

### Building from Source

```bash
# Clone repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Install to system
sudo make install
```

### Manual Installation

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed installation instructions.

## 🔧 Configuration

CasGists can be configured via:
1. Configuration file (`config.yaml`)
2. Environment variables
3. Command-line flags

### Basic Configuration

```yaml
server:
  host: 0.0.0.0
  port: 64001
  url: https://gists.example.com

database:
  driver: postgres
  dsn: postgres://user:pass@localhost/casgists

security:
  secret_key: ${CASGISTS_SECRET_KEY}

features:
  registration: true
  public_gists: true
```

### Environment Variables

```bash
CASGISTS_SERVER_PORT=64001
CASGISTS_SERVER_URL=https://gists.example.com
CASGISTS_DATABASE_DSN=postgres://user:pass@localhost/casgists
CASGISTS_SECRET_KEY=your-secret-key
```

## 🖥️ Usage

### Web Interface

1. Navigate to your server's IP address or domain on port 64001
2. Register a new account or login
3. Click "New Gist" to create your first gist
4. Share the URL or embed in your website

### Command Line

```bash
# Install CLI
casgists cli install

# Login
casgists login

# Create a gist
casgists create file.js --description "My JavaScript code"

# List your gists
casgists list

# Search gists
casgists search "function"
```

### API Usage

```bash
# Get API token from settings
TOKEN="your-api-token"

# Create a gist (replace with your server's IP/domain)
curl -X POST http://your-server:64001/api/v1/gists \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "files": [{
      "filename": "hello.js",
      "content": "console.log(\"Hello, World!\");"
    }]
  }'

# List gists  
curl http://your-server:64001/api/v1/gists \
  -H "Authorization: Bearer $TOKEN"
```

## 🔌 Integrations

### Embedding Gists

```html
<!-- Embed a gist -->
<script src="https://gists.example.com/embed/gist-id.js"></script>

<!-- Embed with specific theme -->
<script src="https://gists.example.com/embed/gist-id.js?theme=dark"></script>
```

### Webhooks

Configure webhooks in settings to receive notifications:
- Gist created/updated/deleted
- User starred/unstarred
- New comments

### Import from Other Services

1. Go to Settings → Import
2. Select service (GitHub/GitLab/Bitbucket)
3. Authenticate and select gists to import

## 🛠️ Development

### Prerequisites

- Go 1.23+
- Node.js 18+ (for frontend development)
- Make
- Git

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/casapps/casgists.git
cd casgists

# Install dependencies
make deps

# Run in development mode
make run-dev

# Run tests
make test

# Generate coverage report
make coverage
```

### Project Structure

```
casgists/
├── cmd/casgists/        # Main application entry
├── internal/            # Internal packages
│   ├── api/            # API handlers
│   ├── auth/           # Authentication
│   ├── database/       # Database models
│   ├── services/       # Business logic
│   └── web/            # Web handlers
├── web/                # Frontend assets
│   ├── static/         # Static files
│   ├── templates/      # HTML templates
│   └── views/          # View components
├── deployments/        # Deployment configs
├── docs/               # Documentation
└── tests/              # Test files
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make coverage

# Run linter
make lint
```

## 📊 Performance

CasGists is optimized for performance:
- **Database query optimization** with indexes
- **In-memory caching** for frequently accessed data
- **CDN-ready** static asset serving
- **Compression** for all responses
- **Connection pooling** for database
- **Background job processing** for heavy tasks

Benchmarks (on modest hardware):
- List gists: < 50ms
- Create gist: < 100ms
- Search: < 200ms
- Concurrent users: 1000+

## 🔒 Security

- **Encrypted passwords** with bcrypt
- **JWT tokens** for API authentication
- **CSRF protection** for web forms
- **Rate limiting** to prevent abuse
- **Input validation** and sanitization
- **SQL injection** prevention
- **XSS protection** with CSP headers
- **Security headers** (HSTS, X-Frame-Options, etc.)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## 📄 License

CasGists is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by GitHub Gist
- Built with Go, Echo, GORM, and modern web technologies
- UI components from DaisyUI and Tailwind CSS
- Syntax highlighting by Prism.js

## 📞 Support

- 📧 Email: support@casgists.com
- 💬 Discord: [Join our community](https://discord.gg/casgists)
- 🐛 Issues: [GitHub Issues](https://github.com/casapps/casgists/issues)
- 📖 Docs: [Documentation](https://docs.casgists.com)

---

Made with ❤️ by the CasGists team