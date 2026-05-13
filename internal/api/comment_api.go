package api

import (
	"strconv"

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
		PostId:          req.PostId,
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

func (ca *CommentApi) List(c *gin.Context) {
	postId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid post id")
		return
	}

	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)

	comments, total, err := commentService.QueryByPost(uint(postId), page, pageSize, viewerId)
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

func (ca *CommentApi) ListReplies(c *gin.Context) {
	rootId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid comment id")
		return
	}

	page, pageSize := parsePagination(c)
	viewerId := GetUserId(c)

	comments, total, err := commentService.QueryReplies(uint(rootId), page, pageSize, viewerId)
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
