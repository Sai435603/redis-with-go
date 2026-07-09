// This happens when thousands of cached items expire at the exact same time (e.g., a batch job populates the cache at midnight with a 24-hour TTL). The DB gets crushed the next night at midnight.

// Solution: Add randomized "Jitter" to your TTLs.

package main

import (
	"math/rand"
	"time"
)

func generateJitterTTL(baseMinutes int) time.Duration {
	// Base time + random jitter between 0 and 5 minutes
	jitter := rand.Intn(5)
	return time.Duration(baseMinutes+jitter) * time.Minute
}

// rdb.Set(ctx, key, val, generateJitterTTL(60))
