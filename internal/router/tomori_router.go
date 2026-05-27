package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

// SetupTomoriRouter 超级管理员（TakamatsuTomori）论坛管理路由
// 所有端点均需 JWT 认证 + 管理员角色校验（在 Handler 中完成）
func SetupTomoriRouter(r *gin.Engine) {
	tomoriApi := api.Api.TomoriApi

	protected := r.Group("/api/tomori")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		// 统计面板
		protected.GET("/stats", tomoriApi.GetStats)

		// 用户管理
		protected.GET("/users", tomoriApi.ListUsers)
		protected.POST("/user/ban", tomoriApi.BanUser)
		protected.POST("/user/role", tomoriApi.SetRole)
		protected.POST("/user/reset_password", tomoriApi.ResetPassword)
		protected.DELETE("/user/:id", tomoriApi.DeleteUser)

		// 帖子管理
		protected.GET("/posts", tomoriApi.ListPosts)
		protected.DELETE("/post/:id", tomoriApi.DeletePost)

		// 评论管理
		protected.GET("/comments", tomoriApi.ListComments)
		protected.DELETE("/comment/:id", tomoriApi.DeleteComment)

		// 文件管理
		protected.GET("/files", tomoriApi.ListFiles)
		protected.DELETE("/file/:id", tomoriApi.DeleteFile)

		// AI 智能体管理
		protected.GET("/agents", tomoriApi.ListAgents)
		protected.DELETE("/agent/:id", tomoriApi.DeleteAgent)

		// 聊天会话管理
		protected.GET("/chats", tomoriApi.ListChats)
		protected.DELETE("/chat/:chat_id", tomoriApi.DeleteChat)

		// 系统维护
		protected.GET("/blacklist", tomoriApi.ListBlacklist)
		protected.POST("/blacklist/clear", tomoriApi.ClearBlacklist)
	}
}
