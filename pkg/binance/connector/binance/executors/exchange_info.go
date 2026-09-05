package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance"
	"github.com/m1xar/scope360-reconstruction/pkg/binance/connector/binance/models"
)

const exchangeInfoPath = "/fapi/v1/exchangeInfo"

func FetchInstruments(client *resty.Client) (map[string]models.Instrument, error) {
	info, err := doWithRateLimit(func() (models.ExchangeInfo, error) {
		return binance.DoGet[models.ExchangeInfo](client, exchangeInfoPath, nil, 1)
	})
	if err != nil {
		return nil, err
	}

	instruments := make(map[string]models.Instrument, len(info.Symbols))
	for _, inst := range info.Symbols {
		instruments[inst.Symbol] = inst
	}
	return instruments, nil
}
