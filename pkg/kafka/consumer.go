package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// EventHandler 事件处理函数签名
type EventHandler func(ctx context.Context, msg kafka.Message) error

// Consumer 封装 kafka.Reader
type Consumer struct {
	reader  *kafka.Reader
	handler EventHandler
	redis   *redis.Client
	topic   string
	groupID string
}

// NewConsumer 创建一个新的 Kafka 消费者
func NewConsumer(brokers []string, topic, groupID string, handler EventHandler, redisClient *redis.Client) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
			StartOffset:    kafka.FirstOffset,
		}),
		handler: handler,
		redis:   redisClient,
		topic:   topic,
		groupID: groupID,
	}
}

// Consume 开始消费消息，阻塞直到 ctx 取消
func (c *Consumer) Consume(ctx context.Context) {
	slog.Info("kafka consumer started", "topic", c.topic, "group", c.groupID)
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("kafka consumer stopped", "topic", c.topic, "group", c.groupID)
				return
			}
			slog.Error("kafka consume error", "topic", c.topic, "err", err)
			continue
		}

		// 幂等处理：通过 Redis SETNX 保证消息不重复消费
		lockKey := fmt.Sprintf("kafka:processed:%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)
		if c.redis != nil {
			added, err := c.redis.SetNX(ctx, lockKey, 1, 24*time.Hour).Result()
			if err != nil {
				slog.Error("kafka idempotent check failed", "err", err)
			}
			if !added {
				continue // 已处理过，跳过
			}
		}

		if err := c.handler(ctx, msg); err != nil {
			slog.Error("kafka handler error",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"err", err,
			)
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.reader.Close()
}
