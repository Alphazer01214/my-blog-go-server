package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

// SetupStockRouter 设置股票讨论区路由
func SetupStockRouter(r *gin.Engine) {
	stockApi := api.StockApi{}

	// 公开路由
	public := r.Group("/api/stock")
	{
		public.GET("/:symbol/board", stockApi.GetBoard)
		public.GET("/:symbol/info", stockApi.GetInfo)
	}

	// 需要认证的路由
	protected := r.Group("/api/stock")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("/:symbol/board", stockApi.CreatePost)
	}
}
