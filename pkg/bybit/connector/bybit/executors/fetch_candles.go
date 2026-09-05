package executors

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	klinePath       = "/v5/market/kline"
	candlesMaxLimit = 1000
)

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

func BybitInterval(interval string) string {
	s := strings.TrimSpace(interval)
	if len(s) < 2 {
		return s
	}
	n, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
	if err != nil {
		return s
	}
	switch s[len(s)-1] {
	case 'm':
		return fmt.Sprint(n)
	case 'h':
		return fmt.Sprint(n * 60)
	case 'd':
		return "D"
	case 'w':
		return "W"
	default:
		return s
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
	nextEnd := endMs

	for {
		if startMs > 0 && nextEnd < startMs {
			break
		}

		limit := candlesMaxLimit
		if step > 0 && startMs > 0 && nextEnd > 0 {
			if need := (nextEnd-startMs)/step + 1; need < int64(limit) {
				limit = int(need)
			}
		}
		if limit < 1 {
			limit = 1
		}

		params := map[string]string{
			"category": models.CategoryLinear,
			"symbol":   symbol,
			"interval": BybitInterval(interval),
			"limit":    fmt.Sprintf("%d", limit),
		}
		if startMs > 0 {
			params["start"] = fmt.Sprint(startMs)
		}
		if nextEnd > 0 {
			params["end"] = fmt.Sprint(nextEnd)
		}

		page, err := doWithRateLimit(func() (models.KlineResult, error) {
			return bybit.DoGet[models.KlineResult](client, klinePath, params)
		})
		if err != nil {
			if len(result) > 0 && isHTTP5xx(err) {
				break
			}
			return nil, err
		}
		if len(page.List) == 0 {
			break
		}

		result = append(result, page.List...)

		oldest := page.List[len(page.List)-1].StartTime
		if len(page.List) < limit || oldest <= 0 || oldest > nextEnd {
			break
		}
		nextEnd = oldest - 1
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})
	return result, nil
}

func FetchLatestClose(client *resty.Client, symbol string) (models.Candle, error) {
	params := map[string]string{
		"category": models.CategoryLinear,
		"symbol":   symbol,
		"interval": "1",
		"limit":    "1",
	}
	page, err := doWithRateLimit(func() (models.KlineResult, error) {
		return bybit.DoGet[models.KlineResult](client, klinePath, params)
	})
	if err != nil {
		return models.Candle{}, err
	}
	if len(page.List) == 0 {
		return models.Candle{}, fmt.Errorf("bybit: no klines for %s", symbol)
	}
	return page.List[0], nil
}
