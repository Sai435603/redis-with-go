package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	fmt.Println("Hello world!!")
	_ = godotenv.Load()
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
	} else {
		fmt.Println("Successfully connected to Redis!")
	}

	r := chi.NewRouter()
	rawJSON := `{
		"name": "Sai",
		"email": "sai@example.com",
		"role": "Developer",
		"skills": ["Go", "Redis", "Docker"]
	}`
	err = rdb.Set(ctx, "user:raw", rawJSON, 0).Err()
	if err != nil {
		fmt.Errorf("%v", err)
		return
	}
	fmt.Println("Saved raw JSON string directly to Redis!")
	val, err := rdb.Get(ctx, "user:raw").Result()
	if err != nil {
		fmt.Printf("Error fetching from Redis: %v\n", err)
		return
	}
	rdb.Set(ctx, "visits", 0, 0)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rdb.Incr(ctx, "visits")
		visit, _ := rdb.Get(ctx, "visits").Result()
		fmt.Println(visit, val)
		// json.NewEncoder(w).Encode(val)
		w.Write([]byte(val))
	})
	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8000", r)

}
