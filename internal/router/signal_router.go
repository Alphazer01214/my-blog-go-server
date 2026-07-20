package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

// SetupSignalRouter 设置交易信号路由
func SetupSignalRouter(r *gin.Engine) {
	signalApi := api.SignalApi{}

	// 公开路由
	public := r.Group("/api")
	{
		public.GET("/signal/:id", signalApi.GetById)
		public.GET("/signals", signalApi.List)
		public.GET("/user/:id/signals", signalApi.ListByUserId)
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("/signal", signalApi.Create)
		protected.PUT("/signal/:id/close", signalApi.CloseSignal)
		protected.POST("/signal/:id/follow", signalApi.Follow)
		protected.DELETE("/signal/:id", signalApi.Delete)
		protected.GET("/signals/feed", signalApi.GetFeed)
	}
}
