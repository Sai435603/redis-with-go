// Here is how you combine Cache Aside, Singleflight, and Jitter TTL into a robust production struct in Go.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type UserService struct {
	redis *redis.Client
	sf    singleflight.Group
}

func NewUserService(r *redis.Client) *UserService {
	return &UserService{redis: r}
}

// Fetch DB Mock
func (s *UserService) fetchFromDB(userID string) (string, error) {
	fmt.Println(">>> [DB HIT] Executing heavy query for:", userID)
	time.Sleep(500 * time.Millisecond) // Simulate DB latency
	return "User_Data_For_" + userID, nil
}

func (s *UserService) GetUserProfile(ctx context.Context, userID string) (string, error) {
	cacheKey := "profile:" + userID

	// 1. Cache Aside pattern
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return val, nil // Cache hit
	}

	// 2. Singleflight to prevent Cache Stampede
	result, err, _ := s.sf.Do(cacheKey, func() (any, error) {
		
		// 3. Fetch from DB
		data, dbErr := s.fetchFromDB(userID)
		if dbErr != nil {
			return nil, dbErr
		}

		// 4. Cache Penetration prevention (Cache empty results short term)
		if data == "" {
			s.redis.Set(ctx, cacheKey, "NULL", 30*time.Second)
			return "", nil
		}

		// 5. Cache Avalanche prevention (Jitter TTL)
		baseTTL := 60
		jitter := rand.Intn(10) // 0-10 minutes jitter
		ttl := time.Duration(baseTTL+jitter) * time.Minute

		s.redis.Set(ctx, cacheKey, data, ttl)
		return data, nil
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}