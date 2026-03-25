package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type PostApi struct {
}

func (pa *PostApi) Create(c *gin.Context) {
	var req request.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}
	post := &entity.Post{
		EnvInfo:  req.Env,
		AuthorId: req.Creator.ID,
		Title:    req.Title,
		Cover:    req.Cover,
		Category: req.Category,
		Keywords: req.Keywords,
		Content:  req.Content,
		Public:   req.Public,
	}
	if err := postService.Create(post); err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to create post",
		})
		return
	}

	response.SuccessWithMsg(c, "post success")
}

func (pa *PostApi) QueryOneById(c *gin.Context) {
	sid := c.Param("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	post, err := postService.QueryOneById(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if post.Public == false {
		response.ErrorWithMsg(c, "post is private")
		return
	}
	response.SuccessWithDetail(c, post, "query success")
}
