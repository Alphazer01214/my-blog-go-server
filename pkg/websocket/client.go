package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写超时
	writeWait = 10 * time.Second

	// 读超时
	pongWait = 60 * time.Second

	// 心跳间隔（pongWait 的 90%）
	pingPeriod = 54 * time.Second

	// 最大消息大小
	maxMessageSize = 8192
)

// Client WebSocket 客户端连接
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uint
	rooms  map[string]bool
	mu     sync.RWMutex
}

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, userID uint) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		rooms:  make(map[string]bool),
	}
}

// ReadPump 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Error("websocket read error", "user_id", c.userID, "err", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump 写入消息到客户端
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError(400, "invalid message format")
		return
	}

	switch msg.Type {
	case MsgTypePing:
		c.sendMessage(Message{Type: MsgTypePong})

	case MsgTypeJoinRoom:
		roomID := msg.RoomID
		if roomID == "" {
			c.sendError(400, "room_id is required")
			return
		}
		c.hub.JoinRoom(c, roomID)
		c.mu.Lock()
		c.rooms[roomID] = true
		c.mu.Unlock()

	case MsgTypeLeaveRoom:
		roomID := msg.RoomID
		if roomID == "" {
			return
		}
		c.hub.LeaveRoom(c, roomID)
		c.mu.Lock()
		delete(c.rooms, roomID)
		c.mu.Unlock()

	default:
		// 广播到房间
		if msg.RoomID != "" {
			c.hub.BroadcastToRoom(msg.RoomID, data)
		}
	}
}

// sendMessage 发送消息给客户端
func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		// 发送缓冲区满，关闭连接
		close(c.send)
	}
}

// sendError 发送错误消息
func (c *Client) sendError(code int, message string) {
	errData, _ := json.Marshal(ErrorData{Code: code, Message: message})
	c.sendMessage(Message{Type: MsgTypeError, Data: errData})
}

// GetUserID 获取用户 ID
func (c *Client) GetUserID() uint {
	return c.userID
}

// GetRooms 获取客户端加入的房间
func (c *Client) GetRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
