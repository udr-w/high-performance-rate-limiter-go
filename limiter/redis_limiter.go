package limiter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrInvalidRedisClient = errors.New("redis client is nil")
var ErrInvalidRedisKey = errors.New("redis key must not be empty")

var fixedWindowScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
local ttl = redis.call("PTTL", KEYS[1])
if current == 1 or ttl < 0 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// RedisLimiter implements a distributed fixed-window rate limiter using Redis.
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

// RedisKey builds a namespaced Redis key for untrusted identifiers. The
// identifier is hashed to keep keys bounded and avoid exposing raw user input.
func RedisKey(namespace, identifier string) string {
	sum := sha256.Sum256([]byte(identifier))
	return "rate_limiter:" + namespace + ":" + hex.EncodeToString(sum[:])
}

// NewRedisLimiterWithValidation creates a Redis limiter and validates inputs.
func NewRedisLimiterWithValidation(client *redis.Client, key string, limit int, duration time.Duration) (*RedisLimiter, error) {
	if client == nil {
		return nil, ErrInvalidRedisClient
	}
	if key == "" {
		return nil, ErrInvalidRedisKey
	}
	if limit <= 0 {
		return nil, ErrInvalidCapacity
	}
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	return NewRedisLimiter(client, key, limit, duration), nil
}

// Allow checks if the request is allowed under the rate limit. Redis errors are
// treated as denied for compatibility with the original API. Use AllowContext
// when callers need the underlying error.
func (rl *RedisLimiter) Allow() bool {
	allowed, err := rl.AllowContext(context.Background())
	if err != nil {
		return false
	}
	return allowed
}

// AllowContext checks if the request is allowed under the rate limit.
func (rl *RedisLimiter) AllowContext(ctx context.Context) (bool, error) {
	count, err := rl.IncrementContext(ctx)
	if err != nil {
		return false, err
	}

	return count <= int64(rl.limit), nil
}

// IncrementContext increments the current window and returns its request count.
func (rl *RedisLimiter) IncrementContext(ctx context.Context) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := rl.validate(); err != nil {
		return 0, err
	}

	ttlMillis := rl.duration.Milliseconds()
	if ttlMillis <= 0 {
		ttlMillis = 1
	}

	result, err := fixedWindowScript.Run(ctx, rl.client, []string{rl.key}, ttlMillis).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (rl *RedisLimiter) validate() error {
	if rl == nil || rl.client == nil {
		return ErrInvalidRedisClient
	}
	if rl.key == "" {
		return ErrInvalidRedisKey
	}
	if rl.limit <= 0 {
		return ErrInvalidCapacity
	}
	if rl.duration <= 0 {
		return ErrInvalidDuration
	}
	return nil
}
