// How it works: Requests are added to a queue (the bucket). A background process pulls requests from the queue at a strictly constant rate (leaking). If the queue is full, new requests overflow and are rejected.
// Pros: Forces traffic to flow at a strictly smooth, consistent rate (great for downstream databases that can't handle bursts).
// Cons: Bursts fill up the bucket with old requests, causing new, potentially more urgent requests to be dropped.

package ratelimit

import "time"

type LeakyBucket struct {
	queue chan struct{}
}

func NewLeakyBucket(capacity int, leakRate time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		queue: make(chan struct{}, capacity),
	}
	// Background goroutine leaks the bucket at a constant rate
	go func() {
		ticker := time.NewTicker(leakRate)
		for range ticker.C {
			select {
			case <-lb.queue:
			default:
			}
		}
	}()

	return lb
}

func (lb *LeakyBucket) Allow() bool {
	select {
	case lb.queue <- struct{}{}:
		return true
	default:
		return false
	}
}
