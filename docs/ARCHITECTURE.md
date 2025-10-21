# CasGists Architecture

## Overview

CasGists follows a clean architecture pattern with clear separation of concerns, dependency injection, and a layered approach that makes the codebase maintainable, testable, and scalable.

## Directory Structure

```
casgists/
├── cmd/
│   └── casgists/        # Application entry point
├── internal/            # Private application code
│   ├── api/            # API endpoints and handlers
│   │   └── v1/         # API version 1
│   ├── auth/           # Authentication & authorization
│   ├── backup/         # Backup and restore functionality
│   ├── cache/          # Caching layer abstraction
│   ├── cli/            # CLI commands and administration
│   ├── config/         # Configuration management
│   ├── database/       # Database connections and migrations
│   ├── email/          # Email notifications and templates
│   ├── metrics/        # Prometheus metrics
│   ├── models/         # Data models and structures
│   ├── performance/    # Performance optimizations
│   ├── repositories/   # Data access layer
│   ├── search/         # Search functionality
│   ├── services/       # Business logic layer
│   ├── utils/          # Utility functions
│   ├── web/           # Web UI and templates
│   └── webhooks/      # Webhook system
├── docs/              # Documentation
├── scripts/           # Build and deployment scripts
└── tests/             # Integration tests
```

## Core Components

### 1. Layered Architecture

```
┌─────────────────────────────────────────┐
│          API / Web UI Layer             │
├─────────────────────────────────────────┤
│          Service Layer                  │
├─────────────────────────────────────────┤
│        Repository Layer                 │
├─────────────────────────────────────────┤
│      Database / Cache Layer            │
└─────────────────────────────────────────┘
```

#### API/Web Layer
- Handles HTTP requests and responses
- Input validation and error formatting
- Authentication middleware
- Rate limiting and CORS

#### Service Layer
- Business logic implementation
- Transaction management
- Event triggering (webhooks, emails)
- Cache invalidation

#### Repository Layer
- Data access abstraction
- Database queries and operations
- Query optimization
- Cache integration

#### Infrastructure Layer
- Database connections
- Cache management
- External services (email, webhooks)
- File storage

### 2. Database Design

```sql
Users
├── id (UUID, Primary Key)
├── username (Unique)
├── email (Unique)
├── password_hash
└── created_at, updated_at

Gists
├── id (UUID, Primary Key)
├── user_id (Foreign Key)
├── title
├── description
├── visibility
└── created_at, updated_at

Files
├── id (UUID, Primary Key)
├── gist_id (Foreign Key)
├── name
├── content
└── language

Sessions
├── id (UUID, Primary Key)
├── user_id (Foreign Key)
├── token_hash
├── expires_at
└── ip_address, user_agent
```

### 3. Authentication Flow

```
Client                  Server
  │                       │
  ├─ Login Request ──────>│
  │                       ├─ Validate Credentials
  │                       ├─ Generate JWT Tokens
  │<─── Access Token ─────┤
  │<── Refresh Token ─────┤
  │                       │
  ├─ API Request ────────>│
  │  (with Access Token)  ├─ Validate Token
  │                       ├─ Process Request
  │<──── Response ────────┤
```

### 4. Caching Strategy

```
Request
  │
  ├─> Check Cache
  │     │
  │     ├─ Hit ──> Return Cached Data
  │     │
  │     └─ Miss ─> Query Database
  │                   │
  │                   ├─> Store in Cache
  │                   │
  │                   └─> Return Data
```

**Cache Layers:**
1. **Redis** (Primary)
   - User sessions
   - Search results
   - Popular gists
   - API responses

2. **In-Memory** (Fallback)
   - Thread-safe LRU cache
   - Automatic expiration
   - Limited size

### 5. Search Architecture

```
Search Query
  │
  ├─> Parse Query Syntax
  │     │
  │     ├─> Extract Filters
  │     └─> Build Search Terms
  │
  ├─> Check Cache
  │
  └─> Search Backend
        │
        ├─> Redis Search (if available)
        │
        └─> SQLite FTS (fallback)
```

### 6. Webhook System

```
Event Occurs
  │
  ├─> Service Triggers Event
  │
  ├─> Webhook Manager
  │     │
  │     ├─> Find Subscribed Webhooks
  │     ├─> Generate Payload
  │     ├─> Sign with HMAC
  │     └─> Queue Delivery
  │
  └─> Delivery Worker
        │
        ├─> HTTP POST Request
        ├─> Handle Response
        └─> Retry if Failed
```

### 7. Email System

```
Email Trigger
  │
  ├─> Create Email Entry
  │
  ├─> Queue in Database
  │
  └─> Email Processor
        │
        ├─> Fetch from Queue
        ├─> Render Template
        ├─> Send via SMTP
        └─> Update Status
```

## Security Architecture

### Authentication
- **Argon2id** for password hashing
- **JWT** with RS256 for tokens
- **TOTP** for 2FA
- **Session management** with device tracking

### Authorization
- Role-based access control (Admin, User)
- Resource-level permissions
- API token scopes

### Data Protection
- SQL injection prevention via GORM
- XSS protection headers
- CSRF protection
- Input validation at all layers

## Performance Optimizations

### Database
- Connection pooling
- Prepared statements
- Efficient indexing
- Query optimization

### Caching
- Multi-level caching
- Cache-aside pattern
- Intelligent invalidation
- Request coalescing

### HTTP
- Gzip compression
- ETag support
- Static asset caching
- Response buffering

## Scalability Considerations

### Horizontal Scaling
```
         Load Balancer
              │
    ┌─────────┴─────────┐
    │         │         │
Server 1   Server 2   Server 3
    │         │         │
    └─────────┬─────────┘
              │
         Shared Redis
              │
         PostgreSQL
```

### Database Scaling
- Read replicas for search
- Connection pooling
- Query caching
- Sharding ready (by user_id)

### Caching Strategy
- Distributed Redis cluster
- Cache warming
- Graceful degradation

## Monitoring & Observability

### Metrics (Prometheus)
- Request rates and latencies
- Error rates
- Database performance
- Cache hit rates
- Business metrics

### Health Checks
- `/health` - Basic health
- `/health/ready` - Readiness probe
- `/health/live` - Liveness probe

### Logging
- Structured JSON logging
- Log levels (Debug, Info, Warn, Error)
- Request tracing with correlation IDs
- Error tracking

## Development Workflow

### Local Development
```bash
# Start dependencies
docker-compose up -d redis postgres

# Run with hot reload
make dev

# Run tests
make test

# Check coverage
make coverage
```

### Code Organization
- Clean architecture principles
- Dependency injection
- Interface-based design
- Testable components

### Testing Strategy
- Unit tests for business logic
- Integration tests for APIs
- Repository tests with test database
- Mocked external services

## Deployment Architecture

### Docker Deployment
```yaml
services:
  casgists:
    image: casgists:latest
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
    depends_on:
      - postgres
      - redis
  
  postgres:
    image: postgres:15
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7
    volumes:
      - redis_data:/data
```

### Production Considerations
- SSL/TLS termination
- Rate limiting
- DDoS protection
- Backup strategies
- Monitoring alerts

## Future Enhancements

### Planned Features
- GraphQL API
- Real-time collaboration
- Version control for gists
- Plugin system
- Mobile apps

### Architecture Evolution
- Microservices consideration
- Event sourcing for history
- CQRS for complex queries
- WebSocket support

---

This architecture provides a solid foundation for a scalable, maintainable, and secure gist management system while remaining simple enough for small deployments.