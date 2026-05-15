package response

import (
	"time"

	"blog.alphazer01214.top/internal/entity"
)

type MarketResponse struct {
	UpdatedAt time.Time          `json:"updated_at"`
	Common    []entity.IndexItem `json:"common"`
	America   []entity.IndexItem `json:"america"`
	Europe    []entity.IndexItem `json:"europe"`
	Asia      []entity.IndexItem `json:"asia"`
	Other     []entity.IndexItem `json:"other"`
}

type HistoryResponse struct {
	Points []entity.HistoryPoint `json:"points"`
}

