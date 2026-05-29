package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/global"
	"blog.alphazer01214.top/internal/response"
	"github.com/gin-gonic/gin"
)

type MarketApi struct{}

func (ma *MarketApi) GetIndices(c *gin.Context) {
	rp, err := marketService.FetchIndices()
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "market data")
}

func (ma *MarketApi) GetHistory(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	to = time.Now()

	if fromStr != "" {
		ts, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			response.ErrorWithMsg(c, "invalid from timestamp")
			return
		}
		from = time.UnixMilli(ts)
	} else {
		from = to.Add(-10 * time.Minute)
	}

	if toStr != "" {
		ts, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			response.ErrorWithMsg(c, "invalid to timestamp")
			return
		}
		to = time.UnixMilli(ts)
	}

	rp, err := marketService.FetchHistory(from, to)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "market history")
}

// StreamMarket SSE 实时推送市场数据
func (ma *MarketApi) StreamMarket(c *gin.Context) {
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

	ticker := time.NewTicker(time.Duration(global.GetConfig().Market.RefreshInterval) * time.Second)
	defer ticker.Stop()

	// 首次推送
	rp, err := marketService.FetchIndices()
	if err == nil {
		if err := writeSSE(rp); err != nil {
			return
		}
	}

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			rp, err := marketService.FetchIndices()
			if err != nil {
				continue
			}
			if err := writeSSE(rp); err != nil {
				return
			}
		}
	}
}

// ExchangeRate 查询汇率
func (ma *MarketApi) ExchangeRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	moneyStr := c.DefaultQuery("money", "1")

	if from == "" || to == "" {
		response.ErrorWithMsg(c, "from and to parameters are required")
		return
	}

	money, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		response.ErrorWithMsg(c, "invalid money parameter")
		return
	}

	rp, err := marketService.FetchExchangeRate(from, to, money)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "exchange rate")
}

// GetCurrencyCodes 获取货币代码大全
func (ma *MarketApi) GetCurrencyCodes(c *gin.Context) {
	codes, err := marketService.FetchCurrencyCodes()
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, codes, "currency codes")
}
