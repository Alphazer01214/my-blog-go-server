package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Hub WebSocket 连接管理器
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 房间映射: roomID -> clients
	rooms map[string]map[*Client]bool

	// 用户映射: userID -> clients
	users map[uint]map[*Client]bool

	// 注册通道
	register chan *Client

	// 注销通道
	unregister chan *Client

	// 互斥锁
	mu sync.RWMutex
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		users:      make(map[uint]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动 Hub 主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)
		}
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// addClient 添加客户端
func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// 添加到用户映射
	if h.users[client.userID] == nil {
		h.users[client.userID] = make(map[*Client]bool)
	}
	h.users[client.userID][client] = true

	slog.Info("websocket client connected", "user_id", client.userID)
}

// removeClient 移除客户端
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)

	// 从用户映射中移除
	if clients, ok := h.users[client.userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.users, client.userID)
		}
	}

	// 从所有房间中移除
	for roomID := range client.rooms {
		if clients, ok := h.rooms[roomID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}

	close(client.send)
	slog.Info("websocket client disconnected", "user_id", client.userID)
}

// JoinRoom 加入房间
func (h *Hub) JoinRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true

	slog.Info("client joined room", "user_id", client.userID, "room_id", roomID)
}

// LeaveRoom 离开房间
func (h *Hub) LeaveRoom(client *Client, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[roomID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}

	slog.Info("client left room", "user_id", client.userID, "room_id", roomID)
}

// BroadcastToRoom 向房间内所有客户端广播消息
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	clients, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.send <- message:
		default:
			// 发送缓冲区满，关闭连接
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// SendToUser 向指定用户发送消息
func (h *Hub) SendToUser(userID uint, message []byte) {
	h.mu.RLock()
	clients, ok := h.users[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.send <- message:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// SendNotification 向用户发送通知消息
func (h *Hub) SendNotification(userID uint, notification NotificationData) {
	data, err := json.Marshal(notification)
	if err != nil {
		slog.Error("marshal notification failed", "err", err)
		return
	}

	msg := Message{
		Type: MsgTypeNotification,
		Data: data,
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.SendToUser(userID, msgData)
}

// GetOnlineCount 获取在线用户数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.users)
}

// GetRoomOnlineCount 获取房间在线用户数
func (h *Hub) GetRoomOnlineCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}

// GetRoomOnlineUsers 获取房间在线用户列表
func (h *Hub) GetRoomOnlineUsers(roomID string) []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return nil
	}

	users := make([]uint, 0, len(clients))
	seen := make(map[uint]bool)
	for client := range clients {
		if !seen[client.userID] {
			users = append(users, client.userID)
			seen[client.userID] = true
		}
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.users[userID]
	return ok && len(clients) > 0
}
