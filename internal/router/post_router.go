package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupPostRouter(r *gin.Engine, apis *api.Apis, userService *service.UserService) {
	postApi := apis.PostApi

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware(userService))
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
		protected.GET("/post/search", postApi.SearchPost)
		protected.POST("/post/ask", postApi.Ask)
	}
}
