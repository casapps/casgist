package performance

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Optimizer manages performance optimizations
type Optimizer struct {
	db              *gorm.DB
	cfg             *viper.Viper
	queryCache      *QueryCache
	connectionPool  *ConnectionPool
	resourceLimiter *ResourceLimiter
	mu              sync.RWMutex
}

// NewOptimizer creates a new performance optimizer
func NewOptimizer(db *gorm.DB, cfg *viper.Viper) *Optimizer {
	return &Optimizer{
		db:              db,
		cfg:             cfg,
		queryCache:      NewQueryCache(cfg),
		connectionPool:  NewConnectionPool(cfg),
		resourceLimiter: NewResourceLimiter(cfg),
	}
}

// OptimizeDatabase applies database optimizations
func (o *Optimizer) OptimizeDatabase() error {
	// Set connection pool settings
	sqlDB, err := o.db.DB()
	if err != nil {
		return err
	}

	// Configure connection pool
	maxOpenConns := o.cfg.GetInt("database.max_open_connections")
	if maxOpenConns == 0 {
		maxOpenConns = runtime.NumCPU() * 2
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := o.cfg.GetInt("database.max_idle_connections")
	if maxIdleConns == 0 {
		maxIdleConns = runtime.NumCPU()
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	connMaxLifetime := o.cfg.GetDuration("database.connection_max_lifetime")
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	connMaxIdleTime := o.cfg.GetDuration("database.connection_max_idle_time")
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 10 * time.Minute
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Enable prepared statements globally
	o.db = o.db.Session(&gorm.Session{
		PrepareStmt: true,
		CreateBatchSize: 100,
	})

	// Create additional indexes for performance
	if err := o.createPerformanceIndexes(); err != nil {
		return err
	}

	return nil
}

// createPerformanceIndexes creates additional performance indexes
func (o *Optimizer) createPerformanceIndexes() error {
	indexes := []string{
		// Composite indexes for common queries
		"CREATE INDEX IF NOT EXISTS idx_gists_user_visibility_created ON gists(user_id, visibility, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_gists_visibility_stars_created ON gists(visibility, star_count DESC, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_gist_files_gist_language ON gist_files(gist_id, language)",
		
		// Full-text search indexes (for SQLite)
		"CREATE INDEX IF NOT EXISTS idx_gists_title_fts ON gists(title)",
		"CREATE INDEX IF NOT EXISTS idx_gists_description_fts ON gists(description)",
		"CREATE INDEX IF NOT EXISTS idx_gist_files_content_fts ON gist_files(content)",
		
		// Performance indexes
		"CREATE INDEX IF NOT EXISTS idx_sessions_token_expires ON sessions(token, expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_api_tokens_hash_user ON api_tokens(token_hash, user_id)",
	}

	for _, idx := range indexes {
		if err := o.db.Exec(idx).Error; err != nil {
			// Log but don't fail on index creation errors
			continue
		}
	}

	// Analyze tables for query optimizer
	o.db.Exec("ANALYZE")

	return nil
}

// QueryCache manages query result caching
type QueryCache struct {
	cache sync.Map
	ttl   time.Duration
}

// NewQueryCache creates a new query cache
func NewQueryCache(cfg *viper.Viper) *QueryCache {
	ttl := cfg.GetDuration("performance.query_cache_ttl")
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	qc := &QueryCache{
		ttl: ttl,
	}

	// Start cleanup goroutine
	go qc.cleanup()

	return qc
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	if val, ok := qc.cache.Load(key); ok {
		entry := val.(*cacheEntry)
		if time.Now().Before(entry.expiry) {
			return entry.data, true
		}
		qc.cache.Delete(key)
	}
	return nil, false
}

// Set stores a query result in cache
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.cache.Store(key, &cacheEntry{
		data:   data,
		expiry: time.Now().Add(qc.ttl),
	})
}

// cleanup removes expired entries periodically
func (qc *QueryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		qc.cache.Range(func(key, value interface{}) bool {
			entry := value.(*cacheEntry)
			if now.After(entry.expiry) {
				qc.cache.Delete(key)
			}
			return true
		})
	}
}

type cacheEntry struct {
	data   interface{}
	expiry time.Time
}

// ConnectionPool manages database connections
type ConnectionPool struct {
	maxSize     int
	minSize     int
	idleTimeout time.Duration
	mu          sync.RWMutex
	connections []*pooledConnection
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(cfg *viper.Viper) *ConnectionPool {
	return &ConnectionPool{
		maxSize:     cfg.GetInt("database.pool_size_max"),
		minSize:     cfg.GetInt("database.pool_size_min"),
		idleTimeout: cfg.GetDuration("database.pool_idle_timeout"),
		connections: make([]*pooledConnection, 0),
	}
}

type pooledConnection struct {
	conn     *gorm.DB
	lastUsed time.Time
	inUse    bool
}

// ResourceLimiter manages resource limits
type ResourceLimiter struct {
	maxCPU      float64
	maxMemory   uint64
	maxRequests int
	mu          sync.RWMutex
	requests    int
}

// NewResourceLimiter creates a new resource limiter
func NewResourceLimiter(cfg *viper.Viper) *ResourceLimiter {
	return &ResourceLimiter{
		maxCPU:      cfg.GetFloat64("performance.max_cpu_percent"),
		maxMemory:   cfg.GetUint64("performance.max_memory_mb") * 1024 * 1024,
		maxRequests: cfg.GetInt("performance.max_concurrent_requests"),
	}
}

// CheckLimits checks if resource limits are exceeded
func (rl *ResourceLimiter) CheckLimits() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Check concurrent requests
	if rl.maxRequests > 0 && rl.requests >= rl.maxRequests {
		return false
	}

	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if rl.maxMemory > 0 && m.Alloc > rl.maxMemory {
		return false
	}

	return true
}

// PerformanceMiddleware returns Echo middleware for performance optimization
func (o *Optimizer) PerformanceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check resource limits
			if !o.resourceLimiter.CheckLimits() {
				return echo.NewHTTPError(503, "Server busy")
			}

			// Track request
			o.resourceLimiter.mu.Lock()
			o.resourceLimiter.requests++
			o.resourceLimiter.mu.Unlock()

			defer func() {
				o.resourceLimiter.mu.Lock()
				o.resourceLimiter.requests--
				o.resourceLimiter.mu.Unlock()
			}()

			// Add performance headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Enable compression
			c.Response().Header().Set("Vary", "Accept-Encoding")

			// Set cache headers for static assets
			if isStaticAsset(c.Request().URL.Path) {
				c.Response().Header().Set("Cache-Control", "public, max-age=86400")
			}

			return next(c)
		}
	}
}

// isStaticAsset checks if path is a static asset
func isStaticAsset(path string) bool {
	staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf"}
	for _, ext := range staticExts {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// BatchProcessor processes operations in batches
type BatchProcessor struct {
	batchSize     int
	flushInterval time.Duration
	processor     func(context.Context, []interface{}) error
	items         []interface{}
	mu            sync.Mutex
	timer         *time.Timer
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, flushInterval time.Duration, processor func(context.Context, []interface{}) error) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		processor:     processor,
		items:         make([]interface{}, 0, batchSize),
	}
	bp.timer = time.AfterFunc(flushInterval, bp.flush)
	return bp
}

// Add adds an item to the batch
func (bp *BatchProcessor) Add(ctx context.Context, item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.items = append(bp.items, item)

	if len(bp.items) >= bp.batchSize {
		return bp.processLocked(ctx)
	}

	// Reset timer
	bp.timer.Reset(bp.flushInterval)
	return nil
}

// flush processes pending items
func (bp *BatchProcessor) flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.items) > 0 {
		ctx := context.Background()
		_ = bp.processLocked(ctx)
	}

	bp.timer.Reset(bp.flushInterval)
}

// processLocked processes items (must be called with lock held)
func (bp *BatchProcessor) processLocked(ctx context.Context) error {
	if len(bp.items) == 0 {
		return nil
	}

	items := bp.items
	bp.items = make([]interface{}, 0, bp.batchSize)

	err := bp.processor(ctx, items)
	if err != nil {
		// Re-add items on error
		bp.items = append(items, bp.items...)
		return err
	}

	return nil
}

// Close flushes remaining items and stops the processor
func (bp *BatchProcessor) Close(ctx context.Context) error {
	bp.timer.Stop()

	bp.mu.Lock()
	defer bp.mu.Unlock()

	return bp.processLocked(ctx)
}