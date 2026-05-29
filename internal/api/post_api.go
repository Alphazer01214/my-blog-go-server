package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
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
		EnvInfo:       req.Env,
		UserId:        userId,
		Title:         req.Title,
		Cover:         req.Cover,
		Tags:          req.Tags,
		Category:      req.Category,
		Keywords:      req.Keywords,
		Content:       req.Content,
		Public:        req.Public,
		ForbidComment: req.ForbidComment,
		ForbidShare:   req.ForbidShare,
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

func (pa *PostApi) SearchPost(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, "auth failed")
		return
	}
	viewerId := cl.UserId

	var req request.PostSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	page, pageSize := parsePagination(c)

	list, err := postService.SearchPost(&req, page, pageSize, viewerId)
	if err != nil {
		response.ErrorWithMsg(c, "query failed")
		return
	}

	for i, item := range list.Items {
		if !item.Public && viewerId != item.UserId {
			list.Items[i].Content = "this is private post"
		}
	}

	response.SuccessWithDetail(c, list, "query success")
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
	post, err := postService.GetPostEntityById(uint(id))
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
	post, err := postService.GetPostEntityById(req.Id)
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
	updated, err := postService.GetPostByPostId(req.Id, userId)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, updated, "update success")
}

func (pa *PostApi) Ask(c *gin.Context) {
	ctx := context.Background()

	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	userId := cl.UserId

	var req request.PostAskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, "invalid request")
		return
	}

	post, err := postService.GetPostEntityById(req.PostId)
	if err != nil {
		response.ErrorWithMsg(c, "post not found")
		return
	}

	chatId := req.ChatId
	if chatId == "" {
		response.ErrorWithMsg(c, "need provide uuid")
		return
	}

	var userPrompt string
	switch req.Mode {
	case "summarize":
		userPrompt = fmt.Sprintf("请总结以下文章：\n\n标题：%s\n\n文章内容：\n%s", post.Title, post.Content)
	case "selected":
		userPrompt = fmt.Sprintf("基于以下文章回答用户问题：\n\n标题：%s\n\n文章内容：\n%s\n\n用户选中文本：\n%s\n\n用户问题：%s",
			post.Title, post.Content, req.SelectedText, req.Prompt)
	default: // "ask"
		userPrompt = fmt.Sprintf("基于以下文章回答用户问题：\n\n标题：%s\n\n文章内容：\n%s\n\n用户问题：%s",
			post.Title, post.Content, req.Prompt)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "no-cache")

	writeSSE := func(payload interface{}) error {
		j, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", j); err != nil {
			return err
		}
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
		return nil
	}

	pushStream := func(data string) error {
		resp := map[string]interface{}{
			"chat_id": chatId,
			"content": data,
			"status":  true,
		}
		return writeSSE(resp)
	}

	tmpReq := &request.TmpChatRequest{
		BaseUrl: global.GetConfig().LLM.BaseUrl,
		ApiKey:  global.GetConfig().LLM.ApiKey,
		Model:   global.GetConfig().LLM.ModelName,
		Prompt:  userPrompt,
	}

	if err := aiService.TmpToolCallingStreamChat(ctx, userId, chatId, tmpReq, pushStream); err != nil {
		resp := map[string]interface{}{
			"chat_id": chatId,
			"content": "",
			"status":  false,
			"message": err.Error(),
		}
		_ = writeSSE(resp)
		return
	}

	_ = writeSSE(map[string]interface{}{
		"chat_id": chatId,
		"content": "",
		"status":  true,
		"message": "done",
	})
}

func (pa *PostApi) GetFavorites(c *gin.Context) {
	cl, err := Authorize(c)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	var req request.PostFavoriteListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	favorites, err := postService.GetFavoritesByUserId(cl.UserId, req.Page, req.PageSize)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithDetail(c, favorites, "query favorites success")
}
