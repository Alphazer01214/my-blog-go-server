package websocket

import "encoding/json"

// MessageType 消息类型
type MessageType string

const (
	// 系统消息
	MsgTypePing      MessageType = "ping"
	MsgTypePong      MessageType = "pong"
	MsgTypeConnected MessageType = "connected"
	MsgTypeError     MessageType = "error"

	// 通知消息
	MsgTypeNotification MessageType = "notification"

	// 房间消息
	MsgTypeJoinRoom  MessageType = "join_room"
	MsgTypeLeaveRoom MessageType = "leave_room"
	MsgTypeRoomMsg   MessageType = "room_message"

	// 在线状态
	MsgTypeOnlineUsers MessageType = "online_users"
)

// Message WebSocket 消息结构
type Message struct {
	Type    MessageType     `json:"type"`
	RoomID  string          `json:"room_id,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NotificationData 通知数据
type NotificationData struct {
	ID         uint   `json:"id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	SourceID   uint   `json:"source_id"`
	SourceType string `json:"source_type"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}

// OnlineUsersData 在线用户数据
type OnlineUsersData struct {
	RoomID string `json:"room_id"`
	Count  int    `json:"count"`
	Users  []uint `json:"users"`
}

// ErrorData 错误数据
type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
