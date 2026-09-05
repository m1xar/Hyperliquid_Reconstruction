package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	positionListPath   = "/v5/position/list"
	positionsPageLimit = 200
)

var SettleCoins = []string{"USDT", "USDC"}

func FetchOpenPositions(client *resty.Client) ([]models.Position, error) {
	var open []models.Position
	for _, settle := range SettleCoins {
		rows, err := collectCursor[models.Position](client, positionListPath, map[string]string{
			"category":   models.CategoryLinear,
			"settleCoin": settle,
		}, positionsPageLimit)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if parse(row.Size) == 0 {
				continue
			}
			open = append(open, row)
		}
	}
	return open, nil
}
