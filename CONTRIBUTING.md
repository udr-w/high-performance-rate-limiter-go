# Contributing

Thanks for helping improve this project. The goal is to keep this repository credible as a production-oriented Go rate limiting library: small API surface, clear semantics, safe concurrency, deterministic tests, and honest documentation.

## Development Setup

Install Go using the version declared in `go.mod`.

Download dependencies:

```sh
go mod download
```

Run the standard quality gate:

```sh
make verify
```

This runs:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Run benchmarks when touching hot paths:

```sh
make bench
```

## Redis Integration Tests

Redis-backed tests use `localhost:6379` when Redis is available and skip cleanly otherwise.

To run them locally against Redis:

```sh
docker run --name rate-limiter-redis -p 6379:6379 -d redis:7
go test ./...
go test -race ./...
docker rm -f rate-limiter-redis
```

Do not use `FlushDB` in tests. Use unique keys and clean up only the keys created by the test.

## Code Standards

- Keep APIs small and explicit.
- Prefer constructors that validate configuration and return errors.
- Preserve backward compatibility unless the change is intentionally breaking and clearly documented.
- Keep limiter implementations safe for concurrent use.
- Use `context.Context` for operations that can block on external systems.
- Do not hide infrastructure errors in new APIs.
- Avoid unbounded memory growth and attacker-controlled key spaces.
- Document algorithm semantics, especially burst behavior and distributed consistency tradeoffs.

## Testing Standards

Every behavior change should include focused tests.

Use deterministic tests for time-based local limiters. Prefer injected clocks over `time.Sleep`.

Concurrent tests must pass under the race detector. Use atomics, mutexes, or channels for shared test state.

Redis tests should verify:

- Allowed and denied behavior.
- Multi-instance behavior with shared Redis keys.
- Context-aware error handling.
- TTL/window reset behavior.
- Isolation from other local Redis data.

## Benchmarks

Benchmark numbers are environment-specific. When changing benchmark claims or publishing results, include:

- Go version.
- CPU and OS.
- Redis version and deployment mode, when applicable.
- Exact benchmark command.
- Allocation counts from `-benchmem`.

Avoid benchmark claims that measure only an already-denied path unless that is the explicit scenario.

## Documentation

Update `README.md` when changing public APIs, behavior, setup, Redis semantics, or benchmark guidance.

Use precise language:

- Token bucket allows bursts up to capacity.
- Leaky bucket smooths traffic through a bounded queue model.
- The current Redis limiter is fixed-window and may allow boundary bursts.

Do not claim support for algorithms, performance numbers, or deployment guarantees that are not implemented and verified.

## Security

Report suspected vulnerabilities according to [SECURITY.md](SECURITY.md).

Security-sensitive changes should be reviewed carefully, especially changes involving Redis scripts, key construction, concurrency, context handling, fallback behavior, or validation.

## Pull Request Checklist

Before opening a pull request:

- Run `gofmt` on changed Go files.
- Run `make verify`.
- Run Redis-backed tests when changing Redis behavior.
- Add or update tests for behavior changes.
- Update documentation for public-facing changes.
- Explain compatibility impact and migration guidance if APIs change.
