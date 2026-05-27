package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupFileRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	fileApi := apis.FileApi

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.POST("/upload/init", fileApi.InitUpload)
		protected.POST("/upload/chunk", fileApi.UploadChunk)
		protected.POST("/upload/complete", fileApi.CompleteUpload)
		protected.GET("/upload/:uploadId/status", fileApi.GetUploadStatus)
		protected.POST("/upload/:uploadId/abort", fileApi.AbortUpload)
		protected.GET("/file/:id", fileApi.GetFile)
		protected.GET("/file/:id/download", fileApi.DownloadFile)
		protected.GET("/files", fileApi.ListFiles)
		protected.GET("/user/:id/files", fileApi.GetFilesByUserId)
		protected.DELETE("/file/:id", fileApi.DeleteFile)
	}
}
