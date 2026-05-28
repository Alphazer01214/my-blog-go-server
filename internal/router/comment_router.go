package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupCommentRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	commentApi := apis.CommentApi

	//public := r.Group("/api")
	//{
	//	public.GET("/post/:id/comments", commentApi.ListCommentsByPostId)
	//	public.GET("/video/:id/comments", commentApi.ListCommentsByVideoId)
	//}

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.GET("/post/:id/comments", commentApi.ListCommentsByPostId)
		protected.GET("/video/:id/comments", commentApi.ListCommentsByVideoId)
		protected.POST("/comment", commentApi.Create)
		protected.DELETE("/comment/:id", commentApi.Delete)
		protected.POST("/comment/like", commentApi.Like)
		protected.POST("/comment/dislike", commentApi.Dislike)
	}
}
