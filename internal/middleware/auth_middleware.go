package middleware

import (
	"errors"
	"strconv"

	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := utils.GetAccessTokenCookie(c)
		refreshToken := utils.GetRefreshTokenCookie(c)

		if ok, err := utils.IsTokenBlacklisted(refreshToken); err != nil || !ok {
			utils.RemoveRefreshTokenCookie(c)
			response.ErrorAuth(c, "Invalid token")
			c.Abort()
			return
		}

		// 先解析 access token, 成功则设置claims, 过期就使用 refresh token 刷新 access token，否则报错
		claims, err := utils.ParseAccessToken(accessToken)
		if err != nil {
			// if access token expired
			if errors.Is(err, jwt.ErrTokenExpired) {
				refreshClaims, err := utils.ParseRefreshToken(refreshToken)
				if err != nil {
					utils.RemoveRefreshTokenCookie(c)
					response.ErrorAuth(c, "Token invalid or expired")
					c.Abort()
					return
				}
				user, err := service.Service.UserService.GetUserInstanceById(refreshClaims.Id)
				if err != nil {
					utils.RemoveRefreshTokenCookie(c)
					response.ErrorAuth(c, "User not exists")
					c.Abort()
					return
				}
				accessClaims := utils.GenerateAccessClaims(request.BaseClaims{
					Id:       user.ID,
					Username: user.Username,
					RoleType: user.Role,
				})
				accessToken := utils.GenerateAccessTokenFromClaims(accessClaims)
				c.Header("access-token", accessToken)
				c.Header("access-expire-at", strconv.FormatInt(accessClaims.ExpiresAt.Unix(), 10))

				c.Set("claims", accessClaims)
				c.Next()
				return
			}

			utils.RemoveRefreshTokenCookie(c)
			response.ErrorAuth(c, "Token access token")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()

	}
}
