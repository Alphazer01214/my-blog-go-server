package router

import (
	"blog.alphazer01214.top/internal/api"
	"blog.alphazer01214.top/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupUserRouter 设置用户相关路由
func SetupUserRouter(r *gin.Engine) {
	userApi := api.Api.UserApi

	// 公开路由（无需认证）
	publicAuth := r.Group("/api/auth")
	{
		publicAuth.POST("/register", userApi.Register)
		publicAuth.POST("/login", userApi.Login)
	}
	protectedAuth := r.Group("/api/auth")
	protectedAuth.Use(middleware.JWTAuthMiddleware())
	{
		protectedAuth.POST("/logout", userApi.Logout) // 暂时的
	}
	public := r.Group("/api")
	{
		public.GET("/user/:id", userApi.QueryUserById)
	}

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.POST("/update_password", userApi.UpdatePassword)
		protected.POST("/update_profile", userApi.UpdateProfile)
	}

}
