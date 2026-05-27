package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupVideoRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	videoApi := apis.VideoApi

	public := r.Group("/api/video")
	{
		public.GET("/:id/stream", videoApi.Stream)
		public.GET("/:id/cover", videoApi.Cover)
	}

	protected := r.Group("/api/video")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.POST("/init", videoApi.InitUpload)
		protected.POST("/chunk", videoApi.UploadChunk)
		protected.POST("/merge", videoApi.Merge)
		protected.POST("/create", videoApi.Create)
		protected.GET("/:id", videoApi.Query)
		protected.GET("", videoApi.QueryAll)
		protected.GET("/user/:userId", videoApi.ListByUserId)
		protected.PUT("/:id", videoApi.Update)
		protected.DELETE("/:id", videoApi.Delete)
		protected.POST("/:id/like", videoApi.Like)
		protected.POST("/:id/dislike", videoApi.Dislike)
		protected.POST("/:id/favorite", videoApi.Favorite)
		protected.POST("/:id/share", videoApi.Share)
	}
}
