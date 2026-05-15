package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/entity"
	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"

	"github.com/redis/go-redis/v9"
)

const (
	marketLatestKey  = "market:latest"
	marketHistoryKey = "market:history"
	marketHistoryTTL = 24 * time.Hour
	httpTimeout      = 8 * time.Second
)

type MarketService struct{}

func (ms *MarketService) CacheDuration() time.Duration {
	cfg := global.GetConfig()
	if cfg.Market != nil && cfg.Market.RefreshInterval > 0 {
		return time.Duration(cfg.Market.RefreshInterval) * time.Second
	}
	return 5 * time.Second
}

func (ms *MarketService) apiURL() string {
	cfg := global.GetConfig()
	if cfg.Market != nil && cfg.Market.ApiUrl != "" {
		return cfg.Market.ApiUrl
	}
	return "https://proxy.finance.qq.com/ifzqgtimg/appstock/app/rank/indexRankDetail2"
}

// PollOnce 由 keepalive 后台定时调用，抓取上游数据并写入 Redis
func (ms *MarketService) PollOnce(ctx context.Context) error {
	rp := ms.doFetch()
	if rp == nil {
		return fmt.Errorf("fetch failed")
	}

	data, err := json.Marshal(rp)
	if err != nil {
		return err
	}

	dur := ms.CacheDuration()
	global.GetRedis().Set(context.Background(), marketLatestKey, data, dur)

	now := time.Now()
	pt := entity.HistoryPoint{
		Time:    now,
		Common:  rp.Common,
		America: rp.America,
		Europe:  rp.Europe,
		Asia:    rp.Asia,
		Other:   rp.Other,
	}
	ptData, _ := json.Marshal(pt)
	if ptData != nil {
		score := float64(now.UnixMilli())
		global.GetRedis().ZAdd(context.Background(), marketHistoryKey, redis.Z{
			Score:  score,
			Member: string(ptData),
		})
		global.GetRedis().Expire(context.Background(), marketHistoryKey, marketHistoryTTL)
		global.GetRedis().ZRemRangeByScore(context.Background(), marketHistoryKey,
			"0", strconv.FormatInt(time.Now().Add(-marketHistoryTTL).UnixMilli(), 10))
	}

	return nil
}

// FetchIndices 从 Redis 缓存读取最新数据
func (ms *MarketService) FetchIndices() (*response.MarketResponse, error) {
	cached, err := global.GetRedis().Get(context.Background(), marketLatestKey).Result()
	if err != nil {
		return &response.MarketResponse{
			UpdatedAt: time.Now(),
		}, nil
	}

	var rp response.MarketResponse
	if err := json.Unmarshal([]byte(cached), &rp); err != nil {
		return &response.MarketResponse{
			UpdatedAt: time.Now(),
		}, nil
	}
	return &rp, nil
}

func (ms *MarketService) FetchHistory(from, to time.Time) (*response.HistoryResponse, error) {
	opt := &redis.ZRangeBy{
		Min: strconv.FormatInt(from.UnixMilli(), 10),
		Max: strconv.FormatInt(to.UnixMilli(), 10),
	}
	raw, err := global.GetRedis().ZRangeByScoreWithScores(context.Background(), marketHistoryKey, opt).Result()
	if err != nil {
		return nil, err
	}

	points := make([]entity.HistoryPoint, 0, len(raw))
	for _, z := range raw {
		var pt entity.HistoryPoint
		if err := json.Unmarshal([]byte(z.Member.(string)), &pt); err != nil {
			continue
		}
		points = append(points, pt)
	}

	return &response.HistoryResponse{Points: points}, nil
}

func (ms *MarketService) doFetch() *response.MarketResponse {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(ms.apiURL())
	if err != nil {
		fmt.Printf("[market] http get error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[market] read body error: %v\n", err)
		return nil
	}

	var raw entity.IndexRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Printf("[market] unmarshal error: %v\n", err)
		return nil
	}

	if raw.Code != 0 {
		fmt.Printf("[market] api error: code=%d msg=%s\n", raw.Code, raw.Msg)
		return nil
	}

	return &response.MarketResponse{
		UpdatedAt: time.Now(),
		Common:    raw.Data.Common,
		America:   raw.Data.America,
		Europe:    raw.Data.Europe,
		Asia:      raw.Data.Asia,
		Other:     raw.Data.Other,
	}
}
