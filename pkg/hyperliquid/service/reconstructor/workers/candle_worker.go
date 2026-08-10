package workers

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/hyperliquid/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/hyperliquid/connector/hyperliquid/executors"
	"github.com/m1xar/scope360-reconstruction/pkg/hyperliquid/connector/hyperliquid/models"
	"github.com/m1xar/scope360-reconstruction/pkg/hyperliquid/service/reconstructor/helpers"
)

func StartCandleWorkers(
	client *resty.Client,
	endpoint string,
	requests <-chan helpers.CandleRequest,
	workerCount int,
) {
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range requests {
				candles, err := fetchCandles(client, endpoint, req)
				req.ReplyCh <- helpers.CandleResponse{Candles: candles, Err: err}
			}
		}()
	}
	go func() {
		wg.Wait()
	}()
}

func fetchCandles(client *resty.Client, endpoint string, req helpers.CandleRequest) ([]models.HyperliquidCandle, error) {
	intervalMs, _ := helpers.IntervalToMs(req.Interval)
	oldestAllowedMs := time.Now().UnixMilli() - intervalMs*5000

	// Recent enough for Hyperliquid candle API — stay on HL.
	if req.StartMs >= oldestAllowedMs {
		return executors.FetchAllCandlesHyperliquid(
			client,
			endpoint,
			req.Coin,
			req.Interval,
			req.StartMs,
			req.EndMs,
		)
	}

	// Older than HL window: pull Binance for the old slice, HL for the recent slice.
	var out []models.HyperliquidCandle

	binanceEnd := req.EndMs
	if binanceEnd > oldestAllowedMs {
		binanceEnd = oldestAllowedMs - 1
	}
	if binanceEnd >= req.StartMs {
		binanceCandles, err := binance.FetchFuturesKlinesPaged(
			client,
			req.Coin,
			req.Interval,
			req.StartMs,
			binanceEnd,
			499,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, binanceCandles...)
	}

	if req.EndMs >= oldestAllowedMs {
		hlStart := oldestAllowedMs
		if hlStart < req.StartMs {
			hlStart = req.StartMs
		}
		hlCandles, err := executors.FetchAllCandlesHyperliquid(
			client,
			endpoint,
			req.Coin,
			req.Interval,
			hlStart,
			req.EndMs,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, hlCandles...)
	}

	return out, nil
}
