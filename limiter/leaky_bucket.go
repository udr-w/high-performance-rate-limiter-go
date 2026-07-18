package limiter

import (
	"sync"
	"time"
)

// LeakyBucket is a thread-safe limiter that drains queued requests at a fixed
// rate. It is useful when callers want a bounded backlog rather than bursty
// token-bucket behavior.
type LeakyBucket struct {
	rate     float64
	capacity float64
	water    float64
	lastLeak time.Time
	clock    Clock
	mutex    sync.Mutex
}

// NewLeakyBucket creates a leaky bucket with rate requests drained per second
// and capacity as the maximum queued request count.
//
// Invalid inputs create a closed limiter. Use NewLeakyBucketWithOptions when
// callers need validation errors.
func NewLeakyBucket(rate, capacity int) *LeakyBucket {
	lb, err := NewLeakyBucketWithOptions(float64(rate), capacity)
	if err != nil {
		return &LeakyBucket{
			clock:    realClock{},
			lastLeak: time.Now(),
		}
	}
	return lb
}

// NewLeakyBucketWithOptions creates a leaky bucket with validation and optional
// configuration.
func NewLeakyBucketWithOptions(rate float64, capacity int, opts ...Option) (*LeakyBucket, error) {
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

	return &LeakyBucket{
		rate:     rate,
		capacity: float64(capacity),
		lastLeak: cfg.clock.Now(),
		clock:    cfg.clock,
	}, nil
}

// Allow reports whether one request can enter the bucket.
func (lb *LeakyBucket) Allow() bool {
	return lb.AllowN(1)
}

// AllowN reports whether n requests can enter the bucket. A non-positive n is
// always allowed and does not mutate the bucket.
func (lb *LeakyBucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.leakLocked()

	requested := float64(n)
	if lb.water+requested <= lb.capacity {
		lb.water += requested
		return true
	}
	return false
}

// Queued returns the current whole number of queued requests.
func (lb *LeakyBucket) Queued() int {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.leakLocked()
	return int(lb.water)
}

func (lb *LeakyBucket) leakLocked() {
	if lb.clock == nil {
		lb.clock = realClock{}
	}

	now := lb.clock.Now()
	elapsed := now.Sub(lb.lastLeak)
	if elapsed <= 0 {
		return
	}

	lb.water = maxFloat(0, lb.water-elapsed.Seconds()*lb.rate)
	lb.lastLeak = now
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
