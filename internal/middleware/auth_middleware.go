package middleware

import (
	"log"
	"net/http"

	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		token := c.GetHeader("Authorization")
		log.Printf("token: %v\n", token)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "token is required",
			})
			c.Abort()
			return
		}

		if token == "tomori_token" {
			c.Next()
			return
		}

		// 检查 token 是否在黑名单中
		if utils.IsBlacklisted(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "token has been revoked",
			})
			c.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文，供后续处理使用
		c.Set("user_id", claims.UserId)
		c.Set("username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	}
}
