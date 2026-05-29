package jwt

import (
	"errors"
	"fmt"

	"blog.alphazer01214.top/internal/global"
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
		fmt.Println("================JWT auth middleware================")
		accessToken := utils.GetAccessTokenCookie(c)
		refreshToken := utils.GetRefreshTokenCookie(c)

		if yes, err := utils.IsTokenBlacklisted(refreshToken); err == nil && yes {
			// 在 blacklist
			fmt.Printf("[JWT auth middleware] blacklist token: %v \n", refreshToken)
			utils.RemoveRefreshTokenCookie(c)
			utils.RemoveAccessTokenCookie(c)
			response.ErrorAuth(c, "Invalid token")
			c.Abort()
			return
		} else if err != nil {
			response.ErrorWithMsg(c, err.Error())
			c.Abort()
			return
		}

		// 检查 accessToken 黑名单
		if yes, err := utils.IsTokenBlacklisted(accessToken); err == nil && yes {
			fmt.Printf("[JWT auth middleware] blacklist access token: %v \n", accessToken)
			utils.RemoveRefreshTokenCookie(c)
			utils.RemoveAccessTokenCookie(c)
			response.ErrorAuth(c, "Invalid token")
			c.Abort()
			return
		}

		// 先解析 access token, 成功则设置claims, 过期就使用 refresh token 刷新 access token，否则报错
		claims, err := utils.ParseAccessToken(accessToken)
		if err != nil {
			// if access token expired
			if errors.Is(err, jwt.ErrTokenExpired) {
				fmt.Printf("expired access token: %v", accessToken)
				refreshClaims, err := utils.ParseRefreshToken(refreshToken)
				if err != nil {
					utils.RemoveRefreshTokenCookie(c)
					utils.RemoveAccessTokenCookie(c)
					response.ErrorAuth(c, "Token invalid or expired")
					c.Abort()
					return
				}
				user, err := service.Service.UserService.GetUserInfoById(refreshClaims.Id, 0)
				if err != nil {
					utils.RemoveRefreshTokenCookie(c)
					utils.RemoveAccessTokenCookie(c)
					response.ErrorAuth(c, "User not exists")
					c.Abort()
					return
				}
				accessClaims := utils.GenerateAccessClaims(request.BaseClaims{
					UserId:   user.UserId,
					Username: user.Username,
					RoleType: user.Role,
				})
				accessToken, err := utils.GenerateAccessTokenFromClaims(accessClaims)
				if err != nil {
					utils.RemoveRefreshTokenCookie(c)
					utils.RemoveAccessTokenCookie(c)
					response.ErrorAuth(c, "Token generation failed")
					c.Abort()
					return
				}
				utils.SetAccessTokenCookie(c, accessToken, global.GetConfig().JWT.AccessTokenExpireTime)

				c.Set("claims", accessClaims)
				c.Set("user_id", accessClaims.UserId)
				c.Next()
				return
			}

			utils.RemoveRefreshTokenCookie(c)
			utils.RemoveAccessTokenCookie(c)
			response.ErrorAuth(c, "Authorization failed: "+err.Error())
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserId)
		fmt.Printf("[JWT auth middleware] claims(base claims, register claims): %v\n", claims)
		c.Next()
	}
}
