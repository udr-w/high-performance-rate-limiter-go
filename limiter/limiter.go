package limiter

import (
	"sync"
	"time"
)

// TokenBucket is a simple rate limiter using the token bucket algorithm
type TokenBucket struct {
	rate       int        // tokens added per second
	capacity   int        // maximum tokens allowed
	tokens     int        // current token count
	lastRefill time.Time  // last refill timestamp
	mutex      sync.Mutex // concurrency safety
}

// NewTokenBucket creates a new Token Bucket rate limiter
func NewTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if available
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	duration := now.Sub(tb.lastRefill).Seconds()
	toAdd := int(duration * float64(tb.rate))
	if toAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+toAdd)
		tb.lastRefill = now
	}

	// Check if we can allow the request
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
