package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupVideoRouter(r *gin.Engine) {
	videoApi := api.Api.VideoApi

	upload := r.Group("/api/video")
	upload.Use(jwt.JWTAuthMiddleware())
	{
		upload.POST("/upload/init", videoApi.InitUpload)
		upload.POST("/upload/chunk", videoApi.UploadChunk)
		upload.POST("/upload/merge", videoApi.Merge)
	}

	protected := r.Group("/api/video")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("", videoApi.Create)
		protected.GET("/:id", videoApi.Query)
		protected.GET("/list", videoApi.QueryAll)
		protected.PUT("/:id", videoApi.Update)
		protected.DELETE("/:id", videoApi.Delete)
		protected.POST("/like", videoApi.Like)
		protected.POST("/dislike", videoApi.Dislike)
		protected.POST("/favorite", videoApi.Favorite)
		protected.POST("/share", videoApi.Share)
	}

	public := r.Group("/api")
	{
		public.GET("/user/:id/videos", videoApi.ListByUserId)
	}

	r.GET("/api/video/:id/stream", videoApi.Stream)
	r.GET("/api/video/:id/cover", videoApi.Cover)
}
