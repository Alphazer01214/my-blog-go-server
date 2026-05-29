package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	maxResults    = 5
	maxContentLen = 500 // 每篇帖子内容截断字符数
)

type SearchPostParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"` // 可选，覆盖默认限制
}

type SearchPost struct{}

func (sp *SearchPost) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params SearchPostParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	query := strings.TrimSpace(params.Query)
	if query == "" {
		return "搜索关键词不能为空，请提供有效的搜索词。", nil
	}

	// 转义 LIKE 通配符，防止用户输入中的 % _ 破坏查询
	safeQuery := escapeLikePattern(query)

	limit := maxResults
	if params.Limit > 0 && params.Limit <= 10 {
		limit = params.Limit
	}

	var posts []entity.Post
	if err := global.GetDB().WithContext(ctx).
		Where("title ILIKE ? OR content ILIKE ?", "%"+safeQuery+"%", "%"+safeQuery+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error; err != nil {
		return "", fmt.Errorf("数据库查询失败: %w", err)
	}

	if len(posts) == 0 {
		return fmt.Sprintf("未找到与「%s」相关的帖子。建议：\n- 尝试更简短的关键词\n- 检查是否有错别字\n- 使用更通用的词汇重试", query), nil
	}

	// 构建结构化结果
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## 搜索结果（共 %d 篇，关键词：「%s」）\n\n", len(posts), query))
	b.WriteString("请基于以下帖子内容回答用户问题。如果帖子内容不足以回答，请如实告知。\n\n")

	for i, post := range posts {
		b.WriteString(fmt.Sprintf("---\n### 帖子 %d\n", i+1))
		b.WriteString(fmt.Sprintf("- **标题**: %s\n", post.Title))
		b.WriteString(fmt.Sprintf("- **ID**: %d\n", post.ID))
		b.WriteString(fmt.Sprintf("- **链接**: http://localhost:5173/post/%d\n", post.ID))

		// 如果有时间字段，输出发布时间
		if !post.CreatedAt.IsZero() {
			b.WriteString(fmt.Sprintf("- **发布时间**: %s\n", post.CreatedAt.Format(time.DateTime)))
		}

		// 内容截断，避免上下文过长
		content := truncateString(post.Content, maxContentLen)
		b.WriteString(fmt.Sprintf("- **内容摘要**:\n%s\n\n", content))
	}

	return b.String(), nil
}

func (sp *SearchPost) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_post",
		Desc: "在博客网站中搜索帖子。当用户询问与博客内容相关的问题时，使用此工具检索相关帖子以提供准确回答。" +
			"适用场景：用户想查找特定主题的帖子、询问某篇文章内容、寻找博客上的信息等。" +
			"不适用场景：用户的问题与博客帖子无关、请求通用知识问答、或要求修改/删除帖子。",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     "string",
					Desc:     "搜索关键词。应提取用户问题中的核心名词或短语，避免过长的句子。例如：用户问「有没有关于 Docker 部署的文章」，query 应为「Docker 部署」。",
					Required: true,
				},
				"limit": {
					Type: "integer",
					Desc: fmt.Sprintf("返回结果数量上限，默认 %d，最大 10。", maxResults),
				},
			},
		),
	}, nil
}

// escapeLikePattern 转义 SQL LIKE 通配符
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// truncateString 按 rune 截断字符串，避免截断多字节字符
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	// 在截断位置向后找到一个合理的断句点（标点或空格）
	cutPos := maxLen
	for i := maxLen; i < maxLen+20 && i < len(runes); i++ {
		if runes[i] == '。' || runes[i] == '！' || runes[i] == '？' ||
			runes[i] == '.' || runes[i] == '!' || runes[i] == '?' ||
			runes[i] == '，' || runes[i] == ',' || runes[i] == ' ' {
			cutPos = i + 1
			break
		}
	}
	return string(runes[:cutPos]) + "..."
}

func NewSearchPostTool() *SearchPost {
	return &SearchPost{}
}
