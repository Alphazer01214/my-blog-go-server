package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/search"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

// PostEventHandler 处理帖子相关事件
type PostEventHandler struct {
	searchService *search.SearchService
}

// NewPostEventHandler 创建帖子事件处理器
func NewPostEventHandler(searchService *search.SearchService) *PostEventHandler {
	return &PostEventHandler{
		searchService: searchService,
	}
}

// Handle 处理消息
func (h *PostEventHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event pkgKafka.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event failed: %w", err)
	}

	// 如果 ES 不可用，跳过
	if h.searchService == nil || !h.searchService.IsAvailable(ctx) {
		slog.Debug("elasticsearch not available, skip post event", "action", event.Action, "post_id", event.TargetID)
		return nil
	}

	switch event.Action {
	case pkgKafka.ActionCreate, pkgKafka.ActionUpdate:
		return h.handleIndex(ctx, event)
	case pkgKafka.ActionDelete:
		return h.handleDelete(ctx, event)
	default:
		slog.Debug("post event: unknown action", "action", event.Action)
	}

	return nil
}

// handleIndex 索引帖子到 ES
func (h *PostEventHandler) handleIndex(ctx context.Context, event pkgKafka.Event) error {
	// 从 DB 获取最新帖子数据
	var post entity.Post
	if err := global.GetDB().Where("id = ?", event.TargetID).First(&post).Error; err != nil {
		slog.Error("post not found for indexing", "post_id", event.TargetID, "err", err)
		return nil // 帖子不存在，跳过
	}

	// 获取作者名
	var profile entity.UserProfile
	if err := global.GetDB().Where("user_id = ?", post.UserId).First(&profile).Error; err != nil {
		slog.Warn("author profile not found", "user_id", post.UserId)
	}

	if err := h.searchService.IndexPost(ctx, &post, profile.Username); err != nil {
		slog.Error("index post to es failed", "post_id", event.TargetID, "err", err)
		return err
	}

	slog.Info("post indexed to es", "post_id", event.TargetID, "action", event.Action)
	return nil
}

// handleDelete 从 ES 删除帖子
func (h *PostEventHandler) handleDelete(ctx context.Context, event pkgKafka.Event) error {
	if err := h.searchService.DeletePost(ctx, event.TargetID); err != nil {
		slog.Error("delete post from es failed", "post_id", event.TargetID, "err", err)
		return err
	}

	slog.Info("post deleted from es", "post_id", event.TargetID)
	return nil
}
