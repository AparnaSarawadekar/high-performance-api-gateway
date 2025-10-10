package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Thread-safe token bucket (float tokens).
type bucket struct {
	mu        sync.Mutex
	capacity  float64
	tokens    float64
	refillRPS float64
	last      time.Time
}

func newBucket(capacity int, refillRPS float64) *bucket {
	now := time.Now()
	return &bucket{
		capacity:  float64(capacity),
		tokens:    float64(capacity), // start full to allow a burst
		refillRPS: refillRPS,
		last:      now,
	}
}

// allow consumes 1 token if available. Returns (allowed, retryAfter, remainingTokensInt).
func (b *bucket) allow() (bool, time.Duration, int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.capacity, b.tokens+(elapsed*b.refillRPS))
		b.last = now
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, 0, int(math.Floor(b.tokens))
	}

	need := 1.0 - b.tokens
	sec := need / b.refillRPS
	if sec < 0 {
		sec = 0
	}
	return false, time.Duration(math.Ceil(sec)) * time.Second, int(math.Floor(b.tokens))
}