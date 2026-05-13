package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupCommentRouter(r *gin.Engine) {
	commentApi := api.Api.CommentApi

	public := r.Group("/api")
	{
		public.GET("/post/:id/comments", commentApi.List)
		public.GET("/comments/:id/replies", commentApi.ListReplies)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.POST("/comment", commentApi.Create)
		protected.DELETE("/comment/:id", commentApi.Delete)
		protected.POST("/comment/like", commentApi.Like)
		protected.POST("/comment/dislike", commentApi.Dislike)
	}
}
