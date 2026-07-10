// How it works: Instead of fixed time frames, it keeps a rolling log of request timestamps. When a new request arrives, it removes outdated timestamps (older than the window size) and checks if the remaining log count is below the limit.
// Pros: Perfectly smooths out traffic, solving the edge-spike problem completely.
// Cons: Highly memory-intensive, as it stores a timestamp for every single request.
package ratelimit

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests []time.Time
}

func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	 return &SlidingWindow{
		 limit: limit,
		 window: window,
	 }
}

func (s *SlidingWindow) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.window)

	//filtering out the old req's

	var validRequests []time.Time
	for _, reqTime := range s.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	s.requests = validRequests
	if len(s.requests) < s.limit {
		s.requests = append(s.requests, now)
		return true
	}
	return false
}
