package api

import (
	"fmt"
	"strconv"

	"blog.alphazer01214.top/internal/constant"
	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type CommentApi struct{}

func (ca *CommentApi) Create(c *gin.Context) {
	var req request.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	comment := &entity.Comment{
		EnvInfo:         req.Env,
		UserId:          cl.UserId,
		TargetType:      constant.TargetPost,
		TargetId:        req.PostId,
		Content:         req.Content,
		RootCommentId:   req.RootCommentId,
		ParentCommentId: req.ParentCommentId,
	}

	r, err := commentService.Create(comment)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, r, "comment success")
}

func (ca *CommentApi) ListCommentsByPostId(c *gin.Context) {
	postId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid post id")
		return
	}

	page, pageSize := parsePagination(c)
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	comments, total, err := commentService.GetCommentsByPostId(uint(postId), page, pageSize, cl.UserId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	msg := fmt.Sprintf("user %v: %v query success", cl.UserId, cl.Username)

	response.SuccessWithDetail(c, response.CommentList{
		Items:    comments,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, msg)
}

func (ca *CommentApi) ListCommentsByUserId(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	page, pageSize := parsePagination(c)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	userSetting, err := userService.GetSetting(uint(userId))
	if err != nil {
		response.ErrorWithMsg(c, "can't get user setting")
		return
	}
	if !userSetting.CommentPublic {
		response.SuccessWithDetail(c, response.CommentList{
			Items:    []*response.Comment{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
		}, "user comment is private")
		return
	}
	comments, total, err := commentService.GetCommentsByUserId(uint(userId), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, response.CommentList{
		Items:    comments,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, "query success")

}

func (ca *CommentApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid comment id")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := commentService.Delete(uint(id), cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "delete success")
}

func (ca *CommentApi) Like(c *gin.Context) {
	var req request.CommentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := commentService.Like(req.CommentId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "like success")
}

func (ca *CommentApi) Dislike(c *gin.Context) {
	var req request.CommentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := commentService.Dislike(req.CommentId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "dislike success")
}

func parsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20
	if rawPage := c.Query("page"); rawPage != "" {
		if p, err := strconv.Atoi(rawPage); err == nil && p > 0 {
			page = p
		}
	}
	if rawSize := c.Query("page_size"); rawSize != "" {
		if p, err := strconv.Atoi(rawSize); err == nil && p > 0 {
			if p > 100 {
				p = 100
			}
			pageSize = p
		}
	}
	return
}
