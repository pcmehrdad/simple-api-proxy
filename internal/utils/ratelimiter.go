package utils

import (
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiting mechanism
type RateLimiter struct {
	limit     int
	window    time.Duration
	requests  int
	lastReset time.Time
	mu        sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified requests per second limit
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		limit:     rps,
		window:    time.Second,
		requests:  0,
		lastReset: time.Now(),
	}
}

// Allow checks if a new request should be allowed based on the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastReset) >= rl.window {
		rl.requests = 0
		rl.lastReset = now
	}

	if rl.requests >= rl.limit {
		return false
	}

	rl.requests++
	return true
}

// IsLimited checks if the rate limiter is currently limited
func (rl *RateLimiter) IsLimited() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastReset)
	if elapsed >= rl.window {
		rl.requests = 0
		rl.lastReset = now
		return false
	}

	return rl.requests >= rl.limit
}

// RateLimiterManager manages multiple rate limiters
type RateLimiterManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewRateLimiterManager creates a new rate limiter manager
// This function is used by the Client in the api package
func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]*RateLimiter),
	}
}

// GetLimiter returns a rate limiter for the specified key and rate limit
func (rm *RateLimiterManager) GetLimiter(key string, rps int) *RateLimiter {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	limiter, exists := rm.limiters[key]
	if !exists {
		limiter = NewRateLimiter(rps)
		rm.limiters[key] = limiter
	}

	return limiter
}

// GetAvailableKey returns an available API key and its rate limit
func (rm *RateLimiterManager) GetAvailableKey(keys map[string]int) (string, int) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Try to find a key that's not limited
	for key, limit := range keys {
		limiter, exists := rm.limiters[key]
		if !exists {
			return key, limit
		}

		// Explicitly call the method using the RateLimiter pointer
		if limiter != nil && !limiter.IsLimited() {
			return key, limit
		}
	}

	// If all are limited, wait a small amount of time and try again
	time.Sleep(50 * time.Millisecond)
	for key, limit := range keys {
		limiter, exists := rm.limiters[key]
		if !exists {
			return key, limit
		}

		// Explicitly call the method using the RateLimiter pointer
		if limiter != nil && !limiter.IsLimited() {
			return key, limit
		}
	}

	return "", 0
}
