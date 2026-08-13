package eventbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/accessible-path/route-service/internal/ports/out"
	"github.com/redis/go-redis/v9"
)

type RedisEventBus struct {
	client *redis.Client
}

func NewRedisEventBus(addr, password string, db int) *RedisEventBus {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisEventBus{client: client}
}

func (b *RedisEventBus) Publish(stream string, event map[string]interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return b.client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"data": string(data)},
	}).Err()
}

func (b *RedisEventBus) Subscribe(stream, group, consumer string, handler func(map[string]interface{})) error {
	ctx := context.Background()

	err := b.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}

	for {
		streams, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    5 * 1000,
		}).Result()
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(msg.Values["data"].(string)), &data); err != nil {
					continue
				}
				handler(data)
				b.client.XAck(ctx, stream.Stream, group, msg.ID)
			}
		}
	}
}

func (b *RedisEventBus) Close() error {
	return b.client.Close()
}

var _ out.EventBus = (*RedisEventBus)(nil)