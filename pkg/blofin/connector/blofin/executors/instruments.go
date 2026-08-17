package executors

import (
	"github.com/go-resty/resty/v2"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin"
	"github.com/m1xar/scope360-reconstruction/pkg/blofin/connector/blofin/models"
)

const instrumentsPath = "/api/v1/market/instruments"

func FetchInstruments(client *resty.Client, baseURL string) (map[string]models.Instrument, error) {
	data, err := doWithRateLimit(func() ([]models.Instrument, error) {
		return blofin.DoGet[[]models.Instrument](client, baseURL, instrumentsPath, nil)
	})
	if err != nil {
		return nil, err
	}

	instruments := make(map[string]models.Instrument, len(data))
	for _, inst := range data {
		instruments[inst.InstID] = inst
	}
	return instruments, nil
}

func Instrument(instruments map[string]models.Instrument, instID string) models.Instrument {
	if inst, ok := instruments[instID]; ok && inst.ContractValue != "" {
		return inst
	}
	return models.DefaultInstrument
}
