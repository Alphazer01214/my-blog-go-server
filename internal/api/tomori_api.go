package api

import (
	"errors"
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

// TomoriApi 超级管理员（TakamatsuTomori）论坛管理 API
type TomoriApi struct {
	tomoriService *service.TomoriService
}

// authorizeTomori 验证当前用户是否为 TakamatsuTomori 超级管理员
func authorizeTomori(c *gin.Context) (request.AccessClaims, error) {
	cl, err := Authorize(c)
	if err != nil {
		return cl, err
	}
	if cl.RoleType != entity.RoleTakamatsuTomori {
		return cl, errors.New("tomori access only")
	}
	return cl, nil
}

// ======================== 统计面板 ========================

// GetStats GET /api/tomori/stats - 获取论坛基础统计数据
func (ta *TomoriApi) GetStats(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	stats, err := ta.tomoriService.GetStats(c.Request.Context())
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, stats, "ok")
}

// ======================== 用户管理 ========================

// ListUsers GET /api/tomori/users - 分页查询所有用户
func (ta *TomoriApi) ListUsers(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// BanUser POST /api/tomori/user/ban - 封禁/解封用户 {user_id, banned}
func (ta *TomoriApi) BanUser(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	var req request.TomoriBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	if err := ta.tomoriService.BanUser(c.Request.Context(), req.UserId, req.Banned); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// SetRole POST /api/tomori/user/role - 修改用户角色 {user_id, role}
func (ta *TomoriApi) SetRole(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	var req request.TomoriRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	if err := ta.tomoriService.SetRole(c.Request.Context(), req.UserId, req.Role); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ResetPassword POST /api/tomori/user/reset_password - 重置用户密码 {user_id, new_password}
func (ta *TomoriApi) ResetPassword(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	var req request.TomoriResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	if err := ta.tomoriService.ResetPassword(c.Request.Context(), req.UserId, req.NewPassword); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// DeleteUser DELETE /api/tomori/user/:id - 删除用户（级联清理 profile/setting/follows）
func (ta *TomoriApi) DeleteUser(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	if err := ta.tomoriService.ForceDeleteUser(c.Request.Context(), uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== 帖子管理 ========================

// ListPosts GET /api/tomori/posts - 分页查询所有帖子（含私密）
func (ta *TomoriApi) ListPosts(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListAllPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// DeletePost DELETE /api/tomori/post/:id - 删除任意帖子
func (ta *TomoriApi) DeletePost(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid post id")
		return
	}
	if err := ta.tomoriService.ForceDeletePost(c.Request.Context(), uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== 评论管理 ========================

// ListComments GET /api/tomori/comments - 分页查询所有评论
func (ta *TomoriApi) ListComments(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListAllComments(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// DeleteComment DELETE /api/tomori/comment/:id - 删除任意评论（含级联子评论）
func (ta *TomoriApi) DeleteComment(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid comment id")
		return
	}
	if err := ta.tomoriService.ForceDeleteComment(c.Request.Context(), uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== 文件管理 ========================

// ListFiles GET /api/tomori/files - 分页查询所有文件（含私密）
func (ta *TomoriApi) ListFiles(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListAllFiles(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// DeleteFile DELETE /api/tomori/file/:id - 删除任意文件
func (ta *TomoriApi) DeleteFile(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid file id")
		return
	}
	if err := ta.tomoriService.ForceDeleteFile(c.Request.Context(), uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== AI 智能体管理 ========================

// ListAgents GET /api/tomori/agents - 查看所有用户的智能体
func (ta *TomoriApi) ListAgents(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	rp, err := ta.tomoriService.ListAllAgents(c.Request.Context())
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// DeleteAgent DELETE /api/tomori/agent/:id - 删除任意智能体
func (ta *TomoriApi) DeleteAgent(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid agent id")
		return
	}
	if err := ta.tomoriService.ForceDeleteAgent(c.Request.Context(), uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== 聊天会话管理 ========================

// ListChats GET /api/tomori/chats - 分页查询所有聊天会话
func (ta *TomoriApi) ListChats(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListAllChats(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// DeleteChat DELETE /api/tomori/chat/:chat_id - 删除任意聊天会话
func (ta *TomoriApi) DeleteChat(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	chatId := c.Param("chat_id")
	if chatId == "" {
		response.ErrorWithMsg(c, "missing chat_id")
		return
	}
	if err := ta.tomoriService.ForceDeleteChat(c.Request.Context(), chatId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}

// ======================== 系统维护 ========================

// ListBlacklist GET /api/tomori/blacklist - 分页查询 Token 黑名单
func (ta *TomoriApi) ListBlacklist(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	page, pageSize := parsePagination(c)
	rp, err := ta.tomoriService.ListBlacklist(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "ok")
}

// ClearBlacklist POST /api/tomori/blacklist/clear - 清空 Token 黑名单
func (ta *TomoriApi) ClearBlacklist(c *gin.Context) {
	if _, err := authorizeTomori(c); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if err := ta.tomoriService.ClearBlacklist(c.Request.Context()); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "ok")
}
