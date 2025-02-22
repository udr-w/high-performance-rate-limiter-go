package limiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// RedisLimiter implements a distributed rate limiter using Redis
type RedisLimiter struct {
	client   *redis.Client
	key      string
	limit    int
	duration time.Duration
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(client *redis.Client, key string, limit int, duration time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client:   client,
		key:      key,
		limit:    limit,
		duration: duration,
	}
}

// Allow checks if the request is allowed under the rate limit
func (rl *RedisLimiter) Allow() bool {
	count, err := rl.client.Incr(ctx, rl.key).Result()
	if err != nil {
		return false
	}

	if count == 1 {
		rl.client.Expire(ctx, rl.key, rl.duration)
	}

	return count <= int64(rl.limit)
}
