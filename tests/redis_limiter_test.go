package limiter_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

var ctx = context.Background()

func redisClientOrSkip(tb testing.TB) *redis.Client {
	tb.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		DialTimeout:  250 * time.Millisecond,
		ReadTimeout:  250 * time.Millisecond,
		WriteTimeout: 250 * time.Millisecond,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		tb.Skipf("Redis integration test skipped; Redis is unavailable at localhost:6379: %v", err)
	}

	return client
}

func cleanupKey(tb testing.TB, client *redis.Client, key string) {
	tb.Helper()

	if err := client.Del(ctx, key).Err(); err != nil {
		tb.Fatalf("failed to clean Redis key %q: %v", key, err)
	}
}

func redisTestKey(tb testing.TB, namespace string) string {
	tb.Helper()

	return limiter.RedisKey(namespace, tb.Name()+":"+strconv.FormatInt(time.Now().UnixNano(), 10))
}

func TestRedisLimiter(t *testing.T) {
	client := redisClientOrSkip(t)
	defer client.Close()

	key := redisTestKey(t, "test")
	cleanupKey(t, client, key)
	defer cleanupKey(t, client, key)

	rl, err := limiter.NewRedisLimiterWithValidation(client, key, 5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewRedisLimiterWithValidation() error = %v", err)
	}

	for i := 0; i < 5; i++ {
		allowed, err := rl.AllowContext(ctx)
		if err != nil {
			t.Fatalf("AllowContext() error = %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	allowed, err := rl.AllowContext(ctx)
	if err != nil {
		t.Fatalf("AllowContext() error = %v", err)
	}
	if allowed {
		t.Errorf("Request 6 should have been denied due to rate limit")
	}

	time.Sleep(150 * time.Millisecond)

	allowed, err = rl.AllowContext(ctx)
	if err != nil {
		t.Fatalf("AllowContext() after reset error = %v", err)
	}
	if !allowed {
		t.Errorf("Request after reset should have been allowed")
	}
}

func TestRedisLimiterUnderLoad(t *testing.T) {
	client := redisClientOrSkip(t)
	defer client.Close()

	key := redisTestKey(t, "test")
	cleanupKey(t, client, key)
	defer cleanupKey(t, client, key)

	rl, err := limiter.NewRedisLimiterWithValidation(client, key, 10, time.Second)
	if err != nil {
		t.Fatalf("NewRedisLimiterWithValidation() error = %v", err)
	}

	var wg sync.WaitGroup
	var allowed atomic.Int64
	var denied atomic.Int64
	numRequests := 100
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			ok, err := rl.AllowContext(ctx)
			if err != nil {
				t.Errorf("AllowContext() error = %v", err)
				return
			}
			if ok {
				allowed.Add(1)
			} else {
				denied.Add(1)
			}
		}()
	}

	wg.Wait()

	t.Logf("Under Load - Allowed: %d, Denied: %d", allowed.Load(), denied.Load())

	if allowed.Load() > 10 {
		t.Errorf("More requests allowed than limit: got %d, want <= 10", allowed.Load())
	}
	if got := allowed.Load() + denied.Load(); got != int64(numRequests) {
		t.Errorf("Allowed + denied = %d, want %d", got, numRequests)
	}
}

func TestRedisLimiterMultiInstance(t *testing.T) {
	client1 := redisClientOrSkip(t)
	defer client1.Close()

	client2 := redisClientOrSkip(t)
	defer client2.Close()

	key := redisTestKey(t, "test")
	cleanupKey(t, client1, key)
	defer cleanupKey(t, client1, key)

	rl1, err := limiter.NewRedisLimiterWithValidation(client1, key, 5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewRedisLimiterWithValidation() error = %v", err)
	}
	rl2, err := limiter.NewRedisLimiterWithValidation(client2, key, 5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewRedisLimiterWithValidation() error = %v", err)
	}

	for i := 0; i < 5; i++ {
		allowed, err := rl1.AllowContext(ctx)
		if err != nil {
			t.Fatalf("AllowContext() error = %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should have been allowed on instance 1", i+1)
		}
	}

	allowed, err := rl2.AllowContext(ctx)
	if err != nil {
		t.Fatalf("AllowContext() error = %v", err)
	}
	if allowed {
		t.Errorf("Request should have been denied due to shared rate limit across instances")
	}

	time.Sleep(150 * time.Millisecond)

	allowed, err = rl1.AllowContext(ctx)
	if err != nil {
		t.Fatalf("AllowContext() after reset error = %v", err)
	}
	if !allowed {
		t.Errorf("Request after reset should have been allowed on instance 1")
	}
}

func TestRedisLimiterValidation(t *testing.T) {
	if _, err := limiter.NewRedisLimiterWithValidation(nil, "key", 1, time.Second); err != limiter.ErrInvalidRedisClient {
		t.Fatalf("expected ErrInvalidRedisClient, got %v", err)
	}

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	if _, err := limiter.NewRedisLimiterWithValidation(client, "", 1, time.Second); err != limiter.ErrInvalidRedisKey {
		t.Fatalf("expected ErrInvalidRedisKey, got %v", err)
	}
	if _, err := limiter.NewRedisLimiterWithValidation(client, "key", 0, time.Second); err != limiter.ErrInvalidCapacity {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}
	if _, err := limiter.NewRedisLimiterWithValidation(client, "key", 1, 0); err != limiter.ErrInvalidDuration {
		t.Fatalf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestRedisLimiterAllowContextReturnsErrors(t *testing.T) {
	rl := limiter.NewRedisLimiter(nil, "key", 1, time.Second)

	allowed, err := rl.AllowContext(ctx)
	if err != limiter.ErrInvalidRedisClient {
		t.Fatalf("expected ErrInvalidRedisClient, got allowed=%v err=%v", allowed, err)
	}
	if rl.Allow() {
		t.Fatal("compatibility Allow() should deny on Redis/configuration errors")
	}
}

func BenchmarkRedisLimiter(b *testing.B) {
	client := redisClientOrSkip(b)
	defer client.Close()

	key := redisTestKey(b, "benchmark")
	if err := client.Del(ctx, key).Err(); err != nil {
		b.Fatalf("failed to clean Redis key %q: %v", key, err)
	}
	defer client.Del(ctx, key)

	rl := limiter.NewRedisLimiter(client, key, b.N+100, time.Second)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rl.AllowContext(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
