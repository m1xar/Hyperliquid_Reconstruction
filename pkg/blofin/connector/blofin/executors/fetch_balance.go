package executors

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const balancePath = "/api/v1/account/balance"

var ProductTypes = []string{"USDT-FUTURES", "USDC-FUTURES", "COIN-FUTURES"}

func FetchBalance(client *resty.Client, baseURL, productType string) (models.Balance, error) {
	params := map[string]string{}
	if productType != "" {
		params["productType"] = productType
	}

	return doWithRateLimit(func() (models.Balance, error) {
		return blofin.DoGet[models.Balance](client, baseURL, balancePath, params)
	})
}

func FetchTotalEquity(client *resty.Client, baseURL string) (float64, error) {
	var total float64
	var firstErr error
	answered := false

	for _, productType := range ProductTypes {
		balance, err := FetchBalance(client, baseURL, productType)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		answered = true
		equity, convErr := strconv.ParseFloat(balance.TotalEquity, 64)
		if convErr != nil {
			continue
		}
		total += equity
	}

	if !answered {
		return 0, firstErr
	}
	return total, nil
}
