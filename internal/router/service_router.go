package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupServiceRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	aiApi := apis.AiApi
	marketApi := apis.MarketApi

	public := r.Group("/api")
	{
		public.GET("/market", marketApi.GetIndices)
		public.GET("/market/history", marketApi.GetHistory)
	}

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.POST("/agents", aiApi.Create)
		protected.PUT("/agents/:id", aiApi.Update)
		protected.GET("/agents", aiApi.QueryAgents)
		protected.GET("/agent/:id", aiApi.QueryAgentById)
		protected.POST("/invoke_agent", aiApi.Invoke)
		protected.POST("/chat", aiApi.OnlineStreamChat)
		protected.GET("/chats", aiApi.ListChats)
		protected.GET("/chat/:session_id", aiApi.GetChatSession)
	}
}
