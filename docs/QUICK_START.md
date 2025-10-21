# CasGists Quick Start Guide

Get CasGists up and running in under 5 minutes!

## Option 1: Docker (Fastest)

```bash
# Run with Docker
docker run -d \
  -p 3000:3000 \
  -v casgists-data:/data \
  --name casgists \
  casapps/casgists:latest

# Access at http://localhost:3000
```

## Option 2: Binary Download

```bash
# Linux/macOS
curl -L https://github.com/casapps/casgists/releases/latest/download/casgists-$(uname -s)-$(uname -m).tar.gz | tar xz
chmod +x casgists

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/casapps/casgists/releases/latest/download/casgists-windows-amd64.zip -OutFile casgists.zip
Expand-Archive casgists.zip -DestinationPath .

# Run setup wizard
./casgists setup

# Start server
./casgists serve
```

## Option 3: Build from Source

```bash
# Requirements: Go 1.21+
git clone https://github.com/casapps/casgists.git
cd casgists
go build -o casgists cmd/casgists/main.go
./casgists setup
./casgists serve
```

## First Time Setup

1. **Access Setup Wizard**: Navigate to http://localhost:3000
2. **Configure Database**: Choose SQLite for quick start
3. **Create Admin Account**: Set username and password
4. **Configure Features**: Enable desired features
5. **Start Using**: Create your first gist!

## Default Credentials

- **Username**: Set during setup
- **Password**: Set during setup
- **Default Port**: 3000

## Essential Commands

```bash
# Check health
curl http://localhost:3000/api/v1/health

# Create backup
./casgists backup

# View logs
./casgists logs

# Stop server
./casgists stop
```

## Quick Configuration

Create a `.env` file for easy configuration:

```bash
# Essential settings
CASGISTS_SERVER_PORT=3000
CASGISTS_SERVER_URL=http://localhost:3000
CASGISTS_SECURITY_SECRET_KEY=change-this-to-a-random-32-char-string

# Enable features
CASGISTS_FEATURES_REGISTRATION=true
CASGISTS_FEATURES_PUBLIC_GISTS=true
```

## Creating Your First Gist

1. **Login**: Use your admin credentials
2. **Click "New Gist"**
3. **Add Content**:
   - Title: "My First Gist"
   - File: `hello.py`
   - Content: `print("Hello, CasGists!")`
4. **Save**: Choose visibility (public/private)

## Next Steps

- Read the [Feature Reference](FEATURE_REFERENCE.md)
- Configure [Email Notifications](DEPLOYMENT_GUIDE.md#email)
- Set up [Webhooks](FEATURE_REFERENCE.md#webhook-system)
- Import from [GitHub](FEATURE_REFERENCE.md#github-import)

## Getting Help

- **Documentation**: [/docs](./docs/)
- **Issues**: [GitHub Issues](https://github.com/casapps/casgists/issues)
- **Health Check**: http://localhost:3000/api/v1/health

## Common Issues

### Port Already in Use
```bash
# Change port
CASGISTS_SERVER_PORT=8080 ./casgists serve
```

### Permission Denied
```bash
# Run without privileges
./casgists serve --user-mode
```

### Database Connection Failed
```bash
# Use SQLite (no setup required)
CASGISTS_DATABASE_TYPE=sqlite ./casgists serve
```