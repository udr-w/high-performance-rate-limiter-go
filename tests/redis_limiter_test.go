package limiter_test

import (
	"context"
	"sync"
	"testing"
	"time"

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

func TestRedisLimiterUnderLoad(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()
	client.FlushDB(ctx)

	rl := limiter.NewRedisLimiter(client, "test_load", 10, 3*time.Second)

	var wg sync.WaitGroup
	allowed := 0
	denied := 0
	numRequests := 100
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			if rl.Allow() {
				allowed++
			} else {
				denied++
			}
		}()
	}

	wg.Wait()

	t.Logf("Under Load - Allowed: %d, Denied: %d", allowed, denied)
}

func TestRedisLimiterMultiInstance(t *testing.T) {
	client1 := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client1.Close()
	client1.FlushDB(ctx)

	client2 := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client2.Close()

	rl1 := limiter.NewRedisLimiter(client1, "test_multi", 5, 2*time.Second)
	rl2 := limiter.NewRedisLimiter(client2, "test_multi", 5, 2*time.Second)

	// First 5 requests should be allowed across instances
	for i := 0; i < 5; i++ {
		if !rl1.Allow() {
			t.Errorf("Request %d should have been allowed on instance 1", i+1)
		}
	}

	// Additional request from second instance should be denied
	if rl2.Allow() {
		t.Errorf("Request should have been denied due to shared rate limit across instances")
	}

	// Wait for rate limiter to reset
	time.Sleep(2 * time.Second)

	// Now requests should be allowed again
	if !rl1.Allow() {
		t.Errorf("Request after reset should have been allowed on instance 1")
	}
}
