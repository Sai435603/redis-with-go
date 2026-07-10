package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type ChatMessage struct {
	Room      string    `json:"room"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatSystem struct {
	rdb *redis.Client
}

func NewChatSystem(redisAddr string) *ChatSystem {
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 3,
	})
	return &ChatSystem{rdb: rdb}
}

func (cs *ChatSystem) PublishMessage(ctx context.Context, msg ChatMessage) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	receivers, err := cs.rdb.Publish(ctx, msg.Room, buf.Bytes()).Result()
	if err != nil {
		return fmt.Errorf("failed to publish to redis: %w", err)
	}

	slog.Debug("Message published successfully", "room", msg.Room, "received_by_count", receivers)
	return nil
}

func (cs *ChatSystem) SubscribePattern(ctx context.Context, pattern string, wg *sync.WaitGroup) {
	defer wg.Done()

	pubsub := cs.rdb.PSubscribe(ctx, pattern)
	defer pubsub.Close()

	slog.Info("Subscribed to pattern", "pattern", pattern)
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down subscription loop due to context cancellation", "pattern", pattern)
			return
		case msg, ok := <-ch:
			if !ok {
				slog.Warn("Subscription channel closed abruptly")
				return
			}
			go cs.processIncomingMessage(msg)
		}
	}
}

func (cs *ChatSystem) processIncomingMessage(msg *redis.Message) {
	var chatMsg ChatMessage
	reader := strings.NewReader(msg.Payload)
	if err := json.NewDecoder(reader).Decode(&chatMsg); err != nil {
		slog.Error("Failed to decode message payload", "error", err, "raw", msg.Payload)
		return
	}

	slog.Info("New Chat Message Received",
		"matched_pattern", msg.Pattern,
		"actual_channel", msg.Channel,
		"sender", chatMsg.Sender,
		"content", chatMsg.Content,
		"sent_at", chatMsg.Timestamp.Format(time.RFC3339),
	)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chatServer := NewChatSystem("localhost:6379")
	if err := chatServer.rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Could not connect to Redis", "error", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go chatServer.SubscribePattern(ctx, "rooms.*", &wg)
	time.Sleep(500 * time.Millisecond)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		rooms := []string{"rooms.gaming", "rooms.tech", "rooms.lobby"}
		count := 1

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentRoom := rooms[count%len(rooms)]
				msg := ChatMessage{
					Room:      currentRoom,
					Sender:    fmt.Sprintf("User-%d", count),
					Content:   fmt.Sprintf("Hello everyone in %s! (Msg sequence: %d)", currentRoom, count),
					Timestamp: time.Now(),
				}

				if err := chatServer.PublishMessage(ctx, msg); err != nil {
					if !errors.Is(err, context.Canceled) {
						slog.Error("Publish failed", "error", err)
					}
				}
				count++
			}
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	slog.Info("Shutdown signal received. Cleaning up resources...")
	cancel()
	wg.Wait()
	_ = chatServer.rdb.Close()
	slog.Info("Application successfully stopped.")
}
