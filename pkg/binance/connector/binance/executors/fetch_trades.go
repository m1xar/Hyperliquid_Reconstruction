package executors

import (
	"fmt"
	"sort"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const userTradesPath = "/fapi/v1/userTrades"

const tradesPageLimit = 1000

func FetchAllUserTrades(client *resty.Client, symbol string) ([]models.Trade, error) {
	fills, err := fetchTradesByID(client, symbol, 0)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(fills, func(i, j int) bool {
		if fills[i].Time == fills[j].Time {
			return fills[i].ID < fills[j].ID
		}
		return fills[i].Time < fills[j].Time
	})
	return fills, nil
}

func fetchTradesByID(client *resty.Client, symbol string, fromID int64) ([]models.Trade, error) {
	var result []models.Trade
	seen := make(map[int64]struct{})

	for {
		params := map[string]string{
			"symbol": symbol,
			"fromId": fmt.Sprint(fromID),
			"limit":  fmt.Sprintf("%d", tradesPageLimit),
		}

		page, err := doWithRateLimit(func() ([]models.Trade, error) {
			return binance.DoGet[[]models.Trade](client, userTradesPath, params, 5)
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

		added := 0
		maxID := int64(0)
		for _, fill := range page {
			if fill.ID > maxID {
				maxID = fill.ID
			}
			if _, ok := seen[fill.ID]; ok {
				continue
			}
			seen[fill.ID] = struct{}{}
			result = append(result, fill)
			added++
		}

		if len(page) < tradesPageLimit || added == 0 || maxID < fromID {
			break
		}
		fromID = maxID + 1
	}

	return result, nil
}
