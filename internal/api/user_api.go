package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserApi struct {
}

func (u *UserApi) Register(c *gin.Context) {
	var req request.UserRegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c)
		return
	}
	if !utils.IsPasswordValid(req.Password) {
		response.Error(c)
		return
	}
	//pass := utils.EncryptPassword(req.Password)
	usr := &entity.User{
		Username: req.Username,
		Password: req.Password, // 明文
	}
	rp, err := userService.Register(usr, req.Env)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "register success")
}

func (u *UserApi) Login(c *gin.Context) {
	var req request.UserLoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c)
		return
	}
	usr := &entity.User{
		Username: req.Username,
		Password: req.Password,
	}
	loginResponse, err := userService.Login(usr, req.Env)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	// 成功登录，把refresh token 和 access token 放进cookies
	utils.SetRefreshTokenCookie(c, loginResponse.Token.RefreshToken, loginResponse.Token.RefreshTokenExpireTime)
	utils.SetAccessTokenCookie(c, loginResponse.Token.AccessToken, loginResponse.Token.AccessTokenExpireTime)
	c.Set("user_id", loginResponse.UserInfo.ID)
	response.SuccessWithDetail(c, loginResponse, "login success")
}

func (u *UserApi) Logout(c *gin.Context) {
	// 从请求头获取 token
	//token := c.GetHeader("Authorization")
	//if token == "" {
	//	response.ErrorWithMsg(c, "token is required")
	//	return
	//}
	refreshToken := utils.GetRefreshTokenCookie(c)
	accessToken := utils.GetAccessTokenCookie(c)
	if refreshToken == "" {
		response.ErrorWithMsg(c, "not login yet")
		return
	}
	if accessToken == "" {
		response.ErrorWithMsg(c, "not login yet")
	}
	// 将 token 加入黑名单
	if err := utils.TokenJoinBlacklist(refreshToken); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	utils.RemoveRefreshTokenCookie(c)
	utils.RemoveAccessTokenCookie(c)

	response.SuccessWithMsg(c, "Logout successful")
}

func (u *UserApi) UpdatePassword(c *gin.Context) {
	var req request.UserUpdatePasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c)
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
	}
	userId := cl.UserId

	rp, err := userService.UpdateUserPassword(userId, req.OldPassword, req.NewPassword, req.Env)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "UserUpdate password successful")
}

func (u *UserApi) UpdateProfile(c *gin.Context) {
	var req request.UserUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c)
		return
	}

	rp, err := userService.UpdateUserProfile(req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "UserUpdate profile successful")
}

func (u *UserApi) QueryUserById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	rp, err := userService.GetUserById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "Query user by id successful")
}
