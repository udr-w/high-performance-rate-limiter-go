# Benchmark Report

This report documents the benchmark methodology and raw results for the current limiter implementations. Benchmarks are intended to be reproducible and transparent, not universal performance guarantees.

## Summary

Collected on 2026-07-18 with five benchmark samples per scenario.

| Benchmark | Median | Mean | Allocations |
| --- | ---: | ---: | ---: |
| TokenBucket allow | 20.55 ns/op | 20.58 ns/op | 0 B/op, 0 allocs/op |
| TokenBucket deny | 22.09 ns/op | 22.07 ns/op | 0 B/op, 0 allocs/op |
| TokenBucket parallel allow | 81.82 ns/op | 82.95 ns/op | 0 B/op, 0 allocs/op |
| LeakyBucket allow | 22.49 ns/op | 23.03 ns/op | 0 B/op, 0 allocs/op |
| LeakyBucket deny | 23.17 ns/op | 24.00 ns/op | 0 B/op, 0 allocs/op |
| LeakyBucket parallel allow | 81.65 ns/op | 82.10 ns/op | 0 B/op, 0 allocs/op |
| RedisLimiter allow | 65.66 us/op | 66.78 us/op | 311 B/op, 9 allocs/op |
| RedisLimiter deny | 65.11 us/op | 65.75 us/op | 311 B/op, 9 allocs/op |
| RedisLimiter parallel allow | 54.06 us/op | 54.43 us/op | 313 B/op, 10 allocs/op |

## Environment

- OS: Ubuntu 22.04 family kernel, Linux `6.8.0-134-generic`
- Architecture: `linux/amd64`
- CPU: 11th Gen Intel Core i7-1165G7 @ 2.80GHz
- Logical CPUs: 8
- Go: `go1.26.5 linux/amd64`
- Redis: `redis:7` Docker container, Redis server `7.4.9`, jemalloc `5.3.0`
- Redis topology: local Docker container exposed on `localhost:6379`
- Go benchmark package: `github.com/udr-w/high-performance-rate-limiter-go/tests`

## Command

```sh
docker run --name rate-limiter-redis -p 6379:6379 -d redis:7
go test -run '^$' -bench='Benchmark(Token|Leaky|Redis)' -benchmem -benchtime=3s -count=5 ./tests
docker rm -f rate-limiter-redis
```

## Methodology

The benchmark suite separates distinct operational paths:

- `Allow`: enough capacity is configured so every request is accepted.
- `Deny`: the limiter is filled or exhausted before timing starts, so every measured request is rejected.
- `Parallel`: `testing.B.RunParallel` measures contention under concurrent calls.

Local limiter benchmarks use an injected deterministic clock. This keeps the benchmark focused on limiter bookkeeping, locking, and decision logic rather than wall-clock syscalls or scheduler timing.

Redis benchmarks include the Go Redis client, network loopback/Docker path, Lua script execution, Redis command processing, response parsing, and allocation behavior. They should be interpreted as an integration-path measurement for this machine, not a universal Redis latency claim.

## Interpretation

Local token and leaky bucket paths are allocation-free and complete in tens of nanoseconds in single-goroutine use. Parallel local benchmarks remain allocation-free, with the expected additional cost from mutex contention.

Redis allow and deny paths are similar because both execute the same atomic Lua script. Redis parallel throughput improves latency per operation in this local setup because multiple requests are in flight through the Redis client pool, but the operation still includes external system round trips and allocations.

Production Redis performance depends heavily on network distance, TLS, server load, Redis persistence settings, client pool sizing, and deployment topology. Benchmark again in the target environment before setting production capacity assumptions.

## Raw Results

```text
goos: linux
goarch: amd64
pkg: github.com/udr-w/high-performance-rate-limiter-go/tests
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkTokenBucketAllow-8        175271474    20.59 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketAllow-8        168252002    19.97 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketAllow-8        175925871    20.12 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketAllow-8        170339718    20.55 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketAllow-8        174519506    21.66 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketDeny-8         171583188    22.09 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketDeny-8         166288752    21.49 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketDeny-8         164090288    21.71 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketDeny-8         169256372    22.73 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketDeny-8         156228106    22.32 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketParallel-8      52594515    80.55 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketParallel-8      42766232    88.49 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketParallel-8      38401435    81.82 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketParallel-8      44769309    80.63 ns/op     0 B/op    0 allocs/op
BenchmarkTokenBucketParallel-8      44339205    83.26 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketAllow-8        158384293    22.16 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketAllow-8        163953944    23.17 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketAllow-8        161651654    22.43 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketAllow-8        160001244    24.92 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketAllow-8        156426312    22.49 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketDeny-8         162414650    22.74 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketDeny-8         163240386    22.05 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketDeny-8         153833312    23.17 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketDeny-8         164204073    23.21 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketDeny-8         128275827    28.85 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketParallel-8      45244908    80.90 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketParallel-8      43884174    83.76 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketParallel-8      43120144    81.14 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketParallel-8      45733867    81.65 ns/op     0 B/op    0 allocs/op
BenchmarkLeakyBucketParallel-8      43943294    83.04 ns/op     0 B/op    0 allocs/op
BenchmarkRedisLimiterAllow-8           55178    64447 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterAllow-8           53506    70343 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterAllow-8           54738    68568 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterAllow-8           51342    65655 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterAllow-8           55236    64863 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterDeny-8            51307    64814 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterDeny-8            55155    63806 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterDeny-8            54078    66473 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterDeny-8            55302    65108 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterDeny-8            55354    68538 ns/op   311 B/op    9 allocs/op
BenchmarkRedisLimiterParallel-8        71172    55351 ns/op   313 B/op   10 allocs/op
BenchmarkRedisLimiterParallel-8        65386    54055 ns/op   313 B/op   10 allocs/op
BenchmarkRedisLimiterParallel-8        65869    53704 ns/op   313 B/op   10 allocs/op
BenchmarkRedisLimiterParallel-8        62911    55683 ns/op   313 B/op   10 allocs/op
BenchmarkRedisLimiterParallel-8        66054    53356 ns/op   313 B/op   10 allocs/op
PASS
ok      github.com/udr-w/high-performance-rate-limiter-go/tests    219.558s
```

## Quality Bar

Before updating this report, run:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Benchmark updates should use multiple samples, include `-benchmem`, document Redis availability, and avoid comparing numbers collected under materially different environments without calling that out.
