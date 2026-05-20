package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupPostRouter(r *gin.Engine) {
	postApi := api.Api.PostApi

	//public := r.Group("/api")
	//{
	//
	//}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/post/:id", postApi.Query)
		protected.GET("/post", postApi.QueryAll)
		protected.POST("/create", postApi.Create)
		protected.POST("/update", postApi.Update)
		protected.DELETE("/post/:id", postApi.Delete)
		protected.POST("/post/like", postApi.Like)
		protected.POST("/post/dislike", postApi.Dislike)
		protected.POST("/post/favorite", postApi.Favorite)
		protected.POST("/post/share", postApi.Share)
	}
}
