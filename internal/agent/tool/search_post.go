package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type SearchPostParams struct {
	Query string `json:"query"`
}

type SearchPost struct {
}

func (sp *SearchPost) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params SearchPostParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}
	var postEntities []entity.Post

	if err := global.GetDB().WithContext(ctx).Where("title ILIKE ? OR content ILIKE ?", "%"+params.Query+"%", "%"+params.Query+"%").Find(&postEntities).Error; err != nil {
		return "", err
	}
	if len(postEntities) == 0 {
		return "未检索到相关帖子", nil
	}

	resp := "检索到以下帖子：\n"
	for idx, post := range postEntities {
		title := post.Title
		content := post.Content
		prompt := fmt.Sprintf("第%d个帖子：标题：%s\n内容：%s\n", idx+1, title, content)
		resp += prompt
	}

	return resp, nil

}

func (sp *SearchPost) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_post",
		Desc: "按照关键词搜索帖子",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     "string",
					Desc:     "当需要查找该网站上的帖子时，使用此工具。",
					Required: true,
				},
			},
		),
	}, nil
}

func NewSearchPostTool() *SearchPost {
	return &SearchPost{}
}
