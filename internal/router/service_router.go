package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupServiceRouter(r *gin.Engine) {
	aiApi := api.Api.AiApi
	marketApi := api.Api.MarketApi

	public := r.Group("/api")
	{
		public.GET("/market", marketApi.GetIndices)
		public.GET("/market/history", marketApi.GetHistory)
	}

	// ai api
	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware())
	{
		protected.POST("/create_agent", aiApi.Create)
		protected.POST("/update_agent", aiApi.Update)
		protected.POST("/invoke_agent", aiApi.Invoke)
		protected.POST("/chat", aiApi.OnlineStreamChat)
		protected.GET("/agents", aiApi.QueryAgents)
		protected.GET("/agent/:id", aiApi.QueryAgentById)
		protected.GET("/chats", aiApi.ListChats)
		protected.GET("/chat/:chat_id", aiApi.GetChatSession)
	}
}
