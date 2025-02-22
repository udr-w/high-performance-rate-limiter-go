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
(TBD: Run benchmarks and add results here)

## License
MIT License. See `LICENSE` file for details.