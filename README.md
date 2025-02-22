# High-Performance Rate Limiter in Golang

## Overview
A high-performance, thread-safe rate limiter built in Golang using token bucket and leaky bucket algorithms. Includes a Redis-backed distributed version for scalability and benchmarking insights.

## Features
- Token Bucket & Leaky Bucket implementations
- Thread-safe and optimized for high concurrency
- Redis-backed version for distributed environments
- Benchmark comparisons with Go’s built-in `rate` package
- Unit tests and performance profiling

## Installation
```sh
git clone https://github.com/yourusername/high-performance-rate-limiter-go.git
cd high-performance-rate-limiter-go
go mod init github.com/yourusername/high-performance-rate-limiter-go
go mod tidy
```

## Run Redis Locally
```sh
# Pull & run Redis container
docker run --name redis -p 6379:6379 -d redis

# Check running Redis containers
docker ps --filter "name=redis"

# If Redis is not running, start it
docker start redis
```

## Run Tests
```sh
go test ./...
```

## Usage Example
```go
package main

import (
    "fmt"
    "time"
    "your/package/rate_limiter"
)

func main() {
    limiter := rate_limiter.NewTokenBucket(5, time.Second)
    
    if limiter.Allow() {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Rate limit exceeded")
    }
}
```

## Benchmarks
goos: linux
goarch: amd64
pkg: github.com/udr-w/high-performance-rate-limiter-go/tests
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
### Token Bucket Limiter
BenchmarkTokenBucket-8          67066435                47.58 ns/op
### Redis Bucket Limiter
BenchmarkRedisLimiter-8           130024             25220 ns/op

## License
MIT License. See `LICENSE` file for details.