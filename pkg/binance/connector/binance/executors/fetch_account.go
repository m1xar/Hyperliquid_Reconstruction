package executors

import (
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const accountPath = "/fapi/v3/account"

var StableAssets = map[string]struct{}{
	"USDT":  {},
	"USDC":  {},
	"FDUSD": {},
	"BUSD":  {},
}

func IsStableAsset(asset string) bool {
	_, ok := StableAssets[strings.ToUpper(strings.TrimSpace(asset))]
	return ok
}

func FetchAccount(client *resty.Client) (models.Account, error) {
	return doWithRateLimit(func() (models.Account, error) {
		return binance.DoGet[models.Account](client, accountPath, nil, 5)
	})
}

func TotalMarginBalance(account models.Account) float64 {
	v, _ := strconv.ParseFloat(account.TotalMarginBalance, 64)
	return v
}

func StableWalletBalance(account models.Account) float64 {
	var total float64
	for _, asset := range account.Assets {
		if !IsStableAsset(asset.Asset) {
			continue
		}
		v, _ := strconv.ParseFloat(asset.WalletBalance, 64)
		total += v
	}
	return total
}
