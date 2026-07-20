package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

// SetupWsRouter 设置 WebSocket 路由
func SetupWsRouter(r *gin.Engine) {
	wsApi := api.WsApi{}

	// WebSocket 连接（需要认证）
	ws := r.Group("/api/ws")
	ws.Use(jwt.JWTAuthMiddleware())
	{
		ws.GET("/connect", wsApi.HandleWebSocket)
	}

	// 在线状态查询（公开）
	r.GET("/api/ws/status", wsApi.GetOnlineStatus)
	r.GET("/api/ws/status/:room_id", wsApi.GetRoomOnlineStatus)
}
