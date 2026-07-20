package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"blog.alphazer01214.top/internal/global"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NewsSearchTool 财经新闻搜索工具
type NewsSearchTool struct{}

// NewsSearchParams 参数
type NewsSearchParams struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
}

// NewsItem 新闻条目
type NewsItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Date    string `json:"date"`
	Source  string `json:"source"`
}

func NewNewsSearchTool() *NewsSearchTool {
	return &NewsSearchTool{}
}

func (ns *NewsSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "news_search",
		Desc: `搜索财经新闻和市场资讯。当用户询问最新财经新闻、市场动态、公司公告、行业资讯时使用此工具。
搜索范围包括：股票新闻、宏观经济、行业动态、公司公告等。`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     "string",
					Desc:     "搜索关键词，如 '茅台财报'、'半导体行业'、'央行降息' 等",
					Required: true,
				},
				"limit": {
					Type: "integer",
					Desc: "返回结果数量，默认 5，最大 10",
				},
			},
		),
	}, nil
}

func (ns *NewsSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params NewsSearchParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	query := strings.TrimSpace(params.Query)
	if query == "" {
		return "搜索关键词不能为空", nil
	}

	limit := 5
	if params.Limit > 0 && params.Limit <= 10 {
		limit = params.Limit
	}

	// 使用百度千帆搜索 API（复用 web_search 的 API）
	apiKey := global.GetConfig().Tools.WebSearch.ApiKey
	if apiKey == "" {
		return "新闻搜索 API 未配置", nil
	}

	results, err := ns.searchNews(ctx, apiKey, query, limit)
	if err != nil {
		return fmt.Sprintf("新闻搜索失败: %v", err), nil
	}

	if len(results) == 0 {
		return fmt.Sprintf("未找到与「%s」相关的财经新闻", query), nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## 财经新闻搜索结果（关键词：「%s」）\n\n", query))
	b.WriteString("请根据以下新闻回答用户问题。\n\n")

	for i, news := range results {
		b.WriteString(fmt.Sprintf("---\n### %d. %s\n", i+1, news.Title))
		if news.Source != "" {
			b.WriteString(fmt.Sprintf("- **来源**: %s\n", news.Source))
		}
		if news.Date != "" {
			b.WriteString(fmt.Sprintf("- **日期**: %s\n", news.Date))
		}
		if news.Content != "" {
			content := news.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			b.WriteString(fmt.Sprintf("- **摘要**: %s\n", content))
		}
		if news.URL != "" {
			b.WriteString(fmt.Sprintf("- **链接**: %s\n", news.URL))
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (ns *NewsSearchTool) searchNews(ctx context.Context, apiKey, query string, limit int) ([]NewsItem, error) {
	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"content": query + " 财经新闻", "role": "user"},
		},
		"search_source": "baidu_search_v2",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		"https://qianfan.baidubce.com/v2/ai_search/web_search",
		strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp struct {
		References []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			URL     string `json:"url"`
			Date    string `json:"date"`
		} `json:"references"`
	}

	if err := json.Unmarshal(respBytes, &searchResp); err != nil {
		return nil, err
	}

	var results []NewsItem
	for i, ref := range searchResp.References {
		if i >= limit {
			break
		}
		results = append(results, NewsItem{
			Title:   ref.Title,
			Content: ref.Content,
			URL:     ref.URL,
			Date:    ref.Date,
		})
	}

	return results, nil
}
