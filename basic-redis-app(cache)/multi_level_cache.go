// Network calls to Redis (L2) still take ~1ms. To get microsecond latency, we use an in-memory local cache (L1) inside the Go application itself, falling back to Redis, then the DB.
package main

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

var localCache = cache.New(5*time.Minute, 10*time.Minute)

func getMultiLevel(rdb *redis.Client, key string) string {
	// 1. Check L1 (Local Go memory - 0 network hops)
	if val, found := localCache.Get(key); found {
		return val.(string)
	}

	// 2. Check L2 (Redis - 1 network hop)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		localCache.Set(key, val, cache.DefaultExpiration) // Populate L1
		return val
	}

	// 3. Fallback to DB...
	return "db_data"
}
