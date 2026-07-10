// How it works: Imagine a bucket with a maximum capacity of tokens. Tokens are added to the bucket at a constant rate. Every request removes one token. If the bucket is empty, the request is rejected.
// Pros: Highly memory efficient and allows for sudden bursts of traffic (up to the bucket's capacity) while enforcing an average rate over time.
// Cons: Can be slightly tricky to tune the relationship between bucket capacity and refill rate.

package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refilRate  float64
	lastRefill time.Time
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refilRate:  refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.tokens += elapsed * tb.refilRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens -= 1
		return true
	}
	return false
}
