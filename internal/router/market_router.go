package router

import (
	"blog.alphazer01214.top/internal/api"
	"github.com/gin-gonic/gin"
)

func SetupMarketRouter(r *gin.Engine) {
	marketApi := api.Api.MarketApi

	public := r.Group("/api")
	{
		public.GET("/market", marketApi.GetIndices)
		public.GET("/market/history", marketApi.GetHistory)
	}
}
