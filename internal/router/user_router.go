package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/pkg/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func SetupUserRouter(r *gin.Engine, apis *api.Apis) {
	userApi := apis.UserApi
	postApi := apis.PostApi
	commentApi := apis.CommentApi

	publicAuth := r.Group("/api/auth")
	{
		publicAuth.POST("/register", userApi.Register)
		publicAuth.POST("/login", userApi.Login)
	}
	protectedAuth := r.Group("/api/auth")
	protectedAuth.Use(jwt.JWTAuthMiddleware())
	{
		protectedAuth.POST("/logout", userApi.Logout)
	}
	public := r.Group("/api")
	{
		public.GET("/user/:id", userApi.QueryUserById)
		public.GET("/user/:id/posts", postApi.ListPostsByUserId)
		public.GET("/user/:id/comments", commentApi.ListCommentsByUserId)
		public.GET("/user/:id/followers", userApi.GetFollowers)
		public.GET("/user/:id/following", userApi.GetFollowing)
		public.GET("/all_users", userApi.GetAllUsers)
	}

	protected := r.Group("/api")
	protected.Use(jwt.JWTAuthMiddleware(userService))
	{
		protected.POST("/update_password", userApi.UpdatePassword)
		protected.POST("/update_profile", userApi.UpdateProfile)
		protected.GET("/me", userApi.CurrentUser)
		protected.POST("/user/follow", userApi.Follow)
		protected.GET("/settings", userApi.GetSettings)
		protected.POST("/settings", userApi.UpdateSettings)
	}
}
