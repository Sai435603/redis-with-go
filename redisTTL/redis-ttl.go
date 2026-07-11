// 1. What is TTL?
// TTL (Time-To-Live) is a mechanism that allows you to set a lifespan for a key in Redis. Once that time elapses, the key and its associated value are automatically deleted from the database. It is a fundamental feature for managing memory, enforcing security policies, and caching transient data without requiring manual deletion.

package redisttl

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis"
)

var ctx context.Context

func main() {
	client := redis.Client(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Setting a key with no initial expiration

	{
		client.Set(ctx, "mykey", "Hello World", 0)

		// Setting the key to expire in 10 seconds using EXPIRE
		err := client.Expire(ctx, "mykey", 10*time.Second).Err()
		if err != nil {
			panic(err)
		}
		fmt.Println("Key will expire in 10 seconds")
	}
	// Remove the expiration from "mykey"
	{
		isPersisted, err := client.Persist(ctx, "mykey").Result()
		if err != nil {
			panic(err)
		}

		if isPersisted {
			fmt.Println("Key is now permanent!")
		}
	}

	// 	The TTL command returns the remaining time to live of a key that has a timeout, measured in seconds.

	// Returns -2 if the key does not exist.

	// Returns -1 if the key exists but has no associated expiration.

	{
		// Check the remaining TTL in seconds
		ttl, err := client.TTL(ctx, "mykey").Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Time remaining: %v\n", ttl)
		// Output might be: Time remaining: 9s
	}

	{

		// Check the remaining TTL in milliseconds for more precision
		pttl, err := client.PTTL(ctx, "mykey").Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Time remaining in ms: %v\n", pttl.Milliseconds())
	}

}

// Real-World Use Cases

func createSession(client *redis.Client, sessionID string, userData string) {
	// Session expires after 30 minutes of inactivity
	sessionDuration := 30 * time.Minute

	err := client.Set(ctx, "session:"+sessionID, userData, sessionDuration).Err()
	if err != nil {
		fmt.Println("Error saving session:", err)
		return
	}
	fmt.Println("Session created for user.")
}

func refreshSession(client *redis.Client, sessionID string) {
	// Reset the expiration back to 30 minutes on user activity
	err := client.Expire(ctx, "session:"+sessionID, 30*time.Minute).Err()
	if err != nil {
		fmt.Println("Error refreshing session:", err)
	}
}

//otp -service

func generateOTP(client *redis.Client, email string, otpCode string) {
	// OTP is only valid for 3 minutes
	otpDuration := 3 * time.Minute
	redisKey := "otp:" + email

	err := client.Set(ctx, redisKey, otpCode, otpDuration).Err()
	if err != nil {
		fmt.Println("Error storing OTP:", err)
		return
	}
	fmt.Printf("OTP generated and sent to %s. It expires in 3 minutes.\n", email)
}

func verifyOTP(client *redis.Client, email string, providedOTP string) bool {
	redisKey := "otp:" + email

	// Fetch the OTP
	storedOTP, err := client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		fmt.Println("OTP has expired or does not exist.")
		return false
	} else if err != nil {
		fmt.Println("Error fetching OTP:", err)
		return false
	}

	// Verify and delete upon single use
	if storedOTP == providedOTP {
		client.Del(ctx, redisKey) // OTPs are single-use!
		return true
	}

	return false
}
