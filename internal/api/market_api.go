package api

import (
	"strconv"
	"time"

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

	rp, err := marketService.FetchHistory(from, to)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}
	response.SuccessWithDetail(c, rp, "market history")
}
