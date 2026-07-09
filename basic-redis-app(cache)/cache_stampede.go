// If a highly popular cache key (e.g., "trending_topics") expires, thousands of concurrent requests will all get a Cache Miss and hit the database simultaneously, crashing it.
package main

import (
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var g singleflight.Group

func getTrendingTopics(rdb *redis.Client) (string, error) {
	val, err := rdb.Get(ctx, "trending").Result()
	if err == nil {
		return val, nil
	}

	// If 10,000 requests hit this simultaneously, Do() ensures the function
	// inside only executes exactly ONCE.
	result, err, _ := g.Do("trending", func() (any, error) {
		// Fetch from DB
		dbVal := "trend1, trend2"
		rdb.Set(ctx, "trending", dbVal, 10*time.Minute)
		return dbVal, nil
	})

	return result.(string), err
}
