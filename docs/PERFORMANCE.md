# Performance Optimization Guide

## Overview

CasGists includes comprehensive performance optimizations to ensure fast response times and efficient resource usage even under high load.

## Database Optimizations

### Connection Pooling
- Automatic connection pool sizing based on CPU cores
- Configurable max open/idle connections
- Connection lifetime and idle time management

```yaml
database:
  max_open_connections: 25
  max_idle_connections: 10
  connection_max_lifetime: 5m
  connection_max_idle_time: 90s
```

### Query Optimizations
- Selective field loading with GORM's `Select()`
- Efficient preloading strategies
- Optimized indexes for common queries
- Batch operations for bulk updates/inserts

### Indexes
The following performance indexes are automatically created:

- Composite indexes for user+visibility queries
- Full-text search optimization indexes
- Time-based sorting indexes
- Foreign key relationship indexes

## Caching Strategy

### Cache Layers
1. **Redis Cache** (if enabled) - Distributed caching
2. **In-Memory Cache** - Local fallback cache
3. **Query Result Cache** - Database query caching

### Cache Durations
- User data: 5 minutes
- Gist data: 10 minutes
- List queries: 2 minutes
- Search results: 1 minute
- User statistics: 15 minutes
- Trending data: 30 minutes

### Cache Invalidation
- Automatic invalidation on data updates
- Cascading invalidation for related data
- Background cache warming for popular content

## HTTP Performance

### Compression
- Gzip compression for text responses
- Configurable compression levels
- Automatic skip for pre-compressed content

```yaml
performance:
  compression_level: 5
  compression_min_size: 1024
```

### Response Optimization
- ETags for conditional requests
- Cache-Control headers for static assets
- Response buffering for small responses
- Content-Length headers when possible

### Static Asset Caching
- 1 year cache for versioned assets
- 5 minute cache for HTML pages
- No-cache for API responses

## Query Optimization Features

### Batch Processing
```go
// Process records in batches
processor := performance.NewBatchProcessor(db, 100)
err := processor.ProcessInBatches(ctx, &models.Gist{}, func(tx *gorm.DB, batch interface{}) error {
    // Process batch
    return nil
})
```

### Optimized Queries
```go
// Use optimized query builder
optimizer := performance.NewQueryOptimizations(db)
query := optimizer.OptimizedGistList(performance.GistFilters{
    UserID:       &userID,
    Visibility:   "public",
    IncludeFiles: true,
    OrderBy:      "stars",
})
```

### Search Performance
- Database-specific full-text search
- PostgreSQL: `ts_vector` and `ts_query`
- MySQL: `MATCH AGAINST`
- SQLite: Optimized LIKE queries

## Middleware Stack

### Performance Middleware Order
1. Request ID generation
2. Compression
3. Cache control headers
4. Response buffering
5. Security headers
6. Rate limiting
7. Request processing

## Configuration

### Recommended Production Settings

```yaml
# Database
database:
  type: postgresql
  max_open_connections: 50
  max_idle_connections: 20
  connection_max_lifetime: 5m

# Redis
redis:
  enabled: true
  pool_size: 100
  min_idle_conns: 10

# Performance
performance:
  compression_level: 6
  compression_min_size: 1024
  cache_warming_enabled: true
  
# Server
server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
```

## Monitoring Performance

### Metrics Available
- Request duration histogram
- Database query count
- Cache hit/miss rates
- Active connection count
- Memory usage
- Goroutine count

### Performance Endpoints
- `/metrics` - Prometheus metrics
- `/api/v1/health` - Health check with timing
- `/debug/pprof/` - Go profiling (dev mode only)

## Best Practices

### For Developers
1. Use pagination for list endpoints
2. Implement selective field loading
3. Cache computed values
4. Use batch operations for bulk changes
5. Optimize N+1 queries with preloading

### For Operators
1. Enable Redis for distributed caching
2. Configure appropriate connection pools
3. Monitor slow query logs
4. Set up CDN for static assets
5. Use reverse proxy caching

## Benchmarks

### Expected Performance
- Gist list: < 50ms
- Single gist: < 30ms
- Search: < 100ms
- Static assets: < 10ms
- API auth: < 20ms

### Load Testing
```bash
# Simple load test
ab -n 10000 -c 100 http://localhost:3000/api/v1/gists

# Authenticated requests
ab -n 10000 -c 100 -H "Authorization: Bearer TOKEN" http://localhost:3000/api/v1/gists
```

## Troubleshooting

### Slow Queries
1. Enable query logging
2. Check EXPLAIN plans
3. Verify indexes exist
4. Monitor connection pool

### High Memory Usage
1. Check cache size limits
2. Monitor goroutine leaks
3. Profile memory allocations
4. Tune GC settings

### Poor Cache Performance
1. Verify Redis connectivity
2. Check cache key conflicts
3. Monitor eviction rates
4. Tune TTL values