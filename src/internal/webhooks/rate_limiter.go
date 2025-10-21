package webhooks

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for webhooks
type RateLimiter struct {
	limiters map[uuid.UUID]*rate.Limiter
	limits   map[uuid.UUID]int
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[uuid.UUID]*rate.Limiter),
		limits:   make(map[uuid.UUID]int),
	}
}

// Allow checks if a webhook request should be allowed
func (rl *RateLimiter) Allow(webhookID uuid.UUID) bool {
	rl.mu.RLock()
	limiter, exists := rl.limiters[webhookID]
	rl.mu.RUnlock()
	
	if !exists {
		// No rate limit set, allow by default
		return true
	}
	
	return limiter.Allow()
}

// SetLimit sets the rate limit for a webhook (requests per minute)
func (rl *RateLimiter) SetLimit(webhookID uuid.UUID, requestsPerMinute int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	if requestsPerMinute <= 0 {
		// Remove rate limiting
		delete(rl.limiters, webhookID)
		delete(rl.limits, webhookID)
		return
	}
	
	// Convert requests per minute to requests per second
	requestsPerSecond := float64(requestsPerMinute) / 60.0
	
	// Create new limiter with burst capacity equal to 10% of per-minute limit
	burst := requestsPerMinute / 10
	if burst < 1 {
		burst = 1
	}
	
	rl.limiters[webhookID] = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
	rl.limits[webhookID] = requestsPerMinute
}

// GetLimit gets the current rate limit for a webhook
func (rl *RateLimiter) GetLimit(webhookID uuid.UUID) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return rl.limits[webhookID]
}

// RemoveLimit removes rate limiting for a webhook
func (rl *RateLimiter) RemoveLimit(webhookID uuid.UUID) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	delete(rl.limiters, webhookID)
	delete(rl.limits, webhookID)
}

// GetRemainingRequests gets the number of requests remaining in the current window
func (rl *RateLimiter) GetRemainingRequests(webhookID uuid.UUID) int {
	rl.mu.RLock()
	limiter, exists := rl.limiters[webhookID]
	rl.mu.RUnlock()
	
	if !exists {
		return -1 // No limit set
	}
	
	// Calculate remaining tokens
	tokens := limiter.Tokens()
	return int(tokens)
}

// WaitN waits until n requests can be made
func (rl *RateLimiter) WaitN(webhookID uuid.UUID, n int) time.Duration {
	rl.mu.RLock()
	limiter, exists := rl.limiters[webhookID]
	rl.mu.RUnlock()
	
	if !exists {
		return 0 // No limit set
	}
	
	// Calculate wait time
	reservation := limiter.Reserve()
	if !reservation.OK() {
		return time.Hour // Very long wait time
	}
	
	delay := reservation.Delay()
	reservation.Cancel() // Cancel the reservation
	return delay
}