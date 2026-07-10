// How it works: Time is divided into fixed intervals (e.g., 1 minute: 12:00 to 12:01). A counter increments for every request. If the counter exceeds the limit, requests are rejected until the next window begins.
// Pros: Very memory efficient and easy to implement.
// Cons: The "Edge Spike" problem. A user can send 100 requests at 12:00:59 and 100 at 12:01:01, effectively bypassing the intended limit within a 2-second span.

package ratelimit 

import (
	"sync"
	"time"
)

type FixedWindow struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	counter     int
	windowStart time.Time
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:       limit,
		window:      window,
		windowStart: time.Now(),
	}
}

func (f *FixedWindow) Allow() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	if now.Sub(f.windowStart) > f.window {
		f.windowStart = now
		f.counter = 0
	}
	if f.counter < f.limit {
		f.counter++
		return true
	}
	return false
}
