package api

import (
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/request"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type PostApi struct{}

func (pa *PostApi) Create(c *gin.Context) {
	var req request.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	post := &entity.Post{
		EnvInfo:  req.Env,
		UserId:   userId,
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

func (pa *PostApi) Query(c *gin.Context) {
	if c.Query("id") != "" {
		pa.QueryOneById(c)
		return
	}

	pa.QueryAll(c)
}

func (pa *PostApi) QueryOneById(c *gin.Context) {
	sid := c.Query("id")
	if sid == "" {
		response.ErrorWithMsg(c, "missing id")
		return
	}
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
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
	}
	userId := cl.UserId
	if post.Public == false && userId != post.UserId {
		response.ErrorWithMsg(c, "post is private")
		return
	}
	response.SuccessWithDetail(c, post, "query success")
}

func (pa *PostApi) QueryAll(c *gin.Context) {
	posts, err := postService.QueryAll()
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	maxQuery := c.Query("max_query")
	if maxQuery != "" {
		limit, err := strconv.ParseUint(maxQuery, 10, 64)
		if err != nil {
			response.ErrorWithMsg(c, "invalid max_query")
			return
		}
		if int(limit) < len(posts) {
			posts = posts[:int(limit)]
		}
	}

	response.SuccessWithDetail(c, posts, "query success")
}

func (pa *PostApi) EvilQuery(c *gin.Context) {

}

func (pa *PostApi) Delete(c *gin.Context) {

}

func (pa *PostApi) Update(c *gin.Context) {
	var req request.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	post, err := postService.QueryOneById(req.Id)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if post.UserId != userId {
		response.ErrorWithMsg(c, "you can't edit others' post")
		return
	}

	if err := postService.Update(req.Id, req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, post, "update success")
}
