package limiter_test

import (
	"testing"
	"time"

	"context"

	"github.com/go-redis/redis/v8"
	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

var ctx = context.Background()

func TestRedisLimiter(t *testing.T) {
	// Set up Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Make sure Redis is running on this port
	})
	defer client.Close()

	// Flush DB to ensure clean state
	client.FlushDB(ctx)

	// Create a Redis-based rate limiter (allow max 5 requests per 2 seconds)
	rl := limiter.NewRedisLimiter(client, "test_key", 5, 2*time.Second)

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// 6th request should be denied
	if rl.Allow() {
		t.Errorf("Request 6 should have been denied due to rate limit")
	}

	// Wait for the rate limit to reset
	time.Sleep(2 * time.Second)

	// Now requests should be allowed again
	if !rl.Allow() {
		t.Errorf("Request after reset should have been allowed")
	}
}
