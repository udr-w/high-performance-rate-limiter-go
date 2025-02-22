package limiter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

func TestTokenBucket(t *testing.T) {
	tb := limiter.NewTokenBucket(5, 10) // 5 tokens per second, max 10 tokens

	// Allow should pass initially (bucket starts full)
	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true, got false")
	}

	// Exhaust the tokens
	for i := 0; i < 9; i++ {
		tb.Allow()
	}

	// Now it should be empty
	if tb.Allow() {
		t.Errorf("Expected Allow() to return false after consuming all tokens")
	}

	// Wait 1 second to refill tokens
	time.Sleep(time.Second)

	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true after refill, got false")
	}
}

func TestTokenBucketUnderLoad(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 20) // 10 tokens per second, max 20 tokens

	var wg sync.WaitGroup
	allowed := 0
	denied := 0

	// Use a large number of goroutines to simulate concurrent requests
	numRequests := 50
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed++
			} else {
				denied++
			}
		}()
	}

	wg.Wait()

	t.Logf("Allowed requests: %d, Denied requests: %d", allowed, denied)

	if allowed > 20 {
		t.Errorf("More requests allowed than max capacity")
	}

	// Wait to allow tokens to refill
	time.Sleep(time.Second)

	// Run another batch of requests after refill
	allowed = 0
	denied = 0
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed++
			} else {
				denied++
			}
		}()
	}

	wg.Wait()
	t.Logf("After refill - Allowed requests: %d, Denied requests: %d", allowed, denied)
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

func BenchmarkTokenBucket(b *testing.B) {
	tb := limiter.NewTokenBucket(100, 200) // 100 tokens per second, max 200 tokens

	b.ResetTimer() // Ignore setup time

	for i := 0; i < b.N; i++ {
		tb.Allow()
	}
}
