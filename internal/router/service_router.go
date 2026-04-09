package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupServiceRouter(r *gin.Engine) {
	aiApi := api.Api.AiApi

	// ai api
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.POST("/create_agent", aiApi.Create)
		protected.POST("/update_agent", aiApi.Update)
		protected.POST("/invoke_agent", aiApi.Invoke)
		protected.POST("/chat", aiApi.OnlineStreamChat)
	}
}
