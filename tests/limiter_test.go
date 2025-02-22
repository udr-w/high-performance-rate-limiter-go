package limiter_test

import (
	"testing"
	"time"

	"github.com/udr-w/high-performance-rate-limiter-go/limiter"
)

func TestTokenBucket(t *testing.T) {
	tb := limiter.NewTokenBucket(5, 10) // 5 tokens per second, max 10 tokens

	// Allow should pass initially (bucket starts full)
	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true, got false")
	}

	// Exhaust the tokens
	for i := 0; i < 9; i++ {
		tb.Allow()
	}

	// Now it should be empty
	if tb.Allow() {
		t.Errorf("Expected Allow() to return false after consuming all tokens")
	}

	// Wait 1 second to refill tokens
	time.Sleep(time.Second)

	if !tb.Allow() {
		t.Errorf("Expected Allow() to return true after refill, got false")
	}
}
