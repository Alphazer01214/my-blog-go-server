package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupFileRouter(r *gin.Engine) {
	fileApi := api.Api.FileApi

	public := r.Group("/api")
	{
		public.GET("/files/:id", fileApi.GetFile)
		public.GET("/files/:id/download", fileApi.DownloadFile)
		public.GET("/files", fileApi.ListFiles)
		public.GET("/user/:id/files", fileApi.GetFilesByUserId)
		public.GET("/upload/status", fileApi.GetUploadStatus)
	}

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("/upload/init", fileApi.InitUpload)
		protected.PUT("/upload/chunk", fileApi.UploadChunk)
		protected.POST("/upload/complete", fileApi.CompleteUpload)
		protected.POST("/upload/abort", fileApi.AbortUpload)
		protected.DELETE("/files/:id", fileApi.DeleteFile)
	}
}
