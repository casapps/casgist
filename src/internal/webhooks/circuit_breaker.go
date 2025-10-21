package webhooks

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"   // Normal operation
	CircuitBreakerOpen     CircuitBreakerState = "open"     // Failing, requests blocked
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open" // Testing if service recovered
)

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening
	RecoveryTimeout  time.Duration // Time to wait before trying half-open
	SuccessThreshold int           // Number of successes needed to close from half-open
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,                // Open after 5 failures
		RecoveryTimeout:  30 * time.Second, // Wait 30 seconds before retry
		SuccessThreshold: 3,                // Need 3 successes to close
	}
}

// CircuitBreakerInfo contains information about a circuit breaker
type CircuitBreakerInfo struct {
	State            CircuitBreakerState `json:"state"`
	Failures         int                 `json:"failures"`
	Successes        int                 `json:"successes"`
	LastFailureTime  time.Time           `json:"last_failure_time"`
	LastSuccessTime  time.Time           `json:"last_success_time"`
	NextRetryTime    time.Time           `json:"next_retry_time"`
	TotalRequests    int64               `json:"total_requests"`
	SuccessfulReqs   int64               `json:"successful_requests"`
	FailedRequests   int64               `json:"failed_requests"`
}

// CircuitBreaker manages circuit breakers for webhooks
type CircuitBreaker struct {
	breakers map[uuid.UUID]*CircuitBreakerInfo
	config   CircuitBreakerConfig
	mu       sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker manager
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		breakers: make(map[uuid.UUID]*CircuitBreakerInfo),
		config:   DefaultCircuitBreakerConfig(),
	}
}

// NewCircuitBreakerWithConfig creates a new circuit breaker with custom config
func NewCircuitBreakerWithConfig(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		breakers: make(map[uuid.UUID]*CircuitBreakerInfo),
		config:   config,
	}
}

// Allow checks if a request should be allowed through the circuit breaker
func (cb *CircuitBreaker) Allow(webhookID uuid.UUID) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	breaker.TotalRequests++
	
	now := time.Now()
	
	switch breaker.State {
	case CircuitBreakerClosed:
		return true // Allow all requests
		
	case CircuitBreakerOpen:
		// Check if recovery timeout has passed
		if now.After(breaker.NextRetryTime) {
			breaker.State = CircuitBreakerHalfOpen
			breaker.Successes = 0
			return true // Allow one test request
		}
		return false // Still in open state
		
	case CircuitBreakerHalfOpen:
		return true // Allow requests to test recovery
		
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess(webhookID uuid.UUID) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	breaker.LastSuccessTime = time.Now()
	breaker.SuccessfulReqs++
	
	switch breaker.State {
	case CircuitBreakerClosed:
		// Reset failure count on success
		breaker.Failures = 0
		
	case CircuitBreakerHalfOpen:
		breaker.Successes++
		if breaker.Successes >= cb.config.SuccessThreshold {
			// Enough successes, close the circuit
			breaker.State = CircuitBreakerClosed
			breaker.Failures = 0
			breaker.Successes = 0
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure(webhookID uuid.UUID) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	now := time.Now()
	breaker.LastFailureTime = now
	breaker.Failures++
	breaker.FailedRequests++
	
	switch breaker.State {
	case CircuitBreakerClosed:
		if breaker.Failures >= cb.config.FailureThreshold {
			// Too many failures, open the circuit
			breaker.State = CircuitBreakerOpen
			breaker.NextRetryTime = now.Add(cb.config.RecoveryTimeout)
		}
		
	case CircuitBreakerHalfOpen:
		// Any failure in half-open state reopens the circuit
		breaker.State = CircuitBreakerOpen
		breaker.NextRetryTime = now.Add(cb.config.RecoveryTimeout)
		breaker.Successes = 0
	}
}

// GetState returns the current state of a circuit breaker
func (cb *CircuitBreaker) GetState(webhookID uuid.UUID) CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	
	// Check if we should transition from open to half-open
	if breaker.State == CircuitBreakerOpen && time.Now().After(breaker.NextRetryTime) {
		return CircuitBreakerHalfOpen
	}
	
	return breaker.State
}

// GetInfo returns detailed information about a circuit breaker
func (cb *CircuitBreaker) GetInfo(webhookID uuid.UUID) *CircuitBreakerInfo {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	
	// Create a copy to avoid race conditions
	info := *breaker
	
	// Update state if necessary
	if info.State == CircuitBreakerOpen && time.Now().After(info.NextRetryTime) {
		info.State = CircuitBreakerHalfOpen
	}
	
	return &info
}

// Reset manually resets a circuit breaker to closed state
func (cb *CircuitBreaker) Reset(webhookID uuid.UUID) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	breaker := cb.getOrCreateBreaker(webhookID)
	breaker.State = CircuitBreakerClosed
	breaker.Failures = 0
	breaker.Successes = 0
	breaker.NextRetryTime = time.Time{}
}

// Remove removes a circuit breaker
func (cb *CircuitBreaker) Remove(webhookID uuid.UUID) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	delete(cb.breakers, webhookID)
}

// GetAllStates returns the state of all circuit breakers
func (cb *CircuitBreaker) GetAllStates() map[uuid.UUID]CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	states := make(map[uuid.UUID]CircuitBreakerState)
	now := time.Now()
	
	for id, breaker := range cb.breakers {
		state := breaker.State
		if state == CircuitBreakerOpen && now.After(breaker.NextRetryTime) {
			state = CircuitBreakerHalfOpen
		}
		states[id] = state
	}
	
	return states
}

// getOrCreateBreaker gets or creates a circuit breaker for a webhook
func (cb *CircuitBreaker) getOrCreateBreaker(webhookID uuid.UUID) *CircuitBreakerInfo {
	breaker, exists := cb.breakers[webhookID]
	if !exists {
		breaker = &CircuitBreakerInfo{
			State: CircuitBreakerClosed,
		}
		cb.breakers[webhookID] = breaker
	}
	return breaker
}

// UpdateConfig updates the circuit breaker configuration
func (cb *CircuitBreaker) UpdateConfig(config CircuitBreakerConfig) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.config = config
}

// GetConfig returns the current configuration
func (cb *CircuitBreaker) GetConfig() CircuitBreakerConfig {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return cb.config
}