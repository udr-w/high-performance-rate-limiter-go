package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

type fakeClock struct {
	now time.Time
}

func (fc *fakeClock) Now() time.Time {
	return fc.now
}

func (fc *fakeClock) Advance(d time.Duration) {
	fc.now = fc.now.Add(d)
}

func TestTokenBucket(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	tb, err := limiter.NewTokenBucketWithOptions(5, 10, limiter.WithClock(clock))
	if err != nil {
		t.Fatalf("NewTokenBucketWithOptions() error = %v", err)
	}

	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true, got false")
	}

	for i := 0; i < 9; i++ {
		tb.Allow()
	}

	if tb.Allow() {
		t.Errorf("Expected Allow() to return false after consuming all tokens")
	}

	clock.Advance(time.Second)

	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true after refill, got false")
	}
}

func TestTokenBucketPreservesPartialRefill(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	tb, err := limiter.NewTokenBucketWithOptions(2, 2, limiter.WithClock(clock))
	if err != nil {
		t.Fatalf("NewTokenBucketWithOptions() error = %v", err)
	}

	if !tb.AllowN(2) {
		t.Fatal("expected initial burst to be allowed")
	}

	clock.Advance(250 * time.Millisecond)
	if tb.Allow() {
		t.Fatal("expected less than one refilled token to deny")
	}

	clock.Advance(250 * time.Millisecond)
	if !tb.Allow() {
		t.Fatal("expected accumulated partial refill to allow")
	}
}

func TestTokenBucketValidation(t *testing.T) {
	if _, err := limiter.NewTokenBucketWithOptions(0, 1); err != limiter.ErrInvalidRate {
		t.Fatalf("expected ErrInvalidRate, got %v", err)
	}
	if _, err := limiter.NewTokenBucketWithOptions(1, 0); err != limiter.ErrInvalidCapacity {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}
}

func TestTokenBucketUnderLoad(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	tb, err := limiter.NewTokenBucketWithOptions(10, 20, limiter.WithClock(clock))
	if err != nil {
		t.Fatalf("NewTokenBucketWithOptions() error = %v", err)
	}

	var wg sync.WaitGroup
	var allowed atomic.Int64
	var denied atomic.Int64

	numRequests := 50
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed.Add(1)
			} else {
				denied.Add(1)
			}
		}()
	}

	wg.Wait()

	t.Logf("Allowed requests: %d, Denied requests: %d", allowed.Load(), denied.Load())

	if allowed.Load() > 20 {
		t.Errorf("More requests allowed than max capacity")
	}

	clock.Advance(time.Second)

	allowed.Store(0)
	denied.Store(0)
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed.Add(1)
			} else {
				denied.Add(1)
			}
		}()
	}

	wg.Wait()
	t.Logf("After refill - Allowed requests: %d, Denied requests: %d", allowed.Load(), denied.Load())

	if allowed.Load() > 10 {
		t.Errorf("More requests allowed than refill rate")
	}
}

func TestTokenBucketWithChannels(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 20) // 10 tokens per second, max 20 tokens

	numRequests := 50
	requests := make(chan bool, numRequests)
	results := make(chan bool, numRequests)
	var wg sync.WaitGroup

	// Spawn goroutines to send requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			requests <- tb.Allow()
		}()
	}

	// Collect results asynchronously
	go func() {
		for result := range requests {
			results <- result
		}
		close(results)
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(requests)

	// Count allowed and denied requests
	allowed := 0
	denied := 0
	for result := range results {
		if result {
			allowed++
		} else {
			denied++
		}
	}

	t.Logf("Using Channels - Allowed: %d, Denied: %d", allowed, denied)

	if allowed > 20 {
		t.Errorf("More requests allowed than max capacity")
	}
}

func TestLeakyBucket(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	lb, err := limiter.NewLeakyBucketWithOptions(2, 2, limiter.WithClock(clock))
	if err != nil {
		t.Fatalf("NewLeakyBucketWithOptions() error = %v", err)
	}

	if !lb.AllowN(2) {
		t.Fatal("expected initial bucket capacity to allow")
	}
	if lb.Allow() {
		t.Fatal("expected full leaky bucket to deny")
	}

	clock.Advance(500 * time.Millisecond)
	if !lb.Allow() {
		t.Fatal("expected leaked capacity to allow")
	}
}

func TestLeakyBucketValidation(t *testing.T) {
	if _, err := limiter.NewLeakyBucketWithOptions(0, 1); err != limiter.ErrInvalidRate {
		t.Fatalf("expected ErrInvalidRate, got %v", err)
	}
	if _, err := limiter.NewLeakyBucketWithOptions(1, 0); err != limiter.ErrInvalidCapacity {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}
}

func BenchmarkTokenBucket(b *testing.B) {
	tb := limiter.NewTokenBucket(100, 200) // 100 tokens per second, max 200 tokens

	b.ResetTimer() // Ignore setup time

	for i := 0; i < b.N; i++ {
		tb.Allow()
	}
}
