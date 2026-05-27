package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupTomoriRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	tomoriApi := apis.TomoriApi

	protected := r.Group("/api/tomori")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.GET("/stats", tomoriApi.GetStats)
		protected.GET("/users", tomoriApi.ListUsers)
		protected.POST("/user/ban", tomoriApi.BanUser)
		protected.POST("/user/role", tomoriApi.SetRole)
		protected.POST("/user/reset_password", tomoriApi.ResetPassword)
		protected.DELETE("/user/:id", tomoriApi.DeleteUser)
		protected.GET("/posts", tomoriApi.ListPosts)
		protected.DELETE("/post/:id", tomoriApi.DeletePost)
		protected.GET("/comments", tomoriApi.ListComments)
		protected.DELETE("/comment/:id", tomoriApi.DeleteComment)
		protected.GET("/files", tomoriApi.ListFiles)
		protected.DELETE("/file/:id", tomoriApi.DeleteFile)
		protected.GET("/agents", tomoriApi.ListAgents)
		protected.DELETE("/agent/:id", tomoriApi.DeleteAgent)
		protected.GET("/chats", tomoriApi.ListChats)
		protected.DELETE("/chat/:id", tomoriApi.DeleteChat)
		protected.GET("/blacklist", tomoriApi.ListBlacklist)
		protected.POST("/blacklist/clear", tomoriApi.ClearBlacklist)
	}
}
