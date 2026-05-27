package api

import (
	"strconv"
	"time"

	"blog.alphazer01214.top/internal/response"
	"blog.alphazer01214.top/internal/service"
	"github.com/gin-gonic/gin"
)

type MarketApi struct {
	marketService *service.MarketService
}

func (ma *MarketApi) GetIndices(c *gin.Context) {
	rp, err := ma.marketService.GetIndices()
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
		from = to.Add(-1 * time.Hour)
	}

	if toStr != "" {
		ts, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			response.ErrorWithMsg(c, "invalid to timestamp")
			return
		}
		to = time.UnixMilli(ts)
	}

	rp, err := ma.marketService.GetHistory(from, to)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "market history")
}
