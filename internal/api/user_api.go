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
	usr := &entity.User{
		Username: req.Username,
		Password: req.Password,
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
	}
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
	info, err := userService.GetUserInfoById(cl.UserId)
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
	rp, err := userService.GetUserInfoById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "Query user by id successful")
}

func (u *UserApi) GetAllUsers(c *gin.Context) {
	scnt := c.Query("tok_k")
	_, err := strconv.ParseInt(scnt, 10, 64)
	if err != nil {
	}
	usrInfo, err := userService.GetAllUserInfo()
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, usrInfo, "Query all users successful")
}

func (u *UserApi) QueryUserPosts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	posts, err := userService.GetPostsByUser(uint(id), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, posts, "query user posts success")
}

func (u *UserApi) QueryUserComments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)
	comments, total, err := userService.GetCommentsByUser(uint(id), viewerId, page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, response.CommentList{
		Items:    comments,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, "query user comments success")
}
