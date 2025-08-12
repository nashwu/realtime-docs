package ws

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"log/slog"
	"realtime-docs/internal/app"
)

type BusMessage struct {
	DocID   string `json:"docId"`
	Payload []byte `json:"payload"`
}

type RedisBus struct {
	rdb *redis.Client
	log *slog.Logger
}

// NewRedisBus connects to redis and verifies connectivity
func NewRedisBus(ctx context.Context, cfg app.Config, log *slog.Logger) (*RedisBus, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   cfg.RedisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisBus{rdb: rdb, log: log}, nil
}

// Publish sends a message to the redis channel for a doc
func (b *RedisBus) Publish(ctx context.Context, m BusMessage) error {
	raw, _ := json.Marshal(m)
	return b.rdb.Publish(ctx, channel(m.DocID), raw).Err()
}

// Subscribe listens to all doc channels and invokes fn for each message
func (b *RedisBus) Subscribe(ctx context.Context, fn func(BusMessage)) {
	pubsub := b.rdb.PSubscribe(ctx, channel("*"))
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			_ = pubsub.Close()
			return
		case msg := <-ch:
			var bm BusMessage
			_ = json.Unmarshal([]byte(msg.Payload), &bm)
			if bm.DocID != "" {
				fn(bm)
			}
		}
	}
}

// Close shuts down the redis connection
func (b *RedisBus) Close() { _ = b.rdb.Close() }

// channel namespacing for doc pub/sub
func channel(docID string) string { return "doc:" + docID }
