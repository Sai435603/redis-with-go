// Redis Streams is one of the most powerful and versatile data structures in Redis. It acts as an append-only log, making it perfect for building robust message brokers, event-sourcing architectures, and job queues.
package redisstreams

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx context.Context

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update with your Redis address
	})
}

func xAddExample(rdb *redis.Client) {
	// Add an entry with key-value pairs to "mystream"
	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "mystream",
		Values: map[string]interface{}{
			"event": "user_signup",
			"user_id": "12345",
		},
	}).Err()

	if err != nil {
		fmt.Println("Error adding to stream:", err)
		return
	}
	fmt.Println("Message successfully added to 'mystream'")
}

func xRangeExample(rdb *redis.Client) {
	// Read all messages from the beginning (-) to the end (+)
	messages, err := rdb.XRange(ctx, "mystream", "-", "+").Result()
	if err != nil {
		fmt.Println("Error reading range:", err)
		return
	}

	for _, msg := range messages {
		fmt.Printf("ID: %s, Values: %v\n", msg.ID, msg.Values)
	}
}

func createConsumerGroup(rdb *redis.Client) {
	// Creates group "mygroup" on "mystream". 
	// The "0" means the group will read from the very beginning of the stream.
	// You can also use "$" to only read newly arriving messages.
	err := rdb.XGroupCreateMkStream(ctx, "mystream", "mygroup", "0").Err()
	if err != nil {
		// Ignore error if group already exists
		fmt.Println("Group creation note:", err)
	}
}

func readAsConsumer(rdb *redis.Client, consumerName string) string {
	// Read up to 1 new message for this specific consumer
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "mygroup",
		Consumer: consumerName,
		Streams:  []string{"mystream", ">"}, // ">" means new messages only
		Count:    1,
		Block:    2 * time.Second, // Wait up to 2 seconds if no message is available
	}).Result()

	if err == redis.Nil {
		fmt.Println(consumerName, "- No new messages")
		return ""
	} else if err != nil {
		fmt.Println("Error reading group:", err)
		return ""
	}

	msgID := streams[0].Messages[0].ID
	fmt.Printf("%s received message ID: %s\n", consumerName, msgID)
	return msgID
}

func acknowledgeMessage(rdb *redis.Client, msgID string) {
	err := rdb.XAck(ctx, "mystream", "mygroup", msgID).Err()
	if err != nil {
		fmt.Println("Error acknowledging message:", err)
		return
	}
	fmt.Printf("Message %s acknowledged successfully!\n", msgID)
}

func checkPending(rdb *redis.Client) {
	// Get a summary of pending messages in the group
	pendingInfo, err := rdb.XPending(ctx, "mystream", "mygroup").Result()
	if err != nil {
		fmt.Println("Error checking pending:", err)
		return
	}
	fmt.Printf("Total pending messages: %d\n", pendingInfo.Count)
	
	// You can also use XPendingExt to get detailed information about specific stuck messages
}

func autoClaimStaleMessages(rdb *redis.Client, activeConsumerName string) {
	// Claim messages that have been pending for more than 1 minute
	// "0-0" is the starting ID to scan from
	claimed, _, err := rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   "mystream",
		Group:    "mygroup",
		Consumer: activeConsumerName,
		MinIdle:  1 * time.Minute, 
		Start:    "0-0",
		Count:    10,
	}).Result()

	if err != nil {
		fmt.Println("Error claiming messages:", err)
		return
	}

	for _, msg := range claimed {
		fmt.Printf("%s successfully claimed stuck message %s\n", activeConsumerName, msg.ID)
		// Process it, then don't forget to ACK it!
		// acknowledgeMessage(rdb, msg.ID)
	}
}