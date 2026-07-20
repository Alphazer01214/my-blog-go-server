package kafka

import "time"

// Topic 常量定义
const (
	TopicNotification = "trading-forum.notifications"
	TopicPostEvent    = "trading-forum.post-events"
	TopicUserEvent    = "trading-forum.user-events"
	TopicAuditLog     = "trading-forum.audit-log"
)

// Event Action 常量
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionLike   = "like"
	ActionUnlike = "unlike"
	ActionFollow = "follow"
	ActionUnfollow = "unfollow"
	ActionComment = "comment"
)

// Event 统一消息结构
type Event struct {
	Action    string      `json:"action"`     // 动作类型
	Topic     string      `json:"-"`          // 不序列化到消息体，由 Producer 设置
	UserID    uint        `json:"user_id"`    // 操作者 ID
	TargetID  uint        `json:"target_id"`  // 目标对象 ID
	TargetType string     `json:"target_type"` // 目标类型: post/comment/user
	Extra     interface{} `json:"extra,omitempty"` // 附加数据
	Timestamp time.Time   `json:"timestamp"`
}

// NotificationEvent 通知事件的附加数据
type NotificationEvent struct {
	ReceiverID uint   `json:"receiver_id"` // 通知接收者
	Content    string `json:"content"`     // 通知内容
}
