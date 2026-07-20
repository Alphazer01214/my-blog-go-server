package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

// SetupWatchlistRouter 设置关注列表路由
func SetupWatchlistRouter(r *gin.Engine) {
	watchlistApi := api.WatchlistApi{}

	// 公开路由
	public := r.Group("/api")
	{
		public.GET("/watchlist/public", watchlistApi.ListPublic)
		public.GET("/user/:id/watchlists", watchlistApi.List)
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("/watchlist", watchlistApi.Create)
		protected.GET("/watchlist/:id", watchlistApi.GetById)
		protected.PUT("/watchlist/:id", watchlistApi.Update)
		protected.DELETE("/watchlist/:id", watchlistApi.Delete)
		protected.POST("/watchlist/:id/stock", watchlistApi.AddStock)
		protected.DELETE("/watchlist/:id/stock/:symbol", watchlistApi.RemoveStock)
	}
}
