// Attackers query for a key that does not exist in the DB (e.g., user:-1). The cache will always miss, forcing the DB to process the query, potentially bringing the system down.

// Solution: Cache null/empty values with a short TTL, or use a Bloom Filter.
package main

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func CachePenetration() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:8000",
	})
	var dbUser = "hello"

	if dbUser == "" {
		rdb.Set(ctx, "user:-1", "NULL", 1*time.Minute)
	}
}
