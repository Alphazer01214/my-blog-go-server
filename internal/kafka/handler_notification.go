package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	pkgKafka "blog.alphazer01214.top/pkg/kafka"
	ws "blog.alphazer01214.top/pkg/websocket"
	"github.com/segmentio/kafka-go"
)

// NotificationHandler 处理通知相关事件
type NotificationHandler struct {
	hub *ws.Hub
}

// NewNotificationHandler 创建通知事件处理器
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// SetHub 设置 WebSocket Hub（用于实时推送）
func (h *NotificationHandler) SetHub(hub *ws.Hub) {
	h.hub = hub
}

// Handle 处理消息
func (h *NotificationHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event pkgKafka.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event failed: %w", err)
	}

	// 从 Extra 中提取通知信息
	extraData, err := json.Marshal(event.Extra)
	if err != nil {
		return fmt.Errorf("marshal extra failed: %w", err)
	}

	var notifEvent pkgKafka.NotificationEvent
	if err := json.Unmarshal(extraData, &notifEvent); err != nil {
		return fmt.Errorf("unmarshal notification event failed: %w", err)
	}

	if notifEvent.ReceiverID == 0 {
		return nil // 没有接收者，跳过
	}

	// 不给自己发通知
	if notifEvent.ReceiverID == event.UserID {
		return nil
	}

	// 写入数据库
	notification := entity.Notification{
		UserId:     notifEvent.ReceiverID,
		Type:       event.Action,
		SourceId:   event.TargetID,
		SourceType: event.TargetType,
		Content:    notifEvent.Content,
		IsRead:     false,
	}

	if err := global.GetDB().Create(&notification).Error; err != nil {
		return fmt.Errorf("create notification failed: %w", err)
	}

	slog.Info("notification created",
		"receiver", notifEvent.ReceiverID,
		"type", event.Action,
		"source", fmt.Sprintf("%s:%d", event.TargetType, event.TargetID),
	)

	// 通过 WebSocket 实时推送给在线用户
	if h.hub != nil && h.hub.IsUserOnline(notifEvent.ReceiverID) {
		h.hub.SendNotification(notifEvent.ReceiverID, ws.NotificationData{
			ID:         notification.ID,
			Type:       notification.Type,
			Content:    notification.Content,
			SourceID:   notification.SourceId,
			SourceType: notification.SourceType,
			IsRead:     false,
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		slog.Info("notification pushed via websocket", "receiver", notifEvent.ReceiverID)
	}

	return nil
}
