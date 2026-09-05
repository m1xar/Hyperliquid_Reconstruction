package executors

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const (
	positionRiskPath = "/fapi/v3/positionRisk"
	symbolConfigPath = "/fapi/v1/symbolConfig"
)

func FetchOpenPositions(client *resty.Client) ([]models.PositionRisk, error) {
	rows, err := doWithRateLimit(func() ([]models.PositionRisk, error) {
		return binance.DoGet[[]models.PositionRisk](client, positionRiskPath, nil, 5)
	})
	if err != nil {
		return nil, err
	}

	open := make([]models.PositionRisk, 0, len(rows))
	for _, row := range rows {
		amt, _ := strconv.ParseFloat(row.PositionAmt, 64)
		if amt == 0 {
			continue
		}
		open = append(open, row)
	}
	return open, nil
}

func FetchSymbolConfig(client *resty.Client) (map[string]models.SymbolConfig, error) {
	rows, err := doWithRateLimit(func() ([]models.SymbolConfig, error) {
		return binance.DoGet[[]models.SymbolConfig](client, symbolConfigPath, nil, 5)
	})
	if err != nil {
		return nil, err
	}

	out := make(map[string]models.SymbolConfig, len(rows))
	for _, row := range rows {
		out[row.Symbol] = row
	}
	return out, nil
}
