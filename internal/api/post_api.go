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
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	post := &entity.Post{
		EnvInfo:    req.Env,
		UserId:     userId,
		Title:      req.Title,
		Cover:      req.Cover,
		Tags:       req.Tags,
		CategoryId: req.CategoryId,
		Keywords:   req.Keywords,
		Content:    req.Content,
		Public:     req.Public,
	}
	r, err := postService.Create(post)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, r, "post success")
}

func (pa *PostApi) Query(c *gin.Context) {
	if val := c.Param("id"); val != "" {
		pa.QueryOneById(c)
		return
	}
	response.ErrorWithMsg(c, "empty query")
}

func (pa *PostApi) QueryOneById(c *gin.Context) {
	sid := c.Param("id") // post id
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	viewerId := cl.UserId
	if sid == "" {
		response.ErrorWithMsg(c, "missing id")
		return
	}
	//fmt.Printf("current viewer id: %v", viewerId)
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	post, err := postService.GetPostByPostId(uint(id), viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if !post.Public && viewerId != post.UserId {
		post.Content = "this is private post"
	}
	response.SuccessWithDetail(c, post, "query success")
}

func (pa *PostApi) QueryAll(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	viewerId := cl.UserId
	page, pageSize := parsePagination(c)

	//fmt.Printf("current viewer id: %v", viewerId)

	postList, err := postService.GetAllPosts(page, pageSize, viewerId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	for i, item := range postList.Items {
		if !item.Public && viewerId != item.UserId {
			postList.Items[i].Content = "this is private post"
		}
	}
	response.SuccessWithDetail(c, postList, "query success")
}
func (pa *PostApi) ListPostsByUserId(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid user id")
		return
	}
	page, pageSize := parsePagination(c)
	postList, err := postService.GetPostsByUserId(uint(userId), page, pageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, postList, "query success")
}

func (pa *PostApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid id")
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	post, err := postService.GetPostByPostId(uint(id), 0)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	if post.UserId != cl.UserId {
		response.ErrorWithMsg(c, "you can't delete others' post")
		return
	}
	if err := postService.DeleteById(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithMsg(c, "delete success")
}

func (pa *PostApi) Like(c *gin.Context) {
	var req request.PostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := postService.Like(req.PostId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "like success")
}

func (pa *PostApi) Dislike(c *gin.Context) {
	var req request.PostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := postService.Dislike(req.PostId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "dislike success")
}

func (pa *PostApi) Favorite(c *gin.Context) {
	var req request.PostActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := postService.Favorite(req.PostId, cl.UserId); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "favorite success")
}

func (pa *PostApi) Share(c *gin.Context) {
	var req request.PostShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if err := postService.Share(req.PostId, cl.UserId, entity.ShareInfo{
		ShareTo:      req.ShareTo,
		ShareMessage: req.ShareMessage,
	}); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "share success")
}

func (pa *PostApi) Update(c *gin.Context) {
	var req request.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId
	post, err := postService.GetPostByPostId(req.Id, 0)
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
