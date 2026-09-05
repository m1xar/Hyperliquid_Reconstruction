package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/bybit/connector/bybit/models"
)

const (
	instrumentsPath      = "/v5/market/instruments-info"
	instrumentsPageLimit = 1000
)

func FetchInstruments(client *resty.Client) (map[string]models.Instrument, error) {
	rows, err := collectCursor[models.Instrument](client, instrumentsPath, map[string]string{
		"category": models.CategoryLinear,
	}, instrumentsPageLimit)
	if err != nil {
		return nil, err
	}

	instruments := make(map[string]models.Instrument, len(rows))
	for _, inst := range rows {
		instruments[inst.Symbol] = inst
	}
	return instruments, nil
}
