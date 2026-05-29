package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WebSearch struct {
	ApiKey string `json:"api_key" yaml:"api_key"`
}

type WebSearchParams struct {
	Query string `json:"query"`
}

type baiduSearchRequest struct {
	Messages     []baiduSearchMessage `json:"messages"`
	SearchSource string               `json:"search_source"`
}

type baiduSearchMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type baiduSearchResponse struct {
	References []baiduSearchReference `json:"references"`
	RequestID  string                 `json:"request_id"`
}

type baiduSearchReference struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	URL       string `json:"url"`
	Date      string `json:"date"`
	Type      string `json:"type"`
	WebAnchor string `json:"web_anchor"`
}

func (ws *WebSearch) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "当用户询问实时信息（如今天的股票行情、天气、最新新闻等）时，必须使用此工具进行搜索。现在是" + time.Now().Format("2006-01-02 15:04:05"),
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     "string",
					Desc:     "搜索关键词",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun 需要把结果变为 prompt 返回
func (ws *WebSearch) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params WebSearchParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	reqBody := baiduSearchRequest{
		Messages: []baiduSearchMessage{
			{Content: params.Query, Role: "user"},
		},
		SearchSource: "baidu_search_v2",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://qianfan.baidubce.com/v2/ai_search/web_search", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ws.ApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var searchResp baiduSearchResponse
	if err := json.Unmarshal(respBytes, &searchResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(searchResp.References) == 0 {
		return "未找到相关结果", nil
	}

	var result string = "以下是搜索结果，请根据这些结果回答用户提问：\n"
	for i, ref := range searchResp.References {
		result += fmt.Sprintf("[%d] %s\n", i+1, ref.Title)
		if ref.Content != "" {
			result += fmt.Sprintf("   内容: %s\n", ref.Content)
		}
		if ref.URL != "" {
			result += fmt.Sprintf("   链接: %s\n", ref.URL)
		}
		if ref.Date != "" {
			result += fmt.Sprintf("   日期: %s\n", ref.Date)
		}
		result += "\n"
	}

	return result, nil
}

func NewWebSearchTool(apiKey string) *WebSearch {
	return &WebSearch{
		ApiKey: apiKey,
	}
}
