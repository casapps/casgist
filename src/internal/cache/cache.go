package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Close() error
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// MemoryCache implements Cache interface using in-memory storage (fallback)
type MemoryCache struct {
	data map[string]cacheItem
	ttls map[string]time.Time
	mu   sync.RWMutex
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

// CacheManager manages cache instances with fallback
type CacheManager struct {
	primary   Cache
	fallback  Cache
	enabled   bool
	keyPrefix string
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cfg *viper.Viper) *CacheManager {
	manager := &CacheManager{
		enabled:   cfg.GetBool("cache.enabled"),
		keyPrefix: cfg.GetString("cache.key_prefix"),
	}

	if manager.keyPrefix == "" {
		manager.keyPrefix = "casgists:"
	}

	// Try to connect to Redis
	if manager.enabled && cfg.GetBool("redis.enabled") {
		redisCache, err := NewRedisCache(cfg)
		if err == nil {
			manager.primary = redisCache
		}
	}

	// Always have memory cache as fallback
	manager.fallback = NewMemoryCache()

	return manager
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg *viper.Viper) (*RedisCache, error) {
	addr := cfg.GetString("redis.addr")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := cfg.GetString("redis.password")
	db := cfg.GetInt("redis.db")
	prefix := cfg.GetString("cache.key_prefix")
	if prefix == "" {
		prefix = "casgists:"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		PoolSize:     10,
		PoolTimeout:  time.Second * 4,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: prefix,
	}, nil
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheItem),
		ttls: make(map[string]time.Time),
	}
}

// CacheManager methods

func (cm *CacheManager) key(key string) string {
	return cm.keyPrefix + key
}

func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	if !cm.enabled {
		return "", fmt.Errorf("cache not enabled")
	}

	fullKey := cm.key(key)

	// Try primary cache first
	if cm.primary != nil {
		value, err := cm.primary.Get(ctx, fullKey)
		if err == nil {
			return value, nil
		}
	}

	// Fallback to memory cache
	return cm.fallback.Get(ctx, fullKey)
}

func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cm.enabled {
		return nil
	}

	fullKey := cm.key(key)

	// Set in primary cache
	if cm.primary != nil {
		if err := cm.primary.Set(ctx, fullKey, value, ttl); err == nil {
			return nil
		}
	}

	// Fallback to memory cache
	return cm.fallback.Set(ctx, fullKey, value, ttl)
}

func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	if !cm.enabled {
		return nil
	}

	fullKey := cm.key(key)

	// Delete from both caches
	if cm.primary != nil {
		cm.primary.Delete(ctx, fullKey)
	}
	cm.fallback.Delete(ctx, fullKey)

	return nil
}

func (cm *CacheManager) DeletePattern(ctx context.Context, pattern string) error {
	if !cm.enabled {
		return nil
	}

	fullPattern := cm.key(pattern)

	// Delete from both caches
	if cm.primary != nil {
		cm.primary.DeletePattern(ctx, fullPattern)
	}
	cm.fallback.DeletePattern(ctx, fullPattern)

	return nil
}

func (cm *CacheManager) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if !cm.enabled {
		return fmt.Errorf("cache not enabled")
	}

	value, err := cm.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func (cm *CacheManager) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cm.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cm.Set(ctx, key, string(data), ttl)
}

func (cm *CacheManager) Close() error {
	if cm.primary != nil {
		cm.primary.Close()
	}
	if cm.fallback != nil {
		cm.fallback.Close()
	}
	return nil
}

// RedisCache methods

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return rc.client.Set(ctx, key, value, ttl).Err()
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return rc.client.Del(ctx, keys...).Err()
	}

	return nil
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (rc *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

func (rc *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return rc.client.Expire(ctx, key, ttl).Err()
}

func (rc *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := rc.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (rc *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rc.Set(ctx, key, string(data), ttl)
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// MemoryCache methods

func (mc *MemoryCache) cleanExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	now := time.Now()
	for key, expiry := range mc.ttls {
		if now.After(expiry) {
			delete(mc.data, key)
			delete(mc.ttls, key)
		}
	}
}

func (mc *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	mc.cleanExpired()
	
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	item, exists := mc.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if time.Now().After(item.expiresAt) {
		return "", fmt.Errorf("key expired")
	}

	return item.value, nil
}

func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.cleanExpired()

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		strValue = string(data)
	}

	expiresAt := time.Now().Add(ttl)
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.data[key] = cacheItem{
		value:     strValue,
		expiresAt: expiresAt,
	}
	mc.ttls[key] = expiresAt

	return nil
}

func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	delete(mc.data, key)
	delete(mc.ttls, key)
	return nil
}

func (mc *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Simple pattern matching for memory cache
	for key := range mc.data {
		// Basic wildcard matching
		if matchPattern(pattern, key) {
			delete(mc.data, key)
			delete(mc.ttls, key)
		}
	}
	return nil
}

func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.cleanExpired()
	
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	_, exists := mc.data[key]
	return exists, nil
}

func (mc *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	// Not implemented for memory cache
	return 0, fmt.Errorf("increment not supported in memory cache")
}

func (mc *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if item, exists := mc.data[key]; exists {
		item.expiresAt = time.Now().Add(ttl)
		mc.data[key] = item
		mc.ttls[key] = item.expiresAt
	}
	return nil
}

func (mc *MemoryCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := mc.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (mc *MemoryCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return mc.Set(ctx, key, string(data), ttl)
}

func (mc *MemoryCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.data = make(map[string]cacheItem)
	mc.ttls = make(map[string]time.Time)
	return nil
}

// Helper functions

func matchPattern(pattern, str string) bool {
	// Very basic wildcard matching for memory cache
	if pattern == "*" {
		return true
	}
	
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(str, prefix)
	}
	
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(str, suffix)
	}
	
	return pattern == str
}

// Cache keys constants
const (
	CacheKeyUser       = "user:%s"
	CacheKeyGist       = "gist:%s"
	CacheKeyGistList   = "gist:list:%s"
	CacheKeyUserGists  = "user:%s:gists"
	CacheKeyUserStats  = "user:%s:stats"
	CacheKeySearch     = "search:%s"
	CacheKeyTrending   = "trending"
	CacheKeyPopular    = "popular"
	CacheKeyConfig     = "config"
)

// Cache TTL constants
const (
	TTLShort   = 5 * time.Minute
	TTLMedium  = 30 * time.Minute
	TTLLong    = 2 * time.Hour
	TTLVeryLong = 24 * time.Hour
)

// Helper functions for common cache operations

func UserKey(userID string) string {
	return fmt.Sprintf(CacheKeyUser, userID)
}

func GistKey(gistID string) string {
	return fmt.Sprintf(CacheKeyGist, gistID)
}

func UserGistsKey(userID string) string {
	return fmt.Sprintf(CacheKeyUserGists, userID)
}

func UserStatsKey(userID string) string {
	return fmt.Sprintf(CacheKeyUserStats, userID)
}

func SearchKey(query string) string {
	return fmt.Sprintf(CacheKeySearch, query)
}