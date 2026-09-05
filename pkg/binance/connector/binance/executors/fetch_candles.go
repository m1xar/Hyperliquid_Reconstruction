package executors

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const klinesPath = "/fapi/v1/klines"

const candlesMaxLimit = 1500

func klineWeight(limit int) int {
	switch {
	case limit < 100:
		return 1
	case limit < 500:
		return 2
	case limit <= 1000:
		return 5
	default:
		return 10
	}
}

func IntervalMs(interval string) int64 {
	s := strings.TrimSpace(interval)
	if len(s) < 2 {
		return 0
	}
	n, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	switch s[len(s)-1] {
	case 'm':
		return n * 60_000
	case 'h':
		return n * 3_600_000
	case 'd':
		return n * 86_400_000
	case 'w':
		return n * 7 * 86_400_000
	default:
		return 0
	}
}

func AlignToInterval(ts int64, interval string) int64 {
	step := IntervalMs(interval)
	if step <= 0 {
		return ts
	}
	return ts - ts%step
}

func FetchCandles(client *resty.Client, symbol, interval string, startMs, endMs int64) ([]models.Candle, error) {
	var result []models.Candle
	step := IntervalMs(interval)
	nextStart := startMs

	for {
		if endMs > 0 && nextStart > endMs {
			break
		}

		limit := candlesMaxLimit
		if step > 0 && endMs > 0 && nextStart > 0 {
			if need := (endMs-nextStart)/step + 1; need < int64(limit) {
				limit = int(need)
			}
		}
		if limit < 1 {
			limit = 1
		}

		params := map[string]string{
			"symbol":   symbol,
			"interval": interval,
			"limit":    fmt.Sprintf("%d", limit),
		}
		if nextStart > 0 {
			params["startTime"] = fmt.Sprint(nextStart)
		}
		if endMs > 0 {
			params["endTime"] = fmt.Sprint(endMs)
		}

		page, err := doWithRateLimit(func() ([]models.Candle, error) {
			return binance.DoGet[[]models.Candle](client, klinesPath, params, klineWeight(limit))
		})
		if err != nil {
			if len(result) > 0 && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		result = append(result, page...)

		lastOpen := page[len(page)-1].OpenTime
		if len(page) < limit || lastOpen <= 0 || lastOpen < nextStart {
			break
		}
		nextStart = lastOpen + 1
	}

	return result, nil
}

func FetchLatestClose(client *resty.Client, symbol string) (models.Candle, error) {
	params := map[string]string{
		"symbol":   symbol,
		"interval": "1m",
		"limit":    "1",
	}
	page, err := doWithRateLimit(func() ([]models.Candle, error) {
		return binance.DoGet[[]models.Candle](client, klinesPath, params, 1)
	})
	if err != nil {
		return models.Candle{}, err
	}
	if len(page) == 0 {
		return models.Candle{}, fmt.Errorf("binance: no klines for %s", symbol)
	}
	return page[len(page)-1], nil
}
