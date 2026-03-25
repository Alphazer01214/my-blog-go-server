package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupPostRouter(r *gin.Engine) {
	postApi := api.Api.PostApi

	public := r.Group("/api")
	{
		public.GET("/post/:id", postApi.QueryOneById)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.POST("/create", postApi.Create)
	}
}
