package api

import (
	"context"
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
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
	usr := &entity.User{
		Username: req.Username,
		Password: req.Password, // original password
		Admin:    false,
		Role:     entity.RoleNormalUser,
		Banned:   false,
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

	utils.SetRefreshTokenCookie(c, loginResponse.Token.RefreshToken, loginResponse.Token.RefreshTokenExpireTime)
	utils.SetAccessTokenCookie(c, loginResponse.Token.AccessToken, loginResponse.Token.AccessTokenExpireTime)
	c.Set("user_id", loginResponse.UserInfo.UserId)
	response.SuccessWithDetail(c, loginResponse, "login success")
}

func (u *UserApi) Logout(c *gin.Context) {
	refreshToken := utils.GetRefreshTokenCookie(c)
	accessToken := utils.GetAccessTokenCookie(c)
	if refreshToken == "" {
		response.ErrorWithMsg(c, "not login yet")
		return
	}
	if accessToken == "" {
		response.ErrorWithMsg(c, "not login yet")
		return
	}
	if err := utils.TokenJoinBlacklist(refreshToken); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	// 同时将 accessToken 加入黑名单，防止在 accessToken 过期前被滥用
	if err := utils.TokenJoinBlacklist(accessToken); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	claims, err := utils.ParseRefreshToken(refreshToken)
	if err == nil {
		global.GetRedis().Del(context.Background(), strconv.Itoa(int(claims.Id)))
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
		return
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
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	rp, err := userService.UpdateUserProfile(cl.UserId, req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "UserUpdate profile successful")
}

func (u *UserApi) CurrentUser(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	info, err := userService.GetUserInfoById(cl.UserId, 0)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, info, "Query current user successful")
}

func (u *UserApi) QueryUserById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	rp, err := userService.GetUserInfoById(uint(id), GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "Query user by id successful")
}

func (u *UserApi) GetAllUsers(c *gin.Context) {
	usrInfo, err := userService.GetAllUserInfo(GetUserId(c))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, usrInfo, "Query all users successful")
}

func (u *UserApi) Follow(c *gin.Context) {
	var req request.UserFollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	rp, err := userService.Follow(cl.UserId, req.FollowingId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "follow toggled")
}

func (u *UserApi) GetFollowers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := userService.GetFollowers(uint(id), GetUserId(c), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "query followers success")
}

func (u *UserApi) GetFollowing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := userService.GetFollowing(uint(id), GetUserId(c), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "query following success")
}

func (u *UserApi) GetSettings(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	setting, err := userService.GetSetting(cl.UserId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, response.UserSettingResponse{
		PostPublic:         setting.PostPublic,
		CommentPublic:      setting.CommentPublic,
		FollowListPublic:   setting.FollowListPublic,
		FollowerListPublic: setting.FollowerListPublic,
	}, "get settings success")
}

func (u *UserApi) UpdateSettings(c *gin.Context) {
	var req request.UserSettingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := userService.UpdateSetting(cl.UserId, req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "update settings success")
}
