// Package limiter provides thread-safe local and Redis-backed rate limiters.
//
// Local limiters are safe for concurrent use by multiple goroutines. RedisLimiter
// provides distributed fixed-window limits and exposes context-aware methods so
// callers can apply request deadlines and handle infrastructure errors
// explicitly.
package limiter
