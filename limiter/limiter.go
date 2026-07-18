package limiter

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidRate     = errors.New("rate must be greater than zero")
	ErrInvalidCapacity = errors.New("capacity must be greater than zero")
	ErrInvalidDuration = errors.New("duration must be greater than zero")
)

// Limiter is the common interface implemented by local limiters.
type Limiter interface {
	Allow() bool
}

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

type tokenBucketConfig struct {
	clock Clock
}

// Option configures local limiters.
type Option func(*tokenBucketConfig)

// WithClock overrides the time source used by a local limiter.
// It is primarily useful for deterministic tests and simulations.
func WithClock(c Clock) Option {
	return func(cfg *tokenBucketConfig) {
		if c != nil {
			cfg.clock = c
		}
	}
}

// TokenBucket is a thread-safe rate limiter using the token bucket algorithm.
type TokenBucket struct {
	rate       float64
	capacity   float64
	tokens     float64
	lastRefill time.Time
	clock      Clock
	mutex      sync.Mutex
}

// NewTokenBucket creates a token bucket with rate tokens added per second and
// capacity as the maximum burst size.
//
// Invalid inputs create a closed limiter. Use NewTokenBucketWithOptions when
// callers need validation errors.
func NewTokenBucket(rate, capacity int) *TokenBucket {
	tb, err := NewTokenBucketWithOptions(float64(rate), capacity)
	if err != nil {
		return &TokenBucket{
			clock:      realClock{},
			lastRefill: time.Now(),
		}
	}
	return tb
}

// NewTokenBucketWithOptions creates a token bucket with validation and optional
// configuration.
func NewTokenBucketWithOptions(rate float64, capacity int, opts ...Option) (*TokenBucket, error) {
	if rate <= 0 {
		return nil, ErrInvalidRate
	}
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	cfg := tokenBucketConfig{clock: realClock{}}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return &TokenBucket{
		rate:       rate,
		capacity:   float64(capacity),
		tokens:     float64(capacity),
		lastRefill: cfg.clock.Now(),
		clock:      cfg.clock,
	}, nil
}

// Allow reports whether one request is allowed and consumes one token when it is.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN reports whether n requests are allowed and consumes n tokens when they
// are. A non-positive n is always allowed and does not mutate the bucket.
func (tb *TokenBucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refillLocked()

	requested := float64(n)
	if tb.tokens >= requested {
		tb.tokens -= requested
		return true
	}
	return false
}

// Available returns the current number of whole tokens without consuming them.
func (tb *TokenBucket) Available() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refillLocked()
	return int(tb.tokens)
}

func (tb *TokenBucket) refillLocked() {
	if tb.clock == nil {
		tb.clock = realClock{}
	}

	now := tb.clock.Now()
	elapsed := now.Sub(tb.lastRefill)
	if elapsed <= 0 {
		return
	}

	tb.tokens = minFloat(tb.capacity, tb.tokens+elapsed.Seconds()*tb.rate)
	tb.lastRefill = now
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
