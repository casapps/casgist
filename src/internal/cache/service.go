package cache

import (
	"context"
	"time"
)

// Service defines the cache service interface
type Service interface {
	// Get retrieves a value from cache
	Get(key string, dest interface{}) error
	
	// Set stores a value in cache
	Set(key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a value from cache
	Delete(key string) error
	
	// Exists checks if a key exists
	Exists(key string) (bool, error)
}

// MemoryCacheService implements Service using in-memory cache
type MemoryCacheService struct {
	*MemoryCache
}

// NewMemoryCacheService creates a new memory cache service
func NewMemoryCacheService() Service {
	return &MemoryCacheService{
		MemoryCache: NewMemoryCache(),
	}
}

// Get retrieves a value from cache
func (m *MemoryCacheService) Get(key string, dest interface{}) error {
	ctx := context.Background()
	return m.GetJSON(ctx, key, dest)
}

// Set stores a value in cache
func (m *MemoryCacheService) Set(key string, value interface{}, ttl time.Duration) error {
	ctx := context.Background()
	return m.SetJSON(ctx, key, value, ttl)
}

// Delete removes a value from cache
func (m *MemoryCacheService) Delete(key string) error {
	ctx := context.Background()
	return m.MemoryCache.Delete(ctx, key)
}

// Exists checks if a key exists
func (m *MemoryCacheService) Exists(key string) (bool, error) {
	ctx := context.Background()
	return m.MemoryCache.Exists(ctx, key)
}