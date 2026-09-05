package helpers

import (
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

type CandleRequest struct {
	Symbol   string
	Interval string
	StartMs  int64
	EndMs    int64
	ReplyCh  chan<- CandleResponse
}

type CandleResponse struct {
	Candles []models.Candle
	Err     error
}
