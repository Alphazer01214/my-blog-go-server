package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"blog.alphazer01214.top/internal/global"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

// CacheHandler 处理缓存失效事件
type CacheHandler struct{}

// NewCacheHandler 创建缓存失效处理器
func NewCacheHandler() *CacheHandler {
	return &CacheHandler{}
}

// Handle 处理消息
func (h *CacheHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event pkgKafka.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event failed: %w", err)
	}

	rdb := global.GetRedis()

	switch event.TargetType {
	case "post":
		// 失效帖子详情缓存
		key := fmt.Sprintf("cache:post:%d", event.TargetID)
		if err := rdb.Del(ctx, key).Err(); err != nil {
			slog.Error("cache invalidate failed", "key", key, "err", err)
		} else {
			slog.Info("cache invalidated", "key", key)
		}

		// 失效帖子列表缓存
		if err := rdb.Del(ctx, "cache:posts:hot").Err(); err != nil {
			slog.Error("cache invalidate failed", "key", "cache:posts:hot", "err", err)
		}

	case "user":
		// 失效用户资料缓存
		key := fmt.Sprintf("cache:user:%d", event.TargetID)
		if err := rdb.Del(ctx, key).Err(); err != nil {
			slog.Error("cache invalidate failed", "key", key, "err", err)
		} else {
			slog.Info("cache invalidated", "key", key)
		}

		// 失效用户信息缓存
		key2 := fmt.Sprintf("cache:userinfo:%d", event.TargetID)
		rdb.Del(ctx, key2)

	case "comment":
		// 评论变更时，失效对应帖子的缓存
		key := fmt.Sprintf("cache:post:%d", event.TargetID)
		if err := rdb.Del(ctx, key).Err(); err != nil {
			slog.Error("cache invalidate failed", "key", key, "err", err)
		} else {
			slog.Info("cache invalidated", "key", key)
		}
	}

	return nil
}
