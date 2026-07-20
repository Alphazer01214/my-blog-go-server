package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/response"
	"github.com/elastic/go-elasticsearch/v8"
)

// SearchService Elasticsearch 搜索服务
type SearchService struct {
	client *elasticsearch.Client
	index  string
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword  string   `json:"keyword"`
	Tags     []string `json:"tags,omitempty"`
	Category string   `json:"category,omitempty"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Items    []PostSearchHit `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// PostSearchHit 搜索命中项
type PostSearchHit struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Tags          []string  `json:"tags"`
	Category      string    `json:"category"`
	UserId        uint      `json:"user_id"`
	AuthorName    string    `json:"author_name"`
	ViewCount     int       `json:"view_count"`
	LikeCount     int       `json:"like_count"`
	CommentCount  int       `json:"comment_count"`
	CreatedAt     time.Time `json:"created_at"`
	Highlight     string    `json:"highlight,omitempty"` // 高亮片段
}

// NewSearchService 创建搜索服务
func NewSearchService(addresses []string, index string) (*SearchService, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	svc := &SearchService{
		client: client,
		index:  index,
	}

	// 初始化索引
	if err := svc.initIndex(context.Background()); err != nil {
		slog.Warn("elasticsearch init index failed", "err", err)
	}

	return svc, nil
}

// initIndex 创建索引（如果不存在）
func (s *SearchService) initIndex(ctx context.Context) error {
	res, err := s.client.Indices.Exists([]string{s.index})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil // 索引已存在
	}

	// 创建索引，配置 IK 中文分词
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"ik_max": {
						"type": "custom",
						"tokenizer": "ik_max_word"
					},
					"ik_smart": {
						"type": "custom",
						"tokenizer": "ik_smart"
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id":           { "type": "integer" },
				"title":        { "type": "text", "analyzer": "ik_max", "search_analyzer": "ik_smart" },
				"content":      { "type": "text", "analyzer": "ik_max", "search_analyzer": "ik_smart" },
				"tags":         { "type": "keyword" },
				"category":     { "type": "keyword" },
				"user_id":      { "type": "integer" },
				"author_name":  { "type": "keyword" },
				"view_count":   { "type": "integer" },
				"like_count":   { "type": "integer" },
				"comment_count": { "type": "integer" },
				"is_public":    { "type": "boolean" },
				"created_at":   { "type": "date" }
			}
		}
	}`

	res, err = s.client.Indices.Create(s.index, s.client.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}

	slog.Info("elasticsearch index created", "index", s.index)
	return nil
}

// IndexPost 索引帖子文档
func (s *SearchService) IndexPost(ctx context.Context, post *entity.Post, authorName string) error {
	doc := map[string]interface{}{
		"id":            post.ID,
		"title":         post.Title,
		"content":       post.Content,
		"tags":          post.Tags,
		"category":      post.Category,
		"user_id":       post.UserId,
		"author_name":   authorName,
		"view_count":    post.ViewCount,
		"like_count":    post.LikeCount,
		"comment_count": post.CommentCount,
		"is_public":     post.Public,
		"created_at":    post.CreatedAt,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := s.client.Index(
		s.index,
		bytes.NewReader(data),
		s.client.Index.WithDocumentID(fmt.Sprintf("%d", post.ID)),
		s.client.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("index post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index post error: %s", res.String())
	}

	return nil
}

// DeletePost 删除帖子文档
func (s *SearchService) DeletePost(ctx context.Context, postID uint) error {
	res, err := s.client.Delete(
		s.index,
		fmt.Sprintf("%d", postID),
		s.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete post error: %s", res.String())
	}

	return nil
}

// SearchPosts 搜索帖子（支持高亮）
func (s *SearchService) SearchPosts(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	// 构建查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  req.Keyword,
							"fields": []string{"title^3", "content", "tags^2", "author_name"},
						},
					},
				},
				"filter": s.buildFilters(req),
			},
		},
		"highlight": map[string]interface{}{
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{"fragment_size": 150, "number_of_fragments": 3},
			},
		},
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		},
		"from": (req.Page - 1) * req.PageSize,
		"size": req.PageSize,
	}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID        string               `json:"_id"`
				Source    PostSearchHit         `json:"_source"`
				Highlight map[string][]string   `json:"highlight"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search result: %w", err)
	}

	items := make([]PostSearchHit, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		item := hit.Source
		// 提取高亮
		if hl, ok := hit.Highlight["title"]; ok && len(hl) > 0 {
			item.Highlight = hl[0]
		} else if hl, ok := hit.Highlight["content"]; ok && len(hl) > 0 {
			item.Highlight = hl[0]
		}
		items = append(items, item)
	}

	return &SearchResult{
		Items:    items,
		Total:    result.Hits.Total.Value,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// buildFilters 构建过滤条件
func (s *SearchService) buildFilters(req *SearchRequest) []map[string]interface{} {
	var filters []map[string]interface{}

	if req.Category != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{"category": req.Category},
		})
	}

	if len(req.Tags) > 0 {
		filters = append(filters, map[string]interface{}{
			"terms": map[string][]string{"tags": req.Tags},
		})
	}

	// 只搜索公开帖子
	filters = append(filters, map[string]interface{}{
		"term": map[string]bool{"is_public": true},
	})

	return filters
}

// ConvertToPostList 将搜索结果转换为 PostList 格式（兼容现有接口）
func (s *SearchService) ConvertToPostList(result *SearchResult, toDetail func(*PostSearchHit) response.PostDetail) response.PostList {
	items := make([]response.PostDetail, 0, len(result.Items))
	for _, hit := range result.Items {
		items = append(items, toDetail(&hit))
	}
	return response.PostList{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

// IsAvailable 检查 ES 是否可用
func (s *SearchService) IsAvailable(ctx context.Context) bool {
	res, err := s.client.Ping(s.client.Ping.WithContext(ctx))
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return !res.IsError()
}
