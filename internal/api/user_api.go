package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	response.SuccessWithDetail(c, loginResponse, "Login successful")
}

func (u *UserApi) Logout(c *gin.Context) {
	// 从请求头获取 token
	token := c.GetHeader("Authorization")
	if token == "" {
		response.ErrorWithMsg(c, "token is required")
		return
	}

	// 将 token 加入黑名单
	utils.AddToBlacklist(token)

	response.SuccessWithMsg(c, "Logout successful")
}

func (u *UserApi) TokenNext(c *gin.Context, user *entity.User) {
	if user.Banned {
		response.ErrorWithMsg(c, "user is banned")
	}
	baseClaims := request.BaseClaims{
		Id:       user.ID,
		Username: user.Username,
		RoleType: user.Role,
	}

	accessClaims := utils.GenerateAccessClaims(baseClaims)
	accessToken := utils.GenerateAccessTokenFromClaims(accessClaims)
	refreshToken := utils.GenerateAccessTokenFromClaims(accessClaims)

}

func (u *UserApi) UpdatePassword(c *gin.Context) {
	var req request.UserUpdatePasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c)
		return
	}

	rp, err := userService.UpdateUserPassword(req.Id, req.OldPassword, req.NewPassword, req.Env)
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

	rp, err := userService.UpdateUserProfile(&entity.User{
		Model:    gorm.Model{ID: req.Id},
		Username: req.NewUsername,
		Email:    req.NewEmail,
		Phone:    req.NewPhone,
		Bio:      req.NewBio,
		Avatar:   req.NewAvatar,
	}, req.Env)
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
