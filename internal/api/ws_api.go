package api

import (
	"log/slog"
	"net/http"

	"blog.alphazer01214.top/internal/response"
	ws "blog.alphazer01214.top/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源
	},
}

// WsHub WebSocket 连接管理器（由 main.go 注入）
var WsHub *ws.Hub

type WsApi struct{}

// HandleWebSocket 处理 WebSocket 连接
func (wa *WsApi) HandleWebSocket(c *gin.Context) {
	// 验证用户身份
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithAppError(c, err)
		return
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "err", err)
		return
	}

	// 创建客户端
	client := ws.NewClient(WsHub, conn, cl.UserId)

	// 注册到 Hub
	WsHub.Register(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// GetOnlineStatus 获取在线状态
func (wa *WsApi) GetOnlineStatus(c *gin.Context) {
	if WsHub == nil {
		response.ErrorWithMsg(c, "websocket not available")
		return
	}

	response.SuccessWithDetail(c, gin.H{
		"online_count": WsHub.GetOnlineCount(),
	}, "success")
}

// GetRoomOnlineStatus 获取房间在线状态
func (wa *WsApi) GetRoomOnlineStatus(c *gin.Context) {
	roomID := c.Param("room_id")
	if roomID == "" {
		response.ErrorWithMsg(c, "room_id is required")
		return
	}

	if WsHub == nil {
		response.ErrorWithMsg(c, "websocket not available")
		return
	}

	response.SuccessWithDetail(c, gin.H{
		"room_id":      roomID,
		"online_count": WsHub.GetRoomOnlineCount(roomID),
		"users":        WsHub.GetRoomOnlineUsers(roomID),
	}, "success")
}
