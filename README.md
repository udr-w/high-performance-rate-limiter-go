# High-Performance Rate Limiter for Go

A concurrency-safe Go rate limiting library for services that need predictable local limits and Redis-backed distributed limits.

## What It Provides

- Thread-safe token bucket limiter for burst-tolerant local limits.
- Thread-safe leaky bucket limiter for bounded backlog and smoother flow.
- Redis-backed distributed fixed-window limiter with atomic counter and TTL updates.
- Context-aware Redis API that returns infrastructure errors separately from rate-limit denials.
- Deterministic tests for local algorithms and race-safe concurrency coverage.
- Redis integration tests that skip cleanly when Redis is not available.

## Install

```sh
go get github.com/udr-w/high-performance-rate-limiter-go
```

## Token Bucket

Token bucket is a good default when a service should allow short bursts while enforcing a sustained average rate.

```go
package main

import (
	"fmt"

	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

func main() {
	tb, err := limiter.NewTokenBucketWithOptions(100, 200)
	if err != nil {
		panic(err)
	}

	if tb.Allow() {
		fmt.Println("request allowed")
		return
	}

	fmt.Println("rate limited")
}
```

For compatibility with the original API, `NewTokenBucket(rate, capacity int)` is also available. Invalid configuration creates a closed limiter. Prefer `NewTokenBucketWithOptions` in production so configuration errors fail early.

## Leaky Bucket

Leaky bucket is useful when the application should smooth traffic and reject requests once a bounded queue is full.

```go
lb, err := limiter.NewLeakyBucketWithOptions(50, 100)
if err != nil {
	panic(err)
}

allowed := lb.Allow()
```

## Redis Distributed Limiter

`RedisLimiter` implements fixed-window limiting. It uses a Lua script so increment and expiration are applied atomically, and it repairs missing TTLs on existing keys.

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	defer client.Close()

	key := limiter.RedisKey("login", "user-123")
	rl, err := limiter.NewRedisLimiterWithValidation(client, key, 10, time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	allowed, err := rl.AllowContext(ctx)
	if err != nil {
		// Choose the policy that fits the endpoint: fail closed, fail open,
		// or fall back to a local limiter.
		log.Fatal(err)
	}
	if !allowed {
		log.Println("rate limited")
	}
}
```

Fixed-window semantics can allow boundary bursts around window rollover. For abuse-sensitive workloads that need stricter smoothing across distributed instances, add a Redis token bucket or sliding-window limiter and benchmark it against your production Redis topology.

## Redis Key Safety

Do not use raw untrusted user input as Redis keys. `RedisKey(namespace, identifier)` hashes identifiers and applies a consistent namespace:

```go
key := limiter.RedisKey("api:v1:checkout", userID)
```

Use dedicated Redis credentials, ACLs, timeout settings, and key eviction policies appropriate for limiter data.

## Tests

Run unit and integration tests:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Redis integration tests use `localhost:6379` when available and skip otherwise.

Run Redis locally:

```sh
docker run --name rate-limiter-redis -p 6379:6379 -d redis:7
```

## Benchmarks

```sh
go test -bench=. -benchmem ./...
```

Latest benchmark report: [docs/benchmarks.md](docs/benchmarks.md).

Summary from the 2026-07-18 run on `linux/amd64`, Go `1.26.5`, Intel i7-1165G7, Redis `7.4.9` in local Docker:

| Scenario | Median | Allocations |
| --- | ---: | ---: |
| TokenBucket allow | 20.55 ns/op | 0 B/op, 0 allocs/op |
| TokenBucket parallel allow | 81.82 ns/op | 0 B/op, 0 allocs/op |
| LeakyBucket allow | 22.49 ns/op | 0 B/op, 0 allocs/op |
| LeakyBucket parallel allow | 81.65 ns/op | 0 B/op, 0 allocs/op |
| RedisLimiter allow | 65.66 us/op | 311 B/op, 9 allocs/op |
| RedisLimiter parallel allow | 54.06 us/op | 313 B/op, 10 allocs/op |

Redis benchmarks require Redis at `localhost:6379`. Treat benchmark numbers as environment-specific; publish results only with hardware, Go version, Redis version, command details, and repeated samples.

## License

MIT License. See [LICENSE](LICENSE).
