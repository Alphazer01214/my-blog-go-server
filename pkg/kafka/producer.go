package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer 封装 kafka.Writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer 创建一个新的 Kafka 生产者
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			BatchSize:    100,
			BatchTimeout: 10 * time.Millisecond,
			Compression:  kafka.Lz4,
			Async:        false, // 同步写入，保证可靠性
		},
	}
}

// Publish 发布单条消息
func (p *Producer) Publish(ctx context.Context, key string, event Event) error {
	event.Timestamp = time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

// PublishBatch 批量发布消息
func (p *Producer) PublishBatch(ctx context.Context, messages []kafka.Message) error {
	return p.writer.WriteMessages(ctx, messages...)
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}
