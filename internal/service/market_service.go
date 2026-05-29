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
	marketHistoryTTL = 10 * time.Minute
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

// exchangeRateAPIURL 获取汇率API地址
func (ms *MarketService) exchangeRateAPIURL() string {
	cfg := global.GetConfig()
	if cfg.Market != nil && cfg.Market.ExchangeRateUrl != "" {
		return cfg.Market.ExchangeRateUrl
	}
	return "https://cn.apihz.cn/api/jinrong/huilv.php"
}

// ExchangeRateRequest 汇率查询请求
type ExchangeRateRequest struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Money float64 `json:"money"`
}

// FlexFloat64 兼容 JSON 字符串和数字的浮点类型
type FlexFloat64 float64

func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexFloat64(num)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		*f = FlexFloat64(num)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into float64", string(data))
}

// ExchangeRateResponse 汇率查询响应
type ExchangeRateResponse struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg,omitempty"`
	Uptime string      `json:"uptime,omitempty"`
	Money  string      `json:"money,omitempty"`
	From   string      `json:"from,omitempty"`
	To     string      `json:"to,omitempty"`
	Result FlexFloat64 `json:"result,omitempty"`
	Rate   FlexFloat64 `json:"rate,omitempty"`
}

// FetchExchangeRate 查询汇率
func (ms *MarketService) FetchExchangeRate(from, to string, money float64) (*response.ExchangeRateData, error) {
	apiURL := fmt.Sprintf("%s?id=88888888&key=88888888&from=%s&to=%s&money=%.2f",
		ms.exchangeRateAPIURL(), from, to, money)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	var rateResp ExchangeRateResponse
	if err := json.Unmarshal(body, &rateResp); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	if rateResp.Code != 200 {
		return nil, fmt.Errorf("api error: %s", rateResp.Msg)
	}

	return &response.ExchangeRateData{
		Uptime: rateResp.Uptime,
		From:   rateResp.From,
		To:     rateResp.To,
		Money:  rateResp.Money,
		Result: float64(rateResp.Result),
		Rate:   float64(rateResp.Rate),
	}, nil
}

// CurrencyCode 货币代码
type CurrencyCode struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// FetchCurrencyCodes 获取货币代码大全
func (ms *MarketService) FetchCurrencyCodes() ([]CurrencyCode, error) {
	apiURL := fmt.Sprintf("%s?id=88888888&key=88888888", ms.exchangeRateAPIURL())

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	var result struct {
		Code    int            `json:"code"`
		Message string         `json:"message,omitempty"`
		Data    []CurrencyCode `json:"data,omitempty"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("api error: %s", result.Message)
	}

	return result.Data, nil
}
